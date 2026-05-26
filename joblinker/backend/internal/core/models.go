package core

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
)
