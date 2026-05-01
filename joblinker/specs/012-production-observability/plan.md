# Plan: Production Observability & Agent Monitoring

**Feature Branch**: `012-production-observability`
**Created**: 2026-05-01
**Status**: Draft

## Technical Context

### Tech Stack
- **Backend**: Go + Gin + GORM
- **Database**: PostgreSQL (metrics, audit logs, errors)
- **Frontend**: Next.js (admin dashboard)
- **Real-time**: WebSocket for live updates

### Project Structure
```
backend/
├── internal/
│   ├── handler/
│   │   └── admin.go          # Admin dashboard endpoint
│   ├── service/
│   │   ├── agent_metrics.go  # Metrics collection
│   │   ├── audit_service.go  # Audit logging
│   │   └── alerting_service.go # Webhook alerts
│   └── model/
│       ├── agent_metrics.go
│       ├── audit_log.go
│       └── error_event.go
frontend/
└── src/app/admin/
    └── page.tsx              # Admin dashboard
```

### Key Design Decisions

1. **Audit logging is async** - don't block agent response for logging
2. **Metrics aggregated in-memory** - expensive to query pgvector for counts
3. **WebSocket broadcast** - dashboard subscribes to metrics topic
4. **Error deduplication** - same error within 5 min = one alert

## Implementation Plan

### Phase 1: Database & Models

- [ ] Create agent_metrics table and model
- [ ] Create audit_logs table and model
- [ ] Create error_events table and model
- [ ] Add AutoMigrate to main.go

### Phase 2: Core Services

- [ ] Implement MetricsService (collect, aggregate, query)
- [ ] Implement AuditService (log events, query by match_id)
- [ ] Implement AlertingService (webhook calls on error spike)

### Phase 3: Integration Points

- [ ] Hook agent_metrics into MessageQueueService
- [ ] Add audit logging to MessageHandler
- [ ] Add error tracking to tool_executor

### Phase 4: Admin API

- [ ] GET /api/admin/metrics - agent health summary
- [ ] GET /api/admin/audit?match_id=xxx - query audit logs
- [ ] GET /api/admin/errors - error events list
- [ ] WebSocket /api/admin/ws - real-time dashboard

### Phase 5: Frontend Dashboard

- [ ] Admin page at /admin
- [ ] Metrics cards (active agents, error rate, response time)
- [ ] Audit log viewer with filters
- [ ] Error list with resolution status

## Data Model

### AgentMetrics
```go
type AgentMetrics struct {
    ID                    uuid.UUID
    AgentID               uuid.UUID
    AgentType             string // "seeker" | "recruiter"
    MessagesProcessed     int
    ErrorsCount           int
    AvgResponseTimeMs     int
    ConversationsActive   int
    ConversationsCompleted int
    LastHeartbeat         time.Time
}
```

### AuditLog
```go
type AuditLog struct {
    ID          uuid.UUID
    AgentID     uuid.UUID
    MatchID     uuid.UUID
    EventType   string // "message_sent" | "tool_call" | "decision" | "error"
    EventData   JSONB
    Timestamp   time.Time
}
```

### ErrorEvent
```go
type ErrorEvent struct {
    ID           uuid.UUID
    ErrorType    string // "ai_failure" | "rabbitmq_failure" | "db_timeout" | "tool_error"
    ErrorMessage string
    StackTrace   string
    Context      JSONB
    Resolved     bool
    Timestamp    time.Time
}
```

## Verification

- [ ] Dashboard loads at /admin
- [ ] Metrics update in real-time via WebSocket
- [ ] Audit log query returns results for given match_id
- [ ] Error events are logged and displayable
- [ ] All existing features (A2A, RabbitMQ, Vector DB) still work