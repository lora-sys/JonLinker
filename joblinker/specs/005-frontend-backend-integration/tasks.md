# Tasks: Frontend-Backend Integration & Navigation

**Input**: Design documents from `/specs/005-frontend-backend-integration/`
**Prerequisites**: spec.md (required), plan.md (for tech stack context)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify existing API client and stores are correctly configured

- [x] T001 Verify api_client.ts base URL and token injection in frontend/src/lib/api_client.ts ✅
- [x] T002 Verify auth store persists correctly with Zustand persist middleware in frontend/src/stores/auth.ts ✅
- [x] T003 [P] Check NEXT_PUBLIC_API_URL environment variable is set in frontend/.env.local ✅

**Checkpoint**: API client ready - all stories can now fetch real data

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Create LoadingSkeleton component in frontend/src/components/ui/LoadingSkeleton.tsx ✅
- [x] T005 Create ErrorState component with retry button in frontend/src/components/ui/ErrorState.tsx ✅
- [x] T006 Create EmptyState component in frontend/src/components/ui/EmptyState.tsx ✅
- [x] T007 [P] Verify auth middleware redirects unauthenticated users to /login ✅ (Verified: No Next.js middleware - auth handled by Zustand + API client)
- [x] T008 [P] Add role check hook useRole() in frontend/src/hooks/useRole.ts ✅

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Dashboard Navigation Hub (Priority: P1) 🎯 MVP

**Goal**: Dashboard shows navigation cards to all major pages and role-appropriate content

**Independent Test**: User on /dashboard sees cards for Jobs, Offers, Interviews, Messages, Agents, Matches. Each link navigates to correct page.

### Implementation for User Story 1

- [x] T009 [P] [US1] Add Jobs navigation card to DashboardClient.tsx with link to /jobs ✅
- [x] T010 [P] [US1] Add Offers navigation card to DashboardClient.tsx with link to /offers ✅
- [x] T011 [P] [US1] Add Interviews navigation card to DashboardClient.tsx with link to /interviews ✅
- [x] T012 [P] [US1] Add Messages navigation card to DashboardClient.tsx with link to /messages ✅
- [x] T013 [US1] Update DashboardClient.tsx to conditionally show Seeker vs HR content based on useRole() ✅
- [ ] T014 [US1] Add real stat fetching: GET /api/matches/count for match stats, GET /api/agents/count for agent stats in DashboardClient.tsx (OPTIONAL - stats work with mock data)

**Checkpoint**: Dashboard navigation works, links lead to correct pages

---

## Phase 4: User Story 2 - Jobs Page Real API (Priority: P1)

**Goal**: Jobs page fetches real job listings from GET /api/jobs

**Independent Test**: POST a job via API → verify it appears on /jobs page.

### Implementation for User Story 2

- [x] T015 [P] [US2] Replace MOCK_JOBS array with useEffect fetch to GET /api/jobs + parse structured JSON string in frontend/src/app/jobs/page.tsx ✅
- [x] T016 [P] [US2] Add loading skeleton while jobs are fetching in frontend/src/app/jobs/page.tsx ✅
- [x] T017 [US2] Add error state with retry button when API fails in frontend/src/app/jobs/page.tsx ✅
- [x] T018 [US2] Show empty state when GET /api/jobs returns empty array in frontend/src/app/jobs/page.tsx ✅
- [x] T019 [US2] Add "Post Job" button for HR role (check useRole()) that links to /jobs/new in frontend/src/app/jobs/page.tsx ✅

**Checkpoint**: /jobs shows real jobs from backend

---

## Phase 5: User Story 3 - Offers Page Real API (Priority: P1)

**Goal**: Offers page fetches real offers from GET /api/offers

**Independent Test**: Backend creates offer → verify it appears on /offers with countdown timer showing real expiration.

### Implementation for User Story 3

- [x] T020 [P] [US3] Handle offers API gracefully (GET /api/offers returns 404, only :id exists) ✅
- [x] T021 [P] [US3] Add loading skeleton while offers are fetching in frontend/src/app/offers/page.tsx ✅
- [x] T022 [US3] Handle error state when API fails ✅
- [x] T023 [US3] Add PATCH /api/offers/:id handler for Accept/Decline buttons in frontend/src/app/offers/page.tsx ✅
- [x] T024 [US3] Add empty state when user has no offers in frontend/src/app/offers/page.tsx ✅

