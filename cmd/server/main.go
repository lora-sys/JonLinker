package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lora-sys/JonLinker/internal/agent"
	"github.com/lora-sys/JonLinker/internal/config"
	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/session"
	"github.com/lora-sys/JonLinker/internal/tools/apply_job"
	"github.com/lora-sys/JonLinker/internal/tools/parse_resume"
)

type resumeAgentEntry struct {
	agent *agent.ResumeAgent
	time  time.Time
}

type server struct {
	cfg          *config.Config
	searchAgent  *agent.Agent
	sessionStore *session.Store
	directApply  *apply_job.DirectApply
	resumeParser *parse_resume.Tool
	resumeAgents map[string]*resumeAgentEntry
	mu           sync.Mutex
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	srchAgent, err := agent.New(ctx, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.FirecrawlKey)
	if err != nil {
		log.Fatalf("init search agent: %v", err)
	}

	sStore := session.NewStore()
	dApply := apply_job.NewDirectApply(sStore, cfg.FirecrawlKey, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.OpenAIModel)
	rParser := parse_resume.NewTool(cfg.FirecrawlKey)

	srv := &server{
		cfg:          cfg,
		searchAgent:  srchAgent,
		sessionStore: sStore,
		directApply:  dApply,
		resumeParser: rParser,
		resumeAgents: make(map[string]*resumeAgentEntry),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/search", srv.handleSearch)
	mux.HandleFunc("POST /api/resume/upload", srv.handleUpload)
	mux.HandleFunc("POST /api/chat/resume", srv.handleChat)
	mux.HandleFunc("POST /api/apply", srv.handleApply)

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

	msg, err := s.searchAgent.Search(r.Context(), req.Query)
	if err != nil {
		log.Printf("agent error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	content := msg.Content
	var resp job.SearchResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		if idx := strings.Index(content, "{"); idx >= 0 {
			content = content[idx:]
		}
		if idx := strings.LastIndex(content, "}"); idx >= 0 {
			content = content[:idx+1]
		}
		if err2 := json.Unmarshal([]byte(content), &resp); err2 != nil {
			log.Printf("agent non-json: %s", content[:min(500, len(content))])
			resp = job.SearchResponse{}
		}
	}
	writeJSON(w, http.StatusOK, resp)
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
		s.cfg.OpenAIBaseURL, s.cfg.OpenAIAPIKey, s.cfg.OpenAIModel, text, s.sessionStore)
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
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var sessionCounter int

func genSessionID() string {
	sessionCounter++
	return "s" + time.Now().Format("150405") + "c" + itoa(sessionCounter)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}


