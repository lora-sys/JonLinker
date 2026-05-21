package sessionstore

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"joblinker/internal/eino/agent"
	"joblinker/internal/model"
)

// SessionService orchestrates the Load → RunTurn → AppendMessages lifecycle.
// It is stateless: all session data lives in the Store.
type SessionService struct {
	store     Store
	seeker    *agent.SeekerAgent
	recruiter *agent.RecruiterAgent
}

// NewSessionService creates a service that reads/writes sessions via store
// and drives conversation via the given agents.
func NewSessionService(store Store, seeker *agent.SeekerAgent, recruiter *agent.RecruiterAgent) *SessionService {
	return &SessionService{
		store:     store,
		seeker:    seeker,
		recruiter: recruiter,
	}
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

	return reply, nil
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
