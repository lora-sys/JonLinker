package model

import (
	"time"

	"github.com/google/uuid"
)

// NegotiationStatus represents the status of a negotiation session
type NegotiationStatus string

const (
	NegotiationStatusActive     NegotiationStatus = "active"
	NegotiationStatusAgreed    NegotiationStatus = "agreed"
	NegotiationStatusFailed     NegotiationStatus = "failed"
	NegotiationStatusCancelled NegotiationStatus = "cancelled"
)

// NegotiationSession tracks dual-agent autonomous negotiation
type NegotiationSession struct {
	ID            uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID       uuid.UUID         `json:"match_id" gorm:"type:uuid;not null;index"`
	SeekerAgentID uuid.UUID         `json:"seeker_agent_id" gorm:"type:uuid;not null"`
	RecruiterAgentID uuid.UUID      `json:"recruiter_agent_id" gorm:"type:uuid;not null"`
	Status        NegotiationStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CurrentRound  int               `json:"current_round" gorm:"default:0"`
	MaxRounds     int               `json:"max_rounds" gorm:"default:20"`
	AgreedTerms   string            `json:"agreed_terms" gorm:"type:jsonb"` // Final agreed terms
	StartedAt     time.Time         `json:"started_at" gorm:"autoCreateTime"`
	CompletedAt   *time.Time        `json:"completed_at"`
}

// TableName overrides the default table name
func (NegotiationSession) TableName() string {
	return "negotiation_sessions"
}