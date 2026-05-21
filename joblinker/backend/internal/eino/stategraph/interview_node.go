package stategraph

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// interviewNode runs one round of interview Q&A between recruiter (asks) and seeker (answers).
func (g *RecruitmentGraph) interviewNode(ctx context.Context, state *RecruitmentState) (*RecruitmentState, error) {
	roundNum := len(state.InterviewRounds) + 1

	// Recruiter asks a question
	recruiterPrompt := fmt.Sprintf(
		"This is interview round %d for the %s position. Ask the candidate a relevant technical or behavioral question. "+
			"Include a question number so the candidate can reference it.",
		roundNum,
		state.JobTitle,
	)
	question, err := g.recruiter.Chat(ctx, nil, recruiterPrompt)
	if err != nil {
		return state, fmt.Errorf("recruiter interview round %d: %w", roundNum, err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Recruiter Interview R%d] %s", roundNum, question))

	// Seeker answers
	seekerPrompt := fmt.Sprintf(
		"This is interview round %d. Answer the recruiter's question about %s position. Be thorough and professional.",
		roundNum,
		state.JobTitle,
	)
	answer, err := g.seeker.Chat(ctx, nil, seekerPrompt)
	if err != nil {
		return state, fmt.Errorf("seeker interview round %d: %w", roundNum, err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Seeker Interview R%d] %s", roundNum, answer))

	// Recruiter evaluates the answer
	evalPrompt := fmt.Sprintf(
		"Evaluate the candidate's answer to your question. Rate it 1-10 and say 'PASS' if the candidate should proceed. "+
			"Question: %s\nAnswer: %s",
		question,
		answer,
	)
	evalResp, err := g.recruiter.Chat(ctx, nil, evalPrompt)
	if err != nil {
		return state, fmt.Errorf("recruiter evaluation round %d: %w", roundNum, err)
	}
	state.Messages = append(state.Messages, fmt.Sprintf("[Recruiter Evaluation R%d] %s", roundNum, evalResp))

	// Record round
	round := InterviewRound{
		Round:     roundNum,
		Question:  question,
		Answer:    answer,
		Evaluated: true,
		Score:     extractScore(evalResp),
		CreatedAt: time.Now(),
	}
	state.InterviewRounds = append(state.InterviewRounds, round)

	// Check if passed (score >= 6 or contains "PASS")
	if round.Score >= 6 || strings.Contains(strings.ToUpper(evalResp), "PASS") {
		state.InterviewPassed = true
	}

	return state, nil
}

// extractScore attempts to parse a numeric score from the evaluation text.
func extractScore(text string) int {
	// Look for patterns like "score: 8", "8/10", "rating: 7"
	patterns := []string{"score:", "rating:", "Rate:"}
	for _, p := range patterns {
		idx := strings.Index(strings.ToLower(text), strings.ToLower(p))
		if idx >= 0 {
			after := text[idx+len(p):]
			after = strings.TrimSpace(after)
			if len(after) > 0 {
				numStr := strings.Split(after, " ")[0]
				numStr = strings.Split(numStr, "/")[0]
				numStr = strings.TrimSpace(numStr)
				if score, err := strconv.Atoi(numStr); err == nil && score >= 1 && score <= 10 {
					return score
				}
			}
		}
	}
	return 5 // default middle score if parsing fails
}
