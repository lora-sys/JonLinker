package repository

import (
	"joblinker/internal/model"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type PreferenceVectorRepository struct {
	db *gorm.DB
}

func NewPreferenceVectorRepository() *PreferenceVectorRepository {
	return &PreferenceVectorRepository{db: nil}
}

func (r *PreferenceVectorRepository) WithDB(db *gorm.DB) *PreferenceVectorRepository {
	r.db = db
	return r
}

func (r *PreferenceVectorRepository) Create(pref *model.UserPreferenceVector) error {
	return r.db.Create(pref).Error
}

func (r *PreferenceVectorRepository) GetByID(id uuid.UUID) (*model.UserPreferenceVector, error) {
	var pref model.UserPreferenceVector
	if err := r.db.First(&pref, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *PreferenceVectorRepository) GetByUserID(userID uuid.UUID) ([]*model.UserPreferenceVector, error) {
	var prefs []*model.UserPreferenceVector
	if err := r.db.Where("user_id = ?", userID).Find(&prefs).Error; err != nil {
		return nil, err
	}
	return prefs, nil
}

func (r *PreferenceVectorRepository) GetByUserAndType(userID uuid.UUID, prefType model.PreferenceType) ([]*model.UserPreferenceVector, error) {
	var prefs []*model.UserPreferenceVector
	if err := r.db.Where("user_id = ? AND preference_type = ?", userID, prefType).Find(&prefs).Error; err != nil {
		return nil, err
	}
	return prefs, nil
}

func (r *PreferenceVectorRepository) Update(pref *model.UserPreferenceVector) error {
	return r.db.Save(pref).Error
}

func (r *PreferenceVectorRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.UserPreferenceVector{}, "id = ?", id).Error
}

func (r *PreferenceVectorRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Delete(&model.UserPreferenceVector{}, "user_id = ?", userID).Error
}

// SearchSimilar finds preference vectors similar to the given embedding
// In production, this would use pgvector's vector similarity search
func (r *PreferenceVectorRepository) SearchSimilar(userID uuid.UUID, prefType model.PreferenceType, embedding pq.Float64Array, limit int) ([]*model.UserPreferenceVector, error) {
	var prefs []*model.UserPreferenceVector
	if err := r.db.Where("user_id = ? AND preference_type = ?", userID, prefType).Limit(limit).Find(&prefs).Error; err != nil {
		return nil, err
	}
	return prefs, nil
}