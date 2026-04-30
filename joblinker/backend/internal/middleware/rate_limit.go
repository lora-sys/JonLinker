package middleware

import (
	"net/http"
	"time"

	"joblinker/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	MaxMessagesPerMinute = 10
	WindowDuration      = time.Minute
)

func RateLimit(rateLimitRepo *repository.RateLimitRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userIDVal.(string)
		if !ok {
			c.Next()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.Next()
			return
		}

		count, err := rateLimitRepo.GetMessageCount(userID, WindowDuration)
		if err != nil {
			LogError(c, err, userIDStr, "")
			c.Next()
			return
		}

		if count >= MaxMessagesPerMinute {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Error:         "Too many messages. Please wait before sending more.",
				Code:          ErrCodeRateLimited,
				CorrelationID: getCorrelationID(c),
			})
			c.Abort()
			return
		}

		_, err = rateLimitRepo.Increment(userID, WindowDuration)
		if err != nil {
			LogError(c, err, userIDStr, "")
		}

		c.Next()
	}
}

func getCorrelationID(c *gin.Context) string {
	if requestID, exists := c.Get("requestID"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
