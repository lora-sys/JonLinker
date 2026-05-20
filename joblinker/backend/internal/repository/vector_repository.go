package repository

import (
	"context"
	"log"

	redispkg "joblinker/pkg/redis"
)

// VectorRepository provides vector storage and similarity search via Redis Stack
type VectorRepository struct {
	rvClient *redispkg.VectorClient
}

// NewVectorRepository creates a VectorRepository backed by Redis Stack
func NewVectorRepository(rvClient *redispkg.VectorClient) *VectorRepository {
	return &VectorRepository{rvClient: rvClient}
}

// StoreVector stores a document with its vector embedding
func (r *VectorRepository) StoreVector(ctx context.Context, collectionID string, id string, vector []float64, document string, metadata map[string]interface{}) error {
	if r.rvClient == nil {
		log.Printf("VectorRepository: Redis client not configured, skipping vector storage")
		return nil
	}
	doc := &redispkg.VectorDocument{
		ID:           id,
		CollectionID: collectionID,
		Content:      document,
		Vector:       vector,
		Metadata:     metadata,
	}
	return r.rvClient.Store(ctx, doc)
}

// SearchSimilar performs vector similarity search using Redis Stack KNN
func (r *VectorRepository) SearchSimilar(ctx context.Context, collectionID string, queryVector []float64, limit int) ([]string, []float64, error) {
	if r.rvClient == nil {
		log.Printf("VectorRepository: Redis client not configured, returning empty results")
		return nil, nil, nil
	}
	results, err := r.rvClient.SearchSimilar(ctx, collectionID, collectionID, queryVector, limit)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, len(results))
	scores := make([]float64, len(results))
	for i, r := range results {
		ids[i] = r.ID
		scores[i] = r.Score
	}
	return ids, scores, nil
}
