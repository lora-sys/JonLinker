<!-- Generated: 2026-05-01 | Files scanned: 187 | Token estimate: ~600 -->

# Backend Architecture

## Routes (cmd/server/main.go:127-183)

```
POST /api/auth/register  → authHandler.Register
POST /api/auth/login     → authHandler.Login
POST /api/auth/refresh   → authHandler.Refresh (JWT)

GET    /api/agents         → agentHandler.List
POST   /api/agents         → agentHandler.Create
GET    /api/agents/:id     → agentHandler.Get
PATCH  /api/agents/:id     → agentHandler.Update
DELETE /api/agents/:id     → agentHandler.Delete

GET    /api/jobs           → jobHandler.List
POST   /api/jobs           → jobHandler.Create
GET    /api/jobs/:id       → jobHandler.Get
PATCH  /api/jobs/:id       → jobHandler.Update

GET    /api/matches         → matchHandler.List
GET    /api/matches/:id     → matchHandler.Get
POST   /api/matches/auto    → matchHandler.AutoCreate
POST   /api/matches/:id/confirm → matchHandler.Confirm

GET    /api/interviews              → interviewHandler.List
POST   /api/interviews              → interviewHandler.Create
PATCH  /api/interviews/:id          → interviewHandler.Update
GET    /api/interviews/:matchId     → interviewHandler.GetByMatchID
POST   /api/interviews/:matchId/confirm → interviewHandler.Confirm
POST   /api/interviews/:matchId/cancel   → interviewHandler.Cancel

GET    /api/offers/:matchId        → offerHandler.GetByMatchID
POST   /api/offers                 → offerHandler.Create
POST   /api/offers/:matchId/accept  → offerHandler.Accept
POST   /api/offers/:matchId/decline → offerHandler.Decline

POST   /api/privacy/export          → privacyHandler.Export
DELETE /api/privacy/account         → privacyHandler.DeleteAccount

GET    /api/admin/metrics       → adminHandler.GetMetrics
GET    /api/admin/agent-metrics → adminHandler.GetAllAgentMetrics
GET    /api/admin/audit         → adminHandler.GetAuditLogs
GET    /api/admin/errors        → adminHandler.GetErrors
POST   /api/admin/errors/:id/resolve → adminHandler.ResolveError

GET    /api/messages/:matchId    → messageHandler.GetConversation
POST   /api/messages/:matchId    → messageHandler.SendMessage
GET    /api/messages/:matchId/ws → messageHandler.HandleWebSocket
```

## Middleware Chain

```
All requests:
  gin.Recovery() → middleware.Logger() → middleware.GatewayMiddleware()

/api/* (except /api/auth/*):
  + middleware.Auth() (JWT Bearer HS256)
```

## Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/server/main.go` | Entry point, wire-up | ~210 |
| `internal/agent/fsm.go` | Agent state machine (idle→searching→negotiating→interviewing→offer→hired/rejected) | ~135 |
| `internal/agent/tool_executor.go` | Tool execution (query_jobs, create_offer, schedule_interview, etc.) | ~200 |
| `internal/service/message_queue_service.go` | RabbitMQ consumer, AI auto-response, bidirectional A2A, FSM integration | ~560 |
| `internal/service/agent_memory_service.go` | Preference recall, real embeddings via AI API | ~187 |
| `internal/service/fsm_integration.go` | Intent→FSM event mapping, state transitions | ~90 |
| `internal/service/observability_service.go` | MetricsService, AuditService, AlertingService | ~200 |
| `internal/middleware/gateway.go` | X-User-ID/X-Agent-ID/X-Tenant-ID/X-Request-ID extraction | ~103 |
| `pkg/ai/client.go` | OpenAI-compatible AI client (Chat, ChatWithTools, GenerateEmbedding) | ~260 |
| `pkg/rabbitmq/rabbitmq.go` | RabbitMQ client (publish, consume, dead-letter, retry) | ~300 |

## Dependencies

- PostgreSQL (GORM + pgvector) — primary data store
- RabbitMQ — async agent messaging, dead-letter exchange
- Chroma (optional) — vector DB for embeddings
- AI API (LongCat/OpenAI-compatible) — chat completions, embeddings
