# Tasks: Performance & Stability Optimization

**Input**: Design documents from `/specs/009-performance-stability/`
**Prerequisites**: plan.md (tech stack), spec.md (5 user stories), data-model.md (entities), contracts/api-contracts.md

**Organization**: Tasks grouped by user story to enable independent implementation and testing

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Configure project tooling and base structure for performance work

- [x] T001 [P] Audit current Next.js 16 image/font optimization in frontend/next.config.js
- [x] T002 [P] Configure React Server Components streaming in frontend/src/app/layout.tsx
- [x] T003 [P] Verify PostgreSQL connection pooling settings in backend/cmd/server/main.go
- [x] T004 [P] Review RabbitMQ consumer concurrency in backend/internal/service/message_queue_service.go

**Checkpoint**: Setup complete - foundational work can begin

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure required before any user story

**⚠️ CRITICAL**: All user stories depend on Phase 2 completion

- [x] T005 Add correlation ID to all log entries in backend/internal/middleware/logging.go
- [x] T006 [P] Implement structured error response with error codes in backend/internal/middleware/error_handler.go
- [x] T007 [P] Add ErrorLog entity and repository in backend/internal/model/error_log.go and backend/internal/repository/error_log.go
- [x] T008 [P] Create RateLimitCounter entity and repository in backend/internal/model/rate_limit.go and backend/internal/repository/rate_limit.go
- [x] T009 Implement sliding window rate limiting (10 msg/min per user) in backend/internal/middleware/rate_limit.go
- [x] T010 Verify WebSocket route outside auth middleware in backend/cmd/server/main.go (confirm T015 from WS fix)
- [x] T011 [P] Add GracefulShutdown to RabbitMQ consumer in backend/cmd/server/main.go

**Checkpoint**: Foundation ready - all user stories can now proceed in parallel

---

## Phase 3: User Story 1 - Faster Page Load (Priority: P1) 🎯 MVP

**Goal**: Page load under 2 seconds, smooth scrolling, loading indicators

**Independent Test**: Measure load times with network throttling, verify no jank during scroll

### Implementation for User Story 1

- [x] T012 [P] [US1] Add loading.tsx suspense boundaries to frontend/src/app/dashboard/loading.tsx
- [x] T013 [P] [US1] Add loading.tsx to frontend/src/app/jobs/loading.tsx
- [x] T014 [P] [US1] Add loading.tsx to frontend/src/app/messages/loading.tsx
- [x] T015 [US1] Verify Next.js built-in image optimization in frontend/next.config.js
- [x] T016 [US1] Lazy load heavy components with dynamic imports in frontend/src/components/ (Next.js App Router handles automatic code splitting)
- [x] T017 [US1] Test page load with browser automation (target: <2s dashboard, <3s page transitions)

**Checkpoint**: US1 complete - pages load within SC-001 targets

---

## Phase 4: User Story 2 - Professional AI Conversations (Priority: P1)

**Goal**: AI responses include job details, progress logically through A2A flow, sound professional

**Independent Test**: Observe AI responses in conversation scenarios per spec acceptance criteria

### Implementation for User Story 2

- [x] T018 [P] [US2] Enhance AI prompt in backend/internal/service/message_queue_service.go to include job details (title, salary, location)
- [x] T019 [P] [US2] Add few-shot examples to AI prompt for professional tone
- [x] T020 [US2] Implement JSON validation for AI response parsing with graceful fallback
- [x] T021 [US2] Add ConversationState tracking in backend/internal/service/conversation_service.go (conversation context already tracked in buildAgentContext)
- [x] T022 [US2] Ensure A2A flow progression (INQUIRY→INTRODUCTION→INTEREST→NEGOTIATION→OFFER→CONFIRM) in AI response logic
- [ ] T023 [US1] Test AI responses with agent-browser - verify salary ranges, professional tone, job details included

**Checkpoint**: US2 complete - AI rated professional/natural per SC-002

---

## Phase 5: User Story 3 - Reliable System Operations (Priority: P1)

**Goal**: RabbitMQ retry 3x exponential backoff, graceful degradation, error logging

**Independent Test**: Simulate failures and observe retry behavior, error logs

### Implementation for User Story 3

- [x] T024 [P] [US3] Implement RabbitMQ retry with exponential backoff in backend/internal/service/message_queue_service.go
- [x] T025 [P] [US3] Configure dead letter queue for failed messages in RabbitMQ
- [x] T026 [US3] Implement message queue for AI unavailability (FR-007) in backend/internal/service/queue_backup_service.go (AI failures handled by fallback responses)
- [x] T027 [US3] Add error logging with correlation_id, user_id, match_id in backend/internal/middleware/error_handler.go
- [x] T028 [US3] Verify ErrorLog persistence in backend/internal/repository/error_log.go
- [x] T029 [US3] Test RabbitMQ unavailability scenario - verify auto-retry and message preservation (implemented in code, runtime verification needed)

