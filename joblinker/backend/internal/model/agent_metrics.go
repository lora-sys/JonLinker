package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgentMetrics struct {
	ID                      uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	AgentID                 uuid.UUID `gorm:"type:uuid;index" json:"agent_id"`
	AgentType               string    `gorm:"size:20;index" json:"agent_type"` // "seeker" | "recruiter"
	MessagesProcessed       int       `gorm:"default:0" json:"messages_processed"`
	ErrorsCount             int       `gorm:"default:0" json:"errors_count"`
	AvgResponseTimeMs       int       `gorm:"default:0" json:"avg_response_time_ms"`
	ConversationsActive     int       `gorm:"default:0" json:"conversations_active"`
	ConversationsCompleted  int       `gorm:"default:0" json:"conversations_completed"`
	Metadata                JSONMap   `gorm:"type:jsonb" json:"metadata"` // Flexible metadata for extensibility
	LastHeartbeat           time.Time `gorm:"index" json:"last_heartbeat"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func (AgentMetrics) TableName() string {
	return "agent_metrics"
}

func (a *AgentMetrics) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.LastHeartbeat.IsZero() {
		a.LastHeartbeat = time.Now()
	}
	return nil
}

type AuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	AgentID   uuid.UUID `gorm:"type:uuid;index" json:"agent_id"`
	MatchID   uuid.UUID `gorm:"type:uuid;index" json:"match_id"`
	EventType string    `gorm:"size:50;index" json:"event_type"` // "message_sent" | "tool_call" | "decision" | "error"
	EventData JSONMap   `gorm:"type:jsonb" json:"event_data"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Timestamp.IsZero() {
		a.Timestamp = time.Now()
	}
	return nil
}

type ErrorEvent struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ErrorType    string    `gorm:"size:50;index" json:"error_type"` // "ai_failure" | "rabbitmq_failure" | "db_timeout" | "tool_error"
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	StackTrace   string    `gorm:"type:text" json:"stack_trace,omitempty"`
	Context      JSONMap   `gorm:"type:jsonb" json:"context"`
	Resolved     bool      `gorm:"default:false;index" json:"resolved"`
	Timestamp    time.Time `gorm:"index" json:"timestamp"`
	CreatedAt    time.Time `json:"created_at"`
}

func (ErrorEvent) TableName() string {
	return "error_events"
}

func (e *ErrorEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	return nil
}

// JSONMap is a helper type for JSONB columns
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface for JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONMap value: %v", value)
	}
	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface for JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}