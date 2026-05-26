package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"joblinker/internal/adapters"
	"joblinker/internal/engine"
	"joblinker/internal/engine/nodes"

	"github.com/google/uuid"
)

// AgentHandler exposes agent conversation endpoints over HTTP.
// It uses the engine Runner for message processing.
type AgentHandler struct {
	runner    *engine.Runner
	rmq       *adapters.RabbitMQ
	toolNode  *nodes.ToolNode
	aiClient  *adapters.AIClient
}

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(runner *engine.Runner, rmq *adapters.RabbitMQ, toolNode *nodes.ToolNode, aiClient *adapters.AIClient) *AgentHandler {
	return &AgentHandler{
		runner:   runner,
		rmq:      rmq,
		toolNode: toolNode,
		aiClient: aiClient,
	}
}

// SendMessageRequest is the JSON payload for sending an agent message.
type SendMessageRequest struct {
	MatchID    string                 `json:"match_id"`
	SenderID   string                 `json:"sender_id"`
	ReceiverID string                 `json:"receiver_id"`
	Intent     string                 `json:"intent"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
}

// SendMessageResponse is the JSON response after sending a message.
type SendMessageResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

// HandleSendMessage handles POST /api/agent/message to send an agent message.
func (h *AgentHandler) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.MatchID == "" || req.SenderID == "" {
		http.Error(w, "match_id and sender_id required", http.StatusBadRequest)
		return
	}

	msg := &adapters.AgentMessage{
		MessageID:  uuid.New().String(),
		MatchID:    req.MatchID,
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Intent:     req.Intent,
		Payload:    req.Payload,
		Timestamp:  time.Now(),
	}

	if err := h.rmq.PublishAgentMessage(r.Context(), msg); err != nil {
		log.Printf("Failed to publish message: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(SendMessageResponse{
		MessageID: msg.MessageID,
		Status:    "queued",
	})
}

// HandleQueryJobs handles POST /api/agent/tools/query-jobs.
func (h *AgentHandler) HandleQueryJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var args struct {
		Location  string   `json:"location"`
		Skills    []string `json:"skills"`
		SalaryMin int      `json:"salary_min"`
		JobType   string   `json:"job_type"`
		Limit     int      `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	input := map[string]interface{}{
		"tool":      "query_jobs",
		"arguments": map[string]interface{}{
			"location":   args.Location,
			"skills":     args.Skills,
			"salary_min": args.SalaryMin,
			"job_type":   args.JobType,
			"limit":      args.Limit,
		},
	}

	output, err := h.toolNode.Process(r.Context(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("tool execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

// HandleChat handles POST /api/agent/chat for direct AI conversation.
func (h *AgentHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message   string `json:"message"`
		AgentType string `json:"agent_type"`
		Context   string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "message required", http.StatusBadRequest)
		return
	}

	if req.AgentType == "" {
		req.AgentType = "recruiter"
	}

	respBody, err := h.aiClient.GenerateAgentResponse(req.Context, req.AgentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("AI generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": respBody,
	})
}

// Healthz handles GET /healthz.
func (h *AgentHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if !h.rmq.IsConnected() {
		status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   status,
		"service":  "agent-engine",
		"version":  "2.0.0",
	})
}

// RegisterRoutes registers all HTTP handler routes.
// Use gin.WrapH() to wrap handlers for gin: r.POST("/path", gin.WrapH(h.HandleXxx))
func (h *AgentHandler) RegisterRoutes(addRoute func(method, path string, handler http.HandlerFunc)) {
	addRoute("POST", "/api/agent/message", h.HandleSendMessage)
	addRoute("POST", "/api/agent/tools/query-jobs", h.HandleQueryJobs)
	addRoute("POST", "/api/agent/chat", h.HandleChat)
	addRoute("GET", "/healthz", h.Healthz)
}

// StartConsumer starts the engine Runner in a goroutine.
func (h *AgentHandler) StartConsumer(ctx context.Context) error {
	log.Println("[AgentHandler] starting consumer")
	go func() {
		if err := h.runner.Start(ctx); err != nil {
			log.Printf("[AgentHandler] runner exited: %v", err)
		}
	}()
	return nil
}
