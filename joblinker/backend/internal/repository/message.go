package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{db: nil}
}

func (r *MessageRepository) WithDB(db *gorm.DB) *MessageRepository {
	r.db = db
	return r
}

func (r *MessageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) GetByID(id uuid.UUID) (*model.Message, error) {
	var message model.Message
	if err := r.db.First(&message, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *MessageRepository) DeleteByMatchID(matchID uuid.UUID) error {
	return r.db.Where("match_id = ?", matchID).Delete(&model.Message{}).Error
}

func (r *MessageRepository) ListByMatchID(matchID uuid.UUID) ([]*model.Message, error) {
	var messages []*model.Message
	if err := r.db.Where("match_id = ?", matchID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
