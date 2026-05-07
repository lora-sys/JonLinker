package service

import (
	"context"
	"log"

	"joblinker/internal/repository"
	"joblinker/pkg/ai"

	"github.com/google/uuid"
)

type VectorService struct {
	vectorRepo *repository.VectorRepository
	aiClient   *ai.Client
}

func NewVectorService(vectorRepo *repository.VectorRepository) *VectorService {
	return &VectorService{vectorRepo: vectorRepo}
}

// NewVectorServiceWithAI creates a VectorService with AI client for real embeddings
func NewVectorServiceWithAI(vectorRepo *repository.VectorRepository, aiClient *ai.Client) *VectorService {
	return &VectorService{vectorRepo: vectorRepo, aiClient: aiClient}
}

func (s *VectorService) GenerateEmbedding(text string) ([]float64, error) {
	// Try AI API first if client is configured
	if s.aiClient != nil {
		embedding, err := s.aiClient.GenerateEmbedding(text)
		if err != nil {
			log.Printf("WARNING: AI embedding failed, falling back to TF: %v", err)
		} else {
			return embedding, nil
		}
	}
	// Fallback to simple TF-based embedding
	words := tokenize(text)
	tf := computeTF(words)
	embedding := make([]float64, len(words))
	i := 0
	for _, count := range tf {
		embedding[i] = count
		i++
	}
	return embedding, nil
}

func tokenize(text string) []string {
	var words []string
	var current []byte
	for _, c := range text {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			current = append(current, byte(c))
		} else if len(current) > 0 {
			words = append(words, string(current))
			current = nil
		}
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}

func computeTF(tokens []string) map[string]float64 {
	tf := make(map[string]float64)
	for _, token := range tokens {
		tf[token]++
	}
	for token := range tf {
		tf[token] /= float64(len(tokens))
	}
	return tf
}

func (s *VectorService) StoreEmbedding(ctx context.Context, collectionID string, text string, metadata map[string]interface{}) (string, error) {
	embedding, err := s.GenerateEmbedding(text)
	if err != nil {
		return "", err
	}

	id := uuid.New().String()
	err = s.vectorRepo.StoreVector(ctx, collectionID, id, embedding, text, metadata)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *VectorService) SearchSimilar(ctx context.Context, collectionID string, query string, limit int) ([]string, []float64, error) {
	embedding, err := s.GenerateEmbedding(query)
	if err != nil {
		return nil, nil, err
	}
	return s.vectorRepo.SearchSimilar(ctx, collectionID, embedding, limit)
}
