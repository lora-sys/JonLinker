package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"joblinker/internal/model"
	"joblinker/internal/repository"
)

type MetricsService struct {
	metricsRepo *repository.AgentMetricsRepository
}

func NewMetricsService(metricsRepo *repository.AgentMetricsRepository) *MetricsService {
	return &MetricsService{metricsRepo: metricsRepo}
}

func (s *MetricsService) GetOrCreateMetrics(agentID uuid.UUID, agentType string) (*model.AgentMetrics, error) {
	metrics, err := s.metricsRepo.GetByAgentID(agentID)
	if err != nil {
		// Create new metrics entry
		metrics = &model.AgentMetrics{
			AgentID:       agentID,
			AgentType:     agentType,
			LastHeartbeat: time.Now(),
		}
		if err := s.metricsRepo.Create(metrics); err != nil {
			return nil, err
		}
		return metrics, nil
	}
	return metrics, nil
}

func (s *MetricsService) RecordMessage(agentID uuid.UUID) error {
	return s.metricsRepo.IncrementMessagesProcessed(agentID)
}

func (s *MetricsService) RecordError(agentID uuid.UUID) error {
	return s.metricsRepo.IncrementErrors(agentID)
}

func (s *MetricsService) UpdateHeartbeat(agentID uuid.UUID) error {
	return s.metricsRepo.UpdateHeartbeat(agentID)
}

func (s *MetricsService) UpdateResponseTime(agentID uuid.UUID, responseTimeMs int) error {
	metrics, err := s.metricsRepo.GetByAgentID(agentID)
	if err != nil {
		return err
	}

	// Running average calculation
	totalResponses := metrics.MessagesProcessed
	if totalResponses == 0 {
		totalResponses = 1
	}
	newAvg := (metrics.AvgResponseTimeMs*totalResponses + responseTimeMs) / (totalResponses + 1)
	metrics.AvgResponseTimeMs = newAvg

	return s.metricsRepo.Update(metrics)
}

func (s *MetricsService) IncrementActiveConversations(agentID uuid.UUID) error {
	metrics, err := s.metricsRepo.GetByAgentID(agentID)
	if err != nil {
		return err
	}
	metrics.ConversationsActive++
	return s.metricsRepo.Update(metrics)
}

func (s *MetricsService) DecrementActiveConversations(agentID uuid.UUID) error {
	metrics, err := s.metricsRepo.GetByAgentID(agentID)
	if err != nil {
		return err
	}
	if metrics.ConversationsActive > 0 {
		metrics.ConversationsActive--
	}
	return s.metricsRepo.Update(metrics)
}

func (s *MetricsService) CompleteConversation(agentID uuid.UUID) error {
	metrics, err := s.metricsRepo.GetByAgentID(agentID)
	if err != nil {
		return err
	}
	if metrics.ConversationsActive > 0 {
		metrics.ConversationsActive--
	}
	metrics.ConversationsCompleted++
	return s.metricsRepo.Update(metrics)
}

func (s *MetricsService) GetAllMetrics() ([]model.AgentMetrics, error) {
	return s.metricsRepo.GetAll()
}

func (s *MetricsService) GetMetricsByType(agentType string) ([]model.AgentMetrics, error) {
	return s.metricsRepo.ListByType(agentType)
}

func (s *MetricsService) GetActiveSummary() (seekerCount, recruiterCount, activeConversations int64, err error) {
	return s.metricsRepo.GetActiveAgentsSummary()
}

type AuditService struct {
	auditRepo *repository.AuditLogRepository
}

func NewAuditService(auditRepo *repository.AuditLogRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

func (s *AuditService) LogEvent(ctx context.Context, agentID, matchID uuid.UUID, eventType string, eventData map[string]interface{}) error {
	log := &model.AuditLog{
		AgentID:   agentID,
		MatchID:   matchID,
		EventType: eventType,
		EventData: eventData,
		Timestamp: time.Now(),
	}
	return s.auditRepo.Create(log)
}

func (s *AuditService) LogMessageSent(agentID, matchID uuid.UUID, content string) error {
	return s.LogEvent(context.Background(), agentID, matchID, "message_sent", map[string]interface{}{
		"content": content,
	})
}

func (s *AuditService) LogToolCall(agentID, matchID uuid.UUID, toolName string, args map[string]interface{}, result string) error {
	return s.LogEvent(context.Background(), agentID, matchID, "tool_call", map[string]interface{}{
		"tool_name": toolName,
		"arguments": args,
		"result":    result,
	})
}

func (s *AuditService) LogDecision(agentID, matchID uuid.UUID, decision string, reasoning string) error {
	return s.LogEvent(context.Background(), agentID, matchID, "decision", map[string]interface{}{
		"decision":  decision,
		"reasoning": reasoning,
	})
}

func (s *AuditService) LogError(agentID, matchID uuid.UUID, errorMsg string, details map[string]interface{}) error {
	eventData := map[string]interface{}{
		"error_message": errorMsg,
	}
	for k, v := range details {
		eventData[k] = v
	}
	return s.LogEvent(context.Background(), agentID, matchID, "error", eventData)
}

func (s *AuditService) GetByMatchID(matchID uuid.UUID, limit int) ([]model.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.auditRepo.GetByMatchID(matchID, limit)
}

func (s *AuditService) GetByAgentID(agentID uuid.UUID, limit int) ([]model.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.auditRepo.GetByAgentID(agentID, limit)
}

func (s *AuditService) Query(filter struct {
	MatchID   *uuid.UUID
	AgentID   *uuid.UUID
	EventType string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
}) ([]model.AuditLog, error) {
	return s.auditRepo.Query(filter)
}

type AlertingService struct {
	errorRepo *repository.ErrorEventRepository
	alertWebhook string
	errorCounts  map[string]int
	errorWindow time.Time
}

func NewAlertingService(errorRepo *repository.ErrorEventRepository, alertWebhook string) *AlertingService {
	return &AlertingService{
		errorRepo:    errorRepo,
		alertWebhook: alertWebhook,
		errorCounts:  make(map[string]int),
		errorWindow:  time.Now(),
	}
}

func (s *AlertingService) RecordError(errorType, errorMsg, stackTrace string, context map[string]interface{}) error {
	event := &model.ErrorEvent{
		ErrorType:    errorType,
		ErrorMessage: errorMsg,
		StackTrace:   stackTrace,
		Context:      context,
		Resolved:     false,
		Timestamp:    time.Now(),
	}

	if err := s.errorRepo.Create(event); err != nil {
		return err
	}

	// Check if we need to alert
	s.incrementErrorCount(errorType)
	return nil
}

func (s *AlertingService) incrementErrorCount(errorType string) {
	// Reset window every minute
	if time.Since(s.errorWindow) > time.Minute {
		s.errorCounts = make(map[string]int)
		s.errorWindow = time.Now()
	}

	s.errorCounts[errorType]++

	// Alert if threshold exceeded
	if s.errorCounts[errorType] >= 10 {
		s.sendAlert(errorType)
	}
}

func (s *AlertingService) sendAlert(errorType string) {
	// TODO: Implement webhook call
	// For now just log
	// In production, this would POST to s.alertWebhook
}

func (s *AlertingService) GetUnresolvedErrors(limit int) ([]model.ErrorEvent, error) {
	return s.errorRepo.ListUnresolved(limit)
}

func (s *AlertingService) ResolveError(id uuid.UUID) error {
	return s.errorRepo.MarkResolved(id)
}

func (s *AlertingService) GetRecentErrorCount(errorType string) (int64, error) {
	return s.errorRepo.CountRecentByType(errorType, time.Now().Add(-5*time.Minute))
}