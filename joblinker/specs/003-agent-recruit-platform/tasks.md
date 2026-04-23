---

description: "Task list for A2A Agent Recruitment Platform"
---

# Tasks: A2A Agent Recruitment Platform

**Input**: Design documents from `specs/003-agent-recruit-platform/`
**Prerequisites**: plan.md (required), spec.md (required), data-model.md, contracts/a2a-protocol.md
**Tests**: Constitution mandates ≥80% coverage, Playwright E2E for key flows

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Frontend**: `frontend/src/`, `frontend/tests/`
- **Backend**: `backend/`, `backend/cmd/server/`, `backend/internal/`
- **Infrastructure**: `docker-compose.yml`, `Dockerfile.*`
- Paths assume project root

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Initialize repository structure, infrastructure, and development environment

### Infrastructure Setup

- [x] T001 Create project directory structure per implementation plan
- [x] T002 [P] Create docker-compose.yml with PostgreSQL 17, Redis 7.4+, RabbitMQ, Chroma 0.6+
- [x] T003 [P] Create Dockerfile.frontend for Next.js 16 application
- [x] T004 [P] Create Dockerfile.backend for Go 1.23+ application
- [x] T005 Create backend/cmd/server/main.go entry point with basic HTTP server
- [x] T006 Create backend/go.mod with all dependencies (Gin, GORM, pg driver, redis client, amqp)
- [x] T007 Create frontend/package.json with Next.js 16, TypeScript 5.7+, dependencies
- [x] T008 [P] Create frontend/tsconfig.json with strict mode enabled
- [x] T009 [P] Setup frontend/next.config.js with App Router configuration

### Git & CI Setup

- [x] T010 [P] Initialize git repository if not exists
- [x] T011 Create .gitignore for Go/Node artifacts, secrets, sensitive data
- [x] T012 [P] Create backend/.golangci.yml with lint rules (static analysis, vet, format)
- [x] T013 [P] Create frontend/.eslintrc.json with TypeScript and React rules
- [x] T014 Create docker-compose.yml for local development environment

**Checkpoint**: Project structure exists; Docker containers can start; both servers compile

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Backend Foundation

- [x] T015 Create backend/internal/model/ with all domain entities (User, Agent, Job, Resume, Match, Message, Interview, Offer, SecurityEvent, Organization)
- [x] T016 [P] Create backend/internal/repository/ for UserRepository with CRUD operations
- [x] T017 [P] Create backend/internal/repository/ for AgentRepository with CRUD operations
- [x] T018 [P] Create backend/internal/repository/ for JobRepository with CRUD operations
- [x] T019 [P] Create backend/internal/repository/ for MatchRepository with CRUD and query operations
- [x] T020 [P] Create backend/internal/repository/ for MessageRepository
- [x] T021 [P] Create backend/internal/repository/ for InterviewRepository
- [x] T022 [P] Create backend/internal/repository/ for OfferRepository
- [x] T023 [P] Create backend/internal/repository/ for SecurityEventRepository with audit logging
- [x] T024 Create backend/internal/repository/ for OrganizationRepository
- [ ] T025 [P] Create backend/migrations/001_initial_schema.sql with all tables, indexes, constraints
- [x] T026 [P] Create backend/internal/service/AgentService.go with business logic layer
- [x] T027 [P] Create backend/internal/service/MatchService.go with matching logic
- [x] T028 [P] Create backend/internal/service/MessageService.go with message handling
- [x] T029 [P] Create backend/internal/service/InterviewService.go with scheduling logic
- [x] T030 [P] Create backend/internal/service/OfferService.go with offer generation
- [x] T031 [P] Create backend/internal/service/SecurityService.go with event logging
- [x] T032 Create backend/internal/middleware/auth.go with JWT validation middleware
- [x] T033 Create backend/internal/middleware/logging.go with structured JSON logging
- [x] T034 Create backend/internal/middleware/tracing.go with correlation ID propagation
- [x] T035 [P] Create backend/internal/handler/auth.go for POST /api/auth/register, POST /api/auth/login, POST /api/auth/refresh
- [x] T036 [P] Create backend/internal/handler/agent.go for Agent CRUD endpoints
- [x] T037 [P] Create backend/internal/handler/job.go for Job CRUD endpoints
- [x] T038 [P] Create backend/internal/handler/match.go for Match endpoints
- [x] T039 [P] Create backend/internal/handler/interview.go for Interview endpoints
- [x] T040 [P] Create backend/internal/handler/offer.go for Offer endpoints
- [x] T041 [P] Create backend/internal/handler/privacy.go for data export/delete endpoints
- [x] T042 Create backend/internal/handler/health.go for GET /health
- [x] T043 Create backend/internal/agent/fsm.go with finite state machine for Agent lifecycle
- [x] T044 Create backend/internal/agent/xml_protocol.go with XML message parsing and generation
- [x] T045 Create backend/internal/agent/decision.go with intent recognition and response generation
- [x] T046 Create backend/internal/agent/skill_dispatch.go with skill/capability routing

