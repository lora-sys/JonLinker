# API Contracts: AI UI Components

**Feature**: specs/011-ai-ui-components
**Date**: 2026-04-30

## HTTP Streaming Endpoint

### POST /api/chat

Streams AI responses using Vercel AI SDK.

**Request Headers**:
```
Content-Type: application/json
Authorization: Bearer <token>
```

**Request Body**:
```typescript
interface ChatRequest {
  matchId: string;
  messages: UIMessage[];      // Conversation history
  toolCall?: ToolCallConfig;  // Optional tool configuration
}
```

**Response** (text/event-stream):
```
data: {"type":"streaming","content":"Hello"}
data: {"type":"streaming","content":"Hello world"}
data: {"type":"tool_call","toolName":"job_query","args":{...}}
data: {"type":"tool_result","toolName":"job_query","result":{...}}
data: {"type":"done","message":{...}}
```

## WebSocket Events (Existing)

WebSocket continues handling non-AI events.

### Incoming (Server → Client)
```typescript
// Job update notification
{ "type": "job_update", "matchId": string, "job": Job }

// Typing indicator
{ "type": "typing", "matchId": string, "userId": string }

// Presence
{ "type": "presence", "matchId": string, "userId": string, "status": "online|offline" }
```

### Outgoing (Client → Server)
```typescript
// Send message
{ "type": "message", "matchId": string, "content": string }

// Typing start/stop
{ "type": "typing", "matchId": string, "isTyping": boolean }
```

## Tool Call Contract

Tools called by AI are routed through existing internal APIs.

### job_query
```typescript
// Request (AI calls internally via OpenAI function calling)
{
  "name": "job_query",
  "parameters": {
    "query": string,          // Search query
    "location": string,       // Optional location filter
    "jobType": string,        // Optional: full-time|part-time|contract
    "salaryMin": number       // Optional minimum salary
  }
}

// Response
{
  "jobs": [{
    "id": string,
    "title": string,
    "company": string,
    "location": string,
    "salary": { "min": number, "max": number },
    "type": string,
    "skills": string[]
  }],
  "total": number
}
```

### offer_create
```typescript
// Request
{
  "name": "offer_create",
  "parameters": {
    "matchId": string,
    "candidateId": string,
    "jobId": string,
    "salary": number,
    "startDate": string,
    "notes": string
  }
}

// Response
{
  "offerId": string,
  "status": "pending",
  "createdAt": string
}
```

## Error Responses

```typescript
interface APIError {
  code: string;        // e.g., "TOOL_CALL_FAILED"
  message: string;    // User-friendly message
  details?: unknown;  // Additional error context
}

// Example
{
  "code": "RATE_LIMITED",
  "message": "Too many messages. Please wait before sending again.",
  "details": { "retryAfter": 30 }
}
```