package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"joblinker/pkg/ai"
)

// DeepRecruiter coordinates multi-agent interactions between Seeker and Recruiter
// It manages the conversation lifecycle and ensures proper message routing
type DeepRecruiter struct {
	seekerAgent    *SeekerAgent
	recruiterAgent *RecruiterAgent
	matchID        uuid.UUID
	state          MatchState
	mu             sync.RWMutex
}

// MatchState represents the current state of a match conversation
type MatchState string

const (
	MatchStateIdle          MatchState = "idle"
	MatchStateInitializing  MatchState = "initializing"
	MatchStateSeekerTurn    MatchState = "seeker_turn"
	MatchStateRecruiterTurn MatchState = "recruiter_turn"
	MatchStateNegotiating   MatchState = "negotiating"
	MatchStateConcluded     MatchState = "concluded"
)

// NewDeepRecruiter creates a new multi-agent coordinator
func NewDeepRecruiter(aiClient *ai.Client, matchID uuid.UUID) *DeepRecruiter {
	return &DeepRecruiter{
		seekerAgent:    NewSeekerAgent(aiClient),
		recruiterAgent: NewRecruiterAgent(aiClient),
		matchID:        matchID,
		state:          MatchStateIdle,
	}
}

// Start initiates a new match conversation
func (d *DeepRecruiter) Start(ctx context.Context, seekerMsg, recruiterMsg string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state != MatchStateIdle {
		return fmt.Errorf("deep recruiter: already started (state=%s)", d.state)
	}

	d.state = MatchStateInitializing

	// Process initial messages if provided
	if seekerMsg != "" {
		if _, err := d.seekerAgent.Chat(ctx, nil, seekerMsg); err != nil {
			return fmt.Errorf("seeker initial message failed: %w", err)
		}
	}

	if recruiterMsg != "" {
		if _, err := d.recruiterAgent.Chat(ctx, nil, recruiterMsg); err != nil {
			return fmt.Errorf("recruiter initial message failed: %w", err)
		}
	}

	d.state = MatchStateSeekerTurn
	return nil
}

// ProcessSeekerMessage handles a message from the job seeker
func (d *DeepRecruiter) ProcessSeekerMessage(ctx context.Context, message string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state == MatchStateConcluded {
		return "", fmt.Errorf("match concluded, cannot process new messages")
	}

	// Generate seeker response
	seekerResp, err := d.seekerAgent.Chat(ctx, nil, message)
	if err != nil {
		return "", fmt.Errorf("seeker response failed: %w", err)
	}

	// Update state
	d.state = MatchStateRecruiterTurn

	return seekerResp, nil
}

// ProcessRecruiterMessage handles a message from the recruiter
func (d *DeepRecruiter) ProcessRecruiterMessage(ctx context.Context, message string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state == MatchStateConcluded {
		return "", fmt.Errorf("match concluded, cannot process new messages")
	}

	// Generate recruiter response
	recruiterResp, err := d.recruiterAgent.Chat(ctx, nil, message)
	if err != nil {
		return "", fmt.Errorf("recruiter response failed: %w", err)
	}

	// Update state
	d.state = MatchStateSeekerTurn

	return recruiterResp, nil
}

// ProcessCoordinatorMessage handles a message from the coordinator/mediator
// This routes between seeker and recruiter based on context
func (d *DeepRecruiter) ProcessCoordinatorMessage(ctx context.Context, message string, target AgentType) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state == MatchStateConcluded {
		return "", fmt.Errorf("match concluded, cannot process new messages")
	}

	switch target {
	case AgentTypeSeeker:
		return d.seekerAgent.Chat(ctx, nil, message)
	case AgentTypeRecruiter:
		return d.recruiterAgent.Chat(ctx, nil, message)
	default:
		return "", fmt.Errorf("unknown agent type: %s", target)
	}
}

// AgentType specifies which agent to target
type AgentType string

const (
	AgentTypeSeeker    AgentType = "seeker"
	AgentTypeRecruiter AgentType = "recruiter"
)

// GetState returns the current match state
func (d *DeepRecruiter) GetState() MatchState {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.state
}

// Conclude ends the match conversation
func (d *DeepRecruiter) Conclude(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state == MatchStateConcluded {
		return fmt.Errorf("already concluded")
	}

	d.state = MatchStateConcluded
	return nil
}

// GetMatchID returns the match UUID
func (d *DeepRecruiter) GetMatchID() uuid.UUID {
	return d.matchID
}

