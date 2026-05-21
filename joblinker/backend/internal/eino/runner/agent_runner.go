package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"joblinker/internal/eino/agent"
	"joblinker/pkg/ai"
)

// PoolConfig holds runner configuration
type PoolConfig struct {
	MaxAgents     int           // Maximum agents per type
	MaxIdleTime   time.Duration // Idle time before agent is recycled
	MinAgents     int           // Minimum agents to keep warm
	AgentTimeout  time.Duration // Max time for agent operation
}

// DefaultPoolConfig returns sensible defaults
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxAgents:    100,
		MaxIdleTime:  5 * time.Minute,
		MinAgents:    5,
		AgentTimeout: 120 * time.Second,
	}
}

// AgentRunner manages a pool of reusable agents
type AgentRunner struct {
	config      *PoolConfig
	aiClient    *ai.Client
	seekerPool  chan *agent.SeekerAgent
	recruiterPool chan *agent.RecruiterAgent
}

// NewAgentRunner creates a new agent runner with pool management
func NewAgentRunner(aiClient *ai.Client, config *PoolConfig) *AgentRunner {
	if config == nil {
		config = DefaultPoolConfig()
	}

	r := &AgentRunner{
		config:        config,
		aiClient:     aiClient,
		seekerPool:   make(chan *agent.SeekerAgent, config.MaxAgents),
		recruiterPool: make(chan *agent.RecruiterAgent, config.MaxAgents),
	}

	// Pre-warm the pools
	r.warmPools()

	return r
}

// warmPools pre-allocates minimum agents
func (r *AgentRunner) warmPools() {
	for i := 0; i < r.config.MinAgents; i++ {
		r.seekerPool <- agent.NewSeekerAgent(r.aiClient)
		r.recruiterPool <- agent.NewRecruiterAgent(r.aiClient)
	}
}

// GetSeeker gets or creates a seeker agent from the pool
func (r *AgentRunner) GetSeeker(ctx context.Context) (*agent.SeekerAgent, error) {
	select {
	case seeker := <-r.seekerPool:
		return seeker, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Pool empty, create new if under limit
		return agent.NewSeekerAgent(r.aiClient), nil
	}
}

// ReturnSeeker returns a seeker agent to the pool
func (r *AgentRunner) ReturnSeeker(seeker *agent.SeekerAgent) {
	if seeker == nil {
		return
	}
	select {
	case r.seekerPool <- seeker:
		// Successfully returned to pool
	default:
		// Pool full, let it be garbage collected
		log.Printf("Seeker pool full, discarding agent")
	}
}

// GetRecruiter gets or creates a recruiter agent from the pool
func (r *AgentRunner) GetRecruiter(ctx context.Context) (*agent.RecruiterAgent, error) {
	select {
	case recruiter := <-r.recruiterPool:
		return recruiter, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return agent.NewRecruiterAgent(r.aiClient), nil
	}
}

// ReturnRecruiter returns a recruiter agent to the pool
func (r *AgentRunner) ReturnRecruiter(recruiter *agent.RecruiterAgent) {
	if recruiter == nil {
		return
	}
	select {
	case r.recruiterPool <- recruiter:
	default:
		log.Printf("Recruiter pool full, discarding agent")
	}
}

// RunSeekerTask executes a task with a seeker agent
func (r *AgentRunner) RunSeekerTask(ctx context.Context, msg string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.AgentTimeout)
	defer cancel()

	seeker, err := r.GetSeeker(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get seeker: %w", err)
	}
	defer r.ReturnSeeker(seeker)

	return seeker.Chat(ctx, nil, msg)
}

// RunRecruiterTask executes a task with a recruiter agent
func (r *AgentRunner) RunRecruiterTask(ctx context.Context, msg string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.AgentTimeout)
	defer cancel()

	recruiter, err := r.GetRecruiter(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get recruiter: %w", err)
	}
	defer r.ReturnRecruiter(recruiter)

	return recruiter.Chat(ctx, nil, msg)
}

// PoolStats returns current pool statistics
type PoolStats struct {
	SeekerPoolSize   int
	RecruiterPoolSize int
}

// Stats returns current pool statistics
func (r *AgentRunner) Stats() *PoolStats {
	return &PoolStats{
		SeekerPoolSize:   len(r.seekerPool),
		RecruiterPoolSize: len(r.recruiterPool),
	}
}

// Close drains agent pools
func (r *AgentRunner) Close() {
	close(r.seekerPool)
	close(r.recruiterPool)
}
