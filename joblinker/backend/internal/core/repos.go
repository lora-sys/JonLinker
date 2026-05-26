package core

import (
	"context"

	"github.com/google/uuid"
)

// JobRepository defines the interface for job data access.
type JobRepository interface {
	Search(ctx context.Context, query string, skills []string, location string, salaryMin int, jobType string, limit int) ([]*Job, error)
}

// AgentRepository defines the interface for agent data access.
type AgentRepository interface {
	GetByID(id uuid.UUID) (*Agent, error)
	SearchBySkills(ctx context.Context, skills []string, location string, experienceMin int, limit int) ([]*Agent, error)
}

// MatchRepository defines the interface for match data access.
type MatchRepository interface {
	GetByID(id uuid.UUID) (*Match, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status MatchStatus) error
	CreateToolCall(ctx context.Context, toolCall *AgentToolCall) error
}

// OfferRepository defines the interface for offer data access.
type OfferRepository interface {
	Create(offer *Offer) error
}

// InterviewRepository defines the interface for interview data access.
type InterviewRepository interface {
	Create(interview *Interview) error
}
