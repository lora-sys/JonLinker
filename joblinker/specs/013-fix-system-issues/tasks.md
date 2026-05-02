# Tasks: Fix System Issues - API Gateway, FSM, Vector, A2A

**Feature**: specs/013-fix-system-issues
**Date**: 2026-05-01
**Total Tasks**: 32

## Phase 1: Backend API Gateway Middleware (P0)

- [x] T001 Create backend/internal/middleware/gateway.go with X-User-ID/X-Agent-ID/X-Tenant-ID extraction
- [x] T002 Register gateway middleware in cmd/server/main.go (apply to all routes)
- [x] T003 Add correlation ID (X-Request-ID) generation and logging
- [x] T004 Update message handler to read userID from context for audit logging

## Phase 2: FSM State Machine Integration (P0)

- [x] T005 [P] Create backend/internal/service/fsm_integration.go - maps intents to FSM events
- [x] T006 [P] Update MessageQueueService.handleAgentMessage to trigger FSM transitions on intent
- [x] T007 Persist FSM state changes to match.fsm_state column
- [x] T008 Add FSM event logging for debugging

## Phase 3: Real Vector Embeddings (P0)

- [x] T009 [P] Add Chroma client dependency (github.com/chroma-core/chroma/go/pkg/api or chromadb/chromadb-go-client)
- [x] T010 [P] Add OpenAI embeddings client (github.com/sashabaranov/go-openai or use existing AI client)
- [x] T011 Create backend/internal/service/embedding_service.go - real embedding generation
- [x] T012 Integrate embedding_service into agent_memory_service.go (replace mock generateTextEmbedding)
- [x] T013 Add Chroma storage fallback to pgvector when CHROMA_HOST is configured
- [x] T014 Test embedding generation and similarity search

## Phase 4: Job Form JSON Fix (P1)

- [x] T015 Update frontend job form to JSON.stringify structured field before submit
- [x] T016 Verify backend job handler correctly parses structured JSON string

## Phase 5: Bidirectional A2A Dialogue (P1)

- [x] T017 [P] Modify handleAgentMessage to detect when receiver is an agent
- [x] T018 [P] Add message routing back to AI when receiver is agent (create agent_message for receiver)
- [x] T019 Track conversation round count per match (max 10 to prevent infinite loops)
- [x] T020 Add guard to prevent duplicate message processing (idempotency)

## Phase 6: Verification & Testing

- [x] T021 Write Go test for gateway middleware header extraction
- [x] T022 Write Go test for FSM state transitions
- [x] T023 Write test for embedding generation (mock OpenAI or test with real key)
- [x] T024 Run full system test: Job creation with structured field
- [x] T025 Run full system test: A2A bidirectional dialogue for 3+ rounds
- [x] T026 Verify vector similarity search returns relevant results

## Phase 7: Documentation

- [ ] T027 Update CLAUDE.md with API Gateway middleware usage
- [ ] T028 Update plan.md with FSM state machine documentation
- [ ] T029 Document embedding configuration (CHROMA_HOST, OPENAI_API_KEY)

---

## Dependency Graph

```
T001,T002,T003,T004 (Gateway middleware - Phase 1)
         ↓
T005,T006,T007,T008 (FSM integration - Phase 2, parallel with Phase 3 setup)
         ↓
T009,T010,T011,T012,T013,T014 (Real embeddings - Phase 3)
         ↓
T015,T016 (Job JSON fix - Phase 4, parallel after embeddings)
         ↓
T017,T018,T019,T020 (Bidirectional A2A - Phase 5)
         ↓
T021,T022,T023,T024,T025,T026 (Verification - Phase 6)
         ↓
T027,T028,T029 (Documentation - Phase 7)
```

## Implementation Strategy

**Critical Path (P0 first)**:
1. T001-T004: Gateway middleware (enables audit logging)
2. T005-T008: FSM integration (enables state tracking)
3. T009-T014: Real embeddings (enables memory recall)

**Secondary Path (P1)**:
4. T015-T016: Job form fix
5. T017-T020: Bidirectional A2A

**Verification Last**:
6. T021-T026: Tests
7. T027-T029: Docs