# Data Model: Performance & Stability Optimization

## Entities

### MatchScore (existing, enhanced)

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| seeker_agent_id | UUID | FK to Agent |
| job_id | UUID | FK to Job |
| score | decimal(5,4) | 0.0000 - 1.0000 |
| skills_match | decimal(5,4) | Skills overlap score |
| location_match | decimal(5,4) | Location fit score |
| experience_match | decimal(5,4) | Experience level match |
| created_at | timestamp | Creation time |
| updated_at | timestamp | Last update |

**Relationships**: Belongs to Match, Agent, Job

---

### ConversationState (existing, enhanced)

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| match_id | UUID | FK to Match |
| current_intent | varchar(30) | Current A2A intent state |
| message_count | integer | Total messages exchanged |
| last_message_at | timestamp | Last activity time |
| context_summary | text | AI context for next response |

**Relationships**: Belongs to Match

---

### RateLimitCounter (new)

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | FK to User |
| window_start | timestamp | Sliding window start |
| message_count | integer | Messages in current window |
| created_at | timestamp | Creation time |

**Relationships**: Belongs to User

---

### ErrorLog (new)

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | User context (nullable) |
| match_id | UUID | Match context (nullable) |
| error_type | varchar(50) | Error classification |
| error_message | text | Error details |
| stack_trace | text | Stack trace if applicable |
| correlation_id | varchar(50) | Request tracing ID |
| created_at | timestamp | Error timestamp |

---

## State Transitions

### MatchState FSM
```
pending → active → offered → accepted/rejected
         ↓
    interview_scheduled → completed
```

### Intent Flow (A2A Conversation)
```
INQUIRY → INTRODUCTION → INTEREST → NEGOTIATION → OFFER → ACCEPT → CONFIRM
                  ↓           ↓           ↓
               DECLINE     DECLINE     DECLINE
```

---

## Validation Rules

1. MatchScore.score must be between 0 and 1
2. RateLimitCounter.message_count must not exceed 10 per window
3. ConversationState.current_intent must be valid A2A intent
4. ErrorLog.correlation_id must be unique for tracing
