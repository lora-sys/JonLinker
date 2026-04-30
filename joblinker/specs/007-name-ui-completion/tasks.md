# Tasks: Frontend UI Completion

**Input**: Design documents from `specs/007-name-ui-completion/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Constitution I (NON-NEGOTIABLE) requires unit + integration tests for business logic and API contracts. agent-browser E2E (spec.md) supplements but does not replace unit/integration tests. Backend tests use Go standard `testing` package; frontend tests use `jest`/`vitest`.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify Aceternity UI components are installed and accessible

- [X] T001 [P] Verify Aceternity signup-form installed at `frontend/src/components/ui/signup-form.tsx`
- [X] T002 [P] Verify Aceternity expandable-cards installed at `frontend/src/components/ui/expandable-cards.tsx`
- [X] T003 [P] Verify Aceternity focus-cards and background-beams installed

**Note**: Phase 1 is minimal — frontend deps already in package.json from prior work

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure needed for ALL user stories — WS hook, backend ping/pong

**⚠️ CRITICAL**: No user story work begins until this phase is complete

- [X] T004 Create `frontend/src/lib/websocket.ts` — WS client utility (connect, disconnect, send, onMessage, onStatusChange)
- [X] T005 Create `frontend/src/hooks/use-websocket.ts` — WS hook with exponential backoff (5s→10s→20s→60s max), connection states (Connecting|Connected|Disconnected|Reconnecting), ping/pong keepalive, pending message queue flush on reconnect
- [X] T006 [P] Modify `backend/internal/handler/message.go` — add ping/pong echo, message echo (for test verification), verify JWT from query param on WS upgrade
- [X] T007 [P] Verify existing RabbitMQ consumer setup — RabbitMQ not configured in this environment, skipping

**Checkpoint**: Foundation ready — WS hook and backend echo working, all 4 stories can proceed

---

## Phase 3: User Story 1 - Recruiter Posts Job (Priority: P1) 🎯 MVP

**Goal**: Recruiter can post a job via `/jobs/new` UI and see it in job board within 5 seconds

**Independent Test**: Recruiter visits `/jobs/new` → fills form → submits → job appears on `/jobs` within 5s

### Tests for User Story 1

- [X] T030 [P] [US1] Unit test for job.go validation (structured.requirements ≥1, type enum, recruiter role check) in `backend/tests/unit/handler/job_test.go`
- [X] T031 [P] [US1] Integration test for `POST /api/jobs` — valid payload returns 201, missing fields return 400, non-recruiter returns 403 — in `backend/tests/integration/job_test.go`

### Implementation for User Story 1

- [X] T008 Create `frontend/src/components/ui/job-form.tsx` — Aceternity signup-form base, fields: title, description, location, type (dropdown: full-time/part-time/contract), salary_range (min/max), structured.requirements (tag input), submit calls `POST /api/jobs` with JWT auth
- [X] T009 Create `frontend/src/app/jobs/new/page.tsx` — recruiter-only page, role guard (redirect if not recruiter), renders job-form.tsx, success redirect to `/jobs`
- [X] T010 [P] Modify `frontend/src/app/jobs/page.tsx` — ensure job listing displays newly created jobs from API
- [X] T011 [P] Modify `backend/internal/handler/job.go` — validate `structured.requirements` has ≥1 item, validate type enum, ensure recruiter role check on `POST /api/jobs`

**Checkpoint**: Recruiter can post job end-to-end via UI

---

## Phase 4: User Story 2 - WebSocket Test Page (Priority: P1)

**Goal**: User can verify WS full-duplex communication in browser via `/ws-test`

**Independent Test**: User opens `/ws-test` → sees Connected status → sends test message → receives echo within 1s

### Tests for User Story 2

- [X] T032 [P] [US2] Unit test for use-websocket.ts hook (connection states, backoff timing, ping/pong, queue flush) in `frontend/src/__tests__/hooks/use-websocket.test.ts`
- [X] T033 [P] [US2] Integration test for WS ping/pong echo in `backend/tests/integration/ws_test.go`

### Implementation for User Story 2

- [X] T012 Create `frontend/src/components/ws-debug-panel.tsx` — connection status badge (yellow/green/red), debug log (timestamps + direction arrows), message input, send button, manual reconnect button
- [X] T013 Create `frontend/src/app/ws-test/page.tsx` — uses use-websocket.ts, renders ws-debug-panel, auto-connect on mount with JWT token from cookie/localStorage
**Checkpoint**: WS debug page shows live bidirectional communication

---

## Phase 5: User Story 3 - Agent Auto-Dialogue (Priority: P2)

**Goal**: When match reaches `expressed_interest`, agents exchange ≥3 messages via RabbitMQ+WS

**Independent Test**: Match status → `expressed_interest` → within 10s, ≥3 agent messages appear in conversation thread

### Tests for User Story 3

- [ ] T034 [P] [US3] Unit test for match_service.go — on `expressed_interest`, verify RabbitMQ publish is called with correct queue — in `backend/tests/unit/service/match_service_test.go`
- [ ] T035 [P] [US3] Integration test for message_service.go `GenerateAgentResponse` — publish message to queue, verify AI response returned — in `backend/tests/integration/message_service_test.go`

### Implementation for User Story 3

- [ ] T015 [P] Review existing `backend/internal/service/match_service.go` — confirm `expressed_interest` status trigger calls RabbitMQ publish
- [ ] T016 [P] Review existing RabbitMQ consumer in `backend/internal/handler/` — confirm consumer publishes to hub broadcast channel, hub fans out to WS clients by matchId
- [ ] T017 Modify `backend/internal/service/match_service.go` — ensure on `expressed_interest`, recruiter agent intro message is queued to `recruiter_queue`
- [ ] T018 Modify `backend/internal/service/message_service.go` — ensure `GenerateAgentResponse` is called when agent message received via RabbitMQ, response published back to queue
- [ ] T019 [P] Modify `frontend/src/app/conversation/[matchId]/page.tsx` — ensure WS subscription for this matchId is established on mount, messages appended to thread in real-time

**Checkpoint**: Auto-dialogue produces ≥3 exchanges when match → expressed_interest

---

## Phase 6: User Story 4 - Interview & Offer Cards (Priority: P2)

**Goal**: When match status → `interview_scheduled` or `offer_sent`, visually distinct cards appear in conversation thread

**Independent Test**: Match status → `interview_scheduled` → interview card appears in conversation with Confirm/Decline buttons → click Confirm → status updates

### Tests for User Story 4

- [X] T036 [P] [US4] Unit test for interview.go handler (confirm/cancel transitions) in `backend/tests/unit/handler/interview_test.go`
- [X] T037 [P] [US4] Unit test for offer.go handler (accept/decline/expire transitions) in `backend/tests/unit/handler/offer_test.go`

### Implementation for User Story 4

- [X] T020 [P] Create `frontend/src/components/ui/interview-card.tsx` — displays scheduled_at (formatted), format icon (🎥📍📞), location (clickable), status badge (yellow/green/red), Confirm/Decline buttons when status=pending, calls `POST /api/interviews/:matchId/confirm|cancel`
- [X] T021 [P] Create `frontend/src/components/ui/offer-card.tsx` — displays salary_amount ($X/year from cents), start_date, expires_at (red if <48h), status badge, Accept/Decline buttons when status=pending, calls `POST /api/offers/:matchId/accept|decline`
- [X] T022 [P] Modify `frontend/src/app/conversation/[matchId]/page.tsx` — fetch match status on load, if `interview_scheduled` render `<InterviewCard>` above message list, if `offer_sent` render `<OfferCard>`, wire up confirm/decline/accept/decline actions to update UI optimistically
- [X] T023 [P] Modify `backend/internal/handler/interview.go` — ensure `POST /api/interviews/:matchId/confirm` and `/cancel` return updated interview with new status
- [X] T038 [US4] Add participant names to interview-card.tsx — fetch recruiter name from Agent API and seeker name from User API by matchId, display names on card per FR-012 in conversation thread with working actions

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup across all stories

- [ ] T024 Run agent-browser E2E for User Story 1 — register recruiter → visit `/jobs/new` → submit form → verify on `/jobs`
- [ ] T025 Run agent-browser E2E for User Story 2 — visit `/ws-test` → verify Connected → send message → verify echo
- [X] T026 [P] Verify all pages have loading states (form submissions, API fetches)
- [X] T027 [P] Verify all pages have error handling (toast/alert on failure, no raw stack traces)
- [X] T028 [P] Verify Aceternity UI background-beams applied to `/jobs/new` and `/ws-test` pages per spec FR-015
- [ ] T029 Run quickstart.md validation — test each feature against success criteria

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Stories (Phase 3–6)**: All depend on Foundational — can proceed in parallel after Phase 2
  - US1 and US2 are frontend-only (fast to implement)
  - US3 and US4 touch backend + frontend
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 — No dependencies on other stories
- **US2 (P1)**: Depends on Phase 2 — No dependencies on other stories
- **US3 (P2)**: Depends on Phase 2 — Can run in parallel with US1/US2/US4
- **US4 (P2)**: Depends on Phase 2 — Can run in parallel with US1/US2/US3

### Within Each User Story

- WS hook (T004, T005) before WS-dependent tasks (T012, T013, T019)
- Backend ping/pong + echo (T006) before WS test (T012, T013)
- Card components (T020, T021) before conversation page integration (T022)

### Parallel Opportunities

- T001, T002, T003 can run in parallel (dependency check only)
- T004 and T005 can run in parallel (different files)
- T006 and T007 can run in parallel (different files/sides)
- T008 and T010 can run in parallel (different files)
- T009 depends on T008 (job-form needed by page)
- T012 and T013 can run in parallel (debug panel and page are separate files, both use same hook)
- T015 and T016 can run in parallel (review only)
- T017 and T018 can run in parallel (different service files)
- T019 depends on T004, T005 (WS hook needed)
- T020 and T021 can run in parallel (different components)
- T022 depends on T020, T021 (card components needed)
- T023 depends on T020 (UI needs working API)
- T026 and T027 can run in parallel (different concerns)

---

## Parallel Example: User Story 1 + User Story 2

```bash
# US1: Job form + page (after T004-T007 foundation):
Task: "Create job-form.tsx at frontend/src/components/ui/job-form.tsx"
Task: "Create jobs/new/page.tsx at frontend/src/app/jobs/new/page.tsx"
Task: "Modify job.go handler at backend/internal/handler/job.go"

