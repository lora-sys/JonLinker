package model

import (
	"time"

	"github.com/google/uuid"
)

// KeyFactType defines types of key facts preserved during compression
type KeyFactType string

const (
	KeyFactSalary          KeyFactType = "salary_expectation"
	KeyFactLocation        KeyFactType = "preferred_location"
	KeyFactJobType         KeyFactType = "job_type"
	KeyFactSkills          KeyFactType = "skills"
	KeyFactRejectedTitles  KeyFactType = "rejected_titles"
)

// KeyFact represents a critical piece of information to preserve
type KeyFact struct {
	FactType KeyFactType `json:"fact_type"`
	Value    string      `json:"value"`
	Priority int         `json:"priority"` // Higher = more important to preserve
}

// ConversationSummary represents a compressed conversation state
type ConversationSummary struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID     uuid.UUID `json:"match_id" gorm:"type:uuid;not null;index"`
	SummaryText string    `json:"summary_text" gorm:"type:text;not null"`
	KeyFacts    []KeyFact `json:"key_facts" gorm:"-"`
	KeyFactsJSON string   `json:"key_facts_json" gorm:"type:jsonb"`
	TokenCount  int       `json:"token_count" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the default table name
func (ConversationSummary) TableName() string {
	return "conversation_summaries"
}