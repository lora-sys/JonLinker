# Tasks: Chroma Integration + Legacy Fixes

## Format: `[ID] [Phase] Description`

---

## Phase 1: P0 — Chroma Client + Core Replace

### T101 [P1] [Phase1] Create pkg/chroma/client.go

**Files**:
- `pkg/chroma/client.go` — Chroma HTTP client

**API**:
```go
type Client struct { host string; port int }
func NewClient(host string, port int) *Client
func (c *Client) CreateCollection(name string) error
func (c *Client) GetOrCreateCollection(name string) error
func (c *Client) Add(collection string, ids, documents []string, metadatas []map[string]interface{}) error
func (c *Client) Query(collection string, queryTexts []string, nResults int, where map[string]interface{}) ([]QueryResult, error)
func (c *Client) Delete(collection string, ids []string) error
```

**Tests**: `pkg/chroma/client_test.go`

---

### T102 [P1] [Phase1] Replace VectorService with Chroma proxy

**Files**:
- `internal/service/vector_service.go` — Replace with Chroma

**Changes**:
```go
func (s *VectorService) StoreEmbedding(ctx context.Context, collectionID string, text string, metadata map[string]interface{}) (string, error) {
    id := uuid.New().String()
    err := s.chromaClient.Add(collectionID, []string{id}, []string{text}, []map[string]interface{}{metadata})
    return id, err
}

func (s *VectorService) SearchSimilar(ctx context.Context, collectionID string, query string, limit int) ([]string, []float64, error) {
    results, err := s.chromaClient.Query(collectionID, []string{query}, limit, nil)
    // extract ids and distances
}
```

**DELETE**: `tokenize()`, `computeTF()`, TF-IDF fallback, `GenerateEmbedding()`

---

### T103 [P1] [Phase1] Delete GenerateEmbedding from ai/client.go

**Files**:
- `pkg/ai/client.go`

**Changes**: Remove `GenerateEmbedding()` method. Keep `Chat()` and `ChatWithTools()`.

---

### T104 [P1] [Phase1] Delete vector_repository.go

**Files**:
- `internal/repository/vector_repository.go` — DELETE

PostgreSQL vector storage replaced by Chroma.

---

### T105 [P1] [Phase1] AgentMemory AutoMigrate in main.go

**Files**:
- `cmd/server/main.go`

**Changes**: Add `&model.AgentMemory{}` to AutoMigrate list.

---

## Phase 2: P1 — Memory System Wiring

### T201 [P2] [Phase2] StoreMemory → Chroma Add

**Files**:
- `internal/service/agent_memory_service.go`

**Changes**:
```go
// REPLACE:
_ = memory // Currently logged, would persist in production

// WITH:
s.chromaClient.Add("agent_memories", []string{memory.ID}, []string{content}, []map[string]interface{}{metadata})
```

---

### T202 [P2] [Phase2] GetRecentMemories → Chroma Query

**Files**:
- `internal/service/agent_memory_service.go`

**Changes**:
```go
// REPLACE:
return []*model.AgentMemory{}, nil

// WITH:
results, err := s.chromaClient.Query("agent_memories", []string{query}, limit, map[string]interface{}{"match_id": matchID.String()})
// reconstruct AgentMemory from results
```

---

### T203 [P2] [Phase2] SearchSimilarPreferences → Chroma Query

**Files**:
- `internal/service/preference_vector_repo.go`

**Changes**:
```go
// REPLACE:
r.db.Where("user_id = ? AND preference_type = ?", userID, prefType).Limit(limit).Find(&prefs)

// WITH:
results, _ := s.chromaClient.Query("user_preferences", []string{queryText}, limit, map[string]interface{}{"user_id": userID})
```

---

### T204 [P2] [Phase2] Implement StorePreference and RecallPreferences

**Files**:
- `internal/service/preference_extraction_service.go`

