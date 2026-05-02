package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PreferenceType defines types of user preferences
type PreferenceType string

const (
	PreferenceSalary  PreferenceType = "salary"
	PreferenceLocation PreferenceType = "location"
	PreferenceJobType  PreferenceType = "job_type"
	PreferenceSkills   PreferenceType = "skills"
)

// UserPreferenceVector stores vectorized user preferences in pgvector
type UserPreferenceVector struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID       `json:"user_id" gorm:"type:uuid;not null;index"`
	PreferenceType PreferenceType  `json:"preference_type" gorm:"type:varchar(20);not null;index"`
	Embedding     pq.Float64Array `json:"embedding" gorm:"type:float8[]"` // Vector embedding for similarity search
	RawValue       string           `json:"raw_value" gorm:"size:500"`    // Original text value (e.g., "London", "$100k+")
	CreatedAt      time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the default table name
func (UserPreferenceVector) TableName() string {
	return "user_preference_vectors"
}