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

// AgentProposal is the structured JSON format for agent responses
type AgentProposal struct {
	Intent   string                 `json:"intent"`   // "proposal", "counter", "accept", "decline", "inquire"
	Topic    string                 `json:"topic"`    // "salary", "start_date", "position", "general"
	Message  string                 `json:"message"`
	Terms    map[string]interface{} `json:"terms,omitempty"`
	Accepted bool                   `json:"accepted,omitempty"`
}

// NegotiationState persists the current state of a negotiation
type NegotiationState struct {
	MatchID      uuid.UUID                `json:"match_id"`
	CurrentRound int                      `json:"current_round"`
	Status       string                   `json:"status"` // "active", "agreed", "failed"
	LastTerms    map[string]interface{}   `json:"last_terms,omitempty"`
}

// InitiateNegotiation starts the autonomous dual-agent negotiation
func (s *DualAgentNegotiationService) InitiateNegotiation(ctx context.Context, matchID uuid.UUID) error {
	log.Printf("Initiating dual-agent negotiation for match %s", matchID)

	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	seekerAgent, err := s.agentRepo.GetByID(match.SeekerAgentID)
	if err != nil {
		return fmt.Errorf("failed to get seeker agent: %w", err)
	}

	go s.runNegotiationLoop(ctx, match, seekerAgent)
	return nil
}

