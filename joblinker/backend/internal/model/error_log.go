package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ErrorLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID       *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	MatchID      *uuid.UUID `gorm:"type:uuid" json:"match_id,omitempty"`
	ErrorType    string     `gorm:"type:varchar(50)" json:"error_type"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	StackTrace   string     `gorm:"type:text" json:"stack_trace,omitempty"`
	CorrelationID string    `gorm:"type:varchar(50);index" json:"correlation_id"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (e *ErrorLog) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
