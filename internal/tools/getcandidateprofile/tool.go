package getcandidateprofile

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/resume"
	"github.com/lora-sys/JonLinker/internal/session"
)

type Tool struct {
	cpStore interface {
		Get(ctx context.Context, key string) ([]byte, bool, error)
	}
}

func NewTool(store interface{ Get(ctx context.Context, key string) ([]byte, bool, error) }) *Tool {
	return &Tool{cpStore: store}
}

func (t *Tool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_candidate_profile",
		Desc: "获取候选人的个人资料和指定职位的申请材料。需要提供职位URL作为参数。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"job_url": {
				Type:     schema.String,
				Desc:     "职位URL",
				Required: true,
			},
		}),
	}, nil
}

type getCandidateProfileArgs struct {
	JobURL string `json:"job_url"`
}

func (t *Tool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args getCandidateProfileArgs
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	sessionID := session.SessionIDFromContext(ctx)
	if sessionID == "" {
		return "session not found", nil
	}

	result := make(map[string]any)

	profileData, ok, err := t.cpStore.Get(ctx, sessionID+":profile")
	if err != nil {
		return "", fmt.Errorf("get profile: %w", err)
	}
	if !ok {
		return fmt.Sprintf("profile not found for session %s", sessionID), nil
	}
	var profile resume.CandidateProfile
	if err := json.Unmarshal(profileData, &profile); err != nil {
		return "", fmt.Errorf("unmarshal profile: %w", err)
	}
	result["profile"] = profile

	appData, ok, err := t.cpStore.Get(ctx, sessionID+":application_"+args.JobURL)
	if err != nil {
		return "", fmt.Errorf("get application: %w", err)
	}
	if !ok {
		return fmt.Sprintf("application not found for job %s", args.JobURL), nil
	}
	var application job.Application
	if err := json.Unmarshal(appData, &application); err != nil {
		return "", fmt.Errorf("unmarshal application: %w", err)
	}
	result["application"] = application

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}
	return string(data), nil
}
