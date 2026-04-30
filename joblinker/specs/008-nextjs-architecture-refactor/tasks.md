# Tasks: Next.js App Router Architecture Refactor

**Input**: specs/008-nextjs-architecture-refactor/plan.md
**Prerequisites**: spec.md (required), plan.md (required), research.md, data-model.md, quickstart.md

**Tests**: Constitution I requires unit + integration tests. Build verification (`npm run build`) serves as primary quality gate for this refactor.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Baseline Verification)

**Purpose**: Establish baseline metrics and verify current state before refactoring

- [ ] T001 [P] Run `npm run build` in frontend/ to capture baseline bundle size and verify no existing errors
- [ ] T002 [P] Count current 'use client' directives: `grep -r "'use client'" src/app --include="*.tsx" | wc -l` and `grep -r "'use client'" src/components --include="*.tsx" | wc -l`
- [ ] T003 [P] Audit current component inventory in frontend/src/ — list all .tsx files with their current 'use client' status

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create shared component library that all stories will use

**CRITICAL**: No user story work begins until this phase is complete

### Shared Component Extraction

- [ ] T004a [US1] Write unit test for Card component — `frontend/src/components/ui/__tests__/Card.test.tsx`
- [ ] T004 [P] [US1] Create `frontend/src/components/ui/Card.tsx` — shared Card component with title, description, metadata, actions, children props
- [ ] T005a [US1] Write unit test for EmptyState component — `frontend/src/components/ui/__tests__/EmptyState.test.tsx`
- [ ] T005 [P] [US1] Create `frontend/src/components/ui/EmptyState.tsx` — shared EmptyState with icon, title, description, action props
- [ ] T006a [US1] Write unit test for LoadingSkeleton component — `frontend/src/components/ui/__tests__/LoadingSkeleton.test.tsx`
- [ ] T006 [P] [US1] Create `frontend/src/components/ui/LoadingSkeleton.tsx` — shared LoadingSkeleton with variant (text/card/list), count, className props
- [ ] T007a [US1] Write unit test for MessageBubble component — `frontend/src/components/ui/__tests__/MessageBubble.test.tsx`
- [ ] T007 [P] [US1] Create `frontend/src/components/ui/MessageBubble.tsx` — shared MessageBubble with content, sender, timestamp, type props
- [ ] T008a [US1] Write unit test for Avatar component — `frontend/src/components/ui/__tests__/Avatar.test.tsx`
- [ ] T008 [P] [US1] Create `frontend/src/components/ui/Avatar.tsx` — shared Avatar with src, name, size, status props
- [ ] T009a [US1] Write unit test for Skeleton component — `frontend/src/components/ui/__tests__/Skeleton.test.tsx`
- [ ] T009 [P] [US1] Create `frontend/src/components/ui/Skeleton.tsx` — shared Skeleton with className, width, height props
- [ ] T010a [US1] Write unit test for JobCard component — `frontend/src/components/feature/__tests__/JobCard.test.tsx`
- [ ] T010 [P] [US1] Create `frontend/src/components/feature/JobCard.tsx` — domain-specific JobCard extracting from jobs/page.tsx
- [ ] T011a [US1] Write unit test for MatchCard component — `frontend/src/components/feature/__tests__/MatchCard.test.tsx`
- [ ] T011 [P] [US1] Create `frontend/src/components/feature/MatchCard.tsx` — domain-specific MatchCard extracting from matches/page.tsx
- [ ] T012a [US1] Write unit test for OfferCard component — `frontend/src/components/feature/__tests__/OfferCard.test.tsx`
- [ ] T012 [P] [US1] Create `frontend/src/components/feature/OfferCard.tsx` — domain-specific OfferCard extracting from offers/page.tsx
- [ ] T013a [US1] Write unit test for InterviewCard component — `frontend/src/components/feature/__tests__/InterviewCard.test.tsx`
- [ ] T013 [P] [US1] Create `frontend/src/components/feature/InterviewCard.tsx` — domain-specific InterviewCard extracting from interviews/page.tsx

**Checkpoint**: Shared component library created, all components importable from their new locations

---

## Phase 3: User Story 1 - Server/Client Component Separation (Priority: P1) 🎯

**Goal**: Convert non-interactive pages to Server Components, reduce client bundle size

**Independent Test**: `npm run build` succeeds with no TypeScript errors, bundle size decreased

### Implementation for User Story 1

