# Implementation Plan: AI UI Components Upgrade

**Branch**: `011-ai-ui-components` | **Date**: 2026-04-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification for frontend AI component optimization using Vercel AI SDK

## Summary

Refactor frontend chat interface to use Vercel AI SDK (`useChat` hook) for streaming conversations, tool call display, and message state management. Integrate with existing WebSocket infrastructure while maintaining backward compatibility.

## Technical Context

**Language/Version**: TypeScript (Next.js 16, React 19)  
**Primary Dependencies**: Vercel AI SDK (`ai` package, `@ai-sdk/react`), existing WebSocket hook (`useChat.ts`)  
**Storage**: In-memory AI state via AI SDK, existing PostgreSQL for message persistence  
**Testing**: Vitest (unit), agent-browser E2E  
**Target Platform**: Web (desktop + mobile)  
**Performance Goals**: Streaming at 30 chars/sec, thread switch under 500ms  
**Constraints**: Must integrate with existing WebSocket backend, not replace it  
**Scale/Scope**: Single-page conversation threads, 1000+ message history per thread

## Constitution Check

| Gate | Status | Notes |
|------|--------|-------|
| Type Safety | ✓ | AI SDK provides typed Message/UIMessage types |
| Test Coverage | ✓ | E2E tests specified in US6 |
| Responsive Design | ✓ | Chat UI must work mobile + desktop |
| Streaming Performance | ✓ | 30 chars/sec target in SC-002 |
| Accessibility | ⚠️ | Need to verify AI SDK streaming is screen-reader compatible |

## Project Structure

### Documentation (this feature)

```text
specs/011-ai-ui-components/
├── plan.md              # This file
├── research.md          # Phase 0: Vercel AI SDK best practices
├── data-model.md        # Phase 1: ChatMessage, ToolCall, Thread entities
├── quickstart.md        # Phase 1: E2E test scenarios
├── contracts/           # Phase 1: API contracts for WebSocket messages
└── tasks.md             # Phase 2: Task breakdown
```

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── components/
│   │   ├── chat/           # NEW: Chat components (ChatWindow, MessageList, StreamingText)
│   │   └── ui/             # Existing UI primitives
│   ├── hooks/
│   │   ├── useChat.ts      # MODIFIED: Integrate AI SDK useChat
│   │   └── useStream.ts    # NEW: Streaming state management
│   ├── lib/
│   │   └── ai/
│   │       └── config.ts  # NEW: AI SDK configuration
│   └── app/
│       └── conversation/
│           └── [matchId]/
│               └── page.tsx  # MODIFIED: Use AI SDK for streaming
```

**Structure Decision**: Add `components/chat/` directory for new AI UI components. Modify existing `useChat.ts` to delegate to AI SDK `useChat` while preserving WebSocket fallback.

## Complexity Tracking

No violations. Feature adds new AI SDK layer without removing existing WebSocket integration.

## Research Required

1. **Vercel AI SDK streaming behavior** - How `useChat` handles concurrent messages and thread isolation
2. **Tool call UI patterns** - Best practices for displaying "calling tool: X" indicators
3. **Message state machine** - How AI SDK handles in-progress/complete/failed states

## Implementation Phases

### Phase 0: Research & Setup
- Research AI SDK `useChat` API, streaming lifecycle, thread management
- Research tool call UI patterns
- Identify all frontend files needing modification

### Phase 1: Design & Contracts
- Define `ChatMessage` entity with AI SDK types
- Define `ToolCall` entity for tool invocation display
- Document WebSocket message contract for streaming
- Write E2E test scenarios for AI conversation flows

### Phase 2: Implementation (via /speckit-tasks)
- Implement `useAIChat` hook wrapping AI SDK
- Create `ChatWindow` component with streaming display
- Create `ToolCallIndicator` component
- Integrate with existing conversation page
- Add E2E tests