package service

import (
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"time"

	"github.com/google/uuid"
)

type InterviewService struct {
	interviewRepo *repository.InterviewRepository
	matchRepo     *repository.MatchRepository
	messageSvc    *MessageService
}

func NewInterviewService(interviewRepo *repository.InterviewRepository, matchRepo *repository.MatchRepository, messageSvc *MessageService) *InterviewService {
	return &InterviewService{
		interviewRepo: interviewRepo,
		matchRepo:     matchRepo,
		messageSvc:    messageSvc,
	}
}

func (s *InterviewService) ScheduleInterview(matchID uuid.UUID, scheduledAt string, format model.InterviewFormat, location string) (*model.Interview, error) {
	interview := &model.Interview{
		MatchID:     matchID,
		ScheduledAt: parseTime(scheduledAt),
		Format:      format,
		Location:    location,
		Status:      model.InterviewStatusScheduled,
	}
	if err := s.interviewRepo.Create(interview); err != nil {
		return nil, err
	}
	return interview, nil
}

func (s *InterviewService) GetInterview(id uuid.UUID) (*model.Interview, error) {
	return s.interviewRepo.GetByID(id)
}

func (s *InterviewService) UpdateInterview(id uuid.UUID, updates map[string]interface{}) (*model.Interview, error) {
	interview, err := s.interviewRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if status, ok := updates["status"].(string); ok {
		interview.Status = model.InterviewStatus(status)
	}
	if location, ok := updates["location"].(string); ok {
		interview.Location = location
	}
	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, err
	}
	return interview, nil
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
