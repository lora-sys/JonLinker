package retriever

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

type RedisRetriever struct {
	client      *redis.Client
	embedder    Embedder
}

type Embedder interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)
}

func NewRedisRetriever(client *redis.Client, embedder Embedder) *RedisRetriever {
	return &RedisRetriever{
		client:   client,
		embedder: embedder,
	}
}

func (r *RedisRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	topK := 10
	scoreThreshold := 0.0
	collection := "default_docs"

	// Generate query embedding
	var queryVec []float64
	if r.embedder != nil && query != "" {
		vectors, err := r.embedder.EmbedStrings(ctx, []string{query})
		if err != nil {
			return nil, fmt.Errorf("redis retriever: embed query: %w", err)
		}
		if len(vectors) > 0 {
			queryVec = vectors[0]
		}
	}

	// If we have a vector, use FT.SEARCH with KNN
	if len(queryVec) > 0 {
		return r.searchByVector(ctx, collection, queryVec, topK, scoreThreshold)
	}

	// Fallback to text search
	return r.searchByText(ctx, collection, query, topK)
}

func (r *RedisRetriever) searchByVector(ctx context.Context, collection string, queryVec []float64, topK int, scoreThreshold float64) ([]*schema.Document, error) {
	// FT.SEARCH with KNN
	// FT.SEARCH {collection} "*=>[KNN {topK} @vector $BLOB]" PARAMS 2 BLOB {vecBlob} RETURN 3 content weight vector SORTBY __vector_score DIALECT 4
	float32Vec := make([]float32, len(queryVec))
	for i, v := range queryVec {
		float32Vec[i] = float32(v)
	}

	vecBlob := float32VecToBytes(float32Vec)

	cmdStr := fmt.Sprintf(
		"FT.SEARCH %s *=>[KNN %d @vector $BLOB] PARAMS 2 BLOB %s RETURN 3 content weight vector SORTBY __vector_score DIALECT 4",
		collection, topK, blobPlaceholder,
	)

	result, err := r.client.Do(ctx, cmdStr, vecBlob).Result()
	if err != nil {
		// Index might not exist, try text search
		return r.searchByText(ctx, collection, "", topK)
	}

	docs, err := parseFTSearchResult(result, scoreThreshold)
	if err != nil {
		return nil, fmt.Errorf("redis retriever: parse result: %w", err)
	}

	return docs, nil
}

func (r *RedisRetriever) searchByText(ctx context.Context, collection string, query string, topK int) ([]*schema.Document, error) {
	// FT.SEARCH {collection} "@content:{query}" RETURN 3 content weight
	cmdStr := fmt.Sprintf("FT.SEARCH %s @content:%s RETURN 3 content weight LIMIT 0 %d DIALECT 4",
		collection, query, topK)

	result, err := r.client.Do(ctx, cmdStr).Result()
	if err != nil {
		return []*schema.Document{}, nil
	}

	docs, err := parseFTSearchResult(result, 0)
	if err != nil {
		return nil, fmt.Errorf("redis retriever: parse text result: %w", err)
	}

	return docs, nil
}

func float32VecToBytes(vec []float32) string {
	data := make([]byte, len(vec)*4)
	for i, v := range vec {
		bits := float32ToBytes(v)
		data[i*4] = bits[0]
		data[i*4+1] = bits[1]
		data[i*4+2] = bits[2]
		data[i*4+3] = bits[3]
	}
	return string(data)
}

func float32ToBytes(v float32) [4]byte {
	n := uint32(v)
	return [4]byte{
		byte(n),
		byte(n >> 8),
		byte(n >> 16),
		byte(n >> 24),
	}
}

func parseFTSearchResult(result interface{}, scoreThreshold float64) ([]*schema.Document, error) {
	results, ok := result.([]interface{})
	if !ok || len(results) < 2 {
		return []*schema.Document{}, nil
	}

	// First element is total count
	totalStr := fmt.Sprintf("%v", results[0])
	total, err := strconv.Atoi(totalStr)
	if err != nil {
		total = 0
	}

	docs := make([]*schema.Document, 0, total)
	for i := 1; i+1 < len(results); i += 2 {
		key := fmt.Sprintf("%v", results[i])
		fields, ok := results[i+1].([]interface{})
		if !ok {
			continue
		}

		doc := &schema.Document{
			ID:       key,
			MetaData: make(map[string]interface{}),
		}

		for j := 0; j+1 < len(fields); j += 2 {
			fieldName := fmt.Sprintf("%v", fields[j])
			fieldVal := fmt.Sprintf("%v", fields[j+1])

			switch fieldName {
			case "content":
				doc.Content = fieldVal
			case "__vector_score":
				score, _ := strconv.ParseFloat(fieldVal, 64)
				doc.MetaData["score"] = score
				if scoreThreshold > 0 && score > scoreThreshold {
					return nil, nil
				}
			default:
				doc.MetaData[fieldName] = fieldVal
			}
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

const blobPlaceholder = "$blob"
