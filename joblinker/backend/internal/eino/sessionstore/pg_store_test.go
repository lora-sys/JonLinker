package sessionstore

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"joblinker/internal/model"
)

// connectTestDB connects to a test PostgreSQL database using DATABASE_URL or a default.
func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Try default joblinker database; the test tables are created via AutoMigrate
		dsn = "host=localhost user=joblinker password=joblinker_dev dbname=joblinker port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Auto-migrate session tables
	if err := db.AutoMigrate(&model.SessionMeta{}, &model.SessionMessage{}); err != nil {
		t.Fatalf("Failed to migrate test tables: %v", err)
	}

	// Clean up any leftover data from previous test runs
	db.Exec("DELETE FROM session_messages")
	db.Exec("DELETE FROM session_meta")

	return db
}

func TestPostgresStore_CreateAndLoad(t *testing.T) {
	db := connectTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)

	// Create
	if err := store.Create(ctx, sid, matchID, 1); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Load
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

func TestPostgresStore_AppendMessages(t *testing.T) {
	db := connectTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)

	msgs := []Message{
		{Role: "user", Content: "Hello from seeker"},
		{Role: "assistant", Content: "Hello from recruiter"},
	}
	if err := store.AppendMessages(ctx, sid, msgs); err != nil {
		t.Fatalf("AppendMessages failed: %v", err)
	}

	session, _ := store.Load(ctx, sid)
	if len(session.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(session.Messages))
	}
	if session.Messages[0].Role != "user" || session.Messages[0].Content != "Hello from seeker" {
		t.Errorf("Message 0 = %+v", session.Messages[0])
	}
	if session.Messages[1].Role != "assistant" || session.Messages[1].Content != "Hello from recruiter" {
		t.Errorf("Message 1 = %+v", session.Messages[1])
	}

	// Append more messages to verify seq continues
	msgs2 := []Message{{Role: "user", Content: "Follow up"}}
	if err := store.AppendMessages(ctx, sid, msgs2); err != nil {
		t.Fatalf("Second AppendMessages failed: %v", err)
	}

	session, _ = store.Load(ctx, sid)
	if len(session.Messages) != 3 {
		t.Fatalf("Expected 3 messages after second append, got %d", len(session.Messages))
	}
	if session.Messages[2].Seq != 3 {
		t.Errorf("Seq should continue; got %d, want 3", session.Messages[2].Seq)
	}
}

func TestPostgresStore_ConcludeBlocksAppend(t *testing.T) {
	db := connectTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)

	// Conclude
	if err := store.Conclude(ctx, sid, "paused"); err != nil {
		t.Fatalf("Conclude failed: %v", err)
	}

	// Verify status
	session, _ := store.Load(ctx, sid)
	if session.Status != SessionStatusConcluded {
		t.Errorf("Status = %s, want concluded", session.Status)
	}

	// Append should fail
	err := store.AppendMessages(ctx, sid, []Message{{Role: "user", Content: "should fail"}})
	if err == nil {
		t.Fatal("Expected error when appending to concluded session, got nil")
	}
}

func TestPostgresStore_ListByMatchID(t *testing.T) {
	db := connectTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	matchID := uuid.New()
	sid1 := SessionID(matchID, 1)
	sid2 := SessionID(matchID, 2)

	store.Create(ctx, sid1, matchID, 1)
	store.Create(ctx, sid2, matchID, 2)

	sessions, err := store.ListByMatchID(ctx, matchID)
	if err != nil {
		t.Fatalf("ListByMatchID failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Version != 1 {
		t.Errorf("First session version = %d, want 1", sessions[0].Version)
	}
	if sessions[1].Version != 2 {
		t.Errorf("Second session version = %d, want 2", sessions[1].Version)
	}
}

func TestPostgresStore_LatestVersion(t *testing.T) {
	db := connectTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	matchID := uuid.New()

	// No sessions yet → version 0
	v, err := store.LatestVersion(ctx, matchID)
	if err != nil {
		t.Fatalf("LatestVersion failed: %v", err)
	}
	if v != 0 {
		t.Errorf("Initial version = %d, want 0", v)
	}

	store.Create(ctx, SessionID(matchID, 1), matchID, 1)
	store.Create(ctx, SessionID(matchID, 2), matchID, 2)

	v, _ = store.LatestVersion(ctx, matchID)
	if v != 2 {
		t.Errorf("Latest version = %d, want 2", v)
	}
}

func TestPostgresStore_Delete(t *testing.T) {
	db := connectTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	matchID := uuid.New()
	sid := SessionID(matchID, 1)
	store.Create(ctx, sid, matchID, 1)
	store.AppendMessages(ctx, sid, []Message{{Role: "user", Content: "test"}})

	// Delete
	if err := store.Delete(ctx, sid); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Load should fail
	_, err := store.Load(ctx, sid)
	if err == nil {
		t.Fatal("Expected error after delete, got nil")
	}
}

func TestPostgresStore_LoadNotFound(t *testing.T) {
	db := connectTestDB(t)
	store := NewPostgresStore(db)
	ctx := context.Background()

	_, err := store.Load(ctx, "nonexistent@v1")
	if err == nil {
		t.Fatal("Expected error for nonexistent session, got nil")
	}
}

func TestSessionIDFormat(t *testing.T) {
	matchID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sid := SessionID(matchID, 1)
	expected := "550e8400-e29b-41d4-a716-446655440000@v1"
	if sid != expected {
		t.Errorf("SessionID = %q, want %q", sid, expected)
	}

	parsedMID, parsedVer, err := ParseSessionID(sid)
	if err != nil {
		t.Fatalf("ParseSessionID failed: %v", err)
	}
	if parsedMID != matchID {
		t.Errorf("Parsed matchID = %v, want %v", parsedMID, matchID)
	}
	if parsedVer != 1 {
		t.Errorf("Parsed version = %d, want 1", parsedVer)
	}
}

func TestParseSessionID_Errors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"invalid"},
		{"not-a-uuid@v1"},
		{"550e8400-e29b-41d4-a716-446655440000@vabc"},
	}
	for _, tt := range tests {
		_, _, err := ParseSessionID(tt.input)
		if err == nil {
			t.Errorf("ParseSessionID(%q) expected error", tt.input)
		}
	}
}
