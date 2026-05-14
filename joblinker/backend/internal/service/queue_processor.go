package service

import (
	"context"
	"log"
	"time"

	"joblinker/pkg/rabbitmq"

	"github.com/google/uuid"
)

// QueueProcessor handles RabbitMQ consumption and publishing
type QueueProcessor struct {
	rmq *rabbitmq.RabbitMQ
}

// NewQueueProcessor creates a new QueueProcessor
func NewQueueProcessor(rmq *rabbitmq.RabbitMQ) *QueueProcessor {
	return &QueueProcessor{rmq: rmq}
}

// StartConsuming starts consuming messages from the queue
func (p *QueueProcessor) StartConsuming(ctx context.Context, handler func(*rabbitmq.AgentMessage) error) error {
	if p.rmq == nil {
		log.Println("RabbitMQ not configured, skipping consumer")
		return nil
	}
	return p.rmq.Consume(ctx, handler)
}

// PublishMessage publishes an agent message with retry
func (p *QueueProcessor) PublishMessage(msg *rabbitmq.AgentMessage) error {
	if p.rmq == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.rmq.PublishAgentMessage(ctx, msg)
}

// PublishToA2A routes a message to the other agent in the A2A dialogue
func (p *QueueProcessor) PublishToA2A(senderID, receiverID, intent, matchID string, payload map[string]interface{}) {
	if receiverID == "" {
		return
	}
	msg := &rabbitmq.AgentMessage{
		MessageID:  uuid.New().String(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		Intent:     intent,
		MatchID:    matchID,
		Payload:    payload,
		Timestamp:  time.Now(),
	}
	if err := p.PublishMessage(msg); err != nil {
		log.Printf("A2A publish failed: %v", err)
	} else {
		log.Printf("A2A message routed: %s -> %s (intent: %s)", senderID, receiverID, intent)
	}
}

// IsConfigured returns whether RabbitMQ is available
func (p *QueueProcessor) IsConfigured() bool {
	return p.rmq != nil
}
