package service

import (
	"log"
	"time"

	"github.com/google/uuid"
	goproto "google.golang.org/protobuf/proto"

	"joblinker/internal/agent"
	"joblinker/internal/config"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/proto"
)

// FSMIntegration maps intents to FSM events and handles state transitions
type FSMIntegration struct {
	matchRepo         *repository.MatchRepository
	wsSerializer      *proto.WebSocketFrameSerializer
	// WebSocket broadcast channel for state change events
	broadcast chan *proto.StateChangeEvent
}

// NewFSMIntegration creates a new FSM integration service
func NewFSMIntegration(matchRepo *repository.MatchRepository) *FSMIntegration {
	s := &FSMIntegration{
		matchRepo:     matchRepo,
		wsSerializer: proto.NewWebSocketFrameSerializer(),
		broadcast:     make(chan *proto.StateChangeEvent, 100),
	}
	// Start broadcast goroutine
	go s.broadcastLoop()
	return s
}

// broadcastLoop handles broadcasting state change events to subscribers
func (s *FSMIntegration) broadcastLoop() {
	for event := range s.broadcast {
		s.publishStateChange(event)
	}
}

// publishStateChange sends a StateChangeEvent to WebSocket subscribers
func (s *FSMIntegration) publishStateChange(event *proto.StateChangeEvent) {
	if event == nil {
		return
	}

	// Serialize as Protobuf if enabled
	if config.IsWebSocketProtobufEnabled() {
		frame := &proto.WebSocketFrame{
			MessageType:   proto.MessageType_STATE_CHANGE,
			SequenceNum:   uint64(time.Now().UnixNano()),
			Timestamp:     time.Now().UnixMilli(),
			SchemaVersion: 1,
		}
		// Embed StateChangeEvent in payload
		data, err := goproto.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal StateChangeEvent: %v", err)
			return
		}
		frame.Payload = data

		frameBytes, err := s.wsSerializer.MarshalFrame(frame)
		if err != nil {
			log.Printf("Failed to marshal WebSocketFrame: %v", err)
			return
		}
		log.Printf("Broadcasting StateChangeEvent as Protobuf: match_id=%s old=%s new=%s",
			event.MatchId, event.OldState, event.NewState)
		// Would send to WebSocket connection manager here
		_ = frameBytes
	} else {
		// JSON fallback
		log.Printf("Broadcasting StateChangeEvent as JSON: match_id=%s old=%s new=%s",
			event.MatchId, event.OldState, event.NewState)
	}
}

// BroadcastStateChange broadcasts a state change event to all subscribers
func (s *FSMIntegration) BroadcastStateChange(matchID string, oldState, newState agent.State, reason string) {
	event := &proto.StateChangeEvent{
		MatchId:   matchID,
		OldState:  proto.AgentState(agentStateToProto(oldState)),
		NewState:  proto.AgentState(agentStateToProto(newState)),
		Reason:    reason,
		Timestamp: time.Now().UnixMilli(),
	}
	select {
	case s.broadcast <- event:
	default:
		log.Printf("State change broadcast channel full, dropping event for match %s", matchID)
	}
}

// MatchStatus <-> FSM State mappings
var matchStatusToFSMState = map[model.MatchStatus]agent.State{
	model.MatchStatusPending:       agent.StateIdle,
	model.MatchStatusSearching:     agent.StateSearching,
	model.MatchStatusMutualInterest: agent.StateSearching,
	model.MatchStatusNegotiating:   agent.StateNegotiating,
	model.MatchStatusInterviewing:  agent.StateInterviewing,
	model.MatchStatusOffered:       agent.StateOfferReceived,
	model.MatchStatusHired:         agent.StateHired,
	model.MatchStatusRejected:      agent.StateRejected,
	model.MatchStatusPaused:        agent.StatePaused,
}

var fsmStateToMatchStatus = map[agent.State]model.MatchStatus{
	agent.StateIdle:          model.MatchStatusPending,
	agent.StateSearching:     model.MatchStatusMutualInterest,
	agent.StateNegotiating:   model.MatchStatusNegotiating,
	agent.StateInterviewing:  model.MatchStatusInterviewing,
	agent.StateOfferReceived: model.MatchStatusOffered,
	agent.StateHired:         model.MatchStatusHired,
	agent.StateRejected:      model.MatchStatusRejected,
	agent.StatePaused:        model.MatchStatusPaused,
}

// IntentToEvent maps AI intent strings to FSM events
var IntentToEvent = map[string]agent.Event{
	"INQUIRY":             agent.EventStartSearch,
	"INTRODUCTION":        agent.EventStartSearch,
	"INTEREST_EXPRESSED": agent.EventInterestExpressed,
	"INTEREST":           agent.EventInterestExpressed,
	"NEGOTIATION":         agent.EventNegotiate,
	"NEGOTIATE":          agent.EventNegotiate,
	"SCHEDULE":           agent.EventScheduleInterview,
	"SCHEDULE_INTERVIEW": agent.EventScheduleInterview,
	"OFFER":              agent.EventOfferReceived,
	"OFFER_CREATED":      agent.EventOfferReceived,
	"CONFIRM":            agent.EventOfferAccepted,
	"CONFIRMATION":       agent.EventOfferAccepted,
	"INTERVIEW_COMPLETE": agent.EventInterviewComplete,
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

	// Map MatchStatus -> FSM state
	fsmState, ok := matchStatusToFSMState[match.Status]
	if !ok {
		fsmState = agent.StateIdle
	}
	fsm.SetState(fsmState)
	oldFSMState := fsmState

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

	newFSMState := fsm.CurrentState()

	// Map FSM state -> MatchStatus
	newStatus, ok := fsmStateToMatchStatus[newFSMState]
	if !ok {
		newStatus = match.Status // keep current if no mapping
	}

	// Persist state change
	if err := s.matchRepo.UpdateStatus(matchID, newStatus); err != nil {
		return newFSMState, false, err
	}

	// Broadcast state change if transition occurred
	if newFSMState != oldFSMState {
		s.BroadcastStateChange(matchID.String(), oldFSMState, newFSMState, intent)
		log.Printf("FSM: match=%s %s -> %s (via %s)", matchID, oldFSMState, newFSMState, intent)
	}

	return newFSMState, true, nil
}

// GetCurrentState returns the current FSM state for a match
func (s *FSMIntegration) GetCurrentState(matchID uuid.UUID) (agent.State, error) {
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return agent.StateIdle, err
	}
	if match.Status == "" {
		return agent.StateIdle, nil
	}
	return agent.State(match.Status), nil
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

// agentStateToProto converts agent state to proto enum
func agentStateToProto(s agent.State) int32 {
	switch s {
	case agent.StateIdle:
		return 1 // AgentState_IDLE = 1
	case agent.StateSearching:
		return 2 // AgentState_SEARCHING = 2
	case agent.StateNegotiating:
		return 3 // AgentState_NEGOTIATING = 3
	case agent.StateInterviewing:
		return 4 // AgentState_INTERVIEWING = 4
	case agent.StateOfferReceived:
		return 5 // AgentState_OFFER_RECEIVED = 5
	case agent.StateHired:
		return 6 // AgentState_HIRED = 6
	case agent.StateRejected:
		return 7 // AgentState_REJECTED = 7
	default:
		return 0 // AgentState_AGENT_STATE_UNSPECIFIED = 0
	}
}