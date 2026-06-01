package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/resume"
)

type ResumeAgent struct {
	chatModel *openai.ChatModel
	store     compose.CheckPointStore
	history   []*schema.Message
	mu        sync.Mutex
}

func NewResumeAgent(ctx context.Context, baseURL, apiKey, model, parsedText string, store compose.CheckPointStore) (*ResumeAgent, error) {
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       model,
		Timeout:     120 * time.Second,
		MaxTokens:   intPtr(4096),
		Temperature: float32Ptr(0.3),
	})
	if err != nil {
		return nil, err
	}

	sysMsg := fmt.Sprintf(`你是AI招聘助手，负责通过对话收集候选人的求职资料。

已有简历文本：
%s

请根据以上简历文本，与用户聊天补充完整的求职资料。
你需要提取或询问以下信息：
- 姓名 (name)
- 求职意向 (title)
- 技能列表 (skills)
- 工作经历 (experience)：公司、职位、时间、描述
- 教育背景 (education)：学校、学历、专业、时间
- 电话 (phone)
- 邮箱 (email)
- 个人简介 (summary)
- 兴趣爱好 (hobbies)

规则：
1. 从简历文本中提取已有信息，不要重复询问
2. 对缺失的信息，逐个询问用户
3. 每次问1-2个问题
4. 所有信息收集完毕，输出JSON：
{"complete":true,"profile":{...}}
未完成时输出：
{"complete":false,"message":"还需要了解..."}`, parsedText)

	return &ResumeAgent{
		chatModel: cm,
		store:     store,
		history:   []*schema.Message{{Role: schema.System, Content: sysMsg}},
	}, nil
}

func (a *ResumeAgent) ChatStream(ctx context.Context, sessionID, userMsg string, onToken func(string)) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.history = append(a.history, &schema.Message{Role: schema.User, Content: userMsg})

	sr, err := a.chatModel.Stream(ctx, a.history)
	if err != nil {
		return "", fmt.Errorf("stream: %w", err)
	}
	defer sr.Close()

	var full strings.Builder
	for {
		msg, err := sr.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("recv: %w", err)
		}
		full.WriteString(msg.Content)
		if onToken != nil {
			onToken(msg.Content)
		}
	}

	reply := full.String()
	a.history = append(a.history, &schema.Message{Role: schema.Assistant, Content: reply})

	jsonPart := reply
	if idx := strings.Index(jsonPart, "{"); idx >= 0 {
		jsonPart = jsonPart[idx:]
	}
	if idx := strings.LastIndex(jsonPart, "}"); idx >= 0 {
		jsonPart = jsonPart[:idx+1]
	}

	var state struct {
		Complete bool                    `json:"complete"`
		Message  string                  `json:"message,omitempty"`
		Profile  resume.CandidateProfile `json:"profile,omitempty"`
	}
	if err := json.Unmarshal([]byte(jsonPart), &state); err == nil && state.Complete && state.Profile.Name != "" {
		if b, err := json.Marshal(state.Profile); err == nil {
			if err := a.store.Set(ctx, sessionID+":profile", b); err != nil {
				log.Printf("checkpoint save failed: %v", err)
			}
		}
	}

	return reply, nil
}
