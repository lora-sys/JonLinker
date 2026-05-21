package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"

	"joblinker/internal/eino/chatmodel"
	"joblinker/internal/eino/prompt"
	"joblinker/pkg/ai"
)

// RecruiterAgent represents the recruiter agent
type RecruiterAgent struct {
	model       *chatmodel.EinoChatModel
	promptLoader *prompt.Loader
	scenario    prompt.Scenario
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

// Chat processes a chat message with optional conversation history and returns the response
func (a *RecruiterAgent) Chat(ctx context.Context, history []*schema.Message, userMessage string) (string, error) {
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

// SetScenario sets the conversation scenario for prompt selection
func (a *RecruiterAgent) SetScenario(scenario prompt.Scenario) {
	a.scenario = scenario
}

// GetModel returns the underlying chat model (for testing)
func (a *RecruiterAgent) GetModel() *chatmodel.EinoChatModel {
	return a.model
}
