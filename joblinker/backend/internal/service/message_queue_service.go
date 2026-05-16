package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"joblinker/internal/eino/chatmodel"
	"joblinker/internal/eino/runner"
	"joblinker/internal/eino/prompt"
	"joblinker/internal/agent"
	"joblinker/internal/cache"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"
	"joblinker/pkg/rabbitmq"

	"github.com/google/uuid"
)

type BroadcastFunc func(matchID string, eventType string, payload interface{})

type MessageQueueService struct {
	rmq                *rabbitmq.RabbitMQ
	messageRepo        *repository.MessageRepository
	matchRepo          *repository.MatchRepository
	agentRepo          *repository.AgentRepository
	jobRepo            *repository.JobRepository
	offerRepo          *repository.OfferRepository
	interviewRepo      *repository.InterviewRepository
	aiClient           *ai.Client
	einoChatModel      *chatmodel.EinoChatModel
	promptService      *AgentPromptService
	toolExecutor       *agent.ToolExecutor
	toolCache          *cache.ToolCache
	fsmIntegration     *FSMIntegration
	// Eino Agent Runner (optional, for new Eino-based processing)
	einoRunner         *runner.AgentRunner
	// ADK Runner (optional, for ADK-based processing)
	adkRunner          *runner.ADKRunner
	// Context optimization for AI prompts
	contextOptimizer   *ContextOptimizerService
	// Observability
	metricsSvc         *MetricsService
	auditSvc           *AuditService
	alertSvc           *AlertingService
	// Bidirectional A2A tracking
	conversationRounds map[string]int
	conversationMu     sync.RWMutex
	// WebSocket broadcast callback
	broadcastFn        BroadcastFunc
}

// SetEinoRunner sets the Eino AgentRunner for Eino-based processing
func (s *MessageQueueService) SetEinoRunner(einoRunner *runner.AgentRunner) {
	s.einoRunner = einoRunner
	log.Printf("MessageQueueService: Eino Runner configured")
}

// SetADKRunner sets the ADK Runner for ADK-based agent processing
func (s *MessageQueueService) SetADKRunner(adkRunner *runner.ADKRunner) {
	s.adkRunner = adkRunner
	log.Printf("MessageQueueService: ADK Runner configured")
}

// SetContextOptimizer sets the context optimizer for AI prompts
func (s *MessageQueueService) SetContextOptimizer(ctxOptimizer *ContextOptimizerService) {
	s.contextOptimizer = ctxOptimizer
	log.Printf("MessageQueueService: Context Optimizer configured")
}

// SetBroadcastCallback sets the WebSocket broadcast callback
func (s *MessageQueueService) SetBroadcastCallback(fn BroadcastFunc) {
	s.broadcastFn = fn
	log.Printf("MessageQueueService: Broadcast callback configured")
}

func NewMessageQueueService(
	rmq *rabbitmq.RabbitMQ,
	messageRepo *repository.MessageRepository,
	matchRepo *repository.MatchRepository,
	agentRepo *repository.AgentRepository,
	jobRepo *repository.JobRepository,
	offerRepo *repository.OfferRepository,
	interviewRepo *repository.InterviewRepository,
	metricsRepo *repository.AgentMetricsRepository,
	auditRepo *repository.AuditLogRepository,
	errorEventRepo *repository.ErrorEventRepository,
	toolCache *cache.ToolCache,
) *MessageQueueService {
	aiClient := ai.NewClient()
	einoChatModel := chatmodel.NewEinoChatModel(aiClient)
	promptService := NewAgentPromptService()
	metricsSvc := NewMetricsService(metricsRepo)
	auditSvc := NewAuditService(auditRepo)
	alertSvc := NewAlertingService(errorEventRepo, "")
	fsmIntegration := NewFSMIntegration(matchRepo)
	toolExecutor := agent.NewToolExecutor(jobRepo, agentRepo, matchRepo, offerRepo, interviewRepo)
	toolExecutor.SetErrorHandler(func(errorType, msg string, ctx map[string]interface{}) {
		alertSvc.RecordError(errorType, msg, "", ctx)
	})
	if toolCache != nil {
		toolExecutor.SetCache(toolCache)
	}
	log.Printf("MessageQueueService initialized with AI client")
	return &MessageQueueService{
		rmq:                rmq,
		messageRepo:        messageRepo,
		matchRepo:          matchRepo,
		agentRepo:          agentRepo,
		jobRepo:            jobRepo,
		offerRepo:          offerRepo,
		interviewRepo:      interviewRepo,
		aiClient:           aiClient,
		einoChatModel:      einoChatModel,
		promptService:      promptService,
		toolExecutor:       toolExecutor,
		toolCache:          toolCache,
		fsmIntegration:     fsmIntegration,
		metricsSvc:         metricsSvc,
		auditSvc:           auditSvc,
		alertSvc:           alertSvc,
		conversationRounds: make(map[string]int),
	}
}

