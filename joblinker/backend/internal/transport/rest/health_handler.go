package rest

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type checkResult struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type healthResponse struct {
	Status string                  `json:"status"`
	Checks map[string]checkResult `json:"checks"`
}

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health handles GET /health.
func (h *HealthHandler) Health(c *gin.Context) {
	overall := "healthy"
	checks := make(map[string]checkResult)

	// Check PostgreSQL
	if h.db != nil {
		if err := h.db.PingContext(c.Request.Context()); err != nil {
			checks["db"] = checkResult{Status: "down", Detail: err.Error()}
			overall = "degraded"
		} else {
			checks["db"] = checkResult{Status: "ok"}
		}
	} else {
		checks["db"] = checkResult{Status: "not_configured"}
	}

	// Check RabbitMQ (best effort - try to connect to AMQP port)
	rmqUser := os.Getenv("RABBITMQ_USER")
	rmqPass := os.Getenv("RABBITMQ_PASS")
	rmqHost := os.Getenv("RABBITMQ_HOST")
	rmqPort := os.Getenv("RABBITMQ_PORT")
	if rmqHost == "" {
		rmqHost = "localhost"
	}
	if rmqPort == "" {
		rmqPort = "5672"
	}
	addr := net.JoinHostPort(rmqHost, rmqPort)
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		checks["rabbitmq"] = checkResult{Status: "down", Detail: fmt.Sprintf("Cannot connect to %s: %v", addr, err)}
		if overall == "healthy" {
			overall = "degraded"
		}
	} else {
		conn.Close()
		if rmqUser != "" && rmqPass != "" {
			checks["rabbitmq"] = checkResult{Status: "ok", Detail: fmt.Sprintf("Configured AMQP %s", addr)}
		} else {
			checks["rabbitmq"] = checkResult{Status: "ok", Detail: fmt.Sprintf("Connected to %s (no credentials)", addr)}
		}
	}

	// Check AI API availability
	aiURL := os.Getenv("AI_BASE_URL")
	if aiURL != "" {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(aiURL)
		if err != nil {
			checks["ai_api"] = checkResult{Status: "down", Detail: "Cannot reach AI API"}
			if overall == "healthy" {
				overall = "degraded"
			}
		} else {
			resp.Body.Close()
			checks["ai_api"] = checkResult{Status: "ok"}
		}
	} else {
		checks["ai_api"] = checkResult{Status: "not_configured"}
	}

	c.JSON(http.StatusOK, healthResponse{
		Status: overall,
		Checks: checks,
	})
}
