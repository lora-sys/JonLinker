package hooks

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// ---- context keys for per-request match data ----

type ctxKey string

const (
	CtxKeyMatchContext ctxKey = "jl_match_context"
	CtxKeyMatchID      ctxKey = "jl_match_id"
	CtxKeyAgentType    ctxKey = "jl_agent_type"
)

// MatchContext is the per-match data injected into agent system prompts.
type MatchContext struct {
	Phase         string  // INTRODUCTION / NEGOTIATION / INTERVIEW / OFFER / COMPLETED
	CandidateName string  `json:"candidate_name,omitempty"`
	CandidateBio  string  `json:"candidate_bio,omitempty"`
	JobTitle      string  `json:"job_title,omitempty"`
	JobDesc       string  `json:"job_desc,omitempty"`
	SalaryAgreed  string  `json:"salary_agreed,omitempty"`
	MatchScore    float64 `json:"match_score,omitempty"`
}

// WithMatchContext sets the match context into the context.Context for ADK agent injection.
func WithMatchContext(ctx context.Context, mc *MatchContext) context.Context {
	return context.WithValue(ctx, CtxKeyMatchContext, mc)
}

// GetMatchContext extracts match context from context.Context (may be nil).
func GetMatchContext(ctx context.Context) *MatchContext {
	mc, _ := ctx.Value(CtxKeyMatchContext).(*MatchContext)
	return mc
}

// WithMatchID sets a match ID into the context.
func WithMatchID(ctx context.Context, matchID string) context.Context {
	return context.WithValue(ctx, CtxKeyMatchID, matchID)
}

// WithAgentType sets the agent type into the context.
func WithAgentType(ctx context.Context, agentType string) context.Context {
	return context.WithValue(ctx, CtxKeyAgentType, agentType)
}

// ---- ContextInjector ChatModelAgentMiddleware ----

// ContextInjector is an ADK ChatModelAgentMiddleware that injects match context
// (candidate profile / job description / phase) into the agent's system instruction
// before each model invocation.
type ContextInjector struct{}

func NewContextInjector() *ContextInjector {
	return &ContextInjector{}
}

func (h *ContextInjector) WrapInvokableToolCall(_ context.Context, endpoint adk.InvokableToolCallEndpoint, _ *adk.ToolContext) (adk.InvokableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextInjector) WrapStreamableToolCall(_ context.Context, endpoint adk.StreamableToolCallEndpoint, _ *adk.ToolContext) (adk.StreamableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextInjector) WrapEnhancedInvokableToolCall(_ context.Context, endpoint adk.EnhancedInvokableToolCallEndpoint, _ *adk.ToolContext) (adk.EnhancedInvokableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextInjector) WrapEnhancedStreamableToolCall(_ context.Context, endpoint adk.EnhancedStreamableToolCallEndpoint, _ *adk.ToolContext) (adk.EnhancedStreamableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextInjector) WrapModel(_ context.Context, m model.BaseChatModel, _ *adk.ModelContext) (model.BaseChatModel, error) {
	return m, nil
}

// BeforeAgent injects match context + phase info into the agent's Instruction.
func (h *ContextInjector) BeforeAgent(ctx context.Context, c *adk.ChatModelAgentContext) (context.Context, *adk.ChatModelAgentContext, error) {
	mc := GetMatchContext(ctx)
	if mc == nil {
		return ctx, c, nil
	}
	block := buildContextBlock(mc)
	c.Instruction = block + "\n\n" + c.Instruction
	return ctx, c, nil
}

func (h *ContextInjector) BeforeModelRewriteState(ctx context.Context, state *adk.ChatModelAgentState, _ *adk.ModelContext) (context.Context, *adk.ChatModelAgentState, error) {
	return ctx, state, nil
}

func (h *ContextInjector) AfterModelRewriteState(ctx context.Context, state *adk.ChatModelAgentState, _ *adk.ModelContext) (context.Context, *adk.ChatModelAgentState, error) {
	return ctx, state, nil
}

func (h *ContextInjector) AfterAgent(ctx context.Context, state *adk.ChatModelAgentState, action *adk.AgentAction) (context.Context, *adk.ChatModelAgentState, *adk.AgentAction, error) {
	return ctx, state, action, nil
}

// buildContextBlock formats match context as a prompt preamble.
func buildContextBlock(mc *MatchContext) string {
	var sb strings.Builder
	sb.WriteString("## CURRENT SESSION CONTEXT\n")
	sb.WriteString(fmt.Sprintf("Phase: %s\n", mc.Phase))

	if mc.CandidateName != "" {
		sb.WriteString(fmt.Sprintf("Candidate Name: %s\n", mc.CandidateName))
	}
	if mc.CandidateBio != "" {
		sb.WriteString(fmt.Sprintf("Candidate Profile: %s\n", mc.CandidateBio))
	}
	if mc.JobTitle != "" {
		sb.WriteString(fmt.Sprintf("Job Title: %s\n", mc.JobTitle))
	}
	if mc.JobDesc != "" {
		sb.WriteString(fmt.Sprintf("Job Description: %s\n", truncateStr(mc.JobDesc, 300)))
	}
	if mc.SalaryAgreed != "" {
		sb.WriteString(fmt.Sprintf("Agreed Salary: %s\n", mc.SalaryAgreed))
	}
	if mc.MatchScore > 0 {
		sb.WriteString(fmt.Sprintf("Match Score: %.1f\n", mc.MatchScore))
	}
	return sb.String()
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// PhaseInstruction returns a phase-specific instruction fragment to append to any prompt.
func PhaseInstruction(phase string, round int) string {
	switch phase {
	case "INTRODUCTION":
		return "This is the introduction phase. Establish rapport and exchange basic information."
	case "NEGOTIATION":
		if round > 0 {
			return fmt.Sprintf("This is negotiation round %d. Negotiate salary and terms in good faith.", round)
		}
		return "This is the negotiation phase. Discuss salary and terms."
	case "INTERVIEW":
		if round > 0 {
			return fmt.Sprintf("This is interview round %d. Ask/answer relevant questions and evaluate fit.", round)
		}
		return "This is the interview phase. Evaluate technical and cultural fit."
	case "OFFER":
		return "This is the offer phase. Present, discuss, and decide on the job offer."
	default:
		return ""
	}
}
