# Research: Next.js App Router Server/Client Component Architecture

**Date**: 2026-04-28
**Feature**: specs/008-nextjs-architecture-refactor

## Decision 1: Component Classification Criteria

### Rule: When 'use client' is REQUIRED

A component MUST be a Client Component if it:
- Uses React hooks: `useState`, `useEffect`, `useRef`, `useCallback`, `useMemo`, `useContext`, etc.
- Uses browser APIs: `localStorage`, `sessionStorage`, `window`, `document`, `navigator`, `location`
- Uses WebSocket or Server-Sent Events
- Uses `zustand` store (client-side state)
- Has event handlers: `onClick`, `onChange`, `onSubmit`, `onKeyDown`, etc.
- Uses Next.js client hooks: `useRouter`, `useSearchParams`, `usePathname`
- Uses `RefObject` with DOM elements

### Rule: When Server Component is POSSIBLE

A component CAN be a Server Component if:
- Only renders data passed as props
- Does pure conditional rendering (no state)
- Renders static UI (no interactivity)
- Uses only shared utility functions (no hooks)
- Uses only other Server Components

### Practical Heuristic

If a component has `'use client'` at the top AND any of these conditions:
- Only uses hooks to pass data to child components
- Uses hooks but renders almost no UI directly
- Wraps a form but the form itself could be the client island

Then it should be restructured: outer shell → server, interactive parts → client islands.

---

## Decision 2: Conversation Page Architecture

### Current State (Problem)

`conversation/[matchId]/page.tsx` is likely marked `'use client'` because:
- WebSocket connection for real-time messages
- Message input state
- Scroll-to-bottom behavior

### Target State (Server Outer + Client Islands)

```
SERVER (page.tsx):
├── Fetch match data via fetch('/api/matches/:id')
├── Render conversation metadata (participants, job title, status)
├── Render message thread container shell
└── Pass message data to MessageThread

CLIENT (MessageThread.tsx):
├── WebSocket subscription for new messages
├── Auto-scroll to bottom on new messages
└── Optimistic message sending

CLIENT (ChatInput.tsx):
├── Message text state
├── Send message handler
└── Loading/error states
```

### Why This Pattern Works

1. **Fast initial load**: Server renders the shell with messages in initial HTML
2. **SEO friendly**: Conversation content visible to search engines
3. **Resumability**: Next.js can resume server-rendered content without re-fetching
4. **Progressive enhancement**: Chat works even if JS fails to load

---

## Decision 3: Shared Component Extraction Candidates

Based on codebase analysis, these patterns appear MULTIPLE times and should be extracted:

### Card Pattern (appears in jobs/page.tsx, matches/page.tsx, offers/page.tsx)
```tsx
// Current: each page has its own card markup
<div className="bg-zinc-900/80 border border-zinc-800 rounded-xl p-6">
  <h3 className="text-lg font-medium">{title}</h3>
  <p className="text-neutral-400">{description}</p>
</div>

// Target: <Card title={title} description={description} />
```

### EmptyState Pattern (appears in multiple pages)
```tsx
// Current: inline empty state markup
<div className="text-center py-12">
  <Briefcase className="h-12 w-12 mx-auto text-neutral-400" />
  <h3 className="mt-4 text-lg font-medium">No items</h3>
</div>

// Target: <EmptyState icon={Briefcase} title="No items" />
```

### LoadingSkeleton Pattern (appears in multiple pages)
```tsx
// Current: inline skeleton
<div className="animate-pulse space-y-4">
  <div className="h-4 bg-zinc-800 rounded w-3/4"></div>
  <div className="h-4 bg-zinc-800 rounded w-1/2"></div>
</div>

// Target: <LoadingSkeleton variant="card" count={3} />
```

### MessageBubble Pattern (appears in conversation/[matchId]/page.tsx)
```tsx
// Current: inline message markup with if/else for user vs agent
<div className={isUser ? "bg-cyan-500/10" : "bg-zinc-800"}>
  {content}
</div>

// Target: <MessageBubble content={content} sender={sender} type={type} />
```

---

## Decision 4: Migration Strategy

### Phase Order (to minimize conflicts)

**Phase 1**: Extract shared UI components
- No dependencies on each other
- Safe to do in parallel
- Extract: Card, Badge, Button, EmptyState, LoadingSkeleton, Avatar, Skeleton

**Phase 2**: Update all pages to use shared components
- Import path changes only
- No logic changes
- Safe batch update

**Phase 3**: Remove unnecessary 'use client'
- Start with pages that clearly don't need it
- Use build errors as guide (TypeScript will catch missing hooks)
- Work from leaves inward (leaf components → pages)

**Phase 4**: Refactor conversation page
- Most complex change
- Requires careful WebSocket hook integration

---

## Decision 5: What's NOT Changing

These are explicitly OUT OF SCOPE for this refactor:

1. **Backend Go code**: RabbitMQ handlers, WebSocket hub, agent services, vector matching — unchanged
2. **API routes**: The `/api/*` Next.js routes that proxy to Go backend — unchanged
3. **Styling approach**: 继续使用 Tailwind CSS — no CSS-in-JS migration
4. **State management library**: zustand stays — only move it to pages that truly need it
5. **WebSocket hook logic**: `use-websocket.ts` stays as client component — its nature requires it
6. **Authentication flow**: Login, register, auth store — client-only by necessity

---

## Alternatives Considered

### Alternative 1: Keep Everything as Client Components
- **Rejected because**: Violates Performance Principle (IV) — bundle size target of <150KB
- Server Components are the primary tool Next.js provides for reducing client bundle

### Alternative 2: Migrate Everything at Once
- **Rejected because**: High risk of breaking things
- Incremental migration with verification at each step is safer

### Alternative 3: Use Next.js `dynamic()` for all client components
- **Rejected because**: Over-engineered — if a component needs 'use client', just mark it
- `dynamic()` is for code-splitting specific imports, not for marking component type

---

## Current Component Inventory (Pre-refactor)

| File | 'use client'? | Reason | Can be Server? |
|------|---------------|---------|-----------------|
| app/layout.tsx | Yes | Uses AppRouter育儿 | Maybe (needs audit) |
| app/page.tsx | Yes | Uses BackgroundBeams | Check |
| app/login/page.tsx | Yes | Uses useState, useRouter | Form could be island |
| app/register/page.tsx | Yes | Uses useState, useRouter | Form could be island |
| app/dashboard/page.tsx | Yes | Uses useState, zustand | Complex |
| app/jobs/page.tsx | Yes | Needs audit | ? |
| app/jobs/new/page.tsx | Yes | Uses auth store, router | Complex (auth) |
| app/conversation/[matchId]/page.tsx | Yes | WebSocket, message state | TARGET |
| app/ws-test/page.tsx | Yes | WebSocket hook | Must stay client |
| components/layout/Header.tsx | Yes | Likely needs audit | ? |
| components/layout/Footer.tsx | Likely no | Static content | Likely |
| components/layout/Sidebar.tsx | Yes | Nav state | Must stay client |
| components/chat/ChatWindow.tsx | Yes | WebSocket, state | Must stay client |
| components/chat/MessageBubble.tsx | Likely no | Pure rendering | Likely |
| components/ui/job-form.tsx | Yes | Form state, API call | Must stay client |
| components/ui/interview-card.tsx | Likely yes | Event handlers | Check |
| components/ui/offer-card.tsx | Likely yes | Event handlers | Check |

*Full inventory to be completed in Phase 0 implementation.*
