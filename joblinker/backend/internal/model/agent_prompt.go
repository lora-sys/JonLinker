package model

import (
	"time"

	"github.com/google/uuid"
)

// PromptScenarioType defines possible scenario types
type PromptScenarioType string

const (
	ScenarioGreeting     PromptScenarioType = "greeting"
	ScenarioNegotiation  PromptScenarioType = "negotiation"
	ScenarioSalary       PromptScenarioType = "salary"
	ScenarioInterview     PromptScenarioType = "interview"
	ScenarioOffer        PromptScenarioType = "offer"
	ScenarioDecline      PromptScenarioType = "decline"
	ScenarioTermination  PromptScenarioType = "termination"
)

// AgentPrompt represents a structured prompt template for agents
type AgentPrompt struct {
	ID        uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Role      AgentType       `json:"role" gorm:"type:varchar(20);not null;index"` // Uses AgentType from user.go
	Scenario  PromptScenarioType `json:"scenario" gorm:"type:varchar(20)"`
	MustDo    string          `json:"must_do" gorm:"type:text;not null"`
	MustNotDo string          `json:"must_not_do" gorm:"type:text;not null"`
	Behavior  string          `json:"behavior" gorm:"type:text;not null"`
	CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// PromptScenario represents scenario-specific prompt fragments
type PromptScenario struct {
	ID                uuid.UUID           `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Scenario          PromptScenarioType `json:"scenario" gorm:"type:varchar(20);not null;index"`
	Description       string              `json:"description" gorm:"size:500"`
	TriggerConditions string             `json:"trigger_conditions" gorm:"type:text"`
	PromptFragment    string              `json:"prompt_fragment" gorm:"type:text;not null"`
	Priority          int                 `json:"priority" gorm:"default:0"`
	CreatedAt         time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
}