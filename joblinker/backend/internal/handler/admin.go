package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"joblinker/internal/repository"
	"joblinker/internal/service"
)

type AdminHandler struct {
	metricsRepo *repository.AgentMetricsRepository
	auditRepo   *repository.AuditLogRepository
	errorRepo   *repository.ErrorEventRepository
	metricsSvc  *service.MetricsService
	auditSvc    *service.AuditService
	alertSvc    *service.AlertingService
}

func NewAdminHandler(
	metricsRepo *repository.AgentMetricsRepository,
	auditRepo *repository.AuditLogRepository,
	errorRepo *repository.ErrorEventRepository,
) *AdminHandler {
	return &AdminHandler{
		metricsRepo: metricsRepo,
		auditRepo:   auditRepo,
		errorRepo:   errorRepo,
		metricsSvc:  service.NewMetricsService(metricsRepo),
		auditSvc:    service.NewAuditService(auditRepo),
		alertSvc:    service.NewAlertingService(errorRepo, ""),
	}
}

type MetricsSummary struct {
	ActiveSeekers       int64 `json:"active_seekers"`
	ActiveRecruiters    int64 `json:"active_recruiters"`
	ActiveConversations int64 `json:"active_conversations"`
	TotalErrors         int64 `json:"total_errors_unresolved"`
}

func (h *AdminHandler) GetMetrics(c *gin.Context) {
	seekers, recruiters, active, err := h.metricsSvc.GetActiveSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	errors, _ := h.alertSvc.GetUnresolvedErrors(100)
	unresolvedCount := int64(len(errors))

	c.JSON(http.StatusOK, MetricsSummary{
		ActiveSeekers:       seekers,
		ActiveRecruiters:    recruiters,
		ActiveConversations: active,
		TotalErrors:         unresolvedCount,
	})
}

func (h *AdminHandler) GetAllAgentMetrics(c *gin.Context) {
	metrics, err := h.metricsSvc.GetAllMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	matchIDStr := c.Query("match_id")
	agentIDStr := c.Query("agent_id")
	eventType := c.Query("event_type")
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")
	limitStr := c.DefaultQuery("limit", "100")

	limit, _ := strconv.Atoi(limitStr)

	filter := struct {
		MatchID   *uuid.UUID
		AgentID   *uuid.UUID
		EventType string
		StartTime *time.Time
		EndTime   *time.Time
		Limit     int
	}{
		Limit: limit,
	}

	if matchIDStr != "" {
		id, err := uuid.Parse(matchIDStr)
		if err == nil {
			filter.MatchID = &id
		}
	}
	if agentIDStr != "" {
		id, err := uuid.Parse(agentIDStr)
		if err == nil {
			filter.AgentID = &id
		}
	}
	if eventType != "" {
		filter.EventType = eventType
	}
	if startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			filter.StartTime = &t
		}
	}
	if endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			filter.EndTime = &t
		}
	}

	logs, err := h.auditSvc.Query(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *AdminHandler) GetErrors(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	errors, err := h.alertSvc.GetUnresolvedErrors(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, errors)
}

func (h *AdminHandler) ResolveError(c *gin.Context) {
	errorID := c.Param("id")
	id, err := uuid.Parse(errorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid error ID"})
		return
	}

	if err := h.alertSvc.ResolveError(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Error marked as resolved"})
}

func (h *AdminHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