func (s *DualAgentNegotiationService) runNegotiationLoop(ctx context.Context, match *model.Match, seekerAgent *model.Agent) {
	maxRounds := 20
	state := &NegotiationState{
		MatchID: match.ID,
		Status:  "active",
	}

	// Get the recruiter agent (from the job owner)
	var recruiterAgent *model.Agent
	var job model.Job
	if err := s.db.First(&job, "id = ?", match.JobID).Error; err == nil {
		recruiterAgent, _ = s.agentRepo.GetByID(job.AgentID)
	}

	for state.CurrentRound < maxRounds && state.Status == "active" {
		select {
		case <-ctx.Done():
			log.Printf("Negotiation cancelled for match %s: %v", match.ID, ctx.Err())
			state.Status = "failed"
			return
		default:
		}

		state.CurrentRound++

		// Round part A: Seeker proposes
		seekerProposal := s.generateProposal(ctx, match, seekerAgent, state, model.AgentTypeSeeker)
		s.storeAgentMessage(ctx, match.ID, seekerAgent.ID, seekerProposal, "PROPOSAL")

		if s.isAgreement(seekerProposal) {
			state.Status = "agreed"
			state.LastTerms = extractTerms(seekerProposal)
			s.matchRepo.UpdateStatus(match.ID, model.MatchStatusHired)
			log.Printf("Negotiation agreed for match %s after %d rounds", match.ID, state.CurrentRound)
			break
		}

		if s.isDecline(seekerProposal) {
			state.Status = "failed"
			s.matchRepo.UpdateStatus(match.ID, model.MatchStatusRejected)
			log.Printf("Negotiation declined by seeker for match %s after %d rounds", match.ID, state.CurrentRound)
			break
		}

		// Round part B: Recruiter responds
		if recruiterAgent != nil {
			recruiterProposal := s.generateProposal(ctx, match, recruiterAgent, state, model.AgentTypeRecruiter)
			s.storeAgentMessage(ctx, match.ID, recruiterAgent.ID, recruiterProposal, "COUNTER")

			if s.isAgreement(recruiterProposal) {
				state.Status = "agreed"
				state.LastTerms = extractTerms(recruiterProposal)
				s.matchRepo.UpdateStatus(match.ID, model.MatchStatusHired)
				log.Printf("Negotiation agreed for match %s after %d rounds", match.ID, state.CurrentRound)
				break
			}

			if s.isDecline(recruiterProposal) {
				state.Status = "failed"
				s.matchRepo.UpdateStatus(match.ID, model.MatchStatusRejected)
				log.Printf("Negotiation declined by recruiter for match %s after %d rounds", match.ID, state.CurrentRound)
				break
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	if state.CurrentRound >= maxRounds && state.Status == "active" {
		state.Status = "failed"
		s.matchRepo.UpdateStatus(match.ID, model.MatchStatusRejected)
		log.Printf("Negotiation max rounds reached for match %s", match.ID)
	}
}

func (s *DualAgentNegotiationService) generateProposal(ctx context.Context, match *model.Match, agent *model.Agent, state *NegotiationState, agentType model.AgentType) *AgentProposal {
	messages, _ := s.messageRepo.ListByMatchID(match.ID)

	contextBuilder := strings.Builder{}
	for _, m := range messages {
		contextBuilder.WriteString(fmt.Sprintf("[%s] %s\n", m.SenderAgentID, m.ContentXML))
	}

	scenario := model.ScenarioNegotiation
	if state.CurrentRound <= 1 {
		scenario = model.ScenarioGreeting
	}

	prompt := s.promptService.BuildFullPrompt(agentType, scenario, contextBuilder.String())

	roleDesc := "a job seeker"
	if agentType == model.AgentTypeRecruiter {
		roleDesc = "a recruiter"
	}

	negotiationPrompt := fmt.Sprintf(`%s

## Negotiation Round %d
You are acting as %s in a dual-agent negotiation.

## Instructions
1. Analyze the current negotiation state
2. Respond with a STRUCTURED JSON object (not plain text)
3. If terms are acceptable, set "accepted": true
4. Include your specific proposal terms

## Response Format (MUST be valid JSON)
{
  "intent": "proposal|counter|accept|decline|inquire",
  "topic": "salary|start_date|position|general",
  "message": "Your natural language message here",
  "terms": {
    "salary": 120000,
    "start_date": "2026-06-01"
  },
  "accepted": false
}`,
		prompt, state.CurrentRound, roleDesc)

	proposal := &AgentProposal{
		Intent:  "proposal",
		Topic:   "general",
		Message: "I'd like to discuss the compensation package further.",
	}

	if s.aiClient == nil {
		return proposal
	}

	systemPrompt := fmt.Sprintf("You are %s in a salary negotiation. ALWAYS respond with valid JSON.", roleDesc)
	response, err := s.aiClient.Chat(systemPrompt, negotiationPrompt)
	if err != nil || response == "" {
		log.Printf("AI call failed in generateProposal: %v", err)
		return proposal
	}

	// Try to parse as JSON
	parsed, err := parseProposalJSON(response)
	if err == nil && parsed != nil {
		return parsed
	}

	// Fallback: natural language analysis
	proposal.Message = response
	lower := strings.ToLower(response)
	if strings.Contains(lower, "agree") || strings.Contains(lower, "accept") || strings.Contains(lower, "deal") || strings.Contains(lower, "confirmed") {
		proposal.Intent = "accept"
		proposal.Accepted = true
	} else if strings.Contains(lower, "decline") || strings.Contains(lower, "reject") || strings.Contains(lower, "walk") || strings.Contains(lower, "unfortunately") {
		proposal.Intent = "decline"
	} else {
		proposal.Intent = "counter"
	}

	return proposal
}

// parseProposalJSON attempts to parse an AI response as a structured AgentProposal JSON
func parseProposalJSON(response string) (*AgentProposal, error) {
	// Try direct JSON parse first
	var proposal AgentProposal
	if err := json.Unmarshal([]byte(response), &proposal); err == nil {
		return &proposal, nil
	}

	// Try to extract JSON from code fences
	cleaned := response
	if idx := strings.Index(response, "```json"); idx >= 0 {
		end := strings.Index(response[idx+7:], "```")
		if end >= 0 {
			cleaned = strings.TrimSpace(response[idx+7 : idx+7+end])
		}
	} else if idx := strings.Index(response, "```"); idx >= 0 {
		end := strings.Index(response[idx+3:], "```")
		if end >= 0 {
			cleaned = strings.TrimSpace(response[idx+3 : idx+3+end])
		}
	}

	if cleaned != response {
		if err := json.Unmarshal([]byte(cleaned), &proposal); err == nil {
			return &proposal, nil
		}
	}

	return nil, fmt.Errorf("invalid JSON")
}

// extractTerms extracts the terms map from a proposal
func extractTerms(proposal *AgentProposal) map[string]interface{} {
	if proposal.Terms != nil {
		return proposal.Terms
	}
	return map[string]interface{}{}
}

func (s *DualAgentNegotiationService) isAgreement(proposal *AgentProposal) bool {
	if proposal.Accepted {
		return true
	}
	if proposal.Intent == "accept" {
		return true
	}
	return false
}

func (s *DualAgentNegotiationService) isDecline(proposal *AgentProposal) bool {
	return proposal.Intent == "decline"
}

func (s *DualAgentNegotiationService) storeAgentMessage(ctx context.Context, matchID, agentID uuid.UUID, proposal *AgentProposal, intent string) {
	termsJSON, _ := json.Marshal(proposal.Terms)
	contentXML := fmt.Sprintf(`<message><intent>%s</intent><topic>%s</topic><message>%s</message><terms>%s</terms></message>`,
		proposal.Intent, proposal.Topic, proposal.Message, string(termsJSON))

	msg := &model.Message{
		ID:            uuid.New(),
		MatchID:       matchID,
		SenderAgentID: agentID,
		ContentXML:    contentXML,
		IntentType:    intent,
	}

	if err := s.messageRepo.Create(msg); err != nil {
		log.Printf("Failed to store agent message: %v", err)
	}
}

// GetNegotiationStatus returns current negotiation state
func (s *DualAgentNegotiationService) GetNegotiationStatus(matchID uuid.UUID) (*NegotiationState, error) {
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
	}
	return nil
}
