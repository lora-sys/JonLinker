package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/agent"
	"github.com/lora-sys/JonLinker/internal/checkpoint"
	"github.com/lora-sys/JonLinker/internal/config"
	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/llm"
	"github.com/lora-sys/JonLinker/internal/memory"
	"github.com/lora-sys/JonLinker/internal/resume"
	"github.com/lora-sys/JonLinker/internal/middleware"
	"github.com/lora-sys/JonLinker/internal/sse"
	"github.com/lora-sys/JonLinker/internal/tools/applyjob"
	"github.com/lora-sys/JonLinker/internal/tools/parseresume"
)

type resumeAgentEntry struct {
	agent *agent.ResumeAgent
	time  time.Time
}

type server struct {
	cfg           *config.Config
	searchAgent   *agent.Agent
	memStore      memory.MemoryStore
	checkpoint    *checkpoint.Store
	directApply   *applyjob.DirectApply
	resumeParser  *parseresume.Tool
	resumeAgents  map[string]*resumeAgentEntry
	mu            sync.Mutex
	cleanupCancel context.CancelFunc
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	cpStore := checkpoint.NewPersistentStore(cfg.SessionFile)
	dApply, err := applyjob.NewDirectApply(cpStore, cfg.FirecrawlKey, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.OpenAIModel)
	if err != nil {
		log.Fatalf("init direct apply: %v", err)
	}

	agentMem := memory.NewStore(cfg)
	srchAgent, err := agent.New(ctx, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.FirecrawlKey, dApply, agentMem, cpStore)
	if err != nil {
		log.Fatalf("init search agent: %v", err)
	}
	rParser := parseresume.NewTool(cfg.FirecrawlKey)

	srv := &server{
		cfg:          cfg,
		searchAgent:  srchAgent,
		memStore:     agentMem,
		checkpoint:   cpStore,
		directApply:  dApply,
		resumeParser: rParser,
		resumeAgents: make(map[string]*resumeAgentEntry),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("health encode: %v", err)
	}
	})
	mux.HandleFunc("POST /api/resume/upload", srv.handleUpload)
	mux.HandleFunc("POST /api/chat/unified", srv.handleUnifiedChat)
	mux.HandleFunc("POST /api/apply", srv.handleApply)

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	srv.cleanupCancel = cleanupCancel
	go srv.cleanupLoop(cleanupCtx)

	handler := middleware.CORS(cfg.FrontendURL)(mux)

	log.Printf("server starting on :%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, handler); err != nil {
		log.Fatalf("serve: %v", err)
	}
}



