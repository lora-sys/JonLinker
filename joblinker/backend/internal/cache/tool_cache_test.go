package cache

import (
	"testing"
	"time"
)

func TestToolCache(t *testing.T) {
	cache := NewToolCache(10, time.Minute)

	// Test GenerateCacheKey
	args := map[string]interface{}{"location": "SF", "skills": []interface{}{"golang"}}
	key := cache.GenerateCacheKey("query_jobs", args)
	if key == "" {
		t.Error("GenerateCacheKey should return non-empty key")
	}
	if len(key) < 10 {
		t.Error("GenerateCacheKey should return reasonable length key")
	}

	// Test Set and Get
	result := map[string]interface{}{
		"jobs":  []interface{}{},
		"count": 5,
	}
	cached := cache.Set(key, result)
	if cached == nil {
		t.Error("Set should return non-nil ToolResult")
	}
	if cached.Summary == "" {
		t.Error("Set should generate summary")
	}

	// Test Get
	got, found := cache.Get(key)
	if !found {
		t.Error("Get should find cached entry")
	}
	if got.Summary != cached.Summary {
		t.Error("Get should return same summary")
	}

	// Test cache miss
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Get should return false for missing key")
	}
}

func TestToolCache_LRU(t *testing.T) {
	cache := NewToolCache(3, time.Minute)

	// Fill beyond capacity
	for i := 0; i < 5; i++ {
		key := cache.GenerateCacheKey("query_jobs", map[string]interface{}{"index": i})
		cache.Set(key, map[string]interface{}{"count": i})
	}

	// Should have at most maxSize entries
	size := len(cache.Keys())
	if size > 3 {
		t.Errorf("Cache should evict, got %d entries", size)
	}
}

func TestGenerateSummary(t *testing.T) {
	cache := NewToolCache(10, time.Minute)

	tests := []struct {
		toolName string
		result   map[string]interface{}
		want     string
	}{
		{"query_jobs", map[string]interface{}{"count": 5, "location": "SF"}, "found 5 jobs in SF"},
		{"create_offer", map[string]interface{}{"salary": 150000.0, "start_date": "2026-07-01"}, "offer created: $150000, start 2026-07-01"},
		{"schedule_interview", map[string]interface{}{"interview_type": "video", "datetime": "2026-06-01T10:00:00Z"}, "interview scheduled: video at 2026-06-01T10:00:00Z"},
	}

	for _, tt := range tests {
		summary := cache.GenerateSummary(tt.toolName, tt.result)
		if summary == "" {
			t.Errorf("GenerateSummary(%s) returned empty", tt.toolName)
		}
	}
}