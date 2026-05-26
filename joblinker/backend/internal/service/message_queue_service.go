package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"joblinker/internal/eino/chatmodel"
	"joblinker/internal/eino/runner"
	"joblinker/internal/eino/sessionstore"
	"joblinker/internal/eino/stategraph"
	"joblinker/internal/eino/prompt"

	"github.com/cloudwego/eino/schema"
	"joblinker/internal/agent"
	"joblinker/internal/cache"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"
	"joblinker/pkg/rabbitmq"

	"github.com/google/uuid"
)

type BroadcastFunc func(matchID string, eventType string, payload interface{})

// AutoResponse represents an AI-generated response
type AutoResponse struct {
	Intent  string
	Payload map[string]interface{}
}

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
	// StateGraph Runner (optional, for Phase 1 full autonomous pipeline)
	stategraphRunner   *runner.StateGraphRunner
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
	// StateGraph tracking: which matches already have a graph running
	runningGraphs   map[string]bool
	runningGraphsMu sync.Mutex
	// Session store for persisting autonomous conversation sessions
	sessionStore sessionstore.Store
	// Session service for higher-level session operations (reopen, etc.)
	sessionSvc *sessionstore.SessionService
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

// SetStateGraphRunner sets the StateGraph Runner for autonomous pipeline processing
func (s *MessageQueueService) SetStateGraphRunner(sgr *runner.StateGraphRunner) {
	s.stategraphRunner = sgr
	log.Printf("MessageQueueService: StateGraph Runner configured")
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

// SetSessionStore sets the session store for persisting autonomous conversation sessions
func (s *MessageQueueService) SetSessionStore(store sessionstore.Store) {
	s.sessionStore = store
	log.Printf("MessageQueueService: Session store configured")
}

// SetSessionService sets the SessionService for higher-level session operations.
func (s *MessageQueueService) SetSessionService(svc *sessionstore.SessionService) {
	s.sessionSvc = svc
	log.Printf("MessageQueueService: SessionService configured")
}

// ReopenSession creates a new session version for a paused match (delegates to SessionService).
// Returns the new session ID or empty string on failure.
func (s *MessageQueueService) ReopenSession(ctx context.Context, matchID uuid.UUID) string {
	if s.sessionSvc == nil {
		log.Printf("[Session] ReopenSession: SessionService not configured")
		return ""
	}
	sid, err := s.sessionSvc.ReopenSession(ctx, matchID)
	if err != nil {
		log.Printf("[Session] ReopenSession failed for match=%s: %v", matchID, err)
		return ""
	}
	log.Printf("[Session] Reopened session %s for match=%s", sid, matchID)
	if s.broadcastFn != nil {
		s.broadcastFn(matchID.String(), "session_reopened", map[string]interface{}{
			"match_id":   matchID.String(),
			"session_id": sid,
			"version":    func() int { _, v, _ := sessionstore.ParseSessionID(sid); return v }(),
			"timestamp":  time.Now(),
		})
	}
	return sid
}

// RecordTurn records a single message turn to the session store.
// Creates a new session if none exists for the match.
// Called from handleAgentMessage for both incoming and outgoing messages.
func (s *MessageQueueService) RecordTurn(ctx context.Context, matchID uuid.UUID, role, content string) {
	if s.sessionStore == nil {
		return
	}

	ver, err := s.sessionStore.LatestVersion(ctx, matchID)
	if err != nil {
		log.Printf("[Session] RecordTurn: failed to get version for match=%s: %v", matchID, err)
		return
	}

	var sessionID string
	if ver == 0 {
		// No session exists — create one
		sessionID = s.CreateSession(ctx, matchID)
		if sessionID == "" {
			log.Printf("[Session] RecordTurn: failed to create session for match=%s", matchID)
			return
		}
	} else {
		sessionID = sessionstore.SessionID(matchID, ver)
		session, loadErr := s.sessionStore.Load(ctx, sessionID)
		if loadErr != nil {
			log.Printf("[Session] RecordTurn: failed to load session %s: %v", sessionID, loadErr)
			return
		}
		if session.Status == sessionstore.SessionStatusConcluded {
			// Session concluded — try to reopen (creates a new version)
			newSid := s.ReopenSession(ctx, matchID)
			if newSid == "" {
				// Reopen failed — create fresh session
				sessionID = s.CreateSession(ctx, matchID)
				if sessionID == "" {
					return
				}
			} else {
				sessionID = newSid
			}
		}
	}

	if err := s.sessionStore.AppendMessages(ctx, sessionID, []sessionstore.Message{
		{Role: role, Content: content},
	}); err != nil {
		log.Printf("[Session] RecordTurn: append to %s failed: %v", sessionID, err)
	}
}

// CreateSession creates a new session for the given match and returns the session ID.
func (s *MessageQueueService) CreateSession(ctx context.Context, matchID uuid.UUID) string {
	if s.sessionStore == nil {
		log.Printf("[Session] sessionStore is nil, cannot create session for match=%s", matchID)
		return ""
	}
	ver, err := s.sessionStore.LatestVersion(ctx, matchID)
	if err != nil {
		log.Printf("[Session] Failed to get latest version for match=%s: %v", matchID, err)
		return ""
	}
	newVer := ver + 1
	sessionID := sessionstore.SessionID(matchID, newVer)
	if err := s.sessionStore.Create(ctx, sessionID, matchID, newVer); err != nil {
		log.Printf("[Session] Failed to create session for match=%s: %v", matchID, err)
		return ""
	}
	log.Printf("[Session] Created session %s for match=%s", sessionID, matchID)

	if s.broadcastFn != nil {
		s.broadcastFn(matchID.String(), "session_created", map[string]interface{}{
			"match_id":   matchID.String(),
			"session_id": sessionID,
			"version":    newVer,
			"timestamp":  time.Now(),
		})
	}

	return sessionID
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
		runningGraphs:      make(map[string]bool),
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

	if currentRound > 20 {
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

	jobCtx := s.getJobContextString(match)

	// Extract incoming message content for session recording
	msgContent := ""
	if msg.Payload != nil {
		if content, ok := msg.Payload["content"].(string); ok {
			msgContent = content
		}
	}
	if msgContent == "" {
		msgContent = msg.Intent
	}
	senderRole := string(senderAgent.Type)

	// Prefer ADK Runner (real Eino ADK agent)
	var response *AutoResponse
	legacyFallback := false

	if s.adkRunner != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		enrichedMsg := msgContent
		if jobCtx != "" {
			enrichedMsg = fmt.Sprintf("%s\n\n%s", jobCtx, msgContent)
		}

		adkResponse, adkErr := s.generateWithADK(ctx, enrichedMsg, senderAgent, matchID.String())
		if adkErr != nil {
			log.Printf("ADK Runner failed: %v, falling back to legacy AI", adkErr)
		} else {
			intent := s.parseIntent(adkResponse, msg.Intent)
			response = &AutoResponse{
				Intent:  intent,
				Payload: map[string]interface{}{"message": adkResponse},
			}
			log.Printf("ADK Runner generated response: intent=%s", intent)
		}
	}

	// Fallback to Eino streaming (preferred), then sync Eino, then legacy AI
	if response == nil && s.einoRunner != nil {
		response = s.generateEinoResponseStream(msg, match, senderAgent, jobCtx)
	}
	if response == nil {
		legacyEinoResponse := s.generateEinoResponse(msg, match, senderAgent, jobCtx)
		if legacyEinoResponse != nil {
			response = legacyEinoResponse
		}
	}
	if response == nil {
		log.Printf("Eino returned nil, falling back to legacy AI")
		legacyFallback = true
		response = s.generateAutoResponse(msg, match, senderAgent)
	}

	// Record both sides of this turn to the session store
	incomingRole := senderRole
	if incomingRole == "seeker" {
		incomingRole = "user"
	}
	s.RecordTurn(context.Background(), matchID, incomingRole, msgContent)
	var responseText string
	if response.Payload != nil {
		if text, ok := response.Payload["message"].(string); ok {
			responseText = text
		}
	}
	if responseText != "" {
		s.RecordTurn(context.Background(), matchID, "assistant", responseText)
	}

	// Ensure DB records are created for OFFER/SCHEDULE intents
	// (ADK and Eino runners only return intent strings without calling tools)
	// Skip if legacy fallback was used — generateAutoResponse already calls tools internally
	if !legacyFallback {
	 if response.Intent == "OFFER" {
		salary := 0
		startDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
		if match.Job != nil {
		 var jobData map[string]interface{}
		 if err := json.Unmarshal(match.Job.StructuredJSON, &jobData); err == nil {
		  if min, ok := jobData["salary_min"].(float64); ok {
		   if max, ok := jobData["salary_max"].(float64); ok && max > min {
		    salary = int((min + max) / 2)
		   }
		  }
		 }
		}
		if salary == 0 {
		 salary = 150000 // fallback
		}
		toolResult, err := s.toolExecutor.ExecuteTool(context.Background(), matchID, "create_offer", map[string]interface{}{
		 "match_id":   matchID.String(),
		 "salary":     float64(salary),
		 "start_date": startDate,
		})
		if err != nil {
		 log.Printf("Failed to create offer via tool: %v", err)
		} else if toolResult != nil && toolResult.Success {
		 log.Printf("Offer created via tool (post-AI): %+v", toolResult.Data)
		}
	}
	if response.Intent == "SCHEDULE" {
	datetime := time.Now().AddDate(0, 0, 14).Format(time.RFC3339) // 2 weeks from now
	toolResult, err := s.toolExecutor.ExecuteTool(context.Background(), matchID, "schedule_interview", map[string]interface{}{
	 "match_id":       matchID.String(),
	 "datetime":       datetime,
	 "interview_type": "video",
	})
	if err != nil {
	 log.Printf("Failed to schedule interview via tool: %v", err)
	} else if toolResult != nil && toolResult.Success {
	 log.Printf("Interview scheduled via tool (post-AI): %+v", toolResult.Data)
	}
}
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
			"message_id":      responseMsg.ID.String(),
			"match_id":        matchID.String(),
			"sender_agent_id": responseAgentID.String(),
			"content_xml":     responseMsg.ContentXML,
			"text":            responseText,
			"intent_type":     response.Intent,
			"created_at":      responseMsg.CreatedAt,
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

	// Get job and candidate details for context (used by both Eino and legacy paths)
	jobContext := s.getJobContextString(match)
	var jobTitle, jobLocation string
	var salaryMin, salaryMax int
	var jobSkills []string
	var seekerSkills []string

	if job, err := s.jobRepo.GetByID(match.JobID); err == nil {
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

	// Try Eino-based agent processing first if runner is configured
	if s.einoRunner != nil {
		einoResponse := s.generateEinoResponse(msg, match, senderAgent, jobContext)
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

	// Build a role-specific system prompt with job context
	var agentRole string
	if agentType == model.AgentTypeSeeker {
		agentRole = "job seeker"
	} else {
		agentRole = "recruiter"
	}
	systemPrompt := fmt.Sprintf(
		"You are a professional %s recruitment agent. Handle the '%s' position. "+
			"Location: %s | Salary: $%d-$%d | Required skills: %v. "+
			"Candidate skills: %v. "+
			"Engage in a realistic, specific conversation — reference actual job requirements and candidate experience. "+
			"Do NOT generate generic responses. Discuss concrete details about the role, tech stack, team, and qualifications. "+
			"Use tools (schedule_interview, create_offer, query_jobs) when the conversation calls for concrete actions.",
		agentRole, jobTitle,
		jobLocation, salaryMin, salaryMax, jobSkills,
		seekerSkills,
	)

	// Call AI for intent recognition and response using ChatWithTools for function calling
	response, toolCall, err := s.aiClient.ChatWithTools(
		systemPrompt,
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

// getJobContextString builds a job context string for enriching AI prompts
func (s *MessageQueueService) getJobContextString(match *model.Match) string {
	var jobTitle, jobLocation string
	var salaryMin, salaryMax int
	var jobSkills []string

	if job, err := s.jobRepo.GetByID(match.JobID); err == nil {
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
	var seekerName string
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
			if name, ok := config["name"].(string); ok {
				seekerName = name
			}
		}
	}

	return fmt.Sprintf(
		"[Job Context] Position: %s | Location: %s | Salary: $%d-$%d | Required: %v | Candidate: %s | Skills: %v",
		jobTitle, jobLocation, salaryMin, salaryMax, jobSkills,
		seekerName, seekerSkills,
	)
}

// generateEinoResponseStream wraps generateEinoResponse and broadcasts each chunk via WebSocket.
// Returns nil if Eino is not configured or fails, triggering fallback.
func (s *MessageQueueService) generateEinoResponseStream(msg *rabbitmq.AgentMessage, match *model.Match, senderAgent *model.Agent, jobContext string) *AutoResponse {
	// Streaming removed in S2 refactor; fall back to non-streaming response.
	// Broadcast the full response as a single chunk for backwards compatibility.
	resp := s.generateEinoResponse(msg, match, senderAgent, jobContext)
	if resp != nil && s.broadcastFn != nil {
		if payload, ok := resp.Payload["message"].(string); ok {
			s.broadcastFn(match.ID.String(), "adk_stream_chunk", map[string]interface{}{
				"chunk":           payload,
				"accumulated":     payload,
				"sender_agent_id": senderAgent.ID.String(),
			})
		}
	}
	return resp
}

// generateEinoResponse generates a response using Eino agents (seeker/recruiter pool)
// Returns nil if Eino is not configured or fails, triggering fallback
func (s *MessageQueueService) generateEinoResponse(msg *rabbitmq.AgentMessage, match *model.Match, senderAgent *model.Agent, jobContext string) *AutoResponse {
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

	// Inject job context into the message so Eino agents see actual role details
	enrichedMsg := msgContent
	if jobContext != "" {
		enrichedMsg = fmt.Sprintf("%s\n\n%s", jobContext, msgContent)
	}

	var response string
	var err error

	if senderAgent.Type == "seeker" {
		response, err = s.einoRunner.RunRecruiterTask(ctx, enrichedMsg)
	} else {
		response, err = s.einoRunner.RunSeekerTask(ctx, enrichedMsg)
	}

	if err != nil {
		log.Printf("Eino response error: %v", err)
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
		b, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Warning: failed to marshal confirmation payload: %v", err)
		} else {
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

	// Bug #10 fix: schedule auto-expiry after 24h
	reqID := cr.ID
	matchIDStr := matchID
	time.AfterFunc(24*time.Hour, func() {
		req, err := s.matchRepo.GetConfirmationRequestByID(reqID)
		if err != nil {
			log.Printf("[Confirm expiry] failed to fetch request %s: %v", reqID, err)
			return
		}
		if req.Status == model.ConfirmationStatusPending {
			req.Status = model.ConfirmationStatusExpired
			if err := s.matchRepo.UpdateConfirmationRequest(req); err != nil {
				log.Printf("[Confirm expiry] failed to expire request %s: %v", reqID, err)
			} else {
				log.Printf("[Confirm expiry] auto-expired confirmation request %s (match=%s)", reqID, matchIDStr)
			}
		}
	})
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

	// Update match status based on confirmation type
	switch cr.Type {
	case model.ConfirmationTypeOffer:
		_ = s.matchRepo.UpdateStatus(matchID, model.MatchStatusHired)
		s.broadcastStateChange(matchID.String(), string(model.MatchStatusHired), "Offer accepted, match completed")
		log.Printf("Match %s status updated to hired after offer confirmation", matchID)
	case model.ConfirmationTypeInterview:
		_ = s.matchRepo.UpdateStatus(matchID, model.MatchStatusInterviewing)
		s.broadcastStateChange(matchID.String(), string(model.MatchStatusInterviewing), "Interview scheduled, conversation resumes")
		log.Printf("Match %s status updated to interviewing after schedule confirmation", matchID)
	default:
		_ = s.matchRepo.UpdateStatus(matchID, model.MatchStatusHired)
	}

	// Only continue conversation for interview scheduling — offer accept (hired) ends the match
	if cr.Type != model.ConfirmationTypeOffer {
		s.continueAgentConversation(matchID, cr)
	}
	return nil
}

func (s *MessageQueueService) broadcastStateChange(matchID string, newState string, reason string) {
	if s.broadcastFn != nil {
		s.broadcastFn(matchID, "fsm_state_change", map[string]interface{}{
			"match_id":  matchID,
			"new_state": newState,
			"reason":    reason,
			"timestamp": time.Now(),
		})
	}
}

// runStateGraphForMatch runs the autonomous StateGraph pipeline for a match in the current goroutine.
// It creates seeker/recruiter agents, runs all 5 phases (INTRODUCTION → COMPLETED),
// and broadcasts each message and phase change in real-time.
func (s *MessageQueueService) runStateGraphForMatch(ctx context.Context, matchID uuid.UUID, msg *rabbitmq.AgentMessage, match *model.Match, senderAgent *model.Agent) {
	graphCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create session for this match if session store is configured
	var sessionID string
	if s.sessionStore != nil {
		sessionID = s.CreateSession(graphCtx, matchID)
	}

	otherAgentID := s.getOtherAgentID(match, uuid.MustParse(msg.SenderID))
	seekerAgentID := match.SeekerAgentID

	// Broadcast callbacks
	onMessage := func(sender, message string, phase stategraph.Phase) {
		// Determine which agent ID sent this
		agentID := otherAgentID
		if sender == "Seeker" || strings.HasPrefix(sender, "Seeker") {
			agentID = seekerAgentID
		} else if sender == "Recruiter" || strings.HasPrefix(sender, "Recruiter") {
			agentID = otherAgentID
		}

		if s.broadcastFn != nil {
			s.broadcastFn(matchID.String(), "state_graph_message", map[string]interface{}{
				"match_id":        matchID.String(),
				"sender_agent_id": agentID.String(),
				"sender":          sender,
				"text":            message,
				"phase":           string(phase),
				"timestamp":       time.Now(),
			})
		}

		// Persist message to database
		contentXML := fmt.Sprintf("<message><text>%s</text></message>", message)
		msgRecord := &model.Message{
			ID:            uuid.New(),
			MatchID:       matchID,
			TenantID:      match.TenantID,
			SenderAgentID: agentID,
			ContentXML:    contentXML,
			IntentType:    string(phase),
		}
		if err := s.messageRepo.Create(msgRecord); err != nil {
			log.Printf("[StateGraph] Failed to persist message for match=%s: %v", matchID, err)
		}

		// Also append to session store if configured
		if s.sessionStore != nil && sessionID != "" {
			s.sessionStore.AppendMessages(graphCtx, sessionID, []sessionstore.Message{
				{Role: sender, Content: message},
			})
		}
	}

	onPhase := func(phase stategraph.Phase) {
		if s.broadcastFn != nil {
			s.broadcastFn(matchID.String(), "fsm_state_change", map[string]interface{}{
				"match_id":  matchID.String(),
				"new_state": string(phase),
				"reason":    "stategraph_progression",
				"timestamp": time.Now(),
			})
		}
	}

	log.Printf("[StateGraph] Running autonomous pipeline for match=%s", matchID)

	state, err := s.stategraphRunner.RunGraph(graphCtx, matchID,
		stategraph.WithCallbacks(onMessage, onPhase),
		stategraph.WithMaxNegotiationRounds(10),
		stategraph.WithMaxInterviewRounds(5),
		stategraph.WithMaxOfferRounds(3),
	)
	if err != nil {
		log.Printf("[StateGraph] Error for match=%s: %v", matchID, err)
		s.broadcastStateChange(matchID.String(), "ERROR", err.Error())
		return
	}

	log.Printf("[StateGraph] Completed for match=%s: final=%s accepted=%v", matchID, state.Phase, state.OfferAccepted)

	// Conclude session if session store is configured
	if s.sessionStore != nil && sessionID != "" {
		if s.broadcastFn != nil {
			s.broadcastFn(matchID.String(), "session_concluded", map[string]interface{}{
				"match_id":   matchID.String(),
				"session_id": sessionID,
				"reason":     "naturally",
				"timestamp":  time.Now(),
			})
		}
		s.sessionStore.Conclude(graphCtx, sessionID, "naturally")
		log.Printf("[Session] Concluded session %s for match=%s", sessionID, matchID)
	}

	// Update match status based on offer outcome
	switch {
	case state.OfferAccepted:
		if err := s.matchRepo.UpdateStatus(matchID, model.MatchStatusHired); err != nil {
			log.Printf("[StateGraph] Failed to update match %s to Hired: %v", matchID, err)
		} else {
			log.Printf("[StateGraph] Match %s status set to Hired (offer accepted)", matchID)
		}

		// Create an offer record so the Offers page shows this offer
		if err := s.createOfferFromState(matchID, match.TenantID, state); err != nil {
			log.Printf("[StateGraph] Failed to create offer record for match %s: %v", matchID, err)
		}
	case state.OfferDeclined:
		if err := s.matchRepo.UpdateStatus(matchID, model.MatchStatusRejected); err != nil {
			log.Printf("[StateGraph] Failed to update match %s to Rejected: %v", matchID, err)
		} else {
			log.Printf("[StateGraph] Match %s status set to Rejected (offer declined)", matchID)
		}
	default:
		log.Printf("[StateGraph] No status update for match %s (no offer decision)", matchID)
	}

	// Broadcast completion event
	if s.broadcastFn != nil {
		s.broadcastFn(matchID.String(), "state_graph_complete", map[string]interface{}{
			"match_id":   matchID.String(),
			"phase":      string(state.Phase),
			"candidate":  state.CandidateName,
			"job_title":  state.JobTitle,
			"accepted":   state.OfferAccepted,
			"declined":   state.OfferDeclined,
			"msg_count":  len(state.Messages),
			"timestamp":  time.Now(),
		})
	}
}

// createOfferFromState creates an offer record in the database from StateGraph completion state.
func (s *MessageQueueService) createOfferFromState(matchID, tenantID uuid.UUID, state *stategraph.RecruitmentState) error {
	// Extract a salary amount from the state
	salaryStr := state.OfferAmount
	if salaryStr == "" {
		// Try to extract from the last negotiation round
		if len(state.NegotiationRounds) > 0 {
			last := state.NegotiationRounds[len(state.NegotiationRounds)-1]
			salaryStr = last.RecruiterOffer
		}
	}
	if salaryStr == "" {
		salaryStr = "80000"
	}

	// Parse salary as integer (best-effort)
	baseSalary := 80000
	if parsed, err := strconv.Atoi(regexp.MustCompile(`\d+`).FindString(salaryStr)); err == nil && parsed > 0 {
		baseSalary = parsed
	}

	compensation := map[string]interface{}{
		"base_salary": baseSalary,
		"currency":    "USD",
		"bonus": map[string]interface{}{
			"amount":      baseSalary * 10 / 100,
			"description": "Annual performance bonus",
		},
		"equity": map[string]interface{}{
			"shares":         1000,
			"vesting_period": "4-year vesting with 1-year cliff",
		},
		"benefits": []string{"Health Insurance", "401(k) matching", "Unlimited PTO"},
	}

	compJSON, err := json.Marshal(compensation)
	if err != nil {
		return fmt.Errorf("marshal compensation: %w", err)
	}

	offer := &model.Offer{
		ID:               uuid.New(),
		MatchID:          matchID,
		TenantID:         tenantID,
		CompensationJSON: json.RawMessage(compJSON),
		StartDate:        time.Now().AddDate(0, 1, 0), // 1 month from now
		Status:           model.OfferStatusAccepted,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.offerRepo.Create(offer); err != nil {
		return fmt.Errorf("create offer: %w", err)
	}

	log.Printf("[StateGraph] Created offer record %s for match %s (salary=%d)", offer.ID, matchID, baseSalary)
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

// generateWithADK uses the ADK Runner to generate a response with real-time streaming
func (s *MessageQueueService) generateWithADK(ctx context.Context, msgContent string, senderAgent *model.Agent, matchID ...string) (string, error) {
	if s.adkRunner == nil {
		return "", fmt.Errorf("adk runner not configured")
	}

	mid := ""
	if len(matchID) > 0 {
		mid = matchID[0]
	}

	events := s.adkRunner.Query(ctx, msgContent)

	var finalResponse string
	var streamBuf string
	toolCallStarts := make(map[string]time.Time)
	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Printf("ADK error for match %s: %v", mid, event.Err)
			if s.broadcastFn != nil && mid != "" {
				s.broadcastFn(mid, "adk_error", map[string]interface{}{
					"error": event.Err.Error(),
				})
			}
			return "", event.Err
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			mo := event.Output.MessageOutput

			// Step 1: Always capture assistant text first (even when tool calls are also present)
			// ADK ChatModelAgent may bundle Content + ToolCalls in a single event.
			// Without this, every tool call iteration discards the accompanying text,
			// and the agent loop ends with finalResponse empty ("adk runner: empty response").
			if mo.Message != nil && mo.Role == schema.Assistant && mo.Message.Content != "" && !mo.IsStreaming {
				finalResponse = mo.Message.Content
			}

			// Tool result → broadcast adk_tool_call with result
			if mo.Message != nil && mo.Role == schema.Tool {
				result := mo.Message.Content
				errorMsg := ""
				if event.Action != nil && event.Action.Exit {
					result = "exit"
				}

				var durationMs int64
				if startTime, ok := toolCallStarts[mo.Message.ToolCallID]; ok {
					durationMs = time.Since(startTime).Milliseconds()
					delete(toolCallStarts, mo.Message.ToolCallID)
				}

				if s.broadcastFn != nil && mid != "" {
					s.broadcastFn(mid, "adk_tool_call", map[string]interface{}{
						"tool_name":       mo.ToolName,
						"call_id":         mo.Message.ToolCallID,
						"result":          result,
						"status":          "completed",
						"error":           errorMsg,
						"duration_ms":     durationMs,
						"sender_agent_id": senderAgent.ID.String(),
						"sender_label":    senderAgent.Type,
					})
				}
				continue
			}

			// Assistant message with tool calls → broadcast adk_tool_call_start for each
			if mo.Message != nil && mo.Role == schema.Assistant && len(mo.Message.ToolCalls) > 0 {
				for _, tc := range mo.Message.ToolCalls {
					toolCallStarts[tc.ID] = time.Now()

					var argsJSON interface{}
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &argsJSON); err != nil {
						argsJSON = tc.Function.Arguments
					}
					if s.broadcastFn != nil && mid != "" {
						s.broadcastFn(mid, "adk_tool_call_start", map[string]interface{}{
							"tool_name":       tc.Function.Name,
							"call_id":         tc.ID,
							"arguments":       argsJSON,
							"sender_agent_id": senderAgent.ID.String(),
							"sender_label":    senderAgent.Type,
						})
					}
				}
				continue
			}

			// Streaming text chunk
			if mo.IsStreaming && mo.Message != nil {
				streamBuf += mo.Message.Content
				if s.broadcastFn != nil && mid != "" {
					s.broadcastFn(mid, "adk_stream_chunk", map[string]interface{}{
						"chunk":            mo.Message.Content,
						"accumulated":      streamBuf,
						"sender_agent_id":  senderAgent.ID.String(),
					})
				}
				continue
			}

			// Non-streaming assistant text → final response
			if mo.Message != nil && mo.Role == schema.Assistant {
				finalResponse = mo.Message.Content
			}
		}

		// Action-based events (exit, transfer, etc.)
		if event.Action != nil {
			log.Printf("ADK action for match %s: %+v", mid, event.Action)
			if s.broadcastFn != nil && mid != "" {
				s.broadcastFn(mid, "adk_tool_call", map[string]interface{}{
					"action":          event.Action,
					"sender_agent_id": senderAgent.ID.String(),
					"sender_label":    senderAgent.Type,
				})
			}
		}
	}

	if finalResponse == "" && streamBuf != "" {
		finalResponse = streamBuf
	}

	if finalResponse == "" {
		// ADK loop completed with no text at all (e.g. all iterations produced
		// tool calls with empty Content). Return a safe fallback rather than
		// an error, so the caller doesn't fall through to an inferior Eino path
		// or push an empty message to the user.
		log.Printf("ADK runner for match %s: no text captured (all %d iterations were tool calls without explanatory text), returning fallback", mid, len(toolCallStarts))
		return "Thank you for your message. I've reviewed the available information. Let me know if you have any specific questions.", nil
	}
	return finalResponse, nil
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
