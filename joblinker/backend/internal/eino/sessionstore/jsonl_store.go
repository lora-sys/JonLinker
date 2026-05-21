package sessionstore

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// JSONLSessionStore implements Store using JSONL files for development.
// Directory structure:
//
//	{dataDir}/{sessionID}/_meta.json      — session metadata
//	{dataDir}/{sessionID}/messages.jsonl  — append-only message log
type JSONLSessionStore struct {
	dataDir string
}

// NewJSONLSessionStore creates a store rooted at dataDir.
// dataDir must be a writable directory path (will be created if missing).
func NewJSONLSessionStore(dataDir string) (*JSONLSessionStore, error) {
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("resolve data dir %s: %w", dataDir, err)
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, fmt.Errorf("create data dir %s: %w", abs, err)
	}
	return &JSONLSessionStore{dataDir: abs}, nil
}

// metaPath returns the metadata file path for a session.
func (s *JSONLSessionStore) metaPath(sessionID string) string {
	return filepath.Join(s.dataDir, sessionID, "_meta.json")
}

// messagesPath returns the messages file path for a session.
func (s *JSONLSessionStore) messagesPath(sessionID string) string {
	return filepath.Join(s.dataDir, sessionID, "messages.jsonl")
}

// sessionDir returns the directory for a session.
func (s *JSONLSessionStore) sessionDir(sessionID string) string {
	return filepath.Join(s.dataDir, sessionID)
}

// metaJSON is the on-disk metadata structure.
type metaJSON struct {
	SessionID        string  `json:"session_id"`
	MatchID          string  `json:"match_id"`
	Version          int     `json:"version"`
	Status           string  `json:"status"`
	State            string  `json:"state"`
	ConclusionReason *string `json:"conclusion_reason,omitempty"`
	MatchStatus      string  `json:"match_status,omitempty"`
}

// messageLine is one line in the JSONL file.
type messageLine struct {
	Seq     int    `json:"seq"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *JSONLSessionStore) Create(ctx context.Context, sessionID string, matchID uuid.UUID, version int) error {
	dir := s.sessionDir(sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create session dir %s: %w", dir, err)
	}

	meta := metaJSON{
		SessionID: sessionID,
		MatchID:   matchID.String(),
		Version:   version,
		Status:    string(SessionStatusActive),
		State:     "idle",
	}
	return writeJSON(s.metaPath(sessionID), meta)
}

func (s *JSONLSessionStore) Load(ctx context.Context, sessionID string) (*Session, error) {
	meta, err := readMeta(s.metaPath(sessionID))
	if err != nil {
		return nil, fmt.Errorf("session %s: %w", sessionID, errNotFound)
	}

	matchID, err := uuid.Parse(meta.MatchID)
	if err != nil {
		return nil, fmt.Errorf("parse match id %s: %w", meta.MatchID, err)
	}

	msgs, err := readMessages(s.messagesPath(sessionID))
	if err != nil {
		// messages file might not exist yet (empty session)
		msgs = nil
	}

	var cr *string
	if meta.ConclusionReason != nil {
		cr = meta.ConclusionReason
	}

	session := &Session{
		SessionID: meta.SessionID,
		MatchID:   matchID,
		Version:   meta.Version,
		Status:    SessionStatus(meta.Status),
		State:     meta.State,
		Messages:  msgs,
	}

	// Restore conclusion reason into state for reopen detection
	if cr != nil {
		reasons := map[string]string{
			"naturally": "concluded_naturally",
			"paused":    "concluded_paused",
		}
		if v, ok := reasons[*cr]; ok {
			session.State = v
		}
	}

	return session, nil
}

func (s *JSONLSessionStore) AppendMessages(ctx context.Context, sessionID string, msgs []Message) error {
	metaPath := s.metaPath(sessionID)
	msgPath := s.messagesPath(sessionID)

	// Verify session exists and is active
	meta, err := readMeta(metaPath)
	if err != nil {
		return fmt.Errorf("session %s: %w", sessionID, errNotFound)
	}
	if meta.Status != string(SessionStatusActive) {
		return fmt.Errorf("session %s is %s, cannot append messages", sessionID, meta.Status)
	}

	// Get current max seq
	existing, _ := readMessages(msgPath)
	maxSeq := 0
	for _, m := range existing {
		if m.Seq > maxSeq {
			maxSeq = m.Seq
		}
	}

	// Append messages
	f, err := os.OpenFile(msgPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open messages file %s: %w", msgPath, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for i, m := range msgs {
		line := messageLine{
			Seq:     maxSeq + i + 1,
			Role:    m.Role,
			Content: m.Content,
		}
		if err := enc.Encode(line); err != nil {
			return fmt.Errorf("write message seq %d: %w", line.Seq, err)
		}
	}
	return nil
}

func (s *JSONLSessionStore) ListByMatchID(ctx context.Context, matchID uuid.UUID) ([]Session, error) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("read data dir %s: %w", s.dataDir, err)
	}

	matchIDStr := matchID.String()
	var sessions []Session

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		// Format: {matchID}@v{version}
		parts := strings.SplitN(dirName, "@v", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] != matchIDStr {
			continue
		}

		meta, err := readMeta(filepath.Join(s.dataDir, dirName, "_meta.json"))
		if err != nil {
			continue
		}

		parsedMID, _ := uuid.Parse(meta.MatchID)
		sessions = append(sessions, Session{
			SessionID: meta.SessionID,
			MatchID:   parsedMID,
			Version:   meta.Version,
			Status:    SessionStatus(meta.Status),
			State:     meta.State,
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Version < sessions[j].Version
	})

	if sessions == nil {
		return []Session{}, nil
	}
	return sessions, nil
}

func (s *JSONLSessionStore) Conclude(ctx context.Context, sessionID string, reason string) error {
	metaPath := s.metaPath(sessionID)
	meta, err := readMeta(metaPath)
	if err != nil {
		return fmt.Errorf("session %s: %w", sessionID, errNotFound)
	}
	meta.Status = string(SessionStatusConcluded)
	meta.ConclusionReason = &reason
	return writeJSON(metaPath, meta)
}

func (s *JSONLSessionStore) LatestVersion(ctx context.Context, matchID uuid.UUID) (int, error) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return 0, nil // Directory not yet created
	}

	matchIDStr := matchID.String()
	maxVer := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		parts := strings.SplitN(entry.Name(), "@v", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] != matchIDStr {
			continue
		}
		var ver int
		if _, err := fmt.Sscanf(parts[1], "%d", &ver); err == nil {
			if ver > maxVer {
				maxVer = ver
			}
		}
	}
	return maxVer, nil
}

func (s *JSONLSessionStore) Delete(ctx context.Context, sessionID string) error {
	dir := s.sessionDir(sessionID)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete session %s: %w", sessionID, err)
	}
	return nil
}

// --- helpers ---

func writeJSON(path string, v interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func readMeta(path string) (*metaJSON, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errNotFound
		}
		return nil, err
	}
	defer f.Close()
	var meta metaJSON
	if err := json.NewDecoder(f).Decode(&meta); err != nil {
		return nil, fmt.Errorf("decode meta: %w", err)
	}
	return &meta, nil
}

func readMessages(path string) ([]Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var msgs []Message
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var ml messageLine
		if err := json.Unmarshal([]byte(line), &ml); err != nil {
			return nil, fmt.Errorf("decode message line: %w", err)
		}
		msgs = append(msgs, Message{
			Seq:     ml.Seq,
			Role:    ml.Role,
			Content: ml.Content,
		})
	}
	return msgs, scanner.Err()
}
