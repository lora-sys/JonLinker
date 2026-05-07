package service

import (
	"context"
	"fmt"

	"joblinker/pkg/chroma"

	"github.com/google/uuid"
)

type VectorService struct {
	chromaClient *chroma.Client
}

func NewVectorService(chromaClient *chroma.Client) *VectorService {
	return &VectorService{chromaClient: chromaClient}
}

func (s *VectorService) StoreEmbedding(ctx context.Context, collectionID string, text string, metadata map[string]interface{}) (string, error) {
	id := uuid.New().String()
	err := s.chromaClient.Add(collectionID, []string{id}, []string{text}, []map[string]interface{}{metadata})
	if err != nil {
		return "", fmt.Errorf("failed to store embedding in Chroma: %w", err)
	}
	return id, nil
}

func (s *VectorService) SearchSimilar(ctx context.Context, collectionID string, query string, limit int) ([]string, []float64, error) {
	result, err := s.chromaClient.Query(collectionID, []string{query}, limit, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("chroma query failed: %w", err)
	}
	if len(result.IDs) == 0 || len(result.IDs[0]) == 0 {
		return []string{}, []float64{}, nil
	}
	return result.IDs[0], result.Distances[0], nil
}