# Frontend Refactoring + Integration Testing

## Feature Name
frontend-refactor-playwright

## Overview
The JobLinker frontend application has accumulated critical bugs, architectural inconsistencies, and dead code that prevent reliable operation. This refactoring fixes authentication failures, unifies the application layout, modernizes the chat system with streaming AI capabilities, redesigns the conversation interface for dual-agent workflows, and validates everything through browser-based integration testing.

## Problem Statement

### Current State
- **Authentication is broken on key pages**: The agents page reads a cookie name that no page ever sets, causing perpetual "not authenticated" errors. The admin page passes raw JSON as a Bearer token instead of extracting the token string.
- **No shared layout exists**: Every page independently renders its own full-page layout. There is no sidebar, header, or navigation consistency. The AppShell component exists but is never used.
- **The conversation page uses browser `alert()` dialogs** for milestone confirmations and casts types with `as any`, bypassing type safety.
- **The chat system uses a custom streaming implementation** that duplicates functionality now available in the AI SDK, with character-by-character animation via requestAnimationFrame.
- **Dead code accumulates**: Vector math utilities (TF-IDF), ChromaDB client, and a WebSocket debug panel are all unused.
- **The messages page has a hardcoded unread count of 0** and constructs WebSocket URLs inconsistently.
- **Sidebar navigation items are rendered as styled `<div>` elements** that do not actually navigate when clicked (except Logout).

### Impact
- Users cannot access the agents page (authentication always fails)
- Users have no consistent navigation across the application
- The conversation experience is fragmented with no visual distinction between agents
- Type safety is compromised in the conversation flow
- The codebase carries ~500 lines of dead code

## User Scenarios & Testing

### Scenario 1: New User Registration and First Agent
**Actor**: New job seeker
**Flow**: Register account -> Land on dashboard -> Navigate to agents page via sidebar -> Create first agent -> Verify agent appears in list
**Expected**: Seamless flow from registration through agent creation with consistent navigation visible throughout.

### Scenario 2: Authenticated Conversation with AI
**Actor**: Registered user with an active match
**Flow**: Log in -> Navigate to conversation via sidebar -> Send a message -> Receive streaming AI response -> See FSM stage progression in side panel -> Receive interview or offer card
**Expected**: Dual-column conversation view with messages on the left, process flow on the right. Two agents are visually distinct. Messages stream in naturally. The FSM stage and tool calls are visible in real-time.

### Scenario 3: Cross-Page Navigation
**Actor**: Authenticated user
**Flow**: From any page, click sidebar items (Dashboard, Agents, Jobs, Matches, Messages, Resume, Settings) -> Each click navigates to the correct page -> On mobile, sidebar collapses after navigation
**Expected**: Every sidebar item navigates reliably. Active page is highlighted. Mobile sidebar closes on navigation.

### Scenario 4: Unauthenticated Access Protection
**Actor**: Visitor or logged-out user
**Flow**: Clear cookies/localStorage -> Attempt to access /dashboard, /agents, /messages, etc.
**Expected**: All protected pages redirect to /login. Public pages (landing, login, register, privacy) remain accessible.

### Scenario 5: Mobile Responsive Experience
**Actor**: Mobile user
**Flow**: Access application on 375px viewport -> Navigate through pages -> Enter conversation view
**Expected**: Sidebar is hidden by default (hamburger toggle). No horizontal overflow. Conversation view adapts to single-column on narrow screens.

## Functional Requirements

### FR-1: Authentication Consistency
All server-side components that read authentication state MUST use the same cookie name and parsing logic. The token extraction MUST consistently parse the JSON cookie value and extract the token field, not pass the raw cookie value as a credential.

### FR-2: Unified Application Layout
All authenticated pages MUST share a common layout that includes a header and sidebar navigation. Public pages (landing, login, register, privacy) MUST NOT include the sidebar. The layout MUST be responsive, collapsing the sidebar on mobile viewports.

### FR-3: Sidebar Navigation
Every sidebar navigation item MUST navigate to its target page when clicked. On mobile viewports, clicking a navigation item MUST also close the sidebar. The currently active page MUST be visually indicated.

### FR-4: Notification System
User-facing confirmations (milestone confirmed, resume saved, etc.) MUST use in-page notification components, not browser `alert()` dialogs. Notifications MUST be dismissible and auto-timeout.

### FR-5: Type Safety
All component props MUST use properly typed interfaces. No `as any` type assertions are permitted in page components or their direct children.

### FR-6: Conversation Page Layout
The conversation page MUST display a dual-column layout: message area on the left (65% width) and a process flow panel on the right (35% width). On viewports below 1024px, the flow panel MUST collapse into an accessible overlay or sheet.

### FR-7: Agent Visual Differentiation
In the conversation view, messages from the Recruiter Agent and Seeker Agent MUST be visually distinct through color scheme, alignment, and agent identification (name, avatar, timestamp).

### FR-8: Process Flow Panel
The right-side panel MUST display: the current FSM negotiation stage (e.g., Introduction, Negotiating, Interviewing, Offer), a visual progress indicator showing completed/current/upcoming stages, a history of tool calls with status indicators, and any Interview or Offer cards.