**Changes**:
```go
func (s *PreferenceExtractionService) StorePreference(ctx context.Context, userID uuid.UUID, prefType string, value string, metadata map[string]interface{}) error {
    id := uuid.New().String()
    doc := fmt.Sprintf("%s: %s", prefType, value)
    meta := map[string]interface{}{"user_id": userID.String(), "preference_type": prefType}
    for k, v := range metadata { meta[k] = v }
    return s.chromaClient.Add("user_preferences", []string{id}, []string{doc}, []map[string]interface{}{meta})
}

func (s *PreferenceExtractionService) RecallPreferences(ctx context.Context, userID uuid.UUID, queryText string, limit int) ([]ExtractedPreference, error) {
    results, err := s.chromaClient.Query("user_preferences", []string{queryText}, limit, map[string]interface{}{"user_id": userID.String()})
    // reconstruct ExtractedPreference from results
}
```

---

### T205 [P2] [Phase2] ExtractAndStorePreferences calls StorePreference

**Files**:
- `internal/service/preference_extraction_service.go`

**Changes**:
```go
// REPLACE:
preferences := s.extractPreferencesFromMessages(messages)
log.Printf("Extracted %d preferences...", len(preferences))
return nil

// WITH:
preferences := s.extractPreferencesFromMessages(messages)
for _, pref := range preferences {
    s.StorePreference(ctx, userID, pref.Type, pref.Value, pref.Metadata)
}
return nil
```

---

### T206 [P2] [Phase2] Connect contextOptimizer via SetContextOptimizer

**Files**:
- `cmd/server/main.go`

**Changes**:
```go
// REPLACE: mqSvc.SetContextOptimizer(...) never called
// WITH:
mqSvc.SetContextOptimizer(contextOptimizer)
```

---

## Phase 3: P2 — WebSocket Flow

### T301 [P3] [Phase3] processMessage → AI + RabbitMQ (remove hardcoded switch)

**Files**:
- `internal/handler/message.go`

**Changes**:
```go
// REPLACE: hardcoded switch/case returning IntentInterest etc.
// WITH: call mqSvc.PublishMessage() → RabbitMQ → AI agent → response
```

---

### T302 [P3] [Phase3] broadcastToMatch → filter by matchID

**Files**:
- `internal/handler/ws_handler.go`

**Changes**:
```go
// REPLACE: clients map[string]*websocket.Conn (keyed by userID)
// WITH: clients map[string]map[string]*websocket.Conn (keyed by matchID → userID)
```

---

### T303 [P3] [Phase3] Timestamp → time.Now()

**Files**:
- `internal/handler/message.go`

**Changes**:
```go
// REPLACE:
Timestamp: "2026-04-23T00:00:00Z",

// WITH:
Timestamp: time.Now().Format(time.RFC3339),
```

---

## Phase 4: P3 — Legacy Fixes

### T401 [P4] [Phase4] Offer start_date → dynamic

**Files**:
- `internal/handler/offer.go`

**Changes**:
```go
// REPLACE:
if startDate == "" {
    startDate = "2026-07-01"
}

// WITH:
if startDate == "" {
    startDate = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
}
```

---

### T402 [P4] [Phase4] Offer salary → from Job.salary_min/max

**Files**:
- `internal/handler/offer.go`

**Changes**:
```go
// REPLACE:
if salary == 0 {
    salary = 120000
}

// WITH:
if salary == 0 {
    // Get job and calculate midpoint of salary range
    job, _ := s.jobRepo.GetByID(jobID)
    if job != nil && job.SalaryMax > 0 {
        salary = (job.SalaryMin + job.SalaryMax) / 2
    }
}
```

---

### T403 [P4] [Phase4] Delete AI_API_KEY length log

**Files**:
- `pkg/ai/client.go`

**Changes**:
```go
// DELETE:
log.Printf("AI_API_KEY length: %d", len(os.Getenv("AI_API_KEY")))
```

---

### T404 [P4] [Phase4] Eino dependency check

**Files**:
- `go.mod`

**Changes**: Verify `github.com/cloudwego/eino` is in go.mod. If `internal/eino/` is a custom wrapper (not real Eino SDK), document the gap.

---

## Verification Commands

```bash
# Build check
cd backend && rtk go build ./... && rtk go vet ./...

# Unit tests
cd backend && rtk go test ./tests/unit/... -count=1

# Integration (needs Chroma)
docker-compose up -d chroma
CHROMA_HOST=localhost:8000 JWT_SECRET=test rtk go test ./tests/integration/... -count=1
```
