package sessionstore

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"joblinker/internal/model"
)

// ---- S4: ReopenSession Tests ----

func TestReopenSession_ConcludedPaused(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	svc := NewSessionService(store, nil, nil)
	ctx := context.Background()

	matchID := uuid.New()
	sid, err := svc.CreateSession(ctx, matchID)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Add some messages so summary has content
	store.AppendMessages(ctx, sid, []Message{
		{Role: "seeker", Content: "I am looking for a software engineering role"},
		{Role: "assistant", Content: "Let me discuss the position with you."},
	})

	// Conclude with paused reason
	if err := store.Conclude(ctx, sid, ConcludedPaused); err != nil {
		t.Fatalf("Conclude failed: %v", err)
	}

	// Reopen
	newSid, err := svc.ReopenSession(ctx, matchID)
	if err != nil {
		t.Fatalf("ReopenSession failed: %v", err)
	}

	// Verify new session ID is different
	if newSid == sid {
		t.Error("New session ID should differ from old")
	}

	// Verify new session is version 2
	newSession, _ := store.Load(ctx, newSid)
	if newSession.Version != 2 {
		t.Errorf("New session version = %d, want 2", newSession.Version)
	}

	// Verify new session has system message injected (from summary)
	if len(newSession.Messages) == 0 {
		t.Fatal("New session should have system message with history summary")
	}
	if newSession.Messages[0].Role != "system" {
		t.Errorf("First message role = %s, want system", newSession.Messages[0].Role)
	}
}

func TestReopenSession_NoSessions(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	svc := NewSessionService(store, nil, nil)
	ctx := context.Background()

	_, err := svc.ReopenSession(ctx, uuid.New())
	if err == nil {
		t.Fatal("Expected error when no sessions exist")
	}
}

func TestReopenSession_NaturalEnd(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	svc := NewSessionService(store, nil, nil)
	ctx := context.Background()

	matchID := uuid.New()
	sid, _ := svc.CreateSession(ctx, matchID)
	store.Conclude(ctx, sid, ConcludedNaturally)

	_, err := svc.ReopenSession(ctx, matchID)
	if err == nil {
		t.Fatal("Expected error when concluded naturally")
	}
}

func TestReopenSession_ActiveSession(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	svc := NewSessionService(store, nil, nil)
	ctx := context.Background()

	matchID := uuid.New()
	svc.CreateSession(ctx, matchID)

	_, err := svc.ReopenSession(ctx, matchID)
	if err == nil {
		t.Fatal("Expected error when session is still active")
	}
}

func TestReopenSession_HistoryInjection(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	svc := NewSessionService(store, nil, nil)
	ctx := context.Background()

	matchID := uuid.New()
	sid, _ := svc.CreateSession(ctx, matchID)

	// Add some messages
	store.AppendMessages(ctx, sid, []Message{
		{Role: "seeker", Content: "I am looking for a software engineering role with competitive salary"},
		{Role: "assistant", Content: "Great, let me discuss the position with you. The role requires Go expertise."},
	})
	store.Conclude(ctx, sid, ConcludedPaused)

	newSid, err := svc.ReopenSession(ctx, matchID)
	if err != nil {
		t.Fatalf("ReopenSession failed: %v", err)
	}

	newSession, _ := store.Load(ctx, newSid)
	if len(newSession.Messages) == 0 {
		t.Fatal("Expected system message with history")
	}
	sysMsg := newSession.Messages[0]
	if sysMsg.Role != "system" {
		t.Errorf("First message role = %s, want system", sysMsg.Role)
	}
	if !contains(sysMsg.Content, "历史 Session 摘要") {
		t.Errorf("System message should contain summary header, got: %s", sysMsg.Content[:min(50, len(sysMsg.Content))])
	}
}

// ---- S4: UpdateMatchStatusForSession Tests ----

