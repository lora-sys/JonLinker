# Research: Frontend UI Completion

**Feature**: specs/007-name-ui-completion
**Date**: 2026-04-28

## Decision: Aceternity signup-form for Job Posting

**Decision**: Use Aceternity's signup-form as base for job posting form.

**Rationale**: Aceternity's signup-form is built on shadcn's input/label with framer-motion animations. It provides a polished starting point that can be customized with additional fields. The form has built-in validation display and is already compatible with the project's shadcn/ui setup.

**Alternatives considered**:
- Build form from scratch with shadcn/ui primitives — more work, inconsistent with design system
- Use plain HTML + Tailwind — lacks animation polish, inconsistent with project aesthetic

**Install command**: `npx shadcn@latest add https://ui.aceternity.com/registry/signup-form-demo.json`

---

## Decision: WS Auto-Reconnect with Exponential Backoff

**Decision**: Implement WS reconnection with exponential backoff: 5s → 10s → 20s → max 60s.

**Rationale**: Exponential backoff is the standard pattern for WS reconnection. It prevents hammering the server on persistent failures while eventually succeeding on transient issues. The 60s max prevents unbounded growth.

**Alternatives considered**:
- Fixed interval (e.g., every 5s) — wastes resources on repeated failures
- No reconnection — forces manual user action, poor UX

**Implementation**: use-websocket.ts hook manages connection state machine:
- States: Connecting | Connected | Disconnected | Reconnecting
- On disconnect: enter Reconnecting, schedule retry with backoff
- On retry success: enter Connected, flush pending queue
- Max 5 retries before showing "Connection failed" with manual retry button

---

## Decision: RabbitMQ → WebSocket Bridge via Go Channel

**Decision**: RabbitMQ consumers publish to a buffered Go channel. Hub reads and pushes to WS clients by matchId.

**Rationale**: Non-blocking bridge pattern. Consumer publishes to channel and returns immediately. Hub goroutine handles WS delivery asynchronously. This prevents slow consumers from blocking the queue.

**Alternatives considered**:
- Consumer directly writes to WS — blocks on slow clients
- Separate pub/sub for each match — complex connection management

**Implementation**:
```go
// Hub manages WS connections and message routing
type Hub struct {
    clients    map[string]map[*Client]bool  // matchId → clients
    broadcast  chan Message
    register  chan *Client
    unregister chan *Client
}

// Consumer publishes to hub.broadcast
// Hub fans out to relevant matchId clients
```

---

## Decision: Interview/Offer Cards as Inline Conversation Components

**Decision**: Render interview and offer cards inline within the conversation thread, above the message list.

**Rationale**: Keeps all context in one place. Cards are visually distinct but contextually grouped with the conversation they're about. Natural scroll position puts pending actions in view.

**Alternatives considered**:
- Separate /interviews and /offers pages — requires navigation, loses conversation context
- Modal/dialog popup — interrupts conversation flow, jarring UX

**Implementation**: Conversation page fetches match status, renders card components conditionally:
```tsx
{match.status === 'interview_scheduled' && <InterviewCard interview={interview} />}
{match.status === 'offer_sent' && <OfferCard offer={offer} />}
```

---

## Notes

- Aceternity UI components used: signup-form (job form base), expandable-cards (potential future use), background-beams (page backgrounds)
- WS test page (`/ws-test`) is a debug/dev tool, not a user-facing feature
- All cards use loading states and error handling per constitution requirements
