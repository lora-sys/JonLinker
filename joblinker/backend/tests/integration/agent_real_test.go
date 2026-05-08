//go:build integration
// +build integration

// Real integration tests for agent CRUD endpoints.
// Uses real HTTP requests. Verifies data isolation between users.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"joblinker/tests/testutil"
)

func TestAgentCreate_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "seeker")

	// Create agent
	payload := testutil.AgentPayload{
		Name:      "Seeker Agent " + testutil.RandomUUID()[:8],
		AgentType: "seeker",
		Skills:    []string{"go", "python", "postgresql"},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/api/agents", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to POST /api/agents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var agent map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify fields
	if agent["id"] == nil || agent["id"] == "" {
		t.Error("expected non-empty agent id")
	}
	testutil.AssertNotHardcoded(t, "agent_name", agent["name"].(string))
	if agent["agent_type"] != "seeker" {
		t.Errorf("expected agent_type 'seeker', got %v", agent["agent_type"])
	}
}

func TestAgentList_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "seeker")

	// Create an agent first
	agentID := testutil.MustCreateAgent(t, baseURL, token, "seeker")

	// List agents
	req, _ := http.NewRequest("GET", baseURL+"/api/agents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to GET /api/agents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var agents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		t.Fatalf("failed to decode agents list: %v", err)
	}

	found := false
	for _, a := range agents {
		if a["id"] == agentID {
			found = true
			testutil.AssertNotHardcoded(t, "agent_name", a["name"].(string))
			break
		}
	}
	if !found {
		t.Error("created agent not found in list")
	}
}

func TestAgentGet_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "recruiter")

	agentID := testutil.MustCreateAgent(t, baseURL, token, "recruiter")

	req, _ := http.NewRequest("GET", baseURL+"/api/agents/"+agentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to GET /api/agents/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var agent map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&agent)

	if agent["id"] != agentID {
		t.Errorf("expected agent id %s, got %v", agentID, agent["id"])
	}
	testutil.AssertNotHardcoded(t, "agent_name", agent["name"].(string))
}

func TestAgentUpdate_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "seeker")

	agentID := testutil.MustCreateAgent(t, baseURL, token, "seeker")

	newName := "Updated Agent " + testutil.RandomUUID()[:8]
	payload := map[string]interface{}{
		"name":   newName,
		"skills": []string{"golang", "kubernetes", "aws"},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", baseURL+"/api/agents/"+agentID, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to PUT /api/agents/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var updated map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updated)

	if updated["name"] != newName {
		t.Errorf("expected name %q, got %v", newName, updated["name"])
	}
}

func TestAgentDelete_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "seeker")

	agentID := testutil.MustCreateAgent(t, baseURL, token, "seeker")

	req, _ := http.NewRequest("DELETE", baseURL+"/api/agents/"+agentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to DELETE /api/agents/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		t.Fatalf("expected 200/204, got %d", resp.StatusCode)
	}

	// Verify agent is gone
	req2, _ := http.NewRequest("GET", baseURL+"/api/agents/"+agentID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, _ := testutil.HTTPClient.Do(req2)
	defer resp2.Body.Close()

	if resp2.StatusCode != 404 {
		t.Errorf("expected 404 after delete, got %d", resp2.StatusCode)
	}
}

func TestAgentDataIsolation(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	// User A registers and creates an agent
	tokenA, _ := testutil.MustRegister(t, baseURL, "seeker")
	agentIDA := testutil.MustCreateAgent(t, baseURL, tokenA, "seeker")

	// User B registers
	tokenB, _ := testutil.MustRegister(t, baseURL, "seeker")

	// User B lists agents — should NOT see User A's agent
	req, _ := http.NewRequest("GET", baseURL+"/api/agents", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	var agents []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&agents)

	for _, a := range agents {
		if a["id"] == agentIDA {
			t.Fatal("FAKE: user B can see user A's agent — data isolation broken")
		}
	}
}

func TestAgentCreate_RecruiterType(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "recruiter")

	agentID := testutil.MustCreateAgent(t, baseURL, token, "recruiter")

	req, _ := http.NewRequest("GET", baseURL+"/api/agents/"+agentID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	var agent map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&agent)

	if agent["agent_type"] != "recruiter" {
		t.Errorf("expected agent_type 'recruiter', got %v", agent["agent_type"])
	}
}
