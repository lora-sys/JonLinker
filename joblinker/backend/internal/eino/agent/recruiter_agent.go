package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"

	"joblinker/internal/eino/chatmodel"
	"joblinker/internal/eino/memory"
	"joblinker/internal/eino/prompt"
	"joblinker/pkg/ai"
)

// RecruiterAgent represents the recruiter agent
type RecruiterAgent struct {
	model       *chatmodel.EinoChatModel
	promptLoader *prompt.Loader
	scenario    prompt.Scenario
	memory      *memory.AgentMemory
}

// NewRecruiterAgent creates a new recruiter agent
func NewRecruiterAgent(aiClient *ai.Client) *RecruiterAgent {
	loader := prompt.NewLoader()
	loader.LoadRecruiterTemplates()
	return &RecruiterAgent{
		model:       chatmodel.NewEinoChatModel(aiClient),
		promptLoader: loader,
		scenario:    prompt.ScenarioGreeting,
	}
}

// NewRecruiterAgentWithMemory creates a recruiter agent with memory support
func NewRecruiterAgentWithMemory(aiClient *ai.Client, agentMemory *memory.AgentMemory) *RecruiterAgent {
	loader := prompt.NewLoader()
	loader.LoadRecruiterTemplates()
	return &RecruiterAgent{
		model:       chatmodel.NewEinoChatModel(aiClient),
		promptLoader: loader,
		scenario:    prompt.ScenarioGreeting,
		memory:      agentMemory,
	}
}

// Chat processes a single chat message and returns the response
func (a *RecruiterAgent) Chat(ctx context.Context, userMessage string) (string, error) {
	messages, err := a.promptLoader.Format(ctx, a.scenario, userMessage, nil)
	if err != nil {
		return "", fmt.Errorf("recruiter agent: failed to format prompt: %w", err)
	}

	resp, err := a.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("recruiter agent: failed to generate response: %w", err)
	}

	return resp.Content, nil
}

// ChatWithHistory processes a message with conversation history
func (a *RecruiterAgent) ChatWithHistory(ctx context.Context, history []*schema.Message, userMessage string) (string, error) {
	messages, err := a.promptLoader.Format(ctx, a.scenario, userMessage, history)
	if err != nil {
		return "", fmt.Errorf("recruiter agent: failed to format prompt: %w", err)
	}

	resp, err := a.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("recruiter agent: failed to generate response: %w", err)
	}

	return resp.Content, nil
}

// ChatWithMemory processes a message using memory for context
func (a *RecruiterAgent) ChatWithMemory(ctx context.Context, userMessage string) (string, error) {
	var history []*schema.Message
	if a.memory != nil {
		history = a.memory.GetMessages()
	}

	messages, err := a.promptLoader.Format(ctx, a.scenario, userMessage, history)
	if err != nil {
		return "", fmt.Errorf("recruiter agent: failed to format prompt: %w", err)
	}

	resp, err := a.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("recruiter agent: failed to generate response: %w", err)
	}

	// Store messages in memory
	if a.memory != nil {
		_ = a.memory.AddMessage(ctx, schema.User, userMessage)
		_ = a.memory.AddMessage(ctx, schema.Assistant, resp.Content)
	}

	return resp.Content, nil
}

// SetScenario sets the conversation scenario for prompt selection
func (a *RecruiterAgent) SetScenario(scenario prompt.Scenario) {
	a.scenario = scenario
}

// GetModel returns the underlying chat model (for testing)
func (a *RecruiterAgent) GetModel() *chatmodel.EinoChatModel {
	return a.model
}

// GetMemory returns the agent memory
func (a *RecruiterAgent) GetMemory() *memory.AgentMemory {
	return a.memory
}

// SetMemory sets the agent memory
func (a *RecruiterAgent) SetMemory(agentMemory *memory.AgentMemory) {
	a.memory = agentMemory
}

// AddMemoryMessage adds a message to conversation memory
func (a *RecruiterAgent) AddMemoryMessage(ctx context.Context, role schema.RoleType, content string) error {
	if a.memory == nil {
		return nil
	}
	return a.memory.AddMessage(ctx, role, content)
}

// GetMemoryMessages returns conversation history from memory
func (a *RecruiterAgent) GetMemoryMessages() []*schema.Message {
	if a.memory == nil {
		return nil
	}
	return a.memory.GetMessages()
}