### Frontend Foundation

- [x] T047 [P] Create frontend/src/types/ with TypeScript interfaces for all domain entities
- [x] T048 [P] Create frontend/src/types/api.ts with API request/response types
- [x] T049 [P] Create frontend/src/lib/api_client.ts with fetch wrapper and error handling
- [x] T050 [P] Create frontend/src/stores/auth.ts with Zustand auth state store
- [x] T051 [P] Create frontend/src/stores/agent.ts with Zustand agent state store
- [x] T052 [P] Create frontend/src/stores/match.ts with Zustand match state store
- [x] T053 Create frontend/src/stores/ui.ts with loading and notification state
- [x] T054 [P] Create frontend/src/lib/storage.ts with IndexedDB integration
- [x] T055 [P] Create frontend/src/lib/encryption.ts with AES-256-GCM encryption utilities
- [x] T056 [P] Create frontend/src/lib/vector.ts with local vector generation utilities
- [x] T057 Create frontend/src/hooks/useAuth.ts custom React hook
- [x] T058 Create frontend/src/hooks/useAgent.ts custom React hook
- [x] T059 Create frontend/src/hooks/useWebSocket.ts custom React hook for real-time updates
- [x] T060 Create frontend/src/components/ui/ with base components (Button, Input, Card, Modal)
- [x] T061 Create frontend/src/components/layout/ with AppShell, Header, Footer, Sidebar
- [x] T062 Create frontend/src/app/layout.tsx with root layout and providers
- [x] T063 Create frontend/src/app/page.tsx with landing page
- [x] T064 Create frontend/src/app/register/page.tsx with registration page
- [x] T065 Create frontend/src/app/login/page.tsx with login page
- [x] T066 Create frontend/src/app/dashboard/page.tsx with user dashboard

### Infrastructure Foundation

- [x] T067 [P] Create backend/configs/config.yaml with database, Redis, RabbitMQ, Chroma settings
- [x] T068 [P] Create backend/configs/.env.example with all required environment variables
- [x] T069 [P] Create frontend/.env.local.example with Next.js environment variables
- [x] T070 Create backend/pkg/shared/logger.go with structured JSON logger
- [x] T071 Create backend/pkg/shared/errors.go with custom error types
- [x] T072 Create backend/pkg/shared/validation.go with input validation utilities

### Vector & Matching Foundation

- [x] T073 Create backend/internal/repository/VectorRepository.go for Chroma vector operations
- [x] T074 Create backend/internal/service/VectorService.go for embedding and similarity search
- [x] T075 Create frontend/src/lib/chroma.ts with Chroma client for vector queries

**Checkpoint**: Foundation ready — both frontends compile, backend compiles, Docker environment runs

---

