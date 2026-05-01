# Tasks: AI Agent Capability Upgrade

**Input**: Design documents from `/specs/010-agent-ai-upgrade/`
**Prerequisites**: plan.md (tech stack), spec.md (10 user stories), data-model.md (entities), contracts/api-contracts.md, research.md (decisions)

**Organization**: Tasks grouped by user story to enable independent implementation and testing

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US10)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Configure project tooling and base structure for AI agent work

- [X] T001 [P] Add pgvector extension check in backend/internal/model/ (verify pgvector available)
- [X] T002 [P] Review existing RabbitMQ consumer in backend/internal/service/message_queue_service.go (session isolation)
- [X] T003 [P] Review existing WebSocket handler in backend/cmd/server/main.go (AI response flow)
- [X] T004 [P] Audit AI API configuration in backend/internal/config/ (AI_API_KEY, AI_BASE_URL, AI_MODEL)

**Checkpoint**: Setup complete - foundational work can begin

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure required before any user story

**CRITICAL**: All user stories depend on Phase 2 completion

- [X] T005 Define AgentPrompt model in backend/internal/model/agent_prompt.go (Role, MustDo, MustNotDo, Behavior)
- [X] T006 [P] Define PromptScenario model in backend/internal/model/prompt_scenario.go (Scenario, TriggerConditions, PromptFragment)
- [X] T007 [P] Define ConversationSummary model in backend/internal/model/conversation_summary.go (SummaryText, KeyFacts, TokenCount)
- [X] T008 [P] Define UserPreferenceVector model in backend/internal/model/user_preference_vector.go (pgvector embedding)
- [X] T009 Define AgentToolCall model in backend/internal/model/agent_tool_call.go (ToolName, Arguments, Result, Status)
- [X] T010 Define AgentMemory model in backend/internal/model/agent_memory.go (MemoryType, Content, Embedding)
- [X] T011 Create function definitions schema in backend/internal/agent/function_definitions.go (5 tools: query_jobs, get_candidate, create_offer, schedule_interview, search_candidates)
- [X] T012 Create tool executor in backend/internal/agent/tool_executor.go (execute API calls, handle errors)
- [X] T013 Define NegotiationSession model in backend/internal/model/negotiation_session.go (AgentId, Status, CurrentRound)
- [X] T014 Define PreferenceExtraction model in backend/internal/model/preference_extraction.go (PreferenceType, RawValue, VectorEmbedding)
- [X] T015 Define ConfirmationRequest model in backend/internal/model/confirmation_request.go (RequestType, Payload, Status)

**Checkpoint**: Foundation ready - all user stories can now proceed in parallel

---

## Phase 3: User Story 1 - Three-Part System Prompt Architecture (Priority: P1)

**Goal**: Agent responses follow three-part prompt structure (must do + must not do + behavior rules)

**Independent Test**: Prompt agent with various scenarios, verify response structure

### Implementation for User Story 1

- [X] T016 [P] [US1] Create AgentPromptService in backend/internal/service/agent_prompt_service.go (generate three-part prompts)
- [X] T017 [P] [US1] Create seeker prompt template in backend/internal/service/prompts/seeker_prompts.go (recruitment context)
- [X] T018 [P] [US1] Create recruiter prompt template in backend/internal/service/prompts/recruiter_prompts.go (candidate evaluation)
- [X] T019 [US1] Implement scenario-specific prompts (greeting, negotiation, salary, interview, offer, decline) in backend/internal/service/prompts/
- [X] T020 [US1] Integrate prompt service into message_queue_service.go (replace existing prompt logic)
- [X] T021 [US1] Add context compression trigger when token count exceeds 80% limit
- [X] T022 [US1] Implement ConversationSummary generation preserving key facts (salary, location, skills)

**Checkpoint**: US1 complete - agent responses follow three-part structure per SC-001

---

## Phase 4: User Story 2 - Function Calling Tool Integration (Priority: P1)

**Goal**: Agent calls APIs before providing information (no fabrication)

**Independent Test**: Ask about jobs/candidates, verify correct API calls made

### Implementation for User Story 2

- [X] T023 [P] [US2] Implement query_jobs tool in backend/internal/agent/tools/job_query_tool.go
- [X] T024 [P] [US2] Implement get_candidate tool in backend/internal/agent/tools/candidate_tool.go
- [X] T025 [P] [US2] Implement create_offer tool in backend/internal/agent/tools/offer_tool.go
- [X] T026 [P] [US2] Implement schedule_interview tool in backend/internal/agent/tools/interview_tool.go
- [X] T027 [P] [US2] Implement search_candidates tool in backend/internal/agent/tools/candidate_search_tool.go
- [X] T028 [US2] Integrate function definitions into message_queue_service.go AI prompt
- [X] T029 [US2] Add FR-011 validation: verify agent never responds without API call for specific facts
- [X] T030 [US2] Test all 5 tools return correct response schemas per contracts/api-contracts.md

**Checkpoint**: US2 complete - zero fabricated information per SC-002

---

## Phase 5: User Story 3 - Short-Term Conversation Memory (Priority: P1)

