# Quickstart: Chroma Integration

## Prerequisites

1. Chroma running: `docker-compose up -d chroma`
2. No external API keys needed (Chroma uses builtin `all-MiniLM-L6-v2`)

## Test Chroma Connection

```bash
curl http://localhost:8000/api/v1/heartbeat
# Should return: {"success": true}
```

## Create Collection

```bash
curl -X POST http://localhost:8000/api/v1/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "test_collection", "get_or_create": true}'
```

## Add Documents (Chroma auto-embeds)

```bash
curl -X POST http://localhost:8000/api/v1/collections/test_collection/add \
  -H "Content-Type: application/json" \
  -d '{
    "ids": ["id1", "id2"],
    "documents": ["Python developer with 5 years experience", "Go developer specializing in distributed systems"],
    "metadatas": [{"skill": "python"}, {"skill": "golang"}]
  }'
```

## Query (Chroma embeds + searches)

```bash
curl -X POST http://localhost:8000/api/v1/collections/test_collection/query \
  -H "Content-Type: application/json" \
  -d '{
    "query_texts": ["backend developer"],
    "n_results": 2
  }'
```

## Go Client Usage

```go
client := chroma.NewClient("localhost", 8000)

// Ensure collection exists
client.GetOrCreateCollection("agent_memories")

// Store
client.Add("agent_memories",
  []string{"mem-123"},
  []string{"User wants remote work"},
  []map[string]interface{}{{"match_id": "match-456"}})

// Search
results, _ := client.Query("agent_memories",
  []string{"remote work preferences"},
  5,
  map[string]interface{}{"match_id": "match-456"})
```
