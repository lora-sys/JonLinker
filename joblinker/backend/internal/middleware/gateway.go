package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GatewayMiddleware extracts and validates tenant headers from all requests
// This middleware provides:
// - X-User-ID → context userID
// - X-Agent-ID → context agentID
// - X-Tenant-ID → context tenantID
// - X-Request-ID → correlation ID for tracing

func GatewayMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Extract or generate correlation ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		// Extract tenant headers
		userID := c.GetHeader("X-User-ID")
		agentID := c.GetHeader("X-Agent-ID")
		tenantID := c.GetHeader("X-Tenant-ID")

		// Set defaults if not provided (for backward compatibility)
		if userID == "" {
			userID = "anonymous"
		}
		if tenantID == "" {
			tenantID = "default"
		}

		// Store in context for handlers to access
		c.Set("userID", userID)
		c.Set("agentID", agentID)
		c.Set("tenantID", tenantID)

		// Log request with correlation ID
		log.Printf("[%s] %s %s - userID=%s agentID=%s tenantID=%s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			userID,
			agentID,
			tenantID,
		)

		// Continue request
		c.Next()

		// Log response time
		duration := time.Since(start)
		log.Printf("[%s] Completed %s %s - %d (%v)",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

// GetUserID extracts userID from Gin context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		return userID.(string)
	}
	return ""
}

// GetAgentID extracts agentID from Gin context
func GetAgentID(c *gin.Context) string {
	if agentID, exists := c.Get("agentID"); exists {
		return agentID.(string)
	}
	return ""
}

// GetTenantID extracts tenantID from Gin context
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenantID"); exists {
		return tenantID.(string)
	}
	return ""
}

// GetCorrelationID extracts correlation ID from Gin context
func GetCorrelationID(c *gin.Context) string {
	if requestID, exists := c.Get("requestID"); exists {
		return requestID.(string)
	}
	return ""
}