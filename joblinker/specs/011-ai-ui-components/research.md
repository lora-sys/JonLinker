# Research: AI UI Components Upgrade

**Feature**: specs/011-ai-ui-components
**Date**: 2026-04-30

## Decision: Vercel AI SDK `useChat` Hook

**What was chosen**: Use `@ai-sdk/react` `useChat` hook for streaming chat interface

**Rationale**:
- Built-in streaming text with character-by-character animation
- Message state management (pending/streaming/done/error)
- Tool call tracking via `toolInvocations` field
- Thread management via `threadId` concept
- Type-safe `UIMessage` and `AIState` types
- Compatible with React Server Components

**Alternatives considered**:
- Custom WebSocket streaming (rejected - would duplicate AI SDK functionality)
- LangChain.js (rejected - overkill for frontend UI state only)
- Raw fetch with streaming (rejected - lacks state management)

---

## Decision: Streaming Architecture

**What was chosen**: AI SDK handles streaming internally, WebSocket carries only non-AI messages

**Rationale**:
- AI SDK `useChat` makes POST to `/api/chat` endpoint
- Backend streams AI responses via AI SDK's `streamText`/`generateText`
- WebSocket continues handling: job updates, presence, typing indicators
- Clear separation: AI SDK = conversation content, WebSocket = real-time events

**Message flow**:
```
User → Frontend useChat → POST /api/chat → Backend AI → Stream response → Frontend useChat → UI update
User → Frontend WebSocket → WebSocket server → Broadcast
```

---

## Decision: Tool Call Display

**What was chosen**: Show tool name + pending state in message bubble, then result inline

**UI Pattern**:
```
[AI] Calling tool: job_query... (spinner)
[AI] Results: Found 5 Python developer jobs in London
```

**Rationale**:
- Transparency: users see exactly what the AI is doing
- Progressive disclosure: tool result becomes part of AI response
- Matches AI SDK `toolInvocations` structure with status tracking

---

## Decision: Thread Management

**What was chosen**: Use `useChat` threadId + matchId mapping

**Rationale**:
- Each conversation (match) has isolated message history
- `matchId` from URL maps to AI SDK thread
- Thread state persisted to localStorage for page navigation
- Page navigation restores correct thread in under 500ms (SC-005)

---

## Key AI SDK Types

```typescript
// From @ai-sdk/react
interface UIMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  toolInvocations?: ToolInvocation[];
  createdAt?: Date;
}

interface ToolInvocation {
  id: string;
  toolName: string;
  args: Record<string, unknown>;
  result?: unknown;
  status: 'pending' | 'done' | 'error';
}

// useChat returns
interface UseChatOptions {
  api?: string;           // POST endpoint, default '/api/chat'
  threadId?: string;      // Thread identifier
  initialMessages?: UIMessage[];
  onError?: (error: Error) => void;
  onFinish?: (message: UIMessage) => void;
}
```

---

## Open Questions Resolved

1. **WebSocket conflict**: No conflict - AI SDK uses HTTP POST/stream, WebSocket handles other events
2. **State sync**: AI SDK owns message list, no sync needed with WebSocket
3. **Authentication**: Pass auth token via headers in `useChat` config
4. **Error recovery**: AI SDK has built-in retry, can customize via `onError` handler