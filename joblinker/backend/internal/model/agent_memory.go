package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MemoryType defines whether memory is short-term or long-term
type MemoryType string

const (
	MemoryTypeShortTerm MemoryType = "short_term"
	MemoryTypeLongTerm  MemoryType = "long_term"
)

// AgentMemory stores memory entries for agent context
type AgentMemory struct {
	ID            uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID       uuid.UUID   `json:"match_id" gorm:"type:uuid;not null;index"`
	MemoryType    MemoryType  `json:"memory_type" gorm:"type:varchar(20);not null;index"`
	Content       string      `json:"content" gorm:"type:text;not null"`
	Embedding     pq.Float64Array `json:"embedding" gorm:"type:float8[]"` // For long-term recall
	RecallCount   int         `json:"recall_count" gorm:"default:0"`
	LastRecalledAt *time.Time `json:"last_recalled_at"`
	CreatedAt     time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the default table name
func (AgentMemory) TableName() string {
	return "agent_memories"
}