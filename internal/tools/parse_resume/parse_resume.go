package parse_resume

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type Tool struct {
	apiKey string
}

func NewTool(apiKey string) *Tool {
	return &Tool{apiKey: apiKey}
}

func (t *Tool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "parse_resume_pdf",
		Desc: "解析 PDF 简历文件，提取文本内容",
	}, nil
}

func (t *Tool) InvokableRun(ctx context.Context, _ string, _ ...tool.Option) (string, error) {
	return "", fmt.Errorf("parse_resume_pdf: use ParseFile instead")
}

func (t *Tool) ParseFile(ctx context.Context, filename string, data []byte) (string, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(fw, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("copy file data: %w", err)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/parse", body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("firecrawl parse: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("firecrawl parse failed (%d): %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Data struct {
			Markdown string `json:"markdown"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.Data.Markdown, nil
}
