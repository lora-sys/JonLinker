package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository() *OrganizationRepository {
	return &OrganizationRepository{db: nil}
}

func (r *OrganizationRepository) WithDB(db *gorm.DB) *OrganizationRepository {
	r.db = db
	return r
}

func (r *OrganizationRepository) Create(org *model.Organization) error {
	return r.db.Create(org).Error
}

func (r *OrganizationRepository) GetByID(id uuid.UUID) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.First(&org, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) Update(org *model.Organization) error {
	return r.db.Save(org).Error
}

func (r *OrganizationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Organization{}, "id = ?", id).Error
}

func (r *OrganizationRepository) ListAll(limit, offset int) ([]*model.Organization, int64, error) {
	var orgs []*model.Organization
	var total int64
	r.db.Model(&model.Organization{}).Count(&total)
	if err := r.db.Limit(limit).Offset(offset).Find(&orgs).Error; err != nil {
		return nil, 0, err
	}
	return orgs, total, nil
}
