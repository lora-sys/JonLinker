package core

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// This file holds core domain model types that have zero external dependencies.
// They are shared across the application via the core package.
//
// Database-backed models (gorm) remain in the model/ package.
// Only pure domain types used by the engine/ and adapters/ layers go here.

// UserRole represents the type of user in the system.
type UserRole string

const (
	UserRoleSeeker    UserRole = "seeker"
	UserRoleRecruiter UserRole = "recruiter"
	UserRoleAdmin     UserRole = "admin"
)

// AgentType represents the type of agent.
type AgentType string

const (
	AgentTypeSeeker    AgentType = "seeker"
	AgentTypeRecruiter AgentType = "recruiter"
)

// MatchStatus represents the current status of a match between agents.
type MatchStatus string

const (
	MatchStatusPending        MatchStatus = "pending"
	MatchStatusSearching      MatchStatus = "searching"
	MatchStatusMutualInterest MatchStatus = "mutual_interest"
	MatchStatusNegotiating    MatchStatus = "negotiating"
	MatchStatusInterviewing   MatchStatus = "interviewing"
	MatchStatusOffered        MatchStatus = "offered"
	MatchStatusHired          MatchStatus = "hired"
	MatchStatusRejected       MatchStatus = "rejected"
	MatchStatusPaused         MatchStatus = "paused"
	MatchStatusWithdrawn      MatchStatus = "withdrawn"
	MatchStatusConfirmed      MatchStatus = "confirmed"
	MatchStatusAccepted       MatchStatus = "accepted"
)

// InterviewFormat represents the format of an interview.
type InterviewFormat string

const (
	InterviewFormatVideo  InterviewFormat = "video"
	InterviewFormatPhone  InterviewFormat = "phone"
	InterviewFormatOnsite InterviewFormat = "onsite"
)

// InterviewStatus represents the status of an interview.
type InterviewStatus string

const (
	InterviewStatusScheduled InterviewStatus = "scheduled"
	InterviewStatusCompleted InterviewStatus = "completed"
	InterviewStatusCancelled InterviewStatus = "cancelled"
)

// OfferStatus represents the status of a job offer.
type OfferStatus string

const (
	OfferStatusPending  OfferStatus = "pending"
	OfferStatusAccepted OfferStatus = "accepted"
	OfferStatusDeclined OfferStatus = "declined"
)

// ToolCallStatus represents the status of a tool execution.
type ToolCallStatus string

const (
	ToolStatusSuccess ToolCallStatus = "success"
	ToolStatusFailed  ToolCallStatus = "failed"
)

// Job represents a job listing in the domain.
type Job struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	CreatedAt     time.Time
	StructuredJSON json.RawMessage
}

// User represents a user in the domain.
type User struct {
	ID    uuid.UUID
	Email string
}

// Agent represents an agent (recruiter or seeker).
type Agent struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	User       *User
	ConfigJSON json.RawMessage
}

// Offer represents a job offer.
type Offer struct {
	ID               uuid.UUID
	MatchID          uuid.UUID
	TenantID         uuid.UUID
	CompensationJSON json.RawMessage
	StartDate        time.Time
	Status           OfferStatus
}

// Interview represents a scheduled interview.
type Interview struct {
	ID          uuid.UUID
	MatchID     uuid.UUID
	TenantID    uuid.UUID
	ScheduledAt time.Time
	Format      InterviewFormat
	Status      InterviewStatus
}

// Match represents a match between a job seeker and a recruiter.
type Match struct {
	ID      uuid.UUID
	TenantID uuid.UUID
	Job     *Job
	Status  MatchStatus
}

// AgentToolCall represents a recorded tool execution.
type AgentToolCall struct {
	ID        uuid.UUID
	MatchID   uuid.UUID
	ToolName  string
	Arguments string
	Result    string
	Status    ToolCallStatus
}
