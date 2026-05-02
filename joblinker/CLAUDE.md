<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
specs/012-production-observability/plan.md

Current feature scope: Production Observability & Agent Monitoring — admin dashboard,
audit logging, error tracking, real-time metrics, WebSocket updates. Key components:
MetricsService, AuditService, AlertingService, admin dashboard at /admin.

## API Gateway Middleware

All HTTP requests pass through `middleware.GatewayMiddleware()` registered in `cmd/server/main.go:47`.
The middleware extracts/stores these headers in Gin context:

| Header | Context Key | Fallback |
|--------|------------|----------|
| `X-Request-ID` | `requestID` | auto-generated UUID |
| `X-User-ID` | `userID` | `"anonymous"` |
| `X-Agent-ID` | `agentID` | `""` (empty) |
| `X-Tenant-ID` | `tenantID` | `"default"` |

Helper accessors: `middleware.GetUserID(c)`, `middleware.GetAgentID(c)`, `middleware.GetTenantID(c)`, `middleware.GetCorrelationID(c)`.

Usage in handlers:
```go
userID := middleware.GetUserID(c)
agentID := middleware.GetAgentID(c)
```

## FSM State Machine

Match lifecycle is tracked via `service.FSMIntegration` using `agent.FSM` states:
`idle → searching → matched → negotiating → interviewing → offered → accepted/rejected`.

Intent-to-event mapping in `fsm_integration.go:22-35`. Transition on each agent message via `handleAgentMessage`.

## Embedding Configuration

Real embeddings via OpenAI-compatible API. Set env vars:
- `AI_API_KEY` — API key for embedding generation
- `AI_BASE_URL` — API base URL (defaults to OpenAI)
- `CHROMA_HOST` — optional, enables Chroma storage fallback (otherwise pgvector)

Embedding dimension: 1536 (OpenAI text-embedding-ada-002 compatible).
Mock fallback with zero vectors if API is unavailable (logged as WARNING).
<!-- SPECKIT END -->
