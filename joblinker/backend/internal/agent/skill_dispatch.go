package agent

import (
	"context"
	"fmt"
)

type Skill string

const (
	SkillJobSearch       Skill = "job_search"
	SkillCandidateSearch Skill = "candidate_search"
	SkillNegotiation     Skill = "negotiation"
	SkillScheduling      Skill = "scheduling"
	SkillResumeParsing   Skill = "resume_parsing"
	SkillOfferGeneration Skill = "offer_generation"
	SkillInterviewPrep   Skill = "interview_prep"
)

type SkillHandler interface {
	CanHandle(intent string) bool
	Execute(ctx context.Context, msg *Message) (*Message, error)
}

type SkillDispatch struct {
	handlers map[Skill]SkillHandler
}

func NewSkillDispatch() *SkillDispatch {
	return &SkillDispatch{
		handlers: make(map[Skill]SkillHandler),
	}
}

func (sd *SkillDispatch) RegisterHandler(skill Skill, handler SkillHandler) {
	sd.handlers[skill] = handler
}

func (sd *SkillDispatch) Dispatch(ctx context.Context, skill Skill, msg *Message) (*Message, error) {
	handler, exists := sd.handlers[skill]
	if !exists {
		return nil, fmt.Errorf("no handler registered for skill: %s", skill)
	}
	return handler.Execute(ctx, msg)
}

func (sd *SkillDispatch) RouteIntent(intent string) Skill {
	switch intent {
	case IntentIntroduction:
		return SkillJobSearch
	case IntentInterest:
		return SkillCandidateSearch
	case IntentNegotiation:
		return SkillNegotiation
	case IntentSchedule:
		return SkillScheduling
	case IntentOffer:
		return SkillOfferGeneration
	default:
		return SkillJobSearch
	}
}

type DefaultSkillHandler struct{}

func (h *DefaultSkillHandler) CanHandle(intent string) bool {
	return true
}

func (h *DefaultSkillHandler) Execute(ctx context.Context, msg *Message) (*Message, error) {
	// Default handler just echoes the message back
	return msg, nil
}

type JobSearchHandler struct{}

func (h *JobSearchHandler) CanHandle(intent string) bool {
	return intent == IntentIntroduction || intent == IntentInquiry
}

func (h *JobSearchHandler) Execute(ctx context.Context, msg *Message) (*Message, error) {
	// Process job search intent
	// In production, this would search the vector database and job listings
	return msg, nil
}

type CandidateSearchHandler struct{}

func (h *CandidateSearchHandler) CanHandle(intent string) bool {
	return intent == IntentInterest
}

func (h *CandidateSearchHandler) Execute(ctx context.Context, msg *Message) (*Message, error) {
	// Process candidate search intent
	return msg, nil
}

type NegotiationHandler struct{}

func (h *NegotiationHandler) CanHandle(intent string) bool {
	return intent == IntentNegotiation || intent == IntentOffer
}

func (h *NegotiationHandler) Execute(ctx context.Context, msg *Message) (*Message, error) {
	// Process negotiation intent
	return msg, nil
}

type SchedulingHandler struct{}

func (h *SchedulingHandler) CanHandle(intent string) bool {
	return intent == IntentSchedule
}

func (h *SchedulingHandler) Execute(ctx context.Context, msg *Message) (*Message, error) {
	// Process scheduling intent
	return msg, nil
}
