package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OfferRepository struct {
	db *gorm.DB
}

func NewOfferRepository() *OfferRepository {
	return &OfferRepository{db: nil}
}

func (r *OfferRepository) WithDB(db *gorm.DB) *OfferRepository {
	r.db = db
	return r
}

func (r *OfferRepository) Create(offer *model.Offer) error {
	return r.db.Create(offer).Error
}

func (r *OfferRepository) GetByID(id uuid.UUID) (*model.Offer, error) {
	var offer model.Offer
	if err := r.db.Preload("Match").First(&offer, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *OfferRepository) GetByMatchID(matchID uuid.UUID) (*model.Offer, error) {
	var offer model.Offer
	if err := r.db.First(&offer, "match_id = ?", matchID).Error; err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *OfferRepository) ListAll(limit, offset int) ([]*model.Offer, int64, error) {
	var offers []*model.Offer
	var total int64
	r.db.Model(&model.Offer{}).Count(&total)
	if err := r.db.Limit(limit).Offset(offset).Preload("Match").Find(&offers).Error; err != nil {
		return nil, 0, err
	}
	return offers, total, nil
}

func (r *OfferRepository) Update(offer *model.Offer) error {
	return r.db.Save(offer).Error
}
