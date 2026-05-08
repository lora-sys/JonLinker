# Quickstart: Eino Framework Integration

## Prerequisites

- Go 1.18+
- Existing JobLinker backend running
- RabbitMQ, PostgreSQL configured

## Installation

```bash
cd backend
rtk go get github.com/cloudwego/eino@latest github.com/cloudwego/eino-ext@latest
```

## Quick Test: Basic Eino Agent

### Step 1: Verify Eino ChatModel Wrapper

```go
// In pkg/ai/client.go, verify Eino interface implementation
import "github.com/cloudwego/eino/chatmodel"

var _ chatmodel.ChatModel = (*EinoAIClient)(nil)
```

### Step 2: Run Basic Agent Test

```bash
cd backend
rtk go test ./internal/eino/... -v -run TestBasicAgent
```

### Step 3: Run Full Integration Test

```bash
# Start backend
rtk go run ./cmd/server/main.go

# Send test A2A message
rtk curl -X POST http://localhost:8080/api/messages/test-match-id \
  -H "Content-Type: application/json" \
  -d '{"content_xml": "<message><intent>INTRODUCTION</intent></message>"}'
```

## Expected Behavior

### Phase 1: Basic Agent
- Agent responds to messages without tools
- Memory context loaded correctly
- Prompt template applied

### Phase 2: With Tools
- `query_jobs` returns job summaries
- `search_candidates` performs vector search
- Tool permissions enforced

### Phase 3: With Memory
- Context window maintained across messages
- Old messages summarized when budget exceeded
- Long-term memories recalled via vector search

### Phase 4: With FSM Graph
- State transitions happen via Eino Graph edges
- DeepAgent coordinates Seeker + Recruiter
- Full A2A conversation flow automated

## Common Issues

### Build Errors
```bash
# Missing Eino dependencies
rtk go get github.com/cloudwego/eino@latest

# Interface not implemented
# Verify pkg/ai/client.go implements chatmodel.ChatModel
```

### Runtime Errors
```
# RabbitMQ connection failed
# → Check RABBITMQ_URL environment variable

# AI API failed
# → Check AI_API_KEY and AI_BASE_URL

# Vector store unavailable
# → Falls back to in-memory mock
```

## Verification Commands

```bash
# 1. Build
rtk go build ./cmd/server/...

# 2. Unit tests
rtk go test ./internal/eino/... -v

# 3. Integration test
rtk go test ./internal/service/... -v -run TestEinoIntegration

# 4. Check Eino agent responds
curl -X POST http://localhost:8080/api/messages/{match_id} \
  -H "Authorization: Bearer {token}" \
  -d '{"content_xml": "<message><intent>INTRODUCTION</intent></message>"}'
```

## Next Steps

1. **Phase 1 Complete** → Move to Tool Migration
2. **Tool Migration Complete** → Move to Prompt Migration
3. **All Phases Complete** → Full E2E regression test