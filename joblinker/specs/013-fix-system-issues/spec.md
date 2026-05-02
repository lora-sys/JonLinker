# Feature Specification: Fix System Issues - API Gateway, FSM, Vector Embeddings

**Feature Branch**: `013-fix-system-issues`
**Created**: 2026-05-01
**Status**: Draft
**Input**: System test results showing critical failures in API Gateway backend, FSM integration, vector embeddings, and bidirectional A2A dialogue

## Context

System test revealed 5 critical issues:
1. **API Gateway backend missing**: Frontend has gateway client but backend has no middleware to read X-User-ID/X-Agent-ID/X-Tenant-ID headers
2. **FSM not integrated**: FSM exists in `agent/fsm.go` with full state machine but not triggered by message handling
3. **Vector embeddings are mock**: `generateTextEmbedding()` returns zero vectors, not real embeddings
4. **Job form JSON bug**: structured field sends JSON object but backend expects JSON string
5. **Unidirectional A2A**: Only seeker→recruiter, no bidirectional dialogueloop

## User Scenarios & Testing

### US1 - Backend API Gateway Middleware (P0 - Critical)

As a system, I need all backend requests to go through a gateway middleware that extracts and validates tenant headers.

**Why critical**: Without this, X-User-ID, X-Agent-ID, X-Tenant-ID are not processed, breaking multi-tenancy and audit logging.

**Acceptance Scenarios**:
1. **Given** a request with `X-User-ID: abc123`, **When** it reaches any handler, **Then** the handler can access `ctx.Get("userID")`
2. **Given** a request without headers, **When** it arrives, **Then** middleware sets default values for tracing
3. **Given** a WebSocket upgrade request, **When** it has token query param, **Then** gateway middleware validates and extracts user context

---

### US2 - FSM State Integration (P0 - Critical)

As a system, I need FSM state transitions to trigger automatically based on message intent during A2A dialogueloop.

**Why critical**: Currently FSM exists but is never triggered - negotiation progresses without state tracking.

**Acceptance Scenarios**:
1. **Given** an INTRODUCTION intent is received, **When** handleAgentMessage processes it, **Then** FSM transitions from IDLE to SEARCHING
2. **Given** a NEGOTIATE intent is received, **When** handleAgentMessage processes it, **Then** FSM transitions to NEGOTIATING
3. **Given** an OFFER_CREATED intent is received, **When** handleAgentMessage processes it, **Then** FSM transitions to OFFER_RECEIVED
4. **Given** state transitions, **When** they occur, **Then** match.fsm_state is updated in database

---

### US3 - Real Vector Embeddings (P0 - Critical)

As a system, I need genuine text embeddings using OpenAI API or pgvector for semantic memory recall.

**Why critical**: Zero vectors mean similarity search returns garbage, breaking memory recall feature.

**Acceptance Scenarios**:
1. **Given** agent stores a preference with embedding, **When** it uses AI API (env: OPENAI_API_KEY or AI_API_KEY), **Then** a real 1536-dim embedding is generated
2. **Given** agent recalls preferences, **When** it calls SearchSimilarPreferences, **Then** results are ranked by cosine similarity with real embeddings
3. **Given** Chroma is available via `CHROMA_HOST:CHROMA_PORT`, **When** embedding is stored, **Then** it persists to Chroma collection

---

### US4 - Job Form JSON Serialization (P1)

As a user, I need job creation forms to submit correctly so jobs are stored properly.

**Acceptance Scenarios**:
1. **Given** structured field contains `{salary: {min: 80000, max: 120000}}`, **When** form is submitted, **Then** backend receives JSON string not object
2. **Given** backend receives job with structured as string, **When** it stores to DB, **Then** the JSON is valid and queryable

---

### US5 - Bidirectional A2A Dialogue (P1)

As a system, I need recruiter agent to auto-respond to seeker messages, forming a true dialogueloop.

**Acceptance Scenarios**:
1. **Given** seeker sends a message, **When** recruiter receives it, **Then** recruiter generates and sends a response back
2. **Given** recruiter sends an offer, **When** seeker receives it, **Then** seeker acknowledges and continues negotiation
3. **Given** both agents are active, **When** messages flow, **Then** FSM state tracks the conversation progression

---

## Edge Cases

- **Chroma unavailable**: Fallback to pgvector or in-memory vector store
- **OpenAI API unavailable**: Use mock embeddings with clear logging, don't fail silently
- **FSM invalid transition**: Log warning, keep current state, don't crash
- **Circular messages**: Max 10 exchanges per match before requiring human confirmation

## Requirements

### FR-001: Backend Gateway Middleware
- Create `backend/internal/middleware/gateway.go`
- Extract `X-User-ID`, `X-Agent-ID`, `X-Tenant-ID` from all requests
- Store in Gin context: `ctx.Set("userID", ...)`
- Log all requests with correlation ID

### FR-002: FSM State Machine Integration
- Integrate FSM into MessageQueueService.handleAgentMessage
- Map intents to FSM events: INTRODUCTION→START_SEARCH, NEGOTIATE→EventNegotiate, OFFER_CREATED→EventOfferReceived
- Persist state changes to match.fsm_state column

### FR-003: Real Embedding Generation
- Add `github.com/sashabaranov/go-openai` or use existing AI client for embeddings
- Integrate with Chroma client `github.com/chroma-core/chroma/go/pkg/api`
- Fallback: Use pgvector `pgvector` extension with `vector` column type
- Generate 1536-dim OpenAI embeddings or 768-dim for local models

### FR-004: Job Form JSON Fix
- Frontend: JSON.stringify structured field before sending
- Backend: Parse string to JSON on job create/update

### FR-005: Bidirectional A2A Router
- In handleAgentMessage, after processing sender message, check if receiver is also an agent
- If receiver is agent, route message back through AI for response
- Track conversation round count to prevent infinite loops

## Success Criteria

- **SC-001**: All backend handlers can read `ctx.Get("userID")` from gateway middleware
- **SC-002**: FSM state transitions are logged and persisted when intents are processed
- **SC-003**: Vector embeddings are real (verified by storing text and retrieving similar)
- **SC-004**: Job form submits without JSON parse errors
- **SC-005**: A2A dialogue completes at least 3 exchange rounds without human intervention

## Assumptions

- Chroma runs in Docker: `docker ps | grep chroma` shows container
- OpenAI API key available via OPENAI_API_KEY or AI_API_KEY env
- PostgreSQL with pgvector extension or Chroma for vector storage
- Existing FSM definitions in `agent/fsm.go` are correct

## Dependencies

- API Gateway frontend (specs/011-api-gateway) - already exists, needs backend counterpart
- Vector service (existing in `service/vector_service.go`) - needs real implementation
- Message queue service (existing in `service/message_queue_service.go`) - needs FSM integration