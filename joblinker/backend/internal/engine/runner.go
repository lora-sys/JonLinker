package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"joblinker/internal/adapters"
	"joblinker/internal/core"
	"joblinker/internal/engine/nodes"

	"github.com/google/uuid"
)

// Runner is a RabbitMQ-driven graph executor.
// It consumes agent messages from the queue, routes them through the
// conversation graph, and publishes responses back.
type Runner struct {
	rmq    *adapters.RabbitMQ
	ai     *adapters.AIClient
	fsm    *core.FSM
	graph  *Graph
	active map[string]context.CancelFunc
	mu     sync.Mutex
}

// NewRunner creates a new Runner with the standard agent graph.
func NewRunner(rmq *adapters.RabbitMQ, ai *adapters.AIClient, toolNode *nodes.ToolNode, confirmNode *nodes.ConfirmNode) *Runner {
	fsm := core.NewFSM(uuid.New())
	memNode := nodes.NewMemoryNode()
	aiNode := nodes.NewAINode(ai, fsm)

	registry := &NodeRegistry{
		AINode:      aiNode,
		ToolNode:    toolNode,
		MemoryNode:  memNode,
		ConfirmNode: confirmNode,
	}

	graph := registry.BuildStandardGraph()

	return &Runner{
		rmq:    rmq,
		ai:     ai,
		fsm:    fsm,
		graph:  graph,
		active: make(map[string]context.CancelFunc),
	}
}

// Start begins consuming messages from the queue and processing them.
func (r *Runner) Start(ctx context.Context) error {
	log.Println("[Runner] starting agent graph runner")

	return r.rmq.Consume(ctx, func(msg *adapters.AgentMessage) error {
		log.Printf("[Runner] received message: %s from %s (intent=%s)", msg.MessageID, msg.SenderID, msg.Intent)
		return r.processMessage(ctx, msg)
	})
}

// processMessage handles a single agent message through the graph.
func (r *Runner) processMessage(ctx context.Context, msg *adapters.AgentMessage) error {
	input := map[string]interface{}{
		"match_id":  msg.MatchID,
		"message":   msg.Payload,
		"intent":    msg.Intent,
		"sender":    msg.SenderID,
		"receiver":  msg.ReceiverID,
		"agent_type": "recruiter",
		"state":     string(r.fsm.CurrentState()),
		"context":   r.buildContext(msg),
	}

	output, err := r.graph.Execute(ctx, input)
	if err != nil {
		log.Printf("[Runner] graph execution error: %v", err)
		return err
	}

	return r.publishResponse(ctx, msg, output)
}

// publishResponse sends the graph execution result back as a response message.
func (r *Runner) publishResponse(ctx context.Context, original *adapters.AgentMessage, output map[string]interface{}) error {
	response := &adapters.AgentMessage{
		MessageID:  uuid.New().String(),
		SenderID:   original.ReceiverID,
		ReceiverID: original.SenderID,
		Intent:     intentFromOutput(output),
		MatchID:    original.MatchID,
		Payload:    payloadFromOutput(output),
		Timestamp:  time.Now(),
		ResponseTo: original.MessageID,
	}

	return r.rmq.PublishAgentMessage(ctx, response)
}

// buildContext constructs a context string from the message payload.
func (r *Runner) buildContext(msg *adapters.AgentMessage) string {
	if msg.Payload == nil {
		return ""
	}

	if ctx, ok := msg.Payload["context"].(string); ok {
		return ctx
	}
	return ""
}

// Stop gracefully shuts down the runner and all active conversations.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, cancel := range r.active {
		cancel()
		delete(r.active, id)
	}
	log.Println("[Runner] stopped")
}

func intentFromOutput(output map[string]interface{}) string {
	if intent, ok := output["intent"].(string); ok {
		return intent
	}
	return "INQUIRY"
}

func payloadFromOutput(output map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{})

	if response, ok := output["response"].(string); ok {
		payload["response"] = response
	}
	if nextState, ok := output["next_state"].(string); ok {
		payload["next_state"] = nextState
	}
	if data, ok := output["data"]; ok {
		payload["data"] = data
	}

	return payload
}
