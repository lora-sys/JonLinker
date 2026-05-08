package agent

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"joblinker/internal/eino/memory"
	"joblinker/internal/model"
	"joblinker/pkg/ai"
)

// DeepRecruiter coordinates multi-agent interactions between Seeker and Recruiter
// It manages the conversation lifecycle and ensures proper message routing
type DeepRecruiter struct {
	seekerAgent *SeekerAgent
	recruiterAgent *RecruiterAgent
	seekerMemory *memory.AgentMemory
	recruiterMemory *memory.AgentMemory
	matchID uuid.UUID
	state MatchState
	mu    sync.RWMutex
}

// MatchState represents the current state of a match conversation
type MatchState string

const (
	MatchStateIdle         MatchState = "idle"
	MatchStateInitializing MatchState = "initializing"
	MatchStateSeekerTurn   MatchState = "seeker_turn"
	MatchStateRecruiterTurn MatchState = "recruiter_turn"
	MatchStateNegotiating  MatchState = "negotiating"
	MatchStateConcluded    MatchState = "concluded"
)

// NewDeepRecruiter creates a new multi-agent coordinator
func NewDeepRecruiter(aiClient *ai.Client, matchID uuid.UUID) *DeepRecruiter {
	dr := &DeepRecruiter{
		seekerAgent:   NewSeekerAgent(aiClient),
		recruiterAgent: NewRecruiterAgent(aiClient),
		matchID:     matchID,
		state:       MatchStateIdle,
	}

	// Initialize separate memories for each agent
	dr.seekerMemory = memory.NewAgentMemory(nil, matchID)
	dr.recruiterMemory = memory.NewAgentMemory(nil, matchID)

	dr.seekerAgent.SetMemory(dr.seekerMemory)
	dr.recruiterAgent.SetMemory(dr.recruiterMemory)

	return dr
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
		if _, err := d.seekerAgent.ChatWithMemory(ctx, seekerMsg); err != nil {
			return fmt.Errorf("seeker initial message failed: %w", err)
		}
	}

	if recruiterMsg != "" {
		if _, err := d.recruiterAgent.ChatWithMemory(ctx, recruiterMsg); err != nil {
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

	// Add seeker message to memory
	if err := d.seekerAgent.AddMemoryMessage(ctx, schema.User, message); err != nil {
		log.Printf("Warning: failed to add seeker message to memory: %v", err)
	}

	// Generate seeker response
	seekerResp, err := d.seekerAgent.ChatWithMemory(ctx, message)
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

	// Add recruiter message to memory
	if err := d.recruiterAgent.AddMemoryMessage(ctx, schema.User, message); err != nil {
		log.Printf("Warning: failed to add recruiter message to memory: %v", err)
	}

	// Generate recruiter response
	recruiterResp, err := d.recruiterAgent.ChatWithMemory(ctx, message)
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
		return d.seekerAgent.ChatWithMemory(ctx, message)
	case AgentTypeRecruiter:
		return d.recruiterAgent.ChatWithMemory(ctx, message)
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

// GetSeekerMessages returns all seeker conversation messages
func (d *DeepRecruiter) GetSeekerMessages() []*schema.Message {
	if d.seekerAgent == nil {
		return nil
	}
	return d.seekerAgent.GetMemoryMessages()
}

// GetRecruiterMessages returns all recruiter conversation messages
func (d *DeepRecruiter) GetRecruiterMessages() []*schema.Message {
	if d.recruiterAgent == nil {
		return nil
	}
	return d.recruiterAgent.GetMemoryMessages()
}

// Conclude ends the match conversation
func (d *DeepRecruiter) Conclude(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state == MatchStateConcluded {
		return fmt.Errorf("already concluded")
	}

	// Store final summary for both agents
	if summary, err := d.generateSummary(ctx); err != nil {
		log.Printf("Warning: failed to generate summary: %v", err)
	} else {
		_ = d.seekerMemory.StoreConversationSummary(ctx, summary)
		_ = d.recruiterMemory.StoreConversationSummary(ctx, summary)
	}

	d.state = MatchStateConcluded
	return nil
}

// generateSummary creates a conversation summary for long-term memory
func (d *DeepRecruiter) generateSummary(ctx context.Context) (*model.ConversationSummary, error) {
	// Gather all messages
	seekerMsgs := d.GetSeekerMessages()
	recruiterMsgs := d.GetRecruiterMessages()

	// Build summary text
	summaryText := fmt.Sprintf(
		"Match %s: %d seeker messages, %d recruiter messages. Final state: %s",
		d.matchID,
		len(seekerMsgs),
		len(recruiterMsgs),
		d.state,
	)

	return &model.ConversationSummary{
		MatchID:     d.matchID,
		SummaryText: summaryText,
		KeyFacts:    []model.KeyFact{},
		TokenCount:  0,
	}, nil
}

// GetMatchID returns the match UUID
func (d *DeepRecruiter) GetMatchID() uuid.UUID {
	return d.matchID
}
