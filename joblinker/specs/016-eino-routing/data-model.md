# Data Model: Eino Agent Routing Layer

## Route Context

### RouteContext (in-memory per request)

```go
type RouteContext struct {
    UserID    string
    AgentID   string
    TenantID  string
    RequestID string
    Timestamp time.Time
}
```

### Header Mapping

| Header | Context Field | Required |
|--------|-------------|---------|
| X-User-ID | UserID | Yes |
| X-Agent-ID | AgentID | Yes |
| X-Tenant-ID | TenantID | Yes |
| X-Request-ID | RequestID | Auto-generated if missing |

## Route Definitions

### Agent Routes

| Route | Method | Handler | Service |
|-------|--------|---------|---------|
| `/agent/:agentId/chat` | POST | HandleAgentChat | messageQueueService |
| `/agent/:agentId/state` | GET | HandleAgentState | agentService |
| `/agent/:agentId/tool/:toolName` | POST | HandleAgentTool | toolExecutor |
| `/agent/:agentId/memory` | GET/POST | HandleAgentMemory | memoryService |
| `/agent/:agentId/profile` | GET | HandleAgentProfile | agentService |

### Tool Routes

| Route | Method | Handler | Tool Name |
|-------|--------|---------|-----------|
| `/tool/fetch-job` | POST | HandleFetchJob | query_jobs |
| `/tool/fetch-resume` | POST | HandleFetchResume | get_candidate |
| `/tool/interview-invite` | POST | HandleInterviewInvite | schedule_interview |
| `/tool/generate-offer` | POST | HandleGenerateOffer | create_offer |
| `/tool/match-vector` | POST | HandleMatchVector | search_candidates |

### FSM Routes

| Route | Method | Handler | Event |
|-------|--------|---------|-------|
| `/fsm/transition` | POST | HandleTransition | FSM event |
| `/fsm/state-sync` | POST | HandleStateSync | State sync |
| `/fsm/event` | POST | HandleFSMEvent | Raw event |

### Memory/Vector Routes

| Route | Method | Handler | Operation |
|-------|--------|---------|-----------|
| `/memory/save` | POST | HandleMemorySave | Store memory entry |
| `/memory/retrieve` | POST | HandleMemoryRetrieve | Get recent memories |
| `/vector/search` | POST | HandleVectorSearch | Similarity search |

## Validation Rules

### Agent ID Validation

```
Route param agentId MUST match header X-Agent-ID
Mismatch → 403 Forbidden
```

### Tenant Isolation

```
All database queries MUST filter by tenant_id
Cross-tenant access → 403 Forbidden
```

## State Transitions (FSM)

```
idle → searching (START_SEARCH)
searching → matched (MATCH_FOUND)
matched → negotiating (INTEREST_EXPRESSED)
negotiating → interviewing (SCHEDULE_INTERVIEW)
interviewing → offer_received (INTERVIEW_COMPLETE)
offer_received → hired (OFFER_ACCEPTED)
offer_received → rejected (OFFER_DECLINED)
```