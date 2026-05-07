package service

import (
	"encoding/json"
	"fmt"
	"log"

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
	// Verify user exists
	if _, err := s.userRepo.GetByID(userID); err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Get user's agents
	agents, err := s.agentRepo.ListByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user agents: %w", err)
	}

	// Delete matches for each agent
	for _, agent := range agents {
		matches, _ := s.matchRepo.ListByAgentIDs([]uuid.UUID{agent.ID})
		for _, match := range matches {
			if err := s.matchRepo.Delete(match.ID); err != nil {
				log.Printf("Failed to delete match %s: %v", match.ID, err)
			}
		}
		// Delete agent
		if err := s.agentRepo.Delete(agent.ID); err != nil {
			log.Printf("Failed to delete agent %s: %v", agent.ID, err)
		}
	}

	// Delete user
	if err := s.userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

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
