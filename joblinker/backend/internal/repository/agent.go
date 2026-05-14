package repository

import (
	"context"
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// tenantScope returns a GORM scope that filters by tenant_id.
// If tenantID is empty, no filtering is applied (admin cross-tenant).
func tenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantID == "" {
			return db
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}

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

func (r *AgentRepository) ListByUserID(userID uuid.UUID, tenantID string) ([]*model.Agent, error) {
	var agents []*model.Agent
	db := r.db.Scopes(tenantScope(tenantID)).Where("user_id = ?", userID)
	if err := db.Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (r *AgentRepository) ListByStatus(status model.AgentStatus, tenantID string, limit, offset int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64
	db := r.db.Scopes(tenantScope(tenantID)).Where("status = ?", status)
	db.Model(&model.Agent{}).Count(&total)
	if err := db.Limit(limit).Offset(offset).Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, total, nil
}

func (r *AgentRepository) ListAll(tenantID string, limit, offset int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64
	db := r.db.Scopes(tenantScope(tenantID))
	db.Model(&model.Agent{}).Count(&total)
	if err := db.Limit(limit).Offset(offset).Preload("User").Find(&agents).Error; err != nil {
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
