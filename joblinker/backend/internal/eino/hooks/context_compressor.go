package hooks

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// CompressorConfig controls compression behaviour.
type CompressorConfig struct {
	// MaxModelCalls is the agent iteration count that triggers a "compress old history" reminder.
	MaxModelCalls int
}

// DefaultCompressorConfig returns sensible defaults for A2A conversations.
func DefaultCompressorConfig() *CompressorConfig {
	return &CompressorConfig{
		MaxModelCalls: 20,
	}
}

// ContextCompressor tracks model-call frequency per match and injects
// a compression instruction into the agent when the conversation gets long.
type ContextCompressor struct {
	cfg    *CompressorConfig
	mu     sync.Mutex
	counts map[string]int // matchID → model-call count
}

func NewContextCompressor(cfg *CompressorConfig) *ContextCompressor {
	if cfg == nil {
		cfg = DefaultCompressorConfig()
	}
	return &ContextCompressor{
		cfg:    cfg,
		counts: make(map[string]int),
	}
}

func (h *ContextCompressor) WrapInvokableToolCall(_ context.Context, endpoint adk.InvokableToolCallEndpoint, _ *adk.ToolContext) (adk.InvokableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextCompressor) WrapStreamableToolCall(_ context.Context, endpoint adk.StreamableToolCallEndpoint, _ *adk.ToolContext) (adk.StreamableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextCompressor) WrapEnhancedInvokableToolCall(_ context.Context, endpoint adk.EnhancedInvokableToolCallEndpoint, _ *adk.ToolContext) (adk.EnhancedInvokableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextCompressor) WrapEnhancedStreamableToolCall(_ context.Context, endpoint adk.EnhancedStreamableToolCallEndpoint, _ *adk.ToolContext) (adk.EnhancedStreamableToolCallEndpoint, error) {
	return endpoint, nil
}

func (h *ContextCompressor) WrapModel(_ context.Context, m model.BaseChatModel, _ *adk.ModelContext) (model.BaseChatModel, error) {
	return m, nil
}

// BeforeAgent injects a compression reminder when the conversation is long.
func (h *ContextCompressor) BeforeAgent(ctx context.Context, c *adk.ChatModelAgentContext) (context.Context, *adk.ChatModelAgentContext, error) {
	matchID, _ := ctx.Value(CtxKeyMatchID).(string)
	if matchID == "" {
		return ctx, c, nil
	}

	h.mu.Lock()
	h.counts[matchID]++
	count := h.counts[matchID]
	h.mu.Unlock()

	if count > h.cfg.MaxModelCalls && count%5 == 0 {
		c.Instruction = "[CONTEXT] The conversation history is long. " +
			"Keep responses concise and reference key points from earlier messages. " +
			"Focus on the most relevant information.\n" + c.Instruction
	}

	return ctx, c, nil
}

func (h *ContextCompressor) BeforeModelRewriteState(_ context.Context, state *adk.ChatModelAgentState, _ *adk.ModelContext) (context.Context, *adk.ChatModelAgentState, error) {
	return nil, state, nil
}

func (h *ContextCompressor) AfterModelRewriteState(_ context.Context, state *adk.ChatModelAgentState, _ *adk.ModelContext) (context.Context, *adk.ChatModelAgentState, error) {
	return nil, state, nil
}

func (h *ContextCompressor) AfterAgent(_ context.Context, state *adk.ChatModelAgentState, action *adk.AgentAction) (context.Context, *adk.ChatModelAgentState, *adk.AgentAction, error) {
	return nil, state, action, nil
}

// ResetMatch clears compression state for a match.
func (h *ContextCompressor) ResetMatch(matchID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.counts, matchID)
}
