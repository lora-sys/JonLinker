package service

import (
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"encoding/json"

	"github.com/google/uuid"
)

type SecurityService struct {
	repo *repository.SecurityEventRepository
}

func NewSecurityService(repo *repository.SecurityEventRepository) *SecurityService {
	return &SecurityService{repo: repo}
}

func (s *SecurityService) LogEvent(userID uuid.UUID, actionType string, details map[string]interface{}, ipAddress string) error {
	detailsJSON, _ := json.Marshal(details)
	event := &model.SecurityEvent{
		UserID:     userID,
		ActionType: actionType,
		DetailsJSON: string(detailsJSON),
		IPAddress:  ipAddress,
	}
	return s.repo.Create(event)
}
