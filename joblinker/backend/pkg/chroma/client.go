package chroma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	host   string
	port   int
	client *http.Client
}

func NewClient(host string, port int) *Client {
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 8000
	}
	return &Client{
		host: host,
		port: port,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) baseURL() string {
	return fmt.Sprintf("http://%s:%d/api/v1", c.host, c.port)
}

func (c *Client) Heartbeat() error {
	resp, err := c.client.Get(c.baseURL() + "/heartbeat")
	if err != nil {
		return fmt.Errorf("chroma heartbeat failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma heartbeat returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

type Collection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Dimension any    `json:"dimension,omitempty"`
	Metadata  any    `json:"metadata,omitempty"`
}

func (c *Client) GetOrCreateCollection(name string) (Collection, error) {
	payload := map[string]interface{}{
		"name":         name,
		"get_or_create": true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Collection{}, fmt.Errorf("failed to marshal collection payload: %w", err)
	}

	url := c.baseURL() + "/collections"
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return Collection{}, fmt.Errorf("failed to create collection %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return Collection{}, fmt.Errorf("create collection returned %d: %s", resp.StatusCode, string(respBody))
	}

	var col Collection
	if err := json.NewDecoder(resp.Body).Decode(&col); err != nil {
		return Collection{}, fmt.Errorf("failed to decode collection response: %w", err)
	}
	return col, nil
}

func (c *Client) Add(collectionID string, ids, documents []string, metadatas []map[string]interface{}) error {
	if len(ids) == 0 || len(documents) == 0 {
		return nil
	}
	if len(ids) != len(documents) {
		return fmt.Errorf("ids (%d) and documents (%d) length mismatch", len(ids), len(documents))
	}

	payload := map[string]interface{}{
		"ids":        ids,
		"documents": documents,
		"metadatas": metadatas,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal add payload: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/add", c.baseURL(), collectionID)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type QueryResult struct {
	IDs        [][]string                  `json:"ids"`
	Distances  [][]float64                 `json:"distances,omitempty"`
	Metadatas  [][]map[string]interface{} `json:"metadatas,omitempty"`
	Documents  [][]string                  `json:"documents,omitempty"`
	Embeddings any                         `json:"embeddings,omitempty"`
}

func (c *Client) Query(collectionID string, queryTexts []string, nResults int, where map[string]interface{}) (QueryResult, error) {
	payload := map[string]interface{}{
		"query_texts": queryTexts,
		"n_results":   nResults,
	}
	if where != nil {
		payload["where"] = where
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return QueryResult{}, fmt.Errorf("failed to marshal query payload: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/query", c.baseURL(), collectionID)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return QueryResult{}, fmt.Errorf("failed to query collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return QueryResult{}, fmt.Errorf("query returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return QueryResult{}, fmt.Errorf("failed to decode query result: %w", err)
	}
	return result, nil
}

func (c *Client) Delete(collectionID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	payload := map[string]interface{}{
		"ids": ids,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal delete payload: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/delete", c.baseURL(), collectionID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (c *Client) Get(collectionID string, ids []string, include []string) ([]map[string]interface{}, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	payload := map[string]interface{}{
		"ids":     ids,
		"include": include,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal get payload: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/get", c.baseURL(), collectionID)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		IDs        []string                   `json:"ids"`
		Metadatas  [][]map[string]interface{} `json:"metadatas,omitempty"`
		Documents  [][]string                 `json:"documents,omitempty"`
		Embeddings any                       `json:"embeddings,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode get result: %w", err)
	}

	docs := make([]map[string]interface{}, len(ids))
	for i, id := range result.IDs {
		docs[i] = map[string]interface{}{"id": id}
		if len(result.Documents) > i {
			docs[i]["document"] = result.Documents[i]
		}
		if len(result.Metadatas) > i {
			docs[i]["metadata"] = result.Metadatas[i]
		}
	}
	return docs, nil
}