// StartConsuming starts the message queue consumer
func (s *MessageQueueService) StartConsuming(ctx context.Context) error {
	if s.rmq == nil {
		log.Println("RabbitMQ not configured, skipping consumer")
		return nil
	}

	return s.rmq.Consume(ctx, s.handleAgentMessage)
}

func (s *MessageQueueService) handleAgentMessage(msg *rabbitmq.AgentMessage) error {
	log.Printf("Received agent message: %s -> %s (intent: %s)", msg.SenderID, msg.ReceiverID, msg.Intent)

	matchID, err := uuid.Parse(msg.MatchID)
	if err != nil {
		log.Printf("Invalid match ID %s: %v", msg.MatchID, err)
		return nil
	}

	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		log.Printf("Match %s not found: %v", matchID, err)
		return nil
	}

	senderAgent, err := s.agentRepo.GetByID(uuid.MustParse(msg.SenderID))
	if err != nil {
		log.Printf("Sender agent %s not found: %v", msg.SenderID, err)
		return nil
	}

	var fsmState interface{}
	if s.fsmIntegration != nil {
		newState, transitioned, fsmErr := s.fsmIntegration.TransitionFSM(matchID, msg.Intent)
		if fsmErr != nil {
			log.Printf("FSM transition error: %v", fsmErr)
		} else if transitioned {
			log.Printf("FSM state changed: match=%s intent=%s -> state=%s", matchID, msg.Intent, newState)
			fsmState = newState
		}
	}

	roundKey := matchID.String()
	s.conversationMu.Lock()
	s.conversationRounds[roundKey]++
	currentRound := s.conversationRounds[roundKey]
	s.conversationMu.Unlock()
	log.Printf("Conversation round: match=%s round=%d", matchID, currentRound)

	if currentRound > 10 {
		log.Printf("Conversation max rounds reached for match %s, pausing", matchID)
		return nil
	}

	if s.metricsSvc != nil {
		s.metricsSvc.UpdateHeartbeat(senderAgent.ID)
		s.metricsSvc.RecordMessage(senderAgent.ID)
	}
	if s.auditSvc != nil {
		s.auditSvc.LogEvent(context.Background(), senderAgent.ID, matchID, "message_sent", map[string]interface{}{
			"intent":         msg.Intent,
			"content_length": len(msg.MessageID),
		})
	}

	response := s.generateEinoResponse(msg, match, senderAgent)
	if response == nil {
		log.Printf("DeepRecruiter returned nil, falling back to legacy AI")
		response = s.generateAutoResponse(msg, match, senderAgent)
	}

	responseAgentID := s.getOtherAgentID(match, senderAgent.ID)

	// FSM transition for the response intent (drives progression)
	if s.fsmIntegration != nil && response.Intent != msg.Intent {
		newState, transitioned, fsmErr := s.fsmIntegration.TransitionFSM(matchID, response.Intent)
		if fsmErr != nil {
			log.Printf("FSM transition error (response): %v", fsmErr)
		} else if transitioned {
			log.Printf("FSM state changed (response): match=%s intent=%s -> state=%s", matchID, response.Intent, newState)
			fsmState = newState
		}
	}

	responseMsg := &model.Message{
		ID:            uuid.New(),
		MatchID:       matchID,
		SenderAgentID: responseAgentID,
		ContentXML:    s.responseToXML(response),
		IntentType:    response.Intent,
	}
	if err := s.messageRepo.Create(responseMsg); err != nil {
		log.Printf("Failed to store response message: %v", err)
	}

	if s.broadcastFn != nil {
		s.broadcastFn(matchID.String(), "agent_response", map[string]interface{}{
			"message_id":     responseMsg.ID.String(),
			"match_id":       matchID.String(),
			"sender_agent_id": responseAgentID.String(),
			"content_xml":    responseMsg.ContentXML,
			"intent_type":    response.Intent,
			"created_at":     responseMsg.CreatedAt,
		})
		if fsmState != nil {
			s.broadcastFn(matchID.String(), "fsm_state_change", map[string]interface{}{
				"match_id":  matchID.String(),
				"new_state": fsmState,
			})
		}
	}

	// Check if human confirmation is required (Offer / Schedule)
	if response.Intent == "OFFER" || response.Intent == "SCHEDULE" {
		s.createConfirmationRequest(matchID, response.Intent, response.Payload, responseAgentID)
		if s.broadcastFn != nil {
			s.broadcastFn(matchID.String(), "confirmation_needed", map[string]interface{}{
				"match_id": matchID.String(),
				"intent":   response.Intent,
			})
		}
		log.Printf("Human confirmation required for match %s (intent: %s), pausing", matchID, response.Intent)
		return nil
	}

	if currentRound < 10 {
		nextMsg := &rabbitmq.AgentMessage{
			MessageID:  responseMsg.ID.String(),
			SenderID:   responseAgentID.String(),
			ReceiverID: senderAgent.ID.String(),
			Intent:     response.Intent,
			MatchID:    msg.MatchID,
			Payload:    response.Payload,
			Timestamp:  time.Now(),
		}
		if s.rmq != nil {
			ctxPub, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.rmq.PublishAgentMessage(ctxPub, nextMsg); err != nil {
				log.Printf("Autonomous A2A: failed to publish next round: %v", err)
			} else {
				log.Printf("Autonomous A2A: round %d/%d published for match %s", currentRound, 10, matchID)
			}
		}
	} else {
		log.Printf("Autonomous A2A: max rounds reached for match %s", matchID)
	}

	log.Printf("Auto-response: %s -> %s (intent: %s)", response.Intent, msg.SenderID, msg.Intent)
	return nil
}

