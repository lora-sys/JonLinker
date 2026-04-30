# Data Model: Next.js Component Taxonomy

**Date**: 2026-04-28
**Feature**: specs/008-nextjs-architecture-refactor

## Component Taxonomy

### Category: Shared UI Primitives (`components/ui/`)

Base components usable anywhere in the application. Server-safe unless noted.

| Component | File | Type | Description | Props |
|-----------|------|------|-------------|-------|
| Button | Button.tsx | CLIENT | Interactive button with variants | `variant`, `size`, `disabled`, `onClick`, `children` |
| Card | Card.tsx | SERVER | Container card with title/description/actions | `title`, `description`, `metadata`, `actions`, `children` |
| Badge | Badge.tsx | SERVER | Status badge with color variants | `variant`, `children` |
| EmptyState | EmptyState.tsx | SERVER | Empty state placeholder with icon | `icon`, `title`, `description`, `action` |
| LoadingSkeleton | LoadingSkeleton.tsx | CLIENT | Animated loading placeholder | `variant`, `count`, `className` |
| Avatar | Avatar.tsx | SERVER | User avatar with fallback initials | `src`, `name`, `size`, `status` |
| Skeleton | Skeleton.tsx | CLIENT | Generic animated skeleton | `className`, `width`, `height` |
| Modal | Modal.tsx | CLIENT | Overlay modal dialog | `open`, `onClose`, `children`, `title` |
| Input | input.tsx | CLIENT | Form input (shadcn) | Standard HTML input props |
| Label | label.tsx | SERVER | Form label (shadcn) | Standard label props |

### Category: Layout Components (`components/layout/`)

Page structure components.

| Component | File | Type | Description | Notes |
|-----------|------|------|-------------|-------|
| Header | Header.tsx | SERVER | Top navigation bar | Mostly static, user menu could be client island |
| Footer | Footer.tsx | SERVER | Page footer | Pure static |
| Sidebar | Sidebar.tsx | CLIENT | Navigation sidebar | Has collapse state, must be client |
| AppShell | AppShell.tsx | CLIENT | Page wrapper with nav | Has auth-dependent rendering |

### Category: Feature Components (`components/feature/`)

Domain-specific components extracted from pages.

| Component | File | Type | Description | Source Page |
|-----------|------|------|-------------|-------------|
| JobCard | JobCard.tsx | SERVER | Job listing card | jobs/page.tsx |
| MatchCard | MatchCard.tsx | SERVER | Match summary card | matches/page.tsx |
| OfferCard | OfferCard.tsx | CLIENT | Offer display with actions | offers/page.tsx |
| InterviewCard | InterviewCard.tsx | CLIENT | Interview display with actions | interviews/page.tsx |
| MessageBubble | MessageBubble.tsx | SERVER | Chat message | conversation/[matchId]/page.tsx |
| ChatInput | ChatInput.tsx | CLIENT | Message input with send | conversation/[matchId]/page.tsx |

### Category: Client-Only Hooks (`hooks/`)

React hooks that MUST remain client-side.

| Hook | File | Reason |
|------|------|--------|
| useWebSocket | use-websocket.ts | WebSocket API, browser-only |
| useOutsideClick | use-outside-click.tsx | DOM event listeners |
| useIntersectionReveal | use-intersection-reveal.ts | IntersectionObserver API |
| useRole | use-role.ts | Uses zustand auth store |
| useAuth | auth.ts (store) | zustand persist, localStorage |

### Category: Page Components (`app/`)

Next.js App Router pages and their target architecture.

| Page | Current | Target | Notes |
|------|---------|--------|-------|
| /page.tsx | CLIENT | SERVER | Homepage, static content |
| /layout.tsx | CLIENT | SERVER | Root layout, mostly static |
| /login/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Form is client island |
| /register/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Form is client island |
| /dashboard/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Overview data, client widgets |
| /jobs/page.tsx | CLIENT | SERVER | Job listings, static render |
| /jobs/new/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Form is client island |
| /agents/page.tsx | CLIENT | SERVER | Agent list |
| /matches/page.tsx | CLIENT | SERVER | Match list |
| /messages/page.tsx | CLIENT | SERVER | Message list |
| /interviews/page.tsx | CLIENT | SERVER | Interview list |
| /offers/page.tsx | CLIENT | SERVER | Offer list |
| /conversation/[matchId]/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Full refactor target |
| /ws-test/page.tsx | CLIENT | CLIENT | Must stay client (WebSocket) |
| /settings/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Settings forms |
| /admin/page.tsx | CLIENT | SERVER | Admin panel |
| /privacy/page.tsx | CLIENT | SERVER | Privacy page |
| /resume/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Resume page |
| /resume/generate/page.tsx | CLIENT | SERVER+CLIENT_ISLAND | Generate page |

## Component Dependency Graph

```
app/layout.tsx (SERVER)
└── AppShell (CLIENT) - has auth-dependent nav
    ├── Header (SERVER) - static
    ├── Sidebar (CLIENT) - has collapse state
    └── children (page content)

app/conversation/[matchId]/page.tsx (SERVER outer)
├── ConversationHeader (SERVER) - match metadata
├── MessageThread (CLIENT) - WebSocket, scroll
│   └── MessageBubble (SERVER) - pure render
└── ChatInput (CLIENT) - message input
```

## State Ownership

| State | Owner | Location | Type |
|-------|-------|----------|------|
| Auth token | Client | stores/auth.ts | Zustand persist |
| WebSocket connection | Client | hooks/use-websocket.ts | React state |
| Message list | Client (WebSocket) | conversation page | useState |
| UI state (modals, dropdowns) | Client | Individual components | useState |
| Page data (jobs, matches) | Server fetch | page.tsx | Server Component |

## Shared Component Props Interfaces

```typescript
// Card
interface CardProps {
  title?: string;
  description?: string;
  metadata?: React.ReactNode;
  actions?: React.ReactNode;
  children?: React.ReactNode;
  className?: string;
}

// EmptyState
interface EmptyStateProps {
  icon?: React.ComponentType<{ className?: string }>;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

// MessageBubble
interface MessageBubbleProps {
  content: string;
  sender: 'user' | 'agent' | 'system';
  timestamp?: Date;
  senderName?: string;
}

// LoadingSkeleton
interface LoadingSkeletonProps {
  variant?: 'text' | 'card' | 'list';
  count?: number;
  className?: string;
}
```
