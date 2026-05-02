package cache

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ConversationEntry struct {
	Role      string
	Content   string
	Timestamp time.Time
	CacheKey  string
}

type ContextOptimizer struct {
	windowSize    int
	messages      []*ConversationEntry
	maxContextTokens int
	cache         *ToolCache
}

func NewContextOptimizer(windowSize int, maxContextTokens int, cache *ToolCache) *ContextOptimizer {
	return &ContextOptimizer{
		windowSize:      windowSize,
		messages:        make([]*ConversationEntry, 0),
		maxContextTokens: maxContextTokens,
		cache:           cache,
	}
}

func (co *ContextOptimizer) AddMessage(role, content string, cacheKey string) {
	entry := &ConversationEntry{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		CacheKey:  cacheKey,
	}
	co.messages = append(co.messages, entry)

	if len(co.messages) > co.windowSize {
		co.messages = co.messages[len(co.messages)-co.windowSize:]
	}
}

func (co *ContextOptimizer) GetRecentMessages(count int) []*ConversationEntry {
	if count > len(co.messages) {
		count = len(co.messages)
	}
	start := len(co.messages) - count
	return co.messages[start:]
}

func (co *ContextOptimizer) ShouldCompress() bool {
	estimatedTokens := co.estimateContextTokens()
	return estimatedTokens > (co.maxContextTokens * 50 / 100)
}

func (co *ContextOptimizer) estimateContextTokens() int {
	total := 0
	for _, msg := range co.messages {
		total += len(msg.Content) / 4
	}
	return total
}

func (co *ContextOptimizer) Compress() string {
	if len(co.messages) == 0 {
		return ""
	}

	var summary strings.Builder
	summary.WriteString("Conversation summary:\n")

	roleCounts := make(map[string]int)
	for _, msg := range co.messages {
		roleCounts[msg.Role]++
	}

	for role, count := range roleCounts {
		summary.WriteString(fmt.Sprintf("- %s messages: %d\n", role, count))
	}

	if lastMsg := co.messages[len(co.messages)-1]; lastMsg != nil {
		summary.WriteString(fmt.Sprintf("- Last message (%s): %s\n", lastMsg.Role, truncate(lastMsg.Content, 100)))
	}

	remaining := co.messages
	if len(remaining) > 3 {
		remaining = remaining[len(remaining)-3:]
	}
	co.messages = remaining

	return summary.String()
}

func (co *ContextOptimizer) BuildOptimizedContext() string {
	if co.ShouldCompress() {
		compressed := co.Compress()
		if compressed != "" {
			return compressed
		}
	}

	var sb strings.Builder
	for _, msg := range co.GetRecentMessages(co.windowSize) {
		if msg.CacheKey != "" {
			if summary, ok := co.cache.GetSummary(msg.CacheKey); ok {
				sb.WriteString(fmt.Sprintf("[%s] %s (cached: %s)\n", msg.Role, msg.Content, summary))
				continue
			}
		}
		sb.WriteString(fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content))
	}
	return sb.String()
}

func (co *ContextOptimizer) ReplaceWithCacheKey(cacheKey string, originalContent string) {
	for i, msg := range co.messages {
		if msg.Content == originalContent {
			co.messages[i].CacheKey = cacheKey
			if summary, ok := co.cache.GetSummary(cacheKey); ok {
				co.messages[i].Content = summary
			}
			break
		}
	}
}

func (co *ContextOptimizer) Reset() {
	co.messages = make([]*ConversationEntry, 0)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type RollingWindowStats struct {
	TotalMessages  int
	CacheHits     int
	CacheMisses   int
	Compressions  int
	AvgContextSize int
}

func (co *ContextOptimizer) Stats() *RollingWindowStats {
	cacheHits := 0
	for _, msg := range co.messages {
		if msg.CacheKey != "" {
			if _, ok := co.cache.GetSummary(msg.CacheKey); ok {
				cacheHits++
			}
		}
	}

	return &RollingWindowStats{
		TotalMessages:  len(co.messages),
		CacheHits:     cacheHits,
		CacheMisses:   len(co.messages) - cacheHits,
		Compressions:  0,
		AvgContextSize: co.estimateContextTokens(),
	}
}

func BuildContextForMatch(matchID uuid.UUID, messages []*ConversationEntry, cache *ToolCache) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Match ID: %s\n\n", matchID.String()))
	sb.WriteString("## Recent Conversation\n")

	for i, msg := range messages {
		if msg == nil {
			continue
		}
		if msg.CacheKey != "" {
			if summary, ok := cache.GetSummary(msg.CacheKey); ok {
				sb.WriteString(fmt.Sprintf("%d. [%s] (cache: %s)\n", i+1, msg.Role, summary))
				continue
			}
		}
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, msg.Role, truncate(msg.Content, 200)))
	}

	return sb.String()
}