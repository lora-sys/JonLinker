package workflow

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"

	"joblinker/internal/agent"
)

// FSMEvent represents an event that can trigger a state transition
type FSMEvent struct {
	Event   agent.Event
	Payload interface{}
}

// FSMTransitionResult represents the result of a state transition
type FSMTransitionResult struct {
	PreviousState agent.State
	CurrentState agent.State
	Event        agent.Event
	Transitioned bool
	Error        error
}

// FSMGraph wraps the FSM state machine with Eino Graph for workflow orchestration
type FSMGraph struct {
	fsm         *agent.FSM
	graph       *compose.Graph[FSMEvent, FSMTransitionResult]
	transition  compose.Runnable[FSMEvent, FSMTransitionResult]
}

// NewFSMGraph creates a new FSM workflow using Eino Graph
func NewFSMGraph(agentID string) (*FSMGraph, error) {
	fsm := agent.NewFSMFromID(agentID)

	g := compose.NewGraph[FSMEvent, FSMTransitionResult]()

	// Add a passthrough node to capture current state
	g.AddPassthroughNode("capture_state")

	// Add branches for each state
	// Each branch condition checks if we're in the current state and returns the next node

	// Branch: idle state transitions
	idleBranch := compose.NewGraphBranch(func(ctx context.Context, in FSMEvent) (endNode string, err error) {
		if fsm.CurrentState() != agent.StateIdle {
			return "", nil // Not in idle state, skip
		}

		switch in.Event {
		case agent.EventStartSearch:
			return "searching", nil
		case agent.EventPause:
			return "paused", nil
		default:
			return "", nil
		}
	}, map[string]bool{"idle": true, "searching": true, "paused": true})
	g.AddBranch("capture_state", idleBranch)

	// Branch: searching state transitions
	searchingBranch := compose.NewGraphBranch(func(ctx context.Context, in FSMEvent) (endNode string, err error) {
		if fsm.CurrentState() != agent.StateSearching {
			return "", nil
		}

		switch in.Event {
		case agent.EventMatchFound, agent.EventInterestExpressed:
			return "negotiating", nil
		case agent.EventPause:
			return "paused", nil
		case agent.EventTimeout:
			return "idle", nil
		default:
			return "", nil
		}
	}, map[string]bool{"searching": true, "negotiating": true, "idle": true, "paused": true})
	g.AddBranch("capture_state", searchingBranch)

	// Branch: negotiating state transitions
	negotiatingBranch := compose.NewGraphBranch(func(ctx context.Context, in FSMEvent) (endNode string, err error) {
		if fsm.CurrentState() != agent.StateNegotiating {
			return "", nil
		}

		switch in.Event {
		case agent.EventScheduleInterview:
			return "interviewing", nil
		case agent.EventOfferReceived:
			return "offer_received", nil
		case agent.EventRejected:
			return "rejected", nil
		case agent.EventPause:
			return "paused", nil
		case agent.EventTimeout:
			return "idle", nil
		default:
			return "", nil
		}
	}, map[string]bool{"negotiating": true, "interviewing": true, "offer_received": true, "rejected": true, "paused": true, "idle": true})
	g.AddBranch("capture_state", negotiatingBranch)

	// Branch: interviewing state transitions
	interviewingBranch := compose.NewGraphBranch(func(ctx context.Context, in FSMEvent) (endNode string, err error) {
		if fsm.CurrentState() != agent.StateInterviewing {
			return "", nil
		}

		switch in.Event {
		case agent.EventInterviewComplete:
			return "negotiating", nil
		case agent.EventOfferReceived:
			return "offer_received", nil
		case agent.EventRejected:
			return "rejected", nil
		case agent.EventPause:
			return "paused", nil
		case agent.EventTimeout:
			return "negotiating", nil
		default:
			return "", nil
		}
	}, map[string]bool{"interviewing": true, "negotiating": true, "offer_received": true, "rejected": true, "paused": true})
	g.AddBranch("capture_state", interviewingBranch)

	// Branch: offer_received state transitions
	offerReceivedBranch := compose.NewGraphBranch(func(ctx context.Context, in FSMEvent) (endNode string, err error) {
		if fsm.CurrentState() != agent.StateOfferReceived {
			return "", nil
		}

		switch in.Event {
		case agent.EventOfferAccepted:
			return "hired", nil
		case agent.EventOfferDeclined:
			return "rejected", nil
		case agent.EventNegotiate:
			return "negotiating", nil
		default:
			return "", nil
		}
	}, map[string]bool{"offer_received": true, "hired": true, "rejected": true, "negotiating": true})
	g.AddBranch("capture_state", offerReceivedBranch)

	// Branch: paused state transitions (resume to previous)
	pausedBranch := compose.NewGraphBranch(func(ctx context.Context, in FSMEvent) (endNode string, err error) {
		if fsm.CurrentState() != agent.StatePaused {
			return "", nil
		}

		if in.Event == agent.EventResume {
			return fsm.GetPreviousStateStr(), nil
		}
		return "", nil
	}, map[string]bool{"paused": true})
	g.AddBranch("capture_state", pausedBranch)

	// Compile the graph
	ctx := context.Background()
	runnable, err := g.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile FSM graph: %w", err)
	}

	return &FSMGraph{
		fsm:        fsm,
		graph:      g,
		transition: runnable,
	}, nil
}

// Transition processes an event and returns the transition result
func (f *FSMGraph) Transition(ctx context.Context, event FSMEvent) (*FSMTransitionResult, error) {
	previousState := f.fsm.CurrentState()

	// Try to handle the event
	if !f.fsm.CanHandle(event.Event) {
		return &FSMTransitionResult{
			PreviousState: previousState,
			CurrentState:  previousState,
			Event:         event.Event,
			Transitioned:  false,
			Error:         fmt.Errorf("invalid transition: event %s not allowed in state %s", event.Event, previousState),
		}, nil
	}

	// Perform the transition
	if err := f.fsm.Handle(event.Event); err != nil {
		return &FSMTransitionResult{
			PreviousState: previousState,
			CurrentState:  f.fsm.CurrentState(),
			Event:         event.Event,
			Transitioned:  false,
			Error:         err,
		}, err
	}

	return &FSMTransitionResult{
		PreviousState: previousState,
		CurrentState:  f.fsm.CurrentState(),
		Event:         event.Event,
		Transitioned:  previousState != f.fsm.CurrentState(),
		Error:         nil,
	}, nil
}

// CurrentState returns the current FSM state
func (f *FSMGraph) CurrentState() agent.State {
	return f.fsm.CurrentState()
}

// GetHistory returns the state history
func (f *FSMGraph) GetHistory() []agent.StateInfo {
	return f.fsm.GetHistory()
}

// StateNames returns a list of all valid state names for the graph
func StateNames() []string {
	return []string{
		"idle", "searching", "negotiating", "interviewing",
		"offer_received", "hired", "rejected", "paused",
	}
}

// EventNames returns a list of all valid event names
func EventNames() []string {
	return []string{
		"START_SEARCH", "MATCH_FOUND", "INTEREST_EXPRESSED", "NEGOTIATE",
		"SCHEDULE_INTERVIEW", "INTERVIEW_COMPLETE", "OFFER_RECEIVED",
		"OFFER_ACCEPTED", "OFFER_DECLINED", "PAUSE", "RESUME",
		"TIMEOUT", "REJECTED",
	}
}
