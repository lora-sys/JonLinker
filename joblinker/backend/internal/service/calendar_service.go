package service

import (
	"errors"
	"time"

	"joblinker/internal/agent"
	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

type CalendarService struct {
	interviewRepo *repository.InterviewRepository
	matchRepo     *repository.MatchRepository
	agentRepo     *repository.AgentRepository
	scheduler     *agent.Scheduler
}

func NewCalendarService(interviewRepo *repository.InterviewRepository, matchRepo *repository.MatchRepository, agentRepo *repository.AgentRepository) *CalendarService {
	return &CalendarService{
		interviewRepo: interviewRepo,
		matchRepo:     matchRepo,
		agentRepo:     agentRepo,
		scheduler:     agent.NewScheduler(),
	}
}

func (s *CalendarService) ScheduleInterviews(matchID uuid.UUID, proposedTime string, format model.InterviewFormat, location string) (*model.Interview, error) {
	// Parse proposed time
	parsedTime, err := parseTime(proposedTime)
	if err != nil {
		return nil, err
	}

	// Validate time slot
	if !s.scheduler.IsWithinNoticePeriod(parsedTime) {
		return nil, ErrInvalidTimeSlot
	}

	if s.scheduler.IsTooFarAhead(parsedTime) {
		return nil, ErrInvalidTimeSlot
	}

	// Create interview
	interview := &model.Interview{
		ID:        uuid.New(),
		MatchID:   matchID,
		ScheduledAt: parsedTime,
		Format:   format,
		Location: location,
		Status:   model.InterviewStatusScheduled,
	}

	if err := s.interviewRepo.Create(interview); err != nil {
		return nil, err
	}

	return interview, nil
}

func (s *CalendarService) RescheduleInterview(interviewID uuid.UUID, newTime string) (*model.Interview, error) {
	interview, err := s.interviewRepo.GetByID(interviewID)
	if err != nil {
		return nil, err
	}

	parsedTime, err := parseTime(newTime)
	if err != nil {
		return nil, err
	}

	interview.ScheduledAt = parsedTime
	interview.Status = model.InterviewStatusRescheduled

	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, err
	}

	return interview, nil
}

func (s *CalendarService) ConfirmInterview(interviewID uuid.UUID) (*model.Interview, error) {
	interview, err := s.interviewRepo.GetByID(interviewID)
	if err != nil {
		return nil, err
	}

	interview.Status = model.InterviewStatusScheduled

	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, err
	}

	return interview, nil
}

func (s *CalendarService) CancelInterview(interviewID uuid.UUID) (*model.Interview, error) {
	interview, err := s.interviewRepo.GetByID(interviewID)
	if err != nil {
		return nil, err
	}

	interview.Status = model.InterviewStatusCancelled

	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, err
	}

	return interview, nil
}

func (s *CalendarService) GenerateICS(interview *model.Interview) (string, error) {
	title := "Job Interview"
	description := "Interview scheduled via JobLinker"
	location := interview.Location

	return agent.GenerateICS(title, description, interview.ScheduledAt, interview.ScheduledAt.Add(time.Hour), location), nil
}

// Error definitions
var (
	ErrInvalidTimeSlot = errors.New("invalid time slot for interview")
)

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}