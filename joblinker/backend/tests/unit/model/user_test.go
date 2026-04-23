package model_test

import (
	"testing"
	"time"

	"joblinker/internal/model"

	"github.com/google/uuid"
)

func TestUserModel(t *testing.T) {
	user := &model.User{
		ID:             uuid.New(),
		Email:          "test@example.com",
		PasswordHash:   "hashedpassword",
		Role:           model.RoleSeeker,
		OrganizationID: nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}

	if user.Role != model.RoleSeeker {
		t.Errorf("expected role %s, got %s", model.RoleSeeker, user.Role)
	}
}

func TestAgentModel(t *testing.T) {
	agent := &model.Agent{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Type:      model.AgentTypeSeeker,
		Status:    model.AgentStatusActive,
		ConfigJSON: "{}",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if agent.Type != model.AgentTypeSeeker {
		t.Errorf("expected type %s, got %s", model.AgentTypeSeeker, agent.Type)
	}

	if agent.Status != model.AgentStatusActive {
		t.Errorf("expected status %s, got %s", model.AgentStatusActive, agent.Status)
	}
}

func TestJobModel(t *testing.T) {
	job := &model.Job{
		ID:           uuid.New(),
		AgentID:      uuid.New(),
		StructuredJSON: "{}",
		VectorID:     "vec-123",
		Status:       model.JobStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if job.Status != model.JobStatusActive {
		t.Errorf("expected status %s, got %s", model.JobStatusActive, job.Status)
	}
}

func TestMatchModel(t *testing.T) {
	match := &model.Match{
		ID:            uuid.New(),
		SeekerAgentID: uuid.New(),
		JobID:         uuid.New(),
		Score:         0.95,
		Status:        model.MatchStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if match.Score != 0.95 {
		t.Errorf("expected score 0.95, got %f", match.Score)
	}

	if match.Status != model.MatchStatusPending {
		t.Errorf("expected status %s, got %s", model.MatchStatusPending, match.Status)
	}
}

func TestInterviewModel(t *testing.T) {
	interview := &model.Interview{
		ID:           uuid.New(),
		MatchID:     uuid.New(),
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Format:      model.InterviewFormatVideo,
		Location:    "https://meet.example.com/123",
		Status:      model.InterviewStatusScheduled,
		ReminderSent: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if interview.Format != model.InterviewFormatVideo {
		t.Errorf("expected format %s, got %s", model.InterviewFormatVideo, interview.Format)
	}

	if interview.Status != model.InterviewStatusScheduled {
		t.Errorf("expected status %s, got %s", model.InterviewStatusScheduled, interview.Status)
	}

	if interview.ReminderSent != false {
		t.Error("expected reminder_sent to be false")
	}
}

func TestOfferModel(t *testing.T) {
	now := time.Now()
	offer := &model.Offer{
		ID:               uuid.New(),
		MatchID:          uuid.New(),
		CompensationJSON: `{"base_salary":100000}`,
		StartDate:        now.Add(30 * 24 * time.Hour),
		Status:           model.OfferStatusPending,
		RespondedAt:     nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if offer.Status != model.OfferStatusPending {
		t.Errorf("expected status %s, got %s", model.OfferStatusPending, offer.Status)
	}

	if offer.RespondedAt != nil {
		t.Error("expected responded_at to be nil")
	}
}

func TestSecurityEventModel(t *testing.T) {
	event := &model.SecurityEvent{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		ActionType: "login_success",
		DetailsJSON: `{"ip":"127.0.0.1"}`,
		IPAddress:  "127.0.0.1",
	}

	if event.ActionType != "login_success" {
		t.Errorf("expected action_type login_success, got %s", event.ActionType)
	}
}
