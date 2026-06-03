package applyjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/http"
	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/llm"
	"github.com/lora-sys/JonLinker/internal/resume"
)

var ErrProfileNotFound = errors.New("profile not found")

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
	store     compose.CheckPointStore
	fcKey     string
	chatModel *openai.ChatModel
}

func NewDirectApply(store compose.CheckPointStore, fcKey, baseURL, apiKey, model string) (*DirectApply, error) {
	cm, err := llm.NewChatModel(context.Background(), baseURL, apiKey, model, 4096, 0.3)
	if err != nil {
		return nil, fmt.Errorf("init chat model: %w", err)
	}
	return &DirectApply{
		store:     store,
		fcKey:     fcKey,
		chatModel: cm,
	}, nil
}

func (d *DirectApply) Generate(ctx context.Context, jobURL, sessionID string) (*job.Application, error) {
	var profile resume.CandidateProfile
	data, ok, err := d.store.Get(ctx, sessionID+":profile")
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("%w: 未找到简历信息，请先上传简历", ErrProfileNotFound)
	}
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}

	reqBody := map[string]any{"url": jobURL}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/scrape", strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.fcKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("firecrawl scrape: %w", err)
	}
	defer resp.Body.Close()

	jobRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("scrape failed (%d): %s", resp.StatusCode, string(jobRaw))
	}

	profileJSON, err := profileToJSON(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal profile: %w", err)
	}

	prompt := fmt.Sprintf(`你是一个AI招聘助手。根据以下职位信息和候选人资料，生成求职申请。

职位信息：
%s

候选人资料：
%s

请生成求职信（中文Markdown格式）和针对该职位的定制简历（中文Markdown格式），以及3-5条匹配亮点。

以JSON格式返回，不要包含其他文字：
{"cover_letter":"...", "resume_md":"...", "highlights":["...","..."]}`,
		string(jobRaw), profileJSON)

	result, err := d.chatModel.Generate(ctx, []*schema.Message{{Role: schema.User, Content: prompt}})
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}

	content := llm.ExtractJSONBlock(result.Content)

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

	appData, err := json.Marshal(app)
	if err != nil {
		return nil, fmt.Errorf("marshal application: %w", err)
	}
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

func profileToJSON(p resume.CandidateProfile) (string, error) {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
