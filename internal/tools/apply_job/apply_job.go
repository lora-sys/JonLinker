package apply_job

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/resume"
	"github.com/lora-sys/JonLinker/internal/session"
)

type Tool struct {
	sessionStore *session.Store
	fcAPIKey     string
}

func NewTool(store *session.Store, fcAPIKey string) *Tool {
	return &Tool{sessionStore: store, fcAPIKey: fcAPIKey}
}

type applyArgs struct {
	JobURL    string `json:"job_url"`
	SessionID string `json:"session_id"`
}

func (t *Tool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "apply_job",
		Desc: "为指定职位生成求职申请（求职信 + 定制简历）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"job_url": {
				Type:     schema.String,
				Desc:     "职位链接",
				Required: true,
			},
			"session_id": {
				Type:     schema.String,
				Desc:     "会话ID",
				Required: true,
			},
		}),
	}, nil
}

func (t *Tool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var args applyArgs
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	var profile resume.CandidateProfile
	if err := t.sessionStore.Get(ctx, args.SessionID+":profile", &profile); err != nil {
		return "", fmt.Errorf("get profile: %w", err)
	}

	jobDetail, err := t.fetchJob(ctx, args.JobURL)
	if err != nil {
		return "", fmt.Errorf("fetch job: %w", err)
	}

	data := map[string]any{
		"candidate": profile,
		"job":       jobDetail,
	}
	b, _ := json.Marshal(data)
	return string(b), nil
}

func (t *Tool) fetchJob(ctx context.Context, url string) (*job.Job, error) {
	body := map[string]any{"url": url}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/scrape", strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.fcAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("firecrawl scrape: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("scrape failed (%d): %s", resp.StatusCode, string(raw))
	}

	title := extractTitle(string(raw))
	return &job.Job{
		Title:       title,
		URL:         url,
		Description: string(raw),
		Source:      "Indeed",
	}, nil
}

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
	store   *session.Store
	fcKey   string
	baseURL string
	apiKey  string
	model   string
}

func NewDirectApply(store *session.Store, fcKey, baseURL, apiKey, model string) *DirectApply {
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
	if err := d.store.Get(ctx, sessionID+":profile", &profile); err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	reqBody := map[string]any{"url": jobURL}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/scrape", strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.fcKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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

	llmResp, err := http.Post(d.baseURL+"/chat/completions",
		"application/json", strings.NewReader(string(llmBody)))
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

	if err := d.store.Set(ctx, sessionID+":application_"+jobURL, app); err != nil {
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
