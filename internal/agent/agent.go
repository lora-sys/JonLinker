package agent

import (
	"context"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/tools/query_jobs"
)

const systemPrompt = `你是AI招聘助手。工作流程：
1. 分析用户需求，提取关键词、城市等
2. 调用 query_jobs 工具搜索职位
3. 对结果排序，给出匹配度评分(0-100)和推荐理由

重要：最终回复必须是纯 JSON，不要包含任何其他文字或标记。格式：
{
  "jobs": [{
    "title": "职位名称",
    "company": "公司",
    "location": "地点",
    "salary": "薪资",
    "url": "链接",
    "description": "描述",
    "tags": [],
    "source": "Indeed",
    "match_score": 85,
    "summary": "一句话推荐理由",
    "highlights": ["React ✓"]
  }],
  "intent": {"keyword": "", "city": "", "salary_min": 0, "experience": ""}
}
搜索无结果时返回 {"jobs":[],"intent":{}}`

type Agent struct {
	inner *react.Agent
}

func New(ctx context.Context, baseURL, apiKey, model, fcKey string) (*Agent, error) {
	timeout := 120 * time.Second

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Model:      model,
		Timeout:    timeout,
		MaxTokens:  intPtr(4096),
		Temperature: float32Ptr(0.1),
	})
	if err != nil {
		return nil, err
	}

	qTool := query_jobs.NewTool(fcKey)

	modifier := func(ctx context.Context, msgs []*schema.Message) []*schema.Message {
		sys := &schema.Message{
			Role:    schema.System,
			Content: systemPrompt,
		}
		return append([]*schema.Message{sys}, msgs...)
	}

	inner, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{qTool},
		},
		MessageModifier: modifier,
		MaxStep:         8,
	})
	if err != nil {
		return nil, err
	}

	return &Agent{inner: inner}, nil
}

func (a *Agent) Search(ctx context.Context, query string) (*schema.Message, error) {
	return a.inner.Generate(ctx, []*schema.Message{
		{Role: schema.User, Content: query},
	})
}

func intPtr(v int) *int { return &v }
func float32Ptr(v float32) *float32 { return &v }
