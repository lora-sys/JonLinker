package applyjob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type SearchTool struct {
	dApply *DirectApply
}

func NewSearchTool(dApply *DirectApply) *SearchTool {
	return &SearchTool{dApply: dApply}
}

func (t *SearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "apply_job",
		Desc: "为指定职位生成求职申请（求职信 + 定制简历 + 匹配亮点）。需要先搜索到职位再调用。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"job_url": {
				Type:     schema.String,
				Desc:     "职位链接",
				Required: true,
			},
			"session_id": {
				Type:     schema.String,
				Desc:     "会话ID，从系统提示中获取",
				Required: true,
			},
		}),
	}, nil
}

type searchToolArgs struct {
	JobURL    string `json:"job_url"`
	SessionID string `json:"session_id"`
}

func (t *SearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var args searchToolArgs
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	app, err := t.dApply.Generate(ctx, args.JobURL, args.SessionID)
	if err != nil {
		return fmt.Sprintf("申请生成失败: %v", err), nil
	}

	data, _ := json.Marshal(app)
	return string(data), nil
}
