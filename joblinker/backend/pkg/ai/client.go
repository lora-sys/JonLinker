package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	Role     string     `json:"role"`
	Content  string     `json:"content"`
	Name     string     `json:"name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Tools       []Tool    `json:"tools,omitempty"`
}

type Tool struct {
	Type     string                 `json:"type"`
	Function ToolFunction            `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message       Message       `json:"message"`
	Finish string `json:"finish_reason"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

func NewClient() *Client {
	if os.Getenv("AI_MOCK_MODE") == "true" {
		log.Println("AI client initialized in MOCK MODE")
	}
	return &Client{
		BaseURL:    getEnv("AI_BASE_URL", "https://api.longcat.chat/openai"),
		APIKey:     getEnv("AI_API_KEY", ""),
		Model:      getEnv("AI_MODEL", "LongCat-Flash-Lite"),
		MaxTokens:  4096,
		Temperature: 0.2,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// ToolExecutor is a function type that executes a tool and returns its result
type ToolExecutor func(toolName string, arguments map[string]interface{}) (string, error)

// ChatWithTools enables function calling with the AI
// It sends tool definitions to the AI, handles tool calls, executes them, and returns final response
// In mock mode (AI_MOCK_MODE=true), returns canned FSM-progression responses
func (c *Client) ChatWithTools(systemPrompt, userPrompt string, tools []Tool, executor ToolExecutor) (string, string, error) {
	if os.Getenv("AI_MOCK_MODE") == "true" {
		return c.mockChat(systemPrompt, userPrompt)
	}
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	reqBody := ChatRequest{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   4096,
		Temperature: 0.2,
		Tools:       tools,
	}

	// First call - ask AI to use tools if needed
	resp, err := c.DoChat(reqBody)
	if err != nil {
		return "", "", err
	}

	// Check if AI wants to call a tool
	choice := resp.Choices[0]

	// Handle tool calls if present
	if len(choice.Message.ToolCalls) > 0 {
		// Add AI's response with tool calls to messages
		messages = append(messages, choice.Message)

		// Execute each tool call
		for _, tc := range choice.Message.ToolCalls {
			var args map[string]interface{}
			json.Unmarshal(tc.Function.Arguments, &args)

			result, err := executor(tc.Function.Name, args)
			if err != nil {
				result = fmt.Sprintf(`{"error": "%v"}`, err)
			}

			// Add tool result to messages
			messages = append(messages, Message{
				Role:    "tool",
				Content: result,
			})
		}

		// Second call - AI generates final response with tool results
		reqBody.Messages = messages
		reqBody.Tools = nil // No more tools needed
		resp, err = c.DoChat(reqBody)
		if err != nil {
			return "", "", err
		}
	}

	return resp.Choices[0].Message.Content, "", nil
}

// DoChat is the internal chat method for making API calls
func (c *Client) DoChat(reqBody ChatRequest) (*ChatResponse, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &chatResp, nil
}

// Chat is the basic chat method without function calling
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

	resp, err := c.DoChat(reqBody)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
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
		if currentOffer < minSalary {
			return -1, "Below minimum requirements.", nil
		}
		if currentOffer >= targetSalary {
			return targetSalary, "Offer accepted.", nil
		}
		counter := currentOffer + (targetSalary-currentOffer)/2
		return counter, "Counter offer proposed.", nil
	}

	return result.CounterOffer, result.Message, nil
}

// mockChat returns canned FSM-progression responses in mock mode
// The response intent progresses through: INQUIRY → INTRODUCTION → INTEREST → NEGOTIATION → OFFER → CONFIRM
var mockRound int

func (c *Client) mockChat(systemPrompt, userPrompt string) (string, string, error) {
	mockRound++

	// Parse job/candidate context from userPrompt to inject into responses
	jobTitle := extractField(userPrompt, "Position:")
	candidateSkills := extractField(userPrompt, "Skills:")
	jobLocation := extractField(userPrompt, "Location:")

	cycle := []struct {
		intent  string
		message string
	}{
		{
			intent: "INTRODUCTION",
			message: fmt.Sprintf(`{"intent":"INTRODUCTION","message":"Thank you for your interest in the %s role! We're looking for someone with strong %s background. The position is based in %s with a competitive compensation package. The team is working on exciting challenges in the distributed systems space.","data":{"title":"%s","location":"%s","salary_min":180000,"salary_max":250000,"skills":["Go","Python","Kubernetes","AWS","Microservices"]}}`,
				jobTitle, jobTitle, jobLocation, jobTitle, jobLocation),
		},
		{
			intent: "INTEREST",
			message: fmt.Sprintf(`{"intent":"INTEREST","message":"I'm excited about the %s opportunity! My %s directly maps to what you're looking for. Could you share more about the team structure, the tech stack you're using, and the biggest technical challenges the team is tackling right now?","data":{"message":"Interested in team structure and tech stack"}}`,
				jobTitle, candidateSkills),
		},
		{
			intent: "NEGOTIATION",
			message: fmt.Sprintf(`{"intent":"NEGOTIATION","message":"Let's discuss the compensation for the %s role. Given my %d years of experience with %s and the market rate for this level, I'd like to talk about the full package — base salary, equity, and benefits. Could you share details on the compensation structure?","data":{"message":"Compensation and benefits discussion"}}`,
				jobTitle, 8, candidateSkills),
		},
		{
			intent: "OFFER",
			message: fmt.Sprintf(`{"intent":"OFFER","message":"We'd like to extend a formal offer for the %s position. Based on your experience with %s, we're offering a competitive base salary of $220,000, significant equity package, and comprehensive benefits including health insurance, 401k matching, and unlimited PTO. We believe this reflects the value you'd bring to our engineering team.","data":{"salary":220000,"start_date":"2026-07-01","message":"Formal offer extended for %s role"}}`,
				jobTitle, candidateSkills, jobTitle),
		},
		{
			intent: "CONFIRM",
			message: fmt.Sprintf(`{"intent":"CONFIRM","message":"I've reviewed the offer for the %s role and I'm excited to accept. The compensation package is competitive and the opportunity to work on distributed systems with the latest technologies is exactly what I'm looking for. Looking forward to joining the team!","data":{"message":"Offer accepted for %s role"}}`,
				jobTitle, jobTitle),
		},
	}

	idx := (mockRound - 1) % len(cycle)
	log.Printf("[Mock AI] Round %d, returning intent=%s", mockRound, cycle[idx].intent)
	return cycle[idx].message, "", nil
}

// extractField extracts a field value from prompt text after "key: "
func extractField(prompt, key string) string {
	idx := strings.Index(prompt, key)
	if idx < 0 {
		return key
	}
	start := idx + len(key)
	// Trim leading whitespace
	for start < len(prompt) && prompt[start] == ' ' {
		start++
	}
	end := strings.Index(prompt[start:], "\n")
	if end < 0 {
		return strings.TrimSpace(prompt[start:])
	}
	val := strings.TrimSpace(prompt[start : start+end])
	// Remove trailing commas
	val = strings.TrimRight(val, ", ")
	// Bound at 80 chars to avoid overly long strings
	if len(val) > 80 {
		val = val[:80]
	}
	return val
}
