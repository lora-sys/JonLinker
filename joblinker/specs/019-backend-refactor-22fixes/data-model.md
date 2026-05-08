# Data Model: Backend Refactor — Fix 22 Issues

**Date**: 2026-05-08
**Plan**: [plan.md](plan.md)

## Entities

### ChromaCollection (in-memory, not persisted)

Represents a cached mapping from a collection name to its Chroma UUID.

| Field | Type | Description |
|-------|------|-------------|
| name | string | Human-readable collection name (e.g., "agent_memories") |
| uuid | string | Chroma-assigned collection UUID |

**Lifecycle**: Created on first `EnsureCollection` call, cached for the process lifetime.

### ExtractedPreference (existing, modified)

Represents a preference extracted from conversation text.

| Field | Type | Description |
|-------|------|-------------|
| Type | string | One of: salary_expectation, preferred_location, job_type, skills |
| Value | string | Extracted value (real data, never a placeholder) |
| Confidence | float64 | Extraction confidence (0.0-1.0) |

**Validation rules**:
- Value MUST NOT equal "extracted_from_conversation"
- Value MUST NOT be empty
- Confidence MUST be between 0.0 and 1.0

### ConfirmationRequest (existing model, added to AutoMigrate)

Represents a pending human confirmation for an AI-negotiated agreement.

| Field | Type | Description |
|-------|------|-------------|
| ID | uuid | Primary key |
| MatchID | uuid | Associated match |
| Terms | string | Negotiated terms summary |
| Status | string | pending, confirmed, rejected |
| CreatedAt | timestamp | Creation time |

### AgentToolCall (existing model, added to AutoMigrate)

Records tool invocations made by AI agents.

| Field | Type | Description |
|-------|------|-------------|
| ID | uuid | Primary key |
| AgentID | uuid | Agent that made the call |
| ToolName | string | Tool that was called |
| Input | json | Tool input parameters |
| Output | json | Tool output result |
| CreatedAt | timestamp | Creation time |

## State Transitions

### ConfirmationRequest

```
pending → confirmed (user accepts)
pending → rejected (user declines)
```

No other transitions are valid. A confirmed/rejected request cannot be changed.