func TestUpdateMatchStatusForSession(t *testing.T) {
	svc := NewSessionService(nil, nil, nil)

	tests := []struct {
		current  string
		state    string
		want     string
	}{
		{"pending", "idle", "pending"},
		{"mutual_interest", "negotiating", "negotiating"},
		{"negotiating", "interviewing", "interviewing"},
		{"interviewing", "offered", "offered"},
		{"offered", "concluded_naturally", "hired"},
		{"pending", "concluded_naturally", "rejected"},
		{"negotiating", "concluded_paused", "paused"},
	}

	for _, tt := range tests {
		got := svc.UpdateMatchStatusForSession(S(tt.current), tt.state)
		if string(got) != tt.want {
			t.Errorf("UpdateMatchStatusForSession(%q, %q) = %q, want %q", tt.current, tt.state, got, tt.want)
		}
	}
}

// ---- S5: Token Estimation Tests ----

func TestAutoCompress_ByDirectAppend(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	svc := NewSessionService(store, nil, nil)
	ctx := context.Background()

	matchID := uuid.New()
	sid, _ := svc.CreateSession(ctx, matchID)

	compressed := false
	svc.SetCompressFunc(func(ctx context.Context, msgs []Message) (string, string, error) {
		compressed = true
		return "Agreed on salary range. Discussed role requirements.", `{"agreed":true}`, nil
	})

	// Short message should NOT trigger compression
	store.AppendMessages(ctx, sid, []Message{
		{Role: "seeker", Content: "Hi"},
	})
	svc.maybeCompress(ctx, sid)
	if compressed {
		t.Fatal("Compression triggered on short message - unexpected")
	}

	// Long message SHOULD trigger (create enough tokens to cross threshold)
	longContent := ""
	for i := 0; i < 3000; i++ {
		longContent += "This is a test sentence to push token count above the threshold for compression. "
	}
	if err := store.AppendMessages(ctx, sid, []Message{
		{Role: "user", Content: longContent},
	}); err != nil {
		t.Fatalf("Append long message failed: %v", err)
	}
	svc.maybeCompress(ctx, sid)

	if !compressed {
		t.Fatal("Compression should have been triggered on long message")
	}
}

func TestEstimateTokenCount_Empty(t *testing.T) {
	if n := EstimateTokenCount(nil); n != 0 {
		t.Errorf("Empty = %d, want 0", n)
	}
	if n := EstimateTokenCount([]Message{}); n != 0 {
		t.Errorf("Empty slice = %d, want 0", n)
	}
}

func TestEstimateTokenCount_Short(t *testing.T) {
	msgs := []Message{
		{Role: "user", Content: "Hello"},
	}
	n := EstimateTokenCount(msgs)
	if n < 1 || n > 5 {
		t.Errorf("Short message = %d, want ~1-5", n)
	}
}

func TestEstimateTokenCount_Long(t *testing.T) {
	content := ""
	for i := 0; i < 100; i++ {
		content += "This is a test sentence with approximately fifteen words in it. "
	}
	msgs := []Message{
		{Role: "user", Content: content},
	}
	n := EstimateTokenCount(msgs)
	// ~1500 words, each ~1.3 tokens → ~2000 tokens expected
	if n < 500 || n > 5000 {
		t.Errorf("Long message = %d, want ~2000", n)
	}
}

// ---- S5: Mock compress function tests ----

func TestCompressFuncCalledOnThreshold(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	svc := NewSessionService(store, nil, nil)
	ctx := context.Background()

	// Inject enough messages to exceed threshold
	bigMsg := ""
	for i := 0; i < 2000; i++ {
		bigMsg += "This is a long message to build token count. "
	}

	matchID := uuid.New()
	sid, _ := svc.CreateSession(ctx, matchID)
	store.AppendMessages(ctx, sid, []Message{
		{Role: "user", Content: bigMsg},
		{Role: "assistant", Content: bigMsg},
	})

	compressCalled := false
	svc.SetCompressFunc(func(ctx context.Context, msgs []Message) (string, string, error) {
		compressCalled = true
		// Verify we got the right number of messages
		if len(msgs) != 2 {
			t.Errorf("compress called with %d messages, want 2", len(msgs))
		}
		return "Summary text", `{"key":"value"}`, nil
	})

	// Call maybeCompress directly to trigger auto-compression check
	svc.maybeCompress(ctx, sid)

	if !compressCalled {
		t.Fatal("Compress function should have been called")
	}
}

// ---- helpers ----

// S is a helper to create MatchStatus from string for tests.
func S(s string) model.MatchStatus { return model.MatchStatus(s) }

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}