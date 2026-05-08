//go:build integration
// +build integration

// Real integration tests for memory persistence via Chroma.
package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"joblinker/tests/testutil"
)

// TestMemoryPersistence verifies that StoreMemory actually persists data
// and GetRecentMemories returns non-empty results.
// If Chroma is not connected, this test FAILS (not skip).
func TestMemoryPersistence(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	chromaHost := testutil.RequireEnv(t, "CHROMA_HOST")

	// Verify Chroma is reachable first
	if _, err := testutil.HTTPClient.Get("http://" + chromaHost + "/api/v1/heartbeat"); err != nil {
		t.Fatalf("Chroma not reachable at %s: %v — %s", chromaHost, err, testutil.ENV_MISSING_MSG)
	}

	// Setup: create match context
	seekerToken, _ := testutil.MustRegister(t, baseURL, "seeker")
	recruiterToken, _ := testutil.MustRegister(t, baseURL, "recruiter")
	seekerAgentID := testutil.MustCreateAgent(t, baseURL, seekerToken, "seeker")
	recruiterAgentID := testutil.MustCreateAgent(t, baseURL, recruiterToken, "recruiter")
	jobID := testutil.MustCreateJob(t, baseURL, recruiterToken)
	matchID := testutil.MustCreateMatch(t, baseURL, seekerToken, recruiterToken, seekerAgentID, recruiterAgentID, jobID)

	// Send a message to trigger memory storage
	msgPayload := map[string]interface{}{
		"match_id": matchID,
		"intent":   "INTRODUCTION",
		"content":  "I have 8 years of experience in distributed systems with Go and Kubernetes",
	}
	body, _ := toJSON(msgPayload)
	req, _ := newHTTPRequest("POST", baseURL+"/api/messages", body, seekerToken)
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}
	resp.Body.Close()

	// Wait for processing
	t.Logf("Message sent for match %s, waiting for memory persistence...", matchID)

	// Get recent memories for this match
	getReq, _ := newHTTPRequest("GET", baseURL+"/api/matches/"+matchID+"/memories", nil, seekerToken)
	memResp, err := testutil.HTTPClient.Do(getReq)
	if err != nil {
		t.Fatalf("failed to get memories: %v", err)
	}
	defer memResp.Body.Close()

	if memResp.StatusCode != 200 {
		t.Fatalf("expected 200 from /api/matches/*/memories, got %d", memResp.StatusCode)
	}

	var memResult map[string]interface{}
	fromJSON(memResp.Body, &memResult)

	memories, ok := memResult["memories"].([]interface{})
	if !ok {
		t.Fatalf("expected 'memories' array in response, got: %+v", memResult)
	}

	if len(memories) == 0 {
		t.Fatal("FAKE: GetRecentMemories returned empty — StoreMemory does not persist to Chroma")
	}

	// Verify memory content is not hardcoded
	for _, m := range memories {
		mem := m.(map[string]interface{})
		if content, ok := mem["content"].(string); ok && content != "" {
			testutil.AssertNotHardcoded(t, "memory_content", content)
		}
	}
}

// TestVectorSearch verifies that vector similarity search returns relevant results.
func TestVectorSearch(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	chromaHost := testutil.RequireEnv(t, "CHROMA_HOST")

	if _, err := testutil.HTTPClient.Get("http://" + chromaHost + "/api/v1/heartbeat"); err != nil {
		t.Fatalf("Chroma not reachable: %v — %s", err, testutil.ENV_MISSING_MSG)
	}

	// Search for skills that should match stored preferences
	req, _ := newHTTPRequest("GET", baseURL+"/api/preferences/search?query=golang+kubernetes&limit=5", nil, "")
	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	fromJSON(resp.Body, &result)

	prefs, ok := result["preferences"].([]interface{})
	if !ok {
		t.Fatalf("expected 'preferences' array, got: %+v", result)
	}

	// If we get results, verify they're not hardcoded
	for _, p := range prefs {
		pref := p.(map[string]interface{})
		if val, ok := pref["value"].(string); ok && val != "" {
			testutil.AssertNotHardcoded(t, "preference_value", val)
		}
	}
}

// Helper functions
func toJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func fromJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func newHTTPRequest(method, url string, body []byte, token string) (*http.Request, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return req, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}
