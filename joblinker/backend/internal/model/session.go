package model

import (
	"time"

	"github.com/google/uuid"
)

// SessionConclusionReason indicates why a session was concluded.
type SessionConclusionReason string

const (
	SessionConcludedNaturally SessionConclusionReason = "naturally"
	SessionConcludedPaused    SessionConclusionReason = "paused"
)

// SessionMeta stores session-level metadata for a single conversation version.
type SessionMeta struct {
	SessionID        string                    `json:"session_id" gorm:"primaryKey;size:128"`
	MatchID          uuid.UUID                 `json:"match_id" gorm:"type:uuid;not null;index"`
	Version          int                       `json:"version" gorm:"not null;default:1"`
	Status           string                    `json:"status" gorm:"type:varchar(20);not null;index;default:'active'"`
	State            string                    `json:"state" gorm:"type:varchar(30);not null;default:'idle'"`
	ConclusionReason *SessionConclusionReason  `json:"conclusion_reason,omitempty" gorm:"type:varchar(20)"`
	MatchStatus      string                    `json:"match_status" gorm:"type:varchar(30)"`
	CreatedAt        time.Time                 `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time                 `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the default table name.
func (SessionMeta) TableName() string {
	return "session_meta"
}

// SessionMessage stores individual messages within a session.
type SessionMessage struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID string    `json:"session_id" gorm:"index;size:128;not null"`
	Seq       int       `json:"seq" gorm:"not null"`
	Role      string    `json:"role" gorm:"type:varchar(20);not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the default table name.
func (SessionMessage) TableName() string {
	return "session_messages"
}
