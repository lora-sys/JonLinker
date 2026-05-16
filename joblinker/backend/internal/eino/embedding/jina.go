package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/embedding"
)

type JinaEmbedding struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewJinaEmbedding(apiKey string) *JinaEmbedding {
	return &JinaEmbedding{
		apiKey:  apiKey,
		model:   "jina-embeddings-v5-text-small",
		baseURL: "https://api.jina.ai/v1/embeddings",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type jinaReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type jinaResp struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (e *JinaEmbedding) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	body := jinaReq{
		Model: e.model,
		Input: texts,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("jina embedding: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("jina embedding: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jina embedding: http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jina embedding: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jina embedding: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var jinaResp jinaResp
	if err := json.Unmarshal(respBody, &jinaResp); err != nil {
		return nil, fmt.Errorf("jina embedding: parse response: %w", err)
	}

	result := make([][]float64, len(texts))
	for i, d := range jinaResp.Data {
		if i < len(result) {
			result[i] = d.Embedding
		}
	}
	return result, nil
}
