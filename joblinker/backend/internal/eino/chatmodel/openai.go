package chatmodel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"joblinker/pkg/ai"
)

// EinoChatModel wraps the existing AI client to implement Eino's ChatModel interface
type EinoChatModel struct {
	client     *ai.Client
	modelName  string
	maxTokens  int
	temperature float64
}

// NewEinoChatModel creates a new Eino-compatible chat model wrapper
func NewEinoChatModel(client *ai.Client) *EinoChatModel {
	return &EinoChatModel{
		client:     client,
		modelName:  client.Model,
		maxTokens:  client.MaxTokens,
		temperature: client.Temperature,
	}
}

// Generate implements the BaseChatModel interface
func (m *EinoChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	log.Printf("Eino AI Generate called: model=%s, msg_count=%d", m.modelName, len(messages))
	// Apply options
	options := model.GetCommonOptions(&model.Options{}, opts...)
	if options.MaxTokens != nil {
		m.maxTokens = *options.MaxTokens
	}
	if options.Temperature != nil {
		m.temperature = float64(*options.Temperature)
	}
	if options.Model != nil {
		m.modelName = *options.Model
	}

	// Convert Eino messages to AI client messages
	aiMessages := convertToAIMessages(messages)

	// Build request
	reqBody := ai.ChatRequest{
		Model:       m.modelName,
		Messages:    aiMessages,
		MaxTokens:   m.maxTokens,
		Temperature: m.temperature,
	}

	// First call
	resp, err := m.client.DoChat(reqBody)
	if err != nil {
		return nil, fmt.Errorf("eino chatmodel: failed to call AI: %w", err)
	}

	choice := resp.Choices[0]

	// Handle tool calls
	if len(choice.Message.ToolCalls) > 0 {
		// Add assistant's tool calls to messages
		aiMessages = append(aiMessages, choice.Message)

		// We don't execute tools here - that would be handled by the agent
		// Just return the tool call request
		return &schema.Message{
			Role:       schema.Assistant,
			Content:    choice.Message.Content,
			ToolCalls:  convertToolCalls(choice.Message.ToolCalls),
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: choice.Message.Content,
	}, nil
}

// Stream implements the BaseChatModel interface (not implemented for MVP)
func (m *EinoChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// For MVP, fall back to non-streaming
	msg, err := m.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}

	// Wrap the single message in a stream
	sr, sw := schema.Pipe[*schema.Message](1)
	sw.Send(msg, nil)
	sw.Close()
	return sr, nil
}

// WithTools returns a new model instance with tools bound
func (m *EinoChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	// For MVP, just return self - tools are handled at a higher level
	return m, nil
}

// convertToAIMessages converts Eino schema messages to AI client messages
func convertToAIMessages(messages []*schema.Message) []ai.Message {
	result := make([]ai.Message, 0, len(messages))
	for _, msg := range messages {
		aiMsg := ai.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			aiMsg.ToolCalls = make([]ai.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				args, _ := json.Marshal(tc.Function.Arguments)
				aiMsg.ToolCalls[i] = ai.ToolCall{
					ID: tc.ID,
					Type: tc.Type,
					Function: struct {
						Name      string `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					}{
						Name:      tc.Function.Name,
						Arguments: args,
					},
				}
			}
		}

		result = append(result, aiMsg)
	}
	return result
}

// convertToolCalls converts AI client tool calls to Eino tool calls
func convertToolCalls(toolCalls []ai.ToolCall) []schema.ToolCall {
	result := make([]schema.ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		result[i] = schema.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: schema.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: string(tc.Function.Arguments),
			},
		}
	}
	return result
}

// Helper to convert role strings
func roleTypeToString(rt schema.RoleType) string {
	switch rt {
	case schema.System:
		return "system"
	case schema.User:
		return "user"
	case schema.Assistant:
		return "assistant"
	case schema.Tool:
		return "tool"
	default:
		return strings.ToLower(string(rt))
	}
}
