//go:build integration
// +build integration

// Real WebSocket integration tests.
// No t.Skip — WebSocket unavailability causes t.Fatal.
package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"joblinker/tests/testutil"
)

func wsDial(t *testing.T, path string) *websocket.Conn {
	serverURL := testutil.RequireEnv(t, "TEST_WS_URL")
	token := testutil.RequireEnv(t, "TEST_WS_TOKEN")

	header := make(http.Header)
	header.Set("Authorization", "Bearer "+token)

	ws, resp, err := websocket.DefaultDialer.Dial(serverURL+path, header)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v (response: %v)", err, resp)
	}
	return ws
}

func TestWebSocket_Connect(t *testing.T) {
	ws := wsDial(t, "/ws")
	defer ws.Close()

	// Read the connected message
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read connected message: %v", err)
	}

	var frame map[string]interface{}
	if err := json.Unmarshal(msg, &frame); err != nil {
		t.Fatalf("connected message is not valid JSON: %v", err)
	}

	// Frame must have a "type" or "message_type" field
	msgType, ok := frame["type"].(string)
	if !ok {
		msgType, _ = frame["message_type"].(string)
	}
	if msgType == "" {
		t.Error("connected message missing type field")
	}
}

func TestWebSocket_AuthRequired(t *testing.T) {
	serverURL := testutil.RequireEnv(t, "TEST_WS_URL")

	// No token — should fail to connect
	_, resp, err := websocket.DefaultDialer.Dial(serverURL+"/ws", nil)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected connection failure without token, but succeeded")
	}
	// err should indicate unauthorized
	if resp != nil {
		resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Errorf("expected 401 for missing token, got %d", resp.StatusCode)
		}
	}
}

func TestWebSocket_InvalidMatchID(t *testing.T) {
	// Should get error or be able to read an error message, not panic
	ws := wsDial(t, "/ws")
	defer ws.Close()

	// Send a message with an invalid match ID
	msg := map[string]interface{}{
		"match_id": "this-is-not-a-valid-match-id",
		"intent":   "INTRODUCTION",
		"content":  "test",
	}
	body, _ := json.Marshal(msg)
	if err := ws.WriteMessage(websocket.TextMessage, body); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Set a read deadline — server should respond with error or close
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		// Expected: either timeout or close
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "close") {
			t.Fatalf("unexpected error reading response: %v", err)
		}
	} else {
		// If we get a response, check it's not a hardcoded error
		var resp map[string]interface{}
		json.Unmarshal(msgBytes, &resp)
		if errMsg, ok := resp["error"].(string); ok {
			testutil.AssertNotHardcoded(t, "error_message", errMsg)
		}
	}
}

func TestWebSocket_MessageIsolation(t *testing.T) {
	// Create two different match contexts
	baseURL := testutil.RequireServerURL(t)
	token1, _ := testutil.MustRegister(t, baseURL, "seeker")
	token2, _ := testutil.MustRegister(t, baseURL, "seeker")

	agent1ID := testutil.MustCreateAgent(t, baseURL, token1, "seeker")
	agent2ID := testutil.MustCreateAgent(t, baseURL, token2, "seeker")

	recruiterToken, _ := testutil.MustRegister(t, baseURL, "recruiter")
	recruiterAgentID := testutil.MustCreateAgent(t, baseURL, recruiterToken, "recruiter")
	jobID := testutil.MustCreateJob(t, baseURL, recruiterToken)

	match1ID := testutil.MustCreateMatch(t, baseURL, token1, recruiterToken, agent1ID, recruiterAgentID, jobID)
	match2ID := testutil.MustCreateMatch(t, baseURL, token2, recruiterToken, agent2ID, recruiterAgentID, jobID)

	// Connect two WS connections
	ws1 := wsDial(t, "/ws?match_id="+match1ID)
	defer ws1.Close()
	ws2 := wsDial(t, "/ws?match_id="+match2ID)
	defer ws2.Close()

	// Read both connected messages
	ws1.SetReadDeadline(time.Now().Add(5 * time.Second))
	ws2.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, _ = ws1.ReadMessage()
	_, _, _ = ws2.ReadMessage()

	// Send message on match1 — verify match2 does NOT receive it
	testMsg := map[string]interface{}{
		"match_id": match1ID,
		"intent":   "INTRODUCTION",
		"content":  "isolated-message-for-match1-" + testutil.RandomUUID()[:8],
	}
	body, _ := json.Marshal(testMsg)
	if err := ws1.WriteMessage(websocket.TextMessage, body); err != nil {
		t.Fatalf("failed to send on ws1: %v", err)
	}

	// Set a short deadline to check if ws2 receives anything
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err := ws2.ReadMessage()

	// ws2 should NOT receive match1's message
	if err == nil {
		t.Fatal("FAKE: match2 received match1's message — WebSocket isolation broken")
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "close") {
		t.Fatalf("unexpected error on ws2: %v", err)
	}
}

func TestWebSocket_TimestampNotHardcoded(t *testing.T) {
	ws := wsDial(t, "/ws")
	defer ws.Close()

	// Read connected message
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, _ := ws.ReadMessage()

	var frame map[string]interface{}
	json.Unmarshal(msg, &frame)

	// Check timestamp is not the hardcoded 2026-04-23 value
	if ts, ok := frame["timestamp"].(string); ok {
		testutil.AssertTimestampNotHardcoded(t, "ws_timestamp", ts)
	}
}
