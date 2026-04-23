package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InterviewRepository struct {
	db *gorm.DB
}

func NewInterviewRepository() *InterviewRepository {
	return &InterviewRepository{db: nil}
}

func (r *InterviewRepository) WithDB(db *gorm.DB) *InterviewRepository {
	r.db = db
	return r
}

func (r *InterviewRepository) Create(interview *model.Interview) error {
	return r.db.Create(interview).Error
}

func (r *InterviewRepository) GetByID(id uuid.UUID) (*model.Interview, error) {
	var interview model.Interview
	if err := r.db.Preload("Match").First(&interview, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &interview, nil
}

func (r *InterviewRepository) GetByMatchID(matchID uuid.UUID) (*model.Interview, error) {
	var interview model.Interview
	if err := r.db.First(&interview, "match_id = ?", matchID).Error; err != nil {
		return nil, err
	}
	return &interview, nil
}

func (r *InterviewRepository) Update(interview *model.Interview) error {
	return r.db.Save(interview).Error
}

func (r *InterviewRepository) ListAll(limit, offset int) ([]*model.Interview, int64, error) {
	var interviews []*model.Interview
	var total int64
	r.db.Model(&model.Interview{}).Count(&total)
	if err := r.db.Limit(limit).Offset(offset).Preload("Match").Find(&interviews).Error; err != nil {
		return nil, 0, err
	}
	return interviews, total, nil
}

func (r *InterviewRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Interview{}, "id = ?", id).Error
}
