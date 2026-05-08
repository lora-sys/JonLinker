# Tasks: Backend Refactor — Fix 22 Issues

**Input**: Design documents from `/specs/019-backend-refactor-22fixes/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: TDD mandatory. Every functional change: write failing test (RED) → implement (GREEN) → refactor (IMPROVE). All tests use real services (DB, Chroma, AI, RabbitMQ). No mocks. No `t.Skip`. No hardcoded assertions.

**Organization**: Tasks grouped by user story. Each story independently testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Ensure compilation and initialize Chroma collection resolution

- [ ] T001 Verify `go build ./...` passes — run build, confirm zero errors in `joblinker/backend/`
- [ ] T002 [P] Write RED test for Chroma EnsureCollection in `joblinker/backend/pkg/chroma/client_test.go` — test that `EnsureCollection("test_collection")` returns a non-empty UUID and subsequent calls return the same UUID
- [ ] T003 Implement `collections` map and `EnsureCollection(name string) (string, error)` method in `joblinker/backend/pkg/chroma/client.go` — cache name→UUID, call `GetOrCreateCollection` on first access (GREEN for T002)
- [ ] T004 Modify `Add`, `Query`, `Delete`, `Get` methods in `joblinker/backend/pkg/chroma/client.go` to call `EnsureCollection` internally before making API calls
- [ ] T005 Add `EnsureCollection("agent_memories")` and `EnsureCollection("user_preferences")` calls in `joblinker/backend/cmd/server/main.go` after Chroma heartbeat success
- [ ] T006 Write integration test for Chroma collection persistence in `joblinker/backend/tests/integration/chroma_real_test.go` — store a document via `Add`, retrieve via `Query`, verify non-empty results. Uses `t.Fatal` if Chroma unreachable.

**Checkpoint**: Chroma collection resolution works. Build passes. `grep -n "EnsureCollection" pkg/chroma/client.go` returns matches.

---

## Phase 2: User Story 1 — Data Extraction Returns Real Values (Priority: P1)

**Goal**: `extractLocationValue`, `extractSkillsValue`, `extractSalaryValue` return real extracted data, never placeholder strings.

**Independent Test**: Call each extractor with sample text containing known values. Verify output contains the known values. Call with garbage text. Verify output is a substring of input (not "extracted_from_conversation").

### Tests (RED phase) ⚠️

> Write these tests FIRST. Verify they FAIL before implementation.

- [ ] T007 [P] [US2] Write unit tests for `extractLocationValue` in `joblinker/backend/internal/service/conversation_service_test.go` — test cases: "remote work in San Francisco" → contains "remote" or "San Francisco"; "onsite in NYC" → contains "onsite" or "NYC"; "hybrid position" → contains "hybrid"; random gibberish → NOT "extracted_from_conversation"
- [ ] T008 [P] [US2] Write unit tests for `extractSkillsValue` in `joblinker/backend/internal/service/conversation_service_test.go` — test cases: "5 years of Go and Python" → contains "Go" and "Python"; "React, Node.js, Docker" → contains "React", "Node", "Docker"; random text → NOT "extracted_from_conversation"
- [ ] T009 [P] [US2] Write unit tests for `extractSalaryValue` fallback in `joblinker/backend/internal/service/conversation_service_test.go` — test case: text with no salary info → NOT "extracted_from_conversation", returns substring of input

### Implementation (GREEN phase)

- [ ] T010 [US2] Implement `extractLocationValue` in `joblinker/backend/internal/service/conversation_service.go` — keyword matching for remote/onsite/hybrid + top 20 tech cities; fallback returns first 50 chars of content
- [ ] T011 [US2] Implement `extractSkillsValue` in `joblinker/backend/internal/service/conversation_service.go` — match 50+ common technical skills (Go, Python, Java, JavaScript, TypeScript, React, Node, Docker, Kubernetes, AWS, etc.); fallback returns first 100 chars
- [ ] T012 [US2] Fix `extractSalaryValue` fallback in `joblinker/backend/internal/service/conversation_service.go` line 265 — change from `"extracted_from_conversation"` to first 100 chars of content

### Refactor (IMPROVE phase)

- [ ] T013 [US2] Run `go test ./internal/service/ -run TestExtract -v` — all tests GREEN. Refactor keyword lists into package-level constants if needed.

**Checkpoint**: `grep -c "extracted_from_conversation" internal/service/conversation_service.go` returns 0.

---

## Phase 3: User Story 2 — AI Agents Generate Real Responses (Priority: P1)

**Goal**: `generateSeekerProposal`, `understandPhase`, `respondPhase` call the real AI service. Fallbacks are generic messages, not templates with round numbers.

**Independent Test**: Call `generateSeekerProposal` with conversation history. Verify response does NOT contain "Seeker proposal in round". Call with AI service down. Verify graceful fallback.

### Tests (RED phase) ⚠️

- [ ] T014 [P] [US3] Write RED test for `generateSeekerProposal` in `joblinker/backend/internal/service/dual_agent_negotiation_service_test.go` — verify response does NOT contain "Seeker proposal in round"; verify response is non-empty; if AI available, verify response length > 20 chars
- [ ] T015 [P] [US3] Write RED test for `understandPhase` AI integration in `joblinker/backend/internal/service/reasoning_service_test.go` — verify that with `aiClient` set, intent classification returns a valid intent (not just "General inquiry" for every input)
- [ ] T016 [P] [US3] Write RED test for `respondPhase` fallback in `joblinker/backend/internal/service/reasoning_service_test.go` — verify response does NOT contain "Based on my analysis:"; verify fallback is "I'll review the details and get back to you." when AI unavailable

### Implementation (GREEN phase)

- [ ] T017 [US3] Add `aiClient *ai.Client` field to `DualAgentNegotiationService` struct and update `NewDualAgentNegotiationService` constructor in `joblinker/backend/internal/service/dual_agent_negotiation_service.go`
- [ ] T018 [US3] Rewrite `generateSeekerProposal` in `joblinker/backend/internal/service/dual_agent_negotiation_service.go` — call `s.aiClient.Chat(systemPrompt, prompt)` with conversation context; fallback: "I'd like to discuss the compensation package further."
- [ ] T019 [US3] Add AI call to `understandPhase` in `joblinker/backend/internal/service/reasoning_service.go` — if `s.aiClient != nil`, call AI for intent classification; keep keyword matching as fallback
- [ ] T020 [US3] Fix `respondPhase` fallback in `joblinker/backend/internal/service/reasoning_service.go` — replace "Based on my analysis:" template with `s.aiClient.Chat()` call; fallback: "I'll review the details and get back to you."

### Refactor (IMPROVE phase)

- [ ] T021 [US3] Run `go test ./internal/service/ -run TestDualAgent -v` and `go test ./internal/service/ -run TestReasoning -v` — all GREEN. Verify no "Seeker proposal in round" or "Based on my analysis" in source.

**Checkpoint**: `grep -c "Seeker proposal in round" internal/service/dual_agent_negotiation_service.go` returns 0. `grep -c "Based on my analysis" internal/service/reasoning_service.go` returns 0.

---

## Phase 4: User Story 3 — No Hardcoded Business Values (Priority: P1)

**Goal**: `fallbackResponse` contains no salary magic numbers. Salary resolution returns error instead of 100000.

**Independent Test**: Call `fallbackResponse("INTEREST")` — verify no 100000/140000 in output. Trigger salary fallback with no job data — verify error returned, not 100000.

### Tests (RED phase) ⚠️

- [ ] T022 [P] [US4] Write RED test for `fallbackResponse` in `joblinker/backend/internal/service/message_queue_service_test.go` — for each intent (INTRODUCTION, INTEREST, NEGOTIATION, OFFER, SCHEDULE), verify response contains NO hardcoded values (100000, 140000, 110000, 120000)
- [ ] T023 [P] [US4] Write RED test for salary fallback in `joblinker/backend/internal/service/message_queue_service_test.go` — when job has no salary data and no structured JSON, verify the system returns an error or uses match score, NOT 100000

### Implementation (GREEN phase)

- [ ] T024 [US4] Rewrite `fallbackResponse` in `joblinker/backend/internal/service/message_queue_service.go` lines 615-631 — INTEREST returns text-only message "Let's discuss compensation details."; NEGOTIATION returns "We'd like to extend an offer."; remove all salary numbers
- [ ] T025 [US4] Fix salary fallback in `joblinker/backend/internal/service/message_queue_service.go` lines 481-483 — replace `salary = 100000` with `return nil, fmt.Errorf("salary could not be determined from job data")`

### Refactor (IMPROVE phase)

- [ ] T026 [US4] Run `go test ./internal/service/ -run TestFallback -v` — all GREEN. Run `grep -n "100000\|140000\|110000" internal/service/message_queue_service.go | grep -v MaxTokens` — returns 0 matches.

**Checkpoint**: Zero hardcoded salary values in message_queue_service.go (excluding MaxTokens constants).

---

## Phase 5: User Story 4 — Database Schema Is Complete (Priority: P2)

**Goal**: AutoMigrate includes all required models. Fresh database startup creates all tables.

**Independent Test**: Start server against fresh DB. Query `information_schema.tables` for ConversationSummary, AgentToolCall, ConfirmationRequest. All must exist.

### Tests (RED phase) ⚠️

- [ ] T027 Write integration test for schema completeness in `joblinker/backend/tests/integration/schema_real_test.go` — connect to test DB, query `information_schema.tables`, verify tables exist: conversation_summaries, agent_tool_calls, confirmation_requests. Uses `t.Fatal` if DB unreachable.

### Implementation (GREEN phase)

- [ ] T028 Add missing models to AutoMigrate in `joblinker/backend/cmd/server/main.go` lines 78-96 — add `&model.ConversationSummary{}`, `&model.AgentToolCall{}`, `&model.ConfirmationRequest{}`. Verify models exist in `internal/model/` first; create if missing.
- [ ] T029 Run `go build ./...` — confirm compilation passes after model additions.

**Checkpoint**: `grep -n "AgentToolCall" cmd/server/main.go` returns match. Integration test T027 passes.

---

## Phase 6: User Story 5 — Integration Tests Are Trustworthy (Priority: P2)

**Goal**: All integration tests use `t.Fatal` (not `t.Skip`), generate random data, include anti-hardcoded assertions.

**Independent Test**: Run `grep -c "t.Skip(" tests/integration/*.go` — returns 0. Run `grep -c "CreateComprehensiveSeedData" tests/integration/*.go` — returns 0.

### Tests (RED phase) ⚠️

- [ ] T030 [P] [US6] Write RED test for AI response integration in `joblinker/backend/tests/integration/ai_response_real_test.go` — send message via POST /api/messages/:matchId, poll for AI response, verify response does NOT contain "Thank you for your introduction" or any KnownHardcodedValues. Uses `t.Fatal` if server unreachable.
- [ ] T031 [P] [US6] Write RED test for extraction integration in `joblinker/backend/tests/integration/extraction_real_test.go` — create match, send message with salary/location/skills info, verify stored preferences contain real values (not "extracted_from_conversation"). Uses `t.Fatal` if services unreachable.

### Implementation (GREEN phase)

- [ ] T032 [US6] Review `joblinker/backend/tests/integration/` for any `t.Skip` calls — replace with `t.Fatal`. Check eino unit test `joblinker/backend/internal/eino/agent/seeker_agent_test.go` lines 17, 50 — change `t.Skip` to `t.Fatal` or add build tag.
- [ ] T033 [US6] Review `joblinker/backend/tests/testutil/anti_cheat.go` — verify KnownHardcodedValues includes all strings from spec: "extracted_from_conversation", "Seeker proposal in round", "Based on my analysis:". Add missing entries.
- [ ] T034 [US6] Add comments to `joblinker/backend/tests/unit/agent/tool_executor_test.go` — clarify that nil-repo tests are error-handling tests, not functional tests. Functional tests live in integration/.

### Refactor (IMPROVE phase)

- [ ] T035 [US6] Run `grep -rc "t\.Skip(" tests/integration/ --include="*.go"` — returns 0. Run `grep -rc "CreateComprehensiveSeedData" tests/integration/ --include="*.go"` — returns 0. Run `go test ./... -count=1 -v` — all unit tests pass.

**Checkpoint**: Zero `t.Skip` in integration tests. All anti-hardcoded assertions in place.

---

## Phase 7: User Story 6 — Dead Code Is Removed (Priority: P3)

**Goal**: Delete dead files, remove dead functions, fix route conflict, fix Protobuf timestamp.

**Independent Test**: `go build ./...` passes. Deleted files don't exist. Route conflict resolved.

### Implementation (no RED phase — deletion/cleanup tasks)

- [ ] T036 Verify no imports of `internal/agent/memory.go` — run `grep -rn "joblinker/internal/agent" --include="*.go" | grep -v "_test.go" | grep "memory"` in `joblinker/backend/`. If no imports found, delete `joblinker/backend/internal/agent/memory.go`.
- [ ] T037 Verify no callers of `GenerateEmbedding` — run `grep -rn "GenerateEmbedding" --include="*.go"` in `joblinker/backend/`. If no callers outside the function itself, delete `GenerateEmbedding`, `EmbeddingRequest`, `EmbeddingResponse`, `EmbeddingData` from `joblinker/backend/pkg/ai/client.go` (lines 327-391).
- [ ] T038 Remove API key length log in `joblinker/backend/internal/service/message_queue_service.go` line 91 — delete `log.Printf("MessageQueueService AI client - BaseURL: %s, APIKey length: %d", ...)`
- [ ] T039 Fix Protobuf frame timestamp in `joblinker/backend/internal/handler/message.go` — change `Timestamp: 0` to `Timestamp: time.Now().UnixMilli()` in the WebSocketFrame construction
- [ ] T040 Fix route conflict in `joblinker/backend/cmd/server/main.go` — change `api.GET("/interviews/:matchId", interviewHandler.GetByMatchID)` to `api.GET("/interviews/match/:matchId", interviewHandler.GetByMatchID)`. Update any frontend callers if needed.
- [ ] T041 Run `go build ./...` — confirm all deletions and changes compile. Run `go vet ./...` — no issues.

**Checkpoint**: `test -f internal/agent/memory.go && echo "FAIL" || echo "PASS"` returns PASS. `grep -n "GenerateEmbedding" pkg/ai/client.go` returns 0. `grep -n "APIKey length" internal/service/message_queue_service.go` returns 0.

---

## Phase 8: Final Validation

**Purpose**: End-to-end verification of all 10 success criteria

- [ ] T042 Run full verification script from `specs/019-backend-refactor-22fixes/quickstart.md` — all 10 checks must print PASS
- [ ] T043 Run `go test ./... -count=1 -v` — all unit tests pass, no SKIP
- [ ] T044 Run `go test ./... -tags=integration -count=1 -v` — all integration tests pass (requires live services: DB, Chroma, RabbitMQ, AI)
- [ ] T045 Run `go vet ./...` and `go build ./...` — zero errors

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — starts immediately
- **Phase 2 (US2 - Extraction)**: Depends on Phase 1 (build must pass)
- **Phase 3 (US3 - AI)**: Depends on Phase 1; independent of Phase 2
- **Phase 4 (US4 - Hardcoded)**: Depends on Phase 1; independent of Phases 2-3
- **Phase 5 (US5 - Schema)**: Depends on Phase 1; independent of Phases 2-4
- **Phase 6 (US6 - Tests)**: Depends on Phases 2-4 (tests verify the fixes)
- **Phase 7 (US7 - Dead Code)**: Depends on Phase 1; can run in parallel with Phases 2-6
- **Phase 8 (Validation)**: Depends on all previous phases

### User Story Dependencies

- **US2 (Extraction)**: Independent — can start after Phase 1
- **US3 (AI)**: Independent — can start after Phase 1
- **US4 (Hardcoded)**: Independent — can start after Phase 1
- **US5 (Schema)**: Independent — can start after Phase 1
- **US6 (Tests)**: Depends on US2, US3, US4 (tests verify those fixes)
- **US7 (Dead Code)**: Independent — can start after Phase 1

### Within Each User Story

- Tests (RED) MUST be written and FAIL before implementation
- Implementation (GREEN) makes tests pass
- Refactor (IMPROVE) cleans up after green

### Parallel Opportunities

- Phases 2, 3, 4, 5, 7 can all run in parallel (different files)
- Within Phase 2: T007, T008, T009 can run in parallel (different test functions)
- Within Phase 3: T014, T015, T016 can run in parallel
- Within Phase 4: T022, T023 can run in parallel
- Within Phase 6: T030, T031 can run in parallel

---

## Parallel Example: Phase 2 (US2 - Extraction)

```bash
# RED phase — launch all 3 test tasks in parallel:
Task T007: "Write unit tests for extractLocationValue"
Task T008: "Write unit tests for extractSkillsValue"
Task T009: "Write unit tests for extractSalaryValue fallback"

# Verify all 3 tests FAIL (RED confirmed)

# GREEN phase — implement sequentially (same file):
Task T010: "Implement extractLocationValue"
Task T011: "Implement extractSkillsValue"
Task T012: "Fix extractSalaryValue fallback"

# IMPROVE phase:
Task T013: "Run tests, refactor keyword lists"
```

## Parallel Example: Phase 3+4+5+7 (Independent Stories)

```bash
# These 4 phases touch different files — can run simultaneously:
Phase 3 (AI):     dual_agent_negotiation_service.go, reasoning_service.go
Phase 4 (Values): message_queue_service.go
Phase 5 (Schema): cmd/server/main.go
Phase 7 (Dead):   agent/memory.go, ai/client.go, handler/message.go
```

---

## Implementation Strategy

### MVP First (Phases 1-2)

1. Complete Phase 1: Chroma collection resolution
2. Complete Phase 2: Data extraction fixes
3. **STOP and VALIDATE**: Run extraction tests, verify no "extracted_from_conversation"

### Incremental Delivery

1. Phase 1 → Chroma works
2. Phase 2 → Extraction works (MVP!)
3. Phase 3 → AI responses work
4. Phase 4 → No hardcoded values
5. Phase 5 → Schema complete
6. Phase 6 → Tests trustworthy
7. Phase 7 → Dead code removed
8. Phase 8 → Full validation

### TDD Cycle Per Task

```
For each task:
  1. Write test that describes desired behavior
  2. Run test — verify it FAILS (RED)
  3. Implement minimum code to pass
  4. Run test — verify it PASSES (GREEN)
  5. Refactor for clarity
  6. Run test — verify still GREEN (IMPROVE)
  7. Commit
```

---

## Notes

- All integration tests require live services (PostgreSQL, Chroma, RabbitMQ, AI API)
- `t.Fatal` for missing dependencies — never `t.Skip`
- Random data for all test inputs (random emails, random UUIDs)
- Anti-hardcoded assertions in every integration test
- Commit after each task or logical group
- Stop at any checkpoint to validate independently
