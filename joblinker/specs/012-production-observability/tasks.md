# Tasks: Production Observability & Agent Monitoring

## Phase 1: Database & Models

- [ ] T001 [P] Create `backend/internal/model/agent_metrics.go`
- [ ] T002 [P] Create `backend/internal/model/audit_log.go`
- [ ] T003 [P] Create `backend/internal/model/error_event.go`
- [ ] T004 Add AutoMigrate for new models in `backend/cmd/server/main.go`

## Phase 2: Repositories

- [ ] T005 [P] Create `backend/internal/repository/agent_metrics_repo.go`
- [ ] T006 [P] Create `backend/internal/repository/audit_log_repo.go`
- [ ] T007 [P] Create `backend/internal/repository/error_event_repo.go`

## Phase 3: Core Services

- [ ] T008 [P] Implement `backend/internal/service/metrics_service.go`
- [ ] T009 [P] Implement `backend/internal/service/audit_service.go`
- [ ] T010 [P] Implement `backend/internal/service/alerting_service.go`

## Phase 4: Integration

- [ ] T011 Hook metrics collection into `message_queue_service.go`
- [ ] T012 Add audit logging to `message_handler.go`
- [ ] T013 Add error tracking to `tool_executor.go`

## Phase 5: Admin API Endpoints

- [ ] T014 [P] GET `/api/admin/metrics` - agent health summary
- [ ] T015 [P] GET `/api/admin/audit` - query audit logs with filters
- [ ] T016 [P] GET `/api/admin/errors` - error events list
- [ ] T017 WebSocket `/api/admin/ws` - real-time dashboard updates

## Phase 6: Admin Handler

- [ ] T018 Create `backend/internal/handler/admin.go` with all admin endpoints

## Phase 7: Frontend Dashboard

- [ ] T019 Create `frontend/src/app/admin/page.tsx` - admin dashboard
- [ ] T020 Create metrics cards component
- [ ] T021 Create audit log viewer with filters
- [ ] T022 Create error list with resolution status

## Phase 8: Verification

- [ ] T023 Verify dashboard loads at /admin
- [ ] T024 Verify WebSocket updates appear in real-time
- [ ] T025 Verify audit log query by match_id works
- [ ] T026 Verify existing features still work (RabbitMQ, A2A, Vector DB)

## Dependencies

- Phase 1 must complete before Phase 2
- Phase 2 must complete before Phase 3
- Phase 3 must complete before Phase 4
- Phase 5, 6, 7 can run in parallel after Phase 3
- Phase 8 is final verification

## Files to Create

### Backend Models
```
backend/internal/model/
├── agent_metrics.go
├── audit_log.go
└── error_event.go
```

### Backend Repositories
```
backend/internal/repository/
├── agent_metrics_repo.go
├── audit_log_repo.go
└── error_event_repo.go
```

### Backend Services
```
backend/internal/service/
├── metrics_service.go
├── audit_service.go
└── alerting_service.go
```

### Backend Handlers
```
backend/internal/handler/
└── admin.go
```

### Frontend
```
frontend/src/app/admin/
└── page.tsx
```