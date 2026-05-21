package sessionstore

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"joblinker/internal/eino/agent"
	"joblinker/internal/model"
)

// CompressionThreshold is the estimated token count beyond which a conversation
// snapshot is automatically compressed.
const CompressionThreshold = 4000

// CompressFunc is called when a conversation exceeds the token threshold,
// allowing callers to plug in an LLM-based summarizer (e.g., S5).
// Return the structured summary text and key facts JSON string.
type CompressFunc func(ctx context.Context, msgs []Message) (summaryText, keyFactsJSON string, err error)

// SessionService orchestrates the Load → RunTurn → AppendMessages lifecycle.
// It is stateless: all session data lives in the Store.
type SessionService struct {
	store      Store
	seeker     *agent.SeekerAgent
	recruiter  *agent.RecruiterAgent
	compressFn CompressFunc // optional; set for auto-compression (S5)
}

// NewSessionService creates a service that reads/writes sessions via store
// and drives conversation via the given agents.
func NewSessionService(store Store, seeker *agent.SeekerAgent, recruiter *agent.RecruiterAgent) *SessionService {
	return &SessionService{
		store:      store,
		seeker:     seeker,
		recruiter:  recruiter,
	}
}

// SetCompressFunc sets the optional compression function (S5 feature).
func (s *SessionService) SetCompressFunc(fn CompressFunc) {
	s.compressFn = fn
}

// ProcessMessage loads the session, runs the agent turn, appends messages, and returns the reply.
// sender is either "seeker" or "recruiter".
func (s *SessionService) ProcessMessage(ctx context.Context, sessionID, sender, content string) (string, error) {
	session, err := s.store.Load(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("load session %s: %w", sessionID, err)
	}

	if session.Status == SessionStatusConcluded {
		return "", fmt.Errorf("session %s is concluded, cannot process messages", sessionID)
	}

	// Convert stored messages to schema history
	history := ToSchemaMessages(session.Messages)

	// Run the agent turn
	var reply string
	switch sender {
	case "seeker":
		reply, err = s.seeker.Chat(ctx, history, content)
	case "recruiter":
		reply, err = s.recruiter.Chat(ctx, history, content)
	default:
		return "", fmt.Errorf("unknown sender: %s", sender)
	}
	if err != nil {
		return "", fmt.Errorf("%s chat: %w", sender, err)
	}

	// Append both the user message and the reply
	msgs := []Message{
		{Role: sender, Content: content},
		{Role: "assistant", Content: reply},
	}
	if err := s.store.AppendMessages(ctx, sessionID, msgs); err != nil {
		return "", fmt.Errorf("append messages: %w", err)
	}

	// Auto-compression check (S5): after append, check total token count
	s.maybeCompress(ctx, sessionID)
	return reply, nil
}

// maybeCompress checks the token count and triggers compression if it exceeds the threshold.
// This is a no-op if no CompressFunc is set. Best-effort: errors are logged, not returned.
func (s *SessionService) maybeCompress(ctx context.Context, sessionID string) {
	if s.compressFn == nil {
		return
	}
	updatedSession, loadErr := s.store.Load(ctx, sessionID)
	if loadErr != nil {
		log.Printf("[SessionService] Failed to reload session for compression check: %v", loadErr)
		return
	}
	totalTokens := EstimateTokenCount(updatedSession.Messages)
	if totalTokens <= CompressionThreshold {
		return
	}
	log.Printf("[SessionService] Auto-compression triggered for %s: %d tokens", sessionID, totalTokens)
	summaryText, keyFactsJSON, compressErr := s.compressFn(ctx, updatedSession.Messages)
	if compressErr != nil {
		log.Printf("[SessionService] Compression failed: %v", compressErr)
		return
	}
	summary := model.SessionSummary{
		SessionID:    sessionID,
		Round:        len(updatedSession.Messages),
		SummaryText:  summaryText,
		KeyFactsJSON: keyFactsJSON,
		TokenCount:   totalTokens,
	}
	if saveErr := s.saveSummary(ctx, &summary); saveErr != nil {
		log.Printf("[SessionService] Failed to save summary: %v", saveErr)
	}
}