// getOtherAgentID returns the ID of the OTHER agent in the match (for A2A alternation)
func (s *MessageQueueService) getOtherAgentID(match *model.Match, senderID uuid.UUID) uuid.UUID {
	// If sender is the seeker, return the recruiter agent (from job)
	if match.SeekerAgentID == senderID {
		job, err := s.jobRepo.GetByID(match.JobID)
		if err == nil && job != nil {
			return job.AgentID
		}
	}
	// Otherwise return the seeker agent
	return match.SeekerAgentID
}

func (s *MessageQueueService) generateAutoResponse(msg *rabbitmq.AgentMessage, match *model.Match, senderAgent *model.Agent) *AutoResponse {
	// Determine scenario type from intent
	scenario := intentToScenario(msg.Intent)

	// Get agent type
	var agentType model.AgentType
	if senderAgent.Type == "seeker" {
		agentType = model.AgentTypeSeeker
	} else {
		agentType = model.AgentTypeRecruiter
	}

	// Try Eino-based agent processing first if runner is configured
	if s.einoRunner != nil {
		einoResponse := s.generateEinoResponse(msg, match, senderAgent)
		if einoResponse != nil {
			log.Printf("Eino agent generated response: intent=%s, response_len=%d", einoResponse.Intent, len(fmt.Sprint(einoResponse.Payload)))
			return einoResponse
		}
		log.Printf("Eino agent returned nil, falling back to legacy AI")
	}

	// Get conversation context for better responses
	conversationContext := s.buildAgentContext(msg, match)

	// Build the full three-part prompt
	fullPrompt := s.promptService.BuildFullPrompt(agentType, scenario, conversationContext)

	// Get job and agent details for context
	var jobTitle, jobLocation string
	var salaryMin, salaryMax int
	var jobSkills []string

	if job, err := s.jobRepo.GetByID(match.JobID); err == nil {
		// Parse structured job data
		var jobData map[string]interface{}
		if json.Unmarshal(job.StructuredJSON, &jobData) == nil {
			if title, ok := jobData["title"].(string); ok {
				jobTitle = title
			}
			if location, ok := jobData["location"].(string); ok {
				jobLocation = location
			}
			if min, ok := jobData["salary_min"].(float64); ok {
				salaryMin = int(min)
			}
			if max, ok := jobData["salary_max"].(float64); ok {
				salaryMax = int(max)
			}
			if skills, ok := jobData["skills"].([]interface{}); ok {
				for _, skill := range skills {
					if sk, ok := skill.(string); ok {
						jobSkills = append(jobSkills, sk)
					}
				}
			}
		}
	}

	var seekerSkills []string
	if seeker, err := s.agentRepo.GetByID(match.SeekerAgentID); err == nil {
		var config map[string]interface{}
		if json.Unmarshal(seeker.ConfigJSON, &config) == nil {
			if skills, ok := config["skills"].([]interface{}); ok {
				for _, skill := range skills {
					if sk, ok := skill.(string); ok {
						seekerSkills = append(seekerSkills, sk)
					}
				}
			}
		}
	}

	// Enhance prompt with job and candidate details
	enhancedPrompt := fmt.Sprintf(`%s

## Current Job Details
- Position: %s
- Location: %s
- Salary Range: $%d - $%d
- Required Skills: %v
- Match ID: %s

## Candidate Profile
- Skills: %v

## Your Task
When asked to schedule an interview, create offers, or query jobs, you MUST use the available tools.
Do NOT make up job details, salary numbers, or candidate information - always use the tools.

## Available Tools
- schedule_interview: Schedule an interview (requires match_id, datetime, interview_type)
- create_offer: Create a job offer (requires match_id, salary, start_date)
- query_jobs: Search for jobs (requires location, optional skills, salary_min)
- get_candidate: Get candidate details (requires candidate_id)
- search_candidates: Search for candidates (requires skills, optional location)

## Response Format
When NOT using tools, respond with ONLY a valid JSON object:
{"intent":"ONE_OF:[INTRODUCTION,INTEREST,NEGOTIATION,OFFER,CONFIRM,SCHEDULE,INQUIRY]","message":"Your professional response message here","data":{"title":"Job title if relevant","location":"Job location if relevant","salary_min":number,"salary_max":number,"skills":["skill1","skill2"],"message":"Short message for display"}}

## Intent Progression Rules
- INQUIRY -> respond with INTRODUCTION, include job details (title, salary, location, skills)
- INTRODUCTION -> respond with INTEREST, acknowledge candidate background
- INTEREST -> respond with NEGOTIATION, discuss salary/benefits if appropriate
- NEGOTIATION -> respond with OFFER, present formal offer terms
- OFFER -> respond with CONFIRM or DECLINE
- SCHEDULE -> respond with CONFIRM with interview details`,
		fullPrompt,
		jobTitle,
		jobLocation,
		salaryMin,
		salaryMax,
		jobSkills,
		match.ID,
		seekerSkills)

	// Call AI for intent recognition and response using ChatWithTools for function calling
	response, toolCall, err := s.aiClient.ChatWithTools(
		"You are a professional AI recruitment agent. Use tools when needed to get real data.",
		enhancedPrompt,
		s.toolExecutor.GetTools(),
		s.toolExecutor.ExecuteToolForAI,
	)
	if err != nil {
		log.Printf("AI response failed, using fallback: %v", err)
		return s.fallbackResponse(msg.Intent)
	}

	log.Printf("DEBUG: AI response received, content=%s, tool_called=%v", response, toolCall != "")

	// Parse AI response
	var aiResp struct {
		Intent  string                 `json:"intent"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data,omitempty"`
	}
	if err := json.Unmarshal([]byte(response), &aiResp); err != nil {
		log.Printf("Failed to parse AI response: %v", err)
		return s.fallbackResponse(msg.Intent)
	}

	// Execute tools based on intent (AI may have already called tools via ChatWithTools)
	log.Printf("DEBUG: Checking incoming intent '%s' vs AI response '%s' for tool execution", msg.Intent, aiResp.Intent)

	// Only execute tools if AI didn't call them via ChatWithTools
	// (toolCall == "" means AI didn't request any tools)
	aiCalledTools := toolCall != ""
	if !aiCalledTools {
		// Execute tools based on USER's intent, not AI's response (AI may progress intent)
		// The user sent SCHEDULE, so we execute schedule_interview tool
		if msg.Intent == "SCHEDULE" || msg.Intent == "CONFIRM" || aiResp.Intent == "SCHEDULE" || aiResp.Intent == "CONFIRM" {
			log.Printf("DEBUG: Executing schedule_interview tool for match %s", match.ID)
			// Extract datetime from AI response - no hardcoded dates
			datetime := ""
			if aiResp.Data != nil {
				if dt, ok := aiResp.Data["datetime"].(string); ok {
					datetime = dt
				}
			}
			// Fallback: generate reasonable future datetime if AI didn't provide one
			if datetime == "" {
				datetime = time.Now().AddDate(0, 0, 14).Format(time.RFC3339) // 2 weeks from now
			}

			log.Printf("DEBUG: Calling toolExecutor.ExecuteTool with match_id=%s, datetime=%s", match.ID.String(), datetime)
			// Note: Using context.Background() as generateAutoResponse doesn't receive context
			// For production, propagate context through the call chain
			toolResult, err := s.toolExecutor.ExecuteTool(context.Background(), match.ID, "schedule_interview", map[string]interface{}{
				"match_id": match.ID.String(),
				"datetime": datetime,
				"interview_type": "video",
			})

			log.Printf("DEBUG: toolResult=%+v, err=%v", toolResult, err)

			if err == nil && toolResult.Success {
				log.Printf("Interview scheduled via tool: %+v", toolResult.Data)
				// Add interview_id to response data
				if aiResp.Data == nil {
					aiResp.Data = make(map[string]interface{})
				}
				if toolResult.Data != nil {
					if id, ok := toolResult.Data["interview_id"]; ok {
						aiResp.Data["interview_id"] = id
					}
				}
			}
		}

		if aiResp.Intent == "OFFER" {
			// Create offer - extract values from AI response, not hardcoded
			salary := 0
			startDate := ""
			if aiResp.Data != nil {
				if s, ok := aiResp.Data["salary"].(float64); ok {
					salary = int(s)
				} else if s, ok := aiResp.Data["salary_max"].(float64); ok {
					salary = int(s)
				}
				if sd, ok := aiResp.Data["start_date"].(string); ok {
					startDate = sd
				}
			}
			// Fallback: get from job's salary range if AI didn't provide
			if salary == 0 {
				if job, err := s.jobRepo.GetByID(match.JobID); err == nil {
					var jobData map[string]interface{}
					json.Unmarshal(job.StructuredJSON, &jobData)
					if min, ok := jobData["salary_min"].(float64); ok {
						if max, ok := jobData["salary_max"].(float64); ok && max > min {
							salary = int((min + max) / 2)
						}
					}
				}
				if salary == 0 {
					log.Printf("WARNING: salary could not be determined from job data, using fallback")
					return s.fallbackResponse("OFFER")
				}
			}
			if startDate == "" {
				startDate = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
			}

			toolResult, err := s.toolExecutor.ExecuteTool(context.Background(), match.ID, "create_offer", map[string]interface{}{
				"match_id": match.ID.String(),
				"salary": salary,
				"start_date": startDate,
			})

			if err == nil && toolResult.Success {
				log.Printf("Offer created via tool: %+v", toolResult.Data)
				if aiResp.Data == nil {
					aiResp.Data = make(map[string]interface{})
				}
				if toolResult.Data != nil {
					if id, ok := toolResult.Data["offer_id"]; ok {
						aiResp.Data["offer_id"] = id
					}
				}
			}
		}
	} else {
		log.Printf("DEBUG: AI called tools via ChatWithTools, skipping manual tool execution")
	}

	return &AutoResponse{
		Intent:  aiResp.Intent,
		Payload: aiResp.Data,
	}
}

