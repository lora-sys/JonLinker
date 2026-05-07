package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"joblinker/internal/agent"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"

	"github.com/google/uuid"
)

type ReasoningService struct {
	toolExecutor *agent.ToolExecutor
	matchRepo    *repository.MatchRepository
	messageRepo  *repository.MessageRepository
	promptService *AgentPromptService
	aiClient     *ai.Client
}

func NewReasoningService(
	toolExecutor *agent.ToolExecutor,
	matchRepo *repository.MatchRepository,
	messageRepo *repository.MessageRepository,
	promptService *AgentPromptService,
	aiClient *ai.Client,
) *ReasoningService {
	return &ReasoningService{
		toolExecutor: toolExecutor,
		matchRepo:    matchRepo,
		messageRepo:  messageRepo,
		promptService: promptService,
		aiClient:     aiClient,
	}
}

// ReasoningStep represents a step in the ReAct reasoning loop
type ReasoningStep struct {
	Phase    string `json:"phase"`    // "understand", "query", "analyze", "respond"
	Input    string `json:"input"`
	Output   string `json:"output"`
	ToolCall string `json:"tool_call,omitempty"`
	Result   string `json:"result,omitempty"`
}

// ReasoningResult holds the complete reasoning output
type ReasoningResult struct {
	Steps     []ReasoningStep `json:"steps"`
	FinalResponse string      `json:"final_response"`
	ToolCalls []string        `json:"tool_calls"`
}

// ProcessWithReasoning processes a user message through the ReAct reasoning loop
func (s *ReasoningService) ProcessWithReasoning(ctx context.Context, matchID uuid.UUID, agentID uuid.UUID, userMessage string) (*ReasoningResult, error) {
	result := &ReasoningResult{
		Steps:     []ReasoningStep{},
		ToolCalls: []string{},
	}

	// Step 1: Understand - extract requirements from user message
	understandStep := s.understandPhase(userMessage)
	result.Steps = append(result.Steps, understandStep)

	// Step 2: Query - determine which tools to call
	queryStep, toolsToCall := s.queryPhase(ctx, understandStep.Output, matchID)
	result.Steps = append(result.Steps, queryStep)

	// Step 3: Execute tool calls
	var toolResults []string
	for _, toolCall := range toolsToCall {
		result.ToolCalls = append(result.ToolCalls, toolCall.ToolName)
		execResult, err := s.toolExecutor.ExecuteTool(ctx, matchID, toolCall.ToolName, toolCall.Arguments)
		if err != nil {
			log.Printf("Tool %s failed: %v", toolCall.ToolName, err)
			toolResults = append(toolResults, fmt.Sprintf("Error: %v", err))
		} else {
			toolResults = append(toolResults, s.formatToolResult(execResult))
		}
	}

	// Step 4: Analyze - score matches against preferences
	analyzeStep := s.analyzePhase(understandStep.Output, toolResults)
	result.Steps = append(result.Steps, analyzeStep)

	// Step 5: Respond - generate response with reasoning explanation
	respondStep := s.respondPhase(ctx, matchID, agentID, result.Steps)
	result.Steps = append(result.Steps, respondStep)
	result.FinalResponse = respondStep.Output

	return result, nil
}

// understandPhase extracts requirements from user message
func (s *ReasoningService) understandPhase(message string) ReasoningStep {
	// Simple keyword extraction for understanding
	// In production, would use NER and intent classification
	lower := strings.ToLower(message)

	var extractedRequirements []string

	// Extract location
	if strings.Contains(lower, "remote") {
		extractedRequirements = append(extractedRequirements, "location: remote")
	} else if strings.Contains(lower, "onsite") {
		extractedRequirements = append(extractedRequirements, "location: onsite")
	} else if strings.Contains(lower, "hybrid") {
		extractedRequirements = append(extractedRequirements, "location: hybrid")
	}

	// Extract job type
	if strings.Contains(lower, "full-time") || strings.Contains(lower, "全职") {
		extractedRequirements = append(extractedRequirements, "job_type: full-time")
	} else if strings.Contains(lower, "part-time") || strings.Contains(lower, "兼职") {
		extractedRequirements = append(extractedRequirements, "job_type: part-time")
	}

	// Extract salary if mentioned
	if strings.Contains(lower, "salary") || strings.Contains(lower, "$") {
		extractedRequirements = append(extractedRequirements, "salary: mentioned")
	}

	// Extract skills
	skillKeywords := []string{"python", "golang", "java", "javascript", "react", "node"}
	for _, skill := range skillKeywords {
		if strings.Contains(lower, skill) {
			extractedRequirements = append(extractedRequirements, fmt.Sprintf("skill: %s", skill))
		}
	}

	output := strings.Join(extractedRequirements, ", ")
	if output == "" {
		output = "General inquiry - no specific requirements extracted"
	}

	return ReasoningStep{
		Phase:  "understand",
		Input:  message,
		Output: output,
	}
}

