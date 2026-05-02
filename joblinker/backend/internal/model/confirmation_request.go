package model

import (
	"time"

	"github.com/google/uuid"
)

// ConfirmationRequestType defines types of confirmation requests
type ConfirmationRequestType string

const (
	ConfirmationTypeOffer     ConfirmationRequestType = "offer"
	ConfirmationTypeInterview ConfirmationRequestType = "interview"
)

// ConfirmationStatus represents the status of a confirmation request
type ConfirmationStatus string

const (
	ConfirmationStatusPending   ConfirmationStatus = "pending"
	ConfirmationStatusApproved   ConfirmationStatus = "approved"
	ConfirmationStatusRejected   ConfirmationStatus = "rejected"
	ConfirmationStatusExpired    ConfirmationStatus = "expired"
)

// ConfirmationRequest represents a human final confirmation request
type ConfirmationRequest struct {
	ID          uuid.UUID               `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID     uuid.UUID               `json:"match_id" gorm:"type:uuid;not null;index"`
	Type        ConfirmationRequestType `json:"type" gorm:"type:varchar(20);not null"`
	Payload     string                  `json:"payload" gorm:"type:jsonb"` // Offer or interview details
	Status      ConfirmationStatus      `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	UserID      uuid.UUID               `json:"user_id" gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time               `json:"created_at" gorm:"autoCreateTime"`
	RespondedAt *time.Time              `json:"responded_at"`
	Feedback    string                   `json:"feedback" gorm:"type:text"` // Rejection feedback if any
}

// TableName overrides the default table name
func (ConfirmationRequest) TableName() string {
	return "confirmation_requests"
}