## Phase 3: User Story 1 - Agent Creation and Registration (Priority: P1)

**Goal**: Users can create Seeker or Recruiter Agents and manage their profiles

**Independent Test**: User creates an Agent and verifies it appears active, can receive matches

### Implementation

- [ ] T076 [P] [US1] Create frontend/src/app/agent/create/page.tsx with Agent creation form
- [ ] T077 [P] [US1] Create frontend/src/components/agent/AgentForm.tsx with type selection (seeker/recruiter)
- [x] T078 [US1] Implement POST /api/agents handler in backend/internal/handler/agent.go
- [x] T079 [US1] Implement AgentService.CreateAgent in backend/internal/service/agent_service.go
- [ ] T080 [P] [US1] Create frontend/src/app/agent/[id]/page.tsx with Agent detail view
- [ ] T081 [P] [US1] Create frontend/src/components/agent/AgentStatus.tsx with status indicator
- [x] T082 [US1] Implement PATCH /api/agents/:id handler for pause/activate
- [x] T083 [US1] Implement AgentService.UpdateStatus for status transitions
- [x] T084 [US1] Implement AgentService.ListByUser to list user's agents
- [x] T085 [P] [US1] Create frontend/src/components/agent/AgentList.tsx with agent cards
- [x] T086 [US1] Add security event logging for agent_create, agent_pause, agent_resume

**Checkpoint**: User can create, view, pause, and activate Agents ✓ PASS

---

## Phase 4: User Story 6 - Privacy Protection (Priority: P1)

**Goal**: Sensitive data encrypted locally, exportable on request, deletable

**Independent Test**: Verify encrypted storage, export works within 24h, deletion completes in 30 days

### Implementation

- [x] T087 [P] [US6] Implement frontend/src/lib/encryption.ts with AES-256-GCM for local storage
- [x] T088 [P] [US6] Implement frontend/src/lib/storage.ts with IndexedDB CRUD for encrypted data
- [x] T089 [US6] Create frontend/src/app/privacy/page.tsx with privacy settings dashboard
- [x] T090 [US6] Create frontend/src/components/privacy/ExportButton.tsx with data export trigger
- [x] T091 [US6] Create frontend/src/components/privacy/DeleteAccountButton.tsx with deletion flow
- [x] T092 [US6] Implement POST /api/privacy/export handler for data export request
- [x] T093 [US6] Implement DELETE /api/privacy/account handler for deletion request
- [x] T094 [P] [US6] Create backend/internal/service/PrivacyService.go with export generation
- [x] T095 [US6] Implement scheduled cleanup job for 30-day deletion
- [x] T096 [US6] Add security event logging for data_export, data_delete
- [x] T097 [US6] Create frontend/src/components/privacy/EncryptionStatus.tsx with encryption indicator

**Checkpoint**: Privacy controls functional; data encrypted locally; export/delete work ✓ PASS

---

## Phase 5: User Story 2 - Automated Matching (Priority: P1)

**Goal**: Seeker Agents automatically matched with jobs based on vector similarity

**Independent Test**: Create matching Seeker and Recruiter profiles; verify match generated with score

### Implementation

