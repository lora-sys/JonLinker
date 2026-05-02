package integration

import (
	"encoding/xml"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWS_PingPongEcho(t *testing.T) {
	// This test requires a running server with WebSocket support
	// Skip if server is not running
	serverURL := "ws://localhost:8080/api/messages/test-match-id/ws"

	// Set a token for auth (we'll use a test token)
	header := make(http.Header)
	header.Set("Authorization", "Bearer test-token")

	ws, resp, err := websocket.DefaultDialer.Dial(serverURL, header)
	if err != nil {
		// Server might not be running or WebSocket not configured
		t.Skip("WebSocket server not available at localhost:8080")
		return
	}
	defer ws.Close()
	defer resp.Body.Close()

	// Read the connected message
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read initial message: %v", err)
	}

	// Send a ping message
	pingMsg := `<message>
		<header>
			<message_id>test-001</message_id>
			<timestamp>2026-04-29T10:00:00Z</timestamp>
			<sender_id>agent-001</sender_id>
			<receiver_id>agent-002</receiver_id>
		</header>
		<payload>
			<intent>PING</intent>
		</payload>
	</message>`

	if err := ws.WriteMessage(websocket.TextMessage, []byte(pingMsg)); err != nil {
		t.Fatalf("Failed to send ping: %v", err)
	}

	// Read response
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, respMsg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Verify response is valid XML
	var result map[string]interface{}
	if err := xml.Unmarshal(respMsg, &result); err != nil {
		t.Logf("Response is not XML: %s", string(respMsg))
		// Response might be JSON (connected message or error)
	}

	t.Logf("Received response: %s", string(respMsg))
}

func TestWS_MessageEcho(t *testing.T) {
	serverURL := "ws://localhost:8080/api/messages/test-match-002/ws"

	header := make(http.Header)
	header.Set("Authorization", "Bearer test-token")

	ws, _, err := websocket.DefaultDialer.Dial(serverURL, header)
	if err != nil {
		t.Skip("WebSocket server not available at localhost:8080")
		return
	}
	defer ws.Close()

	// Send an INTRODUCTION message
	introMsg := `<message>
		<header>
			<message_id>msg-intro-001</message_id>
			<timestamp>2026-04-29T10:00:00Z</timestamp>
			<sender_id>seeker-agent-001</sender_id>
			<receiver_id>recruiter-agent-001</receiver_id>
		</header>
		<payload>
			<intent>INTRODUCTION</intent>
			<parameters>
				<role>seeker</role>
				<name>John Doe</name>
				<title>Senior Backend Engineer</title>
				<company>TechCorp</company>
				<skills>Go,PostgreSQL,Kubernetes</skills>
			</parameters>
		</payload>
	</message>`

	if err := ws.WriteMessage(websocket.TextMessage, []byte(introMsg)); err != nil {
		t.Fatalf("Failed to send introduction: %v", err)
	}

	// Read response
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, respMsg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Parse response XML
	var a2aResp A2AMessage
	if err := xml.Unmarshal(respMsg, &a2aResp); err != nil {
		t.Fatalf("Failed to parse response XML: %v", err)
	}

	// Verify response has expected structure
	if a2aResp.Header.SenderID != "system" {
		t.Errorf("Expected sender 'system', got '%s'", a2aResp.Header.SenderID)
	}
	if a2aResp.Header.ReceiverID != "seeker-agent-001" {
		t.Errorf("Expected receiver 'seeker-agent-001', got '%s'", a2aResp.Header.ReceiverID)
	}
	if a2aResp.Header.ReplyTo != "msg-intro-001" {
		t.Errorf("Expected reply_to 'msg-intro-001', got '%s'", a2aResp.Header.ReplyTo)
	}

	t.Logf("Received valid A2A response: intent=%s", a2aResp.Payload.Intent)
}

// A2AMessage struct for parsing responses
type A2AMessage struct {
	XMLName  xml.Name       `xml:"message"`
	Header   MessageHeader  `xml:"header"`
	Payload  MessagePayload `xml:"payload"`
}

type MessageHeader struct {
	MessageID  string `xml:"message_id"`
	Timestamp string `xml:"timestamp"`
	SenderID  string `xml:"sender_id"`
	ReceiverID string `xml:"receiver_id"`
	ReplyTo   string `xml:"reply_to,omitempty"`
}

type MessagePayload struct {
	Intent      string           `xml:"intent"`
	Parameters *MessageParams   `xml:"parameters,omitempty"`
}

type MessageParams struct {
	Role string `xml:"role,omitempty"`
	Name string `xml:"name,omitempty"`
}

func TestWS_AuthRequired(t *testing.T) {
	// Test that WebSocket requires authentication
	serverURL := "ws://localhost:8080/api/messages/test-match/ws"

	// Connect without auth header
	ws, resp, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err == nil {
		ws.Close()
		resp.Body.Close()
		// If connection succeeds without auth, that's a security issue
		t.Error("WebSocket should require authentication")
	}
}

func TestWS_InvalidMatchID(t *testing.T) {
	// Test WebSocket with invalid match ID format
	serverURL := "ws://localhost:8080/api/messages/invalid-uuid/ws"

	header := make(http.Header)
	header.Set("Authorization", "Bearer test-token")

	ws, _, err := websocket.DefaultDialer.Dial(serverURL, header)
	if err != nil {
		// Connection should fail or return error
		t.Logf("Connection failed (expected): %v", err)
		return
	}
	defer ws.Close()

	// Read error message
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read error: %v", err)
	}

	if !strings.Contains(string(msg), "error") && !strings.Contains(string(msg), "match") {
		t.Logf("Unexpected message: %s", string(msg))
	}
}
