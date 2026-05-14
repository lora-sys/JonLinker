<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:

specs/020-production-hardening/plan.md

Current feature scope: Production Hardening — 5 phases.
Security fix (JWT secret, httpOnly cookie, multi-tenant isolation),
A2A dual-agent dialogue fix, Eino tool real data connection,
WebSocket route split, MQ Service decomposition, observability,
Playwright CLI E2E acceptance testing.

## Key Design Decisions

1. Multi-tenant: tenant_id on all tables, enforced in repository layer
2. A2A: bidirectional loop with structured JSON protocol detection
3. Eino: keep but connect to real repositories (no mock data)
4. WebSocket: /api/messages/:matchId/ws for humans, /api/a2a/:matchId/ws for agents
5. Config: centralized in internal/config/config.go, remove scattered os.Getenv
6. State machine: single source of truth in Match.Status
7. Rate limiting: in-memory sliding window, IP + userID dual key
<!-- SPECKIT END -->
