package stategraph

import (
	"context"
	"fmt"
)

// negotiationNode runs one round of salary negotiation between seeker and recruiter.
func (g *RecruitmentGraph) negotiationNode(ctx context.Context, state *RecruitmentState) (*RecruitmentState, error) {
	roundNum := len(state.NegotiationRounds) + 1

	// Seeker states salary expectation
	seekerPrompt := fmt.Sprintf(
		"This is negotiation round %d. State your expected salary range for the %s position. Be reasonable based on your skills.",
		roundNum,
		state.JobTitle,
	)
	seekerResp, err := g.seeker.Chat(ctx, seekerPrompt)
	if err != nil {
		return state, fmt.Errorf("seeker negotiation round %d: %w", roundNum, err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Seeker Negotiation R%d] %s", roundNum, seekerResp))

	// Recruiter responds with offer/counter
	recruiterPrompt := fmt.Sprintf(
		"This is negotiation round %d. Respond to the candidate's salary expectation for the %s position. "+
			"If acceptable, confirm the agreed salary. If not, make a counter-offer.",
		roundNum,
		state.JobTitle,
	)
	recruiterResp, err := g.recruiter.Chat(ctx, recruiterPrompt)
	if err != nil {
		return state, fmt.Errorf("recruiter negotiation round %d: %w", roundNum, err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Recruiter Negotiation R%d] %s", roundNum, recruiterResp))

	// Record round
	round := NegotiationRound{
		Round:          roundNum,
		SeekerExpect:   seekerResp,
		RecruiterOffer: recruiterResp,
		Agreed:         detectAgreement(seekerResp, recruiterResp),
	}
	state.NegotiationRounds = append(state.NegotiationRounds, round)

	if round.Agreed {
		state.NegotiationDone = true
		state.SalaryAgreed = recruiterResp
	}

	return state, nil
}

// detectAgreement uses simple heuristics to check if both sides agreed.
// In Phase 1 this is a placeholder — replace with intent extraction.
func detectAgreement(seekerResp, recruiterResp string) bool {
	// Simple keyword check (placeholder)
	agreeKeywords := []string{"agree", "accept", "deal", "confirmed", "sounds good", "works for me"}
	for _, kw := range agreeKeywords {
		if contains(seekerResp, kw) && contains(recruiterResp, kw) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsLower(s, substr)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			tc := substr[j]
			if tc >= 'A' && tc <= 'Z' {
				tc += 32
			}
			if sc != tc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
