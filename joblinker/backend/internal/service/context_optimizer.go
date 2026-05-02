package service

import (
	"fmt"
	"strings"

	"joblinker/internal/cache"
	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

type ContextOptimizerService struct {
	optimizer  *cache.ContextOptimizer
	toolCache  *cache.ToolCache
	messageRepo *repository.MessageRepository
	matchRepo   *repository.MatchRepository
	agentRepo   *repository.AgentRepository
}

func NewContextOptimizerService(
	toolCache *cache.ToolCache,
	messageRepo *repository.MessageRepository,
	matchRepo *repository.MatchRepository,
	agentRepo *repository.AgentRepository,
) *ContextOptimizerService {
	optimizer := cache.NewContextOptimizer(10, 4000, toolCache)
	return &ContextOptimizerService{
		optimizer:   optimizer,
		toolCache:   toolCache,
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
		agentRepo:   agentRepo,
	}
}

func (s *ContextOptimizerService) AddMessage(matchID uuid.UUID, role, content string, cacheKey string) {
	s.optimizer.AddMessage(role, content, cacheKey)
}

func (s *ContextOptimizerService) GetRecentMessages(count int) []*cache.ConversationEntry {
	return s.optimizer.GetRecentMessages(count)
}

func (s *ContextOptimizerService) ShouldCompress() bool {
	return s.optimizer.ShouldCompress()
}

func (s *ContextOptimizerService) Compress() string {
	return s.optimizer.Compress()
}

func (s *ContextOptimizerService) BuildOptimizedContext() string {
	return s.optimizer.BuildOptimizedContext()
}

func (s *ContextOptimizerService) GetContextForMatch(matchID uuid.UUID) string {
	messages, _ := s.messageRepo.ListByMatchID(matchID)
	if len(messages) == 0 {
		return ""
	}

	entries := make([]*cache.ConversationEntry, 0, len(messages))
	for _, msg := range messages {
		entry := &cache.ConversationEntry{
			Role:      "agent",
			Content:   msg.ContentXML,
			Timestamp: msg.CreatedAt,
		}
		entries = append(entries, entry)
	}

	return cache.BuildContextForMatch(matchID, entries, s.toolCache)
}

func (s *ContextOptimizerService) CacheToolResult(toolName string, args map[string]interface{}, result map[string]interface{}) *cache.ToolResult {
	key := s.toolCache.GenerateCacheKey(toolName, args)
	return s.toolCache.Set(key, result)
}

func (s *ContextOptimizerService) GetCachedResult(key string) (map[string]interface{}, bool) {
	return s.toolCache.OnDemandLoad(key)
}

func (s *ContextOptimizerService) GetCacheSummary(key string) (string, bool) {
	return s.toolCache.GetSummary(key)
}

func (s *ContextOptimizerService) Stats() map[string]interface{} {
	cacheStats := s.toolCache.Stats()
	optimizerStats := s.optimizer.Stats()

	return map[string]interface{}{
		"cache":    cacheStats,
		"context":  optimizerStats,
	}
}

func (s *ContextOptimizerService) Reset() {
	s.optimizer.Reset()
}

func BuildAgentContext(agentType model.AgentType, scenario model.PromptScenarioType, matchID uuid.UUID, optimizer *ContextOptimizerService) string {
	var context strings.Builder

	context.WriteString(fmt.Sprintf("## Agent Type: %s\n", agentType))
	context.WriteString(fmt.Sprintf("## Scenario: %s\n", scenario))
	context.WriteString(fmt.Sprintf("## Match ID: %s\n\n", matchID.String()))

	if optimizer != nil {
		recentMsgs := optimizer.GetRecentMessages(5)
		if len(recentMsgs) > 0 {
			context.WriteString("## Recent Conversation\n")
			for i, msg := range recentMsgs {
				if msg.CacheKey != "" {
					if summary, ok := optimizer.GetCacheSummary(msg.CacheKey); ok {
						context.WriteString(fmt.Sprintf("%d. [%s] (cached: %s)\n", i+1, msg.Role, summary))
						continue
					}
				}
				context.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, msg.Role, msg.Content))
			}
		}

		if optimizer.ShouldCompress() {
			compressed := optimizer.Compress()
			context.WriteString("\n## Compressed Context\n")
			context.WriteString(compressed)
		}
	}

	return context.String()
}