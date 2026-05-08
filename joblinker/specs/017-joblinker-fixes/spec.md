# Spec: JobLinker Backend Restructuring

## Overview

Backend restructuring from mock/stub data layer to production-ready A2A recruitment system. All tools execute real database queries, AI drives all responses, memory persists to database.

## Scope

### What We Are Building
- Real database integration for all tool calls (jobs, candidates, matches)
- Eino-powered AI orchestration (ChatModel, ToolsNode, Graph)
- Persistent memory system with pgvector similarity search
- AI-driven intent classification (no hardcoded switch/case)
- WebSocket with matchId-scoped broadcasting

### What We Are NOT Building
- Frontend changes (handled in separate spec)
- New API endpoints (existing endpoints suffice)
- Migration from existing Go backend to different language

---

## Test Benchmarks (81 tests)

All tests must pass against **real** infrastructure:
- Real database (PostgreSQL with pgvector)
- Real AI API (LongCat API, no mocks/fallbacks)
- Real RabbitMQ (message queue)

### Test Benchmark Hierarchy

```
P0 CORE (blocks TB-9 A2A):
├── TB-6: AI Real Response (6 tests)
│   └── AI generates valid JSON intent from real context
├── TB-7: Tool Calls Real Data (8 tests)
│   └── All tools query real DB, extract AI values
└── TB-9: A2A Complete Dialogue (10 tests)
    └── Full bidirectional recruitment flow

P1 (independent, can run parallel):
├── TB-2: Auth Flow (8 tests)
├── TB-8: FSM State Transitions (10 tests)
├── TB-10: WebSocket Real-Time (8 tests)
└── TB-11: Memory System Persistence (6 tests)

P2 (foundation, run first):
├── TB-1: Infrastructure Connectivity (4 tests)
├── TB-3: Agent CRUD (7 tests)
├── TB-4: Job CRUD (5 tests)
└── TB-5: Match Auto-Matching (5 tests)

P3:
└── TB-12: Security Baseline (8 tests)
```

### P0 Core Test Detail

#### TB-6: AI Real Response
1. Send INTRODUCTION message → receive 201
2. RabbitMQ consumer processes → AI responds
3. `GET /api/messages/:matchId` returns ≥ 2 messages
4. AI response > 50 chars, valid intent (INTEREST/INTRODUCTION/NEGOTIATION)
5. AI response references real job title or seeker skills
6. AI response is valid JSON with `intent` and `message` fields

#### TB-7: Tool Calls Real Data
1. `query_jobs` → returns jobs from DB (not "Sample Job")
2. `get_candidate` → returns real agent with actual skills
3. `search_candidates` → returns matched seekers by vector similarity
4. `schedule_interview` → creates DB record
5. `create_offer` → creates DB record
6. Repeat `query_jobs` → cache hit on second call
7. `create_offer` salary from AI `parameters.data.salary_max` (not 150000)
8. `schedule_interview` datetime from AI response (not "2026-06-01T10:00:00Z")

#### TB-9: A2A Complete Dialogue
1. Full flow: INTRODUCTION → INTEREST → NEGOTIATION → SCHEDULE → OFFER → ACCEPT
2. 5 consecutive AI responses differ in content
3. Intent progresses per FSM rules (no跳跃, no回退)
4. Bidirectional: seeker → recruiter → seeker → ...
5. Rounds ≤ 10 (compression kicks in if exceeded)
6. All messages persist to DB
7. Interview record created on SCHEDULE
8. Offer record created on OFFER
9. Match final status = hired/rejected
10. Concurrent matches don't interfere

---

## Architecture

### Eino Integration

```
┌─────────────────────────────────────────────────────────┐
│  MessageHandler                                         │
│  POST /api/messages/:matchId                           │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  RabbitMQ Queue: agent.messages                         │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  MessageQueueService.processMessage()                    │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Eino Graph (FSM Orchestration)                   │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │  │
│  │  │ ChatModel   │  │ ToolsNode   │  │ Memory   │  │  │
│  │  │ (LongCat)   │  │ (6 tools)   │  │ Context  │  │  │
│  │  └─────────────┘  └─────────────┘  └─────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  broadcastToMatch(matchID) → WebSocket clients          │
└─────────────────────────────────────────────────────────┘
```

### Tool System

| Tool | Repository | Vector Search |
|------|------------|---------------|
| query_jobs | jobRepo.Search() | No |
| get_candidate | agentRepo.GetByID() | No |
| search_candidates | agentRepo.SearchBySkills() | Yes (pgvector) |
| create_offer | offerRepo.Create() | No |
| schedule_interview | interviewRepo.Create() | No |
| record_tool_call | matchRepo.CreateToolCall() | No |

### Memory System

```
StoreMemory(agentID, content)
  → GenerateEmbedding(content) via AI API
  → INSERT INTO agent_memories (embedding, content, agent_id)

GetRecentMemories(agentID, limit)
  → SELECT FROM agent_memories WHERE agent_id = ? ORDER BY created_at DESC LIMIT ?

SearchSimilar(embedding, topK)
  → SELECT * FROM agent_memories ORDER BY embedding <=> ? LIMIT topK  (cosine distance)
```

---

## Constraints

### MOCKS FORBIDDEN
- AI responses MUST come from real API
- Database queries MUST hit real PostgreSQL
- If AI API unavailable → test FAILS (not skip)
- If DB unavailable → test FAILS (not skip)

### EMBEDDINGS
- All embeddings generated via AI API (OpenAI-compatible)
- Zero vectors only as absolute fallback (log WARNING)
- Embedding dimension: 1536 (text-embedding-ada-002 compatible)

### FSM State Machine
```
idle → searching → matched → negotiating → interviewing → offered → accepted/rejected
```

### WebSocket
- matchId-scoped broadcasting (no cross-match leakage)
- Origin whitelist via ALLOWED_ORIGINS env var
- Protobuf support via PROTOBUF_ENABLED=all

---

## Dependencies

- Go 1.21+
- PostgreSQL with pgvector extension
- RabbitMQ
- LongCat AI API (OpenAI-compatible)

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| DATABASE_URL | Yes | PostgreSQL connection string |
| RABBITMQ_URL | Yes | RabbitMQ connection string |
| AI_API_KEY | Yes | LongCat API key |
| AI_BASE_URL | No | Defaults to OpenAI |
| JWT_SECRET | Yes | Must be non-default in production |
| ALLOWED_ORIGINS | No | WebSocket origin whitelist |
| PROTOBUF_ENABLED | No | "none"\|"internal"\|"websocket"\|"all" |
