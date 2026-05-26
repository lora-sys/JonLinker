package nodes

import (
	"context"
	"fmt"
	"log"
	"time"

	"joblinker/internal/core"

	"github.com/google/uuid"
)

// ConfirmNode handles confirmation workflows and interview scheduling.
type ConfirmNode struct {
	interviewRepo core.InterviewRepository
	matchRepo     core.MatchRepository
}

// NewConfirmNode creates a new ConfirmNode.
func NewConfirmNode(interviewRepo core.InterviewRepository, matchRepo core.MatchRepository) *ConfirmNode {
	return &ConfirmNode{
		interviewRepo: interviewRepo,
		matchRepo:     matchRepo,
	}
}

// Process handles confirmation and scheduling actions.
// Input:
//   - action: "confirm_match", "schedule_interview", "reject", "accept_offer"
//   - match_id: UUID string
//   - interview details for scheduling
//
// Output:
//   - status: result status
//   - details: action-specific result data
func (n *ConfirmNode) Process(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	action, _ := input["action"].(string)
	matchIDStr, _ := input["match_id"].(string)

	if matchIDStr == "" {
		return nil, fmt.Errorf("match_id required")
	}

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid match_id: %w", err)
	}

	switch action {
	case "confirm_match":
		return n.confirmMatch(ctx, matchID)
	case "reject":
		return n.rejectMatch(ctx, matchID)
	case "schedule_interview":
		return n.scheduleInterview(ctx, matchID, input)
	case "accept_offer":
		return n.acceptOffer(ctx, matchID)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (n *ConfirmNode) confirmMatch(ctx context.Context, matchID uuid.UUID) (map[string]interface{}, error) {
	if n.matchRepo != nil {
		if err := n.matchRepo.UpdateStatus(ctx, matchID, core.MatchStatusConfirmed); err != nil {
			return nil, fmt.Errorf("failed to confirm match: %w", err)
		}
		log.Printf("Match %s confirmed", matchID)
	}

	return map[string]interface{}{
		"status":  "confirmed",
		"message": "Match has been confirmed successfully",
	}, nil
}

func (n *ConfirmNode) rejectMatch(ctx context.Context, matchID uuid.UUID) (map[string]interface{}, error) {
	if n.matchRepo != nil {
		if err := n.matchRepo.UpdateStatus(ctx, matchID, core.MatchStatusRejected); err != nil {
			return nil, fmt.Errorf("failed to reject match: %w", err)
		}
		log.Printf("Match %s rejected", matchID)
	}

	return map[string]interface{}{
		"status":  "rejected",
		"message": "Match has been rejected",
	}, nil
}

func (n *ConfirmNode) scheduleInterview(ctx context.Context, matchID uuid.UUID, input map[string]interface{}) (map[string]interface{}, error) {
	datetimeStr, _ := input["datetime"].(string)
	if datetimeStr == "" {
		return nil, fmt.Errorf("datetime required for interview scheduling")
	}

	datetime, err := time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		datetime, err = time.Parse("2006-01-02T15:04", datetimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid datetime format: %w", err)
		}
	}

	interviewType, _ := input["interview_type"].(string)
	var format core.InterviewFormat
	switch interviewType {
	case "video":
		format = core.InterviewFormatVideo
	case "phone":
		format = core.InterviewFormatPhone
	case "onsite":
		format = core.InterviewFormatOnsite
	default:
		format = core.InterviewFormatVideo
	}

	interview := &core.Interview{
		ID:          uuid.New(),
		MatchID:     matchID,
		ScheduledAt: datetime,
		Format:      format,
		Status:      core.InterviewStatusScheduled,
	}

	if n.matchRepo != nil {
		match, err := n.matchRepo.GetByID(matchID)
		if err == nil {
			interview.TenantID = match.TenantID
		}
	}

	if n.interviewRepo != nil {
		if err := n.interviewRepo.Create(interview); err != nil {
			return nil, fmt.Errorf("failed to create interview: %w", err)
		}
	}

	log.Printf("Interview %s scheduled for match %s at %s", interview.ID, matchID, datetimeStr)

	return map[string]interface{}{
		"status":       "scheduled",
		"interview_id": interview.ID.String(),
		"datetime":     datetimeStr,
		"format":       interviewType,
	}, nil
}

func (n *ConfirmNode) acceptOffer(ctx context.Context, matchID uuid.UUID) (map[string]interface{}, error) {
	if n.matchRepo != nil {
		if err := n.matchRepo.UpdateStatus(ctx, matchID, core.MatchStatusAccepted); err != nil {
			return nil, fmt.Errorf("failed to accept offer: %w", err)
		}
		log.Printf("Offer accepted for match %s", matchID)
	}

	return map[string]interface{}{
		"status":  "accepted",
		"message": "Offer has been accepted",
	}, nil
}