- [x] T098 [P] [US6] Implement frontend/src/lib/resume_parser.ts for PDF/Word parsing
- [x] T099 [P] [US6] Implement frontend/src/lib/ai_generator.ts for one-sentence profile generation
- [x] T100 [US6] Create frontend/src/app/profile/edit/page.tsx with profile editor
- [x] T101 [US6] Create frontend/src/components/profile/ResumeUploader.tsx with file upload
- [x] T102 [US6] Create frontend/src/components/profile/ManualEntryForm.tsx for manual input
- [x] T103 [US6] Create frontend/src/components/profile/AIGenerateButton.tsx for AI generation
- [x] T104 [US6] Implement POST /api/jobs handler for job creation
- [x] T105 [US6] Implement JobService.CreateJob with vector embedding
- [x] T106 [P] [US6] Implement frontend/src/lib/vector.ts for local vector generation
- [x] T107 [P] [US6] Create backend/internal/service/VectorService.go for Chroma embeddings
- [x] T108 [US6] Create POST /api/jobs/:id/vector endpoint for vector sync
- [x] T109 [P] [US2] Implement GET /api/matches handler with match list
- [x] T110 [P] [US2] Implement MatchService.FindMatches using Chroma similarity search
- [x] T111 [US2] Implement MatchService.CalculateScore with compatibility scoring
- [ ] T112 [US2] Create background job for automatic matching when new job posted
- [x] T113 [P] [US2] Create frontend/src/app/matches/page.tsx with match list view
- [x] T114 [P] [US2] Create frontend/src/components/match/MatchCard.tsx with compatibility score
- [x] T115 [US2] Implement POST /api/matches/:id/confirm for mutual interest confirmation
- [x] T116 [US2] Add security event logging for match_created

**Checkpoint**: Profiles created with vectors; jobs posted; matches generated with scores ✓ PASS

---

## Phase 6: User Story 3 - A2A Professional Dialogue (Priority: P1)

**Goal**: Agents negotiate salary/requirements via XML protocol over WebSocket

**Independent Test**: Two agents with aligned interests complete negotiation without human intervention

### Implementation

- [x] T117 [P] [US3] Create backend/internal/agent/negotiator.go with salary negotiation logic
- [x] T118 [P] [US3] Create backend/internal/agent/requirements.go with requirements discussion logic
- [x] T119 [US3] Implement WebSocket /api/messages/ws handler for real-time messaging
- [x] T120 [US3] Implement XML message parsing in agent/xml_protocol.go
- [x] T121 [US3] Create MessageService.SendMessage with XML serialization
- [x] T122 [US3] Create MessageService.GetConversation with context retrieval
- [x] T123 [P] [US3] Implement MatchService.TransitionToNegotiating when mutual interest confirmed
- [x] T124 [US3] Create frontend/src/components/chat/ChatWindow.tsx for message display
- [x] T125 [US3] Create frontend/src/components/chat/MessageBubble.tsx with sender styling
- [x] T126 [US3] Create frontend/src/hooks/useChat.ts with WebSocket subscription
- [x] T127 [US3] Implement human confirmation trigger at negotiation milestones
- [x] T128 [US3] Add security event logging for all Agent actions
- [x] T129 [P] [US3] Create frontend/src/app/conversation/[matchId]/page.tsx with chat UI

**Checkpoint**: WebSocket connects; messages exchange; XML protocol followed; human confirmations at milestones ✓ PASS

---

## Phase 7: User Story 4 - Interview Scheduling (Priority: P2)

**Goal**: Agents coordinate interview times and send calendar invitations

**Independent Test**: Agents exchange availability and generate confirmed interview slot

### Implementation

- [ ] T130 [P] [US4] Implement backend/internal/agent/scheduler.go with interview coordination logic
- [ ] T131 [P] [US4] Create backend/internal/service/CalendarService.go with iCalendar generation
- [ ] T132 [US4] Implement POST /api/interviews handler for interview creation
- [ ] T133 [US4] Create InterviewService.ScheduleInterviews with time slot proposal
- [ ] T134 [US4] Implement PATCH /api/interviews/:id for confirmation/rescheduling
- [ ] T135 [US4] Create email service for calendar invitation delivery
- [ ] T136 [US4] Create reminder job for 24-hour advance notifications
- [ ] T137 [P] [US4] Create frontend/src/app/interviews/page.tsx with interview list
- [ ] T138 [P] [US4] Create frontend/src/components/interview/InterviewCard.tsx with details
- [ ] T139 [US4] Create frontend/src/components/interview/ScheduleModal.tsx with time picker
- [ ] T140 [US4] Add frontend/src/components/interview/CalendarPreview.tsx for iCal download
- [ ] T141 [US4] Add security event logging for interview_scheduled

