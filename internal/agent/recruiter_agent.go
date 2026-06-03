package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/checkpoint"
	"github.com/lora-sys/JonLinker/internal/job"
	"github.com/lora-sys/JonLinker/internal/llm"
	"github.com/lora-sys/JonLinker/internal/memory"
	"github.com/lora-sys/JonLinker/internal/resume"
	"github.com/lora-sys/JonLinker/internal/session"
	"github.com/lora-sys/JonLinker/internal/tools/getcandidateprofile"
	"github.com/lora-sys/JonLinker/internal/tools/recordinterviewnote"
)

const recruiterSystemPrompt = `你是一个招聘官 Agent，正在面试一位求职者。

你的职责：
1. 读取候选人的 CandidateProfile 和已申请的职位 Application
2. 针对该职位的需求，向候选人提出面试问题
3. 评估候选人的回答质量
4. 在多轮面试结束后，调用 record_interview_note 工具保存面试记录

行为准则：
- 每次只问 1-2 个问题，等待候选人回答后再继续
- 基于候选人的简历提出有针对性的问题
- 如果候选人问题偏离主题，礼貌拉回
- 不替候选人做决定，不加个人评价
- 所有面试记录结构化保存
- 如果候选人想结束面试，调用 record_interview_note 保存记录

可用工具：
- get_candidate_profile：获取候选人资料和指定职位的申请材料（需要 job_url 参数）
- record_interview_note：保存面试记录（需要 job_url, questions_asked, candidate_answers, overall_assessment）`

const recruiterMemKeySuffix = ":recruiter"

type RecruiterAgent struct {
	inner      *react.Agent
	store      memory.MemoryStore
	cpStore    *checkpoint.Store
	maxHistory int
	locks      *session.Registry
}

func NewRecruiterAgent(ctx context.Context, baseURL, apiKey, model string, cpStore *checkpoint.Store, memStore memory.MemoryStore) (*RecruiterAgent, error) {
	chatModel, err := llm.NewChatModel(ctx, baseURL, apiKey, model, 4096, 0.1)
	if err != nil {
		return nil, err
	}

	profileTool := getcandidateprofile.NewTool(cpStore)
	noteTool := recordinterviewnote.NewTool(cpStore)

	modifier := func(ctx context.Context, msgs []*schema.Message) []*schema.Message {
		var sb strings.Builder
		sb.WriteString(recruiterSystemPrompt)

		sid := session.SessionIDFromContext(ctx)
		if sid != "" {
			profileData, ok, err := cpStore.Get(ctx, sid+":profile")
			if err == nil && ok && len(profileData) > 0 {
				var profile resume.CandidateProfile
				if err := json.Unmarshal(profileData, &profile); err == nil && profile.Name != "" {
					sb.WriteString(fmt.Sprintf("\n\n候选人资料：\n- 姓名: %s\n", profile.Name))
					if profile.Title != "" {
						sb.WriteString(fmt.Sprintf("- 求职意向: %s\n", profile.Title))
					}
					if len(profile.Skills) > 0 {
						sb.WriteString(fmt.Sprintf("- 技能: %s\n", strings.Join(profile.Skills, ", ")))
					}
					if len(profile.Experience) > 0 {
						sb.WriteString("- 工作经历:\n")
						for _, exp := range profile.Experience {
							sb.WriteString(fmt.Sprintf("  - %s at %s (%s)\n", exp.Title, exp.Company, exp.Duration))
						}
					}
					if len(profile.Education) > 0 {
						sb.WriteString("- 教育背景:\n")
						for _, edu := range profile.Education {
							sb.WriteString(fmt.Sprintf("  - %s %s (%s)\n", edu.School, edu.Major, edu.Duration))
						}
					}
					if profile.Summary != "" {
						sb.WriteString(fmt.Sprintf("- 简介: %s\n", profile.Summary))
					}
				}
			}

			appKeys := cpStore.Keys(ctx, sid+":application_")
			if len(appKeys) > 0 {
				sb.WriteString("\n候选人已申请的职位:\n")
				for i, key := range appKeys {
					data, ok, err := cpStore.Get(ctx, key)
					if err != nil || !ok {
						continue
					}
					var app job.Application
					if err := json.Unmarshal(data, &app); err != nil {
						continue
					}
					sb.WriteString(fmt.Sprintf("%d. %s at %s\n", i+1, app.JobTitle, app.Company))
				}
				sb.WriteString("\n请先询问候选人想先讨论哪个职位，然后调用 get_candidate_profile 工具获取该职位的申请材料。")
			}
		}

		sys := &schema.Message{
			Role:    schema.System,
			Content: sb.String(),
		}
		return append([]*schema.Message{sys}, msgs...)
	}

	inner, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{profileTool, noteTool},
		},
		MessageModifier: modifier,
		MaxStep:         15,
	})
	if err != nil {
		return nil, err
	}

	return &RecruiterAgent{
		inner:      inner,
		store:      memStore,
		cpStore:    cpStore,
		maxHistory: defaultMaxHistory,
		locks:      session.NewRegistry(),
	}, nil
}

func (a *RecruiterAgent) ChatStream(ctx context.Context, query, sessionID string, onToken func(string)) (*schema.Message, error) {
	ctx = session.WithSessionID(ctx, sessionID)

	mu := a.locks.Get(sessionID)
	mu.Lock()
	defer mu.Unlock()

	memKey := sessionID + recruiterMemKeySuffix

	history, err := a.store.Read(ctx, memKey)
	if err != nil {
		return nil, fmt.Errorf("read memory: %w", err)
	}

	history = append(history, &schema.Message{Role: schema.User, Content: query})
	if err := a.store.Write(ctx, memKey, history); err != nil {
		return nil, fmt.Errorf("write memory: %w", err)
	}

	stream, err := a.inner.Stream(ctx, history)
	if err != nil {
		return nil, err
	}

	var fullContent strings.Builder
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		fullContent.WriteString(chunk.Content)
		if onToken != nil {
			onToken(chunk.Content)
		}
	}

	msg := &schema.Message{Role: schema.Assistant, Content: fullContent.String()}

	history = append(history, msg)
	history = a.window(history)
	if err := a.store.Write(ctx, memKey, history); err != nil {
		return nil, fmt.Errorf("write memory: %w", err)
	}

	return msg, nil
}

func (a *RecruiterAgent) window(msgs []*schema.Message) []*schema.Message {
	n := a.maxHistory
	if n <= 0 || len(msgs) <= n {
		return msgs
	}
	return msgs[len(msgs)-n:]
}
