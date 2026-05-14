package handler

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type checkResult struct {
	Status  string `json:"status"`
	Detail  string `json:"detail,omitempty"`
}

type healthResponse struct {
	Status string                  `json:"status"`
	Checks map[string]checkResult `json:"checks"`
}

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

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

	// Check RabbitMQ (best effort - check if env is configured)
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL != "" {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://" + rmqURL + "/api/health/checks/alarms")
		if err != nil {
			checks["rabbitmq"] = checkResult{Status: "down", Detail: "Cannot reach RabbitMQ management API"}
			if overall == "healthy" {
				overall = "degraded"
			}
		} else {
			resp.Body.Close()
			checks["rabbitmq"] = checkResult{Status: "ok"}
		}
	} else {
		checks["rabbitmq"] = checkResult{Status: "not_configured"}
	}

	// Check AI API (best effort)
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
