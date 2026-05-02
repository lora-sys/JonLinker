# Implementation Plan: Fix System Issues

**Branch**: `013-fix-system-issues` | **Date**: 2026-05-01 | **Spec**: [spec.md](./spec.md)

## Summary

Fix 5 critical system issues discovered during testing:
1. Add backend API Gateway middleware (extract tenant headers)
2. Integrate FSM state machine with message handling
3. Replace mock embeddings with real OpenAI/Chroma embeddings
4. Fix Job form JSON serialization bug
5. Implement bidirectional A2A dialogue loop

## Technical Context

**Language/Version**: Go 1.23, TypeScript (Next.js 16)
**Primary Dependencies**:
- Backend: `github.com/chroma-core/chroma/go/pkg/api` or `chromadb-go-client`, `github.com/sashabaranov/go-openai`
- Frontend: existing Vercel AI SDK, existing GatewayClient

**Storage**:
- PostgreSQL: existing, adds pgvector extension for fallback
- Chroma: Docker container, for vector embeddings
- Redis: existing, for session/rate limiting

**Constraints**:
- Cannot break existing WebSocket, RabbitMQ, AI API integrations
- Must work when Chroma is unavailable (fallback to pgvector)
- Must not hardcode embeddings (use AI_API_KEY or OPENAI_API_KEY env)

## Project Structure

### New Files (Backend)

```text
backend/internal/
├── middleware/
│   └── gateway.go          # NEW: X-User-ID, X-Agent-ID, X-Tenant-ID extraction
├── service/
│   ├── embedding_service.go # NEW: Real embedding generation (OpenAI/Chroma)
│   └── fsm_integration.go   # NEW: FSM state machine integration
```

### Modified Files (Backend)

```text
backend/
├── cmd/server/main.go      # MODIFIED: Register gateway middleware
├── internal/service/
│   ├── message_queue_service.go  # MODIFIED: FSM trigger, bidirectional routing
│   └── agent_memory_service.go   # MODIFIED: Use real embeddings
├── internal/handler/
│   └── job.go               # MODIFIED: Parse structured JSON string
```

### Modified Files (Frontend)

```text
frontend/src/
├── app/jobs/page.tsx        # MODIFIED: JSON.stringify structured field
├── components/job/JobForm.tsx  # MODIFIED: Serialize before submit
```

## Architecture

### 1. Gateway Middleware

```
Request → GatewayMiddleware → Gin Context → Handlers
                ↓
        Extract headers:
        - X-User-ID → ctx.Set("userID", ...)
        - X-Agent-ID → ctx.Set("agentID", ...)
        - X-Tenant-ID → ctx.Set("tenantID", ...)
        - X-Request-ID → ctx.Set("requestID", ...)
```

### 2. FSM Integration

```
handleAgentMessage(msg)
    ↓
Extract intent from msg.Intent
    ↓
Map intent → FSM event (see fsm_integration.go)
    ↓
Call fsm.Handle(event)
    ↓
Persist new state to match.fsm_state
```

### 3. Real Embedding Service

```
generateTextEmbedding(text)
    ↓
Check CHROMA_HOST configured?
    → Yes: Use Chroma client to generate embedding
    → No: Use OpenAI API (OPENAI_API_KEY env)
        → Fallback: pgvector with 768-dim vector
    ↓
Return []float64 embedding (1536 for OpenAI, 768 for local)
```

### 4. Bidirectional A2A Flow

```
Seeker sends message → RabbitMQ → handleAgentMessage
    ↓
Seeker message processed, intent detected
    ↓
Is receiver an Agent?
    → Yes: Create new agent_message for recruiter, route to AI
    → No: Normal human response
    ↓
Recruiter AI generates response → RabbitMQ → handleAgentMessage
    ↓
(loop continues up to max 10 rounds)
```

## FSM Intent → Event Mapping

| Intent | FSM Event | New State |
|--------|-----------|-----------|
| INTRODUCTION | EventStartSearch | Searching |
| JOB_DETAILS | EventMatchFound | Negotiating |
| NEGOTIATE | EventNegotiate | Negotiating |
| SCHEDULE_INTERVIEW | EventScheduleInterview | Interviewing |
| INTERVIEW_COMPLETE | EventInterviewComplete | Negotiating |
| OFFER_CREATED | EventOfferReceived | OfferReceived |
| OFFER_ACCEPTED | EventOfferAccepted | Hired |
| OFFER_DECLINED | EventOfferDeclined | Rejected |

## Embedding Configuration

| Env Variable | Purpose | Required |
|--------------|---------|----------|
| CHROMA_HOST | Chroma server address (e.g., localhost:8000) | No (fallback to OpenAI) |
| OPENAI_API_KEY | OpenAI API for embeddings | No (fallback to mock) |
| AI_API_KEY | Alternative to OPENAI_API_KEY | No |
| EMBEDDING_MODEL | Model to use (text-embedding-3-small default) | No |

**Fallback Chain**:
1. Chroma (if CHROMA_HOST set)
2. OpenAI embeddings (if API key set)
3. Mock embeddings (log warning, don't fail)

## Edge Cases

- **Chroma down**: Log error, use OpenAI embeddings or mock
- **OpenAI API fail**: Return error to user, don't silently use mock
- **FSM invalid transition**: Log warning, keep current state
- **Circular A2A > 10 rounds**: Pause conversation, require human confirmation
- **Duplicate message**: Check message ID for idempotency before processing

## Verification

1. **Gateway middleware**: `curl -H "X-User-ID: test" /api/agents` → logs show "userID=test"
2. **FSM integration**: Send INTRODUCTION intent → check match.fsm_state = "searching"
3. **Real embeddings**: Store "Java developer" → search "Python developer" → should not be top result
4. **Job form**: Submit job with structured → query DB → structured is JSON string
5. **A2A bidirectional**: Create match → send seeker message → verify recruiter auto-responds