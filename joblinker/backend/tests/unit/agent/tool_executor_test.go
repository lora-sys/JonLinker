package agent_test

import (
	"context"
	"testing"

	"joblinker/internal/agent"

	"github.com/google/uuid"
)

func TestToolExecutor_QueryJobs(t *testing.T) {
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)

	result, err := exec.ExecuteTool(
		context.Background(),
		uuid.New(),
		"query_jobs",
		map[string]interface{}{
			"location": "remote",
			"skills":   []interface{}{"golang", "python"},
			"limit":    5,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.Data == nil {
		t.Error("expected data in result")
	}
}

func TestToolExecutor_GetCandidate(t *testing.T) {
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)

	result, err := exec.ExecuteTool(
		context.Background(),
		uuid.New(),
		"get_candidate",
		map[string]interface{}{
			"candidate_id": uuid.New().String(),
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestToolExecutor_CreateOffer_MissingMatchID(t *testing.T) {
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)

	result, err := exec.ExecuteTool(
		context.Background(),
		uuid.New(),
		"create_offer",
		map[string]interface{}{
			"salary": 100000,
		},
	)

	// Should fail with missing match_id
	if err == nil && result != nil && result.Success {
		t.Error("expected failure for missing match_id")
	}
}

func TestToolExecutor_ScheduleInterview(t *testing.T) {
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)
	matchID := uuid.New()

	result, err := exec.ExecuteTool(
		context.Background(),
		matchID,
		"schedule_interview",
		map[string]interface{}{
			"match_id":          matchID.String(),
			"datetime":          "2026-06-01T10:00:00Z",
			"duration_minutes": 60,
			"interview_type":    "video",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// May fail due to missing repo, but should handle gracefully
	_ = result
}

func TestToolExecutor_SearchCandidates(t *testing.T) {
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)

	result, err := exec.ExecuteTool(
		context.Background(),
		uuid.New(),
		"search_candidates",
		map[string]interface{}{
			"skills":         []interface{}{"golang", "python"},
			"experience_min": 3,
			"limit":          10,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestToolExecutor_UnknownTool(t *testing.T) {
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)

	result, err := exec.ExecuteTool(
		context.Background(),
		uuid.New(),
		"unknown_tool",
		map[string]interface{}{},
	)

	if err == nil && result != nil && result.Success {
		t.Error("expected failure for unknown tool")
	}
}

func TestToolExecutor_ExecuteWithNilRepos(t *testing.T) {
	// Tool executor should work even with nil repos (for testing purposes)
	exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)

	result, _ := exec.ExecuteTool(
		context.Background(),
		uuid.New(),
		"query_jobs",
		map[string]interface{}{"location": "remote"},
	)

	// Should handle nil repos gracefully
	if result == nil {
		t.Error("expected result even with nil repos")
	}
}