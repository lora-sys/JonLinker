package service

import (
	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

type MessageService struct {
	messageRepo *repository.MessageRepository
	matchRepo   *repository.MatchRepository
	agentRepo   *repository.AgentRepository
}

func NewMessageService(messageRepo *repository.MessageRepository, matchRepo *repository.MatchRepository, agentRepo *repository.AgentRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
		agentRepo:   agentRepo,
	}
}

func (s *MessageService) CreateMessage(matchID, senderAgentID uuid.UUID, contentXML, intentType string) (*model.Message, error) {
	message := &model.Message{
		MatchID:      matchID,
		SenderAgentID: senderAgentID,
		ContentXML:  contentXML,
		IntentType:   intentType,
	}
	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}
	return message, nil
}