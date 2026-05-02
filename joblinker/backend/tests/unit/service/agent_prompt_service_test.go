package service_test

import (
	"strings"
	"testing"

	"joblinker/internal/model"
	"joblinker/internal/service"
)

func TestAgentPromptService_GetPromptForScenario(t *testing.T) {
	svc := service.NewAgentPromptService()

	tests := []struct {
		name      string
		agentType model.AgentType
		scenario  model.PromptScenarioType
		wantMUST  bool
		wantNOT   bool
		wantRULES bool
	}{
		{
			name:      "seeker greeting has three-part structure",
			agentType: model.AgentTypeSeeker,
			scenario:  model.ScenarioGreeting,
			wantMUST:   true,
			wantNOT:    true,
			wantRULES:  true,
		},
		{
			name:      "recruiter greeting has three-part structure",
			agentType: model.AgentTypeRecruiter,
			scenario:  model.ScenarioGreeting,
			wantMUST:   true,
			wantNOT:    true,
			wantRULES:  true,
		},
		{
			name:      "seeker negotiation prompt",
			agentType: model.AgentTypeSeeker,
			scenario:  model.ScenarioNegotiation,
			wantMUST:   true,
			wantNOT:    true,
			wantRULES:  true,
		},
		{
			name:      "recruiter salary prompt",
			agentType: model.AgentTypeRecruiter,
			scenario:  model.ScenarioSalary,
			wantMUST:   true,
			wantNOT:    true,
			wantRULES:  true,
		},
		{
			name:      "seeker interview prompt",
			agentType: model.AgentTypeSeeker,
			scenario:  model.ScenarioInterview,
			wantMUST:   true,
			wantNOT:    true,
			wantRULES:  true,
		},
		{
			name:      "recruiter offer prompt",
			agentType: model.AgentTypeRecruiter,
			scenario:  model.ScenarioOffer,
			wantMUST:   true,
			wantNOT:    true,
			wantRULES:  true,
		},
		{
			name:      "seeker decline prompt",
			agentType: model.AgentTypeSeeker,
			scenario:  model.ScenarioDecline,
			wantMUST:   true,
			wantNOT:    true,
			wantRULES:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := svc.GetPromptForScenario(tt.agentType, tt.scenario)

			// Verify three-part structure
			if tt.wantMUST && !strings.Contains(prompt, "MUST DO") {
				t.Errorf("prompt missing MUST DO section")
			}
			if tt.wantNOT && !strings.Contains(prompt, "MUST NOT DO") {
				t.Errorf("prompt missing MUST NOT DO section")
			}
			if tt.wantRULES && !strings.Contains(prompt, "BEHAVIOR RULES") {
				t.Errorf("prompt missing BEHAVIOR RULES section")
			}

			// For greeting scenario, verify agent type context
			if tt.scenario == model.ScenarioGreeting {
				if tt.agentType == model.AgentTypeSeeker {
					if !strings.Contains(strings.ToLower(prompt), "seeker") && !strings.Contains(strings.ToLower(prompt), "job") {
						t.Errorf("seeker prompt missing job context")
					}
				} else {
					if !strings.Contains(strings.ToLower(prompt), "recruit") && !strings.Contains(strings.ToLower(prompt), "employer") {
						t.Errorf("recruiter prompt missing recruiter/employer context")
					}
				}
			}
		})
	}
}

func TestAgentPromptService_BuildFullPrompt(t *testing.T) {
	svc := service.NewAgentPromptService()

	prompt := svc.BuildFullPrompt(
		model.AgentTypeSeeker,
		model.ScenarioNegotiation,
		"User interested in remote Python developer position, salary $120k",
	)

	// Verify base prompt included
	if !strings.Contains(prompt, "MUST DO") {
		t.Errorf("full prompt missing base MUST DO section")
	}
	if !strings.Contains(prompt, "MUST NOT DO") {
		t.Errorf("full prompt missing base MUST NOT DO section")
	}

	// Verify context appended
	if !strings.Contains(prompt, "User interested in remote Python developer position") {
		t.Errorf("full prompt missing provided context")
	}
}

func TestAgentPromptService_FallbackToGreeting(t *testing.T) {
	svc := service.NewAgentPromptService()

	// Request non-existent scenario - should fallback to greeting
	prompt := svc.GetPromptForScenario(model.AgentTypeSeeker, "nonexistent")

	// Should still return valid greeting prompt with three-part structure
	if !strings.Contains(prompt, "MUST DO") {
		t.Errorf("fallback prompt missing MUST DO section")
	}
}