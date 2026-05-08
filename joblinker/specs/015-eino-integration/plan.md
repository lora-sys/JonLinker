# Plan: Eino Framework Integration for JobLinker A2A Recruitment System

## Context

Current JobLinker has a custom-built AI orchestration layer with:
- Hand-written FSM state machine (`agent/fsm.go`) - switch-based state transitions
- Scattered prompt strings in `service/prompts/`
- Manual tool registration in `function_definitions.go`
- No built-in context management (token explosion risk)
- Single-threaded message processing

**Goal**: Integrate [Eino framework](https://github.com/cloudwego/eino) (by ByteDance) to replace the 手工编排层 while preserving WebSocket, RabbitMQ, API Gateway, Protobuf communication layers.

## Technical Context

### Existing Architecture
- **Frontend**: Next.js → API Gateway → Backend
- **Messaging**: RabbitMQ (agent.inbound/outbound) + WebSocket
- **Storage**: PostgreSQL + pgvector / Chroma
- **AI**: OpenAI-compatible API via `pkg/ai/client.go`
- **Agent**: Custom FSM, tool executor, prompt service

### Eino Framework (from cloudwego/eino README)
```
Eino['aino] - Golang LLM application framework

Key Components:
├── ADK (Agent Development Kit)
│   ├── ChatModelAgent     - Simple ReAct agent with tools
│   ├── DeepAgent          - Multi-agent coordinator
│   └── Runner             - Execution runtime with interrupt/resume
├── Composition
│   ├── Graph              - DAG orchestration with streaming
│   └── GraphTool          - Expose graphs as tools
├── Component Ecosystem
│   ├── ChatModel          - OpenAI, Claude, Gemini, Ollama
│   ├── Tool               - Function calling
│   ├── Retriever          - Vector search
│   └── Embedding          - Text embedding
```

## Architecture

### Integration Overview
```
┌─────────────────────────────────────────────────────────────────┐
│                    PRESERVE UNCHANGED                           │
├─────────────────────────────────────────────────────────────────┤
│  Frontend → API Gateway → WebSocket/REST handlers              │
│  RabbitMQ (agent.inbound/outbound)                               │
│  PostgreSQL + pgvector / Chroma                                 │
│  Protobuf content negotiation                                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    REPLACE WITH EINO                            │
├─────────────────────────────────────────────────────────────────┤
│  MessageQueueService.handleAgentMessage()                        │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Eino ADK: ChatModelAgent (Seeker/Recruiter)              │   │
│  │   ├── Model: AI client (Eino ChatModel interface)         │   │
│  │   ├── Tools: query_jobs, search_candidates, etc.        │   │
│  │   ├── Prompt: Eino ChatTemplate                         │   │
│  │   └── Memory: Eino Memory (short-term + long-term)       │   │
│  └─────────────────────────────────────────────────────────┘   │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Eino Composition: Graph for FSM Workflow                │   │
│  │   idle → searching → matched → negotiating → ...         │   │
│  │   (replaces manual switch-based FSM)                    │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Migration Strategy

### Phase 1: Eino Infrastructure
1. Add Eino dependencies (`go get github.com/cloudwego/eino`)
2. Create `backend/internal/eino/` directory structure
3. Wrap existing AI client as Eino ChatModel component
4. Create basic ChatModelAgent for seeker (no tools)
5. Verify agent responds to messages

### Phase 2: Tool Migration
1. Migrate 5 tools to Eino Tool interface:
   - `query_jobs` - Search job listings
   - `search_candidates` - Vector similarity search
   - `get_candidate` - Get candidate profile
   - `create_offer` - Generate offer letter
   - `schedule_interview` - Schedule interview
2. Register tools with Eino agent
3. Add tool permission checking

### Phase 3: Prompt Migration
1. Migrate `SeekerPrompts()` / `RecruiterPrompts()` to Eino ChatTemplate
2. Update agent to use Eino prompt loader

### Phase 4: FSM → Eino Graph + DeepAgent
1. Convert FSM to Eino Composition Graph
2. Implement DeepAgent for multi-agent coordination
3. Add Eino Memory for context management

## Key Design Decisions

1. **Keep Communication Layers Unchanged**: WebSocket, RabbitMQ, API Gateway stay the same
2. **Eino Agent = Existing Agent Types**: ChatModelAgent(seeker) = SeekerAgent
3. **FSM Becomes Eino Graph**: `graph.AddEdge(state, event, nextState)` instead of switch
4. **Memory Integration**: Short-term via Eino Memory, long-term via Eino Retriever → pgvector

## Files to Create

| File | Purpose |
|------|---------|
| `backend/internal/eino/agent/seeker_agent.go` | Eino ChatModelAgent for job seeker |
| `backend/internal/eino/agent/recruiter_agent.go` | Eino ChatModelAgent for recruiter |
| `backend/internal/eino/agent/deep_recruiter.go` | Eino DeepAgent for coordination |
| `backend/internal/eino/workflow/fsm_graph.go` | Eino Graph replacing FSM |
| `backend/internal/eino/prompt/loader.go` | Eino ChatTemplate loader |
| `backend/internal/eino/prompt/templates/seeker.go` | Seeker prompt templates |
| `backend/internal/eino/prompt/templates/recruiter.go` | Recruiter prompt templates |
| `backend/internal/eino/memory/agent_memory.go` | Eino Memory interface |
| `backend/internal/eino/memory/vector_store.go` | Eino Retriever for pgvector |
| `backend/internal/eino/tools/job_tools.go` | Job operation tools |
| `backend/internal/eino/tools/candidate_tools.go` | Candidate operation tools |
| `backend/internal/eino/tools/offer_tools.go` | Offer generation tools |
| `backend/internal/eino/tools/interview_tools.go` | Interview scheduling tools |
| `backend/internal/eino/runner/agent_runner.go` | Eino Runner wrapper |

## Files to Modify

| File | Change |
|------|--------|
| `backend/internal/service/message_queue_service.go` | Delegate to Eino Runner |
| `backend/internal/service/fsm_integration.go` | Use Eino Graph edges |
| `backend/internal/agent/fsm.go` | Keep enums, remove transition logic |
| `backend/internal/agent/function_definitions.go` | Register via Eino Tool |
| `backend/internal/agent/tool_executor.go` | Delegate to Eino Tools |
| `backend/internal/service/prompts/*.go` | Migrate to Eino ChatTemplate |
| `backend/pkg/ai/client.go` | Implement Eino ChatModel interface |
| `backend/cmd/server/main.go` | Initialize Eino agents |

## Verification

1. `go build ./cmd/server/...`
2. `go test ./internal/eino/... -v`
3. E2E: Send A2A message through RabbitMQ, verify Eino agent processes it
4. Regression: Existing WebSocket, Protobuf, RabbitMQ flows unchanged

## References

- Eino GitHub: https://github.com/cloudwego/eino
- Eino Docs: https://www.cloudwego.io/docs/eino/
- Eino Examples: https://github.com/cloudwego/eino-examples