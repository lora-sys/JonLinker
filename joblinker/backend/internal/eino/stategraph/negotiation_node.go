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
		"This is negotiation round %d. State your expected salary range for the %s position. Be reasonable based on your skills and market value.\n\n"+
			"Examples of good responses:\n"+
			"- \"I'm looking for $80,000-$90,000 based on my 5 years of experience.\"\n"+
			"- \"Your offer of $75,000 is reasonable. I'd accept $80,000.\"\n"+
			"- \"After reviewing the benefits package, I'm happy with $85,000.\"\n\n"+
			"At the end of your response, include exactly one of these markers on its own line:\n"+
			"  [AGREED] if you accept the latest offer and want to finalize the deal\n"+
			"  [COUNTER] if you want to propose a different amount and continue negotiating",
		roundNum,
		state.JobTitle,
	)
	seekerResp, err := g.seeker.Chat(ctx, nil, seekerPrompt)
	if err != nil {
		return state, fmt.Errorf("seeker negotiation round %d: %w", roundNum, err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Seeker Negotiation R%d] %s", roundNum, seekerResp))

	// Recruiter responds with offer/counter
	recruiterPrompt := fmt.Sprintf(
		"This is negotiation round %d. Respond to the candidate's salary expectation for the %s position. "+
			"The candidate previously said: \"%s\"\n\n"+
			"Examples of appropriate recruiter responses:\n"+
			"- \"We can offer $82,000 which is within your expected range. Welcome aboard!\" + [AGREED]\n"+
			"- \"Our budget caps at $78,000. Can we meet at $78,000 with an extra week of vacation?\" + [COUNTER]\n"+
			"- \"$85,000 is above our range. Our best offer is $80,000.\" + [COUNTER $80000]\n\n"+
			"Policies:\n"+
			"- If this is a late round (round 8+), you should aim to reach agreement rather than prolonging.\n"+
			"- If the candidate's request is within the company budget, agree and finalize.\n"+
			"- If not, make a reasonable counter-offer that's fair for both sides.\n\n"+
			"At the end of your response, include exactly one of these markers on its own line:\n"+
			"  [AGREED] if you accept the candidate's expectation and finalize the deal\n"+
			"  [COUNTER $AMOUNT] if you want to negotiate further, replacing $AMOUNT with your offer",
		roundNum,
		state.JobTitle,
		seekerResp,
	)
	recruiterResp, err := g.recruiter.Chat(ctx, nil, recruiterPrompt)
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
	// Check for structured [AGREED] marker (from prompt instructions)
	if contains(seekerResp, "[AGREED]") || contains(recruiterResp, "[AGREED]") {
		return true
	}

	// Simple keyword check (placeholder) — both sides must express agreement
	agreeKeywords := []string{"agree", "accept", "deal", "confirmed", "sounds good", "works for me", "同意", "接受", "可以", "成交", "没问题"}
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
