package stategraph

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// completedNode finalises the recruitment process and generates a summary.
func (g *RecruitmentGraph) completedNode(ctx context.Context, state *RecruitmentState) (*RecruitmentState, error) {
	state.Phase = PhaseCompleted

	var summary string
	switch {
	case state.OfferAccepted:
		summary = fmt.Sprintf(
			"Match %s completed successfully. %s accepted the offer for %s at %s.",
			state.MatchID, state.CandidateName, state.JobTitle, state.SalaryAgreed,
		)
	case state.OfferDeclined:
		summary = fmt.Sprintf(
			"Match %s ended. %s declined the offer for %s.",
			state.MatchID, state.CandidateName, state.JobTitle,
		)
	case state.InterviewPassed && !state.OfferMade:
		summary = fmt.Sprintf(
			"Match %s paused after interview. %s passed but no offer was made for %s.",
			state.MatchID, state.CandidateName, state.JobTitle,
		)
	default:
		summary = fmt.Sprintf(
			"Match %s ended. Final phase: %s.",
			state.MatchID, state.Phase,
		)
	}

	state.Messages = append(state.Messages, "[System] "+summary)

	// Store in both agents' memory for future reference
	if g.seeker != nil {
		_ = g.seeker.AddMemoryMessage(ctx, schema.Assistant, "[Summary] "+summary)
	}
	if g.recruiter != nil {
		_ = g.recruiter.AddMemoryMessage(ctx, schema.Assistant, "[Summary] "+summary)
	}

	return state, nil
}
