package agent

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidTransition = errors.New("invalid transition")

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

type Event string

const (
	EventStartSearch           Event = "START_SEARCH"
	EventMatchFound           Event = "MATCH_FOUND"
	EventInterestExpressed    Event = "INTEREST_EXPRESSED"
	EventNegotiate            Event = "NEGOTIATE"
	EventScheduleInterview    Event = "SCHEDULE_INTERVIEW"
	EventInterviewComplete    Event = "INTERVIEW_COMPLETE"
	EventOfferReceived        Event = "OFFER_RECEIVED"
	EventOfferAccepted        Event = "OFFER_ACCEPTED"
	EventOfferDeclined        Event = "OFFER_DECLINED"
	EventPause                Event = "PAUSE"
	EventResume               Event = "RESUME"
	EventTimeout              Event = "TIMEOUT"
	EventRejected             Event = "REJECTED"
)

type FSM struct {
	mu      sync.Mutex
	agentID uuid.UUID
	state   State
	history []State
}

func NewFSM(agentID uuid.UUID) *FSM {
	return &FSM{
		agentID: agentID,
		state:   StateIdle,
		history: []State{StateIdle},
	}
}

func (f *FSM) SetState(state State) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}

func (f *FSM) CurrentState() State {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state
}

func (f *FSM) CanHandle(event Event) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	transitions := getTransitionsForState(f.state)
	_, exists := transitions[event]
	return exists
}

func (f *FSM) Handle(event Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	transitions := getTransitionsForState(f.state)
	nextState, exists := transitions[event]
	if !exists {
		return ErrInvalidTransition
	}

	f.state = nextState
	f.history = append(f.history, nextState)
	return nil
}

// getTransitionsForState returns valid transitions for a given state
// This function is the single source of truth for FSM transitions
func getTransitionsForState(state State) map[Event]State {
	switch state {
	case StateIdle:
		return map[Event]State{
			EventStartSearch: StateSearching,
			EventPause:      StatePaused,
		}
	case StateSearching:
		return map[Event]State{
			EventMatchFound:        StateNegotiating,
			EventInterestExpressed: StateNegotiating,
			EventPause:            StatePaused,
			EventTimeout:          StateIdle,
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
			EventResume:   StateSearching,
			EventRejected: StateRejected,
			EventTimeout:  StateIdle,
		}
	default:
		return map[Event]State{}
	}
}

func (f *FSM) getPreviousState() State {
	if len(f.history) < 2 {
		return StateIdle
	}
	return f.history[len(f.history)-2]
}

// NewFSMFromID creates a new FSM from a string agent ID
func NewFSMFromID(agentID string) *FSM {
	id, err := uuid.Parse(agentID)
	if err != nil {
		id = uuid.New()
	}
	return NewFSM(id)
}

// GetPreviousStateStr returns the previous state as a string for graph routing
func (f *FSM) GetPreviousStateStr() string {
	return string(f.getPreviousState())
}

type StateInfo struct {
	State     State
	EnteredAt time.Time
	ExitedAt  time.Time
	Event     Event
}

func (f *FSM) GetHistory() []StateInfo {
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
