package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	MaxTokens  int
	Temperature float64
	HTTPClient *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message Message `json:"message"`
	Finish string `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

func NewClient() *Client {
	return &Client{
		BaseURL:    getEnv("AI_BASE_URL", "https://api.longcat.chat/openai"),
		APIKey:     getEnv("AI_API_KEY", ""),
		Model:      getEnv("AI_MODEL", "LongCat-Flash-Lite"),
		MaxTokens:  4096,
		Temperature: 0.2,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (c *Client) Chat(systemPrompt, userPrompt string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	reqBody := ChatRequest{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   c.MaxTokens,
		Temperature: c.Temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GenerateAgentResponse uses AI to generate an agent response based on context
func (c *Client) GenerateAgentResponse(context string, agentType string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are an AI agent representing a %s in a recruitment platform.
Your role is to engage in professional dialogue about job opportunities.
Respond concisely and professionally. Format responses as plain text.
Never reveal sensitive personal information.`, agentType)

	userPrompt := fmt.Sprintf(`Context: %s

Based on the context above, generate an appropriate professional response for the recruitment dialogue.
Keep responses brief (1-3 sentences) and focused on the recruitment topic.`, context)

	return c.Chat(systemPrompt, userPrompt)
}

// EvaluateMatch uses AI to evaluate a job match
func (c *Client) EvaluateMatch(seekerProfile, jobDescription string) (float64, string, error) {
	systemPrompt := `You are an AI recruitment assistant. Evaluate job matches objectively.
Return a JSON object with:
- score: float 0.0-1.0 indicating match quality
- reasoning: brief explanation of the evaluation

Example: {"score": 0.85, "reasoning": "Strong skills alignment with 2/3 required skills matched"}`

	userPrompt := fmt.Sprintf(`Seeker Profile: %s

Job Description: %s

Evaluate how well the seeker matches this job. Return JSON only.`, seekerProfile, jobDescription)

	response, err := c.Chat(systemPrompt, userPrompt)
	if err != nil {
		return 0, "", err
	}

	// Try to parse JSON response
	var result struct {
		Score     float64 `json:"score"`
		Reasoning string  `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If parsing fails, return a default
		return 0.5, "AI evaluation unavailable", nil
	}

	return result.Score, result.Reasoning, nil
}

// NegotiateSalary uses AI to determine salary negotiation strategy
func (c *Client) NegotiateSalary(currentOffer, targetSalary, minSalary int, agentType string) (int, string, error) {
	systemPrompt := fmt.Sprintf(`You are an AI agent negotiating salary on behalf of a %s.
Be professional but firm. Return a JSON object with:
- counter_offer: the salary figure you'd counter with (or -1 to walk away)
- message: brief professional message explaining the counter

Example: {"counter_offer": 95000, "message": "Based on my experience, I'd need at least $95K to consider this offer."}`, agentType)

	userPrompt := fmt.Sprintf(`Current offer: $%d
Target salary: $%d
Minimum acceptable: $%d

Generate a professional counter-offer or decision to accept/walk away. Return JSON only.`, currentOffer, targetSalary, minSalary)

	response, err := c.Chat(systemPrompt, userPrompt)
	if err != nil {
		return 0, "", err
	}

	var result struct {
		CounterOffer int    `json:"counter_offer"`
		Message     string `json:"message"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback to rule-based negotiation
		neg := NewSalaryNegotiatorSimple(minSalary, targetSalary, currentOffer)
		return neg.GenerateCounter()
	}

	return result.CounterOffer, result.Message, nil
}

// SalaryNegotiatorSimple is a simple rule-based negotiator
type SalaryNegotiatorSimple struct {
	MinSalary    int
	TargetSalary int
	CurrentOffer int
}

func NewSalaryNegotiatorSimple(min, target, current int) *SalaryNegotiatorSimple {
	return &SalaryNegotiatorSimple{
		MinSalary:    min,
		TargetSalary: target,
		CurrentOffer: current,
	}
}

func (n *SalaryNegotiatorSimple) GenerateCounter() (int, string, error) {
	if n.CurrentOffer < n.MinSalary {
		return -1, "This offer is below my minimum requirements.", nil
	}
	if n.CurrentOffer >= n.TargetSalary {
		return n.TargetSalary, "I can accept this offer.", nil
	}
	// Generate counter between current and target
	counter := n.CurrentOffer + (n.TargetSalary-n.CurrentOffer)/2
	if counter < n.MinSalary {
		counter = n.MinSalary
	}
	return counter, fmt.Sprintf("I'd need at least $%d to proceed.", counter), nil
}