func (s *MessageQueueService) buildAgentContext(msg *rabbitmq.AgentMessage, match *model.Match) string {
	// Use context optimizer if available
	if s.contextOptimizer != nil {
		optimized := s.contextOptimizer.GetContextForMatch(match.ID)
		if optimized != "" {
			return optimized
		}
	}

	// Fallback to simple context building
	var context string

	// Get recent messages for context
	messages, _ := s.messageRepo.ListByMatchID(match.ID)
	if len(messages) > 0 {
		context = fmt.Sprintf("Recent conversation (%d messages):\n", len(messages))
		start := 0
		if len(messages) > 5 {
			start = len(messages) - 5
		}
		for _, m := range messages[start:] {
			context += fmt.Sprintf("- [%s] %s: %s\n", m.IntentType, m.SenderAgentID, m.ContentXML)
		}
	}

	return context
}

// generateEinoResponse uses DeepRecruiter to generate response
// Returns nil if Eino is not configured or fails, triggering fallback
func (s *MessageQueueService) generateEinoResponse(msg *rabbitmq.AgentMessage, match *model.Match, senderAgent *model.Agent) *AutoResponse {
	if s.einoRunner == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var msgContent string
	if msg.Payload != nil {
		if content, ok := msg.Payload["content"].(string); ok {
			msgContent = content
		}
	}
	if msgContent == "" {
		msgContent = msg.Intent
	}

	deepRecruiter := s.einoRunner.GetDeepRecruiter(match.ID)

	var response string
	var err error

	if senderAgent.Type == "seeker" {
		response, err = deepRecruiter.ProcessRecruiterMessage(ctx, msgContent)
	} else {
		response, err = deepRecruiter.ProcessSeekerMessage(ctx, msgContent)
	}

	if err != nil {
		log.Printf("DeepRecruiter error: %v", err)
		return nil
	}

	intent := s.parseIntent(response, msg.Intent)

	return &AutoResponse{
		Intent:  intent,
		Payload: map[string]interface{}{"message": response},
	}
}