// saveSummary persists a SessionSummary to the store's underlying DB.
// By default this is a no-op unless a concrete store (Postgres) is wired.
func (s *SessionService) saveSummary(ctx context.Context, summary *model.SessionSummary) error {
	// If the store has a DB, save it. For JSONL store we log but don't persist.
	// This is a best-effort operation.
	if pg, ok := s.store.(*PostgresStore); ok {
		return pg.db.WithContext(ctx).Create(summary).Error
	}
	log.Printf("[SessionService] Summary saved (store type: %T): session=%s round=%d tokens=%d",
		s.store, summary.SessionID, summary.Round, summary.TokenCount)
	return nil
}

// CreateSession creates a new session for a match and returns the session ID.
func (s *SessionService) CreateSession(ctx context.Context, matchID uuid.UUID) (string, error) {
	ver, err := s.store.LatestVersion(ctx, matchID)
	if err != nil {
		return "", fmt.Errorf("get latest version: %w", err)
	}
	newVer := ver + 1
	sid := SessionID(matchID, newVer)
	if err := s.store.Create(ctx, sid, matchID, newVer); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return sid, nil
}

// --- S4: Session Versioning + Reopen ---

// ConclusionReason constants.
const (
	ConcludedNaturally = "naturally"
	ConcludedPaused    = "paused"
)

// ReopenSession creates a new session version for a paused match.
// Rules:
//   - Load latest session for the match
//   - Verify it is concluded_paused (not concluded_naturally)
//   - Create new version, inject system prompt with v1 summary
//   - New session has empty Messages
//
// Returns the new session ID.
func (s *SessionService) ReopenSession(ctx context.Context, matchID uuid.UUID) (string, error) {
	// Find latest session
	ver, err := s.store.LatestVersion(ctx, matchID)
	if err != nil {
		return "", fmt.Errorf("latest version for %s: %w", matchID, err)
	}
	if ver == 0 {
		return "", fmt.Errorf("no sessions found for match %s", matchID)
	}

	// Load the latest session
	oldSid := SessionID(matchID, ver)
	oldSession, err := s.store.Load(ctx, oldSid)
	if err != nil {
		return "", fmt.Errorf("load latest session %s: %w", oldSid, err)
	}

	// Verify it is concluded
	if oldSession.Status != SessionStatusConcluded {
		return "", fmt.Errorf("session %s is not concluded (status=%s), cannot reopen", oldSid, oldSession.Status)
	}

	// Verify it was paused (not naturally ended)
	if oldSession.State == "concluded_naturally" {
		return "", fmt.Errorf("session %s ended naturally, cannot reopen", oldSid)
	}

	// Build summary from old session's messages
	summary := s.buildSessionSummary(oldSession.Messages)

	// Create new version
	newVer := ver + 1
	newSid := SessionID(matchID, newVer)
	if err := s.store.Create(ctx, newSid, matchID, newVer); err != nil {
		return "", fmt.Errorf("create new session %s: %w", newSid, err)
	}

	// If there's a summary, inject it as the first system message
	if summary != "" {
		systemPrompt := fmt.Sprintf(
			"## 历史 Session 摘要 (v%d)\n%s\n\n## 当前匹配状态\nmatch.status = %s",
			ver, summary, string(model.MatchStatusPaused),
		)
		if err := s.store.AppendMessages(ctx, newSid, []Message{
			{Role: "system", Content: systemPrompt},
		}); err != nil {
			log.Printf("[SessionService] Failed to inject history summary into %s: %v", newSid, err)
		}
	}

	return newSid, nil
}

// buildSessionSummary creates a concise plain-text summary of a session's messages.
func (s *SessionService) buildSessionSummary(msgs []Message) string {
	if len(msgs) == 0 {
		return ""
	}

	var b strings.Builder
	// Extract key information from messages
	keyPoints := extractKeyPoints(msgs)
	for _, kp := range keyPoints {
		b.WriteString("- ")
		b.WriteString(kp)
		b.WriteString("\n")
	}

	summary := b.String()
	if len(summary) > 2000 {
		summary = summary[:2000] + "..."
	}
	return summary
}

