package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	allowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001")
	allowedMethods := getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	allowedHeaders := getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Request-ID,X-User-ID,X-Agent-ID,X-Tenant-ID")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		origins := strings.Split(allowedOrigins, ",")
		allowed := false
		for _, o := range origins {
			if strings.TrimSpace(o) == origin || strings.TrimSpace(o) == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", allowedMethods)
			c.Header("Access-Control-Allow-Headers", allowedHeaders)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
