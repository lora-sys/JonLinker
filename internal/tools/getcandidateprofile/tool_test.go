package getcandidateprofile

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/resume"
	"github.com/lora-sys/JonLinker/internal/session"
)

type mockStore struct {
	data map[string][]byte
}

func (m *mockStore) Get(_ context.Context, key string) ([]byte, bool, error) {
	v, ok := m.data[key]
	if !ok {
		return nil, false, nil
	}
	return v, true, nil
}

func TestInvokableRun_Success(t *testing.T) {
	profile := resume.CandidateProfile{
		Name:  "张三",
		Title: "Go开发工程师",
		Skills: []string{"Go", "Kubernetes", "Docker"},
	}
	application := job.Application{
		JobTitle:    "Senior Go Developer",
		Company:     "TechCo",
		CoverLetter: "I am a great fit...",
		ResumeMD:    "# Resume\n...",
	}

	profileData, _ := json.Marshal(profile)
	appData, _ := json.Marshal(application)

	store := &mockStore{
		data: map[string][]byte{
			"test-sid:profile":                        profileData,
			"test-sid:application_https://example.com/job/1": appData,
		},
	}

	tool := NewTool(store)
	ctx := session.WithSessionID(context.Background(), "test-sid")

	result, err := tool.InvokableRun(ctx, `{"job_url": "https://example.com/job/1"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal([]byte(result), &got); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if _, ok := got["profile"]; !ok {
		t.Error("missing profile in result")
	}
	if _, ok := got["application"]; !ok {
		t.Error("missing application in result")
	}
}

func TestInvokableRun_ProfileNotFound(t *testing.T) {
	appData, _ := json.Marshal(job.Application{
		JobTitle: "Senior Go Developer",
		Company:  "TechCo",
	})

	store := &mockStore{
		data: map[string][]byte{
			"test-sid:application_https://example.com/job/1": appData,
		},
	}

	tool := NewTool(store)
	ctx := session.WithSessionID(context.Background(), "test-sid")

	result, err := tool.InvokableRun(ctx, `{"job_url": "https://example.com/job/1"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "profile not found for session test-sid"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestInvokableRun_ApplicationNotFound(t *testing.T) {
	profileData, _ := json.Marshal(resume.CandidateProfile{
		Name:  "张三",
		Title: "Go开发工程师",
	})

	store := &mockStore{
		data: map[string][]byte{
			"test-sid:profile": profileData,
		},
	}

	tool := NewTool(store)
	ctx := session.WithSessionID(context.Background(), "test-sid")

	result, err := tool.InvokableRun(ctx, `{"job_url": "https://example.com/job/1"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "application not found for job https://example.com/job/1"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestInvokableRun_EmptySessionID(t *testing.T) {
	store := &mockStore{data: make(map[string][]byte)}
	tool := NewTool(store)
	ctx := context.Background()

	result, err := tool.InvokableRun(ctx, `{"job_url": "https://example.com/job/1"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "session not found" {
		t.Errorf("expected %q, got %q", "session not found", result)
	}
}
