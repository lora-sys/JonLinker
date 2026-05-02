package agent_test

import (
	"testing"

	"joblinker/internal/agent"
)

func TestFunctionDefinitions_GetAllTools(t *testing.T) {
	defs := agent.GetFunctionDefinitions()

	if len(defs) != 5 {
		t.Errorf("expected 5 function definitions, got %d", len(defs))
	}

	// Verify all expected tools are present
	toolNames := make(map[string]bool)
	for _, def := range defs {
		toolNames[def.Name] = true
	}

	expectedTools := []string{
		"query_jobs",
		"get_candidate",
		"create_offer",
		"schedule_interview",
		"search_candidates",
	}

	for _, tool := range expectedTools {
		if !toolNames[tool] {
			t.Errorf("expected tool %s not found", tool)
		}
	}
}

func TestFunctionDefinitions_QueryJobsSchema(t *testing.T) {
	def := agent.GetToolByName("query_jobs")
	if def == nil {
		t.Fatal("query_jobs tool not found")
	}

	// Verify required parameters
	params := def.Parameters
	if params["type"] != "object" {
		t.Error("expected object type for parameters")
	}

	properties := params["properties"].(map[string]interface{})

	// Check location is required
	location := properties["location"].(map[string]interface{})
	if location["type"] != "string" {
		t.Error("expected location to be string type")
	}

	// Check skills is array
	skills := properties["skills"].(map[string]interface{})
	if skills["type"] != "array" {
		t.Error("expected skills to be array type")
	}
}

func TestFunctionDefinitions_GetCandidateSchema(t *testing.T) {
	def := agent.GetToolByName("get_candidate")
	if def == nil {
		t.Fatal("get_candidate tool not found")
	}

	params := def.Parameters
	properties := params["properties"].(map[string]interface{})

	// Verify candidate_id is required
	candidateID := properties["candidate_id"].(map[string]interface{})
	if candidateID["type"] != "string" {
		t.Error("expected candidate_id to be string type")
	}
}

func TestFunctionDefinitions_CreateOfferSchema(t *testing.T) {
	def := agent.GetToolByName("create_offer")
	if def == nil {
		t.Fatal("create_offer tool not found")
	}

	params := def.Parameters
	properties := params["properties"].(map[string]interface{})

	// Verify required fields
	matchID := properties["match_id"].(map[string]interface{})
	if matchID["type"] != "string" {
		t.Error("expected match_id to be string type")
	}

	salary := properties["salary"].(map[string]interface{})
	if salary["type"] != "integer" {
		t.Error("expected salary to be integer type")
	}
}

func TestFunctionDefinitions_ScheduleInterviewSchema(t *testing.T) {
	def := agent.GetToolByName("schedule_interview")
	if def == nil {
		t.Fatal("schedule_interview tool not found")
	}

	params := def.Parameters
	properties := params["properties"].(map[string]interface{})

	// Verify match_id and datetime are required
	matchID := properties["match_id"].(map[string]interface{})
	if matchID["type"] != "string" {
		t.Error("expected match_id to be string type")
	}

	datetime := properties["datetime"].(map[string]interface{})
	if datetime["type"] != "string" {
		t.Error("expected datetime to be string type")
	}

	// Verify interview_type enum
	interviewType := properties["interview_type"].(map[string]interface{})
	enum := interviewType["enum"].([]string)
	if len(enum) != 3 {
		t.Errorf("expected 3 interview type options, got %d", len(enum))
	}
}

func TestFunctionDefinitions_SearchCandidatesSchema(t *testing.T) {
	def := agent.GetToolByName("search_candidates")
	if def == nil {
		t.Fatal("search_candidates tool not found")
	}

	params := def.Parameters
	properties := params["properties"].(map[string]interface{})

	// Verify skills is required
	skills := properties["skills"].(map[string]interface{})
	if skills["type"] != "array" {
		t.Error("expected skills to be array type")
	}
}

func TestFunctionDefinitions_GetToolByName(t *testing.T) {
	def := agent.GetToolByName("query_jobs")
	if def == nil {
		t.Error("expected to find query_jobs tool")
	}

	unknown := agent.GetToolByName("nonexistent_tool")
	if unknown != nil {
		t.Error("expected nil for nonexistent tool")
	}
}

func TestFunctionDefinitions_GetToolNames(t *testing.T) {
	names := agent.GetToolNames()

	if len(names) != 5 {
		t.Errorf("expected 5 tool names, got %d", len(names))
	}

	// Verify all names are strings
	for _, name := range names {
		if name == "" {
			t.Error("expected non-empty tool name")
		}
	}
}