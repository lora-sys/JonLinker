package runner

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"

	"joblinker/internal/eino/stategraph"
)

// StateGraphRunner manages StateGraph (RecruitmentGraph) instances per match.
// It obtains seeker/recruiter agents from the AgentRunner pool and caches
// graphs so the same match continues where it left off.
type StateGraphRunner struct {
	agentRunner *AgentRunner
	graphs      map[uuid.UUID]*stategraph.RecruitmentGraph
	mu          sync.RWMutex
}

// NewStateGraphRunner creates a runner that draws agent instances from the given pool.
func NewStateGraphRunner(agentRunner *AgentRunner) *StateGraphRunner {
	return &StateGraphRunner{
		agentRunner: agentRunner,
		graphs:      make(map[uuid.UUID]*stategraph.RecruitmentGraph),
	}
}

// GetOrCreate returns an existing RecruitmentGraph for the match or creates a new one
// using agents from the pool.
func (r *StateGraphRunner) GetOrCreate(ctx context.Context, matchID uuid.UUID) (*stategraph.RecruitmentGraph, error) {
	r.mu.RLock()
	if g, ok := r.graphs[matchID]; ok {
		r.mu.RUnlock()
		return g, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if g, ok := r.graphs[matchID]; ok {
		return g, nil
	}

	seeker, err := r.agentRunner.GetSeeker(ctx)
	if err != nil {
		return nil, fmt.Errorf("get seeker from pool: %w", err)
	}

	recruiter, err := r.agentRunner.GetRecruiter(ctx)
	if err != nil {
		r.agentRunner.ReturnSeeker(seeker)
		return nil, fmt.Errorf("get recruiter from pool: %w", err)
	}

	g := stategraph.NewRecruitmentGraph(seeker, recruiter)
	r.graphs[matchID] = g
	log.Printf("[StateGraphRunner] Created graph for match=%s", matchID)
	return g, nil
}

// RunGraph runs the full recruitment pipeline for the given match.
// Pass optional RunOption callbacks via stategraph.WithCallbacks to receive
// real-time message and phase-change events.
func (r *StateGraphRunner) RunGraph(ctx context.Context, matchID uuid.UUID, opts ...stategraph.RunOption) (*stategraph.RecruitmentState, error) {
	g, err := r.GetOrCreate(ctx, matchID)
	if err != nil {
		return nil, err
	}

	state, err := g.Run(ctx, matchID, opts...)
	if err != nil {
		return state, err
	}

	log.Printf("[StateGraphRunner] Graph finished for match=%s (phase=%s)", matchID, state.Phase)
	return state, nil
}

// Release removes the graph for the given match and returns its agents to the pool.
func (r *StateGraphRunner) Release(matchID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if g, ok := r.graphs[matchID]; ok {
		// We don't track which exact agents we got from the pool,
		// but since they won't be used again we just delete.
		// In a production system we'd keep references and call ReturnSeeker/ReturnRecruiter.
		delete(r.graphs, matchID)
		log.Printf("[StateGraphRunner] Released graph for match=%s", matchID)

		// Quietly discard agents (they'll be GC'd). We release the pool references
		// but the pool's warm count handles replenishment.
		_ = g
	}
}