**Goal**: Agent maintains context across 10+ message conversations

**Independent Test**: Multi-turn conversation, verify context retention

### Implementation for User Story 3

- [X] T031 [P] [US3] Implement ConversationContext tracking in backend/internal/service/conversation_service.go
- [X] T032 [US3] Implement token counting for conversation messages (system + history + tool results)
- [X] T033 [US3] Implement auto-compression at 80% token limit (FR-004)
- [X] T034 [US3] Implement key facts extraction and preservation in compression (FR-005)
- [X] T035 [US3] Test conversation with 20+ messages, verify agent recalls earlier context

**Checkpoint**: US3 complete - context compression reduces tokens 40%+ per SC-003

---

## Phase 6: User Story 4 - Long-Term Vector Memory (Priority: P2)

**Goal**: Agent recalls user preferences across sessions (salary, location, job type)

**Independent Test**: Set preference in session A, recall in session B

### Implementation for User Story 4

- [X] T036 [P] [US4] Create preference_vector repository in backend/internal/repository/preference_vector_repo.go
- [X] T037 [US4] Create AgentMemoryService in backend/internal/service/agent_memory_service.go
- [X] T038 [US4] Implement recall_preferences at session start (FR-015)
- [X] T039 [US4] Implement store_preference when user states preference (FR-014)
- [X] T040 [US4] Add pgvector similarity search for preference recall
- [X] T041 [US4] Test: preference set in session A, recalled in session B (85% accuracy per SC-004)

**Checkpoint**: US4 complete - 85% preference recall accuracy per SC-004

---

## Phase 7: User Story 5 - Long-Chain Reasoning & Task Planning (Priority: P2)

**Goal**: Agent follows understand → query → analyze → respond flow

**Independent Test**: Complex scenario, verify reasoning steps followed

### Implementation for User Story 5

- [X] T042 [P] [US5] Create ReasoningService in backend/internal/service/reasoning_service.go (ReAct loop)
- [X] T043 [US5] Implement understand phase: extract requirements from user message
- [X] T044 [US5] Implement query phase: determine which tools to call
- [X] T045 [US5] Implement analyze phase: score matches against preferences
- [X] T046 [US5] Implement respond phase: generate response with reasoning explanation
- [X] T047 [US5] Add "Let's think step by step" for complex reasoning (Chain-of-Thought)
- [X] T048 [US5] Test: verify 90% of complex scenarios follow understand→query→analyze→respond per SC-005

**Checkpoint**: US5 complete - 90% reasoning flow compliance per SC-005

---

## Phase 8: User Story 6 - Multi-Agent Parallel Scheduling (Priority: P3)

**Goal**: 100 concurrent conversations without cross-contamination

**Independent Test**: Multiple concurrent conversations, verify isolation

### Implementation for User Story 6

- [X] T049 [P] [US6] Add match_id correlation to RabbitMQ message headers
- [X] T050 [US6] Implement session isolation in message consumer (FR-018, FR-019)
- [X] T051 [US6] Implement idempotency check using message_id hash (FR-020)
- [X] T052 [US6] Test 100 concurrent conversations without cross-contamination per SC-006

**Checkpoint**: US6 complete - 100 concurrent agents without leakage per SC-006

---

## Phase 9: User Story 7 - Dual-Agent Autonomous Negotiation (Priority: P1)

**Goal**: 求职Agent and 招聘Agent negotiate autonomously without human intervention

**Independent Test**: Observe complete negotiation flow with no human input mid-process

### Implementation for User Story 7

- [X] T053 [P] [US7] Create DualAgentNegotiationService in backend/internal/service/dual_agent_negotiation_service.go
- [X] T054 [US7] Implement automatic session initiation when match is created (FR-023)
- [X] T055 [US7] Implement agent-to-agent message routing between求职Agent and 招聘Agent
- [X] T056 [US7] Implement autonomous岗位 negotiation between agents
- [X] T057 [US7] Implement autonomous薪资 negotiation between agents
- [X] T058 [US7] Implement autonomous入职条件 negotiation between agents
- [X] T059 [US7] Implement negotiation round tracking and agreement detection
- [X] T060 [US7] Implement negotiation failure/impasse detection (max 20 rounds)
- [X] T061 [US7] Test: complete dual-agent negotiation without any human intervention per SC-010

**Checkpoint**: US7 complete - dual-agent autonomous negotiation per SC-010

---

## Phase 10: User Story 8 - Autonomous Memory Extraction (Priority: P1)

**Goal**: Agent automatically extracts preferences from conversation and stores in vector DB

**Independent Test**: Verify vector store contains auto-extracted preferences after conversations

### Implementation for User Story 8

- [X] T062 [P] [US8] Create PreferenceExtractionService in backend/internal/service/preference_extraction_service.go
- [X] T063 [US8] Implement automatic salary preference extraction from conversation text
- [X] T064 [US8] Implement automatic location preference extraction from conversation text
- [X] T065 [US8] Implement automatic job type preference extraction (remote/hybrid/onsite)
- [X] T066 [US8] Implement automatic vector embedding generation for extracted preferences
- [X] T067 [US8] Implement automatic storage to pgvector without human intervention (FR-026)
- [X] T068 [US8] Implement automatic recall of stored preferences in subsequent sessions (FR-027)
- [X] T069 [US8] Test: verify 90% of explicit preferences auto-extracted and stored per SC-011

