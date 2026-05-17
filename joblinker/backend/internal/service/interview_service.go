package service

import (
	"encoding/json"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"time"

	"github.com/google/uuid"
)

type InterviewService struct {
	interviewRepo *repository.InterviewRepository
	matchRepo     *repository.MatchRepository
	messageSvc    *MessageService
	securitySvc   *SecurityService
}

func NewInterviewService(
	interviewRepo *repository.InterviewRepository,
	matchRepo *repository.MatchRepository,
	messageSvc *MessageService,
	securitySvc *SecurityService,
) *InterviewService {
	return &InterviewService{
		interviewRepo: interviewRepo,
		matchRepo:     matchRepo,
		messageSvc:    messageSvc,
		securitySvc:   securitySvc,
	}
}

func (s *InterviewService) ScheduleInterview(matchID uuid.UUID, scheduledAt string, format model.InterviewFormat, location string) (*model.Interview, error) {
	interview := &model.Interview{
		MatchID:     matchID,
		ScheduledAt: parseTimeStr(scheduledAt),
		Format:      format,
		Location:    location,
		Status:      model.InterviewStatusScheduled,
		Feedback:    json.RawMessage("null"),
	}
	if err := s.interviewRepo.Create(interview); err != nil {
		return nil, err
	}

	if s.securitySvc != nil {
		s.securitySvc.LogEvent(matchID, "interview_scheduled", map[string]interface{}{
			"interview_id": interview.ID.String(),
			"format":       format,
			"location":     location,
		}, "")
	}

	return interview, nil
}

func (s *InterviewService) ListAll(limit, offset int) ([]*model.Interview, int64, error) {
	return s.interviewRepo.ListAll(limit, offset)
}

func (s *InterviewService) GetInterview(id uuid.UUID) (*model.Interview, error) {
	return s.interviewRepo.GetByID(id)
}

func (s *InterviewService) GetInterviewByMatchID(matchID uuid.UUID) (*model.Interview, error) {
	return s.interviewRepo.GetByMatchID(matchID)
}

func (s *InterviewService) ConfirmInterview(matchID uuid.UUID) (*model.Interview, error) {
	interview, err := s.interviewRepo.GetByMatchID(matchID)
	if err != nil {
		return nil, err
	}
	interview.Status = model.InterviewStatusConfirmed
	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, err
	}
	return interview, nil
}

func (s *InterviewService) CancelInterview(matchID uuid.UUID) (*model.Interview, error) {
	interview, err := s.interviewRepo.GetByMatchID(matchID)
	if err != nil {
		return nil, err
	}
	interview.Status = model.InterviewStatusCancelled
	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, err
	}
	return interview, nil
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

func parseTimeStr(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