// parseIntent tries to extract intent from AI response JSON, falls back to advancing
func (s *MessageQueueService) parseIntent(response string, currentIntent string) string {
	var parsed struct {
		Intent string `json:"intent"`
	}
	if json.Unmarshal([]byte(response), &parsed) == nil && parsed.Intent != "" {
		return parsed.Intent
	}
	return s.advanceIntent(currentIntent)
}

// createConfirmationRequest persists a human confirmation request
func (s *MessageQueueService) createConfirmationRequest(matchID uuid.UUID, intent string, payload map[string]interface{}, agentID uuid.UUID) {
	var cType model.ConfirmationRequestType
	switch intent {
	case "OFFER":
		cType = model.ConfirmationTypeOffer
	case "SCHEDULE":
		cType = model.ConfirmationTypeInterview
	default:
		cType = model.ConfirmationTypeOffer
	}

	payloadStr := "{}"
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			payloadStr = string(b)
		}
	}

	cr := &model.ConfirmationRequest{
		ID:        uuid.New(),
		MatchID:   matchID,
		Type:      cType,
		Payload:   payloadStr,
		Status:    model.ConfirmationStatusPending,
		UserID:    agentID,
		CreatedAt: time.Now(),
	}
	if err := s.matchRepo.CreateConfirmationRequest(cr); err != nil {
		log.Printf("Failed to create confirmation request: %v", err)
	}
}

