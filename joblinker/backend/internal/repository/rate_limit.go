package repository

import (
	"joblinker/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RateLimitRepository struct {
	db *gorm.DB
}

func NewRateLimitRepository() *RateLimitRepository {
	return &RateLimitRepository{}
}

func (r *RateLimitRepository) WithDB(db *gorm.DB) *RateLimitRepository {
	return &RateLimitRepository{db: db}
}

func (r *RateLimitRepository) GetActiveWindow(userID uuid.UUID, windowDuration time.Duration) (*model.RateLimitCounter, error) {
	windowStart := time.Now().Add(-windowDuration)
	var counter model.RateLimitCounter
	err := r.db.Where("user_id = ? AND window_start > ?", userID, windowStart).
		Order("window_start DESC").
		First(&counter).Error
	if err != nil {
		return nil, err
	}
	return &counter, nil
}

func (r *RateLimitRepository) Increment(userID uuid.UUID, windowDuration time.Duration) (*model.RateLimitCounter, error) {
	counter, err := r.GetActiveWindow(userID, windowDuration)
	if err == gorm.ErrRecordNotFound {
		counter = &model.RateLimitCounter{
			UserID:       userID,
			WindowStart:  time.Now(),
			MessageCount: 0,
		}
		if err := r.db.Create(counter).Error; err != nil {
			return nil, err
		}
		return counter, nil
	}
	if err != nil {
		return nil, err
	}

	counter.MessageCount++
	if err := r.db.Save(counter).Error; err != nil {
		return nil, err
	}
	return counter, nil
}

func (r *RateLimitRepository) GetMessageCount(userID uuid.UUID, windowDuration time.Duration) (int, error) {
	counter, err := r.GetActiveWindow(userID, windowDuration)
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return counter.MessageCount, nil
}
