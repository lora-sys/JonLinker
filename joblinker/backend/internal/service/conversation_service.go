package service

import (
	"encoding/json"
	"fmt"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"
	"strings"

	"github.com/google/uuid"
)

const (
	// MaxTokens is the typical token limit for AI models (e.g., GPT-4 128k)
	MaxTokens = 100000
	// CompressionThreshold triggers compression at 80% of max tokens
	CompressionThreshold = 0.8
	// TargetTokensAfterCompression aims for 50% of max tokens after compression
	TargetTokensAfterCompression = 0.5
)

type ConversationService struct {
	messageRepo *repository.MessageRepository
	summaryRepo *repository.ConversationSummaryRepository
	aiClient    *ai.Client
}

func NewConversationService(
	messageRepo *repository.MessageRepository,
	summaryRepo *repository.ConversationSummaryRepository,
	aiClient *ai.Client,
) *ConversationService {
	return &ConversationService{
		messageRepo: messageRepo,
		summaryRepo: summaryRepo,
		aiClient:    aiClient,
	}
}

// GetConversationMessages returns all messages for a match, with token count
func (s *ConversationService) GetConversationMessages(matchID uuid.UUID) ([]*model.Message, int, error) {
	messages, err := s.messageRepo.ListByMatchID(matchID)
	if err != nil {
		return nil, 0, err
	}

	totalTokens := s.estimateTokenCount(messages)
	return messages, totalTokens, nil
}

// ShouldCompress returns true when token count exceeds 80% limit
func (s *ConversationService) ShouldCompress(matchID uuid.UUID) (bool, error) {
	_, totalTokens, err := s.GetConversationMessages(matchID)
	if err != nil {
		return false, err
	}
	return float64(totalTokens) > MaxTokens*CompressionThreshold, nil
}

// CompressConversation compresses conversation history while preserving key facts
func (s *ConversationService) CompressConversation(matchID uuid.UUID) error {
	messages, _, err := s.GetConversationMessages(matchID)
	if err != nil {
		return err
	}

	// Extract key facts from conversation
	keyFacts := s.extractKeyFacts(messages)

	// Generate summary text
	summaryText := s.generateSummaryText(messages, keyFacts)

	// Calculate token count after compression
	newTokenCount := s.estimateTokenCountFromText(summaryText)

	summary := &model.ConversationSummary{
		ID:           uuid.New(),
		MatchID:      matchID,
		SummaryText:  summaryText,
		KeyFacts:     keyFacts,
		TokenCount:   newTokenCount,
	}

	// Serialize key facts to JSON
	keyFactsJSON, _ := json.Marshal(keyFacts)
	summary.KeyFactsJSON = string(keyFactsJSON)

	return s.summaryRepo.Create(summary)
}

// Summarize uses AI to generate a summary of the conversation if available
func (s *ConversationService) Summarize(matchID uuid.UUID) (string, error) {
	messages, _, err := s.GetConversationMessages(matchID)
	if err != nil {
		return "", err
	}

	if s.aiClient == nil {
		return "", fmt.Errorf("AI client not configured")
	}

	// Build conversation text for AI
	var convText strings.Builder
	for _, m := range messages {
		convText.WriteString(fmt.Sprintf("[%s] %s\n", m.IntentType, m.ContentXML))
	}

	// Use AI to summarize
	summary, err := s.aiClient.GenerateAgentResponse(convText.String(), "conversation summarizer")
	if err != nil {
		return "", fmt.Errorf("AI summarization failed: %w", err)
	}

	return summary, nil
}

// extractKeyFacts extracts important information from conversation
func (s *ConversationService) extractKeyFacts(messages []*model.Message) []model.KeyFact {
	var keyFacts []model.KeyFact

	for _, msg := range messages {
		content := msg.ContentXML

		// Extract salary mentions
		if strings.Contains(strings.ToLower(content), "salary") ||
			strings.Contains(strings.ToLower(content), "compensation") ||
			strings.Contains(strings.ToLower(content), "$") {
			keyFacts = append(keyFacts, model.KeyFact{
				FactType: model.KeyFactSalary,
				Value:    extractSalaryValue(content),
				Priority: 10,
			})
		}

		// Extract location mentions
		if strings.Contains(strings.ToLower(content), "location") ||
			strings.Contains(strings.ToLower(content), "city") ||
			strings.Contains(strings.ToLower(content), "remote") ||
			strings.Contains(strings.ToLower(content), "onsite") {
			keyFacts = append(keyFacts, model.KeyFact{
				FactType: model.KeyFactLocation,
				Value:    extractLocationValue(content),
				Priority: 8,
			})
		}

		// Extract job type preferences
		if strings.Contains(strings.ToLower(content), "remote") ||
			strings.Contains(strings.ToLower(content), "hybrid") ||
			strings.Contains(strings.ToLower(content), "onsite") {
			keyFacts = append(keyFacts, model.KeyFact{
				FactType: model.KeyFactJobType,
				Value:    extractJobTypeValue(content),
				Priority: 7,
			})
		}

		// Extract skills
		if strings.Contains(strings.ToLower(content), "skill") ||
			strings.Contains(strings.ToLower(content), "experience") ||
			strings.Contains(strings.ToLower(content), "technolog") {
			keyFacts = append(keyFacts, model.KeyFact{
				FactType: model.KeyFactSkills,
				Value:    extractSkillsValue(content),
				Priority: 6,
			})
		}
	}

	return keyFacts
}

