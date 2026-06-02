package queryjobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/http"
	"github.com/lora-sys/JonLinker/internal/job"
)

var jobRe = regexp.MustCompile(`^\|\s*###\s+\[(.+?)\]\((.+?)\)(?:<br>(.+?)(?:<br>(.+?))?)?\s*\|`)

type Tool struct {
	apiKey string
}

func NewTool(apiKey string) *Tool {
	return &Tool{apiKey: apiKey}
}

func (t *Tool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "query_jobs",
		Desc: "搜索招聘职位，按关键词和城市筛选。支持中文搜索，返回 JSON 格式的职位列表。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keyword": {
				Type:     schema.String,
				Desc:     "职位关键词，如 前端、Go开发、产品经理",
				Required: true,
			},
			"city": {
				Type: schema.String,
				Desc: "工作城市，如 北京、上海、杭州（可选）",
			},
		}),
	}, nil
}

type queryArgs struct {
	Keyword string `json:"keyword"`
	City    string `json:"city"`
}

func (t *Tool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args queryArgs
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	jobs, err := t.searchIndeed(ctx, args.Keyword, args.City)
	if err != nil {
		return fmt.Sprintf("搜索职位出错: %v", err), nil
	}

	data, err := json.Marshal(jobs)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}
	return string(data), nil
}

func (t *Tool) searchIndeed(ctx context.Context, keyword, city string) ([]job.Job, error) {
	searchURL := fmt.Sprintf("https://cn.indeed.com/jobs?q=%s", url.QueryEscape(keyword))
	if city != "" {
		searchURL += "&l=" + url.QueryEscape(city)
	}

	body, err := json.Marshal(map[string]any{
		"url":     searchURL,
		"formats": []string{"markdown"},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.firecrawl.dev/v1/scrape", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("firecrawl request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var fcResp struct {
		Success bool `json:"success"`
		Data    struct {
			Markdown string `json:"markdown"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &fcResp); err != nil {
		return nil, fmt.Errorf("parse firecrawl response: %w", err)
	}
	if !fcResp.Success {
		return nil, fmt.Errorf("firecrawl failed: %s", string(respBody))
	}

	return parseIndeedJobs(fcResp.Data.Markdown), nil
}

func parseIndeedJobs(md string) []job.Job {
	var jobs []job.Job
	lines := strings.Split(md, "\n")

	var current *job.Job
	collectDesc := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			collectDesc = false
			continue
		}

		if match := jobRe.FindStringSubmatch(trimmed); len(match) >= 3 {
			if current != nil && current.Title != "" {
				jobs = append(jobs, *current)
			}
			current = &job.Job{
				Title:  match[1],
				Source: "Indeed",
			}
			relURL := match[2]
			if strings.HasPrefix(relURL, "//") {
				current.URL = "https:" + relURL
			} else if !strings.HasPrefix(relURL, "http") {
				current.URL = "https://cn.indeed.com" + relURL
			} else {
				current.URL = relURL
			}
			if len(match) >= 4 {
				current.Company = match[3]
			}
			if len(match) >= 5 {
				current.Location = strings.SplitN(match[4], " |", 2)[0]
			}
			collectDesc = true
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, "- ") {
			continue
		}

		if collectDesc {
			if current.Description != "" {
				current.Description += " "
			}
			current.Description += trimmed
		}
	}

	if current != nil && current.Title != "" {
		jobs = append(jobs, *current)
	}

	return jobs
}


