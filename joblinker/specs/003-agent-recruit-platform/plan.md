# Implementation Plan: A2A Agent Recruitment Platform

**Branch**: `003-agent-recruit-platform` | **Date**: 2026-04-22 | **Spec**: specs/003-agent-recruit-platform/spec.md
**Input**: Feature specification and PRD from prd.md

## Summary

An A2A (Agent-to-Agent) intelligent recruitment platform where autonomous Seeker Agents and Recruiter Agents communicate via XML protocol to automate job matching, professional dialogue, interview scheduling, and offer confirmation. Privacy-first architecture: sensitive user data encrypted and stored locally (IndexedDB + AES-256), only anonymized vectors transmitted to server. Full-stack Next.js 16 frontend with Go backend, PostgreSQL for business data, Chroma for vector metadata, Redis for caching, RabbitMQ for task orchestration.

## Technical Context

**Language/Version**: TypeScript 5.7+ (frontend), Go 1.23+ (backend)
**Primary Dependencies**:
- Frontend: Next.js 16, Zustand (state), Tailwind CSS v5 (styling), Aceternity UI (components)
- Backend: Gin (HTTP framework), GORM (ORM)
- Data: PostgreSQL 17 (business data), Chroma 0.6+ (vector search), Redis 7.4+ (cache), RabbitMQ (task queue)
- Agent: Custom FSM state machine, XML protocol
**Storage**: PostgreSQL 17 (server), IndexedDB + AES-256 (client-local privacy data)
**Testing**: Playwright (E2E), Go testing package (unit/integration)
**Target Platform**: Web browsers (desktop + mobile), Linux server (backend)
**Project Type**: Full-stack web application with autonomous Agent engine
**Performance Goals**:
- Page load TTI: ≤ 2 seconds
- Agent dialogue response: ≤ 1 second
- Vector match response: ≤ 3 seconds
- API p95 read latency: ≤ 200ms
**Constraints**: All raw privacy data stays local; server receives only anonymized vectors and structured metadata
**Scale/Scope**: MVP targets 100 concurrent users, 10k job postings, 50k candidate profiles

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status |
|-----------|-------------|--------|
| Test-First | Unit tests for all business logic; integration tests for API contracts; ≥80% coverage | ✅ Plan includes test phases |
| Code Quality | Static typing (TypeScript strict, Go types); lint rules enforced; SRP modules | ✅ TypeScript + Go both typed |
| UX Consistency | Responsive (mobile 320px+); WCAG 2.1 AA; loading states; user-friendly errors | ✅ PRD specifies these targets |
| Performance | TTI ≤3s; Core Web Vitals LCP<2.5s/FID<100ms/CLS<0.1; API ≤200ms p95 | ✅ PRD specifies ≤2s TTI, ≤1s Agent response |
| Observability | Structured JSON logs; metrics; tracing; /health endpoint | ✅ Backend observability part of Phase 4 |
| Security | Secrets in env vars; input validation; token expiration; dependency audit | ✅ Privacy architecture local-first |
| Privacy | Sensitive data encrypted at rest/in transit; local storage for raw data | ✅ Core architectural constraint |
| Deployment | Docker + Docker Compose; repeatable automation | ✅ PRD Phase 6 specifies containerization |

**Gate Status**: ✅ PASS — All constitution principles addressed in plan

## Project Structure

### Documentation (this feature)

```text
specs/003-agent-recruit-platform/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── a2a-protocol.md  # XML communication contract
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
# Frontend (Next.js 16 + TypeScript)
frontend/
├── src/
│   ├── app/                    # Next.js App Router pages
│   ├── components/             # Reusable UI components
│   ├── pages/                  # Pages (if using Pages Router)
│   ├── lib/                    # Utilities, API clients
│   ├── stores/                 # Zustand state stores
│   ├── hooks/                  # Custom React hooks
│   └── types/                  # TypeScript type definitions
├── public/
├── tests/
│   ├── e2e/                    # Playwright E2E tests
│   └── unit/                   # Unit tests
└── package.json

# Backend (Go + Gin)
backend/
├── cmd/
│   └── server/                 # Main entry point
├── internal/
│   ├── handler/                # HTTP handlers (presentation)
│   ├── service/                # Business logic layer
│   ├── repository/             # Data access layer (GORM)
│   ├── model/                  # Domain entities
│   ├── agent/                  # FSM engine, XML protocol
│   └── middleware/             # Auth, logging, tracing
├── pkg/
│   └── shared/                 # Shared utilities
├── migrations/                 # Database migrations
├── configs/                    # Configuration files
├── tests/
│   ├── integration/            # Integration tests
│   └── unit/                   # Unit tests
└── go.mod

# Infrastructure
├── docker-compose.yml          # Local dev environment
├── Dockerfile.frontend
├── Dockerfile.backend
└── infra/
    └── chroma/                 # Chroma volume (if persistent)
```

