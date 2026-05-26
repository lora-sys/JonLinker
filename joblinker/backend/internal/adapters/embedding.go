package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// EmbeddingClient generates embeddings via Jina AI.
type EmbeddingClient struct {
	apiKey     string
	model      string
	task       string
	httpClient *http.Client
}

// NewEmbeddingClient creates a new EmbeddingClient.
func NewEmbeddingClient() *EmbeddingClient {
	return &EmbeddingClient{
		apiKey: getEnv("JINA_API_KEY", ""),
		model:  getEnv("JINA_EMBEDDING_MODEL", "jina-embeddings-v3"),
		task:   getEnv("JINA_EMBEDDING_TASK", "text-matching"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EmbeddingRequest is the request payload for Jina AI.
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Task  string   `json:"task"`
	Input []string `json:"input"`
}

// EmbeddingResponse is the response from Jina AI.
type EmbeddingResponse struct {
	Model  string      `json:"model"`
	Object string      `json:"object"`
	Usage  EmbedUsage  `json:"usage"`
	Data   []EmbedData `json:"data"`
}

// EmbedUsage holds token usage info.
type EmbedUsage struct {
	TotalTokens int `json:"total_tokens"`
}

// EmbedData holds a single embedding vector.
type EmbedData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// GenerateEmbedding generates an embedding vector for the given text.
func (c *EmbeddingClient) GenerateEmbedding(text string) ([]float64, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("JINA_API_KEY not set")
	}

	reqBody := EmbeddingRequest{
		Model: c.model,
		Task:  c.task,
		Input: []string{text},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.jina.ai/v1/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode, string(body))
	}

	var embedResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embedResp.Data[0].Embedding, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