// HandleHumanConfirm processes a human confirmation response
func (s *MessageQueueService) HandleHumanConfirm(matchID uuid.UUID, approved bool, feedback string) error {
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return fmt.Errorf("match not found: %w", err)
	}

	// Find pending confirmation request
	cr, err := s.matchRepo.GetPendingConfirmation(matchID)
	if err != nil {
		return fmt.Errorf("no pending confirmation: %w", err)
	}

	now := time.Now()
	if approved {
		cr.Status = model.ConfirmationStatusApproved
		log.Printf("Human APPROVED confirmation for match %s (type: %s)", matchID, cr.Type)
	} else {
		cr.Status = model.ConfirmationStatusRejected
		cr.Feedback = feedback
		match.Status = model.MatchStatusRejected
		if err := s.matchRepo.UpdateStatus(matchID, match.Status); err != nil {
			log.Printf("Failed to update match status: %v", err)
		}
		log.Printf("Human REJECTED confirmation for match %s (type: %s): %s", matchID, cr.Type, feedback)
		return nil
	}
	cr.RespondedAt = &now
	if err := s.matchRepo.UpdateConfirmationRequest(cr); err != nil {
		log.Printf("Failed to update confirmation request: %v", err)
	}

	// Reset conversation round counter so autonomous dialogue can resume
	s.conversationMu.Lock()
	s.conversationRounds[matchID.String()] = 0
	s.conversationMu.Unlock()
	log.Printf("Conversation rounds reset for match %s after human approval", matchID)

	// Update match status: approved offer -> hired
	_ = s.matchRepo.UpdateStatus(matchID, model.MatchStatusHired)

	// Continue the conversation loop
	s.continueAgentConversation(matchID, cr)
	return nil
}