// generateSummaryText creates a compressed summary of the conversation
func (s *ConversationService) generateSummaryText(messages []*model.Message, keyFacts []model.KeyFact) string {
	if len(messages) == 0 {
		return "No conversation history."
	}

	// Build summary with key facts first
	var summary strings.Builder
	summary.WriteString("## Conversation Summary\n\n")

	// Add key facts section
	if len(keyFacts) > 0 {
		summary.WriteString("### Key Facts\n")
		for _, fact := range keyFacts {
			summary.WriteString(fmt.Sprintf("- %s: %s\n", fact.FactType, fact.Value))
		}
		summary.WriteString("\n")
	}

	// Add recent messages (last 3)
	recentStart := 0
	if len(messages) > 3 {
		recentStart = len(messages) - 3
	}
	summary.WriteString("### Recent Messages\n")
	for i := recentStart; i < len(messages); i++ {
		m := messages[i]
		summary.WriteString(fmt.Sprintf("- [%s] %s\n", m.IntentType, truncateContent(m.ContentXML, 100)))
	}

	return summary.String()
}

// estimateTokenCount estimates total tokens in messages
func (s *ConversationService) estimateTokenCount(messages []*model.Message) int {
	total := 0
	for _, msg := range messages {
		// Rough estimate: 1 token ≈ 4 chars for English
		total += len(msg.ContentXML) / 4
	}
	return total
}

// estimateTokenCountFromText estimates tokens in text
func (s *ConversationService) estimateTokenCountFromText(text string) int {
	return len(text) / 4
}

// GetSummary retrieves existing summary for a match
func (s *ConversationService) GetSummary(matchID uuid.UUID) (*model.ConversationSummary, error) {
	return s.summaryRepo.GetByMatchID(matchID)
}

// Helper functions for extraction

func extractSalaryValue(content string) string {
	// Look for salary patterns: $XXX,XXX or "salary: XXX" or numbers followed by "k" or "thousand"
	lower := strings.ToLower(content)

	// Pattern for $XXX,XXX or $XXXk
	if strings.Contains(lower, "$") {
		// Find dollar sign and extract number
		for i, c := range content {
			if c == '$' && i+1 < len(content) {
				var num []byte
				for j := i + 1; j < len(content) && (content[j] >= '0' && content[j] <= '9' || content[j] == ',' || content[j] == 'k' || content[j] == 'K'); j++ {
					if content[j] != ',' {
						num = append(num, content[j])
					}
				}
				if len(num) > 0 {
					return "$" + string(num)
				}
			}
		}
	}

	// Pattern for "salary" followed by number
	idx := strings.Index(lower, "salary")
	if idx >= 0 && idx+7 < len(content) {
		var num []byte
		for j := idx + 7; j < len(content) && (content[j] >= '0' && content[j] <= '9' || content[j] == ' ' || content[j] == ':'); j++ {
			if content[j] >= '0' && content[j] <= '9' {
				num = append(num, content[j])
			}
		}
		if len(num) >= 4 { // At least 4 digits (e.g., 100k)
			return string(num)
		}
	}

	if len(content) > 100 {
		return content[:100]
	}
	return content
}

var locationKeywords = []string{
	"remote", "onsite", "on-site", "hybrid",
	"san francisco", "new york", "nyc", "seattle", "austin", "boston", "chicago",
	"los angeles", "la", "denver", "portland", "san diego", "miami", "atlanta",
	"dallas", "houston", "phoenix", "philadelphia", "minneapolis", "detroit",
	"london", "berlin", "paris", "tokyo", "singapore", "toronto", "vancouver",
	"amsterdam", "dublin", "munich", "sydney", "melbangalore",
}

func extractLocationValue(content string) string {
	lower := strings.ToLower(content)
	for _, kw := range locationKeywords {
		if strings.Contains(lower, kw) {
			return kw
		}
	}
	if len(content) > 50 {
		return content[:50]
	}
	return content
}

func extractJobTypeValue(content string) string {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "remote") {
		return "remote"
	}
	if strings.Contains(lower, "hybrid") {
		return "hybrid"
	}
	if strings.Contains(lower, "onsite") {
		return "onsite"
	}
	return "unspecified"
}

var skillKeywords = []string{
	"Go", "Python", "Java", "JavaScript", "TypeScript", "C++", "C#", "Rust", "Ruby", "PHP",
	"Swift", "Kotlin", "Scala", "Elixir", "Haskell", "Clojure", "Perl", "R", "MATLAB",
	"React", "Vue", "Angular", "Svelte", "Next.js", "Nuxt", "Remix",
	"Node.js", "Express", "Django", "Flask", "FastAPI", "Spring", "Rails", "Gin", "Fiber",
	"Docker", "Kubernetes", "K8s", "Terraform", "Ansible", "Helm",
	"AWS", "GCP", "Azure", "Cloudflare", "Vercel", "Netlify",
	"PostgreSQL", "MySQL", "MongoDB", "Redis", "Elasticsearch", "DynamoDB", "Cassandra",
	"gRPC", "GraphQL", "REST", "WebSocket", "Kafka", "RabbitMQ", "NATS",
	"Git", "CI/CD", "Jenkins", "GitHub Actions", "GitLab CI",
	"Linux", "Bash", "Shell", "Nginx", "Apache",
	"Machine Learning", "ML", "AI", "Deep Learning", "NLP", "PyTorch", "TensorFlow",
	"Microservices", "Serverless", "FaaS", "API", "OAuth", "JWT",
}

func extractSkillsValue(content string) string {
	var found []string
	lower := strings.ToLower(content)
	seen := make(map[string]bool)
	for _, skill := range skillKeywords {
		if strings.Contains(lower, strings.ToLower(skill)) && !seen[strings.ToLower(skill)] {
			found = append(found, skill)
			seen[strings.ToLower(skill)] = true
		}
	}
	if len(found) > 0 {
		return strings.Join(found, ", ")
	}
	if len(content) > 100 {
		return content[:100]
	}
	return content
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}