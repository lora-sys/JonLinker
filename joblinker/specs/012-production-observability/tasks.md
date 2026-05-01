# Tasks: Production Observability & Agent Monitoring

## Phase 1: Database & Models

- [x] T001 [P] Create `backend/internal/model/agent_metrics.go`
- [x] T002 [P] Create `backend/internal/model/audit_log.go` (combined in agent_metrics.go)
- [x] T003 [P] Create `backend/internal/model/error_event.go` (combined in agent_metrics.go)
- [x] T004 Add AutoMigrate for new models in `backend/cmd/server/main.go`

## Phase 2: Repositories

- [x] T005 [P] Create `backend/internal/repository/agent_metrics_repo.go` (includes AuditLogRepository, ErrorEventRepository)

## Phase 3: Core Services

- [x] T008 [P] Implement `backend/internal/service/observability_service.go` (MetricsService, AuditService, AlertingService)

## Phase 4: Admin Handler & API Endpoints

- [x] T014 [P] GET `/api/admin/metrics` - agent health summary
- [x] T015 [P] GET `/api/admin/audit` - query audit logs with filters
- [x] T016 [P] GET `/api/admin/errors` - error events list
- [x] T017 [P] GET `/api/admin/agent-metrics` - all agent metrics
- [x] T018 POST `/api/admin/errors/:id/resolve` - mark error resolved

## Phase 5: Integration (Pending)

- [ ] T011 Hook metrics collection into `message_queue_service.go`
- [ ] T012 Add audit logging to `message_handler.go`
- [ ] T013 Add error tracking to `tool_executor.go`

## Phase 6: Frontend Dashboard (Pending)

- [ ] T019 Create `frontend/src/app/admin/page.tsx` - admin dashboard
- [ ] T020 Create metrics cards component
- [ ] T021 Create audit log viewer with filters
- [ ] T022 Create error list with resolution status

## Phase 7: Verification (Pending)

- [ ] T023 Verify dashboard loads at /admin
- [ ] T024 Verify WebSocket updates appear in real-time
- [ ] T025 Verify audit log query by match_id works
- [ ] T026 Verify existing features still work (RabbitMQ, A2A, Vector DB)

## Summary

### Completed (Backend Core)
- Models: AgentMetrics, AuditLog, ErrorEvent
- Repository: Combined in agent_metrics_repo.go
- Services: MetricsService, AuditService, AlertingService
- Handler: AdminHandler with 5 endpoints

### Pending
- Integration hooks into existing code
- Frontend admin dashboard
- Real-time WebSocket updates

## Files Created

```
backend/internal/model/agent_metrics.go         # AgentMetrics, AuditLog, ErrorEvent
backend/internal/repository/agent_metrics_repo.go  # All 3 repos
backend/internal/service/observability_service.go # All 3 services
backend/internal/handler/admin.go               # Admin API handler
backend/cmd/server/main.go                      # Updated with new routes
```