**Checkpoint**: US8 complete - 90% preference extraction rate per SC-011

---

## Phase 11: User Story 9 - Autonomous Tool Execution (Priority: P1)

**Goal**: Agent autonomously decides timing and executes tool calls without human confirmation

**Independent Test**: Observe agent executing tools (search, schedule, offer) without prompts

### Implementation for User Story 9

- [X] T070 [P] [US9] Extend tool executor with autonomous decision capability in backend/internal/agent/tool_executor.go
- [X] T071 [US9] Implement autonomous job search execution (no human trigger)
- [X] T072 [US9] Implement autonomous interview scheduling execution (no human trigger)
- [X] T073 [US9] Implement autonomous offer generation execution (no human trigger)
- [X] T074 [US9] Implement autonomous secondary job-person matching execution
- [X] T075 [US9] Implement graceful error handling with autonomous recovery
- [X] T076 [US9] Test: verify agents execute tool calls autonomously without confirmation per SC-013

**Checkpoint**: US9 complete - autonomous tool execution per SC-013

---

## Phase 12: User Story 10 - Human Final Confirmation Only (Priority: P1)

**Goal**: Human receives only final confirmation requests, no mid-negotiation participation

**Independent Test**: Verify human receives zero mid-negotiation messages

### Implementation for User Story 10

- [X] T077 [P] [US10] Create ConfirmationService in backend/internal/service/confirmation_service.go
- [X] T078 [US10] Implement final offer confirmation request (only when agents reach agreement)
- [X] T079 [US10] Implement final interview confirmation request (only when agents agree)
- [X] T080 [US10] Implement human rejection feedback incorporation into agent renegotiation
- [X] T081 [US10] Implement blocking of all mid-negotiation messages from human view
- [X] T082 [US10] Test: verify human receives only final confirmation requests per SC-014

**Checkpoint**: US10 complete - human confirmation boundary per SC-014

---

## Phase 13: Polish & Cross-Cutting Concerns

**Purpose**: E2E validation, integration testing, documentation

- [X] T083 [P] Add E2E test for AI-driven flows using agent-browser (SC-009)
- [X] T084 [P] Verify all 32 functional requirements implemented (FR-001 to FR-032)
- [X] T085 Update SPEC.md with implementation notes
- [X] T086 Verify existing E2E tests still pass (backward compatibility per SC-008)
- [X] T087 Integration test: full autonomous negotiation flow with human confirmation at end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-12)**: All depend on Foundational phase
- **Polish (Phase 13)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Starts after Phase 2 - No dependencies on other stories
- **US2 (P1)**: Starts after Phase 2 - Can run in parallel with US1, US3
- **US3 (P1)**: Starts after Phase 2 - Can run in parallel with US1, US2
- **US4 (P2)**: Starts after Phase 2 - Can run in parallel with all
- **US5 (P2)**: Starts after Phase 2 - Can run in parallel with all
- **US6 (P3)**: Starts after Phase 2 - Can run in parallel with all
- **US7 (P1)**: Starts after Phase 2 - Can run in parallel with US1, US2, US3
- **US8 (P1)**: Starts after Phase 2 - Can run in parallel with all
- **US9 (P1)**: Starts after Phase 2 - Can run in parallel with all
- **US10 (P1)**: Starts after Phase 2 - Can run in parallel with all

### Within Each User Story

- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel
- Once Foundational complete, all 10 user stories can proceed in parallel
- Within each story, [P] tasks can run in parallel

---

## Implementation Strategy

### MVP First (US1 + US2 + US3 + US7 as core P1 group)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: US1 (Three-Part Prompt)
4. Complete Phase 4: US2 (Function Calling)
5. Complete Phase 5: US3 (Short-Term Memory)
6. Complete Phase 9: US7 (Dual-Agent Autonomous Negotiation)
7. **STOP and VALIDATE**: Test core P1 stories together
8. Deploy/demo if ready

### Incremental Delivery

1. Phase 1 + 2 → Foundation ready
2. US1 → Test independently → Deploy
3. US2 → Test independently → Deploy
4. US3 → Test independently → Deploy
5. US7 → Test independently → Deploy (core for autonomous mode)
6. US8 → Test independently → Deploy (memory extraction)
7. US9 → Test independently → Deploy (tool execution)
8. US10 → Test independently → Deploy (human confirmation boundary)
9. US4 → Test independently → Deploy
10. US5 → Test independently → Deploy
11. US6 → Test independently → Deploy
12. Polish → Final validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story independently testable per "Checkpoint" criteria
- MVP scope: US1 + US2 + US3 + US7 (core P1 group for autonomous mode)
- US4 + US5 + US6 are P2/P3 - can ship in second iteration
- US8 + US9 + US10 are P1 for A2A autonomous mode
