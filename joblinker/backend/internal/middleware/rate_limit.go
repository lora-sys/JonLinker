package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MaxRequestsPerMinute = 60
	WindowDuration       = time.Minute
)

type rateEntry struct {
	count   int
	windowStart time.Time
}

type MemoryRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
}

var globalRateLimiter = &MemoryRateLimiter{
	entries: make(map[string]*rateEntry),
}

func (rl *MemoryRateLimiter) allow(key string, max int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.Sub(entry.windowStart) > window {
		rl.entries[key] = &rateEntry{count: 1, windowStart: now}
		return true
	}

	entry.count++
	return entry.count <= max
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rate limit by IP + userID (if available)
		key := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			if uid, ok := userID.(string); ok && uid != "" && uid != "anonymous" {
				key = uid
			}
		}

		if !globalRateLimiter.allow(key, MaxRequestsPerMinute, WindowDuration) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please wait before sending more requests.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
