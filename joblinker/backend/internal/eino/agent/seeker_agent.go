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

// SeekerAgent represents the job seeker agent
type SeekerAgent struct {
	model       *chatmodel.EinoChatModel
	promptLoader *prompt.Loader
	scenario    prompt.Scenario
	memory      *memory.AgentMemory
}

// NewSeekerAgent creates a new seeker agent
func NewSeekerAgent(aiClient *ai.Client) *SeekerAgent {
	loader := prompt.NewLoader()
	loader.LoadSeekerTemplates()
	return &SeekerAgent{
		model:       chatmodel.NewEinoChatModel(aiClient),
		promptLoader: loader,
		scenario:    prompt.ScenarioGreeting,
	}
}

// NewSeekerAgentWithMemory creates a seeker agent with memory support
func NewSeekerAgentWithMemory(aiClient *ai.Client, agentMemory *memory.AgentMemory) *SeekerAgent {
	loader := prompt.NewLoader()
	loader.LoadSeekerTemplates()
	return &SeekerAgent{
		model:       chatmodel.NewEinoChatModel(aiClient),
		promptLoader: loader,
		scenario:    prompt.ScenarioGreeting,
		memory:      agentMemory,
	}
}

// Chat processes a single chat message and returns the response
func (a *SeekerAgent) Chat(ctx context.Context, userMessage string) (string, error) {
	messages, err := a.promptLoader.Format(ctx, a.scenario, userMessage, nil)
	if err != nil {
		return "", fmt.Errorf("seeker agent: failed to format prompt: %w", err)
	}

	resp, err := a.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("seeker agent: failed to generate response: %w", err)
	}

	return resp.Content, nil
}

// ChatWithHistory processes a message with conversation history
func (a *SeekerAgent) ChatWithHistory(ctx context.Context, history []*schema.Message, userMessage string) (string, error) {
	messages, err := a.promptLoader.Format(ctx, a.scenario, userMessage, history)
	if err != nil {
		return "", fmt.Errorf("seeker agent: failed to format prompt: %w", err)
	}

	resp, err := a.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("seeker agent: failed to generate response: %w", err)
	}

	return resp.Content, nil
}

// SetScenario sets the conversation scenario for prompt selection
func (a *SeekerAgent) SetScenario(scenario prompt.Scenario) {
	a.scenario = scenario
}

// GetModel returns the underlying chat model (for testing)
func (a *SeekerAgent) GetModel() *chatmodel.EinoChatModel {
	return a.model
}

// GetMemory returns the agent memory (nil if not configured)
func (a *SeekerAgent) GetMemory() *memory.AgentMemory {
	return a.memory
}

// SetMemory sets the agent memory
func (a *SeekerAgent) SetMemory(agentMemory *memory.AgentMemory) {
	a.memory = agentMemory
}

// AddMemoryMessage adds a message to conversation memory
func (a *SeekerAgent) AddMemoryMessage(ctx context.Context, role schema.RoleType, content string) error {
	if a.memory == nil {
		return nil
	}
	return a.memory.AddMessage(ctx, role, content)
}

// GetMemoryMessages returns conversation history from memory
func (a *SeekerAgent) GetMemoryMessages() []*schema.Message {
	if a.memory == nil {
		return nil
	}
	return a.memory.GetMessages()
}

// ChatWithMemory processes a message using memory for context
func (a *SeekerAgent) ChatWithMemory(ctx context.Context, userMessage string) (string, error) {
	var history []*schema.Message
	if a.memory != nil {
		history = a.memory.GetMessages()
	}

	messages, err := a.promptLoader.Format(ctx, a.scenario, userMessage, history)
	if err != nil {
		return "", fmt.Errorf("seeker agent: failed to format prompt: %w", err)
	}

	resp, err := a.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("seeker agent: failed to generate response: %w", err)
	}

	// Store user message and assistant response in memory
	if a.memory != nil {
		_ = a.memory.AddMessage(ctx, schema.User, userMessage)
		_ = a.memory.AddMessage(ctx, schema.Assistant, resp.Content)
	}

	return resp.Content, nil
}
