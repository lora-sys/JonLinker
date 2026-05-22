package stategraph

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	// emitPhase helper: called before each phase's messages
	emitPhase := func(phase Phase) {
		cfg.emitPhase(phase)
	}

	// emitNewMessages broadcasts any messages added since the given count.
	emitNewMessages := func(prevCount int) {
		for i := prevCount; i < len(state.Messages); i++ {
			cfg.emitMessage(state.Messages[i], state.Phase)
		}
	}

	// ── Node 1: INTRODUCTION (single pass) ──
	state.Phase = PhaseIntroduction
	emitPhase(PhaseIntroduction)
	prevMsgCount := len(state.Messages)
	state, err := g.introductionNode(ctx, state)
	if err != nil {
		return state, fmt.Errorf("introduction node: %w", err)
	}
	emitNewMessages(prevMsgCount)
	log.Printf("[Stategraph] Introduction complete for match=%s", matchID)

	// ── Node 2: NEGOTIATION (loop until done or max rounds) ──
	state.Phase = PhaseNegotiation
	emitPhase(PhaseNegotiation)
	maxNeg := cfg.maxNegotiationRounds
	if maxNeg <= 0 {
		maxNeg = 10
	}
	for round := 1; round <= maxNeg; round++ {
		if state.NegotiationDone {
			break
		}
		prevCount := len(state.Messages)
		state, err = g.negotiationNode(ctx, state)
		if err != nil {
			return state, fmt.Errorf("negotiation node (round %d): %w", round, err)
		}
		emitNewMessages(prevCount)
	}
	log.Printf("[Stategraph] Negotiation complete for match=%s (done=%v, rounds=%d)", matchID, state.NegotiationDone, len(state.NegotiationRounds))

	// ── Node 3: INTERVIEW (loop until passed or max rounds) ──
	state.Phase = PhaseInterview
	emitPhase(PhaseInterview)
	maxIV := cfg.maxInterviewRounds
	if maxIV <= 0 {
		maxIV = 5
	}
	for round := 1; round <= maxIV; round++ {
		if state.InterviewPassed {
			break
		}
		prevCount := len(state.Messages)
		state, err = g.interviewNode(ctx, state)
		if err != nil {
			return state, fmt.Errorf("interview node (round %d): %w", round, err)
		}
		emitNewMessages(prevCount)
	}
	log.Printf("[Stategraph] Interview complete for match=%s (passed=%v, rounds=%d)", matchID, state.InterviewPassed, len(state.InterviewRounds))

	// ── Node 4: OFFER (loop until accepted/declined or max rounds) ──
	state.Phase = PhaseOffer
	emitPhase(PhaseOffer)
	maxOffer := cfg.maxOfferRounds
	if maxOffer <= 0 {
		maxOffer = 5
	}
	for round := 1; round <= maxOffer; round++ {
		if state.OfferAccepted || state.OfferDeclined {
			break
		}
		prevCount := len(state.Messages)
		state, err = g.offerNode(ctx, state)
		if err != nil {
			return state, fmt.Errorf("offer node (round %d): %w", round, err)
		}
		emitNewMessages(prevCount)
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
	// Seeker introduces themselves with actual profile context
	profileCtx := ""
	if state.CandidateName != "" {
		profileCtx = fmt.Sprintf("Your name is %s.", state.CandidateName)
	}
	if state.CandidateBio != "" {
		profileCtx = fmt.Sprintf("Your background: %s", state.CandidateBio)
	}
	seekerPrompt := fmt.Sprintf(
		"[Phase: INTRODUCTION]\n%s\n\nIntroduce yourself as a job seeker. State your name, background, and what kind of role you're looking for.",
		profileCtx,
	)
	seekerResp, err := g.seeker.Chat(ctx, nil, seekerPrompt)
	if err != nil {
		return state, fmt.Errorf("seeker introduction: %w", err)
	}
	state.Messages = append(state.Messages, "[Seeker] "+seekerResp)
	state.CandidateName = extractName(seekerResp)

	// Recruiter introduces the job with actual job context
	jobCtx := ""
	if state.JobTitle != "" {
		jobCtx = fmt.Sprintf("The position is: %s.", state.JobTitle)
	}
	if state.JobDesc != "" {
		desc := state.JobDesc
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		jobCtx = fmt.Sprintf("The position is: %s. Description: %s", state.JobTitle, desc)
	}
	recruiterPrompt := fmt.Sprintf(
		"[Phase: INTRODUCTION]\n%s\n\nIntroduce the job position. State the job title, key responsibilities, and what the company is looking for.",
		jobCtx,
	)
	recruiterResp, err := g.recruiter.Chat(ctx, nil, recruiterPrompt)
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
	onMessage            func(sender string, message string, phase Phase)
	onPhase              func(phase Phase)
}

// RunOption configures graph execution.
type RunOption func(*runConfig)

// WithMaxNegotiationRounds sets the maximum negotiation loop iterations.
func WithMaxNegotiationRounds(n int) RunOption {
	return func(c *runConfig) {
		c.maxNegotiationRounds = n
	}
}

// WithMaxInterviewRounds sets the maximum interview loop iterations.
func WithMaxInterviewRounds(n int) RunOption {
	return func(c *runConfig) {
		c.maxInterviewRounds = n
	}
}

// WithMaxOfferRounds sets the maximum offer loop iterations.
func WithMaxOfferRounds(n int) RunOption {
	return func(c *runConfig) {
		c.maxOfferRounds = n
	}
}

// WithCallbacks sets streaming callbacks for real-time message and phase-change events.
// onMessage is called for each message generated by a graph node.
// onPhase is called when the graph transitions to a new phase.
func WithCallbacks(onMessage func(sender string, message string, phase Phase), onPhase func(phase Phase)) RunOption {
	return func(c *runConfig) {
		c.onMessage = onMessage
		c.onPhase = onPhase
	}
}

// emitMessage calls the onMessage callback if set. It parses the raw "[Sender] text" format
// from state.Messages into separate sender and text fields.
func (cfg *runConfig) emitMessage(rawMsg string, phase Phase) {
	if cfg.onMessage == nil {
		return
	}
	sender, text := parseMessage(rawMsg)
	cfg.onMessage(sender, text, phase)
}

// emitPhase calls the onPhase callback if set.
func (cfg *runConfig) emitPhase(phase Phase) {
	if cfg.onPhase == nil {
		return
	}
	cfg.onPhase(phase)
}

// parseMessage splits a raw "[Sender] text" message into sender and text.
// If the message doesn't have bracket prefix, sender defaults to "system".
func parseMessage(raw string) (sender, text string) {
	if len(raw) > 0 && raw[0] == '[' {
		if idx := strings.Index(raw, "] "); idx > 0 {
			return raw[1:idx], raw[idx+2:]
		}
	}
	return "system", raw
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
