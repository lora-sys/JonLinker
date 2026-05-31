package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/lora-sys/JonLinker/internal/agent"
	"github.com/lora-sys/JonLinker/internal/config"
	"github.com/lora-sys/JonLinker/internal/job"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	agt, err := agent.New(ctx, cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.FirecrawlKey)
	if err != nil {
		log.Fatalf("init agent: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/search", handleSearch(agt))

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

func handleSearch(agt *agent.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req job.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		msg, err := agt.Search(r.Context(), req.Query)
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
