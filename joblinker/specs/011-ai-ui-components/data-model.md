# Data Model: AI UI Components

**Feature**: specs/011-ai-ui-components
**Date**: 2026-04-30

## Entities

### ChatMessage

Represents a single message in a conversation thread.

```typescript
interface ChatMessage {
  id: string;                    // Unique message ID (UUID)
  matchId: string;               // Conversation thread identifier
  role: 'user' | 'assistant' | 'system';
  content: string;               // Message text content
  toolInvocations?: ToolCall[];  // Tools called during this message
  attachments?: Attachment[];    // File/image attachments
  createdAt: Date;               // Timestamp
  status?: MessageStatus;        // For assistant: streaming/pending/done/error
}

type MessageStatus = 'streaming' | 'pending' | 'done' | 'error';
```

### ToolCall

Represents a single tool invocation by the AI.

```typescript
interface ToolCall {
  id: string;                    // Unique tool call ID
  toolName: string;              // Tool name (e.g., 'job_query', 'offer_create')
  args: Record<string, unknown>; // Tool input arguments
  result?: unknown;              // Tool output (populated after execution)
  status: 'pending' | 'in_progress' | 'done' | 'error';
  error?: string;               // Error message if failed
}
```

### Attachment

Represents a file or image attachment.

```typescript
interface Attachment {
  id: string;
  type: 'image' | 'pdf' | 'document';
  url: string;                  // URL or base64 data
  name: string;                 // Original filename
  size?: number;                // File size in bytes
}
```

### ConversationThread

Represents an isolated conversation session.

```typescript
interface ConversationThread {
  id: string;                   // Same as matchId
  messages: ChatMessage[];     // Message history
  createdAt: Date;
  updatedAt: Date;
  status: 'active' | 'archived';
}
```

### AIStreamState

Tracks streaming state for in-progress AI responses.

```typescript
interface AIStreamState {
  messageId: string;
  partialContent: string;      // Accumulated streaming text
  toolCalls: ToolCall[];        // In-progress tool calls
  isComplete: boolean;
  error?: string;
}
```

## Relationships

```
ConversationThread (1) ─────< ChatMessage (many)
ChatMessage (1) ─────< ToolCall (many, optional)
ChatMessage (1) ─────< Attachment (many, optional)
```

## Validation Rules

- `ChatMessage.role` must be one of: user, assistant, system
- `ChatMessage.content` max length: 10,000 characters
- `ToolCall.toolName` must match registered tool names
- `Attachment.type` must be one of: image, pdf, document
- `Attachment.size` max: 10MB for images, 25MB for documents

## State Transitions

### Message Status
```
(user sends) → pending → streaming → done
                        ↘ error
(assistant sends) → done
```

### Tool Call Status
```
(called) → pending → in_progress → done
                              ↘ error
```

## Notes

- ChatMessage aligns with Vercel AI SDK's UIMessage type
- Thread isolation enforced by matchId routing in API layer
- Message ordering guaranteed by createdAt timestamp