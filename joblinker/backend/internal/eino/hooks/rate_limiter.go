package hooks

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// ---- per-match token bucket rate limiter ----

type matchBucket struct {
	tokens    float64
	lastCheck time.Time
}

// RateLimiter is an ADK ChatModelAgentMiddleware that enforces per-match rate limits
// on agent model invocations.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*matchBucket
	rate     float64 // tokens per second
	burst    int     // max burst
	maxTurns int     // max total turns per match (0 = unlimited)
	turns    map[string]int
}

// NewRateLimiter creates a rate limiter with the given rate (tokens/sec) and burst.
// maxTurns caps total agent iterations per match (0 = unlimited).
func NewRateLimiter(rate float64, burst, maxTurns int) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*matchBucket),
		rate:     rate,
		burst:    burst,
		maxTurns: maxTurns,
		turns:    make(map[string]int),
	}
}

func (r *RateLimiter) WrapInvokableToolCall(_ context.Context, endpoint adk.InvokableToolCallEndpoint, _ *adk.ToolContext) (adk.InvokableToolCallEndpoint, error) {
	return endpoint, nil
}

func (r *RateLimiter) WrapStreamableToolCall(_ context.Context, endpoint adk.StreamableToolCallEndpoint, _ *adk.ToolContext) (adk.StreamableToolCallEndpoint, error) {
	return endpoint, nil
}

func (r *RateLimiter) WrapEnhancedInvokableToolCall(_ context.Context, endpoint adk.EnhancedInvokableToolCallEndpoint, _ *adk.ToolContext) (adk.EnhancedInvokableToolCallEndpoint, error) {
	return endpoint, nil
}

func (r *RateLimiter) WrapEnhancedStreamableToolCall(_ context.Context, endpoint adk.EnhancedStreamableToolCallEndpoint, _ *adk.ToolContext) (adk.EnhancedStreamableToolCallEndpoint, error) {
	return endpoint, nil
}

func (r *RateLimiter) WrapModel(_ context.Context, m model.BaseChatModel, _ *adk.ModelContext) (model.BaseChatModel, error) {
	return m, nil
}

// BeforeAgent checks rate limits before each agent iteration.
// If the rate limit is exceeded, it injects a throttle instruction.
func (r *RateLimiter) BeforeAgent(ctx context.Context, c *adk.ChatModelAgentContext) (context.Context, *adk.ChatModelAgentContext, error) {
	matchID, _ := ctx.Value(CtxKeyMatchID).(string)
	if matchID == "" {
		return ctx, c, nil
	}

	if r.maxTurns > 0 {
		r.mu.Lock()
		turnCount := r.turns[matchID]
		r.mu.Unlock()
		if turnCount >= r.maxTurns {
			c.Instruction = "[RATE LIMIT] Maximum conversation turns reached for this match.\n" + c.Instruction
			return ctx, c, nil
		}
	}

	if r.rate > 0 {
		r.mu.Lock()
		b, ok := r.buckets[matchID]
		now := time.Now()
		if !ok {
			b = &matchBucket{tokens: float64(r.burst), lastCheck: now}
			r.buckets[matchID] = b
		}
		// Refill
		elapsed := now.Sub(b.lastCheck).Seconds()
		b.tokens += elapsed * r.rate
		if b.tokens > float64(r.burst) {
			b.tokens = float64(r.burst)
		}
		b.lastCheck = now

		if b.tokens < 1.0 {
			r.mu.Unlock()
			c.Instruction = "[RATE LIMIT] Please wait before sending more requests.\n" + c.Instruction
			return ctx, c, nil
		}
		b.tokens--
		r.mu.Unlock()
	}

	if r.maxTurns > 0 {
		r.mu.Lock()
		r.turns[matchID]++
		r.mu.Unlock()
	}

	return ctx, c, nil
}

func (r *RateLimiter) BeforeModelRewriteState(ctx context.Context, state *adk.ChatModelAgentState, _ *adk.ModelContext) (context.Context, *adk.ChatModelAgentState, error) {
	return ctx, state, nil
}

func (r *RateLimiter) AfterModelRewriteState(ctx context.Context, state *adk.ChatModelAgentState, _ *adk.ModelContext) (context.Context, *adk.ChatModelAgentState, error) {
	return ctx, state, nil
}

func (r *RateLimiter) AfterAgent(ctx context.Context, state *adk.ChatModelAgentState, action *adk.AgentAction) (context.Context, *adk.ChatModelAgentState, *adk.AgentAction, error) {
	return ctx, state, action, nil
}

// ResetMatch clears rate limit state for a match (e.g., when match completes).
func (r *RateLimiter) ResetMatch(matchID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, matchID)
	delete(r.turns, matchID)
}
