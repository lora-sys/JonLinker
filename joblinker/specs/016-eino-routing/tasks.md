# Tasks: Eino Agent Routing Layer

## Summary

- **Total Tasks**: 18
- **User Stories**: 5 (US1-US5)
- **Parallel Opportunities**: 2 groups (middleware + route files)

---

## Phase 1: Router Infrastructure (Setup)

- [ ] T001 [P] Create `backend/internal/router/router.go` - Main router entry point
- [ ] T002 [P] Create `backend/internal/router/middleware.go` - Header extraction & validation
- [ ] T003 [P] Create `backend/internal/router/context.go` - Request context utilities
- [ ] T004 Create `backend/internal/router/errors.go` - Route error definitions

---

## Phase 2: Agent Routes (US1 - Agent 专属路由)

- [ ] T005 [US1] Create `backend/internal/router/agent.go` - Agent-specific routes
- [ ] T006 [US1] Implement `/agent/:agentId/chat` handler
- [ ] T007 [US1] Implement `/agent/:agentId/state` handler
- [ ] T008 [US1] Implement `/agent/:agentId/tool/:toolName` handler
- [ ] T009 [US1] Implement `/agent/:agentId/memory` handler

---

## Phase 3: MCP Tool Routes (US2 - 工具调用路由)

- [ ] T010 [US2] Create `backend/internal/router/tool.go` - Tool routes
- [ ] T011 [US2] Implement `/tool/fetch-job` handler
- [ ] T012 [US2] Implement `/tool/fetch-resume` handler
- [ ] T013 [US2] Implement `/tool/interview-invite` handler
- [ ] T014 [US2] Implement `/tool/generate-offer` handler
- [ ] T015 [US2] Implement `/tool/match-vector` handler

---

## Phase 4: FSM Routes (US3 - FSM 事件路由)

- [ ] T016 [US3] Create `backend/internal/router/fsm.go` - FSM routes
- [ ] T017 [US3] Implement `/fsm/transition` handler
- [ ] T018 [US3] Implement `/fsm/state-sync` handler
- [ ] T019 [US3] Implement `/fsm/event` handler

---

## Phase 5: Memory/Vector Routes (US4 - 记忆向量路由)

- [ ] T020 [US4] Create `backend/internal/router/memory.go` - Memory routes
- [ ] T021 [US4] Create `backend/internal/router/vector.go` - Vector routes
- [ ] T022 [US4] Implement `/memory/save` handler
- [ ] T023 [US4] Implement `/memory/retrieve` handler
- [ ] T024 [US4] Implement `/vector/search` handler

---

## Phase 6: Integration (US5 - 路由注册)

- [ ] T025 [US5] Update `backend/cmd/server/main.go` - Register router (additive, not replacing)
- [ ] T026 [US5] Test router registration and basic routing

---

## Phase 7: Verification

- [ ] T027 [P] Isolation test: Agent ID mismatch returns 403
- [ ] T028 [P] Header propagation test: X-User-ID, X-Agent-ID, X-Tenant-ID present in context
- [ ] T029 Regression test: WebSocket, RabbitMQ, Protobuf unchanged

---

## Dependency Graph

```
Phase 1 (T001-T004) - Router Infrastructure
    │
    ├──────────────────────────────────────────┐
    ▼                                        ▼
Phase 2 (T005-T009)                    Phase 3 (T010-T015)
    Agent Routes                        Tool Routes
    │                                        │
    └────────────────────────────────────────┴──────────┐
                                                        │
                                                        ▼
                                            Phase 4 (T016-T019)
                                                FSM Routes
                                                        │
                                                        ▼
                                            Phase 5 (T020-T024)
                                            Memory/Vector Routes
                                                        │
                                                        ▼
                                            Phase 6 (T025-T026)
                                               Integration
                                                        │
                                                        ▼
                                            Phase 7 (T027-T029)
                                             Verification
```

## Independent Test Criteria

| Story | Test Criteria |
|-------|---------------|
| US1 (Agent Routes) | /agent/:agentId/chat routes to correct handler |
| US2 (Tool Routes) | /tool/* routes invoke correct tool handlers |
| US3 (FSM Routes) | /fsm/* routes handle state transitions |
| US4 (Memory/Vector) | /memory/* and /vector/* routes work |
| US5 (Integration) | Router registers without breaking existing routes |

## Agent Isolation Verification

### Test Case 1: Header vs Route Mismatch
```
Request: POST /agent/agent-1/chat
Header: X-Agent-ID: agent-2

Expected: 403 Forbidden
Actual: Should reject
```

### Test Case 2: Correct Agent Access
```
Request: POST /agent/agent-1/chat
Header: X-Agent-ID: agent-1

Expected: 200 OK, route to handler
Actual: Should allow
```

## Implementation Notes

- Router is **additive only** - does not replace existing handlers
- Existing WebSocket, RabbitMQ, Protobuf code paths unchanged
- New router wraps existing logic via middleware
- All routes extract and validate X-User-ID, X-Agent-ID, X-Tenant-ID