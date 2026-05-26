package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeAgent  = "agent.exchange"
	QueueAgentIn   = "agent.inbound"
	QueueAgentOut  = "agent.outbound"
	RoutingKeyIn   = "agent.in"
	RoutingKeyOut  = "agent.out"
	MaxRetries     = 3
	BaseRetryDelay = time.Second
)

// Config holds RabbitMQ connection configuration.
type Config struct {
	URL      string
	Username string
	Password string
	Host     string
	Port     string
}

// DefaultConfig returns a Config populated from environment variables.
func DefaultConfig() *Config {
	return &Config{
		URL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Username: getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASS", "guest"),
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnv("RABBITMQ_PORT", "5672"),
	}
}

// AgentMessage represents a message in the agent queue.
type AgentMessage struct {
	MessageID  string                 `json:"message_id"`
	SenderID   string                 `json:"sender_id"`
	ReceiverID string                 `json:"receiver_id"`
	Intent     string                 `json:"intent"`
	MatchID    string                 `json:"match_id"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	ResponseTo string                 `json:"response_to,omitempty"`
}

// RabbitMQ manages connections, channels, queues, and message routing.
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *Config
	mu      sync.RWMutex
	done    chan struct{}
}

// New creates and initializes a new RabbitMQ client.
func NewRabbitMQ(cfg *Config) (*RabbitMQ, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		channel: ch,
		config:  cfg,
		done:    make(chan struct{}),
	}

	if err := rmq.setup(); err != nil {
		rmq.Close()
		return nil, err
	}

	return rmq, nil
}

func (r *RabbitMQ) setup() error {
	// Declare exchange
	err := r.channel.ExchangeDeclare(
		ExchangeAgent,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare inbound queue with dead letter exchange
	args := amqp.Table{
		"x-dead-letter-exchange": ExchangeAgent + ".dlx",
	}
	_, err = r.channel.QueueDeclare(
		QueueAgentIn,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare inbound queue: %w", err)
	}

	// Declare dead letter exchange and queue
	err = r.channel.ExchangeDeclare(
		ExchangeAgent+".dlx",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	_, err = r.channel.QueueDeclare(
		QueueAgentIn+".dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	err = r.channel.QueueBind(
		QueueAgentIn+".dlq",
		RoutingKeyIn,
		ExchangeAgent+".dlx",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// Declare outbound queue
	_, err = r.channel.QueueDeclare(
		QueueAgentOut,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare outbound queue: %w", err)
	}

	// Bind queues to exchange
	err = r.channel.QueueBind(
		QueueAgentIn,
		RoutingKeyIn,
		ExchangeAgent,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind inbound queue: %w", err)
	}

	err = r.channel.QueueBind(
		QueueAgentOut,
		RoutingKeyOut,
		ExchangeAgent,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind outbound queue: %w", err)
	}

	return nil
}

// Close cleanly shuts down the RabbitMQ connection.
func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	close(r.done)

	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// PublishAgentMessage publishes an agent message with retry.
// Retries up to MaxRetries times with exponential backoff on failure.
func (r *RabbitMQ) PublishAgentMessage(ctx context.Context, msg *AgentMessage) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("channel is closed")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         body,
	}

	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		err = ch.PublishWithContext(
			ctx,
			ExchangeAgent,
			RoutingKeyIn,
			false,
			false,
			publishing,
		)
		if err == nil {
			log.Printf("Published agent message: %s -> %s", msg.SenderID, msg.Intent)
			return nil
		}

		lastErr = err
		log.Printf("Publish attempt %d/%d failed for message %s: %v", attempt, MaxRetries, msg.MessageID, err)

		if attempt < MaxRetries {
			delay := BaseRetryDelay * time.Duration(1<<(attempt-1))
			log.Printf("Retrying publish in %v...", delay)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed to publish message after %d retries: %w", MaxRetries, lastErr)
}

// Consume starts consuming messages from the inbound queue with retry support.
// Failed messages go to the dead letter queue after exhausting retries.
func (r *RabbitMQ) Consume(ctx context.Context, handler func(*AgentMessage) error) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("channel is closed")
	}

	msgs, err := ch.Consume(
		QueueAgentIn,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-r.done:
				return
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				var msg AgentMessage
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					d.Nack(false, false)
					continue
				}

				if err := r.processWithRetry(d, &msg, handler); err != nil {
					log.Printf("Message failed after %d retries: %v", MaxRetries, err)
					d.Nack(false, false) // goes to DLQ
				}
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) processWithRetry(d amqp.Delivery, msg *AgentMessage, handler func(*AgentMessage) error) error {
	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		if err := handler(msg); err != nil {
			lastErr = err
			log.Printf("Attempt %d/%d failed for message %s: %v", attempt, MaxRetries, msg.MessageID, err)

			if attempt < MaxRetries {
				delay := BaseRetryDelay * time.Duration(1<<(attempt-1))
				log.Printf("Retrying in %v...", delay)
				time.Sleep(delay)
			}
			continue
		}
		d.Ack(false)
		return nil
	}
	return lastErr
}

// IsConnected returns true if the RabbitMQ connection is active.
func (r *RabbitMQ) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn != nil && !r.conn.IsClosed()
}
