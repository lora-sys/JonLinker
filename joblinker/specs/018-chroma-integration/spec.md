# Spec: Chroma Integration — Unified Vector Storage

## Problem

Go backend currently generates embeddings via:
1. OpenAI `/embeddings` API (requires API key, costs money)
2. TF-IDF fallback (poor quality, stored in PostgreSQL)

Both paths are **fully custom** and inconsistent with frontend which already has a Chroma client.

## Solution

Replace all Go-side embedding/vector code with **Chroma HTTP client**. Chroma 0.6+ has builtin `all-MiniLM-L6-v2` embedding function — no API key needed.

## Chroma HTTP API (v1)

Base URL: `http://{CHROMA_HOST}:8000/api/v1`

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/collections` | POST | Create collection |
| `/collections/{name}` | GET | Get or create |
| `/collections/{name}/add` | POST | Add documents (Chroma auto-embeds) |
| `/collections/{name}/query` | POST | Query with text (Chroma embeds + searches) |
| `/collections/{name}/delete` | POST | Delete by ID |

**Key**: `Add` and `Query` accept **plain text**, not pre-computed embeddings. Chroma handles embedding generation internally.

## Collections

```json
{
  "agent_memories": {
    "description": "Conversation memory per match",
    "partition": "metadata.match_id",
    "documents": "memory content"
  },
  "user_preferences": {
    "description": "Extracted user preferences",
    "partition": "metadata.user_id", 
    "documents": "preference_type: value"
  },
  "job_embeddings": {
    "description": "Job semantic index for matching",
    "partition": "metadata.job_id",
    "documents": "job title + description"
  }
}
```

## Data Flow

### Before (current)
```
Go → OpenAI /embeddings → pgvector (custom cosine)
Go → TF-IDF → pgvector (custom cosine)
```

### After (target)
```
Go → Chroma /add (text only, Chroma auto-embeds)
Go → Chroma /query (text only, Chroma auto-embeds + search)
```

## Files Changed

| Action | File |
|--------|------|
| CREATE | `pkg/chroma/client.go` |
| MODIFY | `internal/service/vector_service.go` |
| MODIFY | `pkg/ai/client.go` (delete GenerateEmbedding) |
| DELETE | `internal/repository/vector_repository.go` |
| MODIFY | `internal/service/agent_memory_service.go` |
| MODIFY | `internal/service/preference_extraction_service.go` |
| MODIFY | `internal/service/preference_vector_repo.go` |
| MODIFY | `internal/service/context_optimizer.go` |
| MODIFY | `cmd/server/main.go` |
| MODIFY | `internal/handler/offer.go` |
| MODIFY | `internal/handler/message.go` |
| MODIFY | `internal/handler/ws_handler.go` |
| DELETE | `frontend/src/lib/vector.ts` |

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `CHROMA_HOST` | `localhost:8000` | Chroma server address |