- [ ] T014 [P] [US1] Refactor `frontend/src/app/page.tsx` — remove 'use client' if not needed, verify still works
- [ ] T015 [P] [US1] Refactor `frontend/src/app/layout.tsx` — remove 'use client' if not needed, keep only AppShell as client
- [ ] T016 [P] [US1] Refactor `frontend/src/app/jobs/page.tsx` — remove 'use client', use JobCard from feature/ components
- [ ] T017 [P] [US1] Refactor `frontend/src/app/matches/page.tsx` — remove 'use client', use MatchCard from feature/ components
- [ ] T018 [P] [US1] Refactor `frontend/src/app/offers/page.tsx` — remove 'use client', use OfferCard from feature/ components
- [ ] T019 [P] [US1] Refactor `frontend/src/app/interviews/page.tsx` — remove 'use client', use InterviewCard from feature/ components
- [ ] T020 [P] [US1] Refactor `frontend/src/app/messages/page.tsx` — remove 'use client' if no client hooks needed
- [ ] T021 [P] [US1] Refactor `frontend/src/app/privacy/page.tsx` — remove 'use client' if static content
- [ ] T022 [P] [US1] Refactor `frontend/src/app/admin/page.tsx` — remove 'use client' if no interactivity
- [ ] T023 [P] [US1] Refactor `frontend/src/app/agents/page.tsx` — remove 'use client' if no client hooks
- [ ] T024 [P] [US1] Refactor `frontend/src/app/settings/page.tsx` — convert to server outer + client islands pattern
- [ ] T025 [P] [US1] Refactor `frontend/src/app/resume/page.tsx` — convert to server outer + client islands pattern
- [ ] T026 [P] [US1] Refactor `frontend/src/app/resume/generate/page.tsx` — convert to server outer + client islands pattern
- [ ] T026b [US1] Remove zustand `_hasRehydrated` hydration pattern — replace with cookie-based session check; update auth store and all pages using `_hasRehydrated` to rely on server-side auth via API route or middleware instead

**Checkpoint**: All static pages converted to Server Components, client bundle reduced

---

## Phase 4: User Story 2 - Conversation/Chat Architecture Refactor (Priority: P2)

**Goal**: Implement server outer + client inner pattern for conversation pages

**Independent Test**: Open conversation page → messages visible in initial HTML (SSR), WebSocket shows new messages in real-time

### Implementation for User Story 2

- [ ] T027 [P] [US2] Create `frontend/src/components/chat/ChatInput.tsx` — client component for message input with onSend callback
- [ ] T028 [P] [US2] Create `frontend/src/components/chat/MessageThread.tsx` — client component for WebSocket subscription and auto-scroll
- [ ] T029 [P] [US2] Create `frontend/src/components/chat/ConversationHeader.tsx` — server component for match metadata display
- [ ] T030 [US2] Refactor `frontend/src/app/conversation/[matchId]/page.tsx` — server outer layout + client inner islands pattern
- [ ] T031 [P] [US2] Refactor `frontend/src/components/chat/ChatWindow.tsx` — ensure WebSocket hook properly integrated

**Checkpoint**: Conversation page renders messages via SSR, real-time updates work via WebSocket

---

## Phase 5: User Story 3 - Shared Component Extraction (Priority: P2)

**Goal**: Ensure all pages use shared components, eliminate duplicate markup

**Independent Test**: No duplicate card/empty-state/skeleton markup in any page file

### Implementation for User Story 3

- [ ] T032 [P] [US3] Update `frontend/src/app/jobs/page.tsx` — replace inline card markup with JobCard component
- [ ] T033 [P] [US3] Update `frontend/src/app/matches/page.tsx` — replace inline card markup with MatchCard component
- [ ] T034 [P] [US3] Update `frontend/src/app/offers/page.tsx` — replace inline card markup with OfferCard component
- [ ] T035 [P] [US3] Update `frontend/src/app/interviews/page.tsx` — replace inline card markup with InterviewCard component
- [ ] T036 [P] [US3] Audit all pages for inline EmptyState markup — replace with EmptyState component
- [ ] T037 [P] [US3] Audit all pages for inline skeleton/loading markup — replace with LoadingSkeleton component
- [ ] T038 [P] [US3] Audit all pages for inline Avatar usage — ensure Avatar component is used consistently
- [ ] T039 [P] [US3] Audit all pages for inline MessageBubble usage in conversation/[matchId] — ensure MessageBubble component is used

**Checkpoint**: All duplicate UI patterns eliminated, pages use shared components exclusively

---

## Phase 6: User Story 4 - Project Structure Normalization (Priority: P3)

**Goal**: Standardize naming conventions and import patterns

**Independent Test**: All components follow PascalCase naming, all imports use @/ alias consistently

### Implementation for User Story 4

- [ ] T040 [P] [US4] Audit import statements across all pages — ensure @/ alias used (not ../../.. patterns)
- [ ] T041 [P] [US4] Normalize component naming — ensure all .tsx files use PascalCase matching their component name
- [ ] T042 [P] [US4] Create `frontend/src/components/feature/` directory structure if missing — domain-specific components
- [ ] T043 [P] [US4] Verify `components/ui/` has consistent structure — all base primitives in same directory

**Checkpoint**: Project structure normalized, imports consistent

---

## Phase 7: Polish & Verification

**Purpose**: Final validation and cleanup

