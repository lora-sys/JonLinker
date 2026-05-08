# Contracts: Eino Framework Integration

## 1. Eino ChatModel Interface (AI Client Wrapping)

### Interface: `chatmodel.ChatModel`

```go
type ChatModel interface {
    Generate(ctx context.Context, messages []*Message) (*Message, error)
    Stream(ctx context.Context, messages []*Message) (*Stream, error)
}
```

### Implementation: `pkg/ai/client.go` → Eino Wrapper

```go
// Wraps existing AI client as Eino ChatModel
type EinoAIClient struct {
    client *ai.Client
}

func (e *EinoAIClient) Generate(ctx context.Context, messages []*chatmodel.Message) (*chatmodel.Message, error) {
    // Convert Eino messages → AI client format
    // Call existing client.Chat()
    // Convert response → Eino Message
}
```

## 2. Eino Tool Interface

### Interface: `tool.BaseTool`

```go
type BaseTool interface {
    Name() string
    Description() string
    Parameters() *schema.Parameters
    Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error)
    Strict() bool
}
```

### Tool Definitions

| Tool Name | Parameters | Return |
|-----------|------------|--------|
| `query_jobs` | `location: string, skills: string[], salary_min: int` | Job list summary |
| `search_candidates` | `skills: string[], location: string, experience_min: int` | Candidate list summary |
| `get_candidate` | `candidate_id: string` | Candidate profile |
| `create_offer` | `match_id: string, salary: int, start_date: string` | Offer confirmation |
| `schedule_interview` | `match_id: string, datetime: string, interview_type: string` | Interview confirmation |

## 3. Eino Memory Interface

### Interface: `memory.Memory`

```go
type Memory interface {
    Save(ctx context.Context, messages ...*Message) error
    Load(ctx context.Context, lastN int) ([]*Message, error)
    Search(ctx context.Context, query string, lastN int) ([]*Message, error)
}
```

### Implementation: `agent_memory.go`

- Short-term: In-memory rolling window
- Long-term: Delegates to VectorStore (Retriever interface)

## 4. Eino Retriever Interface

### Interface: `retriever.Retriever`

```go
type Retriever interface {
    Retrieve(ctx context.Context, query string, options *Options) ([]*Document, error)
}
```

### Implementation: `vector_store.go`

- Wraps existing `vector_repository.go`
- Uses pgvector / Chroma for similarity search

## 5. Eino ChatTemplate Interface

### Interface: `chatprompt.ChatTemplate`

```go
type ChatTemplate interface {
    Format(messages []*Message) (string, error)
    FormatMessages(messages []*Message) ([]*Message, error)
}
```

## 6. Eino Runner Interface

### Interface: `adk/Runner`

```go
type Runner struct {
    agent Agent
    // ...
}

func (r *Runner) Query(ctx context.Context, input string) *Iterator
func (r *Runner) Stream(ctx context.Context, input string) *Stream
```

## 7. Integration Contract: MessageQueueService → Eino

### Before (existing):
```go
func (s *MessageQueueService) handleAgentMessage(msg *AgentMessage) {
    // Manual FSM transitions
    // Manual tool execution
    // Hardcoded prompt strings
}
```

### After (with Eino):
```go
func (s *MessageQueueService) handleAgentMessage(msg *AgentMessage) {
    runner := s.einoRunnerPool.GetRunner(msg.AgentType)
    iter := runner.Query(ctx, msg.XMLContent)

    for {
        event, ok := iter.Next()
        if !ok { break }
        // Process events: tool calls, messages, state changes
    }

    s.einoRunnerPool.PutRunner(runner)
}
```

## 8. External Interface: Existing Systems Unchanged

### WebSocket Handler
- **Unchanged**: `handler/message.go` continues to call `messageQueueService.handleAgentMessage()`
- Eino integration is internal to MessageQueueService

### RabbitMQ
- **Unchanged**: Queues `agent.inbound`, `agent.outbound` remain same
- Message format (XML/Protobuf) unchanged

### Database
- **Unchanged**: Schema same, only agent logic changes