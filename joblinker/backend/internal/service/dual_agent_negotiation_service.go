package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"joblinker/internal/agent"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DualAgentNegotiationService struct {
	db            *gorm.DB
	matchRepo     *repository.MatchRepository
	messageRepo   *repository.MessageRepository
	agentRepo     *repository.AgentRepository
	promptService *AgentPromptService
	toolExecutor  *agent.ToolExecutor
	aiClient      *ai.Client
}

func NewDualAgentNegotiationService(
	db *gorm.DB,
	matchRepo *repository.MatchRepository,
	messageRepo *repository.MessageRepository,
	agentRepo *repository.AgentRepository,
	promptService *AgentPromptService,
	toolExecutor *agent.ToolExecutor,
	aiClient *ai.Client,
) *DualAgentNegotiationService {
	return &DualAgentNegotiationService{
		db:            db,
		matchRepo:     matchRepo,
		messageRepo:   messageRepo,
		agentRepo:     agentRepo,
		promptService: promptService,
		toolExecutor:  toolExecutor,
		aiClient:      aiClient,
	}
}

type NegotiationRound struct {
	RoundNumber       int
	SeekerMessage     string
	RecruiterMessage  string
	AgreementReached  bool
	Topic             string // "position", "salary", "start_date", etc.
}

type NegotiationState struct {
	MatchID      uuid.UUID
	CurrentRound int
	Status       string // "active", "agreed", "failed"
	LastProposal map[string]interface{}
}

// InitiateNegotiation starts the autonomous dual-agent negotiation
func (s *DualAgentNegotiationService) InitiateNegotiation(ctx context.Context, matchID uuid.UUID) error {
	log.Printf("Initiating dual-agent negotiation for match %s", matchID)

	// Get match details
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	// Get seeker agent
	seekerAgent, err := s.agentRepo.GetByID(match.SeekerAgentID)
	if err != nil {
		return fmt.Errorf("failed to get seeker agent: %w", err)
	}

	// Start autonomous negotiation loop
	go s.runNegotiationLoop(context.Background(), match, seekerAgent, s.promptService)

	return nil
}

func (s *DualAgentNegotiationService) runNegotiationLoop(ctx context.Context, match *model.Match, seekerAgent *model.Agent, promptService *AgentPromptService) {
	maxRounds := 20
	state := &NegotiationState{
		MatchID: match.ID,
		Status:  "active",
	}

	for state.CurrentRound < maxRounds && state.Status == "active" {
		state.CurrentRound++

		// Round: Seeker proposes
		seekerProposal := s.generateSeekerProposal(ctx, match, seekerAgent, state.CurrentRound, promptService)
		s.storeAgentMessage(ctx, match.ID, seekerAgent.ID, seekerProposal, "PROPOSAL")

		// Check if agreement reached
		if s.isAgreement(seekerProposal) {
			state.Status = "agreed"
			s.requestHumanConfirmation(ctx, match.ID, "offer", state.LastProposal)
			break
		}

		// Small delay to prevent tight loop
		time.Sleep(100 * time.Millisecond)
	}

	if state.CurrentRound >= maxRounds && state.Status == "active" {
		state.Status = "failed"
		log.Printf("Negotiation failed after %d rounds for match %s", maxRounds, match.ID)
	}
}

func (s *DualAgentNegotiationService) generateSeekerProposal(ctx context.Context, match *model.Match, agent *model.Agent, round int, promptService *AgentPromptService) string {
	// Get conversation context
	messages, _ := s.messageRepo.ListByMatchID(match.ID)

	// Build context
	var contextBuilder strings.Builder
	for _, m := range messages {
		contextBuilder.WriteString(fmt.Sprintf("[%s] %s\n", m.SenderAgentID, m.ContentXML))
	}

	// Get prompt for scenario
	scenario := model.ScenarioNegotiation
	if round == 1 {
		scenario = model.ScenarioGreeting
	}

	prompt := promptService.BuildFullPrompt(model.AgentTypeSeeker, scenario, contextBuilder.String())

	// Add negotiation context
	prompt = fmt.Sprintf(`%s

## Negotiation Round %d
You are acting as a job seeker in a dual-agent negotiation.

## Instructions
1. Analyze the current negotiation state
2. Make a reasonable proposal based on your interests
3. If terms are acceptable, indicate AGREEMENT
4. If not acceptable, propose counter-offer

## Response Format
Respond with your proposal text. If agreeing to terms, say "I agree to these terms."`,
		prompt, round)

	// Call AI service if available
	if s.aiClient != nil {
		systemPrompt := "You are a job seeker agent in a salary negotiation. Respond naturally based on the conversation context."
		response, err := s.aiClient.Chat(systemPrompt, prompt)
		if err == nil && response != "" {
			return response
		}
		log.Printf("AI call failed in generateSeekerProposal: %v", err)
	}

	return "I'd like to discuss the compensation package further."
}

func (s *DualAgentNegotiationService) storeAgentMessage(ctx context.Context, matchID, agentID uuid.UUID, content, intent string) {
	msg := &model.Message{
		ID:            uuid.New(),
		MatchID:       matchID,
		SenderAgentID: agentID,
		ContentXML:    content,
		IntentType:    intent,
	}

	if err := s.messageRepo.Create(msg); err != nil {
		log.Printf("Failed to store agent message: %v", err)
	}
}

func (s *DualAgentNegotiationService) isAgreement(proposal string) bool {
	return strings.Contains(strings.ToLower(proposal), "agree") &&
		strings.Contains(strings.ToLower(proposal), "terms")
}

// requestHumanConfirmation creates a confirmation request when agents reach agreement
func (s *DualAgentNegotiationService) requestHumanConfirmation(ctx context.Context, matchID uuid.UUID, confirmationType string, payload map[string]interface{}) {
	payloadJSON, _ := json.Marshal(payload)

	// Get user ID from match
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		log.Printf("Failed to get match for confirmation: %v", err)
		return
	}

	confReq := &model.ConfirmationRequest{
		ID:     uuid.New(),
		MatchID: matchID,
		Type:   model.ConfirmationRequestType(confirmationType),
		Payload: string(payloadJSON),
		Status: model.ConfirmationStatusPending,
		UserID: match.SeekerAgentID, // Would be SeekerUserID in full model
	}

	log.Printf("Human confirmation requested for match %s: type=%s, payload=%s", matchID, confirmationType, confReq.Payload)
}

// GetNegotiationStatus returns current negotiation state
func (s *DualAgentNegotiationService) GetNegotiationStatus(matchID uuid.UUID) (*NegotiationState, error) {
	// Would query database for actual state
	return &NegotiationState{
		MatchID: matchID,
		Status:  "active",
	}, nil
}

// ProcessHumanResponse handles human approval/rejection
func (s *DualAgentNegotiationService) ProcessHumanResponse(ctx context.Context, confirmationID uuid.UUID, approved bool, feedback string) error {
	if approved {
		log.Printf("Human approved confirmation %s", confirmationID)
	} else {
		log.Printf("Human rejected confirmation %s: %s", confirmationID, feedback)
		// Trigger renegotiation based on feedback
	}

	return nil
}

// NegotiationMessage represents a message in the negotiation
type NegotiationMessage struct {
	From     string `json:"from"` // "seeker" or "recruiter"
	To       string `json:"to"`   // "seeker" or "recruiter"
	Content  string `json:"content"`
	Round    int    `json:"round"`
	Topic    string `json:"topic"`
	Proposal map[string]interface{} `json:"proposal,omitempty"`
}