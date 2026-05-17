package service

import (
	"encoding/json"
	"fmt"
	"log"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"time"

	"github.com/google/uuid"
)

type OfferService struct {
	offerRepo   *repository.OfferRepository
	matchRepo   *repository.MatchRepository
	jobRepo     *repository.JobRepository
	securitySvc *SecurityService
}

func NewOfferService(
	offerRepo *repository.OfferRepository,
	matchRepo *repository.MatchRepository,
	jobRepo *repository.JobRepository,
	securitySvc *SecurityService,
) *OfferService {
	return &OfferService{
		offerRepo:   offerRepo,
		matchRepo:   matchRepo,
		jobRepo:     jobRepo,
		securitySvc: securitySvc,
	}
}

func (s *OfferService) GenerateOffer(matchID uuid.UUID, compensationJSON string, startDate string) (*model.Offer, error) {
	offer := &model.Offer{
		MatchID:           matchID,
		CompensationJSON: json.RawMessage(compensationJSON),
		StartDate:        parseDate(startDate),
		Status:           model.OfferStatusPending,
	}
	if err := s.offerRepo.Create(offer); err != nil {
		return nil, err
	}
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match for status update: %w", err)
	}
	match.Status = model.MatchStatusOffered
	if err := s.matchRepo.Update(match); err != nil {
		return nil, fmt.Errorf("failed to update match status: %w", err)
	}

	if s.securitySvc != nil {
		s.securitySvc.LogEvent(matchID, "offer_sent", map[string]interface{}{
			"offer_id":     offer.ID.String(),
			"compensation": compensationJSON,
			"start_date":   startDate,
		}, "")
	}

	return offer, nil
}

func (s *OfferService) ListAll(limit, offset int) ([]*model.Offer, int64, error) {
	return s.offerRepo.ListAll(limit, offset)
}

func (s *OfferService) GetOffer(id uuid.UUID) (*model.Offer, error) {
	return s.offerRepo.GetByID(id)
}

func (s *OfferService) GetOfferByMatchID(matchID uuid.UUID) (*model.Offer, error) {
	return s.offerRepo.GetByMatchID(matchID)
}

func (s *OfferService) AcceptOffer(matchID uuid.UUID) (*model.Offer, error) {
	offer, err := s.offerRepo.GetByMatchID(matchID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	offer.RespondedAt = &now
	offer.Status = model.OfferStatusAccepted
	if err := s.offerRepo.Update(offer); err != nil {
		return nil, err
	}
	s.handleOfferAccepted(offer)
	return offer, nil
}

func (s *OfferService) DeclineOffer(matchID uuid.UUID) (*model.Offer, error) {
	offer, err := s.offerRepo.GetByMatchID(matchID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	offer.RespondedAt = &now
	offer.Status = model.OfferStatusDeclined
	if err := s.offerRepo.Update(offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *OfferService) RespondToOffer(id uuid.UUID, response string, counterAmount ...float64) (*model.Offer, error) {
	offer, err := s.offerRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	offer.RespondedAt = &now
	details := map[string]interface{}{
		"offer_id": id.String(),
		"response": response,
	}
	switch response {
	case "accept":
		offer.Status = model.OfferStatusAccepted
		s.handleOfferAccepted(offer)
	case "decline":
		offer.Status = model.OfferStatusDeclined
	case "negotiate":
		offer.Status = model.OfferStatusNegotiating
		if len(counterAmount) > 0 && counterAmount[0] > 0 {
			details["counter_amount"] = counterAmount[0]
			log.Printf("Counter-offer submitted for offer %s: %.2f", id, counterAmount[0])
		}
	}
	if err := s.offerRepo.Update(offer); err != nil {
		return nil, err
	}

	if s.securitySvc != nil {
		actionType := "offer_" + response
		s.securitySvc.LogEvent(offer.MatchID, actionType, details, "")
	}

	return offer, nil
}

func (s *OfferService) handleOfferAccepted(offer *model.Offer) {
	if s.jobRepo == nil {
		return
	}
	match, err := s.matchRepo.GetByID(offer.MatchID)
	if err != nil {
		return
	}
	job, err := s.jobRepo.GetByID(match.JobID)
	if err != nil {
		return
	}
	job.Status = model.JobStatusFilled
	s.jobRepo.Update(job)
}

func (s *OfferService) UpdateStatus(id uuid.UUID, status model.OfferStatus) (*model.Offer, error) {
	offer, err := s.offerRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	offer.Status = status
	if err := s.offerRepo.Update(offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Now()
	}
	return t
}