# US2: WS debug components (after T004-T007 foundation):
Task: "Create ws-debug-panel.tsx at frontend/src/components/ws-debug-panel.tsx"
Task: "Create ws-test/page.tsx at frontend/src/app/ws-test/page.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify deps)
2. Complete Phase 2: Foundational (WS hook + backend echo)
3. Complete Phase 3: User Story 1 — Job posting page
4. **STOP and VALIDATE**: Test job posting end-to-end
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Test → Deploy (MVP!)
3. Add US2 → Test → Deploy
4. Add US3 → Test → Deploy
5. Add US4 → Test → Deploy
6. Polish → Final validation

---

## File Inventory

| File | Action | Task |
|------|--------|------|
| `frontend/src/components/ui/job-form.tsx` | CREATE | T008 |
| `frontend/src/app/jobs/new/page.tsx` | CREATE | T009 |
| `frontend/src/app/jobs/page.tsx` | MODIFY | T010 |
| `frontend/src/components/ws-debug-panel.tsx` | CREATE | T012 |
| `frontend/src/app/ws-test/page.tsx` | CREATE | T013 |
| `frontend/src/lib/websocket.ts` | CREATE | T004 |
| `frontend/src/hooks/use-websocket.ts` | CREATE | T005 |
| `frontend/src/components/ui/interview-card.tsx` | CREATE | T020, T038 |
| `frontend/src/components/ui/offer-card.tsx` | CREATE | T021 |
| `frontend/src/app/conversation/[matchId]/page.tsx` | MODIFY | T019, T022 |
| `backend/internal/handler/message.go` | MODIFY | T006 |
| `backend/internal/handler/job.go` | MODIFY | T011 |
| `backend/internal/handler/interview.go` | MODIFY | T023 |
| `backend/internal/service/match_service.go` | MODIFY | T017 |
| `backend/internal/service/message_service.go` | MODIFY | T018 |
| `backend/tests/unit/handler/job_test.go` | CREATE | T030 |
| `backend/tests/integration/job_test.go` | CREATE | T031 |
| `frontend/src/__tests__/hooks/use-websocket.test.ts` | CREATE | T032 |
| `backend/tests/integration/ws_test.go` | CREATE | T033 |
| `backend/tests/unit/service/match_service_test.go` | CREATE | T034 |
| `backend/tests/integration/message_service_test.go` | CREATE | T035 |
| `backend/tests/unit/handler/interview_test.go` | CREATE | T036 |
| `backend/tests/unit/handler/offer_test.go` | CREATE | T037 |

---

## Notes

- Constitution I (NON-NEGOTIABLE) requires unit + integration tests — agent-browser E2E supplements but does not replace
- Tests use Go `testing` package (backend) and `jest`/`vitest` (frontend)
- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
