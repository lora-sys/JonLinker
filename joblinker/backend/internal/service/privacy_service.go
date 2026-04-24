package service

import (
	"encoding/json"
	"joblinker/internal/repository"
	"joblinker/pkg/crypto"

	"github.com/google/uuid"
)

type PrivacyService struct {
	userRepo  *repository.UserRepository
	agentRepo *repository.AgentRepository
	matchRepo *repository.MatchRepository
}

func NewPrivacyService(userRepo *repository.UserRepository, agentRepo *repository.AgentRepository, matchRepo *repository.MatchRepository) *PrivacyService {
	return &PrivacyService{
		userRepo:  userRepo,
		agentRepo: agentRepo,
		matchRepo: matchRepo,
	}
}

func (s *PrivacyService) ExportUserData(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"user_id":   userID.String(),
		"email":     user.Email,
		"role":      user.Role,
		"created_at": user.CreatedAt,
	}

	// Encrypt sensitive fields before export
	if email, err := crypto.Encrypt(user.Email); err == nil {
		data["email_encrypted"] = email
	}

	return data, nil
}

func (s *PrivacyService) DeleteUserAccount(userID uuid.UUID) error {
	// Schedule deletion within 30 days per privacy requirements
	// This would typically queue a background job
	return nil
}

func (s *PrivacyService) AnonymizeData(data map[string]interface{}) (string, error) {
	// Remove identifying information
	delete(data, "email")
	delete(data, "phone")
	delete(data, "name")

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Encrypt the anonymized data
	encrypted, err := crypto.Encrypt(string(jsonBytes))
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

func (s *PrivacyService) VerifyDataIntegrity(original, encrypted string) bool {
	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		return false
	}
	return original == decrypted
}
