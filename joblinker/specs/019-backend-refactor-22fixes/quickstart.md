# Quickstart: Backend Refactor — Fix 22 Issues

**Date**: 2026-05-08

## Prerequisites

- Go 1.22+
- PostgreSQL 15+ (running, database `joblinker` created)
- Chroma vector database (running on port 8000)
- RabbitMQ (optional — system degrades gracefully without it)
- AI API key set in `AI_API_KEY` env var

## Build & Verify

```bash
cd joblinker/backend
go build ./...
```

## Run Tests

```bash
# Unit tests
go test ./... -count=1 -v

# Integration tests (requires live services)
go test ./... -tags=integration -count=1 -v
```

## Verification Checklist

After all phases are complete, run these checks:

```bash
cd joblinker/backend

# Phase 0: Compilation
go build ./...

# Phase 1: Chroma
grep -n "EnsureCollection" pkg/chroma/client.go && echo "PASS" || echo "FAIL"

# Phase 2: Extraction
grep -c "extracted_from_conversation" internal/service/conversation_service.go && echo "FAIL" || echo "PASS"

# Phase 3: AI integration
grep -c "Seeker proposal in round" internal/service/dual_agent_negotiation_service.go && echo "FAIL" || echo "PASS"

# Phase 4: Hardcoded values
grep -n "100000" internal/service/message_queue_service.go | grep -v "MaxTokens" && echo "FAIL" || echo "PASS"

# Phase 5: AutoMigrate
grep -n "AgentToolCall" cmd/server/main.go && echo "PASS" || echo "FAIL"

# Phase 6: Tests
grep -c "t.Skip" tests/integration/*.go && echo "FAIL" || echo "PASS"

# Phase 7: Dead code
test -f internal/agent/memory.go && echo "FAIL" || echo "PASS"
grep -n "GenerateEmbedding" pkg/ai/client.go && echo "FAIL" || echo "PASS"
grep -n "APIKey length" internal/service/message_queue_service.go && echo "FAIL" || echo "PASS"
```

All checks must print PASS.
