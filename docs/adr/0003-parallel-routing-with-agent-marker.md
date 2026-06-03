# ADR 0003: Parallel 3-Way Routing with Agent State Marker

## Status

Accepted

## Context

Phase 1-2 routing (`internal/agent/router.go`) was a simple 2-way fork between Search Agent and Resume Agent, based on profile completeness + keyword detection. Phase 3 introduces a Recruiter Agent that simulates interview conversations.

The three agents have different characteristics:

| Agent | Pattern | Trigger | Memory |
|-------|---------|---------|--------|
| Resume | ChatModel | no profile / resume keywords | own endpoint, no MemoryStore |
| Search | ReAct | default | `sessionID` key in MemoryStore |
| Recruiter | ReAct | has application + interview keywords | needs isolated memory |

Key constraint: user must be able to switch between Search and Recruiter mid-session (e.g. search → apply → interview → search more). This rules out replacement mode.

## Decision

Adopt **parallel routing with a `last_agent` state marker**:

1. **CheckpointStore** stores `sessionID+":last_agent"` — written by each agent after processing, read by router on the next request
2. **MemoryStore** uses per-agent namespaces: `sessionID` for Search, `sessionID+":recruiter"` for Recruiter
3. **Router** evaluates on every request:
   - Keyword match → route to the matching agent
   - Ambiguous/no keyword → fall back to `last_agent`
4. **Default**: new sessions have `last_agent = "search"` (backward compatible)

### Routing Flow

```
hasApplication := checkpoint.HasPrefix(sessionID+":application_")
keyword := detectAgentKeyword(msg)

switch {
case resume-related keywords / no profile:
    return AgentResume
case keyword == "search" || (!hasApplication && lastAgent != "recruiter"):
    return AgentSearch
case keyword == "recruiter" || lastAgent == "recruiter":
    return AgentRecruiter
default:
    return lastAgent
}
```

### Agent Marker Lifecycle

- Written by each agent's first tool call or text output
- Read by router at the start of each request
- Cleaned up by session TTL (30-minute goroutine in main.go)

## Consequences

### Positive

- Zero frontend changes — no new endpoints, no new UI components
- Backward compatible — existing sessions (no `last_agent` key) default to Search
- User stays in one chat thread; no mode switching UI needed
- Agent-specific memory isolation prevents cross-agent context pollution

### Negative

- CheckpointStore now carries routing state, not just profile/application data (impurity)
- Router complexity increases — now evaluates agent availability, keywords, and last-agent state
- `last_agent` must be kept in sync with session TTL cleanup
- Debugging routing issues requires tracing checkpoint reads at each request

## Alternatives Considered

### 1. Separate Endpoint + New UI (`/api/chat/recruiter`)
- **Pros**: clean separation, no routing complexity
- **Cons**: new frontend component, new route, user must switch contexts explicitly
- **Rejected**: violates "no new UI" constraint, creates disjointed UX

### 2. Replacement Mode
- **Pros**: simplest router, single memory namespace
- **Cons**: user cannot return to search after interview — violates parallel mode requirement
- **Rejected**: does not meet requirements

### 3. Stateless Keyword-Only Routing
- **Pros**: no marker to persist/clean up
- **Cons**: ambiguous utterances ("继续", "好的") cannot be routed without context
- **Rejected**: poor UX — user would need to repeat keywords on every message

### 4. Confidence-Scoring Router
- **Pros**: extensible, ML-friendly
- **Cons**: over-engineered for 3 agents, introduces scoring threshold tuning
- **Deferred**: revisit when Phase 4 adds a 4th agent
