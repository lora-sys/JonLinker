# Feature Specification: Next.js App Router Architecture Refactor

**Feature Branch**: `008-nextjs-architecture-refactor`
**Created**: 2026-04-28
**Status**: Draft
**Input**: User description: "全面扫描并重构整个 Next.js 项目，严格遵循 App Router 最佳实践：删除不必要 'use client'，仅保留强交互模块为客户端组件，重构对话页面/聊天线程/会话列表/消息面板，批量抽取可复用公共组件，统一组件命名和目录结构"

## User Scenarios & Testing

### User Story 1 - Server/Client Component Separation (Priority: P1)

Developer or maintainer reviews the codebase and identifies all components that can be converted from client to server components. Pages that don't need browser APIs, user interaction, or state management are converted to Server Components, reducing client-side JavaScript bundle size.

**Why this priority**: P1 - This is the foundational refactor that enables all other improvements. Incorrectly placing 'use client' is a common performance anti-pattern in Next.js.

**Independent Test**: Run `npm run build` and verify build succeeds with no TypeScript errors. Bundle size should decrease compared to baseline.

**Acceptance Scenarios**:

1. **Given** a page component without browser API usage, **When** the page is reviewed, **Then** it should be a Server Component without 'use client' directive
2. **Given** a component with WebSocket usage, **When** it's reviewed, **Then** it should be marked 'use client'
3. **Given** a page with mixed server/client rendering needs, **When** it's structured, **Then** the outer layout should be server with client islands for interactive parts

---

### User Story 2 - Conversation/Chat Architecture Refactor (Priority: P2)

The conversation pages and message threads are restructured using the "server outer layout + client inner interaction" pattern. The outer shell (header, conversation list, message thread container) renders on the server for fast initial load and SEO, while interactive elements (message input, real-time updates, scroll behavior) are isolated client components.

**Why this priority**: P2 - These pages are performance-critical and frequently used; they benefit most from the layered architecture.

**Independent Test**: Open a conversation page, verify messages load quickly, WebSocket reconnection shows new messages in real-time, input field accepts and sends messages.

**Acceptance Scenarios**:

1. **Given** a user opens a conversation page, **When** the page loads, **Then** the conversation thread should display within 2 seconds with server-side rendered structure
2. **Given** a new message arrives via WebSocket, **When** the user is viewing the conversation, **Then** the message should appear in real-time without page refresh
3. **Given** a user scrolls through long message history, **When** they scroll, **Then** the scroll position should be maintained and messages lazy-loaded as needed

---

### User Story 3 - Shared Component Extraction (Priority: P2)

Common UI elements are identified, extracted, and consolidated into a shared components library. Cards, list items, message bubbles, empty states, loading indicators, buttons, dialogs, pagination, and date formatting utilities are standardized and reused across pages, eliminating duplication.

**Why this priority**: P2 - Reduces maintenance burden and ensures consistency across the application.

**Independent Test**: Search for duplicate implementations of the same UI pattern across different files - none should exist after refactoring. All instances should use shared components.

**Acceptance Scenarios**:

1. **Given** a new card component is needed, **When** a developer looks at the shared components directory, **Then** a reusable Card component exists with consistent styling
2. **Given** an empty state is displayed, **When** the user sees it, **Then** it should use a consistent EmptyState component across all pages
3. **Given** loading indicators are shown, **When** data is being fetched, **Then** they should all use the same LoadingSkeleton or Spinner component

---

### User Story 4 - Project Structure Normalization (Priority: P3)

Component naming conventions, directory structure, and import patterns are standardized. The codebase follows a consistent pattern for file organization (e.g., `components/ui/`, `components/layout/`, `components/features/`) and naming (`PascalCase` for components, `camelCase` for utilities).

**Why this priority**: P3 - Important for long-term maintainability but doesn't affect user-facing functionality.

**Independent Test**: All components follow the same naming convention and directory structure. Imports are consistent (absolute paths via `@/` alias).

**Acceptance Scenarios**:

1. **Given** a developer adds a new component, **When** they place it in the correct directory, **Then** it follows the established naming convention
2. **Given** imports are used across files, **When** they're reviewed, **Then** they should use `@/` alias consistently without relative path chains like `../../..`

---

## Requirements

### Functional Requirements

- **FR-001**: All pages that don't use browser APIs, user events, state, or WebSocket MUST be Server Components (no 'use client' directive)
- **FR-002**: Only the following types of modules SHOULD retain 'use client': WebSocket hook, real-time message renderer, chat input component, auth state store, scroll listeners, event handlers
- **FR-003**: Conversation pages MUST use server component for outer layout (conversation metadata, message thread container) and client component for inner interactive elements (input, real-time updates)
- **FR-004**: All shared UI primitives MUST be extracted: Card, ListItem, MessageBubble, EmptyState, LoadingSkeleton, Button, Dialog, Pagination, DateTime formatting
- **FR-005**: Component directory structure MUST follow: `components/ui/` (base primitives), `components/layout/` (page shells), `components/feature/` (domain-specific)
- **FR-006**: All imports MUST use `@/` alias consistently; no multi-level relative paths like `../../..`
- **FR-007**: Original business functionality MUST be preserved: RabbitMQ message processing, WebSocket real-time communication, multi-agent dialogue, vector matching
- **FR-008**: Build MUST succeed with `npm run build` producing no TypeScript or lint errors

### Key Entities

- **Page Component**: The Next.js App Router page.tsx file - either a Server Component (default) or Client Component ('use client')
- **Shared UI Component**: A reusable component extracted for use across multiple pages (e.g., Card, Button, EmptyState)
- **Client Island**: A small, focused 'use client' component embedded within a server-rendered page for specific interactivity
- **Conversation Thread**: The message list within a conversation, combining server layout + client real-time updates

## Success Criteria

### Measurable Outcomes

- **SC-001**: Client-side JavaScript bundle size decreases by at least 20% compared to baseline (measured via `npm run build` output)
- **SC-002**: All pages render initial HTML within 1 second on a standard connection (measured via browser devtools)
- **SC-003**: Conversation pages use server-side rendering for structure and client-side for interactivity - verified via View Source showing server-rendered content
- **SC-004**: Zero duplicate UI component implementations exist (e.g., if Card component exists, no page should have its own card markup)
- **SC-005**: Build completes successfully with no TypeScript errors, no lint warnings in modified files
- **SC-006**: WebSocket functionality, RabbitMQ processing, and agent dialogue continue to work as before (verified via manual testing or existing E2E tests)

## Assumptions

- Users access the application via modern browsers that support React Server Components (Chrome 85+, Safari 15.4+, Firefox 90+)
- The existing API routes (Go backend) are unaffected by this refactor - only frontend component architecture changes
- WebSocket connection handling is already implemented in hooks and will be preserved as client-side code
- No changes to the actual styling (CSS classes, Tailwind utilities) are needed - only component boundaries and architecture
- The `zustand` auth store will remain a client component but its usage will be minimized to only pages that need client-side auth checks
