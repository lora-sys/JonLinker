# Plan: JobLinker Backend Restructuring (6 Phases)

## Overview

Backend restructuring from mock/stub data layer to production-ready system with real database queries, AI-powered tool execution, and proper Eino integration.

## Problem Statement

**P0 Core Issues** (from dogfood testing):

| Issue | Location | Impact |
|-------|----------|--------|
| `executeQueryJobs` returns hardcoded mock | `tool_executor.go:221-231` | TB-7.1 fails |
| `executeGetCandidate` returns fake candidate | `tool_executor.go:248-259` | TB-7.2 fails |
| `executeSearchCandidates` is vector search stub | `tool_executor.go:404-412` | TB-7.3 fails |
| `recordToolCall` logs only, no persistence | `tool_executor.go:423-445` | TB-11.1 fails |
| `NewToolExecutor` receives nil matchRepo | `main.go:60` | crash on tool call |

---

## Phase 1: Security + Basic Cleanup

**Independence**: Can proceed without any other phase.

| Task | Description |
|------|-------------|
| T101 | JWT Secret single source, remove hardcoded default |
| T102 | Admin endpoint role authorization middleware |
| T103 | Register/Login filter PasswordHash from responses |
| T104 | Handle token signing errors in `generateToken` |
| T105 | Mount RateLimit and ErrorHandler middleware |
| T106 | Move Refresh endpoint to auth route group |
| T107 | Add `sync.RWMutex` to `conversationRounds` |
| T108 | Fix route conflicts in main.go |
| T109 | Implement real DeleteUserAccount |
| T110 | Fix ExportUserData to not return plaintext email |
| T111 | WebSocket CheckOrigin restricts origins |
| T112 | Remove DEBUG logs with sensitive info |

**Deliverable**: Backend passes TB-12 (Security Baseline)

---

## Phase 2: Eino Basic Integration

**Dependency**: Phase 0 (P0-Core) must complete first.

| Task | Description |
|------|-------------|
| T201 | Add Eino dependencies to go.mod |
| T202 | Create Eino ChatModel wrapper for LongCat API |
| T203 | Replace ai.Client.Chat() with Eino ChatModel |
| T204 | Replace tool_executor switch/case with Eino ToolsNode |
| T205 | Replace FSM编排 with Eino Graph |

**Deliverable**: Backend uses Eino for AI orchestration, passes TB-6 (AI Real Response)

---

## Phase 3: Tool System Realization

**Dependency**: Phase 2 complete.

| Task | Description |
|------|-------------|
| T301 | `executeQueryJobs` calls real jobRepo |
| T302 | `executeGetCandidate` calls real agentRepo |
| T303 | `executeSearchCandidates` real DB + vector query |
| T304 | `executeCreateOffer` reads salary from AI response |
| T305 | `generateAutoResponse` uses AI + ToolsNode |
| T306 | Remove hardcoded Offer/Interview values |

**Deliverable**: All tools return real DB data, passes TB-7 (Tool Calls Real Data)

---

## Phase 4: Memory System Persistence

**Dependency**: Can run parallel to Phase 3.

| Task | Description |
|------|-------------|
| T401 | Implement StoreMemory real DB persistence |
| T402 | Implement GetRecentMemories real DB query |
| T403 | Implement SearchSimilar with pgvector cosine similarity |
| T404 | VectorService.GenerateEmbedding uses AI API |
| T405 | Implement PreferenceExtractionService real storage |
| T406 | Store conversation summaries |

**Deliverable**: Memory system persists to DB, passes TB-11 (Memory System Persistence)

---

## Phase 5: Context Management

**Dependency**: Phase 4 complete.

| Task | Description |
|------|-------------|
| T501 | buildAgentContext uses ContextOptimizerService |
| T502 | ConversationService compression/summarization |
| T503 | Implement `extractSalaryValue` regex extraction |
| T504 | ReasoningService.respondPhase calls AI |

**Deliverable**: AI context includes memory + compressed history, passes TB-9.5 (Rounds ≤ 10)

---

## Phase 6: Message Flow + WebSocket

**Dependency**: Phase 3 + Phase 5 complete.

| Task | Description |
|------|-------------|
| T601 | processMessage AI + RabbitMQ (remove hardcoded switch/case) |
| T602 | broadcastToMatch filters by matchID |
| T603 | Complete IntentToEvent missing mappings |
| T604 | FSM state history persistence |
| T605 | WebSocket Protobuf support |
| T606 | Fix response timestamps (no hardcoded time) |

**Deliverable**: WebSocket broadcasts correctly, AI-driven intent, passes TB-10 (WebSocket Real-Time)

---

## Test Benchmark Summary

| Benchmark | Tests | Priority | Blocked By |
|-----------|-------|----------|------------|
| TB-1: Infrastructure | 4 | P1 | Phase 0 |
| TB-2: Auth | 8 | P1 | T101, T103, T104 |
| TB-3: Agent CRUD | 7 | P2 | TB-2 |
| TB-4: Job CRUD | 5 | P2 | TB-3 |
| TB-5: Match | 5 | P2 | TB-4 |
| TB-6: AI Response | 6 | **P0** | T000.1, T000.2, T201, T203, T205, T305 |
| TB-7: Tool Calls | 8 | **P0** | T000.1–T000.5, T301–T306 |
| TB-8: FSM | 10 | P1 | T603, T604 |
| TB-9: A2A Dialogue | 10 | **P0** | TB-6, TB-7, TB-8, T107, T502, T601 |
| TB-10: WebSocket | 8 | P1 | TB-5, T111, T601, T602, T605, T606 |
| TB-11: Memory | 6 | P1 | T401–T406 |
| TB-12: Security | 8 | P1 | T101, T102, T105, T106, T109, T110, T111 |

**Total: 81 tests**

---

## Key Architecture Decisions

### Eino Integration Strategy
- Use Eino ChatModel wrapper around existing LongCat API (no API change)
- Use Eino ToolsNode for tool dispatch (replaces switch/case)
- Use Eino Graph for FSM orchestration (replaces hardcoded编排)

### Database Strategy
- All repositories must implement real queries (no mocks)
- Tool calls persisted to `agent_tool_calls` table
- Memory persisted to `agent_memories` table with pgvector embeddings
- Conversation summaries persisted to `conversation_summaries` table

### AI Response Strategy
- All responses from real AI API (no fallback/hardcoded responses)
- If AI API unavailable, test FAILS (not skip)
- Embeddings generated via AI API, not zero vectors

---

## Verification

Each phase produces verifiable test results:
- Phase 0: `go test ./tests/... -run TB-7 -v`
- Phase 1: `go test ./tests/... -run TB-12 -v`
- Phase 2: `go test ./tests/... -run TB-6 -v`
- Phase 3: `go test ./tests/... -run TB-7 -v`
- Phase 4: `go test ./tests/... -run TB-11 -v`
- Phase 5: `go test ./tests/... -run TB-9.5 -v`
- Phase 6: `go test ./tests/... -run TB-10 -v`

Final: `go test ./tests/...` (all 81 tests must pass)
