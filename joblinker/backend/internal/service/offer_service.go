package service

import (
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
		CompensationJSON: compensationJSON,
		StartDate:        parseDate(startDate),
		Status:           model.OfferStatusPending,
	}
	if err := s.offerRepo.Create(offer); err != nil {
		return nil, err
	}
	match, _ := s.matchRepo.GetByID(matchID)
	if match != nil {
		match.Status = model.MatchStatusOffered
		s.matchRepo.Update(match)
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

func (s *OfferService) GetOffer(id uuid.UUID) (*model.Offer, error) {
	return s.offerRepo.GetByID(id)
}

func (s *OfferService) RespondToOffer(id uuid.UUID, response string) (*model.Offer, error) {
	offer, err := s.offerRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	offer.RespondedAt = &now
	switch response {
	case "accept":
		offer.Status = model.OfferStatusAccepted
		s.handleOfferAccepted(offer)
	case "decline":
		offer.Status = model.OfferStatusDeclined
	case "negotiate":
		offer.Status = model.OfferStatusNegotiating
	}
	if err := s.offerRepo.Update(offer); err != nil {
		return nil, err
	}

	if s.securitySvc != nil {
		actionType := "offer_" + response
		s.securitySvc.LogEvent(offer.MatchID, actionType, map[string]interface{}{
			"offer_id": id.String(),
			"response": response,
		}, "")
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
