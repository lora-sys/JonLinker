package repository

import (
	"context"
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository() *AgentRepository {
	return &AgentRepository{db: nil}
}

func (r *AgentRepository) WithDB(db *gorm.DB) *AgentRepository {
	r.db = db
	return r
}

func (r *AgentRepository) Create(agent *model.Agent) error {
	return r.db.Create(agent).Error
}

func (r *AgentRepository) GetByID(id uuid.UUID) (*model.Agent, error) {
	var agent model.Agent
	if err := r.db.Preload("User").First(&agent, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) Update(agent *model.Agent) error {
	return r.db.Save(agent).Error
}

func (r *AgentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Agent{}, "id = ?", id).Error
}

func (r *AgentRepository) ListByUserID(userID uuid.UUID) ([]*model.Agent, error) {
	var agents []*model.Agent
	if err := r.db.Where("user_id = ?", userID).Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (r *AgentRepository) ListByStatus(status model.AgentStatus, limit, offset int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64
	r.db.Model(&model.Agent{}).Where("status = ?", status).Count(&total)
	if err := r.db.Where("status = ?", status).Limit(limit).Offset(offset).Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, total, nil
}

func (r *AgentRepository) ListAll(limit, offset int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64
	r.db.Model(&model.Agent{}).Count(&total)
	if err := r.db.Limit(limit).Offset(offset).Preload("User").Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, total, nil
}

func (r *AgentRepository) SearchBySkills(ctx context.Context, skills []string, location string, experienceMin int, limit int) ([]*model.Agent, error) {
	var agents []*model.Agent
	db := r.db.WithContext(ctx).Model(&model.Agent{})

	// Filter by agent type (seeker agents only for candidate search)
	db = db.Where("type = ?", "seeker")

	if location != "" {
		db = db.Where("location ILIKE ?", "%"+location+"%")
	}

	if err := db.Limit(limit).Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}
