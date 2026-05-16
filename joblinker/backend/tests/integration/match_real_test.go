//go:build integration
// +build integration

// Real integration tests for match auto-matching endpoints.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"joblinker/tests/testutil"
)

func TestAutoMatch_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	seekerToken, _ := testutil.MustRegister(t, baseURL, "seeker")
	recruiterToken, _ := testutil.MustRegister(t, baseURL, "recruiter")

	seekerAgentID := testutil.MustCreateAgent(t, baseURL, seekerToken, "seeker")
	recruiterAgentID := testutil.MustCreateAgent(t, baseURL, recruiterToken, "recruiter")
	jobID := testutil.MustCreateJob(t, baseURL, recruiterToken)

	// Seeker calls auto-match with the job
	payload := map[string]interface{}{
		"job_ids": []string{jobID},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/api/matches/auto", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seekerToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to POST /api/matches/auto: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode auto-match response: %v", err)
	}

	matches, ok := result["matches"].([]interface{})
	if !ok || len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	match := matches[0].(map[string]interface{})

	if match["id"] == nil || match["id"] == "" {
		t.Error("expected non-empty match id")
	}
	// recruiter_agent_id is now populated by the backend
	if match["recruiter_agent_id"] == nil || match["recruiter_agent_id"] == "" {
		t.Error("expected non-empty recruiter_agent_id")
	}

	score, ok := match["score"].(float64)
	if !ok {
		t.Error("expected numeric score")
	}
	if score <= 0 || score > 1 {
		t.Errorf("expected score between 0 and 1, got %f", score)
	}

	testutil.AssertScoreNotHardcoded(t, "match_score", score)
}

func TestAutoMatch_List(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	seekerToken, _ := testutil.MustRegister(t, baseURL, "seeker")
	recruiterToken, _ := testutil.MustRegister(t, baseURL, "recruiter")

	seekerAgentID := testutil.MustCreateAgent(t, baseURL, seekerToken, "seeker")
	recruiterAgentID := testutil.MustCreateAgent(t, baseURL, recruiterToken, "recruiter")
	jobID := testutil.MustCreateJob(t, baseURL, recruiterToken)

	matchID := testutil.MustCreateMatch(t, baseURL, seekerToken, recruiterToken, seekerAgentID, recruiterAgentID, jobID)

	req, _ := http.NewRequest("GET", baseURL+"/api/matches", nil)
	req.Header.Set("Authorization", "Bearer "+recruiterToken)
	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var matches []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&matches)

	found := false
	for _, m := range matches {
		if m["id"] == matchID {
			found = true
			score := m["score"].(float64)
			testutil.AssertScoreNotHardcoded(t, "match_score", score)
			break
		}
	}
	if !found {
		t.Error("created match not found in list")
	}
}

func TestAutoMatch_Get(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	seekerToken, _ := testutil.MustRegister(t, baseURL, "seeker")
	recruiterToken, _ := testutil.MustRegister(t, baseURL, "recruiter")

	seekerAgentID := testutil.MustCreateAgent(t, baseURL, seekerToken, "seeker")
	recruiterAgentID := testutil.MustCreateAgent(t, baseURL, recruiterToken, "recruiter")
	jobID := testutil.MustCreateJob(t, baseURL, recruiterToken)

	matchID := testutil.MustCreateMatch(t, baseURL, seekerToken, recruiterToken, seekerAgentID, recruiterAgentID, jobID)

	req, _ := http.NewRequest("GET", baseURL+"/api/matches/"+matchID, nil)
	req.Header.Set("Authorization", "Bearer "+recruiterToken)
	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var match map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&match)

	if match["id"] != matchID {
		t.Errorf("expected match id %s, got %v", matchID, match["id"])
	}
}

func TestAutoMatch_Confirm(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	seekerToken, _ := testutil.MustRegister(t, baseURL, "seeker")
	recruiterToken, _ := testutil.MustRegister(t, baseURL, "recruiter")

	seekerAgentID := testutil.MustCreateAgent(t, baseURL, seekerToken, "seeker")
	recruiterAgentID := testutil.MustCreateAgent(t, baseURL, recruiterToken, "recruiter")
	jobID := testutil.MustCreateJob(t, baseURL, recruiterToken)

	matchID := testutil.MustCreateMatch(t, baseURL, seekerToken, recruiterToken, seekerAgentID, recruiterAgentID, jobID)

	req, _ := http.NewRequest("POST", baseURL+"/api/matches/"+matchID+"/confirm", nil)
	req.Header.Set("Authorization", "Bearer "+recruiterToken)
	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 && resp.StatusCode != 201 {
		t.Fatalf("expected 200/204/201, got %d", resp.StatusCode)
	}
}

func TestMatch_SeekerCannotCreateJob(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	seekerToken, _ := testutil.MustRegister(t, baseURL, "seeker")

	payload := map[string]interface{}{
		"title":       "Seeker Trying to Post Job",
		"description": "This should be rejected",
		"skills":      []string{"go"},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/api/jobs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seekerToken)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("expected 403 for seeker creating job, got %d", resp.StatusCode)
	}
}
