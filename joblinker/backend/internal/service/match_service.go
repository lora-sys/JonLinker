package service

import (
	"errors"

	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

var ErrInvalidMatchState = errors.New("invalid match state transition")

type MatchService struct {
	matchRepo *repository.MatchRepository
	agentRepo *repository.AgentRepository
	jobRepo   *repository.JobRepository
}

func NewMatchService(matchRepo *repository.MatchRepository, agentRepo *repository.AgentRepository, jobRepo *repository.JobRepository) *MatchService {
	return &MatchService{
		matchRepo: matchRepo,
		agentRepo: agentRepo,
		jobRepo:   jobRepo,
	}
}

func (s *MatchService) CreateMatch(seekerAgentID, jobID uuid.UUID, score float64) (*model.Match, error) {
	match := &model.Match{
		SeekerAgentID: seekerAgentID,
		JobID:         jobID,
		Score:         score,
		Status:        model.MatchStatusPending,
	}
	if err := s.matchRepo.Create(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) GetMatch(id uuid.UUID) (*model.Match, error) {
	return s.matchRepo.GetByID(id)
}

func (s *MatchService) ConfirmMatch(id uuid.UUID) (*model.Match, error) {
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if match.Status != model.MatchStatusPending {
		return nil, ErrInvalidMatchState
	}
	match.Status = model.MatchStatusMutualInterest
	if err := s.matchRepo.Update(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) TransitionToNegotiating(id uuid.UUID) (*model.Match, error) {
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if match.Status != model.MatchStatusMutualInterest {
		return nil, ErrInvalidMatchState
	}
	match.Status = model.MatchStatusNegotiating
	if err := s.matchRepo.Update(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) ListUserMatches(userID uuid.UUID) ([]*model.Match, error) {
	agents, err := s.agentRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	var allMatches []*model.Match
	for _, agent := range agents {
		if agent.Type == model.AgentTypeSeeker {
			matches, err := s.matchRepo.ListBySeekerAgentID(agent.ID)
			if err != nil {
				continue
			}
			allMatches = append(allMatches, matches...)
		}
	}
	return allMatches, nil
}

func (s *MatchService) FindMatches(agentID uuid.UUID, limit int) ([]*model.Match, error) {
	return s.matchRepo.ListBySeekerAgentID(agentID)
}

func (s *MatchService) CalculateScore(seekerAgentID, jobID uuid.UUID) (float64, error) {
	return 0.85, nil
}
