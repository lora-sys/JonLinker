package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"joblinker/internal/adapters"
)

// TestSendMessageHandler tests the HandleSendMessage endpoint.
func TestSendMessageHandler(t *testing.T) {
	// Create a handler with nil dependencies (will fail with actual RMQ, but we test routing)
	h := &AgentHandler{}

	body := SendMessageRequest{
		MatchID:    "test-match-id",
		SenderID:   "test-sender",
		ReceiverID: "test-receiver",
		Intent:     "INQUIRY",
	}

	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/message", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleSendMessage(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Without RabbitMQ, we expect a 503 error
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 without RMQ, got %d", resp.StatusCode)
	}
}

// TestHealthz tests the Healthz endpoint.
func TestHealthz(t *testing.T) {
	h := &AgentHandler{}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	h.Healthz(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["status"] != "degraded" {
		t.Errorf("expected degraded status without RMQ, got %s", result["status"])
	}
	if result["service"] != "agent-engine" {
		t.Errorf("expected agent-engine, got %s", result["service"])
	}
}

// TestSendMessageBadMethod tests method validation.
func TestSendMessageBadMethod(t *testing.T) {
	h := &AgentHandler{}

	req := httptest.NewRequest(http.MethodGet, "/api/agent/message", nil)
	w := httptest.NewRecorder()

	h.HandleSendMessage(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

// TestRegisterRoutes ensures the route registration function works.
func TestRegisterRoutes(t *testing.T) {
	h := &AgentHandler{}

	routes := make(map[string]string)
	h.RegisterRoutes(func(method, path string, handler http.HandlerFunc) {
		routes[method+" "+path] = "registered"
	})

	expected := []string{
		"POST /api/agent/message",
		"POST /api/agent/tools/query-jobs",
		"POST /api/agent/chat",
		"GET /healthz",
	}

	for _, route := range expected {
		if _, ok := routes[route]; !ok {
			t.Errorf("missing route: %s", route)
		}
	}
}

// TestHealthzIntegration tests the health check with actual components.
func TestHealthzIntegration(t *testing.T) {
	ai := adapters.NewAIClient()
	h := &AgentHandler{
		aiClient: ai,
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	h.Healthz(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
