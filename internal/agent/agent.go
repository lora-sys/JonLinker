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

	"github.com/lora-sys/JonLinker/internal/llm"
	"github.com/lora-sys/JonLinker/internal/memory"
	"github.com/lora-sys/JonLinker/internal/resume"
	"github.com/lora-sys/JonLinker/internal/session"
	"github.com/lora-sys/JonLinker/internal/tools/applyjob"
	"github.com/lora-sys/JonLinker/internal/tools/queryjobs"
)

const systemPrompt = `你是AI招聘助手，帮助候选人搜索职位和生成求职申请。

你可以与用户自然对话，了解他们的求职需求。
打招呼、闲聊或询问能力时，直接用自然语言回复，不要调用任何工具。
当你需要搜索职位时，调用 query_jobs 工具展示结果给用户。
当用户确认要申请某个职位时，调用 apply_job 工具生成定制求职信和简历。

注意：
- 使用工具后，用自然语言向用户展示结果
- 当搜索到职位时，在自然语言回复后附加JSON格式结果：
  ===JSON===
  {"jobs":[{"title":"...","company":"...","location":"...","salary":"...","url":"...","description":"...","tags":[],"source":"Indeed","match_score":85,"summary":"推荐理由","highlights":["React ✓"]}]}
  ===END===
- 搜索无结果时附加：===JSON==={"jobs":[]}===END===
- 申请结果由 apply_job 工具自动处理，无需手动添加JSON

当前会话ID: {{SESSION_ID}}

可用工具：
- query_jobs：搜索职位（支持关键词、城市筛选）
- apply_job：生成求职申请（需要 job_url，无需手动传入 session_id）`

const defaultMaxHistory = 20

type Agent struct {
	inner      *react.Agent
	store      memory.MemoryStore
	cpStore    compose.CheckPointStore
	maxHistory int
	locks      *session.Registry
}

func New(ctx context.Context, baseURL, apiKey, model, fcKey string, dApply *applyjob.DirectApply, store memory.MemoryStore, cpStore compose.CheckPointStore) (*Agent, error) {
	chatModel, err := llm.NewChatModel(ctx, baseURL, apiKey, model, 4096, 0.1)
	if err != nil {
		return nil, err
	}

	qTool := queryjobs.NewTool(fcKey)
	aTool := applyjob.NewSearchTool(dApply)

	modifier := func(ctx context.Context, msgs []*schema.Message) []*schema.Message {
		sid := session.SessionIDFromContext(ctx)
		prompt := systemPrompt
		if sid != "" {
			prompt = strings.ReplaceAll(prompt, "{{SESSION_ID}}", sid)
			if cpStore != nil {
				data, ok, err := cpStore.Get(ctx, sid+":profile")
				if err == nil && ok && len(data) > 0 {
					var profile resume.CandidateProfile
					if err := json.Unmarshal(data, &profile); err == nil && profile.Name != "" {
						var summary strings.Builder
						summary.WriteString(fmt.Sprintf("\n\n当前候选人资料：\n- 姓名: %s\n", profile.Name))
						if profile.Title != "" {
							summary.WriteString(fmt.Sprintf("- 求职意向: %s\n", profile.Title))
						}
						if len(profile.Skills) > 0 {
							summary.WriteString(fmt.Sprintf("- 技能: %s\n", strings.Join(profile.Skills, ", ")))
						}
						if profile.Summary != "" {
							summary.WriteString(fmt.Sprintf("- 简介: %s\n", profile.Summary))
						}
						summary.WriteString("\n搜索职位时应优先匹配以上技能组合。")
						prompt += summary.String()
					}
				}
			}
		}
		sys := &schema.Message{
			Role:    schema.System,
			Content: prompt,
		}
		return append([]*schema.Message{sys}, msgs...)
	}

	inner, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{qTool, aTool},
		},
		MessageModifier: modifier,
		MaxStep:         15,
	})
	if err != nil {
		return nil, err
	}

	return &Agent{
		inner:      inner,
		store:      store,
		cpStore:    cpStore,
		maxHistory: defaultMaxHistory,
		locks:      session.NewRegistry(),
	}, nil
}

func (a *Agent) Search(ctx context.Context, query, sessionID string) (*schema.Message, error) {
	ctx = session.WithSessionID(ctx, sessionID)

	mu := a.locks.Get(sessionID)
	mu.Lock()
	defer mu.Unlock()

	history, err := a.store.Read(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("read memory: %w", err)
	}

	history = append(history, &schema.Message{Role: schema.User, Content: query})
	if err := a.store.Write(ctx, sessionID, history); err != nil {
		return nil, fmt.Errorf("write memory: %w", err)
	}

	msg, err := a.inner.Generate(ctx, history)
	if err != nil {
		return nil, err
	}

	history = append(history, msg)
	history = a.window(history)
	if err := a.store.Write(ctx, sessionID, history); err != nil {
		return nil, fmt.Errorf("write memory: %w", err)
	}

	return msg, nil
}

func (a *Agent) SearchStream(ctx context.Context, query, sessionID string, onToken func(string)) (*schema.Message, error) {
	ctx = session.WithSessionID(ctx, sessionID)

	mu := a.locks.Get(sessionID)
	mu.Lock()
	defer mu.Unlock()

	history, err := a.store.Read(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("read memory: %w", err)
	}

	history = append(history, &schema.Message{Role: schema.User, Content: query})
	if err := a.store.Write(ctx, sessionID, history); err != nil {
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
	if err := a.store.Write(ctx, sessionID, history); err != nil {
		return nil, fmt.Errorf("write memory: %w", err)
	}

	return msg, nil
}

func (a *Agent) window(msgs []*schema.Message) []*schema.Message {
	n := a.maxHistory
	if n <= 0 || len(msgs) <= n {
		return msgs
	}
	return msgs[len(msgs)-n:]
}


