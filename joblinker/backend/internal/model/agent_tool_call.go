package model

import (
	"time"

	"github.com/google/uuid"
)

// ToolStatus represents the status of a tool call
type ToolStatus string

const (
	ToolStatusPending ToolStatus = "pending"
	ToolStatusSuccess ToolStatus = "success"
	ToolStatusFailed  ToolStatus = "failed"
)

// AgentToolCall records an API call made during agent reasoning
type AgentToolCall struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID     uuid.UUID  `json:"match_id" gorm:"type:uuid;not null;index"`
	ToolName    string     `json:"tool_name" gorm:"type:varchar(50);not null;index"` // query_jobs, get_candidate, etc.
	Arguments   string     `json:"arguments" gorm:"type:jsonb"`                       // JSON string of arguments
	Result      string     `json:"result" gorm:"type:text"`                          // JSON string of result
	Status      ToolStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	ExecutedAt  time.Time  `json:"executed_at" gorm:"autoCreateTime"`
}

// TableName overrides the default table name
func (AgentToolCall) TableName() string {
	return "agent_tool_calls"
}