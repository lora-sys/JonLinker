// Package stategraph defines the Phase 1 StateGraph for the JobLinker recruitment pipeline.
//
// The graph has 4 nodes (INTRODUCTION → NEGOTIATION → INTERVIEW → OFFER → COMPLETED)
// with loop-back edges on NEGOTIATION, INTERVIEW, and OFFER nodes.
package stategraph

import (
	"time"

	"github.com/google/uuid"
)

// Phase represents the current node position in the recruitment graph.
type Phase string

const (
	PhaseIntroduction  Phase = "INTRODUCTION"
	PhaseNegotiation   Phase = "NEGOTIATION"
	PhaseInterview     Phase = "INTERVIEW"
	PhaseOffer         Phase = "OFFER"
	PhaseCompleted     Phase = "COMPLETED"
)

// InterviewRound captures one round of interview Q&A.
type InterviewRound struct {
	Round     int       `json:"round"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Evaluated bool      `json:"evaluated"`
	Score     int       `json:"score"` // 1-10, 0 if not yet evaluated
	CreatedAt time.Time `json:"created_at"`
}

// NegotiationRound captures one round of salary negotiation.
type NegotiationRound struct {
	Round         int    `json:"round"`
	SeekerExpect  string `json:"seeker_expect"`  // expected salary range from seeker
	RecruiterOffer string `json:"recruiter_offer"` // offered amount from recruiter
	Agreed        bool   `json:"agreed"`
}

// RecruitmentState is the shared state that flows through graph nodes.
type RecruitmentState struct {
	MatchID      uuid.UUID `json:"match_id"`
	Phase        Phase     `json:"phase"`

	// Introduction
	CandidateName string `json:"candidate_name,omitempty"`
	CandidateBio  string `json:"candidate_bio,omitempty"`
	JobTitle      string `json:"job_title,omitempty"`
	JobDesc       string `json:"job_desc,omitempty"`

	// Negotiation
	NegotiationRounds []NegotiationRound `json:"negotiation_rounds,omitempty"`
	NegotiationDone   bool               `json:"negotiation_done"`
	SalaryAgreed      string             `json:"salary_agreed,omitempty"`

	// Interview
	InterviewRounds []InterviewRound `json:"interview_rounds,omitempty"`
	InterviewPassed bool             `json:"interview_passed"`

	// Offer
	OfferMade    bool   `json:"offer_made"`
	OfferAmount  string `json:"offer_amount,omitempty"`
	OfferAccepted bool  `json:"offer_accepted"`
	OfferDeclined bool  `json:"offer_declined"`

	// Conversation messages accumulated across all nodes
	Messages []string `json:"messages,omitempty"`

	// Error tracking
	LastError string `json:"last_error,omitempty"`
}

// NewRecruitmentState creates a fresh state for the given match.
func NewRecruitmentState(matchID uuid.UUID) *RecruitmentState {
	return &RecruitmentState{
		MatchID:           matchID,
		Phase:             PhaseIntroduction,
		NegotiationRounds: make([]NegotiationRound, 0),
		InterviewRounds:   make([]InterviewRound, 0),
		Messages:          make([]string, 0),
	}
}

// Clone returns a shallow copy (slices are still shared — OK for read); use for immutable-style updates.
func (s *RecruitmentState) Clone() *RecruitmentState {
	c := *s
	return &c
}
