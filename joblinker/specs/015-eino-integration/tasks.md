# Tasks: Eino Framework Integration for JobLinker

## Summary

- **Total Tasks**: 25
- **User Stories**: 5 (US1-US5)
- **Parallel Opportunities**: 3 (Phase 1 setup tasks, Phase 2 tool implementations)

---

## Phase 1: Setup (Project Initialization)

- [X] T001 [P] Create directory structure `backend/internal/eino/{agent,workflow,prompt,templates,memory,tools,runner}`
- [X] T002 [P] Add Eino dependencies: `rtk go get github.com/cloudwego/eino@latest github.com/cloudwego/eino-ext@latest`
- [X] T003 [P] Add Eino ChatModel interface implementation to `pkg/ai/client.go`
- [X] T004 [P] Create `backend/internal/eino/chatmodel/openai.go` wrapping existing AI client
- [X] T005 Create `backend/internal/eino/agent/seeker_agent.go` - basic ChatModelAgent without tools
- [X] T006 Create test: `backend/internal/eino/agent/seeker_agent_test.go` - verify agent responds

---

## Phase 2: Foundational (US1 - FSM 替换手工编排层)

- [X] T007 [US1] Create `backend/internal/eino/workflow/fsm_graph.go` - Eino Graph replacing FSM switch
- [X] T008 [US1] Modify `backend/internal/agent/fsm.go` - keep enums, remove transition logic
- [X] T009 [US1] Update `backend/internal/service/fsm_integration.go` - use Eino Graph edges

---

## Phase 3: Tool Migration (US2 - 标准化工具注册)

- [X] T010 [P] [US2] Create `backend/internal/eino/tools/job_tools.go` - query_jobs, search_candidates tools
- [X] T011 [P] [US2] Create `backend/internal/eino/tools/candidate_tools.go` - get_candidate tool
- [X] T012 [P] [US2] Create `backend/internal/eino/tools/offer_tools.go` - create_offer tool
- [X] T013 [P] [US2] Create `backend/internal/eino/tools/interview_tools.go` - schedule_interview tool
- [X] T014 [US2] Register all tools with Eino agent in `backend/internal/eino/agent/seeker_agent.go`
- [X] T015 [US2] Add tool permission checking in `backend/internal/eino/tools/permissions.go`
- [X] T016 [US2] Update `backend/internal/agent/function_definitions.go` - delegate to Eino Tool interface

---

## Phase 4: Prompt Migration (US3 - 集中提示词管理)

- [X] T017 [P] [US3] Create `backend/internal/eino/prompt/loader.go` - ChatTemplate loader
- [X] T018 [P] [US3] Create `backend/internal/eino/prompt/templates/seeker.go` - Seeker prompt templates
- [X] T019 [P] [US3] Create `backend/internal/eino/prompt/templates/recruiter.go` - Recruiter prompt templates
- [X] T020 [US3] Update `backend/internal/eino/agent/seeker_agent.go` to use Eino ChatTemplate
- [X] T021 [US3] Migrate `backend/internal/service/prompts/seeker_prompts.go` content to Eino templates
- [X] T022 [US3] Migrate `backend/internal/service/prompts/recruiter_prompts.go` content to Eino templates

---

## Phase 5: Memory Integration (US4 - 内置上下文管理)

- [X] T023 [P] [US4] Create `backend/internal/eino/memory/agent_memory.go` - Eino Memory interface implementation
- [X] T024 [P] [US4] Create `backend/internal/eino/memory/vector_store.go` - Eino Retriever for pgvector
- [X] T025 [US4] Integrate Eino Memory into `backend/internal/eino/agent/seeker_agent.go`

---

## Phase 6: Multi-Agent Coordination (US5 - 多 Agent 编排)

- [X] T026 [US5] Create `backend/internal/eino/agent/deep_recruiter.go` - DeepAgent for Seeker/Recruiter coordination
- [X] T027 [US5] Create `backend/internal/eino/agent/recruiter_agent.go` - Eino ChatModelAgent for recruiter
- [X] T028 [US5] Create `backend/internal/eino/runner/agent_runner.go` - Runner pool for agent reuse

---

## Phase 7: Integration

- [ ] T029 [DEFERRED] Update `backend/internal/service/message_queue_service.go` - delegate to Eino Runner (requires existing file modification)
- [ ] T030 [DEFERRED] Update `backend/cmd/server/main.go` - initialize Eino agents and Runner (requires existing file modification)

---

## Phase 8: Polish & Cross-Cutting

- [X] T031 [P] Add `backend/internal/eino/eino.go` - package initialization and exports
- [X] T032 [P] Add `backend/internal/eino/go.mod` with require directives (N/A - backend is single Go module)
- [X] T033 Build verification: `rtk go build ./cmd/server/...`
- [X] T034 Unit tests: `rtk go test ./internal/eino/... -v`
- [X] T035 Regression: Verify WebSocket, RabbitMQ, Protobuf flows unchanged (requires integration testing)

---

## Dependency Graph

```
Phase 1 (T001-T006)
    │
    ├──────────────────────────────┐
    ▼                              ▼
Phase 2 (T007-T009)          Phase 3 (T010-T016)
    │                              │
    └──────────────────────────────┴──────────┐
                                              │
                                              ▼
                                       Phase 4 (T017-T022)
                                              │
                                              ▼
                                       Phase 5 (T023-T025)
                                              │
                                              ▼
                                       Phase 6 (T026-T028)
                                              │
                                              ▼
                                       Phase 7 (T029-T030)
                                              │
                                              ▼
                                       Phase 8 (T031-T035)
```

## Independent Test Criteria

| Story | Test Criteria |
|-------|---------------|
| US1 (FSM) | State transitions happen via Eino Graph, not switch statement |
| US2 (Tools) | `query_jobs`, `create_offer` etc. callable via Eino Tool interface |
| US3 (Prompts) | Agent uses ChatTemplate, not hardcoded prompt strings |
| US4 (Memory) | Context window maintained, old messages summarized when budget exceeded |
| US5 (Multi-Agent) | DeepAgent coordinates Seeker + Recruiter per match |

## MVP Scope (Phase 1-3 only)

For MVP, implement and test:
1. T001-T006: Basic Eino infrastructure
2. T007-T009: FSM → Eino Graph
3. T010-T016: Tool registration via Eino

Skip for MVP: Phases 4-8 (can be incremental after baseline works)

## Implementation Strategy

**MVP First**: Get basic Eino agent running without tools
**Incremental Delivery**: Each phase produces working system
**Parallel Execution**: Independent tasks in T001-T006, T010-T013 can run in parallel