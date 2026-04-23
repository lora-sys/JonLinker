package service

import (
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"log"
	"time"
)

type ReminderJob struct {
	interviewRepo *repository.InterviewRepository
	emailSvc      *EmailService
	matchRepo     *repository.MatchRepository
	agentRepo     *repository.AgentRepository
	userRepo      *repository.UserRepository
}

func NewReminderJob(
	interviewRepo *repository.InterviewRepository,
	emailSvc *EmailService,
	matchRepo *repository.MatchRepository,
	agentRepo *repository.AgentRepository,
	userRepo *repository.UserRepository,
) *ReminderJob {
	return &ReminderJob{
		interviewRepo: interviewRepo,
		emailSvc:      emailSvc,
		matchRepo:     matchRepo,
		agentRepo:     agentRepo,
		userRepo:      userRepo,
	}
}

func (r *ReminderJob) Run() error {
	interviews, err := r.interviewRepo.GetUpcomingReminders()
	if err != nil {
		return err
	}

	for _, interview := range interviews {
		if err := r.sendReminder(&interview); err != nil {
			log.Printf("Failed to send reminder for interview %s: %v", interview.ID, err)
			continue
		}

		if err := r.interviewRepo.MarkReminderSent(interview.ID); err != nil {
			log.Printf("Failed to mark reminder sent for interview %s: %v", interview.ID, err)
		}
	}

	return nil
}

func (r *ReminderJob) sendReminder(interview *model.Interview) error {
	if r.emailSvc == nil {
		return nil
	}

	match, err := r.matchRepo.GetByID(interview.MatchID)
	if err != nil {
		return err
	}

	agent, err := r.agentRepo.GetByID(match.SeekerAgentID)
	if err != nil {
		return err
	}

	user, err := r.userRepo.GetByID(agent.UserID)
	if err != nil {
		return err
	}

	return r.emailSvc.SendReminder(interview, user.Email)
}

func (r *ReminderJob) Start() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			if err := r.Run(); err != nil {
				log.Printf("Reminder job error: %v", err)
			}
		}
	}()
}
