package middleware

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"joblinker/internal/config"
	"joblinker/pkg/proto"
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

// ResponseWriter wraps gin.ResponseWriter to capture response for Protobuf encoding
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// ProtoResponseWriter creates a response writer that captures the body for Protobuf encoding
func ProtoResponseWriter(c *gin.Context) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
	}
}

// ShouldUseProtobuf determines if the response should be Protobuf based on Accept header
func ShouldUseProtobuf(c *gin.Context) bool {
	if !config.IsInternalProtobufEnabled() {
		return false
	}
	accept := c.GetHeader("Accept")
	return proto.ParseAcceptHeader(accept) == proto.ContentTypeProtobuf
}

// EncodeProtobuf marshals the response body as Protobuf
func EncodeProtobuf(c *gin.Context, data []byte) error {
	c.Header("Content-Type", string(proto.ContentTypeProtobuf))
	_, err := c.Writer.Write(data)
	return err
}

// EncodeJSON encodes data as JSON response
func EncodeJSON(c *gin.Context, data interface{}) {
	c.Header("Content-Type", "application/json")
	c.JSON(200, data)
}

// ProtoMiddleware wraps response writer and handles Protobuf encoding
func ProtoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ShouldUseProtobuf(c) {
			w := ProtoResponseWriter(c)
			c.Writer = w
			c.Set("protoWriter", w)
		}
		c.Next()
	}
}

// ProtoResponseMiddleware captures JSON response body and re-encodes as Protobuf
func ProtoResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		useProto := ShouldUseProtobuf(c)
		log.Printf("ProtoResponseMiddleware: path=%s accept=%s useProto=%v", c.Request.URL.Path, c.GetHeader("Accept"), useProto)
		if !useProto {
			c.Next()
			return
		}

		// Set Content-Type header BEFORE the handler writes the response
		c.Header("Content-Type", string(proto.ContentTypeProtobuf))
		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// GetProtoBody returns the captured response body from Protobuf writer
func GetProtoBody(c *gin.Context) []byte {
	if w, ok := c.Get("protoWriter"); ok {
		if pw, ok := w.(*ResponseWriter); ok {
			return pw.body.Bytes()
		}
	}
	return nil
}

// ReadBody reads and restores the request body for downstream handlers
func ReadBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}