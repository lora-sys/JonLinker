package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NegotiationSessionRepository struct {
	db *gorm.DB
}

func NewNegotiationSessionRepository() *NegotiationSessionRepository {
	return &NegotiationSessionRepository{db: nil}
}

func (r *NegotiationSessionRepository) WithDB(db *gorm.DB) *NegotiationSessionRepository {
	r.db = db
	return r
}

func (r *NegotiationSessionRepository) Create(session *model.NegotiationSession) error {
	return r.db.Create(session).Error
}

func (r *NegotiationSessionRepository) GetByID(id uuid.UUID) (*model.NegotiationSession, error) {
	var session model.NegotiationSession
	if err := r.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *NegotiationSessionRepository) GetByMatchID(matchID uuid.UUID) (*model.NegotiationSession, error) {
	var session model.NegotiationSession
	if err := r.db.Where("match_id = ?", matchID).Order("created_at DESC").First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *NegotiationSessionRepository) Update(session *model.NegotiationSession) error {
	return r.db.Save(session).Error
}

func (r *NegotiationSessionRepository) DeleteByMatchID(matchID uuid.UUID) error {
	return r.db.Where("match_id = ?", matchID).Delete(&model.NegotiationSession{}).Error
}

func (r *NegotiationSessionRepository) ListByStatus(status model.NegotiationStatus) ([]*model.NegotiationSession, error) {
	var sessions []*model.NegotiationSession
	if err := r.db.Where("status = ?", status).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}