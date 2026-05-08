package agent

import (
	"context"
	"os"
	"testing"

	"github.com/cloudwego/eino/schema"

	"joblinker/pkg/ai"
)

func TestSeekerAgentChat(t *testing.T) {
	// Skip if AI_API_KEY is not set
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		t.Fatal("AI_API_KEY not set — required for integration test")
	}

	// Create AI client
	aiClient := &ai.Client{
		BaseURL:    "https://api.longcat.chat/openai",
		APIKey:     apiKey,
		Model:      "LongCat-Flash-Lite",
		MaxTokens:  1024,
		Temperature: 0.2,
	}

	// Create seeker agent
	agent := NewSeekerAgent(aiClient)

	// Test basic chat
	ctx := context.Background()
	resp, err := agent.Chat(ctx, "Hello, I'm looking for a software engineering job.")
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp == "" {
		t.Error("Empty response from agent")
	}

	t.Logf("Agent response: %s", resp)
}

func TestSeekerAgentChatWithHistory(t *testing.T) {
	// Skip if AI_API_KEY is not set
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		t.Fatal("AI_API_KEY not set — required for integration test")
	}

	// Create AI client
	aiClient := &ai.Client{
		BaseURL:    "https://api.longcat.chat/openai",
		APIKey:     apiKey,
		Model:      "LongCat-Flash-Lite",
		MaxTokens:  1024,
		Temperature: 0.2,
	}

	// Create seeker agent
	agent := NewSeekerAgent(aiClient)

	// Test chat with history
	ctx := context.Background()
	history := []*schema.Message{
		schema.AssistantMessage("Hello! I see you're interested in a software engineering position. What skills are you most proficient in?", nil),
	}

	resp, err := agent.ChatWithHistory(ctx, history, "I'm proficient in Go and Python.")
	if err != nil {
		t.Fatalf("ChatWithHistory failed: %v", err)
	}

	if resp == "" {
		t.Error("Empty response from agent")
	}

	t.Logf("Agent response: %s", resp)
}
