# Data Model: Eino Framework Integration

## New Entities

### 1. EinoAgent (interface reference in code)

```go
type EinoAgent struct {
    ID       uuid.UUID
    Type     AgentType  // "seeker" or "recruiter"
    Model    ChatModel  // Eino ChatModel interface
    Tools    []tool.BaseTool
    Prompt   *ChatTemplate
    Memory   Memory
    Runner   *Runner
}
```

### 2. Tool Registration (Eino Tool interface)

```go
// Eino Tool interface implementation
type JobTool struct {
    name        string        // "query_jobs"
    description string
    params      schema.Parameters
    handler     func(ctx context.Context, args map[string]interface{}) (string, error)
}

// Tool definitions
- query_jobs: Search job listings
- search_candidates: Vector similarity search for candidates
- get_candidate: Retrieve candidate profile
- create_offer: Generate compensation offer
- schedule_interview: Schedule interview appointment
```

### 3. FSM State Graph (Eino Composition)

```go
// Eino Graph replacing switch-based FSM
type FSMGraph struct {
    graph *compose.Graph
    nodes map[agent.State]compose.Node
    edges map[agent.State]map[agent.Event]agent.State
}

// State nodes
- idle: Initial state
- searching: Job/candidate search in progress
- matched: Match found
- negotiating: Salary/terms negotiation
- interviewing: Interview scheduled/completed
- offer_received: Offer generated
- hired: Offer accepted
- rejected: Match declined
```

### 4. ChatTemplate (Eino Prompt Management)

```go
// Eino ChatTemplate structure
type ChatTemplate struct {
    SystemPrompt string
    UserPrompt   string
    Examples     []Example
}

// Template scenarios
- greeting: Initial introduction
- negotiation: Salary/terms discussion
- interview: Interview scheduling/feedback
- offer: Offer generation/acceptance
- decline: Polite rejection
```

### 5. Agent Memory (Eino Memory interface)

```go
// Short-term memory (Eino built-in)
type AgentMemory struct {
    WindowSize   int           // Max messages in rolling window
    SummaryThreshold int       // When to trigger summarization
}

// Long-term memory (via Retriever)
type VectorMemory struct {
    CollectionID string
    Retriever   Retriever
}
```

## State Transitions (Eino Graph Edges)

```
idle + START_SEARCH → searching
searching + MATCH_FOUND → matched
matched + INTEREST_EXPRESSED → negotiating
negotiating + SCHEDULE_INTERVIEW → interviewing
interviewing + INTERVIEW_COMPLETE → offer_received
offer_received + OFFER_ACCEPTED → hired
offer_received + OFFER_DECLINED → rejected
```

## Tool Permissions by Agent Type

| Tool | Seeker Agent | Recruiter Agent |
|------|-------------|-----------------|
| query_jobs | ✓ Enabled | ✗ Disabled |
| search_candidates | ✓ Enabled | ✓ Enabled |
| get_candidate | ✓ Enabled | ✓ Enabled |
| create_offer | ✗ Disabled | ✓ Enabled |
| schedule_interview | ✗ Disabled | ✓ Enabled |

## Integration Points

### Existing → Eino Interface Wrapping

| Existing Component | Eino Interface | File |
|--------------------|----------------|------|
| `pkg/ai/client.go` | `chatmodel.ChatModel` | `eino/chatmodel/openai.go` |
| `vector_repository.go` | `retriever.Retriever` | `eino/memory/vector_store.go` |
| `agent/memory.go` | `memory.Memory` | `eino/memory/agent_memory.go` |
| `prompts/*.go` | `chatprompt.ChatTemplate` | `eino/prompt/templates/*.go` |

### Message Flow with Eino

```
1. RabbitMQ message received
2. MessageQueueService.handleAgentMessage()
3. Create Eino Runner with Seeker/Recruiter agent
4. runner.Query(ctx, xmlMessage)
5. Eino Agent processes:
   a. Load memory context
   b. Select appropriate prompt template
   c. Call ChatModel (AI)
   d. If tool call → execute via Eino Tool
   e. Generate response
6. Store result in memory
7. Return XML response
8. Publish to outbound queue / Send via WebSocket
```