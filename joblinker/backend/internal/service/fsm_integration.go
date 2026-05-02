package service

import (
	"log"

	"joblinker/internal/agent"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

// FSMIntegration maps intents to FSM events and handles state transitions
type FSMIntegration struct {
	matchRepo *repository.MatchRepository
}

func NewFSMIntegration(matchRepo *repository.MatchRepository) *FSMIntegration {
	return &FSMIntegration{matchRepo: matchRepo}
}

// IntentToEvent maps AI intent strings to FSM events
var IntentToEvent = map[string]agent.Event{
	"INTRODUCTION":        agent.EventStartSearch,
	"INTEREST_EXPRESSED": agent.EventInterestExpressed,
	"INTEREST":           agent.EventInterestExpressed,
	"NEGOTIATE":          agent.EventNegotiate,
	"SCHEDULE_INTERVIEW": agent.EventScheduleInterview,
	"INTERVIEW_COMPLETE": agent.EventInterviewComplete,
	"OFFER_CREATED":      agent.EventOfferReceived,
	"OFFER_ACCEPTED":     agent.EventOfferAccepted,
	"OFFER_DECLINED":     agent.EventOfferDeclined,
	"REJECTED":           agent.EventRejected,
	"PAUSE":              agent.EventPause,
	"RESUME":             agent.EventResume,
}

// TransitionFSM transitions the match's FSM state based on intent
// Returns the new state and whether a transition occurred
func (s *FSMIntegration) TransitionFSM(matchID uuid.UUID, intent string) (agent.State, bool, error) {
	// Get match to find current FSM state
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return agent.StateIdle, false, err
	}

	// Create FSM from current state - use match ID as agentID for routing
	fsm := agent.NewFSM(matchID)

	// Load existing state from match if not idle
	if match.FSMState != "" && match.FSMState != "idle" {
		fsm.SetState(agent.State(match.FSMState))
	}

	// Map intent to event
	event, exists := IntentToEvent[intent]
	if !exists {
		return agent.StateIdle, false, nil
	}

	// Attempt transition
	if !fsm.CanHandle(event) {
		return fsm.CurrentState(), false, nil
	}

	if err := fsm.Handle(event); err != nil {
		return fsm.CurrentState(), false, err
	}

	newState := fsm.CurrentState()

	// Persist state change
	if err := s.matchRepo.UpdateFSMState(matchID, string(newState)); err != nil {
		return newState, false, err
	}

	return newState, true, nil
}

// GetCurrentState returns the current FSM state for a match
func (s *FSMIntegration) GetCurrentState(matchID uuid.UUID) (agent.State, error) {
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return agent.StateIdle, err
	}
	if match.FSMState == "" {
		return agent.StateIdle, nil
	}
	return agent.State(match.FSMState), nil
}

// ValidateTransition checks if a transition is valid without executing it
func (s *FSMIntegration) ValidateTransition(matchID uuid.UUID, intent string) (bool, error) {
	event, exists := IntentToEvent[intent]
	if !exists {
		return false, nil
	}

	fsm := agent.NewFSM(matchID)
	return fsm.CanHandle(event), nil
}