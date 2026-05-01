package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"joblinker/internal/agent"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"
	"joblinker/pkg/rabbitmq"

	"github.com/google/uuid"
)

type MessageQueueService struct {
	rmq                *rabbitmq.RabbitMQ
	messageRepo        *repository.MessageRepository
	matchRepo          *repository.MatchRepository
	agentRepo          *repository.AgentRepository
	jobRepo            *repository.JobRepository
	offerRepo          *repository.OfferRepository
	interviewRepo      *repository.InterviewRepository
	aiClient           *ai.Client
	promptService      *AgentPromptService
	toolExecutor       *agent.ToolExecutor
}

func NewMessageQueueService(
	rmq *rabbitmq.RabbitMQ,
	messageRepo *repository.MessageRepository,
	matchRepo *repository.MatchRepository,
	agentRepo *repository.AgentRepository,
	jobRepo *repository.JobRepository,
	offerRepo *repository.OfferRepository,
	interviewRepo *repository.InterviewRepository,
) *MessageQueueService {
	aiClient := ai.NewClient()
	promptService := NewAgentPromptService()
	toolExecutor := agent.NewToolExecutor(jobRepo, nil, offerRepo, interviewRepo)
	log.Printf("MessageQueueService AI client - BaseURL: %s, APIKey length: %d", aiClient.BaseURL, len(aiClient.APIKey))
	return &MessageQueueService{
		rmq:                rmq,
		messageRepo:        messageRepo,
		matchRepo:          matchRepo,
		agentRepo:          agentRepo,
		jobRepo:            jobRepo,
		offerRepo:          offerRepo,
		interviewRepo:      interviewRepo,
		aiClient:           aiClient,
		promptService:      promptService,
		toolExecutor:       toolExecutor,
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

	// Parse match ID
	matchID, err := uuid.Parse(msg.MatchID)
	if err != nil {
		log.Printf("Invalid match ID %s: %v", msg.MatchID, err)
		return nil // Don't requeue - invalid match ID is permanent error
	}

	// Get match to verify access
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		log.Printf("Match %s not found: %v", matchID, err)
		return nil // Don't requeue - match not found is permanent
	}

	// Get sender agent
	senderAgent, err := s.agentRepo.GetByID(uuid.MustParse(msg.SenderID))
	if err != nil {
		log.Printf("Sender agent %s not found: %v", msg.SenderID, err)
		return nil // Don't requeue - agent not found is permanent
	}

	// Note: We don't store the incoming message again here.
	// The REST API handler already stored it when the user sent it.
	// We only need to generate and store the auto-response.

	// Generate auto-response based on intent
	response := s.generateAutoResponse(msg, match, senderAgent)

	// Determine who responds - recruiter for INTRODUCTION, seeker for others
	responseAgentID := senderAgent.ID
	if msg.Intent == "INTRODUCTION" {
		// For introduction, recruiter should respond - get from job's agent
		job, err := s.jobRepo.GetByID(match.JobID)
		if err == nil && job != nil {
			responseAgentID = job.AgentID
		}
	}

	responseMsg := &model.Message{
		ID:            uuid.New(),
		MatchID:       matchID,
		SenderAgentID: responseAgentID,
		ContentXML:     s.responseToXML(response),
		IntentType:    response.Intent,
	}
	if err := s.messageRepo.Create(responseMsg); err != nil {
		log.Printf("Failed to store response message: %v", err)
	}

	// Note: We do NOT publish response back to queue - that would cause a loop
	// The WebSocket handler will push the response to connected clients directly

	log.Printf("Auto-response: %s -> %s (intent: %s)", response.Intent, msg.SenderID, msg.Intent)
	return nil
}

type AutoResponse struct {
	Intent  string
	Payload map[string]interface{}
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
		if json.Unmarshal([]byte(job.StructuredJSON), &jobData) == nil {
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
		if json.Unmarshal([]byte(seeker.ConfigJSON), &config) == nil {
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

	// Call AI for intent recognition and response
	response, err := s.aiClient.Chat(
		"You are a professional AI recruitment agent.",
		enhancedPrompt,
	)

	if err != nil {
		log.Printf("AI response failed, using fallback: %v", err)
		return s.fallbackResponse(msg.Intent)
	}

	log.Printf("DEBUG: AI response received, content=%s", response)

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

	// Execute tools based on intent (since AI API doesn't support function calling natively)
	log.Printf("DEBUG: Checking incoming intent '%s' vs AI response '%s' for tool execution", msg.Intent, aiResp.Intent)

	// Execute tools based on USER's intent, not AI's response (AI may progress intent)
	// The user sent SCHEDULE, so we execute schedule_interview tool
	if msg.Intent == "SCHEDULE" || msg.Intent == "CONFIRM" || aiResp.Intent == "SCHEDULE" || aiResp.Intent == "CONFIRM" {
		log.Printf("DEBUG: Executing schedule_interview tool for match %s", match.ID)
		// Schedule interview
		datetime := "2026-06-01T10:00:00Z"
		if aiResp.Data != nil {
			if dt, ok := aiResp.Data["datetime"].(string); ok {
				datetime = dt
			}
		}

		log.Printf("DEBUG: Calling toolExecutor.ExecuteTool with match_id=%s, datetime=%s", match.ID.String(), datetime)
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
		// Create offer
		salary := 150000
		startDate := "2026-07-01"
		if aiResp.Data != nil {
			if s, ok := aiResp.Data["salary_max"].(float64); ok {
				salary = int(s)
			}
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

	return &AutoResponse{
		Intent:  aiResp.Intent,
		Payload: aiResp.Data,
	}
}

func (s *MessageQueueService) buildAgentContext(msg *rabbitmq.AgentMessage, match *model.Match) string {
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

func (s *MessageQueueService) fallbackResponse(currentIntent string) *AutoResponse {
	// Simple fallback responses if AI fails
	switch currentIntent {
	case "INTRODUCTION":
		return &AutoResponse{Intent: "INTEREST", Payload: map[string]interface{}{"message": "Thank you for your introduction. We are interested in your profile."}}
	case "INTEREST":
		return &AutoResponse{Intent: "NEGOTIATION", Payload: map[string]interface{}{"type": "salary", "current": 120000, "target": 150000}}
	case "NEGOTIATION":
		return &AutoResponse{Intent: "OFFER", Payload: map[string]interface{}{"type": "offer", "base_salary": 130000}}
	case "OFFER":
		return &AutoResponse{Intent: "CONFIRM", Payload: map[string]interface{}{"type": "acceptance"}}
	case "SCHEDULE":
		return &AutoResponse{Intent: "CONFIRM", Payload: map[string]interface{}{"type": "interview_confirmed"}}
	default:
		return &AutoResponse{Intent: "INQUIRY", Payload: map[string]interface{}{"message": "Thank you for your message."}}
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

	return s.rmq.PublishAgentMessage(context.Background(), msg)
}
