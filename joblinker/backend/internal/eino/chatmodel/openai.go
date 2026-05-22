package chatmodel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"joblinker/pkg/ai"
)

var functionCallRe = regexp.MustCompile(`<function_call>\s*<function_name>([^<]+)</function_name>\s*<parameters>(.*?)</parameters>\s*</function_call>`)
var xmlParamRe = regexp.MustCompile(`<([^>/]+)>([^<]*)</[^>]+>`)

// EinoChatModel wraps the existing AI client to implement Eino's ChatModel interface
type EinoChatModel struct {
	client     *ai.Client
	modelName  string
	maxTokens  int
	temperature float64
	tools      []*schema.ToolInfo
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

	// Build request — LongCat API does not support native tools parameter,
	// tools are injected via system prompt XML format instead
	reqBody := ai.ChatRequest{
		Model:       m.modelName,
		Messages:    aiMessages,
		MaxTokens:   m.maxTokens,
		Temperature: m.temperature,
	}

	// Use streaming API internally to capture the complete model output,
	// including any <function_call> XML that may appear mid-stream.
	// The ADK Runner may use Stream() directly (which lacks XML parsing),
	// so Generate() must do its own streaming to see the full response.
	content, err := m.collectStreamResponse(ctx, reqBody)
	if err != nil {
		// Fallback to non-streaming
		log.Printf("Eino AI Generate: stream failed (%v), falling back to non-streaming", err)
		resp, err2 := m.client.DoChat(reqBody)
		if err2 != nil {
			return nil, fmt.Errorf("eino chatmodel: failed to call AI: %w", err2)
		}
		content = resp.Choices[0].Message.Content
	}
	log.Printf("Eino AI Generate response: content_len=%d, preview=%q", len(content), truncate(content, 200))

	// LongCat API does not return native ToolCalls; model outputs
	// <function_call> XML as text. Parse it into proper ToolCall objects.
	toolCalls := parseFunctionCallXML(content)
	if len(toolCalls) > 0 {
		cleanContent := stripFunctionCallXML(content)
		log.Printf("Eino AI Generate: parsed %d tool calls from XML, content_len=%d", len(toolCalls), len(cleanContent))
		return &schema.Message{
			Role:       schema.Assistant,
			Content:    cleanContent,
			ToolCalls:  toolCalls,
		}, nil
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
	}, nil
}

// collectStreamResponse calls the streaming API and buffers all chunks
// into a single string. This ensures Generate() sees the complete model
// output including any mid-stream <function_call> XML.
func (m *EinoChatModel) collectStreamResponse(ctx context.Context, reqBody ai.ChatRequest) (string, error) {
	ch, err := m.client.DoChatStream(ctx, reqBody)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	for chunk := range ch {
		buf.WriteString(chunk)
	}
	return buf.String(), nil
}

// parseFunctionCallXML extracts <function_call> XML blocks from text content
// and converts them to schema.ToolCall objects.
// Format: <function_call><function_name>NAME</function_name><parameters><key>val</key></parameters></function_call>
func parseFunctionCallXML(content string) []schema.ToolCall {
	matches := functionCallRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	toolCalls := make([]schema.ToolCall, 0, len(matches))
	for _, match := range matches {
		args := xmlParamsToJSON(match[2])
		toolCalls = append(toolCalls, schema.ToolCall{
			ID:   uuid.New().String(),
			Type: "function",
			Function: schema.FunctionCall{
				Name:      strings.TrimSpace(match[1]),
				Arguments: string(args),
			},
		})
	}
	return toolCalls
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// stripFunctionCallXML removes <function_call> XML blocks from text content.
func stripFunctionCallXML(content string) string {
	return strings.TrimSpace(functionCallRe.ReplaceAllString(content, ""))
}

// xmlParamsToJSON converts <key>value</key> XML to a JSON object.
func xmlParamsToJSON(paramsXML string) []byte {
	matches := xmlParamRe.FindAllStringSubmatch(paramsXML, -1)
	result := make(map[string]interface{}, len(matches))
	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		val := strings.TrimSpace(match[2])
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			result[key] = num
		} else if val == "true" {
			result[key] = true
		} else if val == "false" {
			result[key] = false
		} else {
			result[key] = val
		}
	}
	jsonBytes, _ := json.Marshal(result)
	return jsonBytes
}

