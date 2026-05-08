# Plan: Eino Agent Routing Layer

## Context

Current JobLinker A2A system has no formal routing layer for agent operations. All agent communication goes through:
- WebSocket handlers directly
- RabbitMQ queues directly
- No agent-level isolation for parallel operations

**Goal**: Add three-layer routing system without modifying any existing business logic.

## Architecture Overview

### Route Hierarchy

```
┌─────────────────────────────────────────────────────────────────┐
│                    PRESENT (PRESERVE UNCHANGED)                  │
├─────────────────────────────────────────────────────────────────┤
│  Frontend → API Gateway → WebSocket/REST handlers               │
│  RabbitMQ (agent.inbound/outbound)                               │
│  PostgreSQL + pgvector / Chroma                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    NEW: ROUTING LAYER (ADDITIVE ONLY)             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Layer 1: Unified Business Router (router.js)                  │
│  ├── /agent/*    → Agent routes                                 │
│  ├── /chat/*     → Chat routes                                  │
│  ├── /tool/*     → MCP tool routes                              │
│  ├── /fsm/*      → FSM event routes                             │
│  ├── /memory/*   → Memory routes                                │
│  └── /vector/*   → Vector search routes                         │
│                                                                  │
│  Layer 2: Agent-Specific Routes                                 │
│  ├── /agent/:agentId/chat      - Agent conversation             │
│  ├── /agent/:agentId/state     - Agent state                   │
│  ├── /agent/:agentId/tool      - Agent tool calls               │
│  ├── /agent/:agentId/memory    - Agent memory                   │
│  └── /agent/:agentId/profile   - Agent config                   │
│                                                                  │
│  Layer 3: Functional Routes                                     │
│  ├── /tool/fetch-job                                           │
│  ├── /tool/fetch-resume                                        │
│  ├── /tool/interview-invite                                     │
│  ├── /tool/generate-offer                                        │
│  ├── /tool/match-vector                                          │
│  ├── /fsm/transition                                            │
│  ├── /fsm/state-sync                                            │
│  ├── /fsm/event                                                  │
│  ├── /memory/save                                                │
│  ├── /memory/retrieve                                            │
│  └── /vector/search                                             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Route Definitions

### Layer 1: Unified Router Entry

File: `backend/internal/router/router.go`

```go
func NewRouter() *Router {
    r := &Router{}
    // Agent routes
    r.AgentRouter = r.NewAgentRouter()
    // Chat routes
    r.ChatRouter = r.NewChatRouter()
    // Tool routes (MCP)
    r.ToolRouter = r.NewToolRouter()
    // FSM routes
    r.FSMRouter = r.NewFSMRouter()
    // Memory routes
    r.MemoryRouter = r.NewMemoryRouter()
    // Vector routes
    r.VectorRouter = r.NewVectorRouter()
    return r
}
```

### Layer 2: Agent-Specific Routes

| Route | Method | Description | Handler |
|-------|--------|-------------|---------|
| `/agent/:agentId/chat` | POST | Agent conversation | `HandleAgentChat` |
| `/agent/:agentId/state` | GET | Get agent state | `HandleAgentState` |
| `/agent/:agentId/tool/:toolName` | POST | Execute tool | `HandleAgentTool` |
| `/agent/:agentId/memory` | GET/POST | Memory operations | `HandleAgentMemory` |
| `/agent/:agentId/profile` | GET | Agent config | `HandleAgentProfile` |

### Layer 3: MCP Tool Routes

| Route | Method | Description | Handler |
|-------|--------|-------------|---------|
| `/tool/fetch-job` | POST | Get job details | `HandleFetchJob` |
| `/tool/fetch-resume` | POST | Get resume | `HandleFetchResume` |
| `/tool/interview-invite` | POST | Schedule interview | `HandleInterviewInvite` |
| `/tool/generate-offer` | POST | Generate offer | `HandleGenerateOffer` |
| `/tool/match-vector` | POST | Vector matching | `HandleMatchVector` |

### Layer 4: FSM Routes

| Route | Method | Description | Handler |
|-------|--------|-------------|---------|
| `/fsm/transition` | POST | State transition | `HandleTransition` |
| `/fsm/state-sync` | POST | Sync state | `HandleStateSync` |
| `/fsm/event` | POST | Publish event | `HandleFSMEvent` |

### Layer 5: Memory/Vector Routes

| Route | Method | Description | Handler |
|-------|--------|-------------|---------|
| `/memory/save` | POST | Save memory | `HandleMemorySave` |
| `/memory/retrieve` | POST | Retrieve memory | `HandleMemoryRetrieve` |
| `/vector/search` | POST | Vector similarity | `HandleVectorSearch` |

## Header Propagation

All routes receive and forward these headers:
- `X-User-ID` → Context user ID
- `X-Agent-ID` → Context agent ID (validated against route param)
- `X-Tenant-ID` → Context tenant ID

### Header Validation Flow

```
Request → Router → Middleware(extract headers) → Handler(validate agentId matches) → Service
```

## Agent Isolation

### How Isolation Works

1. **Route Parameter vs Header Validation**:
   - Route has `:agentId` parameter
   - Header has `X-Agent-ID`
   - Handler validates: `route.agentId == header.X-Agent-ID`

2. **In-Memory Context Per Request**:
   - Each request gets isolated context
   - No shared state between agent requests
   - Go routines handle concurrently safely

3. **Database Queries Scoped by Agent**:
   - All queries filter by agent_id
   - Never cross-agent data leakage

## Files to Create

| File | Purpose |
|------|---------|
| `backend/internal/router/router.go` | Main router entry |
| `backend/internal/router/agent.go` | Agent-specific routes |
| `backend/internal/router/chat.go` | Chat routes |
| `backend/internal/router/tool.go` | MCP tool routes |
| `backend/internal/router/fsm.go` | FSM event routes |
| `backend/internal/router/memory.go` | Memory routes |
| `backend/internal/router/vector.go` | Vector routes |
| `backend/internal/router/middleware.go` | Header extraction, validation |
| `backend/internal/router/context.go` | Request context utilities |

## Files to Modify

| File | Change |
|------|--------|
| `backend/cmd/server/main.go` | Register new router (additive, not replacing) |

## Verification

1. **Isolation Test**:
   - Send request to /agent/agent-1/chat with X-Agent-ID: agent-2
   - Expect: 403 Forbidden (agent ID mismatch)

2. **Header Propagation Test**:
   - Call any route with X-User-ID, X-Agent-ID, X-Tenant-ID
   - Verify headers present in handler context

3. **Regression Test**:
   - Existing WebSocket, RabbitMQ flows work unchanged
   - Existing Protobuf, AI calls work unchanged

## Design Decisions

### Decision 1: Go Standard Library or Gin?

Using **Gin** - aligns with existing backend (`github.com/gin-gonic/gin`)

### Decision 2: Router as Middleware or Standalone?

Router as **standalone wrapper** that calls existing handlers, not replacing them. This ensures:
- Existing handlers unchanged
- Router adds routing layer on top
- Easy to remove if needed

### Decision 3: How to Ensure No Existing Logic Modification?

Router wraps existing handlers:
```go
// New approach
r.POST("/agent/:agentId/chat", func(c *gin.Context) {
    // Add routing layer
    agentID := c.Param("agentId")
    validateHeaderAgent(c, agentID)
    // Call existing handler logic
    existingHandler.HandleChat(c)
})
```

This is **additive** - existing code paths remain untouched.