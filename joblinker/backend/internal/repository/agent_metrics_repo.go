package repository

import (
	"time"

	"github.com/google/uuid"
	"joblinker/internal/model"
	"gorm.io/gorm"
)

type AgentMetricsRepository struct {
	db *gorm.DB
}

func NewAgentMetricsRepository() *AgentMetricsRepository {
	return &AgentMetricsRepository{}
}

func (r *AgentMetricsRepository) WithDB(db *gorm.DB) *AgentMetricsRepository {
	return &AgentMetricsRepository{db: db}
}

func (r *AgentMetricsRepository) Create(metrics *model.AgentMetrics) error {
	return r.db.Create(metrics).Error
}

func (r *AgentMetricsRepository) Update(metrics *model.AgentMetrics) error {
	return r.db.Save(metrics).Error
}

func (r *AgentMetricsRepository) GetByAgentID(agentID uuid.UUID) (*model.AgentMetrics, error) {
	var metrics model.AgentMetrics
	err := r.db.Where("agent_id = ?", agentID).First(&metrics).Error
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}

func (r *AgentMetricsRepository) ListByType(agentType string) ([]model.AgentMetrics, error) {
	var metrics []model.AgentMetrics
	err := r.db.Where("agent_type = ?", agentType).Find(&metrics).Error
	return metrics, err
}

func (r *AgentMetricsRepository) GetAll() ([]model.AgentMetrics, error) {
	var metrics []model.AgentMetrics
	err := r.db.Find(&metrics).Error
	return metrics, err
}

func (r *AgentMetricsRepository) IncrementMessagesProcessed(agentID uuid.UUID) error {
	return r.db.Model(&model.AgentMetrics{}).
		Where("agent_id = ?", agentID).
		UpdateColumn("messages_processed", gorm.Expr("messages_processed + 1")).
		Error
}

func (r *AgentMetricsRepository) IncrementErrors(agentID uuid.UUID) error {
	return r.db.Model(&model.AgentMetrics{}).
		Where("agent_id = ?", agentID).
		UpdateColumn("errors_count", gorm.Expr("errors_count + 1")).
		Error
}

func (r *AgentMetricsRepository) UpdateHeartbeat(agentID uuid.UUID) error {
	return r.db.Model(&model.AgentMetrics{}).
		Where("agent_id = ?", agentID).
		Update("last_heartbeat", time.Now()).
		Error
}

func (r *AgentMetricsRepository) GetActiveAgentsSummary() (seekerCount, recruiterCount, activeConversations int64, err error) {
	err = r.db.Model(&model.AgentMetrics{}).Where("agent_type = ? AND last_heartbeat > ?", "seeker", time.Now().Add(-5*time.Minute)).Count(&seekerCount).Error
	if err != nil {
		return
	}
	err = r.db.Model(&model.AgentMetrics{}).Where("agent_type = ? AND last_heartbeat > ?", "recruiter", time.Now().Add(-5*time.Minute)).Count(&recruiterCount).Error
	if err != nil {
		return
	}
	err = r.db.Model(&model.AgentMetrics{}).Select("COALESCE(SUM(conversations_active), 0)").Where("last_heartbeat > ?", time.Now().Add(-5*time.Minute)).Scan(&activeConversations).Error
	return
}

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{}
}

func (r *AuditLogRepository) WithDB(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditLogRepository) GetByMatchID(matchID uuid.UUID, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.db.Where("match_id = ?", matchID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *AuditLogRepository) GetByAgentID(agentID uuid.UUID, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.db.Where("agent_id = ?", agentID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *AuditLogRepository) GetByTimeRange(start, end time.Time, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.db.Where("timestamp BETWEEN ? AND ?", start, end).
		Order("timestamp DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *AuditLogRepository) GetByEventType(eventType string, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.db.Where("event_type = ?", eventType).
		Order("timestamp DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

type ErrorEventRepository struct {
	db *gorm.DB
}

func NewErrorEventRepository() *ErrorEventRepository {
	return &ErrorEventRepository{}
}

func (r *ErrorEventRepository) WithDB(db *gorm.DB) *ErrorEventRepository {
	return &ErrorEventRepository{db: db}
}

func (r *ErrorEventRepository) Create(event *model.ErrorEvent) error {
	return r.db.Create(event).Error
}

func (r *ErrorEventRepository) GetByID(id uuid.UUID) (*model.ErrorEvent, error) {
	var event model.ErrorEvent
	err := r.db.Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *ErrorEventRepository) ListUnresolved(limit int) ([]model.ErrorEvent, error) {
	var events []model.ErrorEvent
	err := r.db.Where("resolved = ?", false).
		Order("timestamp DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *AuditLogRepository) Query(filter struct {
	MatchID   *uuid.UUID
	AgentID   *uuid.UUID
	EventType string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
}) ([]model.AuditLog, error) {
	query := r.db.Model(&model.AuditLog{})

	if filter.MatchID != nil {
		query = query.Where("match_id = ?", *filter.MatchID)
	}
	if filter.AgentID != nil {
		query = query.Where("agent_id = ?", *filter.AgentID)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(100)
	}

	var logs []model.AuditLog
	err := query.Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

func (r *ErrorEventRepository) CountRecentByType(errorType string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.ErrorEvent{}).
		Where("error_type = ? AND timestamp >= ?", errorType, since).
		Count(&count).Error
	return count, err
}

func (r *ErrorEventRepository) MarkResolved(id uuid.UUID) error {
	return r.db.Model(&model.ErrorEvent{}).
		Where("id = ?", id).
		Update("resolved", true).
		Error
}