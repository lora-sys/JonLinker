package repository

import (
	"context"
	"math"

	"gorm.io/gorm"
)

type VectorRepository struct {
	db *gorm.DB
}

func NewVectorRepository() *VectorRepository {
	return &VectorRepository{db: nil}
}

func (r *VectorRepository) WithDB(db *gorm.DB) *VectorRepository {
	r.db = db
	return r
}

type VectorRecord struct {
	ID           string  `json:"id" gorm:"column:id"`
	CollectionID string  `json:"collection_id" gorm:"column:collection_id"`
	Vector       []float64 `json:"vector" gorm:"column:vector"`
	Document     string  `json:"document" gorm:"column:document"`
	Metadata     string  `json:"metadata" gorm:"column:metadata"`
}

func (r *VectorRepository) StoreVector(ctx context.Context, collectionID string, id string, vector []float64, document string, metadata map[string]interface{}) error {
	// This would integrate with Chroma in production
	// For now, we store the mapping in PostgreSQL
	record := &VectorRecord{
		ID:           id,
		CollectionID: collectionID,
		Vector:       vector,
		Document:     document,
	}
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *VectorRepository) GetVector(ctx context.Context, id string) ([]float64, error) {
	var record VectorRecord
	if err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return record.Vector, nil
}

func (r *VectorRepository) DeleteVector(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&VectorRecord{}, "id = ?", id).Error
}

func (r *VectorRepository) SearchSimilar(ctx context.Context, collectionID string, queryVector []float64, limit int) ([]string, []float64, error) {
	// This is a simplified implementation
	// In production, this would call Chroma's API
	var records []VectorRecord
	if err := r.db.WithContext(ctx).
		Where("collection_id = ?", collectionID).
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, nil, err
	}

	ids := make([]string, len(records))
	scores := make([]float64, len(records))
	for i, record := range records {
		ids[i] = record.ID
		scores[i] = cosineSimilarity(queryVector, record.Vector)
	}
	return ids, scores, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	denom := sqrt(normA) * sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dotProduct / denom
}

func sqrt(x float64) float64 {
	return math.Sqrt(x)
}
