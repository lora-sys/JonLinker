package recordinterviewnote

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lora-sys/JonLinker/internal/session"
)

type mockStore struct {
	data map[string][]byte
}

func (m *mockStore) Set(_ context.Context, key string, data []byte) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = data
	return nil
}

func TestRecordInterviewNote_Success(t *testing.T) {
	store := &mockStore{}
	tool := NewTool(store)

	args := recordInterviewNoteArgs{
		JobURL:            "https://example.com/job/123",
		QuestionsAsked:    []string{"介绍一下自己", "为什么选择我们公司"},
		CandidateAnswers:  "候选人回答了自己5年Go开发经验...",
		OverallAssessment: "技术能力扎实，沟通清晰",
	}

	argsJSON, _ := json.Marshal(args)
	ctx := session.WithSessionID(context.Background(), "test-session-1")
	result, err := tool.InvokableRun(ctx, string(argsJSON))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "面试记录已保存" {
		t.Fatalf("expected 面试记录已保存, got %s", result)
	}

	expectedKey := "test-session-1:interview_https://example.com/job/123"
	data, ok := store.data[expectedKey]
	if !ok {
		t.Fatalf("key %s not found in store", expectedKey)
	}

	var note InterviewNote
	if err := json.Unmarshal(data, &note); err != nil {
		t.Fatalf("unmarshal note: %v", err)
	}

	if note.JobURL != args.JobURL {
		t.Errorf("expected JobURL %s, got %s", args.JobURL, note.JobURL)
	}
	if len(note.QuestionsAsked) != 2 || note.QuestionsAsked[0] != "介绍一下自己" {
		t.Errorf("unexpected questions: %v", note.QuestionsAsked)
	}
	if note.CandidateAnswers != args.CandidateAnswers {
		t.Errorf("unexpected candidate answers")
	}
	if note.OverallAssessment != args.OverallAssessment {
		t.Errorf("unexpected overall assessment")
	}
	if _, err := time.Parse(time.RFC3339, note.CreatedAt); err != nil {
		t.Errorf("invalid created_at timestamp: %v", err)
	}
}

func TestRecordInterviewNote_MissingFields(t *testing.T) {
	store := &mockStore{}
	tool := NewTool(store)

	tests := []struct {
		name string
		args recordInterviewNoteArgs
	}{
		{"empty job_url", recordInterviewNoteArgs{QuestionsAsked: []string{"q"}, CandidateAnswers: "a", OverallAssessment: "o"}},
		{"empty questions_asked", recordInterviewNoteArgs{JobURL: "https://example.com/job/1", CandidateAnswers: "a", OverallAssessment: "o"}},
		{"empty candidate_answers", recordInterviewNoteArgs{JobURL: "https://example.com/job/1", QuestionsAsked: []string{"q"}, OverallAssessment: "o"}},
		{"empty overall_assessment", recordInterviewNoteArgs{JobURL: "https://example.com/job/1", QuestionsAsked: []string{"q"}, CandidateAnswers: "a"}},
		{"all empty", recordInterviewNoteArgs{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, _ := json.Marshal(tt.args)
			ctx := session.WithSessionID(context.Background(), "test-session")
			_, err := tool.InvokableRun(ctx, string(argsJSON))
			if err == nil {
				t.Fatal("expected error for missing fields, got nil")
			}
			if !strings.Contains(err.Error(), "missing required fields") {
				t.Errorf("expected missing required fields error, got: %v", err)
			}
		})
	}
}

func TestRecordInterviewNote_EmptySessionID(t *testing.T) {
	store := &mockStore{}
	tool := NewTool(store)

	args := recordInterviewNoteArgs{
		JobURL:            "https://example.com/job/1",
		QuestionsAsked:    []string{"q"},
		CandidateAnswers:  "a",
		OverallAssessment: "o",
	}

	argsJSON, _ := json.Marshal(args)
	ctx := context.Background()
	result, err := tool.InvokableRun(ctx, string(argsJSON))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "session not found" {
		t.Fatalf("expected 'session not found', got %s", result)
	}
}
