package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		logEntry := map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration_ms": duration.Milliseconds(),
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}
		if userID, exists := c.Get("userID"); exists {
			logEntry["user_id"] = userID
		}
		logJSON, _ := json.Marshal(logEntry)
		log.Println(string(logJSON))
	}
}
