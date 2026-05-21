package sessionstore

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestJSONLStore_CreateAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, err := NewJSONLSessionStore(dir)
	if err != nil {
		t.Fatalf("NewJSONLSessionStore failed: %v", err)
	}

	ctx := context.Background()
	matchID := uuid.New()
	sid := SessionID(matchID, 1)

	if err := store.Create(ctx, sid, matchID, 1); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	session, err := store.Load(ctx, sid)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if session.SessionID != sid {
		t.Errorf("SessionID = %q, want %q", session.SessionID, sid)
	}
	if session.MatchID != matchID {
		t.Errorf("MatchID = %v, want %v", session.MatchID, matchID)
	}
	if session.Version != 1 {
		t.Errorf("Version = %d, want 1", session.Version)
	}
	if session.Status != SessionStatusActive {
		t.Errorf("Status = %s, want active", session.Status)
	}
	if len(session.Messages) != 0 {
		t.Errorf("Messages = %d, want 0", len(session.Messages))
	}
}

func TestJSONLStore_AppendMessages(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)

	msgs := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}
	if err := store.AppendMessages(ctx, sid, msgs); err != nil {
		t.Fatalf("AppendMessages failed: %v", err)
	}

	session, _ := store.Load(ctx, sid)
	if len(session.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(session.Messages))
	}
	if session.Messages[0].Seq != 1 || session.Messages[1].Seq != 2 {
		t.Errorf("Seq not sequential: %+v", session.Messages)
	}

	// Append again
	store.AppendMessages(ctx, sid, []Message{{Role: "user", Content: "Hey again"}})
	session, _ = store.Load(ctx, sid)
	if len(session.Messages) != 3 {
		t.Fatalf("Expected 3 messages after second append, got %d", len(session.Messages))
	}
	if session.Messages[2].Seq != 3 {
		t.Errorf("Seq should continue; got %d, want 3", session.Messages[2].Seq)
	}
}

func TestJSONLStore_ConcludeBlocksAppend(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)

	if err := store.Conclude(ctx, sid, "paused"); err != nil {
		t.Fatalf("Conclude failed: %v", err)
	}

	err := store.AppendMessages(ctx, sid, []Message{{Role: "user", Content: "should fail"}})
	if err == nil {
		t.Fatal("Expected error when appending to concluded session, got nil")
	}
}

func TestJSONLStore_ListByMatchID(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()
	store.Create(ctx, SessionID(matchID, 1), matchID, 1)
	store.Create(ctx, SessionID(matchID, 2), matchID, 2)

	sessions, err := store.ListByMatchID(ctx, matchID)
	if err != nil {
		t.Fatalf("ListByMatchID failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(sessions))
	}
}

func TestJSONLStore_LatestVersion(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()

	v, _ := store.LatestVersion(ctx, matchID)
	if v != 0 {
		t.Errorf("Initial version = %d, want 0", v)
	}

	store.Create(ctx, SessionID(matchID, 2), matchID, 2)
	store.Create(ctx, SessionID(matchID, 5), matchID, 5)

	v, _ = store.LatestVersion(ctx, matchID)
	if v != 5 {
		t.Errorf("Latest version = %d, want 5", v)
	}
}

func TestJSONLStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)
	store.AppendMessages(ctx, sid, []Message{{Role: "user", Content: "test"}})

	if err := store.Delete(ctx, sid); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := store.Load(ctx, sid)
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestJSONLStore_LoadNotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	_, err := store.Load(context.Background(), "nonexistent@v1")
	if err == nil {
		t.Fatal("Expected error for nonexistent session")
	}
}

func TestJSONLStore_ConcludeReason(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)
	store.Conclude(ctx, sid, "naturally")

	session, _ := store.Load(ctx, sid)
	if session.Status != SessionStatusConcluded {
		t.Errorf("Status = %s, want concluded", session.Status)
	}
	// State restored from reason
	if session.State != "concluded_naturally" {
		t.Errorf("State = %s, want concluded_naturally", session.State)
	}
}

func TestJSONLStore_EmptyMessagesFile(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)

	// Load without any messages appended
	session, err := store.Load(ctx, sid)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(session.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(session.Messages))
	}
}

func TestJSONLStore_PersistenceAcrossStores(t *testing.T) {
	dir := t.TempDir()
	store1, _ := NewJSONLSessionStore(dir)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store1.Create(ctx, sid, matchID, 1)
	store1.AppendMessages(ctx, sid, []Message{{Role: "user", Content: "Hello"}})

	// Create a new store pointing to same dir
	store2, _ := NewJSONLSessionStore(dir)
	session, err := store2.Load(ctx, sid)
	if err != nil {
		t.Fatalf("Load from second store failed: %v", err)
	}
	if len(session.Messages) != 1 || session.Messages[0].Content != "Hello" {
		t.Errorf("Data not persisted: %+v", session.Messages)
	}
}

// Test against os.CreateTemp used by t.TempDir
var _ = os.CreateTemp
