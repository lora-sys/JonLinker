package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleSeeker   UserRole = "seeker"
	RoleRecruiter UserRole = "recruiter"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email          string     `json:"email" gorm:"uniqueIndex;not null;size:255"`
	PasswordHash   string     `json:"-" gorm:"not null;size:255"`
	Role           UserRole   `json:"role" gorm:"type:varchar(20);not null"`
	TenantID       uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	OrganizationID *uuid.UUID `json:"organization_id" gorm:"type:uuid"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type AgentType string

const (
	AgentTypeSeeker   AgentType = "seeker"
	AgentTypeRecruiter AgentType = "recruiter"
)

type AgentStatus string

const (
	AgentStatusActive AgentStatus = "active"
	AgentStatusPaused AgentStatus = "paused"
)

type AgentFSMState string

const (
	AgentFSMIdle           AgentFSMState = "idle"
	AgentFSMIntro          AgentFSMState = "intro"
	AgentFSMInterest       AgentFSMState = "interest"
	AgentFSMNegotiation    AgentFSMState = "negotiation"
	AgentFSMOffer          AgentFSMState = "offer"
	AgentFSMConfirmed      AgentFSMState = "confirmed"
	AgentFSMRejected      AgentFSMState = "rejected"
)

type Agent struct {
	ID        uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	TenantID  uuid.UUID    `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Type      AgentType    `json:"type" gorm:"type:varchar(20);not null"`
	Status    AgentStatus  `json:"status" gorm:"type:varchar(20);not null;index"`
	FSMState  AgentFSMState `json:"fsm_state" gorm:"type:varchar(20);not null;default:'idle'"`
	ConfigJSON json.RawMessage `json:"config" gorm:"type:jsonb"`
	CreatedAt time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	User      *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ToolConfigJSON returns the tools configuration from ConfigJSON