**Structure Decision**: Full-stack web application with clear frontend/backend separation. Frontend: Next.js 16 with native TypeScript/HTML/CSS. Backend: Go with layered architecture (Handler/Service/Repo). Infrastructure: Docker Compose for PostgreSQL, Redis, RabbitMQ, Chroma.

## Development Phases (from PRD)

### Phase 1: Foundation (1-2 weeks)
- Frontend: Next.js 16 init, routing, base layout
- Backend: Go + Gin layered architecture setup
- Data: PostgreSQL 17 + Chroma deployment
- Infra: Redis + RabbitMQ setup
- Auth: Registration, login, JWT implementation

### Phase 2: Local Privacy & Info Generation (2 weeks)
- Frontend: IndexedDB + AES-256 local storage
- Info input: Manual, AI generation, PDF/Word parsing
- Local vector generation (data stays on device)
- Pages: Profile entry, privacy settings, dashboard

### Phase 3: Agent Core Engine (3 weeks)
- FSM state machine for Agent lifecycle
- XML communication protocol
- Agent decision logic (intent recognition, response generation)
- Skill dispatch module

### Phase 4: Core Business Flow (3 weeks)
- Vector matching with Chroma
- WebSocket real-time A2A dialogue
- Interview scheduling with calendar invites
- Offer generation and status tracking

### Phase 5: Testing & Polish (2 weeks)
- Unit tests ≥85% coverage
- E2E tests for key flows
- Performance optimization
- Error handling and edge cases

### Phase 6: Deploy & Deliver (1 week)
- Docker + Docker Compose
- Multi-environment config
- Documentation
- Demo and acceptance

## Complexity Tracking

No complexity violations requiring justification. Architecture follows standard full-stack web patterns with appropriate separation. Local-first privacy is a core architectural constraint aligned with constitution principles.

## Phase 0: Research

Research is largely complete from the provided PRD. Key technical decisions documented:

| Decision | Rationale | Alternatives Considered |
|----------|-----------|------------------------|
| Next.js 16 frontend | App Router, React Server Components, fast builds | Plain React, Remix |
| Go + Gin backend | Performance, strong typing, mature HTTP ecosystem | Rust/Axia, Node.js |
| PostgreSQL + Chroma | SQL for business data, vectors for matching | Pinecone, Weaviate |
| IndexedDB + AES-256 client-side | Privacy-first: raw data never leaves device | Server-side encryption only |
| XML for A2A protocol | Structured, parseable, extensible | JSON, Protobuf |
| RabbitMQ for orchestration | Reliable async task delivery | Direct WebSocket, Kafka |
| Zustand state management | Minimal, TypeScript-native | Redux, Jotai |
| Playwright for E2E | Supports Next.js, good DX | Cypress, Selenium |

**Output**: research.md — No additional research needed; PRD provides sufficient technical direction.

## Phase 1: Design & Contracts

### Data Model (data-model.md)

Entities derived from spec requirements:

- **User**: id, email, password_hash, role, organization_id, created_at, updated_at
- **Agent**: id, user_id, type (seeker|recruiter), status (active|paused), config_json, created_at
- **Resume**: id, agent_id, structured_json, vector_id (Chroma), created_at, updated_at
- **Job**: id, agent_id, structured_json, vector_id (Chroma), status, created_at, updated_at
- **Match**: id, seeker_agent_id, job_id, score, status, created_at, updated_at
- **Message**: id, match_id, sender_agent_id, content_xml, created_at
- **Interview**: id, match_id, scheduled_at, format, location, status, created_at
- **Offer**: id, match_id, compensation_json, start_date, status, created_at, updated_at
- **SecurityEvent**: id, user_id, action_type, details_json, ip_address, created_at
- **Organization**: id, name, admin_user_id, created_at

Relationships: User 1:N Organization, User 1:N Agent, Agent 1:1 Resume/Job, SeekerAgent N:1 Match N:1 Job, Match 1:N Message, Match 1:1 Interview, Match 1:1 Offer

### Contracts

**A2A XML Protocol** (contracts/a2a-protocol.md):
- Message envelope: `<a2a-message type="negotiate|schedule|offer" id="..." timestamp="...">`
- Intent identification: `<intent type="salary|schedule|requirements" confidence="0.0-1.0">`
- Response format: `<a2a-response ref-id="..." status="accepted|rejected|counter">`

**API Endpoints** (REST + WebSocket):
- Auth: POST /api/auth/register, POST /api/auth/login, POST /api/auth/refresh
- Agents: GET/POST /api/agents, GET/PATCH/DELETE /api/agents/:id
- Jobs: GET/POST /api/jobs, GET/PATCH /api/jobs/:id
- Matches: GET /api/matches, POST /api/matches/:id/confirm
- Messages: WebSocket /api/messages/ws
- Interviews: GET/POST /api/interviews, PATCH /api/interviews/:id
- Offers: GET/POST /api/offers/:id/respond
- Privacy: POST /api/privacy/export, DELETE /api/privacy/account
- Health: GET /health

**Output**: data-model.md, contracts/a2a-protocol.md, quickstart.md