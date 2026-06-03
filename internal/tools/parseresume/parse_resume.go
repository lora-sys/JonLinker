package parseresume

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Tool struct {
	apiKey string
}

func NewTool(apiKey string) *Tool {
	return &Tool{apiKey: apiKey}
}

func (t *Tool) ParseFile(ctx context.Context, filename string, data []byte) (string, error) {
	// Try Firecrawl API for parsing
	result, err := t.parseViaAPI(ctx, filename, data)
	if err == nil {
		return result, nil
	}
	// Fall back to raw text
	return string(data), nil
}

func (t *Tool) parseViaAPI(ctx context.Context, filename string, data []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("write file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.firecrawl.dev/v1/scrape", body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("firecrawl API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("firecrawl API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}
