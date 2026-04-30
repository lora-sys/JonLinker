# Quickstart: Frontend UI Completion

**Feature**: specs/007-name-ui-completion
**Branch**: `007-name-ui-completion`

## Prerequisites

- Next.js 16 dev server running (`cd frontend && npm run dev`)
- Go backend running (`cd backend && go run ./cmd/server`)
- Docker services: PostgreSQL, RabbitMQ, Chroma
- Aceternity UI components installed

## Installation

### Aceternity UI Components

```bash
cd frontend
npx shadcn@latest add https://ui.aceternity.com/registry/signup-form-demo.json
npx shadcn@latest add https://ui.aceternity.com/registry/expandable-cards-demo.json
npx shadcn@latest add https://ui.aceternity.com/registry/focus-cards-demo.json
npx shadcn@latest add https://ui.aceternity.com/registry/background-beams-demo.json
```

### Backend Changes Required

1. **WS echo support** — `backend/internal/handler/message.go`: Add ping/pong handling
2. **Auto-dialogue trigger** — `backend/internal/service/match_service.go`: On `expressed_interest`, publish to RabbitMQ

## Running the Features

### 1. Job Posting (`/jobs/new`)

```bash
# As recruiter user, visit:
http://localhost:3000/jobs/new
```

Form fields: title, description, location, type, salary range min/max, requirements (tags)

### 2. WebSocket Test (`/ws-test`)

```bash
# Any logged-in user:
http://localhost:3000/ws-test
```

- Shows connection status (Connecting/Connected/Disconnected/Reconnecting)
- Auto-reconnects with exponential backoff (5s → 10s → 20s → 60s max)
- Send test messages, see echoed responses

### 3. Agent Auto-Dialogue

Triggered automatically when match status → `expressed_interest`. Requires:

- RabbitMQ queues `job_seeker_queue` and `recruiter_queue` active
- Backend RabbitMQ consumer running
- WS connection established

### 4. Interview/Offer Cards

Visit any conversation where match status is `interview_scheduled` or `offer_sent`:

```bash
http://localhost:3000/conversation/<matchId>
```

## Testing

```bash
# Run backend tests
cd backend && go test ./...

# Run frontend tests
cd frontend && npm test
```

## File Inventory

| File | Status |
|------|--------|
| `frontend/src/app/jobs/new/page.tsx` | **TODO** |
| `frontend/src/app/ws-test/page.tsx` | **TODO** |
| `frontend/src/components/ui/job-form.tsx` | **TODO** |
| `frontend/src/components/ui/interview-card.tsx` | **TODO** |
| `frontend/src/components/ui/offer-card.tsx` | **TODO** |
| `frontend/src/components/ws-debug-panel.tsx` | **TODO** |
| `frontend/src/hooks/use-websocket.ts` | **TODO** |
| `frontend/src/lib/websocket.ts` | **TODO** |
| `backend/internal/handler/message.go` | **MODIFY** |
| `backend/internal/service/match_service.go` | **MODIFY** |
| `frontend/src/app/conversation/[matchId]/page.tsx` | **MODIFY** |