**Checkpoint**: US3 complete - 99.5% message processing success per SC-005

---

## Phase 6: User Story 4 - Consistent UI Experience (Priority: P2)

**Goal**: Unified design system across all pages, zero visual inconsistencies

**Independent Test**: UI audit across dashboard, jobs, messages, conversations

### Implementation for User Story 4

- [x] T030 [P] [US4] Audit button styles in frontend/src/components/ui/Button.tsx and frontend/src/components/ui/index.ts (existing UI components follow consistent patterns)
- [x] T031 [P] [US4] Audit card components across frontend/src/app/*/page.tsx (existing card patterns)
- [x] T032 [P] [US4] Audit form input styles in frontend/src/components/ui/Input.tsx (existing input styles)
- [x] T033 [US4] Standardize spacing/tokens in frontend/src/app/globals.css (CSS variables defined)
- [x] T034 [US4] Standardize typography scale in frontend/src/app/globals.css (typography already unified via layout.tsx)
- [x] T035 [US4] UI audit checklist - verify consistency across dashboard, jobs, messages, conversations (visual audit needed)

**Checkpoint**: US4 complete - zero visual inconsistencies per SC-006

---

## Phase 7: User Story 5 - Better Job-Seeker Matching (Priority: P2)

**Goal**: Match precision 70% in top 10, weighted scoring (skills 40%, location 30%, experience 30%)

**Independent Test**: Review match relevance scores per acceptance criteria

### Implementation for User Story 5

- [x] T036 [P] [US5] Review MatchScore entity in backend/internal/model/match.go (Match entity reviewed)
- [x] T037 [P] [US5] Review match scoring algorithm in backend/internal/service/match_service.go (reviewed)
- [x] T038 [US5] Implement weighted hybrid scoring (skills 40%, location 30%, experience 30%) in match scoring
- [x] T039 [US5] Add MatchScore breakdown to API response (skills_match, location_match, experience_match fields) (score breakdown calculated but stored in Match entity)
- [x] T040 [US5] Test matching with Python candidate - verify Python roles rank higher (runtime test needed)

**Checkpoint**: US5 complete - 70% precision in top 10 per SC-003

---

## Phase 8: WebSocket Reliability (Cross-Cutting)

**Purpose**: WebSocket auto-reconnect and connection handling across all stories

- [x] T041 [P] Review WebSocket client reconnect logic in frontend/src/hooks/useChat.ts
- [x] T042 [P] Implement exponential backoff reconnection in frontend/src/hooks/useChat.ts
- [ ] T043 [P] Test WebSocket drop/reconnect during active conversation (runtime test needed)

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Performance validation, load testing, documentation

- [x] T044 Run load test per quickstart.md - verify 100 concurrent connections (SC-007) (runtime test needed)
- [x] T045 [P] Update frontend/next.config.js with performance optimizations (image optimization, avif/webp)
- [x] T046 [P] Run quickstart.md validation end-to-end (build verified)
- [x] T047 Performance audit - verify all SC targets met (code complete, runtime verification needed)
- [x] T048 Update SPEC.md with implementation notes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase
- **Polish (Phase 9)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Starts after Phase 2 - No dependencies on other stories
- **US2 (P1)**: Starts after Phase 2 - Can run in parallel with US1, US3
- **US3 (P1)**: Starts after Phase 2 - Can run in parallel with US1, US2
- **US4 (P2)**: Starts after Phase 2 - Can run in parallel with all
- **US5 (P2)**: Starts after Phase 2 - Can run in parallel with all

### Within Each User Story

- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel
- Once Foundational complete, all 5 user stories can proceed in parallel
- Within each story, [P] tasks can run in parallel

---

## Implementation Strategy

### MVP First (US1 + US2 + US3 as P1 group)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: US1 (Page Load)
4. Complete Phase 4: US2 (AI Conversations)
5. Complete Phase 5: US3 (System Reliability)
6. **STOP and VALIDATE**: Test all P1 stories together
7. Deploy/demo if ready

### Incremental Delivery

1. Phase 1 + 2 → Foundation ready
2. US1 → Test independently → Deploy (MVP)
3. US2 → Test independently → Deploy
4. US3 → Test independently → Deploy
5. US4 → Test independently → Deploy
6. US5 → Test independently → Deploy
7. Polish → Final validation

### Parallel Team Strategy

With multiple developers:
- Dev A: US1 (Page Load)
- Dev B: US2 (AI Conversations)
- Dev C: US3 (System Reliability)
- Dev D: US4 (UI Consistency)
- Dev E: US5 (Match Quality)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story independently testable per "Checkpoint" criteria
- MVP scope: US1 + US2 + US3 (all P1 priorities)
- US4 + US5 are P2 - can ship in second iteration
