package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConfirmationService struct {
	db        *gorm.DB
	matchRepo *repository.MatchRepository
}

func NewConfirmationService(
	db *gorm.DB,
	matchRepo *repository.MatchRepository,
) *ConfirmationService {
	return &ConfirmationService{
		db:        db,
		matchRepo: matchRepo,
	}
}

// CreateConfirmationRequest creates a final confirmation request for human approval
func (s *ConfirmationService) CreateConfirmationRequest(ctx context.Context, matchID uuid.UUID, confirmationType model.ConfirmationRequestType, payload map[string]interface{}) (*model.ConfirmationRequest, error) {
	payloadJSON, _ := json.Marshal(payload)

	// Get match to find the appropriate user
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	// Determine user ID based on confirmation type
	var userID uuid.UUID
	switch confirmationType {
	case model.ConfirmationTypeOffer:
		userID = match.SeekerAgentID
	case model.ConfirmationTypeInterview:
		userID = match.SeekerAgentID
	default:
		userID = match.SeekerAgentID
	}

	confReq := &model.ConfirmationRequest{
		ID:        uuid.New(),
		MatchID:   matchID,
		Type:      confirmationType,
		Payload:   string(payloadJSON),
		Status:    model.ConfirmationStatusPending,
		UserID:    userID,
	}

	// Store in database
	if err := s.db.Create(confReq).Error; err != nil {
		return nil, fmt.Errorf("failed to create confirmation request: %w", err)
	}

	log.Printf("Created confirmation request %s for match %s: type=%s", confReq.ID, matchID, confirmationType)
	return confReq, nil
}

// GetPendingConfirmations retrieves all pending confirmation requests for a user
func (s *ConfirmationService) GetPendingConfirmations(ctx context.Context, userID uuid.UUID) ([]*model.ConfirmationRequest, error) {
	var confirmations []*model.ConfirmationRequest
	if err := s.db.Where("user_id = ? AND status = ?", userID, model.ConfirmationStatusPending).Find(&confirmations).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending confirmations: %w", err)
	}
	return confirmations, nil
}

// ProcessConfirmation processes a human response to a confirmation request
func (s *ConfirmationService) ProcessConfirmation(ctx context.Context, confirmationID uuid.UUID, approved bool, feedback string) error {
	var confReq model.ConfirmationRequest
	if err := s.db.First(&confReq, "id = ?", confirmationID).Error; err != nil {
		return fmt.Errorf("confirmation request not found: %w", err)
	}

	now := time.Now()
	confReq.RespondedAt = &now
	confReq.Feedback = feedback

	if approved {
		confReq.Status = model.ConfirmationStatusApproved
		log.Printf("Confirmation %s approved by user", confirmationID)
	} else {
		confReq.Status = model.ConfirmationStatusRejected
		log.Printf("Confirmation %s rejected by user: %s", confirmationID, feedback)
	}

	return s.db.Save(&confReq).Error
}

// BlockMidNegotiationMessages is a placeholder for filtering mid-negotiation messages
func (s *ConfirmationService) BlockMidNegotiationMessages(ctx context.Context, userID uuid.UUID) error {
	log.Printf("Blocking mid-negotiation messages for user %s", userID)
	return nil
}

// GetConfirmationStatus returns the status of a confirmation request
func (s *ConfirmationService) GetConfirmationStatus(ctx context.Context, confirmationID uuid.UUID) (*model.ConfirmationRequest, error) {
	var confReq model.ConfirmationRequest
	if err := s.db.First(&confReq, "id = ?", confirmationID).Error; err != nil {
		return nil, fmt.Errorf("confirmation request not found: %w", err)
	}
	return &confReq, nil
}

// ListUserConfirmations lists all confirmation requests for a user with pagination
func (s *ConfirmationService) ListUserConfirmations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.ConfirmationRequest, error) {
	var confirmations []*model.ConfirmationRequest
	query := s.db.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&confirmations).Error; err != nil {
		return nil, fmt.Errorf("failed to list confirmations: %w", err)
	}
	return confirmations, nil
}