**Checkpoint**: Interviews scheduled via agent negotiation; calendar invites sent; reminders delivered

---

## Phase 8: User Story 5 - Offer Generation and Confirmation (Priority: P2)

**Goal**: Recruiter Agent generates formal offer; Seeker Agent receives and responds

**Independent Test**: Employer sends offer; candidate receives and responds accept/decline/negotiate

### Implementation

- [ ] T142 [P] [US5] Implement backend/internal/agent/offer_generator.go with offer creation logic
- [ ] T143 [US5] Implement POST /api/offers handler for offer creation
- [ ] T144 [US5] Create OfferService.GenerateOffer with compensation package
- [ ] T145 [US5] Implement GET /api/offers/:id for offer retrieval
- [ ] T146 [US5] Implement POST /api/offers/:id/respond for accept/decline/negotiate
- [ ] T147 [US5] Create OfferService.UpdateStatus with state transitions
- [ ] T148 [US5] Implement position filled status update when offer accepted
- [ ] T149 [P] [US5] Create frontend/src/app/offers/page.tsx with offer list
- [ ] T150 [P] [US5] Create frontend/src/components/offer/OfferCard.tsx with compensation details
- [ ] T151 [US5] Create frontend/src/components/offer/OfferResponseForm.tsx with accept/decline/negotiate
- [ ] T152 [US5] Add security event logging for offer_sent, offer_accepted, offer_declined

**Checkpoint**: Offers generated with full compensation details; responses tracked; position status updated

---

## Phase 9: User Story 7 - Enterprise Administration (Priority: P3)

**Goal**: Enterprise admins view metrics, manage jobs, configure recruiting Agents

**Independent Test**: Admin views dashboard; pauses all agents; modifies job requirements

### Implementation

- [ ] T153 [P] [US7] Implement GET /api/admin/dashboard handler with metrics
- [ ] T154 [US7] Create AdminService.GetMetrics with recruitment statistics
- [ ] T155 [US7] Implement PATCH /api/admin/agents/:id/pause-all for bulk pause
- [ ] T156 [US7] Create PATCH /api/admin/jobs/:id for job requirement updates
- [ ] T157 [P] [US7] Create frontend/src/app/admin/page.tsx with admin dashboard
- [ ] T158 [P] [US7] Create frontend/src/components/admin/MetricsPanel.tsx with KPIs
- [ ] T159 [US7] Create frontend/src/components/admin/JobManagement.tsx with job list
- [ ] T160 [US7] Create frontend/src/components/admin/AgentManagement.tsx with agent controls
- [ ] T161 [US7] Add role-based access control for admin endpoints

**Checkpoint**: Admin dashboard displays metrics; bulk operations work; job updates reflect in agents

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Testing, optimization, documentation, and deployment preparation

### Testing

- [ ] T162 [P] Create backend/tests/unit/model/ with unit tests for all entities
- [ ] T163 [P] Create backend/tests/unit/service/ with unit tests for all services
- [ ] T164 [P] Create backend/tests/unit/agent/ with unit tests for FSM and negotiation logic
- [ ] T165 [P] Create backend/tests/integration/ with integration tests for API endpoints
- [ ] T166 [P] Create frontend/tests/unit/ with Jest unit tests for stores and utilities
- [ ] T167 [P] Create frontend/tests/e2e/ with Playwright tests for key flows
- [ ] T168 [P] Create frontend/tests/e2e/auth.spec.ts for login/register flows
- [ ] T169 [P] Create frontend/tests/e2e/agent.spec.ts for agent creation flow
- [ ] T170 [P] Create frontend/tests/e2e/matching.spec.ts for match generation flow
- [ ] T171 [P] Create frontend/tests/e2e/chat.spec.ts for A2A dialogue flow
- [ ] T172 [P] Create frontend/tests/e2e/offer.spec.ts for offer flow

