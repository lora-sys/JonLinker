# API Gateway Tasks

## Phase 1: Setup (Gateway Core Infrastructure)

- [x] T001 [P] Create gateway directory structure `frontend/src/lib/gateway/`
- [x] T002 [P] Define route mapping configuration in `routes.ts`
- [x] T003 [P] Implement GatewayClient class in `client.ts`
- [x] T004 Create gateway middleware in `middleware.ts`

## Phase 2: Frontend Integration (Replace Direct API Calls)

- [x] T005 [P] Replace `frontend/src/lib/api_client.ts` to use GatewayClient
- [x] T006 [P] Replace `frontend/src/stores/auth.ts` to use GatewayClient
- [x] T007 [P] Replace `frontend/src/app/agents/page.tsx` API calls (uses apiClient → gateway headers)
- [x] T008 [P] Replace `frontend/src/app/jobs/page.tsx` API calls (uses apiClient → gateway headers)
- [x] T009 [P] Replace `frontend/src/app/matches/page.tsx` API calls (uses apiClient → gateway headers)
- [x] T010 [P] Replace `frontend/src/app/interviews/page.tsx` API calls (uses apiClient → gateway headers)
- [x] T011 [P] Replace `frontend/src/app/offers/page.tsx` API calls (uses apiClient → gateway headers)
- [x] T012 [P] Replace `frontend/src/app/messages/page.tsx` API calls (uses apiClient → gateway headers)
- [x] T013 [P] Replace `frontend/src/app/settings/page.tsx` API calls (uses apiClient → gateway headers)

## Phase 3: Backend Agent Integration (Agent Tool Calls Through Gateway)

- [x] T014 [P] Review `backend/internal/agent/tool_executor.go` (uses direct repo calls, no HTTP needed)
- [x] T015 [P] Review `backend/internal/service/dual_agent_negotiation_service.go` (uses services, not HTTP)

## Phase 4: WebSocket Integration

- [x] T016 [P] Replace `frontend/src/hooks/useWebSocket.ts` to route through gateway WebSocket with token query param

## Phase 5: Extension Points (Optional Polish)

- [x] T017 Add rate limiting hook to GatewayClient
- [x] T018 Add circuit breaker to GatewayClient
- [x] T019 Add request/response logging interceptor

## Phase 6: Verification

- [x] T020 Verify RabbitMQ message queue still works
- [x] T021 Verify WebSocket connections work for A2A dialogue
- [x] T022 Verify Vector DB memory operations work
- [x] T023 Verify dual-agent autonomous dialogue works
- [x] T024 Verify preference vector storage and retrieval works
- [x] T025 Update CLAUDE.md to reference gateway plan

## Summary

All tasks completed. Gateway provides:
- **Route mapping**: ServiceRoutes for all endpoints
- **Header injection**: X-User-ID, X-Agent-ID, X-Tenant-ID
- **Rate limiting**: 100 req/min per client
- **Circuit breaker**: 5 failures triggers open state
- **WebSocket**: Token query param for auth

All existing frontend pages use `apiClient` which now includes gateway headers automatically.

## Next: Phase 12

See `specs/012-production-observability/` for Production Observability & Agent Monitoring features.