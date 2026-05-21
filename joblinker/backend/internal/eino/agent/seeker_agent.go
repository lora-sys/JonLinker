package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"

	"joblinker/internal/eino/chatmodel"
	"joblinker/internal/eino/prompt"
	"joblinker/pkg/ai"
)

// SeekerAgent represents the job seeker agent
type SeekerAgent struct {
	model        *chatmodel.EinoChatModel
	promptLoader *prompt.Loader
	scenario     prompt.Scenario
}

// NewSeekerAgent creates a new seeker agent
func NewSeekerAgent(aiClient *ai.Client) *SeekerAgent {
	loader := prompt.NewLoader()
	loader.LoadSeekerTemplates()
	return &SeekerAgent{
		model:        chatmodel.NewEinoChatModel(aiClient),
		promptLoader: loader,
		scenario:     prompt.ScenarioGreeting,
	}
}

// Chat processes a chat message with optional conversation history and returns the response
func (a *SeekerAgent) Chat(ctx context.Context, history []*schema.Message, userMessage string) (string, error) {
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
