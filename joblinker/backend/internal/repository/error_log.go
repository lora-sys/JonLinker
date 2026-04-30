package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ErrorLogRepository struct {
	db *gorm.DB
}

func NewErrorLogRepository() *ErrorLogRepository {
	return &ErrorLogRepository{}
}

func (r *ErrorLogRepository) WithDB(db *gorm.DB) *ErrorLogRepository {
	return &ErrorLogRepository{db: db}
}

func (r *ErrorLogRepository) Create(log *model.ErrorLog) error {
	return r.db.Create(log).Error
}

func (r *ErrorLogRepository) GetByCorrelationID(correlationID string) ([]model.ErrorLog, error) {
	var logs []model.ErrorLog
	err := r.db.Where("correlation_id = ?", correlationID).Find(&logs).Error
	return logs, err
}

func (r *ErrorLogRepository) GetByUserID(userID uuid.UUID) ([]model.ErrorLog, error) {
	var logs []model.ErrorLog
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *ErrorLogRepository) GetByMatchID(matchID uuid.UUID) ([]model.ErrorLog, error) {
	var logs []model.ErrorLog
	err := r.db.Where("match_id = ?", matchID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}
