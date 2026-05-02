package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ToolResult struct {
	Key     string
	Summary string
	Full    map[string]interface{}
	Created time.Time
	TTL     time.Duration
	Accesses int
}

type ToolCache struct {
	mu      sync.RWMutex
	data    map[string]*ToolResult
	maxSize int
	ttl     time.Duration
}

func NewToolCache(maxSize int, ttl time.Duration) *ToolCache {
	c := &ToolCache{
		data:    make(map[string]*ToolResult),
		maxSize: maxSize,
		ttl:     ttl,
	}
	go c.cleanupLoop()
	return c
}

func (c *ToolCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.cleanup()
	}
}

func (c *ToolCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.Sub(entry.Created) > entry.TTL {
			delete(c.data, key)
		}
	}

	if len(c.data) > c.maxSize {
		c.evictLRU()
	}
}

func (c *ToolCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range c.data {
		if oldestTime.IsZero() || entry.Created.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Created
		}
	}
	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

func (c *ToolCache) GenerateCacheKey(toolName string, args map[string]interface{}) string {
	argsBytes, _ := json.Marshal(args)
	hash := sha256.Sum256(argsBytes)
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("tool:%s:%s", toolName, hashStr[:16])
}

func (c *ToolCache) GenerateSummary(toolName string, result map[string]interface{}) string {
	switch toolName {
	case "query_jobs":
		return generateQueryJobsSummary(result)
	case "search_candidates":
		return generateSearchCandidatesSummary(result)
	case "get_candidate":
		return generateGetCandidateSummary(result)
	case "create_offer":
		return generateCreateOfferSummary(result)
	case "schedule_interview":
		return generateScheduleInterviewSummary(result)
	default:
		return fmt.Sprintf("%s result", toolName)
	}
}

func generateQueryJobsSummary(result map[string]interface{}) string {
	count := len(result)
	var location string
	var salaryMin, salaryMax int

	if jobs, ok := result["jobs"].([]interface{}); ok {
		count = len(jobs)
	}
	if loc, ok := result["location"].(string); ok {
		location = loc
	}
	if min, ok := result["salary_min"].(float64); ok {
		salaryMin = int(min)
	}
	if max, ok := result["salary_max"].(float64); ok {
		salaryMax = int(max)
	}

	summary := fmt.Sprintf("found %d jobs", count)
	if location != "" {
		summary += fmt.Sprintf(" in %s", location)
	}
	if salaryMin > 0 || salaryMax > 0 {
		summary += fmt.Sprintf(", salaries $%dk-$%dk", salaryMin/1000, salaryMax/1000)
	}
	return summary
}

func generateSearchCandidatesSummary(result map[string]interface{}) string {
	count := 0
	var skills []string

	if candidates, ok := result["candidates"].([]interface{}); ok {
		count = len(candidates)
	}
	if sk, ok := result["skills"].([]interface{}); ok {
		for _, s := range sk {
			if str, ok := s.(string); ok {
				skills = append(skills, str)
			}
		}
	}

	summary := fmt.Sprintf("found %d candidates", count)
	if len(skills) > 0 {
		summary += fmt.Sprintf(" matching %v", skills)
	}
	return summary
}

func generateGetCandidateSummary(result map[string]interface{}) string {
	var name string
	var years int
	var skills []string

	if n, ok := result["name"].(string); ok {
		name = n
	}
	if y, ok := result["years_of_experience"].(float64); ok {
		years = int(y)
	}
	if sk, ok := result["skills"].([]interface{}); ok {
		for _, s := range sk {
			if str, ok := s.(string); ok {
				skills = append(skills, str)
			}
		}
	}

	summary := "candidate: "
	if name != "" {
		summary += name
	}
	if years > 0 {
		summary += fmt.Sprintf(", %d years", years)
	}
	if len(skills) > 0 {
		summary += fmt.Sprintf(", skills: %v", skills)
	}
	return summary
}

func generateCreateOfferSummary(result map[string]interface{}) string {
	var salary int
	var startDate string

	if s, ok := result["salary"].(float64); ok {
		salary = int(s)
	}
	if sd, ok := result["start_date"].(string); ok {
		startDate = sd
	}

	summary := "offer created"
	if salary > 0 {
		summary += fmt.Sprintf(": $%d", salary)
	}
	if startDate != "" {
		summary += fmt.Sprintf(", start %s", startDate)
	}
	return summary
}

func generateScheduleInterviewSummary(result map[string]interface{}) string {
	var interviewType string
	var datetime string

	if t, ok := result["interview_type"].(string); ok {
		interviewType = t
	}
	if dt, ok := result["datetime"].(string); ok {
		datetime = dt
	}

	summary := "interview scheduled"
	if interviewType != "" {
		summary += fmt.Sprintf(": %s", interviewType)
	}
	if datetime != "" {
		summary += fmt.Sprintf(" at %s", datetime)
	}
	return summary
}

func (c *ToolCache) Set(key string, result map[string]interface{}) *ToolResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	summary := c.GenerateSummary(extractToolName(key), result)
	entry := &ToolResult{
		Key:      key,
		Summary:  summary,
		Full:     result,
		Created:  time.Now(),
		TTL:      c.ttl,
		Accesses: 0,
	}
	c.data[key] = entry

	if len(c.data) > c.maxSize {
		c.evictLRU()
	}

	return entry
}

func (c *ToolCache) Get(key string) (*ToolResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, ok := c.data[key]; ok {
		if time.Since(entry.Created) <= entry.TTL {
			entry.Accesses++
			return entry, true
		}
		delete(c.data, key)
	}
	return nil, false
}

func (c *ToolCache) OnDemandLoad(key string) (map[string]interface{}, bool) {
	entry, found := c.Get(key)
	if found {
		return entry.Full, true
	}
	return nil, false
}

func (c *ToolCache) GetSummary(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, ok := c.data[key]; ok {
		return entry.Summary, true
	}
	return "", false
}

func extractToolName(key string) string {
	var toolName string
	fmt.Sscanf(key, "tool:%s", &toolName)
	return toolName
}

func (c *ToolCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

func (c *ToolCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalAccesses := 0
	for _, entry := range c.data {
		totalAccesses += entry.Accesses
	}

	return map[string]interface{}{
		"size":         len(c.data),
		"max_size":    c.maxSize,
		"ttl_seconds": c.ttl.Seconds(),
		"total_accesses": totalAccesses,
	}
}

func CacheKeyForUUID(id uuid.UUID, toolName string) string {
	h := sha256.New()
	h.Write([]byte(id.String() + toolName))
	return fmt.Sprintf("tool:%s:%s", toolName, hex.EncodeToString(h.Sum(nil))[:16])
}