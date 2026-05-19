package redis

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	goredis "github.com/redis/go-redis/v9"
)

// VectorClient wraps a Redis Stack client for vector storage and search
type VectorClient struct {
	client    *goredis.Client
	prefix    string
	dimension int
}

// NewVectorClient creates a new Redis Stack vector client
func NewVectorClient(client *goredis.Client, prefix string, dimension int) *VectorClient {
	return &VectorClient{
		client:    client,
		prefix:    prefix,
		dimension: dimension,
	}
}

// EnsureIndex creates a RediSearch index for vector similarity search if it doesn't exist
func (c *VectorClient) EnsureIndex(ctx context.Context, indexName string) error {
	fullIndex := c.prefix + indexName

	_, err := c.client.Do(ctx, "FT.INFO", fullIndex).Result()
	if err == nil {
		log.Printf("Redis vector index %s already exists", fullIndex)
		return nil
	}

	args := []interface{}{
		"FT.CREATE", fullIndex,
		"ON", "JSON",
		"PREFIX", "1", c.prefix + "doc:",
		"SCHEMA",
		"$.content", "AS", "content", "TEXT",
		"$.collection_id", "AS", "collection_id", "TAG",
		"$.metadata", "AS", "metadata", "TEXT",
		"$.vector", "AS", "vector", "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT64",
		"DIM", c.dimension,
		"DISTANCE_METRIC", "COSINE",
	}

	_, err = c.client.Do(ctx, args...).Result()
	if err != nil {
		return fmt.Errorf("failed to create redis vector index %s: %w", fullIndex, err)
	}
	log.Printf("Created Redis vector index %s (dim=%d, HNSW, COSINE)", fullIndex, c.dimension)
	return nil
}

// VectorDocument represents a document stored in Redis with vector embedding
type VectorDocument struct {
	ID           string                 `json:"id"`
	CollectionID string                 `json:"collection_id"`
	Content      string                 `json:"content"`
	Vector       []float64              `json:"vector"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Store stores a document with its vector embedding in Redis
func (c *VectorClient) Store(ctx context.Context, doc *VectorDocument) error {
	key := c.prefix + "doc:" + doc.ID
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal vector document: %w", err)
	}
	return c.client.Do(ctx, "JSON.SET", key, "$", string(data)).Err()
}

// Get retrieves a document by ID
func (c *VectorClient) Get(ctx context.Context, id string) (*VectorDocument, error) {
	key := c.prefix + "doc:" + id
	result, err := c.client.Do(ctx, "JSON.GET", key).Text()
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}
	var doc VectorDocument
	if err := json.Unmarshal([]byte(result), &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}
	return &doc, nil
}

// Delete removes a document by ID
func (c *VectorClient) Delete(ctx context.Context, id string) error {
	key := c.prefix + "doc:" + id
	return c.client.Del(ctx, key).Err()
}

// SearchResult represents a single search result
type SearchResult struct {
	ID      string
	Score   float64
	Content string
}

// SearchSimilar performs vector similarity search using RediSearch KNN
func (c *VectorClient) SearchSimilar(ctx context.Context, indexName string, collectionID string, queryVector []float64, limit int) ([]SearchResult, error) {
	fullIndex := c.prefix + indexName

	vectorBytes := float64SliceToBytes(queryVector)

	filterQuery := "*"
	if collectionID != "" {
		filterQuery = fmt.Sprintf("@collection_id:{%s}", collectionID)
	}

	query := fmt.Sprintf("(%s)=>[KNN %d @vector $BLOB AS score]", filterQuery, limit)

	args := []interface{}{
		"FT.SEARCH", fullIndex,
		query,
		"PARAMS", "2", "BLOB", vectorBytes,
		"SORTBY", "score",
		"LIMIT", "0", strconv.Itoa(limit),
		"RETURN", "2", "content", "score",
		"DIALECT", "2",
	}

	result, err := c.client.Do(ctx, args...).Slice()
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	return parseSearchResults(result)
}

// Ping checks Redis connectivity
func (c *VectorClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func float64SliceToBytes(vec []float64) []byte {
	buf := make([]byte, len(vec)*8)
	for i, v := range vec {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
	}
	return buf
}

func parseSearchResults(raw []interface{}) ([]SearchResult, error) {
	if len(raw) < 1 {
		return nil, nil
	}

	var results []SearchResult
	for i := 1; i+1 < len(raw); i += 2 {
		fields, ok := raw[i+1].([]interface{})
		if !ok {
			continue
		}
		sr := SearchResult{}
		if id, ok := raw[i].(string); ok {
			sr.ID = id
		}
		for j := 0; j+1 < len(fields); j += 2 {
			fieldName, _ := fields[j].(string)
			fieldVal, _ := fields[j+1].(string)
			switch fieldName {
			case "content":
				sr.Content = fieldVal
			case "score":
				if v, err := strconv.ParseFloat(fieldVal, 64); err == nil {
					sr.Score = v
				}
			}
		}
		results = append(results, sr)
	}
	return results, nil
}
