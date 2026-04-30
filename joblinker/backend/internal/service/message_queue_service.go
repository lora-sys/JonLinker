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
	rmq          *rabbitmq.RabbitMQ
	messageRepo  *repository.MessageRepository
	matchRepo    *repository.MatchRepository
	agentRepo    *repository.AgentRepository
	jobRepo      *repository.JobRepository
	aiClient     *ai.Client
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
	Intent   string
	Payload  map[string]interface{}
}

func (s *MessageQueueService) generateAutoResponse(msg *rabbitmq.AgentMessage, match *model.Match, senderAgent *model.Agent) *AutoResponse {
	// Get job and agent details for context
	var jobDesc string
	var seekerSkills []string

	if job, err := s.jobRepo.GetByID(match.JobID); err == nil {
		jobDesc = job.StructuredJSON
	}
	if seeker, err := s.agentRepo.GetByID(match.SeekerAgentID); err == nil {
		var config map[string]interface{}
		if json.Unmarshal([]byte(seeker.ConfigJSON), &config) == nil {
			if skills, ok := config["skills"].([]interface{}); ok {
				for _, s := range skills {
					seekerSkills = append(seekerSkills, fmt.Sprintf("%v", s))
				}
			}
		}
	}
	// Generate AI response
	log.Printf("DEBUG: Calling AI client with BaseURL=%s, APIKeyLen=%d", s.aiClient.BaseURL, len(s.aiClient.APIKey))

	// Get conversation context for better responses
	conversationContext := s.buildAgentContext(msg, match)

	prompt := fmt.Sprintf(`You are an AI recruitment agent in an A2A (Agent-to-Agent) recruitment platform.
Current intent from incoming message: %s

%s

Job details (use these to personalize your response):
%s

Your role: %s

IMPORTANT: You must respond with ONLY a valid JSON object in this exact format (no markdown, no explanation):
{"intent":"ONE_OF:[INTRODUCTION,INTEREST,NEGOTIATION,OFFER,CONFIRM,SCHEDULE,INQUIRY],"message":"Your professional response message here","data":{"title":"Job title if relevant","location":"Job location if relevant","salary_min":number,"salary_max":number,"skills":["skill1","skill2"],"message":"Short message for display"}}

The intent should progress the conversation naturally. For example:
- If intent is INQUIRY -> respond with INTRODUCTION and include job details
- If intent is INTRODUCTION -> respond with INTEREST
- If intent is INTEREST -> respond with NEGOTIATION and include salary details
- If intent is NEGOTIATION -> respond with OFFER
- If intent is OFFER -> respond with CONFIRM
- If intent is SCHEDULE -> respond with CONFIRM
`,
		msg.Intent,
		conversationContext,
		jobDesc,
		senderAgent.Type)

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
