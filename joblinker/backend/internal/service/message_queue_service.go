package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"
	"joblinker/pkg/rabbitmq"

	"github.com/google/uuid"
)

type MessageQueueService struct {
	rmq         *rabbitmq.RabbitMQ
	messageRepo *repository.MessageRepository
	matchRepo   *repository.MatchRepository
	agentRepo   *repository.AgentRepository
	jobRepo     *repository.JobRepository
	aiClient    *ai.Client
}

func NewMessageQueueService(
	rmq *rabbitmq.RabbitMQ,
	messageRepo *repository.MessageRepository,
	matchRepo *repository.MatchRepository,
	agentRepo *repository.AgentRepository,
	jobRepo *repository.JobRepository,
) *MessageQueueService {
	aiClient := ai.NewClient()
	log.Printf("MessageQueueService AI client - BaseURL: %s, APIKey length: %d", aiClient.BaseURL, len(aiClient.APIKey))
	return &MessageQueueService{
		rmq:         rmq,
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
		agentRepo:   agentRepo,
		jobRepo:     jobRepo,
		aiClient:    aiClient,
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
					if s, ok := skill.(string); ok {
						jobSkills = append(jobSkills, s)
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
					if s, ok := skill.(string); ok {
						seekerSkills = append(seekerSkills, s)
					}
				}
			}
		}
	}

	// Get conversation context for better responses
	conversationContext := s.buildAgentContext(msg, match)

	// Enhanced prompt with professional tone and job details
	prompt := fmt.Sprintf(`You are a professional AI recruitment agent in an A2A (Agent-to-Agent) recruitment platform.

## Current Conversation State
- Incoming intent: %s
- Your role: %s

## Conversation History
%s

## Job Details (include in response when relevant)
- Position: %s
- Location: %s
- Salary Range: $%d - $%d
- Required Skills: %v

## Candidate Profile
- Skills: %v

## Response Requirements
You MUST respond with ONLY a valid JSON object (no markdown, no explanation):
{"intent":"ONE_OF:[INTRODUCTION,INTEREST,NEGOTIATION,OFFER,CONFIRM,SCHEDULE,INQUIRY]","message":"Your professional response message here","data":{"title":"Job title if relevant","location":"Job location if relevant","salary_min":number,"salary_max":number,"skills":["skill1","skill2"],"message":"Short message for display"}}

## Intent Progression Rules
- INQUIRY -> respond with INTRODUCTION, include job details (title, salary, location, skills)
- INTRODUCTION -> respond with INTEREST, acknowledge candidate background
- INTEREST -> respond with NEGOTIATION, discuss salary/benefits if appropriate
- NEGOTIATION -> respond with OFFER, present formal offer terms
- OFFER -> respond with CONFIRM or DECLINE
- SCHEDULE -> respond with CONFIRM with interview details

## Professional Tone Guidelines
- Be concise but informative
- Use natural language, avoid robotic phrasing
- Include specific details (salary numbers, location, skills) when relevant
- For off-hours messages, acknowledge timing professionally
`,
		msg.Intent,
		senderAgent.Type,
		conversationContext,
		jobTitle,
		jobLocation,
		salaryMin,
		salaryMax,
		jobSkills,
		seekerSkills)

	response, err := s.aiClient.Chat(
		"You are a professional AI recruitment agent.",
		prompt,
	)

	if err != nil {
		log.Printf("AI response failed, using fallback: %v (API key length: %d)", err, len(s.aiClient.APIKey))
		return s.fallbackResponse(msg.Intent)
	}

	log.Printf("DEBUG: AI response received, length=%d, content=%s", len(response), response)

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
