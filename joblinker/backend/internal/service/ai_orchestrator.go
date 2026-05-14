package service

import (
	"context"
	"log"

	"joblinker/internal/agent"
	"joblinker/internal/cache"
	"joblinker/internal/eino/chatmodel"
	"joblinker/internal/eino/runner"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"
	"joblinker/pkg/rabbitmq"
)

// AIOrchestrator handles AI response generation and tool execution
type AIOrchestrator struct {
	aiClient         *ai.Client
	einoChatModel    *chatmodel.EinoChatModel
	promptService    *AgentPromptService
	toolExecutor     *agent.ToolExecutor
	toolCache        *cache.ToolCache
	einoRunner       *runner.AgentRunner
	contextOptimizer *ContextOptimizerService
	agentRepo        *repository.AgentRepository
	jobRepo          *repository.JobRepository
}

// NewAIOrchestrator creates a new AIOrchestrator
func NewAIOrchestrator(
	aiClient *ai.Client,
	einoChatModel *chatmodel.EinoChatModel,
	promptService *AgentPromptService,
	toolExecutor *agent.ToolExecutor,
	toolCache *cache.ToolCache,
	agentRepo *repository.AgentRepository,
	jobRepo *repository.JobRepository,
) *AIOrchestrator {
	return &AIOrchestrator{
		aiClient:      aiClient,
		einoChatModel: einoChatModel,
		promptService: promptService,
		toolExecutor:  toolExecutor,
		toolCache:     toolCache,
		agentRepo:     agentRepo,
		jobRepo:       jobRepo,
	}
}

// SetEinoRunner configures the optional Eino runner for advanced AI processing
func (o *AIOrchestrator) SetEinoRunner(einoRunner *runner.AgentRunner) {
	o.einoRunner = einoRunner
	log.Printf("AIOrchestrator: Eino Runner configured")
}

// SetContextOptimizer configures context optimization
func (o *AIOrchestrator) SetContextOptimizer(ctxOptimizer *ContextOptimizerService) {
	o.contextOptimizer = ctxOptimizer
	log.Printf("AIOrchestrator: Context Optimizer configured")
}

// AutoResponse represents an AI-generated response
type AutoResponse struct {
	Intent  string
	Payload map[string]interface{}
}

// GenerateResponse generates an AI response for an incoming agent message
func (o *AIOrchestrator) GenerateResponse(
	ctx context.Context,
	msg *rabbitmq.AgentMessage,
	match *model.Match,
	senderAgent *model.Agent,
	conversationCtx string,
) *AutoResponse {
	scenario := intentToScenario(msg.Intent)
	var agentType model.AgentType
	if senderAgent.Type == "seeker" {
		agentType = model.AgentTypeSeeker
	} else {
		agentType = model.AgentTypeRecruiter
	}

	// Try Eino first if configured
	if o.einoRunner != nil {
		einoResponse := o.generateEinoResponse(msg, match, senderAgent, scenario)
		if einoResponse != nil {
			return einoResponse
		}
	}

	// Legacy AI fallback
	fullPrompt := o.promptService.BuildFullPrompt(agentType, scenario, conversationCtx)
	return o.generateLegacyResponse(msg, match, senderAgent, fullPrompt)
}

func (o *AIOrchestrator) generateEinoResponse(
	msg *rabbitmq.AgentMessage,
	match *model.Match,
	senderAgent *model.Agent,
	scenario model.PromptScenarioType,
) *AutoResponse {
	return nil
}

func (o *AIOrchestrator) generateLegacyResponse(
	msg *rabbitmq.AgentMessage,
	match *model.Match,
	senderAgent *model.Agent,
	prompt string,
) *AutoResponse {
	systemPrompt := "You are a helpful recruitment assistant."
	response, err := o.aiClient.Chat(systemPrompt, prompt)
	if err != nil {
		log.Printf("Legacy AI response failed: %v", err)
		return &AutoResponse{
			Intent:  "INTRODUCTION",
			Payload: map[string]interface{}{"message": "Thank you for your interest. Let me find relevant information for you."},
		}
	}
	return &AutoResponse{
		Intent:  msg.Intent,
		Payload: map[string]interface{}{"message": response},
	}
}

// BuildConversationContext builds context string from message history
func (o *AIOrchestrator) BuildConversationContext(match *model.Match) string {
	return ""
}
