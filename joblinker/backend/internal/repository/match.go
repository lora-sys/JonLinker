package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MatchRepository struct {
	db *gorm.DB
}

func NewMatchRepository() *MatchRepository {
	return &MatchRepository{db: nil}
}

func (r *MatchRepository) WithDB(db *gorm.DB) *MatchRepository {
	r.db = db
	return r
}

func (r *MatchRepository) Create(match *model.Match) error {
	return r.db.Create(match).Error
}

func (r *MatchRepository) GetByID(id uuid.UUID) (*model.Match, error) {
	var match model.Match
	if err := r.db.Preload("SeekerAgent").Preload("Job").First(&match, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MatchRepository) Update(match *model.Match) error {
	return r.db.Save(match).Error
}

func (r *MatchRepository) ListBySeekerAgentID(seekerAgentID uuid.UUID) ([]*model.Match, error) {
	var matches []*model.Match
	if err := r.db.Where("seeker_agent_id = ?", seekerAgentID).Preload("Job").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *MatchRepository) ListByJobID(jobID uuid.UUID) ([]*model.Match, error) {
	var matches []*model.Match
	if err := r.db.Where("job_id = ?", jobID).Preload("SeekerAgent").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *MatchRepository) ListAll(limit, offset int) ([]*model.Match, int64, error) {
	var matches []*model.Match
	var total int64
	r.db.Model(&model.Match{}).Count(&total)
	if err := r.db.Limit(limit).Offset(offset).Preload("SeekerAgent").Preload("Job").Find(&matches).Error; err != nil {
		return nil, 0, err
	}
	return matches, total, nil
}

func (r *MatchRepository) ListByStatus(status model.MatchStatus, limit, offset int) ([]*model.Match, int64, error) {
	var matches []*model.Match
	var total int64
	r.db.Model(&model.Match{}).Where("status = ?", status).Count(&total)
	if err := r.db.Where("status = ?", status).Limit(limit).Offset(offset).Preload("SeekerAgent").Preload("Job").Find(&matches).Error; err != nil {
		return nil, 0, err
	}
	return matches, total, nil
}