// continueAgentConversation re-publishes to RabbitMQ after human approval
func (s *MessageQueueService) continueAgentConversation(matchID uuid.UUID, cr *model.ConfirmationRequest) {
	if s.rmq == nil {
		log.Printf("RabbitMQ not available, cannot continue conversation")
		return
	}

	nextMsg := &rabbitmq.AgentMessage{
		MessageID:  uuid.New().String(),
		SenderID:   cr.UserID.String(),
		Intent:     "CONFIRM",
		MatchID:    matchID.String(),
		Payload:    map[string]interface{}{"approved": true, "confirmation_type": string(cr.Type)},
		Timestamp:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.rmq.PublishAgentMessage(ctx, nextMsg); err != nil {
		log.Printf("Failed to continue conversation: %v", err)
	} else {
		log.Printf("Conversation continued after human approval for match %s", matchID)
	}
}

// generateWithADK uses the ADK Runner to generate a response
func (s *MessageQueueService) generateWithADK(ctx context.Context, msgContent string, senderAgent *model.Agent) (string, error) {
	if s.adkRunner == nil {
		return "", fmt.Errorf("adk runner not configured")
	}

	// Build query context — include agent identity
	query := msgContent

	events := s.adkRunner.Query(ctx, query)

	// Collect the final response from the event stream
	var response string
	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			mo := event.Output.MessageOutput
			if !mo.IsStreaming && mo.Message != nil {
				response = mo.Message.Content
			}
		}
	}

	if response == "" {
		return "", fmt.Errorf("adk runner: empty response")
	}
	return response, nil
}

// scenarioFromPromptType converts model.PromptScenarioType to prompt.Scenario
func scenarioFromPromptType(scenario model.PromptScenarioType) prompt.Scenario {
	switch scenario {
	case model.ScenarioNegotiation:
		return prompt.ScenarioNegotiation
	case model.ScenarioSalary:
		return prompt.ScenarioSalary
	case model.ScenarioInterview:
		return prompt.ScenarioInterview
	case model.ScenarioOffer:
		return prompt.ScenarioOffer
	case model.ScenarioDecline:
		return prompt.ScenarioDecline
	case model.ScenarioTermination:
		return prompt.ScenarioTermination
	default:
		return prompt.ScenarioGreeting
	}
}

