package service

import (
	"encoding/json"
	"log"

	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

type AgentService struct {
	agentRepo   *repository.AgentRepository
	userRepo    *repository.UserRepository
	securityRepo *repository.SecurityEventRepository
}

func NewAgentService(agentRepo *repository.AgentRepository, userRepo *repository.UserRepository, securityRepo *repository.SecurityEventRepository) *AgentService {
	return &AgentService{
		agentRepo:   agentRepo,
		userRepo:    userRepo,
		securityRepo: securityRepo,
	}
}

func (s *AgentService) CreateAgent(userID uuid.UUID, agentType model.AgentType, configJSON string) (*model.Agent, error) {
	if configJSON == "" {
		configJSON = "null"
	}
	agent := &model.Agent{
		UserID:     userID,
		Type:       agentType,
		Status:     model.AgentStatusActive,
		FSMState:   model.AgentFSMIdle,
		ConfigJSON: configJSON,
	}
	if err := s.agentRepo.Create(agent); err != nil {
		return nil, err
	}
	s.logSecurityEvent(userID, "agent_create", map[string]interface{}{"agent_id": agent.ID.String(), "type": agentType})
	return agent, nil
}

func (s *AgentService) GetAgent(id uuid.UUID) (*model.Agent, error) {
	return s.agentRepo.GetByID(id)
}

func (s *AgentService) UpdateAgent(id uuid.UUID, updates map[string]interface{}) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if status, ok := updates["status"].(string); ok {
		agent.Status = model.AgentStatus(status)
	}
	if config, ok := updates["config"].(string); ok {
		agent.ConfigJSON = config
	}
	if err := s.agentRepo.Update(agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *AgentService) DeleteAgent(id uuid.UUID) error {
	return s.agentRepo.Delete(id)
}

func (s *AgentService) ListUserAgents(userID uuid.UUID) ([]*model.Agent, error) {
	return s.agentRepo.ListByUserID(userID, "")
}

func (s *AgentService) UpdateStatus(id uuid.UUID, status model.AgentStatus) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	oldStatus := agent.Status
	agent.Status = status
	if err := s.agentRepo.Update(agent); err != nil {
		return nil, err
	}

	actionType := "agent_resume"
	if status == model.AgentStatusPaused {
		actionType = "agent_pause"
	}
	s.logSecurityEvent(agent.UserID, actionType, map[string]interface{}{
		"agent_id":    agent.ID.String(),
		"old_status": oldStatus,
		"new_status": status,
	})
	return agent, nil
}

func (s *AgentService) logSecurityEvent(userID uuid.UUID, actionType string, details map[string]interface{}) {
	if s.securityRepo != nil {
		if err := s.securityRepo.Create(&model.SecurityEvent{
			UserID:     userID,
			ActionType: actionType,
			DetailsJSON: toJSON(details),
		}); err != nil {
			log.Printf("security event log failed: %v", err)
		}
	}
}

func toJSON(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
