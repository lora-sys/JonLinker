package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationSummaryRepository struct {
	db *gorm.DB
}

func NewConversationSummaryRepository() *ConversationSummaryRepository {
	return &ConversationSummaryRepository{db: nil}
}

func (r *ConversationSummaryRepository) WithDB(db *gorm.DB) *ConversationSummaryRepository {
	r.db = db
	return r
}

func (r *ConversationSummaryRepository) Create(summary *model.ConversationSummary) error {
	return r.db.Create(summary).Error
}

func (r *ConversationSummaryRepository) GetByMatchID(matchID uuid.UUID) (*model.ConversationSummary, error) {
	var summary model.ConversationSummary
	if err := r.db.Where("match_id = ?", matchID).Order("created_at DESC").First(&summary).Error; err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *ConversationSummaryRepository) Update(summary *model.ConversationSummary) error {
	return r.db.Save(summary).Error
}

func (r *ConversationSummaryRepository) DeleteByMatchID(matchID uuid.UUID) error {
	return r.db.Where("match_id = ?", matchID).Delete(&model.ConversationSummary{}).Error
}