type chatMessage struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
	// AI SDK v6 sends parts instead of content
	Parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"parts"`
}

func (m chatMessage) text() string {
	if m.Content != "" {
		return m.Content
	}
	for _, p := range m.Parts {
		if p.Type == "text" {
			return p.Text
		}
	}
	return ""
}

func parseAgentResponse(content string) job.SearchResponse {
	var resp job.SearchResponse

	// Try full JSON parse first
	if err := json.Unmarshal([]byte(content), &resp); err == nil {
		return resp
	}

	// Extract first valid JSON block between ===JSON=== and ===END===
	// Use absolute positions throughout — never mix offsets from different substrings
	searchFrom := 0
	for {
		start := strings.Index(content[searchFrom:], "===JSON===")
		if start < 0 {
			break
		}
		start += searchFrom

		blockStart := start + 10
		end := strings.Index(content[blockStart:], "===END===")
		if end < 0 {
			break
		}
		end += blockStart

		raw := strings.TrimSpace(content[blockStart:end])
		if err := json.Unmarshal([]byte(raw), &resp); err == nil {
			break
		}

		searchFrom = end + 9
	}

	// If marker extraction failed, try greedy JSON extraction (agent sometimes omits markers)
	if len(resp.Jobs) == 0 && resp.Application == nil {
		searchFrom = 0
		for _, prefix := range []string{`{"jobs":`, `{"application":`} {
			start := strings.Index(content[searchFrom:], prefix)
			if start < 0 {
				continue
			}
			start += searchFrom
			if end := strings.LastIndex(content[start:], "}"); end >= 0 {
				raw := content[start : start+end+1]
				if err := json.Unmarshal([]byte(raw), &resp); err == nil {
					// Strip extracted JSON from text to prevent raw JSON leaking into chat output
					content = strings.Replace(content, raw, "", 1)
					break
				}
			}
		}
	}

	// Remove all ===JSON===...===END=== markers to extract clean human-readable text
	text := content
	for {
		start := strings.Index(text, "===JSON===")
		if start < 0 {
			break
		}
		end := strings.Index(text[start:], "===END===")
		if end < 0 {
			break
		}
		text = text[:start] + text[start+end+9:]
	}
	text = strings.TrimSpace(text)
	if text != "" {
		resp.Message = text
	}

	// If agent returned an empty jobs list with no text, set a message
	if resp.Message == "" && len(resp.Jobs) == 0 {
		resp.Jobs = nil
		if resp.Message == "" {
			resp.Message = "暂时没有找到匹配的职位，试试其他关键词？"
		}
	}

	return resp
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		sse.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		sse.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		sse.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "read file failed"})
		return
	}

	text, err := s.resumeParser.ParseFile(ctx, header.Filename, data)
	if err != nil {
		log.Printf("parse resume: %v", err)
		sse.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "parse resume failed: " + err.Error()})
		return
	}

	sessionID := genSessionID()

	ragent, err := agent.NewResumeAgent(ctx,
		s.cfg.OpenAIBaseURL, s.cfg.OpenAIAPIKey, s.cfg.OpenAIModel, text, s.checkpoint, s.memStore)
	if err != nil {
		log.Printf("init resume agent: %v", err)
		sse.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "init agent failed"})
		return
	}

	s.mu.Lock()
	s.resumeAgents[sessionID] = &resumeAgentEntry{agent: ragent, time: time.Now()}
	s.mu.Unlock()

	// Immediately extract a preliminary profile so /api/apply works without prior chat
	if cm, err := llm.NewChatModel(ctx, s.cfg.OpenAIBaseURL, s.cfg.OpenAIAPIKey, s.cfg.OpenAIModel, 1024, 0.3); err == nil {
		extractPrompt := fmt.Sprintf(`Extract candidate profile from resume text as JSON. Use empty strings/arrays for missing fields.

%s

{"name":"","title":"","skills":[],"experience":[],"education":[],"phone":"","email":"","summary":"","hobbies":[]}`, text)
		if result, genErr := cm.Generate(ctx, []*schema.Message{{Role: schema.User, Content: extractPrompt}}); genErr == nil {
			content := llm.ExtractJSONBlock(result.Content)
			var profile resume.CandidateProfile
			if err := json.Unmarshal([]byte(content), &profile); err == nil && profile.Name != "" {
				if b, err := json.Marshal(profile); err == nil {
					if err := s.checkpoint.Set(ctx, sessionID+":profile", b); err != nil {
						log.Printf("save initial profile: %v", err)
					}
				}
			}
		}
	}

	sse.WriteJSON(w, http.StatusOK, job.UploadResponse{
		SessionID: sessionID,
		Text:      text[:min(len(text), 500)],
	})
}

func (s *server) handleUnifiedChat(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	var req struct {
		Messages  []chatMessage `json:"messages"`
		ID        string        `json:"id"`
		SessionID string        `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sse.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if len(req.Messages) == 0 {
		sse.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no messages"})
		return
	}

	last := req.Messages[len(req.Messages)-1]
	if last.Role != "user" {
		sse.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "last message must be from user"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		sse.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	userMsg := last.text()

	profileData, hasProfile, _ := s.checkpoint.Get(ctx, req.SessionID+":profile")

	s.mu.Lock()
	entry, hasResumeAgent := s.resumeAgents[req.SessionID]
	s.mu.Unlock()

	kind := agent.Route(hasResumeAgent, profileData, userMsg)

	// Ensure profile exists for search agent
	if kind == agent.AgentSearch && !hasProfile {
		kind = agent.AgentResume
	}

	msgID := "msg_" + strconv.Itoa(int(msgCounter.Add(1)))
	_ = sse.WriteEvent(w, map[string]string{"type": "start"})
	flusher.Flush()
	_ = sse.WriteEvent(w, map[string]string{"type": "text-start", "id": msgID})
	flusher.Flush()

	switch kind {
	case agent.AgentResume:
		if entry == nil {
			_ = sse.WriteEvent(w, map[string]any{
				"type":  "text-delta",
				"id":    msgID,
				"delta": "请先上传简历后再开始完善资料。",
			})
			flusher.Flush()
			_ = sse.WriteEvent(w, map[string]string{"type": "text-end", "id": msgID})
			_ = sse.WriteEvent(w, map[string]string{"type": "finish", "finishReason": "stop"})
			flusher.Flush()
			return
		}
		responseText, err := entry.agent.ChatStream(ctx, req.SessionID, userMsg, func(token string) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if token == "" {
				return
			}
			_ = sse.WriteEvent(w, map[string]any{
				"type":  "text-delta",
				"id":    msgID,
				"delta": token,
			})
			flusher.Flush()
		})
		if err != nil {
			log.Printf("resume chat: %v", err)
			_ = sse.WriteEvent(w, map[string]string{"type": "error", "errorText": err.Error()})
			flusher.Flush()
			_ = sse.WriteEvent(w, map[string]string{"type": "finish", "finishReason": "error"})
			flusher.Flush()
			return
		}

		_ = sse.WriteEvent(w, map[string]string{"type": "text-end", "id": msgID})
		flusher.Flush()

		jsonPart := llm.ExtractJSONBlock(responseText)

		var resumeState struct {
			Complete bool                    `json:"complete"`
			Message  string                  `json:"message,omitempty"`
			Profile  resume.CandidateProfile `json:"profile,omitempty"`
		}
		if err := json.Unmarshal([]byte(jsonPart), &resumeState); err == nil {
			_ = sse.WriteEvent(w, map[string]any{
				"type": "data-resume",
				"id":   msgID,
				"data": resumeState,
			})
			flusher.Flush()
		}

	case agent.AgentSearch:
		var resp job.SearchResponse
		msg, err := s.searchAgent.SearchStream(ctx, userMsg, req.SessionID, func(token string) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if token == "" {
				return
			}
			_ = sse.WriteEvent(w, map[string]any{
				"type":  "text-delta",
				"id":    msgID,
				"delta": token,
			})
			flusher.Flush()
		})
		if err != nil {
			_ = sse.WriteEvent(w, map[string]string{"type": "error", "errorText": err.Error()})
			flusher.Flush()
			_ = sse.WriteEvent(w, map[string]string{"type": "finish", "finishReason": "error"})
			flusher.Flush()
			return
		}

		resp = parseAgentResponse(msg.Content)

		_ = sse.WriteEvent(w, map[string]string{"type": "text-end", "id": msgID})
		flusher.Flush()

		if len(resp.Jobs) > 0 {
			_ = sse.WriteEvent(w, map[string]any{
				"type": "data-jobs",
				"id":   msgID,
				"data": resp.Jobs,
			})
			flusher.Flush()
		}
		if resp.Application != nil {
			_ = sse.WriteEvent(w, map[string]any{
				"type": "data-application",
				"id":   msgID,
				"data": resp.Application,
			})
			flusher.Flush()
		}
	}

	_ = sse.WriteEvent(w, map[string]string{"type": "finish", "finishReason": "stop"})
	flusher.Flush()
}

func (s *server) handleApply(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	var req job.ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sse.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	app, err := s.directApply.Generate(ctx, req.JobURL, req.SessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, applyjob.ErrProfileNotFound) {
			status = http.StatusBadRequest
		}
		log.Printf("apply error: %v", err)
		sse.WriteJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	sse.WriteJSON(w, http.StatusOK, app)
}



var sessionCounter atomic.Int64
var msgCounter atomic.Int64

func genSessionID() string {
	n := sessionCounter.Add(1)
	return "s" + time.Now().Format("150405") + "c" + strconv.Itoa(int(n))
}

func (s *server) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			for id, entry := range s.resumeAgents {
				if time.Since(entry.time) > 30*time.Minute {
					delete(s.resumeAgents, id)
					if err := s.checkpoint.DeletePrefix(ctx, id); err != nil {
						log.Printf("cleanup checkpoint: %v", err)
					}
				}
			}
			s.mu.Unlock()
		}
	}
}