- [ ] T044 Run `npm run build` — verify no TypeScript errors, no lint warnings in modified files
- [ ] T045 Compare bundle size before/after — verify at least 20% reduction in client JS
- [ ] T046 [P] Run agent-browser E2E smoke tests — login flow, job posting, conversation page
- [ ] T047 Verify WebSocket functionality — real-time messages work in conversation page

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Stories (Phase 3–6)**: All depend on Foundational — can proceed in parallel after Phase 2
  - US1 (Server/Client Separation) and US3 (Shared Components) are closely related
  - US2 (Conversation Refactor) depends on having ChatInput/MessageThread extracted
  - US4 (Structure Normalization) can run in parallel with all
- **Polish (Phase 7)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 — No dependencies on other stories
- **US2 (P2)**: Depends on Phase 2 — Can run in parallel with US1/US3/US4
- **US3 (P2)**: Depends on Phase 2 — Can run in parallel with US1/US2/US4
- **US4 (P3)**: Depends on Phase 2 — Can run in parallel with all

### Within Each User Story

- T004-T013 (shared components) before any US1/US2/US3 page refactoring
- US2 (T027-T031) depends on T007 (MessageBubble) only
- US3 (T032-T039) depends on T010-T013 (feature components) and T004-T009 (UI primitives)

### Parallel Opportunities

- T004, T005, T006, T007, T008, T009 can run in parallel (different component files)
- T014-T026 can run in parallel (different page files, all use shared components)
- T032-T039 can run in parallel (different page files)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (baseline verification)
2. Complete Phase 2: Foundational (shared components)
3. Complete Phase 3: US1 — Server/Client separation
4. **STOP and VALIDATE**: Run `npm run build`, verify bundle size reduction
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Test → Deploy (bundle size improved!)
3. Add US3 → Test → Deploy (consistency improved!)
4. Add US2 → Test → Deploy (conversation SSR working!)
5. Add US4 → Test → Deploy (structure normalized!)
6. Polish → Final validation

---

## File Inventory

| File | Action | Task |
|------|--------|------|
| `frontend/src/components/ui/Card.tsx` | CREATE | T004 |
| `frontend/src/components/ui/EmptyState.tsx` | CREATE | T005 |
| `frontend/src/components/ui/LoadingSkeleton.tsx` | CREATE | T006 |
| `frontend/src/components/ui/MessageBubble.tsx` | CREATE | T007 |
| `frontend/src/components/ui/Avatar.tsx` | CREATE | T008 |
| `frontend/src/components/ui/Skeleton.tsx` | CREATE | T009 |
| `frontend/src/components/feature/JobCard.tsx` | CREATE | T010 |
| `frontend/src/components/feature/MatchCard.tsx` | CREATE | T011 |
| `frontend/src/components/feature/OfferCard.tsx` | CREATE | T012 |
| `frontend/src/components/feature/InterviewCard.tsx` | CREATE | T013 |
| `frontend/src/components/chat/ChatInput.tsx` | CREATE | T027 |
| `frontend/src/components/chat/MessageThread.tsx` | CREATE | T028 |
| `frontend/src/components/chat/ConversationHeader.tsx` | CREATE | T029 |
| `frontend/src/app/page.tsx` | REFACTOR | T014 |
| `frontend/src/app/layout.tsx` | REFACTOR | T015 |
| `frontend/src/app/jobs/page.tsx` | REFACTOR | T016, T032 |
| `frontend/src/app/matches/page.tsx` | REFACTOR | T017, T033 |
| `frontend/src/app/offers/page.tsx` | REFACTOR | T018, T034 |
| `frontend/src/app/interviews/page.tsx` | REFACTOR | T019, T035 |
| `frontend/src/app/messages/page.tsx` | REFACTOR | T020, T036 |
| `frontend/src/app/privacy/page.tsx` | REFACTOR | T021 |
| `frontend/src/app/admin/page.tsx` | REFACTOR | T022 |
| `frontend/src/app/agents/page.tsx` | REFACTOR | T023 |
| `frontend/src/app/settings/page.tsx` | REFACTOR | T024 |
| `frontend/src/app/resume/page.tsx` | REFACTOR | T025 |
| `frontend/src/app/resume/generate/page.tsx` | REFACTOR | T026 |
| `frontend/src/app/conversation/[matchId]/page.tsx` | REFACTOR | T030 |
| `frontend/src/components/chat/ChatWindow.tsx` | REFACTOR | T031 |
| `frontend/src/components/chat/MessageBubble.tsx` | REFACTOR | T039 |

---

## Notes

- Constitution I (NON-NEGOTIABLE): Build verification (`npm run build`) must pass
- All pages must maintain original business functionality (WebSocket, RabbitMQ, agents, vector matching)
- WebSocket hook (`use-websocket.ts`) and auth store (`stores/auth.ts`) MUST remain client components
- Shared component props interfaces defined in data-model.md
- Migration order: extract shared components first → refactor pages → remove unnecessary 'use client'
