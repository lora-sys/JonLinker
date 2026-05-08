<!-- SPECKIT START -->
## ⚠️ MANDATORY COMMAND RULE

**所有命令必须使用 `rtk` 前缀**

所有 Bash/Shell 命令必须以 `rtk` 开头，禁止直接执行未使用 `rtk` 的命令。

```bash
# ✅ 正确
rtk git status
rtk go build ./...
rtk curl -s http://localhost:8080/health

# ❌ 错误
git status
go build ./...
curl -s http://localhost:8080/health
```

**原因**: RTK hook 仅自动重写已知命令，其他命令可能不被拦截导致token浪费。

---

For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
specs/019-backend-refactor-22fixes/plan.md

Current feature scope: Backend Refactor — Fix 22 Issues across 7 phases.
TDD-driven (Red-Green-Refactor). Chroma collection resolution, data extraction,
AI integration, hardcoded value removal, schema completion, test cleanup, dead code removal.

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

## Protobuf Configuration

- `PROTOBUF_ENABLED` — "none" (dev), "internal", "websocket", "all" (prod)
- `QUEUE_PROTOBUF_ENABLED` — "true"/"false" for RabbitMQ binary payloads
- Code gen: `./scripts/generate-proto.sh`
- Generated code: `backend/pkg/proto/` (Go), `frontend/src/lib/proto/` (TypeScript)
<!-- SPECKIT END -->
