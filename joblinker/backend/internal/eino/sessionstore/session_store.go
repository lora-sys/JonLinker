package sessionstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// errNotFound is returned when a session is not found.
var errNotFound = errors.New("session not found")

// SessionID builds a versioned session identifier: "{matchID}@v{version}".
func SessionID(matchID uuid.UUID, version int) string {
	return fmt.Sprintf("%s@v%d", matchID.String(), version)
}

// ParseSessionID extracts matchID and version from a session ID string.
// Returns an error if the format is invalid.
func ParseSessionID(sessionID string) (matchID uuid.UUID, version int, err error) {
	parts := strings.SplitN(sessionID, "@v", 2)
	if len(parts) != 2 {
		return uuid.Nil, 0, fmt.Errorf("invalid session ID format: %s", sessionID)
	}
	matchID, err = uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("invalid match UUID in session ID %s: %w", sessionID, err)
	}
	version, err = strconv.Atoi(parts[1])
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("invalid version in session ID %s: %w", sessionID, err)
	}
	return matchID, version, nil
}

// SessionStatus represents the lifecycle status of a session.
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusConcluded SessionStatus = "concluded"
)

// Session wraps metadata and messages for a single session.
type Session struct {
	SessionID string
	MatchID   uuid.UUID
	Version   int
	Status    SessionStatus
	State     string
	Messages  []Message
}

// Message is a single message within a session.
type Message struct {
	Seq     int
	Role    string
	Content string
}

// Store defines the storage contract for session data.
// Implementations: PostgreSQL (production), JSONL (development).
type Store interface {
	// Create persists a new session. The session starts with Status=active and empty messages.
	Create(ctx context.Context, sessionID string, matchID uuid.UUID, version int) error

	// Load retrieves a session (meta + all messages) by its session ID.
	// Returns ErrNotFound if the session does not exist.
	Load(ctx context.Context, sessionID string) (*Session, error)

	// AppendMessages appends one or more messages to an existing session.
	// Returns an error if the session is not in active status.
	AppendMessages(ctx context.Context, sessionID string, msgs []Message) error

	// ListByMatchID returns all session metadata for a given match, ordered by version ascending.
	ListByMatchID(ctx context.Context, matchID uuid.UUID) ([]Session, error)

	// Conclude marks a session as concluded with the given reason.
	// Once concluded, AppendMessages will return an error.
	Conclude(ctx context.Context, sessionID string, reason string) error

	// LatestVersion returns the latest version number for a given match.
	// Returns 0 if no sessions exist for this match.
	LatestVersion(ctx context.Context, matchID uuid.UUID) (int, error)

	// Delete removes a session (both meta and messages).
	Delete(ctx context.Context, sessionID string) error
}
