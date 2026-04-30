package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RateLimitCounter struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	WindowStart   time.Time `gorm:"type:timestamp" json:"window_start"`
	MessageCount  int       `gorm:"default:0" json:"message_count"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (r *RateLimitCounter) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