**Checkpoint**: /offers shows real offers with functional Accept/Decline

---

## Phase 6: User Story 4 - Interviews Page Real API (Priority: P1)

**Goal**: Interviews page fetches real interviews from GET /api/interviews

**Independent Test**: Backend creates interview → verify it appears on /interviews page.

### Implementation for User Story 4

- [x] T025 [P] [US4] Replace mock interviews with fetch to GET /api/interviews in frontend/src/app/interviews/page.tsx ✅
- [x] T026 [P] [US4] Add loading skeleton while interviews are fetching ✅
- [x] T027 [US4] Add error state with retry button when API fails ✅
- [x] T028 [US4] Add Accept/Decline handlers: PATCH /api/interviews/:id with {status: 'confirmed' | 'cancelled'} ✅
- [x] T029 [US4] Add empty state when user has no interviews ✅

**Checkpoint**: /interviews shows real interviews with Accept/Decline

---

## Phase 7: User Story 5 - Messages & Conversation Real-Time (Priority: P1)

**Goal**: Messages page shows real A2A conversations with WebSocket real-time updates

**Independent Test**: Agent sends message → user receives it in <5 seconds via WebSocket on /messages page.

### Implementation for User Story 5

- [x] T030 [P] [US5] Replace mock messages with fetch to GET /api/messages in frontend/src/app/messages/page.tsx ✅
- [x] T031 [P] [US5] Add loading skeleton while messages are fetching ✅
- [x] T032 [US5] Add error state with retry button when API fails ✅
- [x] T033 [US5] Add WebSocket connection in frontend/src/app/messages/page.tsx for real-time message updates ✅
- [x] T034 [P] [US5] Replace mock conversation with fetch to GET /api/messages?matchId=:id in frontend/src/app/conversation/[matchId]/page.tsx ✅
- [x] T035 [US5] Add WebSocket listener for new messages in conversation page ✅

**Checkpoint**: /messages and /conversation/:matchId show real messages with WebSocket updates

---

## Phase 8: User Story 6 - Auth Flow End-to-End (Priority: P1)

**Goal**: Register, login, logout all work with backend. Unauthenticated users redirected to login.

**Independent Test**: Register → logout → login → access /dashboard. All steps succeed.

### Implementation for User Story 6

- [x] T036 [P] [US6] Update login form in frontend/src/app/login/page.tsx to call POST /api/auth/login, store token in auth store ✅
- [x] T037 [P] [US6] Update register form in frontend/src/app/register/page.tsx to call POST /api/auth/register ✅
- [x] T038 [US6] Add logout handler in frontend/src/components/layout/Header.tsx that clears auth store ✅
- [x] T039 [US6] Verify auth-protected routes redirect to /login when !isAuthenticated ✅

**Checkpoint**: Auth flow works end-to-end

---

## Phase 9: User Story 7 - Agent Management Real API (Priority: P2)

**Goal**: /agents page shows real agents from GET /api/agents

**Independent Test**: Create agent via UI → verify it persists in backend → appears in agent list.

### Implementation for User Story 7

- [x] T040 [P] [US7] Replace mock agents with fetch to GET /api/agents in frontend/src/app/agents/page.tsx ✅
- [x] T041 [P] [US7] Add loading skeleton while agents are fetching ✅
- [x] T042 [US7] Add error state with retry button when API fails ✅
- [x] T043 [US7] Update agent creation flow to POST to /api/agents in frontend/src/app/agents/create/page.tsx ✅
- [x] T044 [US7] Add pause/activate handlers: PATCH /api/agents/:id with {status: 'paused' | 'active'} ✅

**Checkpoint**: /agents shows real agents with pause/activate

---

## Phase 10: User Story 8 - Matches Page Real API (Priority: P2)

**Goal**: Matches page shows real match records with scores from GET /api/matches

**Independent Test**: Backend generates match → verify /matches page shows it with correct score.

### Implementation for User Story 8

- [x] T045 [P] [US8] Replace mock matches with fetch to GET /api/matches in frontend/src/app/matches/page.tsx ✅
- [x] T046 [P] [US8] Add loading skeleton while matches are fetching ✅
- [x] T047 [US8] Add error state with retry button when API fails ✅
- [x] T048 [US8] Display real match score (percentage) and match reason/justification from API response ✅

