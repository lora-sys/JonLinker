package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisIndexer struct {
	client      *redis.Client
	embedder    Embedder
	initialized map[string]bool
}

type Embedder interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)
}

func NewRedisIndexer(client *redis.Client, embedder Embedder) *RedisIndexer {
	return &RedisIndexer{
		client:      client,
		embedder:    embedder,
		initialized: make(map[string]bool),
	}
}

func (idx *RedisIndexer) ensureIndex(ctx context.Context, collection string, dim int) error {
	if idx.initialized[collection] {
		return nil
	}

	exists, err := idx.client.Exists(ctx, collection+":idx").Result()
	if err != nil {
		return fmt.Errorf("redis indexer: check index: %w", err)
	}

	if exists == 0 {
		// Create the index flag key and set up the vector index
		// FT.CREATE {collection} ON HASH PREFIX 1 {collection}: SCHEMA doc JSON vector VECTOR FLAT 6 TYPE FLOAT32 DIM 1024 DISTANCE_METRIC COSINE
		createCmd := fmt.Sprintf(
			"FT.CREATE %s ON HASH PREFIX 1 %s: SCHEMA content TEXT weight TEXT vector VECTOR FLAT 6 TYPE FLOAT32 DIM %d DISTANCE_METRIC COSINE",
			collection, collection, dim,
		)
		if err := idx.client.Do(ctx, createCmd).Err(); err != nil {
			// Index might already exist (race condition)
			// Just log and continue
		}
		idx.client.Set(ctx, collection+":idx", "1", 0)
	}

	idx.initialized[collection] = true
	return nil
}

func (idx *RedisIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	if len(docs) == 0 {
		return nil, nil
	}

	collection := "default_docs"

	dim := 1024 // Jina AI v5 is 1024 dimensions
	if err := idx.ensureIndex(ctx, collection, dim); err != nil {
		return nil, err
	}

	ids := make([]string, len(docs))
	for i, doc := range docs {
		docID := doc.ID
		if docID == "" {
			docID = uuid.New().String()
		}
		ids[i] = docID

		// Generate embedding if embedder is available and doc has content
		var vector []float64
		if idx.embedder != nil && doc.Content != "" {
			vectors, err := idx.embedder.EmbedStrings(ctx, []string{doc.Content})
			if err == nil && len(vectors) > 0 {
				vector = vectors[0]
			}
		}

		// Store document metadata as JSON
		meta := map[string]interface{}{
			"content":   doc.Content,
			"createdAt": time.Now().UTC().Format(time.RFC3339),
		}
		if doc.MetaData != nil {
			for k, v := range doc.MetaData {
				meta[k] = v
			}
		}

		metaJSON, _ := json.Marshal(meta)

		// Use JSON.SET for the document
		key := fmt.Sprintf("%s:%s", collection, docID)
		if err := idx.client.JSONSet(ctx, key, "$", string(metaJSON)).Err(); err != nil {
			return nil, fmt.Errorf("redis indexer: store doc %s: %w", docID, err)
		}

		// Store vector as a separate hash field for FT.SEARCH compatibility
		if len(vector) > 0 {
			float32Vec := make([]float32, len(vector))
			for j, v := range vector {
				float32Vec[j] = float32(v)
			}
			if err := idx.client.HSet(ctx, key+":vec", "vector", float32Vec).Err(); err != nil {
				return nil, fmt.Errorf("redis indexer: store vector %s: %w", docID, err)
			}
		}
	}

	return ids, nil
}
