package applyjob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"

	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/resume"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

func extractTitle(raw string) string {
	idx := strings.Index(raw, "### [")
	if idx < 0 {
		return ""
	}
	end := strings.Index(raw[idx:], "](")
	if end < 0 {
		return ""
	}
	return raw[idx+5 : idx+end]
}

type DirectApply struct {
	store   compose.CheckPointStore
	fcKey   string
	baseURL string
	apiKey  string
	model   string
}

func NewDirectApply(store compose.CheckPointStore, fcKey, baseURL, apiKey, model string) *DirectApply {
	return &DirectApply{
		store:   store,
		fcKey:   fcKey,
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
	}
}

func (d *DirectApply) Generate(ctx context.Context, jobURL, sessionID string) (*job.Application, error) {
	var profile resume.CandidateProfile
	data, ok, err := d.store.Get(ctx, sessionID+":profile")
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("profile not found")
	}
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}

	reqBody := map[string]any{"url": jobURL}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/scrape", strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.fcKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("firecrawl scrape: %w", err)
	}
	defer resp.Body.Close()

	jobRaw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("scrape failed (%d): %s", resp.StatusCode, string(jobRaw))
	}

	prompt := fmt.Sprintf(`你是一个AI招聘助手。根据以下职位信息和候选人资料，生成求职申请。

职位信息：
%s

候选人资料：
%s

请生成求职信（中文Markdown格式）和针对该职位的定制简历（中文Markdown格式），以及3-5条匹配亮点。

以JSON格式返回，不要包含其他文字：
{"cover_letter":"...", "resume_md":"...", "highlights":["...","..."]}`,
		string(jobRaw), profileToJSON(profile))

	llmReq := map[string]any{
		"model": d.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  4096,
	}
	llmBody, _ := json.Marshal(llmReq)

	llmReq2, _ := http.NewRequestWithContext(ctx, "POST", d.baseURL+"/chat/completions",
		strings.NewReader(string(llmBody)))
	llmReq2.Header.Set("Content-Type", "application/json")
	llmReq2.Header.Set("Authorization", "Bearer "+d.apiKey)
	llmResp, err := httpClient.Do(llmReq2)
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}
	defer llmResp.Body.Close()

	llmRaw, _ := io.ReadAll(llmResp.Body)
	if llmResp.StatusCode != 200 {
		return nil, fmt.Errorf("llm failed (%d): %s", llmResp.StatusCode, string(llmRaw))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(llmRaw, &result); err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in llm response")
	}

	content := result.Choices[0].Message.Content

	if idx := strings.Index(content, "{"); idx >= 0 {
		content = content[idx:]
	}
	if idx := strings.LastIndex(content, "}"); idx >= 0 {
		content = content[:idx+1]
	}

	var gen struct {
		CoverLetter string   `json:"cover_letter"`
		ResumeMD    string   `json:"resume_md"`
		Highlights  []string `json:"highlights"`
	}
	if err := json.Unmarshal([]byte(content), &gen); err != nil {
		return nil, fmt.Errorf("parse generation: %w", err)
	}

	title := extractTitle(string(jobRaw))
	app := &job.Application{
		JobTitle:    title,
		Company:     extractCompany(string(jobRaw), title),
		CoverLetter: gen.CoverLetter,
		ResumeMD:    gen.ResumeMD,
		Highlights:  gen.Highlights,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	appData, _ := json.Marshal(app)
	if err := d.store.Set(ctx, sessionID+":application_"+jobURL, appData); err != nil {
		return nil, fmt.Errorf("save application: %w", err)
	}

	return app, nil
}

func extractCompany(raw, title string) string {
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		if strings.Contains(line, title) && i+1 < len(lines) {
			next := strings.TrimSpace(lines[i+1])
			if next != "" && !strings.HasPrefix(next, "|") && !strings.HasPrefix(next, "#") {
				return next
			}
		}
	}
	return ""
}

func profileToJSON(p resume.CandidateProfile) string {
	b, _ := json.MarshalIndent(p, "", "  ")
	return string(b)
}
