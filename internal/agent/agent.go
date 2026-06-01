package agent

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/tools/applyjob"
	"github.com/lora-sys/JonLinker/internal/tools/queryjobs"
)

type sessionIDKey struct{}

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
- 生成申请后附加JSON：
  ===JSON===
  {"application":{"job_title":"...","company":"...","cover_letter":"...","resume_md":"...","highlights":["..."],"generated_at":"..."}}
  ===END===
- 搜索无结果时附加：===JSON==={"jobs":[]}===END===

当前会话ID: {{SESSION_ID}}

可用工具：
- query_jobs：搜索职位（支持关键词、城市筛选）
- apply_job：生成求职申请（需要 job_url 和 session_id，session_id 使用当前会话ID）`

type Agent struct {
	inner        *react.Agent
	mu           sync.Mutex
	sessions     map[string][]*schema.Message
	sessionLocks sync.Map
}

func (a *Agent) getSessionLock(sessionID string) *sync.Mutex {
	v, _ := a.sessionLocks.LoadOrStore(sessionID, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func New(ctx context.Context, baseURL, apiKey, model, fcKey string, dApply *applyjob.DirectApply) (*Agent, error) {
	timeout := 120 * time.Second

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       model,
		Timeout:     timeout,
		MaxTokens:   intPtr(4096),
		Temperature: float32Ptr(0.1),
	})
	if err != nil {
		return nil, err
	}

	qTool := queryjobs.NewTool(fcKey)
	aTool := applyjob.NewSearchTool(dApply)

	modifier := func(ctx context.Context, msgs []*schema.Message) []*schema.Message {
		sid, _ := ctx.Value(sessionIDKey{}).(string)
		prompt := systemPrompt
		if sid != "" {
			prompt = strings.ReplaceAll(prompt, "{{SESSION_ID}}", sid)
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
		inner:    inner,
		sessions: make(map[string][]*schema.Message),
	}, nil
}

func (a *Agent) Search(ctx context.Context, query, sessionID string) (*schema.Message, error) {
	ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)

	slock := a.getSessionLock(sessionID)
	slock.Lock()
	a.mu.Lock()
	history := a.sessions[sessionID]
	history = append(history, &schema.Message{Role: schema.User, Content: query})
	a.sessions[sessionID] = history
	a.mu.Unlock()
	slock.Unlock()

	msg, err := a.inner.Generate(ctx, history)
	if err != nil {
		return nil, err
	}

	slock.Lock()
	a.mu.Lock()
	history = a.sessions[sessionID]
	history = append(history, msg)
	a.sessions[sessionID] = history
	a.mu.Unlock()
	slock.Unlock()

	return msg, nil
}

func (a *Agent) SearchStream(ctx context.Context, query, sessionID string, onThinking func(string)) (*schema.Message, error) {
	ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)

	slock := a.getSessionLock(sessionID)
	slock.Lock()
	a.mu.Lock()
	history := a.sessions[sessionID]
	history = append(history, &schema.Message{Role: schema.User, Content: query})
	a.sessions[sessionID] = history
	a.mu.Unlock()
	slock.Unlock()

	if onThinking != nil {
		onThinking("正在搜索职位...")
	}

	msg, err := a.inner.Generate(ctx, history)
	if err != nil {
		return nil, err
	}

	if onThinking != nil {
		onThinking("分析完成")
	}

	slock.Lock()
	a.mu.Lock()
	history = a.sessions[sessionID]
	history = append(history, msg)
	a.sessions[sessionID] = history
	a.mu.Unlock()
	slock.Unlock()

	return msg, nil
}

func intPtr(v int) *int { return &v }
func float32Ptr(v float32) *float32 { return &v }
