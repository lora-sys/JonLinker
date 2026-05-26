package core

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// State represents the current status of an agent or match process.
type State string

const (
	StateIdle          State = "idle"
	StateSearching     State = "searching"
	StateNegotiating   State = "negotiating"
	StateInterviewing  State = "interviewing"
	StateOfferReceived State = "offer_received"
	StateHired         State = "hired"
	StateRejected      State = "rejected"
	StatePaused        State = "paused"
)

// Event represents an action that triggers a state transition.
type Event string

const (
	EventStartSearch        Event = "START_SEARCH"
	EventMatchFound         Event = "MATCH_FOUND"
	EventInterestExpressed  Event = "INTEREST_EXPRESSED"
	EventNegotiate          Event = "NEGOTIATE"
	EventScheduleInterview  Event = "SCHEDULE_INTERVIEW"
	EventInterviewComplete  Event = "INTERVIEW_COMPLETE"
	EventOfferReceived      Event = "OFFER_RECEIVED"
	EventOfferAccepted      Event = "OFFER_ACCEPTED"
	EventOfferDeclined      Event = "OFFER_DECLINED"
	EventPause              Event = "PAUSE"
	EventResume             Event = "RESUME"
	EventTimeout            Event = "TIMEOUT"
	EventRejected           Event = "REJECTED"
)

// FSM is a thread-safe finite state machine for agent negotiation flows.
type FSM struct {
	mu      sync.Mutex
	agentID uuid.UUID
	state   State
	history []State
}

// NewFSM creates a new FSM in the initial Idle state.
func NewFSM(agentID uuid.UUID) *FSM {
	return &FSM{
		agentID: agentID,
		state:   StateIdle,
		history: []State{StateIdle},
	}
}

// NewFSMFromID creates a new FSM from a string agent ID.
func NewFSMFromID(agentID string) *FSM {
	id, err := uuid.Parse(agentID)
	if err != nil {
		id = uuid.New()
	}
	return NewFSM(id)
}

// SetState sets the current state directly (use with caution).
func (f *FSM) SetState(state State) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}

// CurrentState returns the current state.
func (f *FSM) CurrentState() State {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state
}

// CanHandle returns true if the given event is valid in the current state.
func (f *FSM) CanHandle(event Event) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	transitions := getTransitionsForState(f.state)
	_, exists := transitions[event]
	return exists
}

// Handle transitions the FSM to the next state based on the given event.
// Returns an error if the transition is not allowed in the current state.
func (f *FSM) Handle(event Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Special case: Resume from Paused returns to previous state
	if f.state == StatePaused && event == EventResume {
		f.state = f.getPreviousState()
		f.history = append(f.history, f.state)
		return nil
	}

	transitions := getTransitionsForState(f.state)
	nextState, exists := transitions[event]
	if !exists {
		return errors.New("invalid transition: " + string(event) + " from " + string(f.state))
	}

	f.state = nextState
	f.history = append(f.history, nextState)
	return nil
}

// getTransitionsForState returns valid transitions for a given state.
// This is the single source of truth for FSM transitions.
func getTransitionsForState(state State) map[Event]State {
	switch state {
	case StateIdle:
		return map[Event]State{
			EventStartSearch: StateSearching,
			EventPause:       StatePaused,
		}
	case StateSearching:
		return map[Event]State{
			EventMatchFound:        StateNegotiating,
			EventInterestExpressed: StateNegotiating,
			EventPause:             StatePaused,
			EventTimeout:           StateIdle,
		}
	case StateNegotiating:
		return map[Event]State{
			EventScheduleInterview: StateInterviewing,
			EventOfferReceived:     StateOfferReceived,
			EventRejected:          StateRejected,
			EventPause:             StatePaused,
			EventTimeout:           StateIdle,
		}
	case StateInterviewing:
		return map[Event]State{
			EventInterviewComplete: StateNegotiating,
			EventOfferReceived:     StateOfferReceived,
			EventRejected:          StateRejected,
			EventPause:             StatePaused,
			EventTimeout:           StateNegotiating,
		}
	case StateOfferReceived:
		return map[Event]State{
			EventOfferAccepted: StateHired,
			EventOfferDeclined: StateRejected,
			EventNegotiate:     StateNegotiating,
		}
	case StatePaused:
		return map[Event]State{
			EventResume:   StateIdle, // Handle() provides special Resume logic
			EventRejected: StateRejected,
			EventTimeout:  StateIdle,
		}
	case StateHired:
		return map[Event]State{
			EventRejected: StateRejected,
		}
	default:
		return map[Event]State{}
	}
}

// getPreviousState returns the state before the current one.
func (f *FSM) getPreviousState() State {
	if len(f.history) < 2 {
		return StateIdle
	}
	return f.history[len(f.history)-2]
}

// GetPreviousStateStr returns the previous state as a string for graph routing.
func (f *FSM) GetPreviousStateStr() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return string(f.getPreviousState())
}

// StateInfo holds a snapshot of a single state transition in FSM history.
type StateInfo struct {
	State     State
	EnteredAt time.Time
	ExitedAt  time.Time
	Event     Event
}

// GetHistory returns a copy of the full state transition history.
func (f *FSM) GetHistory() []StateInfo {
	f.mu.Lock()
	defer f.mu.Unlock()

	history := make([]StateInfo, 0, len(f.history))
	for i, state := range f.history {
		info := StateInfo{
			State: state,
		}
		if i > 0 {
			info.ExitedAt = time.Now()
		}
		history = append(history, info)
	}
	return history
}