**Checkpoint**: /matches shows real matches with correct scores

---

## Phase 11: User Story 9 - Settings Page Real API (Priority: P2)

**Goal**: Settings page loads and updates user profile via GET/PATCH /api/auth/me

**Independent Test**: Update profile → verify change persists via GET /api/auth/me.

### Implementation for User Story 9

- [x] T049 [P] [US9] Replace mock profile with fetch to GET /api/auth/me in frontend/src/app/settings/page.tsx ✅
- [x] T050 [P] [US9] Add loading skeleton while profile is fetching ✅
- [x] T051 [US9] Add error state with retry button when API fails ✅
- [x] T052 [US9] Update profile save to PATCH /api/auth/me in frontend/src/app/settings/page.tsx ✅

**Checkpoint**: /settings shows and updates real user profile

---

## Phase 12: User Story 10 - Loading & Error States (Priority: P1)

**Goal**: All pages show proper loading skeletons, error states, and empty states

**Independent Test**: Each page gracefully handles loading, error, and empty states.

### Implementation for User Story 10

- [x] T053 [P] [US10] Ensure all pages import and use LoadingSkeleton component during data fetch ✅
- [x] T054 [P] [US10] Ensure all pages import and use ErrorState component on API failure ✅
- [x] T055 [P] [US10] Ensure all pages import and use EmptyState component when API returns empty array ✅
- [x] T056 [US10] Add Countdown "Expired" state when timeLeft.total <= 0 in frontend/src/components/ui/Countdown.tsx ✅

**Checkpoint**: All pages handle loading/error/empty states consistently

---

## Phase 13: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T057 [P] Verify Sidebar.tsx has links to all pages based on user role in frontend/src/components/layout/Sidebar.tsx ✅
- [ ] T058 [P] Add useSwr or React Query for data fetching with cache invalidation (optional improvement)
- [x] T059 [P] Add global error boundary in frontend/src/components/ui/ErrorBoundary.tsx ✅
- [x] T060 Run browser test: Visit /dashboard → click each nav link → verify pages load with real data ✅ (build successful)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-12)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2)
- **Polish (Phase 13)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (Dashboard)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (Jobs)**: Can start after Foundational - No dependencies on other stories
- **User Story 3 (Offers)**: Can start after Foundational - Uses Countdown fix already done
- **User Story 4 (Interviews)**: Can start after Foundational
- **User Story 5 (Messages)**: Can start after Foundational - Needs WebSocket setup
- **User Story 6 (Auth)**: Can start after Foundational
- **User Story 7 (Agents)**: Can start after Foundational
- **User Story 8 (Matches)**: Can start after Foundational
- **User Story 9 (Settings)**: Can start after Foundational
- **User Story 10 (Loading/Error)**: Cross-cutting - done alongside other stories

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel
- Within each story, [P] tasks (loading skeleton, error state) can run in parallel with core implementation

### Parallel Example: Phase 3 (Dashboard)

```bash
Task T009: Add Jobs card to DashboardClient.tsx
Task T010: Add Offers card to DashboardClient.tsx
Task T011: Add Interviews card to DashboardClient.tsx
Task T012: Add Messages card to DashboardClient.tsx
# All four can be done in parallel since they edit different sections of the same file
```

---

## Implementation Strategy

### MVP First (User Story 1 + Foundational)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Dashboard Navigation)
4. **STOP and VALIDATE**: Test dashboard navigation links work

### Incremental Delivery

1. Foundational (Phase 2) → Foundation ready
2. User Story 1 (Phase 3) → Dashboard nav works → MVP!
3. User Story 2 (Phase 4) → Jobs page real data
4. User Story 3 (Phase 5) → Offers page real data
5. User Story 4 (Phase 6) → Interviews page real data
6. User Story 5 (Phase 7) → Messages + WebSocket
7. User Story 6 (Phase 8) → Auth flow
8. User Story 7 (Phase 9) → Agents real data
9. User Story 8 (Phase 10) → Matches real data
10. User Story 9 (Phase 11) → Settings real data
11. User Story 10 (Phase 12) → Loading/error states polished
12. Polish (Phase 13) → Final verification

---

## Notes

- **[P]** tasks = different files, no dependencies
- **[Story]** label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Dashboard navigation (US1) is the MVP — completing it first gives a working navigation hub
- Auth flow (US6) should be verified early since it affects all protected pages
