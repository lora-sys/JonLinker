# Tasks: AI UI Components Upgrade

**Feature**: specs/011-ai-ui-components
**Date**: 2026-05-01
**Total Tasks**: 28

## Phase 1: Setup

- [x] T001 Install and configure Vercel AI SDK dependencies (`ai`, `@ai-sdk/react`) in frontend/package.json
- [x] T002 Create frontend/src/lib/ai/config.ts for AI SDK configuration (model, streaming settings)
- [x] T003 Create frontend/src/types/ai.ts for ChatMessage, ToolCall, ConversationThread, AIStreamState types

## Phase 2: Foundational

- [x] T004 [P] Create frontend/src/components/chat/MessageList.tsx - base message list with role-based styling
- [x] T005 [P] Create frontend/src/components/chat/StreamingText.tsx - character-by-character streaming animation
- [x] T006 [P] Create frontend/src/components/chat/ToolCallIndicator.tsx - tool call pending/result display
- [x] T007 Create frontend/src/hooks/useAIChat.ts - AI SDK useChat hook wrapper with tool call support
- [x] T008 Create frontend/src/hooks/useStream.ts - streaming state management (partialContent, isComplete, error)

## Phase 3: [US1] Chat Interface (P1)

- [x] T009 [US1] Implement ChatWindow component in frontend/src/components/chat/ChatWindow.tsx
- [x] T010 [US1] Add message input with send button (disabled during streaming)
- [x] T011 [US1] Integrate useAIChat hook with ChatWindow for streaming responses
- [x] T012 [US1] Verify streaming text renders at 30+ chars/sec (SC-002)

## Phase 4: [US2] Message State Management (P1)

- [x] T013 [US2] Style user messages (right-aligned, distinct background)
- [x] T014 [US2] Style assistant messages (left-aligned, streaming animation)
- [x] T015 [US2] Style system messages (centered, muted)
- [x] T016 [US2] Handle message status states: streaming/pending/done/error

## Phase 5: [US3] Tool Call Display (P2)

- [x] T017 [US3] Render "calling tool: X" indicator during tool execution
- [x] T018 [US3] Display tool result inline after execution completes
- [x] T019 [US3] Handle tool call error state with retry option
- [x] T020 [US3] Support multiple concurrent tool calls in sequence

## Phase 6: [US4] Conversation Thread Management (P2)

- [x] T021 [US4] Implement thread isolation by matchId (no cross-talk)
- [x] T022 [US4] Persist conversation state to localStorage/API on navigation
- [x] T023 [US4] Restore conversation state on page load
- [x] T024 [US4] Implement progressive message loading (scroll to load older)

## Phase 7: [US5] AI SDK Integration with Backend (P2)

- [x] T025 [US5] Integrate with existing WebSocket hook (useWebSocket.ts) for real-time updates
- [x] T026 [US5] Route AI tool calls through existing internal APIs (job_query, offer_create)
- [x] T027 [US5] Ensure message continuity on WebSocket reconnection

## Phase 8: [US6] E2E AI Testing (P1)

- [x] T028 [US6] Write Playwright E2E tests for streaming chat flow (Scenario 1 in quickstart.md)
- [x] T029 [US6] Write Playwright E2E tests for tool call display verification
- [x] T030 [US6] Write Playwright E2E tests for thread isolation (Scenario 3)
- [x] T031 [US6] Write Playwright E2E tests for conversation persistence (Scenario 4)
- [x] T032 [US6] Verify all existing manual conversation tests pass (backward compatibility)

## Phase 9: Polish

- [x] T033 Mobile responsive layout (375px width minimum)
- [x] T034 Keyboard navigation (Tab through messages)
- [x] T035 Performance optimization: thread switch under 500ms
- [x] T036 Accessibility: screen reader compatibility for streaming content

---

## Dependency Graph

```
T001 → T002 → T003 → T004,T005,T006 → T007,T008 → T009,T010,T011 → T012
                                                              ↓
T013,T014,T015,T016 (US2 parallel with US1 integration) ←←←←←←←
                            ↓
T017,T018,T019,T020 (US3 depends on US1)
                            ↓
T021,T022,T023,T024 (US4 depends on US2)
                            ↓
T025,T026,T027 (US5 depends on US4)
                            ↓
T028,T029,T030,T031,T032 (US6 E2E tests)
                            ↓
T033,T034,T035,T036 (Polish)
```

## Parallel Execution Examples

**Example 1 (Setup → Foundational)**:
```
Phase 1: T001 → T002 → T003
Phase 2: T004, T005, T006 can run in parallel
```

**Example 2 (US1 components)**:
```
T004, T005, T006, T007, T008 → T009,T010,T011 → T012
```

**Example 3 (US1 + US2 parallel)**:
```
Track A: T004,T005,T006 → T007,T008 → T009,T010,T011
Track B: T013,T014,T015,T016 (parallel after foundation)
```

## Implementation Strategy

**MVP Scope (US1 + US2)**: T001-T016
- Core streaming chat with role-based styling
- Basic useAIChat hook with AI SDK

**Increment 2 (US3 + US4)**: T017-T024
- Tool call indicators and results
- Thread isolation and persistence

**Increment 3 (US5)**: T025-T027
- Backend integration with WebSocket and internal APIs

**Increment 4 (US6 + Polish)**: T028-T036
- E2E tests and polish
