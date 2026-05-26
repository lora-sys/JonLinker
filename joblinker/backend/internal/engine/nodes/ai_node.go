package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"joblinker/internal/adapters"
	"joblinker/internal/core"
)

// AINode wraps the AI client with FSM state awareness to drive conversation flow.
type AINode struct {
	client *adapters.AIClient
	fsm    *core.FSM
}

// NewAINode creates a new AINode.
func NewAINode(client *adapters.AIClient, fsm *core.FSM) *AINode {
	return &AINode{
		client: client,
		fsm:    fsm,
	}
}

// Process generates an AI response based on the current FSM state and input context.
// Input:
//   - state: current FSM state name (string representation of core.State)
//   - match_id: conversation/match identifier
//   - agent_type: "recruiter" or "seeker"
//   - context: additional conversation context
//
// Output:
//   - response: AI-generated text
//   - next_state: FSM state transition
//   - intent: detected conversation intent
func (n *AINode) Process(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	stateStr, _ := input["state"].(string)
	agentType, _ := input["agent_type"].(string)
	contextStr, _ := input["context"].(string)

	if stateStr == "" {
		stateStr = string(core.StateIdle)
	}
	if agentType == "" {
		agentType = "recruiter"
	}

	systemPrompt := n.buildSystemPrompt(stateStr, agentType)
	userPrompt := n.buildUserPrompt(stateStr, contextStr, input)

	log.Printf("[AINode] state=%s agent=%s", stateStr, agentType)

	response, err := n.client.Chat(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI chat failed: %w", err)
	}

	intent := n.detectIntent(response)
	nextState := n.applyTransition(core.State(stateStr), intent)

	return map[string]interface{}{
		"response":   response,
		"next_state": nextState,
		"intent":     intent,
	}, nil
}

// ProcessWithTools generates an AI response with tool-calling support.
func (n *AINode) ProcessWithTools(ctx context.Context, input map[string]interface{}, tools []adapters.Tool, executor adapters.ToolExecutor) (map[string]interface{}, error) {
	stateStr, _ := input["state"].(string)
	agentType, _ := input["agent_type"].(string)
	contextStr, _ := input["context"].(string)

	if stateStr == "" {
		stateStr = string(core.StateIdle)
	}
	if agentType == "" {
		agentType = "recruiter"
	}

	systemPrompt := n.buildSystemPrompt(stateStr, agentType)
	userPrompt := n.buildUserPrompt(stateStr, contextStr, input)

	log.Printf("[AINode+Tools] state=%s agent=%s tools=%d", stateStr, agentType, len(tools))

	response, _, err := n.client.ChatWithTools(systemPrompt, userPrompt, tools, executor)
	if err != nil {
		return nil, fmt.Errorf("AI chat with tools failed: %w", err)
	}

	intent := n.detectIntent(response)
	nextState := n.applyTransition(core.State(stateStr), intent)

	return map[string]interface{}{
		"response":   response,
		"next_state": nextState,
		"intent":     intent,
	}, nil
}

func (n *AINode) buildSystemPrompt(stateStr string, agentType string) string {
	roleDesc := "a recruiter looking for qualified candidates"
	if agentType == "seeker" {
		roleDesc = "a job seeker looking for opportunities"
	}

	return fmt.Sprintf(`You are an AI agent representing %s in a recruitment platform.
Your role is to engage in professional dialogue about job opportunities.
Respond concisely and professionally. Format responses as plain text.
Never reveal sensitive personal information.

Current conversation stage: %s`, roleDesc, stateStr)
}

func (n *AINode) buildUserPrompt(stateStr string, contextStr string, input map[string]interface{}) string {
	prompt := ""
	if contextStr != "" {
		prompt += fmt.Sprintf("Context: %s\n\n", contextStr)
	}

	if jobTitle, ok := input["job_title"].(string); ok && jobTitle != "" {
		prompt += fmt.Sprintf("Position: %s\n", jobTitle)
	}
	if skills, ok := input["skills"].(string); ok && skills != "" {
		prompt += fmt.Sprintf("Skills: %s\n", skills)
	}
	if location, ok := input["location"].(string); ok && location != "" {
		prompt += fmt.Sprintf("Location: %s\n", location)
	}

	prompt += "\nBased on the context above, generate an appropriate professional response for the recruitment dialogue. Keep responses brief (1-3 sentences) and focused on the recruitment topic."

	return prompt
}

// detectIntent maps the AI response to a conversation intent and Event.
func (n *AINode) detectIntent(response string) string {
	// Try JSON structure first
	var structured struct {
		Intent string `json:"intent"`
	}
	if err := json.Unmarshal([]byte(response), &structured); err == nil && structured.Intent != "" {
		return structured.Intent
	}

	lower := strings.ToLower(response)
	if len(lower) > 200 {
		lower = lower[:200]
	}

	switch {
	case strings.Contains(lower, "interested") || strings.Contains(lower, "exciting") || strings.Contains(lower, "great opportunity"):
		return "INTEREST_EXPRESSED"
	case strings.Contains(lower, "salary") || strings.Contains(lower, "compensation") || strings.Contains(lower, "offer") || strings.Contains(lower, "package"):
		return "NEGOTIATE"
	case strings.Contains(lower, "interview") || strings.Contains(lower, "schedule") || strings.Contains(lower, "meet") || strings.Contains(lower, "discuss further"):
		return "SCHEDULE_INTERVIEW"
	case strings.Contains(lower, "accept") || strings.Contains(lower, "confirm") || strings.Contains(lower, "agree") || strings.Contains(lower, "looking forward"):
		return "OFFER_ACCEPTED"
	case strings.Contains(lower, "reject") || strings.Contains(lower, "decline") || strings.Contains(lower, "not interested") || strings.Contains(lower, "not a good fit"):
		return "OFFER_DECLINED"
	default:
		return "MATCH_FOUND"
	}
}

// applyTransition maps an intent string to an FSM event and applies the transition.
// Returns the new state as a string, or the current state if the transition is invalid.
func (n *AINode) applyTransition(currentState core.State, intent string) string {
	eventMap := map[string]core.Event{
		"MATCH_FOUND":        core.EventMatchFound,
		"INTEREST_EXPRESSED": core.EventInterestExpressed,
		"NEGOTIATE":          core.EventNegotiate,
		"SCHEDULE_INTERVIEW": core.EventScheduleInterview,
		"OFFER_ACCEPTED":     core.EventOfferAccepted,
		"OFFER_DECLINED":     core.EventOfferDeclined,
		"REJECTED":           core.EventRejected,
		"PAUSE":              core.EventPause,
		"RESUME":             core.EventResume,
		"TIMEOUT":            core.EventTimeout,
	}

	event, ok := eventMap[intent]
	if !ok {
		return string(currentState)
	}

	if err := n.fsm.Handle(event); err != nil {
		log.Printf("[AINode] invalid transition: %v", err)
		return string(currentState)
	}

	return string(n.fsm.CurrentState())
}
