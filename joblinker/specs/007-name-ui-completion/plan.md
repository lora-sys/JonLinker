# Implementation Plan: Frontend UI Completion

**Branch**: `007-name-ui-completion` | **Date**: 2026-04-28 | **Spec**: specs/007-name-ui-completion/spec.md
**Input**: Feature specification from `specs/007-name-ui-completion/spec.md`

## Summary

Implement 4 frontend features for JobLinker: (1) `/jobs/new` recruiter job posting page with Aceternity UI form, (2) `/ws-test` WebSocket debug page for real-time communication verification, (3) Agent auto-dialogue trigger via RabbitMQ+WebSocket, (4) Interview and Offer cards in conversation thread.

## Technical Context

**Language/Version**: TypeScript (Next.js 16), Go 1.21+
**Primary Dependencies**: Next.js 16, shadcn/ui, Tailwind CSS, Aceternity UI (signup-form, expandable-cards, focus-cards, background-beams), Gorilla WebSocket (backend), RabbitMQ (AMQP)
**Storage**: PostgreSQL with GORM, Chroma vector DB, IndexedDB (client-side encryption)
**Testing**: agent-browser for E2E, jest/vitest for unit
**Target Platform**: Web browser (Chrome/Firefox/Safari), responsive 320px+
**Project Type**: Full-stack web application (Next.js frontend + Go/Gin backend)
**Performance Goals**: WS message delivery <1s, page load <3s TTI, AI response <10s
**Constraints**: JWT auth via query param for WS, browser must be online for real-time
**Scale/Scope**: 4 new pages/components, 2 backend endpoints

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Type Safety | ✅ PASS | TypeScript strict mode, no `any` |
| Tests Required | ✅ PASS | agent-browser E2E for all 4 features |
| Lint Pass | ✅ PASS | Next.js lint configured |
| Loading States | ✅ PASS | All async ops must show loading |
| Error Handling | ✅ PASS | User-friendly error messages, no raw stack traces |
| Accessibility | ✅ PASS | WCAG 2.1 AA, keyboard nav, screen reader |
| Performance | ✅ PASS | WS <1s, page load <3s |

## Project Structure

### Documentation (this feature)

```text
specs/007-name-ui-completion/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (API contracts)
└── checklists/          # Quality checklists
```

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── app/
│   │   ├── jobs/
│   │   │   └── new/
│   │   │       └── page.tsx          # NEW: Job posting page
│   │   ├── ws-test/
│   │   │   └── page.tsx              # NEW: WebSocket test page
│   │   └── conversation/
│   │       └── [matchId]/
│   │           └── page.tsx          # MODIFIED: Add Interview/Offer cards
│   ├── components/
│   │   ├── ui/
│   │   │   ├── job-form.tsx          # NEW: Aceternity signup-form based
│   │   │   ├── interview-card.tsx    # NEW: Interview card component
│   │   │   └── offer-card.tsx        # NEW: Offer card component
│   │   └── ws-debug-panel.tsx        # NEW: WS debug component
│   ├── hooks/
│   │   └── use-websocket.ts         # NEW: WS connection hook
│   └── lib/
│       └── websocket.ts             # NEW: WS client utility
backend/
├── internal/
│   ├── handler/
│   │   ├── job.go                    # MODIFIED: Validate structured JSON
│   │   └── message.go                # MODIFIED: Add WS echo for /ws-test
│   └── service/
│       ├── message_service.go       # MODIFIED: Add auto-dialogue trigger
│       └── match_service.go         # MODIFIED: Add express-interest → dialogue
```

**Structure Decision**: Next.js frontend (app router) + Go/Gin backend. Job posting adds new page only (no new routes). WS test is a debug page. Interview/Offer cards modify existing conversation page. New shared components go in `components/ui/`.

## Complexity Tracking

No violations. Simpler approach rejected because:
- WS test page is necessary for verifying real-time communication works in browser
- Auto-dialogue requires RabbitMQ consumer integration (cannot be simplified away)
- Cards require new components but are minimal additions to existing conversation page

## Phase 0: Research

### Research Tasks

1. **Aceternity signup-form integration**: How to integrate Aceternity's signup-form component with shadcn/ui and existing form patterns
2. **WebSocket browser reconnection**: Best practices for WS auto-reconnect with exponential backoff in browser
3. **RabbitMQ → WebSocket relay**: How to bridge RabbitMQ consumer messages to WS push without blocking

### Research Findings (consolidated)

**Aceternity signup-form**: Install via `npx shadcn@latest add https://ui.aceternity.com/registry/signup-form-demo.json`. Built on shadcn's input/label with framer-motion. Can customize fields. Use as base for job posting form.

