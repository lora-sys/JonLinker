package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	allowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001")
	allowedMethods := getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	allowedHeaders := getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Request-ID,X-User-ID,X-Agent-ID,X-Tenant-ID,Sec-WebSocket-Protocol")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is in the allowed list (wildcard "*" not allowed)
		origins := strings.Split(allowedOrigins, ",")
		allowed := false
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
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

		// Security headers (set for all responses)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
