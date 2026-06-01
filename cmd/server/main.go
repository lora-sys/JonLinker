package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lora-sys/JonLinker/internal/agent"
	"github.com/lora-sys/JonLinker/internal/checkpoint"
	"github.com/lora-sys/JonLinker/internal/config"
	"github.com/lora-sys/JonLinker/internal/job"
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
	dApply := applyjob.NewDirectApply(cpStore, cfg.FirecrawlKey, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.OpenAIModel)

	srchAgent, err := agent.New(ctx, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.FirecrawlKey, dApply)
	if err != nil {
		log.Fatalf("init search agent: %v", err)
	}
	rParser := parseresume.NewTool(cfg.FirecrawlKey)

	srv := &server{
		cfg:          cfg,
		searchAgent:  srchAgent,
		checkpoint:   cpStore,
		directApply:  dApply,
		resumeParser: rParser,
		resumeAgents: make(map[string]*resumeAgentEntry),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/search", srv.handleSearch)
	mux.HandleFunc("POST /api/search/stream", srv.handleSearchStream)
	mux.HandleFunc("POST /api/chat", srv.handleAIChat)
	mux.HandleFunc("POST /api/resume/upload", srv.handleUpload)
	mux.HandleFunc("POST /api/chat/resume", srv.handleChat)
	mux.HandleFunc("POST /api/apply", srv.handleApply)

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	srv.cleanupCancel = cleanupCancel
	go srv.cleanupLoop(cleanupCtx)

	handler := corsMiddleware(mux, cfg.FrontendURL)

	log.Printf("server starting on :%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, handler); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req job.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	msg, err := s.searchAgent.Search(r.Context(), req.Query, req.SessionID)
	if err != nil {
		log.Printf("agent error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	content := msg.Content
	resp := parseAgentResponse(content)
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleSearchStream(w http.ResponseWriter, r *http.Request) {
	var req job.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	msg, err := s.searchAgent.SearchStream(r.Context(), req.Query, req.SessionID, func(thinking string) {
		_, _ = w.Write([]byte("data: ===THINKING===\n"))
		_, _ = w.Write([]byte("data: " + strings.ReplaceAll(thinking, "\n", "\\n") + "\n\n"))
		flusher.Flush()
	})
	if err != nil {
		log.Printf("agent error: %v", err)
		writeSSE(w, flusher, map[string]string{"type": "error", "errorText": err.Error()})
		writeSSE(w, flusher, map[string]string{"type": "finish", "finishReason": "error"})
		return
	}

	resp := parseAgentResponse(msg.Content)

	if resp.Message != "" {
		runes := []rune(resp.Message)
		chunkSize := 3
		for i := 0; i < len(runes); i += chunkSize {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			_, _ = w.Write([]byte("data: " + string(runes[i:end]) + "\n\n"))
			flusher.Flush()
			time.Sleep(15 * time.Millisecond)
		}
	}

	if resp.Jobs != nil || resp.Application != nil {
		jsonData, _ := json.Marshal(resp)
		_, _ = w.Write([]byte("data: ===JSON===\n"))
		_, _ = w.Write([]byte("data: " + string(jsonData) + "\n\n"))
		flusher.Flush()
	}

	_, _ = w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
}

func (s *server) handleAIChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Messages  []struct {
			ID      string `json:"id"`
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		ID        string `json:"id"`
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if len(req.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no messages"})
		return
	}

	// Last user message is the query
	last := req.Messages[len(req.Messages)-1]
	if last.Role != "user" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "last message must be from user"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Send start event
	msgID := "msg_" + strconv.Itoa(int(time.Now().UnixNano()))
	writeSSE(w, flusher, map[string]string{"type": "start"})
	writeSSE(w, flusher, map[string]string{"type": "text-start", "id": msgID})

	// Process through agent
	msg, err := s.searchAgent.Search(r.Context(), last.Content, req.SessionID)
	if err != nil {
		writeSSE(w, flusher, map[string]string{"type": "error", "errorText": err.Error()})
		writeSSE(w, flusher, map[string]string{"type": "finish", "finishReason": "error"})
		return
	}

	resp := parseAgentResponse(msg.Content)

	// Stream text in chunks
	if resp.Message != "" {
		runes := []rune(resp.Message)
		chunkSize := 3
		for i := 0; i < len(runes); i += chunkSize {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			writeSSE(w, flusher, map[string]any{
				"type":  "text-delta",
				"id":    msgID,
				"delta": string(runes[i:end]),
			})
			time.Sleep(15 * time.Millisecond)
		}
	} else {
		writeSSE(w, flusher, map[string]any{
			"type":  "text-delta",
			"id":    msgID,
			"delta": "",
		})
	}

	writeSSE(w, flusher, map[string]string{"type": "text-end", "id": msgID})

	// Send structured data as custom data events (type starts with "data-")
	// The `data` field of the event is what onData receives in the frontend.
	if len(resp.Jobs) > 0 {
		writeSSE(w, flusher, map[string]any{
			"type": "data-jobs",
			"id":   msgID,
			"data": resp.Jobs,
		})
	}
	if resp.Application != nil {
		writeSSE(w, flusher, map[string]any{
			"type": "data-application",
			"id":   msgID,
			"data": resp.Application,
		})
	}

	writeSSE(w, flusher, map[string]string{"type": "finish", "finishReason": "stop"})
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, v any) {
	b, _ := json.Marshal(v)
	_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
	flusher.Flush()
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

		blockStart := start + 9
		end := strings.Index(content[blockStart:], "===END===")
		if end < 0 {
			break
		}
		end += blockStart

		raw := strings.TrimSpace(content[blockStart:end])
		if err := json.Unmarshal([]byte(raw), &resp); err == nil {
			break
		}

		searchFrom = end + 8
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
		text = text[:start] + text[start+end+8:]
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
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read file failed"})
		return
	}

	text, err := s.resumeParser.ParseFile(r.Context(), header.Filename, data)
	if err != nil {
		log.Printf("parse resume: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "parse resume failed: " + err.Error()})
		return
	}

	sessionID := genSessionID()

	ragent, err := agent.NewResumeAgent(r.Context(),
		s.cfg.OpenAIBaseURL, s.cfg.OpenAIAPIKey, s.cfg.OpenAIModel, text, s.checkpoint)
	if err != nil {
		log.Printf("init resume agent: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "init agent failed"})
		return
	}

	s.mu.Lock()
	s.resumeAgents[sessionID] = &resumeAgentEntry{agent: ragent, time: time.Now()}
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, job.UploadResponse{
		SessionID: sessionID,
		Text:      text[:min(len(text), 500)],
	})
}

func (s *server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.mu.Lock()
	entry, ok := s.resumeAgents[req.SessionID]
	s.mu.Unlock()
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "session not found"})
		return
	}

	var err error
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	_, err = entry.agent.ChatStream(r.Context(), req.SessionID, req.Message, func(token string) {
		select {
		case <-r.Context().Done():
			return
		default:
		}
		if token == "" {
			return
		}
		_, _ = w.Write([]byte("data: " + strings.ReplaceAll(token, "\n", "\\n") + "\n\n"))
		flusher.Flush()
	})
	if err != nil {
		log.Printf("resume chat: %v", err)
	}
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
}

func (s *server) handleApply(w http.ResponseWriter, r *http.Request) {
	var req job.ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	app, err := s.directApply.Generate(r.Context(), req.JobURL, req.SessionID)
	if err != nil {
		log.Printf("apply error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func corsMiddleware(next http.Handler, frontendURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var sessionCounter atomic.Int64

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
				}
			}
			s.mu.Unlock()
		}
	}
}


