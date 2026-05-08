# Plan: Chroma Integration + Legacy Fixes

## Overview

Replace all Go-side vector/embedding code with Chroma HTTP client. Chroma 0.6+ uses `all-MiniLM-L6-v2` builtin — no more OpenAI embedding calls or TF-IDF fallback from Go.

## Architecture

```
Go Service                    Chroma Server (localhost:8000)
   |                                    |
   |-- Add(collection, docs) ---------->| Chroma auto-embeds via all-MiniLM-L6-v2
   |-- Query(collection, queryText) --->| Chroma embeds + cosine search
   |-- Delete/CreateCollection -------->|
```

## Collections

| Collection | Purpose | Partition Key |
|------------|---------|---------------|
| `agent_memories` | Conversation memory | matchID in metadata |
| `user_preferences` | Extracted preferences | userID in metadata |
| `job_embeddings` | Job semantic index | jobID in metadata |

## Tech Stack

- **Language**: Go
- **HTTP Client**: Standard library `net/http`
- **Chroma**: Docker `chroma/chroma:0.6.3` on port 8000
- **Env Var**: `CHROMA_HOST` (default: `localhost:8000`)

## Files to CREATE

```
pkg/chroma/
  client.go          # Chroma HTTP client (CreateCollection, GetOrCreateCollection, Add, Query, Delete)
  client_test.go     # Unit tests
```

## Files to MODIFY

```
pkg/ai/client.go                         # DELETE GenerateEmbedding() method
internal/service/vector_service.go       # Replace with Chroma proxy (StoreEmbedding, SearchSimilar)
internal/repository/vector_repository.go  # DELETE entire file
internal/service/agent_memory_service.go # Replace generateTextEmbedding with Chroma Add/Query
internal/service/preference_extraction_service.go  # Implement StorePreference/RecallPreferences
internal/service/context_optimizer.go    # Connect to mqSvc via SetContextOptimizer
cmd/server/main.go                       # Init Chroma client, inject into services, add AgentMemory to AutoMigrate
internal/handler/offer.go                # Fix start_date and salary fallbacks
internal/handler/message.go              # Fix hardcoded timestamp, replace switch/case with AI+RabbitMQ
internal/handler/ws_handler.go           # Fix broadcastToMatch per matchID
frontend/src/lib/vector.ts              # DELETE entire file (if exists)
```

## Files to DELETE

- `pkg/ai/client.go` — `GenerateEmbedding()` method only (keep Chat/ChatWithTools)
- `vector_repository.go` — PostgreSQL vector storage replaced by Chroma
- `agent_memory_service.go` — `generateTextEmbedding()` only
- `frontend/src/lib/vector.ts` — Frontend TF-IDF vector generation

## Implementation Phases

```
Phase 1: P0 Chroma Client + Core Replace
  T101  Create pkg/chroma/client.go
  T102  Replace VectorService with Chroma proxy
  T103  Delete GenerateEmbedding from ai/client.go
  T104  Delete vector_repository.go
  T105  AgentMemory AutoMigrate in main.go

Phase 2: P1 Memory System Wiring
  T201  StoreMemory → Chroma Add
  T202  GetRecentMemories → Chroma Query
  T203  SearchSimilarPreferences → Chroma Query
  T204  StorePreference/RecallPreferences implementation
  T205  ExtractAndStorePreferences calls StorePreference
  T206  Connect contextOptimizer via SetContextOptimizer

Phase 3: P2 WebSocket Flow
  T301  processMessage → AI + RabbitMQ (remove hardcoded switch)
  T302  broadcastToMatch → filter by matchID
  T303  Timestamp → time.Now()

Phase 4: P3 Legacy Fixes
  T401  Offer start_date → dynamic
  T402  Offer salary → from Job.salary_min/max
  T403  Delete AI_API_KEY length log
  T404  Eino dependency check (verify real Eino SDK usage)
```

## Dependencies

- `docker-compose.yml` already has Chroma 0.6.3
- `config.yaml` already has `chroma: {host: localhost, port: 8000}`
- Frontend `lib/chroma.ts` already has complete client

## Verification

```bash
# Build
rtk go build ./...

# Unit tests
rtk go test ./tests/unit/... -count=1

# Integration (requires Chroma running)
docker-compose up -d chroma
CHROMA_HOST=localhost:8000 JWT_SECRET=test rtk go test ./...
```