// jsonArgsToXML converts a JSON object string back to <key>value</key> XML.
func jsonArgsToXML(argsJSON string) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil || len(args) == 0 {
		return ""
	}
	var sb strings.Builder
	for k, v := range args {
		sb.WriteString(fmt.Sprintf("<%s>%v</%s>", k, v, k))
	}
	return sb.String()
}

// Stream implements the BaseChatModel interface with real SSE streaming.
// Buffers all chunks, then checks for <function_call> XML. If found,
// strips the XML from text content and injects a final chunk with
// ToolCalls so the ReAct graph branches to tool execution.
func (m *EinoChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
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

	ch, err := m.client.DoChatStream(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("eino chatmodel: failed to start stream: %w", err)
	}

	// Buffer all chunks, then check for XML tool calls
	sr, sw := schema.Pipe[*schema.Message](64)
	go func() {
		defer sw.Close()
		var chunks []string
		for {
			select {
			case chunk, ok := <-ch:
				if !ok {
					// Stream ended — check accumulated content for tool calls
					fullContent := strings.Join(chunks, "")
					toolCalls := parseFunctionCallXML(fullContent)
					if len(toolCalls) > 0 {
						cleanText := stripFunctionCallXML(fullContent)
						if cleanText != "" {
							sw.Send(&schema.Message{
								Role:    schema.Assistant,
								Content: cleanText,
							}, nil)
						}
						sw.Send(&schema.Message{
							Role:      schema.Assistant,
							ToolCalls: toolCalls,
						}, nil)
						log.Printf("Eino AI Stream: parsed %d tool calls from XML (total=%d chunks)", len(toolCalls), len(chunks))
					} else {
						for _, c := range chunks {
							sw.Send(&schema.Message{
								Role:    schema.Assistant,
								Content: c,
							}, nil)
						}
					}
					return
				}
				chunks = append(chunks, chunk)
			case <-ctx.Done():
				return
			}
		}
	}()

	return sr, nil
}

// WithTools returns a new model instance with tools bound
func (m *EinoChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return &EinoChatModel{
		client:      m.client,
		modelName:   m.modelName,
		maxTokens:   m.maxTokens,
		temperature: m.temperature,
		tools:       tools,
	}, nil
}

// convertToAIMessages converts Eino schema messages to AI client messages.
// LongCat API does not support native tool_calls or tool role — convert them
// to XML in text content that the model understands.
func convertToAIMessages(messages []*schema.Message) []ai.Message {
	result := make([]ai.Message, 0, len(messages))
	for _, msg := range messages {
		// Tool role not supported by LongCat; format as user message
		if msg.Role == schema.Tool {
			toolName := msg.ToolName
			if toolName == "" {
				toolName = "function"
			}
			result = append(result, ai.Message{
				Role:    "user",
				Content: fmt.Sprintf("<tool_result>\n<tool_name>%s</tool_name>\n<output>\n%s\n</output>\n</tool_result>", toolName, msg.Content),
			})
			continue
		}

		aiMsg := ai.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Assistant messages with ToolCalls: reconstruct <function_call> XML
		// in content since LongCat ignores native tool_calls field
		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				paramsXML := jsonArgsToXML(tc.Function.Arguments)
				aiMsg.Content += fmt.Sprintf("\n<function_call><function_name>%s</function_name><parameters>%s</parameters></function_call>",
					tc.Function.Name, paramsXML)
			}
		}

		result = append(result, aiMsg)
	}
	return result
}


