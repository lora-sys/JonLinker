package model

import (
	"time"
)

// SessionSummary stores compression snapshots of a conversation.
type SessionSummary struct {
	SessionID    string    `json:"session_id" gorm:"primaryKey;size:128"`
	Round        int       `json:"round" gorm:"primaryKey;autoIncrement:false"`
	SummaryText  string    `json:"summary_text" gorm:"type:text;not null"`
	KeyFactsJSON string    `json:"key_facts_json" gorm:"type:jsonb"`
	TokenCount   int       `json:"token_count" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the default table name.
func (SessionSummary) TableName() string {
	return "session_summaries"
}