// extractKeyPoints extracts meaningful points from conversation messages.
func extractKeyPoints(msgs []Message) []string {
	var points []string
	for i, m := range msgs {
		// Only capture substantive messages (ignore very short or system)
		if m.Role == "system" {
			continue
		}
		content := strings.TrimSpace(m.Content)
		if len(content) < 10 {
			continue
		}
		// Truncate long messages
		preview := content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		prefix := m.Role
		if prefix == "assistant" {
			prefix = "回复"
		}
		points = append(points, fmt.Sprintf("[%d] %s: %s", i+1, prefix, preview))
		// Max 50 summary points
		if len(points) >= 50 {
			break
		}
	}
	return points
}

// --- Match Status Auto-Update ---

// UpdateMatchStatusForSession determines the appropriate MatchStatus based on
// the latest session's state and the current match status.
// This implements the PRD Decision 8 rules.
func (s *SessionService) UpdateMatchStatusForSession(currentStatus model.MatchStatus, sessionState string) model.MatchStatus {
	switch sessionState {
	case "idle":
		return model.MatchStatusPending
	case "negotiating":
		return model.MatchStatusNegotiating
	case "interviewing":
		return model.MatchStatusInterviewing
	case "offered":
		return model.MatchStatusOffered
	case "concluded_naturally":
		// If concluded naturally, check current status
		if currentStatus == model.MatchStatusOffered {
			return model.MatchStatusHired
		}
		return model.MatchStatusRejected
	case "concluded_paused":
		return model.MatchStatusPaused
	}
	return currentStatus
}

// GetSessionHistory returns all messages for a session (for streaming or display).
func (s *SessionService) GetSessionHistory(ctx context.Context, sessionID string) ([]Message, error) {
	session, err := s.store.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return session.Messages, nil
}

// GetSeekerMessages returns messages sent by the seeker in this session.
func (s *SessionService) GetSeekerMessages(ctx context.Context, sessionID string) ([]Message, error) {
	session, err := s.store.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	var seekerMsgs []Message
	for _, m := range session.Messages {
		if m.Role == "seeker" {
			seekerMsgs = append(seekerMsgs, m)
		}
	}
	return seekerMsgs, nil
}

// GetRecruiterMessages returns messages sent by the recruiter in this session.
func (s *SessionService) GetRecruiterMessages(ctx context.Context, sessionID string) ([]Message, error) {
	session, err := s.store.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	var recruiterMsgs []Message
	for _, m := range session.Messages {
		if m.Role == "recruiter" || m.Role == "assistant" {
			recruiterMsgs = append(recruiterMsgs, m)
		}
	}
	return recruiterMsgs, nil
}

// RunTurn is a stateless operation: given conversation history and a user message,
// it picks the right agent, runs it, and returns the reply.
// This replaces DeepRecruiter's stateful ProcessSeekerMessage/ProcessRecruiterMessage.
func (s *SessionService) RunTurn(ctx context.Context, history []Message, sender, content string) (string, error) {
	schemaHistory := ToSchemaMessages(history)
	switch sender {
	case "seeker":
		return s.seeker.Chat(ctx, schemaHistory, content)
	case "recruiter":
		return s.recruiter.Chat(ctx, schemaHistory, content)
	default:
		return "", fmt.Errorf("unknown sender: %s", sender)
	}
}

// MatchStateFromStatus maps model.MatchStatus to agent.MatchState.
func MatchStateFromStatus(status model.MatchStatus) agent.MatchState {
	switch status {
	case model.MatchStatusPending, model.MatchStatusSearching:
		return agent.MatchStateIdle
	case model.MatchStatusMutualInterest:
		return agent.MatchStateInitializing
	case model.MatchStatusNegotiating:
		return agent.MatchStateNegotiating
	case model.MatchStatusInterviewing:
		return agent.MatchStateSeekerTurn
	case model.MatchStatusOffered:
		return agent.MatchStateRecruiterTurn
	case model.MatchStatusHired, model.MatchStatusRejected:
		return agent.MatchStateConcluded
	default:
		return agent.MatchStateIdle
	}
}

// ToSchemaMessages converts our internal Message slice to schema.Message slice.
func ToSchemaMessages(msgs []Message) []*schema.Message {
	result := make([]*schema.Message, 0, len(msgs))
	for _, m := range msgs {
		var role schema.RoleType
		switch m.Role {
		case "seeker", "user":
			role = schema.User
		case "recruiter", "assistant":
			role = schema.Assistant
		case "system":
			role = schema.System
		default:
			role = schema.User
		}
		result = append(result, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return result
}

// Ensure imports used
var _ = log.Printf