package stategraph

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// offerNode runs one round of offer presentation and response between recruiter and seeker.
func (g *RecruitmentGraph) offerNode(ctx context.Context, state *RecruitmentState) (*RecruitmentState, error) {
	roundNum := 1
	if state.OfferMade {
		roundNum = 2 // second attempt
	}

	// Recruiter presents the offer
	var recruiterPrompt string
	if !state.OfferMade {
		recruiterPrompt = fmt.Sprintf(
			"The interview is complete and the candidate passed. Present a formal job offer for the %s position. "+
				"Include the agreed salary (%s), start date expectations, and key benefits. "+
				"Ask the candidate to accept or decline.",
			state.JobTitle,
			state.SalaryAgreed,
		)
	} else {
		recruiterPrompt = fmt.Sprintf(
			"The candidate hasn't decided yet. Reiterate the offer for the %s position with salary %s. "+
				"Address any concerns and ask for a final decision.",
			state.JobTitle,
			state.SalaryAgreed,
		)
	}
	recruiterResp, err := g.recruiter.Chat(ctx, recruiterPrompt)
	if err != nil {
		return state, fmt.Errorf("recruiter offer round %d: %w", roundNum, err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Recruiter Offer R%d] %s", roundNum, recruiterResp))

	// For Phase 1, we simulate the candidate's response via seeker agent
	// In production, this would be a human-in-the-loop confirmation
	seekerPrompt := fmt.Sprintf(
		"You have received a job offer for the %s position. "+
			"The details were: %s. "+
			"Decide whether to accept or decline the offer. State your decision clearly.",
		state.JobTitle,
		recruiterResp,
	)
	seekerResp, err := g.seeker.Chat(ctx, seekerPrompt)
	if err != nil {
		return state, fmt.Errorf("seeker offer response: %w", err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Seeker Offer R%d] %s", roundNum, seekerResp))

	// Also store in seeker agent's memory for context
	if err := g.seeker.AddMemoryMessage(ctx, schema.Assistant, recruiterResp); err != nil {
		return state, fmt.Errorf("store recruiter offer in seeker memory: %w", err)
	}
	if err := g.seeker.AddMemoryMessage(ctx, schema.User, seekerResp); err != nil {
		return state, fmt.Errorf("store seeker response in seeker memory: %w", err)
	}

	state.OfferMade = true

	// Detect accept/decline
	upper := toUpper(seekerResp)
	if contains(upper, "ACCEPT") || contains(upper, "ACCEPTED") || contains(upper, "YES") {
		state.OfferAccepted = true
	} else if contains(upper, "DECLINE") || contains(upper, "DECLINED") || contains(upper, "NO") || contains(upper, "REJECT") {
		state.OfferDeclined = true
	}

	return state, nil
}

func toUpper(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		b[i] = c
	}
	return string(b)
}
