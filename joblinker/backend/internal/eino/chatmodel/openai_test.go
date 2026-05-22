package chatmodel

import (
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestParseFunctionCallXML_simple(t *testing.T) {
	content := `<function_call><function_name>get_match_progress</function_name><parameters><match_id>abc-123</match_id></parameters></function_call>`
	calls := parseFunctionCallXML(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	if calls[0].Function.Name != "get_match_progress" {
		t.Errorf("expected name get_match_progress, got %s", calls[0].Function.Name)
	}
	if calls[0].ID == "" {
		t.Error("expected non-empty ID")
	}
	if calls[0].Type != "function" {
		t.Errorf("expected type function, got %s", calls[0].Type)
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(calls[0].Function.Arguments), &args); err != nil {
		t.Fatalf("failed to parse args JSON: %v", err)
	}
	if args["match_id"] != "abc-123" {
		t.Errorf("expected match_id abc-123, got %v", args["match_id"])
	}
}

func TestParseFunctionCallXML_withText(t *testing.T) {
	content := `Let me check the match progress for you.

<function_call><function_name>get_match_progress</function_name><parameters><match_id>abc-123</match_id></parameters></function_call>`
	calls := parseFunctionCallXML(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	clean := stripFunctionCallXML(content)
	if clean != "Let me check the match progress for you." {
		t.Errorf("unexpected cleaned content: %q", clean)
	}
}

func TestParseFunctionCallXML_multipleCalls(t *testing.T) {
	content := `<function_call><function_name>get_match_progress</function_name><parameters><match_id>abc</match_id></parameters></function_call>
<function_call><function_name>get_interview_details</function_name><parameters><match_id>def</match_id></parameters></function_call>`
	calls := parseFunctionCallXML(content)
	if len(calls) != 2 {
		t.Fatalf("expected 2 tool calls, got %d", len(calls))
	}
	if calls[0].Function.Name != "get_match_progress" {
		t.Errorf("expected first call get_match_progress, got %s", calls[0].Function.Name)
	}
	if calls[1].Function.Name != "get_interview_details" {
		t.Errorf("expected second call get_interview_details, got %s", calls[1].Function.Name)
	}
}

func TestParseFunctionCallXML_noMatch(t *testing.T) {
	content := `Hello, how are you?`
	calls := parseFunctionCallXML(content)
	if len(calls) != 0 {
		t.Errorf("expected 0 tool calls, got %d", len(calls))
	}
}

func TestParseFunctionCallXML_numericParam(t *testing.T) {
	content := `<function_call><function_name>create_offer</function_name><parameters><match_id>abc</match_id><salary>150000</salary><start_date>2026-07-01</start_date></parameters></function_call>`
	calls := parseFunctionCallXML(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	var args map[string]interface{}
	json.Unmarshal([]byte(calls[0].Function.Arguments), &args)
	if args["salary"] != float64(150000) {
		t.Errorf("expected salary 150000, got %v (type %T)", args["salary"], args["salary"])
	}
	if args["match_id"] != "abc" {
		t.Errorf("expected match_id abc, got %v", args["match_id"])
	}
}

func TestXMLParamsToJSON(t *testing.T) {
	xml := "<match_id>abc</match_id><salary>100000</salary><remote>true</remote>"
	jsonBytes := xmlParamsToJSON(xml)
	var args map[string]interface{}
	json.Unmarshal(jsonBytes, &args)
	if args["match_id"] != "abc" {
		t.Errorf("expected match_id abc, got %v", args["match_id"])
	}
	if args["salary"] != float64(100000) {
		t.Errorf("expected salary 100000, got %v", args["salary"])
	}
}

func TestJSONArgsToXML(t *testing.T) {
	args := `{"match_id":"abc","salary":100000}`
	xml := jsonArgsToXML(args)
	if xml == "" {
		t.Fatal("expected non-empty XML")
	}
	if xml != "<match_id>abc</match_id><salary>100000</salary>" && xml != "<salary>100000</salary><match_id>abc</match_id>" {
		t.Logf("got XML: %q (order may vary, checking keys)", xml)
		if !contains(xml, "match_id") || !contains(xml, "salary") {
			t.Errorf("missing expected keys in XML: %s", xml)
		}
	}
}

func TestConvertToAIMessages_toolRole(t *testing.T) {
	input := []*schema.Message{
		{
			Role:     schema.Tool,
			Content:  `{"progress": "in_progress"}`,
			ToolName: "get_match_progress",
		},
	}
	result := convertToAIMessages(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0].Role != "user" {
		t.Errorf("expected role 'user', got %s", result[0].Role)
	}
	if !contains(result[0].Content, "get_match_progress") {
		t.Errorf("expected tool_name in content, got: %s", result[0].Content)
	}
	if !contains(result[0].Content, "in_progress") {
		t.Errorf("expected result in content, got: %s", result[0].Content)
	}
}

func TestConvertToAIMessages_assistantWithToolCalls(t *testing.T) {
	input := []*schema.Message{
		{
			Role:    schema.Assistant,
			Content: "Let me check that for you.",
			ToolCalls: []schema.ToolCall{
				{
					ID:   "call-1",
					Type: "function",
					Function: schema.FunctionCall{
						Name:      "get_match_progress",
						Arguments: `{"match_id":"abc"}`,
					},
				},
			},
		},
	}
	result := convertToAIMessages(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0].Role != "assistant" {
		t.Errorf("expected role 'assistant', got %s", result[0].Role)
	}
	if !contains(result[0].Content, "<function_call>") {
		t.Errorf("expected <function_call> XML in content, got: %s", result[0].Content)
	}
	if !contains(result[0].Content, "get_match_progress") {
		t.Errorf("expected function name in content, got: %s", result[0].Content)
	}
	if !contains(result[0].Content, "Let me check") {
		t.Errorf("expected original text in content, got: %s", result[0].Content)
	}
}

func TestRoundTrip_XMLThroughGenerate(t *testing.T) {
	modelOutput := `Let me look that up.

<function_call><function_name>get_match_progress</function_name><parameters><match_id>abc-123</match_id></parameters></function_call>`

	// Simulate Generate(): parse XML from output
	calls := parseFunctionCallXML(modelOutput)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	cleanContent := stripFunctionCallXML(modelOutput)

	// Simulate storing the result as an Eino message (what Generate returns)
	msg := &schema.Message{
		Role:      schema.Assistant,
		Content:   cleanContent,
		ToolCalls: calls,
	}

	// Simulate convertToAIMessages on the NEXT iteration (feeding history back)
	aiMsgs := convertToAIMessages([]*schema.Message{msg})
	if len(aiMsgs) != 1 {
		t.Fatalf("expected 1 AI message, got %d", len(aiMsgs))
	}

	// Verify the reconstructed XML still parses correctly
	reparsed := parseFunctionCallXML(aiMsgs[0].Content)
	if len(reparsed) != 1 {
		t.Errorf("expected 1 tool call in round-tripped content, got %d; content=%q", len(reparsed), aiMsgs[0].Content)
	}
	if reparsed[0].Function.Name != "get_match_progress" {
		t.Errorf("expected name get_match_progress, got %s", reparsed[0].Function.Name)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
