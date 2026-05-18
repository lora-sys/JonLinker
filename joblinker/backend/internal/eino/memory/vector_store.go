package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/schema"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"
)

// VectorRetriever implements Eino Retriever interface for vector search
type VectorRetriever struct {
	topK    int
	vecRepo *repository.VectorRepository
}

// NewVectorRetriever creates a new Eino Retriever connected to a vector repository
func NewVectorRetriever(vecRepo *repository.VectorRepository) *VectorRetriever {
	return &VectorRetriever{
		topK:    5,
		vecRepo: vecRepo,
	}
}

// NewVectorRetrieverWithTopK creates with custom topK
func NewVectorRetrieverWithTopK(vecRepo *repository.VectorRepository, topK int) *VectorRetriever {
	return &VectorRetriever{
		topK:    topK,
		vecRepo: vecRepo,
	}
}

// Retrieve searches for relevant documents using vector similarity
// Implements Eino Retriever interface
func (r *VectorRetriever) Retrieve(ctx context.Context, query string, opts ...Option) ([]*schema.Document, error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	topK := r.topK
	if options.TopK != nil {
		topK = *options.TopK
	}

	if r.vecRepo == nil {
		log.Printf("VectorRetriever: no vector repository configured, returning empty results for query: %q", query)
		return []*schema.Document{}, nil
	}

	// Generate embedding from the query using AI API
	embedding, err := r.generateEmbedding(ctx, query)
	if err != nil {
		log.Printf("VectorRetriever: failed to generate embedding for query: %v", err)
		return []*schema.Document{}, nil
	}

	ids, scores, err := r.vecRepo.SearchSimilar(ctx, "agent_memories", embedding, topK)
	if err != nil {
		log.Printf("VectorRetriever search failed: %v", err)
		return []*schema.Document{}, nil
	}

	documents := make([]*schema.Document, 0, len(ids))
	for i, id := range ids {
		meta := map[string]interface{}{
			"id":    id,
			"score": scores[i],
		}
		metaBytes, _ := json.Marshal(meta)
		documents = append(documents, &schema.Document{
			ID:       id,
			Content:  string(metaBytes),
			MetaData: meta,
		})
	}

	log.Printf("VectorRetriever: found %d documents for query: %q", len(documents), query)
	return documents, nil
}

// Option configures the Retriever
type Option func(*Options)

// Options holds Retriever configuration
type Options struct {
	TopK            *int
	ScoreThreshold  *float64
}

// WithTopK sets the maximum number of results to return
func WithTopK(k int) Option {
	return func(o *Options) {
		o.TopK = &k
	}
}

// WithScoreThreshold sets minimum similarity score
func WithScoreThreshold(threshold float64) Option {
	return func(o *Options) {
		o.ScoreThreshold = &threshold
	}
}

// generateEmbedding generates an embedding vector for the given text
func (r *VectorRetriever) generateEmbedding(ctx context.Context, text string) ([]float64, error) {
	aiClient := ai.NewClient()
	embedding, err := aiClient.GenerateEmbedding(text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	return embedding, nil
}