func (a *Agent) ToolConfigJSON() (map[string]interface{}, error) {
	if len(a.ConfigJSON) == 0 || string(a.ConfigJSON) == "null" {
		return nil, nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(a.ConfigJSON, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// CanUseTool checks if the agent has permission to use a specific tool
func (a *Agent) CanUseTool(toolName string) bool {
	cfg, err := a.ToolConfigJSON()
	if err != nil || cfg == nil {
		return true // Default allow if no config
	}
	tools, ok := cfg["tools"].(map[string]interface{})
	if !ok {
		return true
	}
	toolPerm, ok := tools[toolName].(map[string]interface{})
	if !ok {
		return true
	}
	enabled, ok := toolPerm["enabled"].(bool)
	return enabled
}

// GetToolConfig returns the tool configuration for a specific tool
func (a *Agent) GetToolConfig(toolName string) (map[string]interface{}, bool) {
	cfg, err := a.ToolConfigJSON()
	if err != nil || cfg == nil {
		return nil, false
	}
	tools, ok := cfg["tools"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	toolPerm, ok := tools[toolName].(map[string]interface{})
	if !ok {
		return nil, false
	}
	return toolPerm, true
}

// GetCacheBudget returns the cache budget from config
func (a *Agent) GetCacheBudget() int {
	cfg, err := a.ToolConfigJSON()
	if err != nil || cfg == nil {
		return 50 // Default 50
	}
	if budget, ok := cfg["cache_budget"].(float64); ok {
		return int(budget)
	}
	return 50
}

type Organization struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null;size:255"`
	AdminUserID uuid.UUID `json:"admin_user_id" gorm:"type:uuid"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type JobStatus string

const (
	JobStatusDraft   JobStatus = "draft"
	JobStatusActive  JobStatus = "active"
	JobStatusPaused  JobStatus = "paused"
	JobStatusFilled  JobStatus = "filled"
	JobStatusClosed  JobStatus = "closed"
)

type Job struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AgentID      uuid.UUID `json:"agent_id" gorm:"type:uuid;not null;index"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	StructuredJSON json.RawMessage `json:"structured" gorm:"type:jsonb;not null"`
	VectorID     string    `json:"vector_id" gorm:"size:255"`
	Status       JobStatus `json:"status" gorm:"type:varchar(20);not null;index"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Agent        *Agent    `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

type Resume struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AgentID        uuid.UUID `json:"agent_id" gorm:"type:uuid;uniqueIndex;not null"`
	TenantID       uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	StructuredJSON json.RawMessage `json:"structured" gorm:"type:jsonb;not null"`
	VectorID       string    `json:"vector_id" gorm:"size:255"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Agent          *Agent    `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

type MatchStatus string

const (
	MatchStatusPending       MatchStatus = "pending"
	MatchStatusSearching     MatchStatus = "searching"
	MatchStatusMutualInterest MatchStatus = "mutual_interest"
	MatchStatusNegotiating   MatchStatus = "negotiating"
	MatchStatusInterviewing  MatchStatus = "interviewing"
	MatchStatusOffered       MatchStatus = "offered"
	MatchStatusHired         MatchStatus = "hired"
	MatchStatusRejected      MatchStatus = "rejected"
	MatchStatusPaused        MatchStatus = "paused"
)

type Match struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SeekerAgentID   uuid.UUID   `json:"seeker_agent_id" gorm:"type:uuid;not null;index"`
	RecruiterAgentID *uuid.UUID `json:"recruiter_agent_id,omitempty" gorm:"type:uuid;index"`
	JobID           uuid.UUID   `json:"job_id" gorm:"type:uuid;not null;index"`
	TenantID        uuid.UUID   `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Score           float64     `json:"score" gorm:"type:decimal(5,4);not null"`
	Status          MatchStatus `json:"status" gorm:"type:varchar(30);not null;index;default:'pending'"`
	CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	SeekerAgent     *Agent      `json:"seeker_agent,omitempty" gorm:"foreignKey:SeekerAgentID"`
	RecruiterAgent  *Agent      `json:"recruiter_agent,omitempty" gorm:"foreignKey:RecruiterAgentID"`
	Job             *Job        `json:"job,omitempty" gorm:"foreignKey:JobID"`
}

type Message struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID      uuid.UUID `json:"match_id" gorm:"type:uuid;not null;index"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	SenderAgentID uuid.UUID `json:"sender_agent_id" gorm:"type:uuid;not null"`
	ContentXML   string    `json:"content_xml" gorm:"type:text;not null"`
	IntentType   string    `json:"intent_type" gorm:"size:50"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	Match        *Match    `json:"match,omitempty" gorm:"foreignKey:MatchID"`
}

type InterviewFormat string

const (
	InterviewFormatVideo   InterviewFormat = "video"
	InterviewFormatPhone   InterviewFormat = "phone"
	InterviewFormatOnsite  InterviewFormat = "onsite"
)

type InterviewStatus string

const (
	InterviewStatusScheduled   InterviewStatus = "scheduled"
	InterviewStatusConfirmed   InterviewStatus = "confirmed"
	InterviewStatusCompleted   InterviewStatus = "completed"
	InterviewStatusCancelled   InterviewStatus = "cancelled"
	InterviewStatusRescheduled InterviewStatus = "rescheduled"
)

type Interview struct {
	ID           uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID      uuid.UUID        `json:"match_id" gorm:"type:uuid;uniqueIndex;not null"`
	TenantID     uuid.UUID        `json:"tenant_id" gorm:"type:uuid;not null;index"`
	ScheduledAt  time.Time        `json:"scheduled_at" gorm:"not null"`
	Format       InterviewFormat  `json:"format" gorm:"type:varchar(20);not null"`
	Location     string           `json:"location" gorm:"size:500"`
	Status       InterviewStatus `json:"status" gorm:"type:varchar(20);not null"`
	Feedback     json.RawMessage  `json:"feedback" gorm:"type:jsonb;default:null"`
	ReminderSent bool             `json:"reminder_sent" gorm:"default:false"`
	CreatedAt    time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	Match        *Match           `json:"match,omitempty" gorm:"foreignKey:MatchID"`
}

type OfferStatus string

const (
	OfferStatusPending    OfferStatus = "pending"
	OfferStatusAccepted   OfferStatus = "accepted"
	OfferStatusDeclined   OfferStatus = "declined"
	OfferStatusNegotiating OfferStatus = "negotiating"
	OfferStatusWithdrawn   OfferStatus = "withdrawn"
)

type Offer struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID          uuid.UUID       `json:"match_id" gorm:"type:uuid;uniqueIndex;not null"`
	TenantID         uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompensationJSON json.RawMessage `json:"compensation" gorm:"type:jsonb;not null"`
	StartDate        time.Time       `json:"start_date" gorm:"type:date;not null"`
	Status           OfferStatus     `json:"status" gorm:"type:varchar(20);not null"`
	ExpiresAt        *time.Time      `json:"expires_at,omitempty"`
	RespondedAt      *time.Time      `json:"responded_at"`
	CreatedAt        time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	Match            *Match          `json:"match,omitempty" gorm:"foreignKey:MatchID"`
}

type SecurityEvent struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ActionType  string    `json:"action_type" gorm:"size:50;not null;index"`
	DetailsJSON string    `json:"details" gorm:"type:jsonb"`
	IPAddress   string    `json:"ip_address" gorm:"size:45"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}