**WS Reconnection**: Use exponential backoff: 5s → 10s → 20s → max 60s. On reconnect, flush any pending messages. Show status indicator (Connecting/Connected/Disconnected/Reconnecting).

**RabbitMQ → WS**: Consumer goroutines publish to a channel. Hub (connection manager) reads from channel and pushes to all subscribed WS clients by matchId. Non-blocking publish from consumer.

## Phase 1: Design & Contracts

### Data Model

**Job** (extends existing):
- `id`: UUID
- `agent_id`: UUID (recruiter)
- `structured`: JSON { title, description, requirements[], nice_to_have[], location, salary_range, work_type }
- `vector_id`: UUID (Chroma)
- `status`: active/inactive
- `created_at`, `updated_at`

**Interview** (existing entity):
- `id`: UUID
- `match_id`: UUID
- `scheduled_at`: timestamp
- `format`: video | onsite | phone
- `location`: string
- `status`: pending | confirmed | cancelled
- `reminder_sent`: boolean

**Offer** (existing entity):
- `id`: UUID
- `match_id`: UUID
- `salary_amount`: integer (cents)
- `start_date`: date
- `expires_at`: timestamp
- `status`: pending | accepted | declined | expired

### API Contracts

**POST /api/jobs** (existing, validates structured field)
```json
Request: {
  "title": "string",
  "description": "string",
  "location": "string",
  "type": "full-time | part-time | contract",
  "salary_range": "string (e.g. '100k-150k')",
  "structured": "{ requirements: string[], location: string, work_type: string, salary_range: { min: number, max: number } }"
}
Response: { "id": "uuid", ... }
```

**WebSocket /api/messages/ws?token=<jwt>**
```json
// Client → Server
{ "type": "ping" }
{ "type": "message", "match_id": "uuid", "content_xml": "string", "intent_type": "string" }

// Server → Client
{ "type": "pong" }
{ "type": "message", "match_id": "uuid", "content": "string", "sender": "seeker | recruiter", "timestamp": "ISO8601" }
{ "type": "error", "message": "string" }
```

**GET /api/messages** (existing, returns conversations)
```json
Response: { "conversations": [{ "match_id": "uuid", "job_title": "string", "last_message": "string", "unread_count": number }] }
```

## Phase 2: Implementation Notes

### Priority Order

1. **Job Posting Page** (`/jobs/new`) — P1, most impactful for recruiters
2. **WS Debug Page** (`/ws-test`) — P1, enables verification of real-time
3. **WS Hook + Reconnection** — P1, foundation for auto-dialogue
4. **Interview/Offer Cards** — P2, UI polish on existing conversation page
5. **Auto-Dialogue Trigger** — P2, backend RabbitMQ integration

### Key Files to Create/Modify

**Create**:
- `frontend/src/app/jobs/new/page.tsx`
- `frontend/src/app/ws-test/page.tsx`
- `frontend/src/components/ui/job-form.tsx`
- `frontend/src/components/ui/interview-card.tsx`
- `frontend/src/components/ui/offer-card.tsx`
- `frontend/src/components/ws-debug-panel.tsx`
- `frontend/src/hooks/use-websocket.ts`
- `frontend/src/lib/websocket.ts`

**Modify**:
- `frontend/src/app/conversation/[matchId]/page.tsx` — add card rendering
- `backend/internal/handler/message.go` — add ping/pong + message echo for testing
- `backend/internal/service/message_service.go` — add auto-dialogue trigger on express-interest

### Testing Strategy

1. **Job Posting**: agent-browser → register recruiter → `/jobs/new` → fill form → submit → verify on `/jobs`
2. **WS Test**: agent-browser → `/ws-test` → verify Connected status → send test message → verify echo
3. **Cards**: Create match → trigger interview/offer status → verify card appears in conversation
4. **Auto-Dialogue**: Set match to expressed_interest → wait 10s → verify messages in thread