### FR-9: Chat Input
The message input MUST be a multi-line textarea that supports Shift+Enter for newlines and Enter to send. It MUST auto-resize between 1 and 5 lines. Send, Stop, and Retry controls MUST be available.

### FR-10: AI Response Streaming
AI responses MUST stream progressively to the user, appearing word-by-word or chunk-by-chunk rather than waiting for the complete response. The streaming implementation MUST support interruption (Stop button).

### FR-11: Real-Time Updates
The conversation MUST receive real-time updates for: new messages from the A2A system, FSM state transitions, tool call lifecycle events, and new interview/offer creation. These arrive via WebSocket and MUST be reflected in the UI without page refresh.

### FR-12: Virtual Scrolling for Messages
The message list MUST use virtual scrolling to handle long conversation histories efficiently. Only visible messages plus a small buffer are rendered. New messages MUST auto-scroll to bottom. Scrolling up MUST load earlier messages.

### FR-13: Dead Code Removal
The following unused modules MUST be removed: TF-IDF vector math utilities, ChromaDB client, WebSocket debug panel. Any dependency exclusively used by removed code MUST also be removed.

### FR-14: Messages Page Unread Count
The messages page MUST display the actual unread message count from the backend API, not a hardcoded zero.

### FR-15: Messages Page WebSocket
The messages page MUST connect to the correct backend WebSocket endpoint for real-time conversation list updates, or fall back to REST API polling if WebSocket is unavailable.

### FR-16: Token Management
A single utility module MUST provide token access for both client-side (from state management) and server-side (from cookies) contexts. All components MUST use this utility rather than directly accessing cookies or localStorage.

## Success Criteria

### SC-1: Authentication Reliability
Every authenticated page successfully loads user-specific data without "not authenticated" errors. Zero cookie name mismatches across the codebase.

### SC-2: Navigation Consistency
100% of sidebar items navigate to the correct page. Active page indicator works on all pages. Mobile sidebar closes on navigation.

### SC-3: Build Success
The application compiles with zero errors. No `as any` assertions remain. No `alert()` calls remain in application code. No references to deleted dead code modules exist.

### SC-4: Conversation Usability
Users can send messages and receive streaming AI responses. Two agents are visually distinguishable. The FSM stage and progress are visible. Tool calls display with status indicators.

### SC-5: Integration Test Coverage
All 9 integration test scenarios pass: registration, login, sidebar navigation, agent CRUD, conversation layout, message sending, auth protection, responsive layout, and full end-to-end flow.

### SC-6: Code Quality
Dead code (vector.ts, chroma.ts, ws-debug-panel.tsx) is removed. Duplicate cookie parsing logic is consolidated. The custom streaming hook is replaced by the AI SDK built-in streaming.

## Scope

### In Scope
- Fix cookie name mismatches across all server components
- Fix sidebar navigation to actually navigate
- Replace all alert() calls with toast notifications
- Remove all `as any` type assertions
- Create unified dashboard layout with AppShell
- Consolidate token management into a single utility
- Delete dead code modules and unused dependencies
- Create Vercel AI SDK streaming endpoint
- Replace custom chat hooks with AI SDK useChat
- Redesign conversation page with dual-column layout
- Add FlowPanel, AgentMessageBubble, ConversationHeader, TextareaInput components
- Implement virtual scrolling for message list
- Fix messages page WebSocket URL and unread count
- Execute 9 integration test scenarios via Playwright browser automation

### Out of Scope
- Backend API changes or new endpoints
- Mobile native application
- End-to-end testing framework setup (test files)
- Performance benchmarking
- Accessibility audit (beyond basic keyboard navigation)
- Internationalization
- Dark mode or theming changes

## Key Entities

| Entity | Description |
|--------|-------------|
| User | Authenticated person with email, password, role (seeker/recruiter) |
| Agent | AI agent created by a user, with name, role, and capabilities |
| Match | Connection between a seeker agent and recruiter agent for a job |
| Message | Individual chat message within a conversation, with sender, content, timestamp |
| Interview | Scheduled interview created during a conversation, with date, time, participants |
| Offer | Job offer created during a conversation, with compensation, terms, status |
| FSM Stage | Current negotiation phase: Introduction, JobDescription, SalaryNegotiation, Interviewing, Offer, Completed |
| Tool Call | AI tool invocation during conversation, with name, arguments, result, status |

## Assumptions

1. The backend API at localhost:8080 is running and accessible during development and testing.
2. PostgreSQL, Chroma, and RabbitMQ are available as backend dependencies.
3. The Vercel AI SDK (`ai` and `@ai-sdk/react`) is already installed in package.json (confirmed: it is).
4. `@tanstack/react-virtual` is already installed (confirmed: it is).
5. The Zustand state management library is the canonical client-side state solution.
6. The `joblinker-auth` cookie with JSON format `{token, userId}` is the canonical cookie format.
7. The frontend dev server runs on localhost:3000 during testing.
8. Browser-based integration testing via Playwright CLI skill is preferred over writing test code files.

## Dependencies

- Vercel AI SDK (`ai`, `@ai-sdk/react`) for streaming chat
- `@tanstack/react-virtual` for virtual scrolling
- Zustand for client-side state management
- Radix UI / shadcn for notification/toast components
- Playwright CLI skill for browser-based integration testing
