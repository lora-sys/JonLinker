package repository

import (
	"context"
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository() *JobRepository {
	return &JobRepository{db: nil}
}

func (r *JobRepository) WithDB(db *gorm.DB) *JobRepository {
	r.db = db
	return r
}

func (r *JobRepository) Create(job *model.Job) error {
	return r.db.Create(job).Error
}

func (r *JobRepository) GetByID(id uuid.UUID) (*model.Job, error) {
	var job model.Job
	if err := r.db.Preload("Agent").First(&job, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *JobRepository) Update(job *model.Job) error {
	return r.db.Save(job).Error
}

func (r *JobRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Job{}, "id = ?", id).Error
}

func (r *JobRepository) ListByAgentID(agentID uuid.UUID) ([]*model.Job, error) {
	var jobs []*model.Job
	if err := r.db.Where("agent_id = ?", agentID).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *JobRepository) ListByStatus(status model.JobStatus, limit, offset int) ([]*model.Job, int64, error) {
	var jobs []*model.Job
	var total int64
	r.db.Model(&model.Job{}).Where("status = ?", status).Count(&total)
	if err := r.db.Where("status = ?", status).Limit(limit).Offset(offset).Preload("Agent").Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (r *JobRepository) ListAll(limit, offset int) ([]*model.Job, int64, error) {
	var jobs []*model.Job
	var total int64
	r.db.Model(&model.Job{}).Count(&total)
	if err := r.db.Limit(limit).Offset(offset).Preload("Agent").Find(&jobs).Error; err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

func (r *JobRepository) Search(ctx context.Context, query string, skills []string, location string, salaryMin int, jobType string, limit int) ([]*model.Job, error) {
	var jobs []*model.Job
	db := r.db.WithContext(ctx).Model(&model.Job{})

	if query != "" {
		db = db.Where("title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if location != "" {
		db = db.Where("location ILIKE ?", "%"+location+"%")
	}
	if salaryMin > 0 {
		db = db.Where("salary_max >= ?", salaryMin)
	}
	if jobType != "" {
		db = db.Where("job_type = ?", jobType)
	}

	if err := db.Limit(limit).Preload("Agent").Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