// advanceIntent progresses the conversation to the next logical intent
func (s *MessageQueueService) advanceIntent(currentIntent string) string {
	switch currentIntent {
	case "INQUIRY":
		return "INTRODUCTION"
	case "INTRODUCTION":
		return "INTEREST"
	case "INTEREST":
		return "NEGOTIATION"
	case "NEGOTIATION":
		return "OFFER"
	case "OFFER", "SCHEDULE":
		return "CONFIRM"
	default:
		return "INTRODUCTION"
	}
}

func (s *MessageQueueService) fallbackResponse(currentIntent string) *AutoResponse {
	nextIntent := s.advanceIntent(currentIntent)
	switch nextIntent {
	case "INTRODUCTION":
		return &AutoResponse{Intent: nextIntent, Payload: map[string]interface{}{"message": "Thank you for your interest. Let me share the job details with you."}}
	case "INTEREST":
		return &AutoResponse{Intent: nextIntent, Payload: map[string]interface{}{"message": "Great! Let me tell you more about the position and discuss next steps."}}
	case "NEGOTIATION":
		return &AutoResponse{Intent: nextIntent, Payload: map[string]interface{}{"message": "Let's discuss compensation and benefits."}}
	case "OFFER":
		return &AutoResponse{Intent: nextIntent, Payload: map[string]interface{}{"message": "We'd like to extend an offer."}}
	case "CONFIRM":
		return &AutoResponse{Intent: nextIntent, Payload: map[string]interface{}{"type": "confirmation"}}
	default:
		return &AutoResponse{Intent: "INTRODUCTION", Payload: map[string]interface{}{"message": "Thank you for your message."}}
	}
}

// intentToScenario maps intent strings to PromptScenarioType
func intentToScenario(intent string) model.PromptScenarioType {
	switch intent {
	case "NEGOTIATION":
		return model.ScenarioNegotiation
	case "SALARY", "COMPENSATION":
		return model.ScenarioSalary
	case "INTERVIEW", "SCHEDULE":
		return model.ScenarioInterview
	case "OFFER":
		return model.ScenarioOffer
	case "DECLINE", "REJECT":
		return model.ScenarioDecline
	case "TERMINATION", "WITHDRAW":
		return model.ScenarioTermination
	default:
		return model.ScenarioGreeting
	}
}

func (s *MessageQueueService) payloadToXML(msg *rabbitmq.AgentMessage) string {
	payload := ""
	if msg.Payload != nil {
		data, _ := json.Marshal(msg.Payload)
		payload = string(data)
	}
	return fmt.Sprintf(`<message>
	<header>
		<message_id>%s</message_id>
		<timestamp>%s</timestamp>
		<sender_id>%s</sender_id>
		<receiver_id>%s</receiver_id>
	</header>
	<payload>
		<intent>%s</intent>
		<parameters>%s</parameters>
	</payload>
</message>`, msg.MessageID, msg.Timestamp.Format(time.RFC3339), msg.SenderID, msg.ReceiverID, msg.Intent, payload)
}

func (s *MessageQueueService) responseToXML(resp *AutoResponse) string {
	payload := ""
	if resp.Payload != nil {
		data, _ := json.Marshal(resp.Payload)
		payload = string(data)
	}
	return fmt.Sprintf(`<message>
	<header>
		<message_id>%s</message_id>
		<timestamp>%s</timestamp>
		<sender_id>system</sender_id>
	</header>
	<payload>
		<intent>%s</intent>
		<parameters>%s</parameters>
	</payload>
</message>`, uuid.New().String(), time.Now().Format(time.RFC3339), resp.Intent, payload)
}

// PublishMessage publishes a message to the agent queue
func (s *MessageQueueService) PublishMessage(matchID, senderID, intent string, payload map[string]interface{}) error {
	if s.rmq == nil {
		return fmt.Errorf("RabbitMQ not configured")
	}

	msg := &rabbitmq.AgentMessage{
		MessageID:  uuid.New().String(),
		SenderID:   senderID,
		ReceiverID: "",
		Intent:     intent,
		MatchID:    matchID,
		Payload:    payload,
		Timestamp:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.rmq.PublishAgentMessage(ctx, msg)
}
