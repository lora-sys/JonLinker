package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SecurityEventRepository struct {
	db *gorm.DB
}

func NewSecurityEventRepository() *SecurityEventRepository {
	return &SecurityEventRepository{db: nil}
}

func (r *SecurityEventRepository) WithDB(db *gorm.DB) *SecurityEventRepository {
	r.db = db
	return r
}

func (r *SecurityEventRepository) Create(event *model.SecurityEvent) error {
	return r.db.Create(event).Error
}

func (r *SecurityEventRepository) ListByUserID(userID uuid.UUID, limit, offset int) ([]*model.SecurityEvent, int64, error) {
	var events []*model.SecurityEvent
	var total int64
	r.db.Model(&model.SecurityEvent{}).Where("user_id = ?", userID).Count(&total)
	if err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}
