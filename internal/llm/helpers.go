package llm

import (
	"context"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

func NewChatModel(ctx context.Context, baseURL, apiKey, model string, maxTokens int, temperature float32) (*openai.ChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       model,
		Timeout:     120 * time.Second,
		MaxTokens:   IntPtr(maxTokens),
		Temperature: Float32Ptr(temperature),
	})
}

func IntPtr(v int) *int { return &v }

func Float32Ptr(v float32) *float32 { return &v }