// queryPhase determines which tools to call based on requirements
func (s *ReasoningService) queryPhase(ctx context.Context, requirements string, matchID uuid.UUID) (ReasoningStep, []agent.ToolCallParams) {
	var toolsToCall []agent.ToolCallParams

	// Parse requirements and determine needed tools
	if strings.Contains(requirements, "skill") || strings.Contains(requirements, "job") {
		toolsToCall = append(toolsToCall, agent.ToolCallParams{
			MatchID:   matchID,
			ToolName:  "query_jobs",
			Arguments: map[string]interface{}{"limit": 5},
		})
	}

	if strings.Contains(requirements, "candidate") || strings.Contains(requirements, "profile") {
		toolsToCall = append(toolsToCall, agent.ToolCallParams{
			MatchID:   matchID,
			ToolName:  "get_candidate",
			Arguments: map[string]interface{}{},
		})
	}

	if strings.Contains(requirements, "offer") || strings.Contains(requirements, "salary") {
		toolsToCall = append(toolsToCall, agent.ToolCallParams{
			MatchID:   matchID,
			ToolName:  "create_offer",
			Arguments: map[string]interface{}{},
		})
	}

	if strings.Contains(requirements, "interview") || strings.Contains(requirements, "schedule") {
		toolsToCall = append(toolsToCall, agent.ToolCallParams{
			MatchID:   matchID,
			ToolName:  "schedule_interview",
			Arguments: map[string]interface{}{},
		})
	}

	if len(toolsToCall) == 0 {
		// Default to query_jobs if no specific tool identified
		toolsToCall = append(toolsToCall, agent.ToolCallParams{
			MatchID:   matchID,
			ToolName:  "query_jobs",
			Arguments: map[string]interface{}{"limit": 5},
		})
	}

	toolNames := make([]string, len(toolsToCall))
	for i, t := range toolsToCall {
		toolNames[i] = t.ToolName
	}

	return ReasoningStep{
		Phase:    "query",
		Input:    requirements,
		Output:   fmt.Sprintf("Calling tools: %s", strings.Join(toolNames, ", ")),
		ToolCall: strings.Join(toolNames, ", "),
	}, toolsToCall
}

// analyzePhase scores matches against preferences
func (s *ReasoningService) analyzePhase(requirements string, toolResults []string) ReasoningStep {
	// Simple analysis - in production would score against user preferences
	var analysis strings.Builder
	analysis.WriteString(fmt.Sprintf("Analyzed requirements: %s\n", requirements))
	analysis.WriteString(fmt.Sprintf("Tool results received: %d\n", len(toolResults)))

	for i, result := range toolResults {
		analysis.WriteString(fmt.Sprintf("Result %d: %s\n", i+1, truncateString(result, 100)))
	}

	return ReasoningStep{
		Phase:  "analyze",
		Input:  requirements,
		Output: analysis.String(),
	}
}

// respondPhase generates the final response with reasoning explanation
func (s *ReasoningService) respondPhase(ctx context.Context, matchID, agentID uuid.UUID, steps []ReasoningStep) ReasoningStep {
	// Build context from conversation history
	messages, _ := s.messageRepo.ListByMatchID(matchID)

	var contextBuilder strings.Builder
	for _, m := range messages {
		contextBuilder.WriteString(fmt.Sprintf("[%s] %s\n", m.IntentType, m.ContentXML))
	}

	// Get the conversation context
	conversationContext := contextBuilder.String()

	// Use AI if available for response generation
	if s.aiClient != nil {
		// Use AI client to generate response with reasoning context
		response, err := s.aiClient.GenerateAgentResponse(conversationContext, "seeker")
		if err == nil && response != "" {
			return ReasoningStep{
				Phase:  "respond",
				Input:  fmt.Sprintf("%d reasoning steps", len(steps)),
				Output: response,
			}
		}
		log.Printf("AI response generation failed: %v, falling back", err)
	}

	// Fallback: generate response from reasoning steps
	// Use prompt service to generate response
	prompt := s.promptService.BuildFullPrompt(model.AgentTypeSeeker, model.ScenarioGreeting, conversationContext)

	// Add chain-of-thought instruction for complex scenarios
	if len(steps) > 3 {
		prompt = prompt + "\n\nLet's think step by step about this situation."
	}

	// In production, would call AI client with this prompt
	// For now, generate a simple response based on the reasoning steps
	var response strings.Builder
	response.WriteString("Based on my analysis:\n\n")

	for _, step := range steps {
		if step.Phase != "respond" {
			response.WriteString(fmt.Sprintf("- [%s] %s\n", strings.ToUpper(step.Phase), step.Output))
		}
	}

	return ReasoningStep{
		Phase:  "respond",
		Input:  fmt.Sprintf("%d reasoning steps", len(steps)),
		Output: response.String(),
	}
}

// formatToolResult formats a tool execution result for display
func (s *ReasoningService) formatToolResult(result *agent.ToolExecutionResult) string {
	if result == nil {
		return "No result"
	}
	if !result.Success {
		return fmt.Sprintf("Error: %s", result.Error)
	}
	data, _ := json.Marshal(result.Data)
	return string(data)
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}