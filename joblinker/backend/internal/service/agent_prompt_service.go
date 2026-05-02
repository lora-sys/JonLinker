package service

import (
	"fmt"
	"joblinker/internal/model"
	"joblinker/internal/service/prompts"
)

// AgentPromptService generates three-part prompts for agents
type AgentPromptService struct {
	seekerPrompts    map[model.PromptScenarioType]string
	recruiterPrompts map[model.PromptScenarioType]string
}

// NewAgentPromptService creates a new prompt service
func NewAgentPromptService() *AgentPromptService {
	return &AgentPromptService{
		seekerPrompts:    prompts.SeekerPrompts(),
		recruiterPrompts: prompts.RecruiterPrompts(),
	}
}

// GetPromptForScenario returns the three-part prompt for a given agent type and scenario
func (s *AgentPromptService) GetPromptForScenario(agentType model.AgentType, scenario model.PromptScenarioType) string {
	var promptMap map[model.PromptScenarioType]string
	if agentType == model.AgentTypeSeeker {
		promptMap = s.seekerPrompts
	} else {
		promptMap = s.recruiterPrompts
	}

	if prompt, ok := promptMap[scenario]; ok {
		return prompt
	}
	// Fallback to greeting if scenario not found
	return promptMap[model.ScenarioGreeting]
}

// BuildFullPrompt constructs the complete system prompt for an agent
func (s *AgentPromptService) BuildFullPrompt(agentType model.AgentType, scenario model.PromptScenarioType, context string) string {
	basePrompt := s.GetPromptForScenario(agentType, scenario)
	return fmt.Sprintf("%s\n\n## Current Context\n%s", basePrompt, context)
}