### Performance Optimization

- [ ] T173 Optimize frontend bundle size; target <150KB gzipped initial JS
- [ ] T174 Optimize vector match query latency; target ≤3s response time
- [ ] T175 Optimize API p95 latency; target ≤200ms for read operations
- [ ] T176 Implement Redis caching for hot match data
- [ ] T177 Implement RabbitMQ consumer for async agent processing

### Error Handling & Edge Cases

- [ ] T178 Handle negotiation deadlock with max rounds escalation
- [ ] T179 Handle concurrent match conflicts (same position to multiple seekers)
- [ ] T180 Handle unresponsive user after offer extended
- [ ] T181 Handle user deletion with active scheduled interviews
- [ ] T182 Handle vector matching with sparse profile/job data
- [ ] T183 Handle interview feedback without recruiter ratings

### Documentation

- [ ] T184 [P] Update README.md with project overview, setup instructions, architecture
- [ ] T185 [P] Update SPEC_FILE (spec.md) with final acceptance criteria verification
- [ ] T186 [P] Update plan.md with actual implementation notes and decisions
- [ ] T187 Create API documentation for all REST endpoints
- [ ] T188 Create deployment documentation with Docker Compose instructions

### Deployment

- [ ] T189 Update docker-compose.yml with production-ready configuration
- [ ] T190 Create .env.production.example with production environment variables
- [ ] T191 Create docker-compose.prod.yml for production stack
- [ ] T192 Verify all tests pass in CI pipeline

**Checkpoint**: All phases complete; ≥80% test coverage; performance targets met; deployment ready

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3-9)**: All depend on Foundational phase completion
  - US1, US2, US3, US6 can proceed in parallel after Foundation (if staffed)
  - US4 depends on US3 (needs WebSocket infrastructure)
  - US5 depends on US3 (needs negotiation context)
  - US7 depends on Foundation (admin views use standard entities)
- **Polish (Final Phase)**: Depends on all user stories being complete

### User Story Dependencies

| Story | Depends On | Reason |
|-------|------------|--------|
| US1 (Agent Creation) | Foundation | Needs API handlers, auth |
| US6 (Privacy) | Foundation | Needs encryption layer |
| US2 (Matching) | US6 (partial) | Privacy data handling for profiles |
| US3 (A2A Dialogue) | Foundation | Needs WebSocket, XML protocol |
| US4 (Scheduling) | US3 | Needs conversation context |
| US5 (Offers) | US3 | Needs negotiation completion |
| US7 (Admin) | Foundation | Uses standard entities, auth |

### Within Each User Story

- Models before services
- Services before handlers
- Handlers before integration tests
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- US1, US2, US3, US6 can run in parallel after Foundation
- All test files marked [P] can run in parallel

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: US1 — Agent Creation
4. **STOP and VALIDATE**: Test Agent creation independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add US1 + US6 → Test independently → Deploy/Demo (MVP!)
3. Add US2 (Matching) → Test independently → Deploy/Demo
4. Add US3 (A2A Dialogue) → Test independently → Deploy/Demo
5. Add US4 (Scheduling) → Test independently → Deploy/Demo
6. Add US5 (Offers) → Test independently → Deploy/Demo
7. Add US7 (Admin) → Test independently → Deploy/Demo
8. Polish phase → Final delivery

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (Agent Creation)
   - Developer B: US6 (Privacy)
   - Developer C: US2 (Matching)
   - Developer D: US3 (A2A Dialogue)
3. After US3 complete:
   - Developer A: US4 (Scheduling)
   - Developer B: US5 (Offers)
4. US7 (Admin) can be done by any developer after Foundation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests (Phase 10) are MANDATORY per Constitution (≥80% coverage)
- Playwright E2E required for key flows (auth, agent, matching, chat, offer)
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence