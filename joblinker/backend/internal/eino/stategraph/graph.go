package stategraph

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"joblinker/internal/eino/agent"
)

// RecruitmentGraph is the Phase 1 pipeline runner.
// It orchestrates the 4-phase flow: Introduction → Negotiation → Interview → Offer → Completed.
// Each phase is a node that calls seeker/recruiter ADK agents internally.
type RecruitmentGraph struct {
	seeker    *agent.SeekerAgent
	recruiter *agent.RecruiterAgent
}

// NewRecruitmentGraph creates a graph with the given ADK agents.
func NewRecruitmentGraph(seeker *agent.SeekerAgent, recruiter *agent.RecruiterAgent) *RecruitmentGraph {
	return &RecruitmentGraph{
		seeker:    seeker,
		recruiter: recruiter,
	}
}

// Run executes the full graph from INTRODUCTION to COMPLETED (or until an error terminates early).
// It returns the final state after all reachable nodes have been processed.
func (g *RecruitmentGraph) Run(ctx context.Context, matchID uuid.UUID, opts ...RunOption) (*RecruitmentState, error) {
	state := NewRecruitmentState(matchID)

	cfg := &runConfig{}
	for _, o := range opts {
		o(cfg)
	}

	log.Printf("[Stategraph] Starting graph for match=%s", matchID)

	// ── Node 1: INTRODUCTION (single pass) ──
	state.Phase = PhaseIntroduction
	state, err := g.introductionNode(ctx, state)
	if err != nil {
		return state, fmt.Errorf("introduction node: %w", err)
	}
	log.Printf("[Stategraph] Introduction complete for match=%s", matchID)

	// ── Node 2: NEGOTIATION (loop until done or max rounds) ──
	state.Phase = PhaseNegotiation
	maxNeg := cfg.maxNegotiationRounds
	if maxNeg <= 0 {
		maxNeg = 10
	}
	for round := 1; round <= maxNeg; round++ {
		if state.NegotiationDone {
			break
		}
		state, err = g.negotiationNode(ctx, state)
		if err != nil {
			return state, fmt.Errorf("negotiation node (round %d): %w", round, err)
		}
	}
	log.Printf("[Stategraph] Negotiation complete for match=%s (done=%v, rounds=%d)", matchID, state.NegotiationDone, len(state.NegotiationRounds))

	// ── Node 3: INTERVIEW (loop until passed or max rounds) ──
	state.Phase = PhaseInterview
	maxIV := cfg.maxInterviewRounds
	if maxIV <= 0 {
		maxIV = 5
	}
	for round := 1; round <= maxIV; round++ {
		if state.InterviewPassed {
			break
		}
		state, err = g.interviewNode(ctx, state)
		if err != nil {
			return state, fmt.Errorf("interview node (round %d): %w", err)
		}
	}
	log.Printf("[Stategraph] Interview complete for match=%s (passed=%v, rounds=%d)", matchID, state.InterviewPassed, len(state.InterviewRounds))

	// ── Node 4: OFFER (loop until accepted/declined or max rounds) ──
	state.Phase = PhaseOffer
	maxOffer := cfg.maxOfferRounds
	if maxOffer <= 0 {
		maxOffer = 5
	}
	for round := 1; round <= maxOffer; round++ {
		if state.OfferAccepted || state.OfferDeclined {
			break
		}
		state, err = g.offerNode(ctx, state)
		if err != nil {
			return state, fmt.Errorf("offer node (round %d): %w", err)
		}
	}
	log.Printf("[Stategraph] Offer complete for match=%s (accepted=%v, declined=%v)", matchID, state.OfferAccepted, state.OfferDeclined)

	// ── Node 5: COMPLETED ──
	state, err = g.completedNode(ctx, state)
	if err != nil {
		return state, fmt.Errorf("completed node: %w", err)
	}
	log.Printf("[Stategraph] Graph completed for match=%s (final phase=%s)", matchID, state.Phase)

	return state, nil
}

// --- graph node handlers (internal) ---

// introductionNode introduces the candidate and job to each other.
func (g *RecruitmentGraph) introductionNode(ctx context.Context, state *RecruitmentState) (*RecruitmentState, error) {
	// Seeker introduces themselves
	seekerPrompt := "Introduce yourself as a job seeker. State your name, background, and what kind of role you're looking for."
	seekerResp, err := g.seeker.Chat(ctx, seekerPrompt)
	if err != nil {
		return state, fmt.Errorf("seeker introduction: %w", err)
	}
	state.Messages = append(state.Messages, "[Seeker] "+seekerResp)
	state.CandidateName = extractName(seekerResp)

	// Recruiter introduces the job
	recruiterPrompt := "Introduce the job position. State the job title, key responsibilities, and what the company is looking for."
	recruiterResp, err := g.recruiter.Chat(ctx, recruiterPrompt)
	if err != nil {
		return state, fmt.Errorf("recruiter introduction: %w", err)
	}
	state.Messages = append(state.Messages, "[Recruiter] "+recruiterResp)
	state.JobTitle = extractJobTitle(recruiterResp)
	state.JobDesc = recruiterResp

	return state, nil
}

// --- run options ---

type runConfig struct {
	maxNegotiationRounds int
	maxInterviewRounds   int
	maxOfferRounds       int
}

// RunOption configures graph execution.
type RunOption func(*runConfig)

// WithMaxNegotiationRounds sets the maximum negotiation loop iterations.
func WithMaxNegotiationRounds(n int) RunOption {
	return func(c *runConfig) {
		c.maxNegotiationRounds = n
	}
}

// simple extractors (stubs — replace with proper NLP later)

func extractName(text string) string {
	if len(text) > 60 {
		return text[:60] + "..."
	}
	return text
}

func extractJobTitle(text string) string {
	if len(text) > 60 {
		return text[:60] + "..."
	}
	return text
}
