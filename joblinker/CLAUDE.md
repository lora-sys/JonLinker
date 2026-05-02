<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:
specs/014-protobuf-protocol/plan.md

Current feature scope: Protobuf Communication Protocol — binary serialization for
internal services, WebSocket agent dialogue, FSM state changes, and queue tasks.
4-phase rollout with zero-downtime JSON backward compatibility.

Key components: 4 .proto files (agent, tools, websocket, queue), Go + TypeScript
code generation via protoc, content negotiation via Accept header, cache-key
integration with existing tool cache layer.

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
