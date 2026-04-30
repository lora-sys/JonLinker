# Data Model: Frontend UI Completion

**Feature**: specs/007-name-ui-completion
**Date**: 2026-04-28

## Entities

### Job (existing, extended)

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| id | UUID | auto-generated | Primary key |
| agent_id | UUID | required, must exist | Recruiter agent reference |
| title | string | required, max 200 chars | Job title |
| description | string | required, max 5000 chars | Full job description |
| location | string | optional, max 200 chars | City/remote/etc |
| type | enum | required | full-time, part-time, contract |
| salary_range | string | optional | Display string (e.g. "100k-150k") |
| structured | JSON | required | { requirements[], nice_to_have[], location, work_type, salary_range: { min, max } } |
| vector_id | UUID | auto-generated | Chroma embedding reference |
| status | enum | default: active | active, inactive |
| created_at | timestamp | auto | |
| updated_at | timestamp | auto | |

**Relationships**: belongs to Agent (recruiter), has many Matches

---

### Interview (existing)

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| id | UUID | auto-generated | Primary key |
| match_id | UUID | required, must exist | Match reference |
| scheduled_at | timestamp | required | Interview datetime |
| format | enum | required | video, onsite, phone |
| location | string | optional | URL for video, address for onsite |
| status | enum | default: pending | pending, confirmed, cancelled |
| reminder_sent | boolean | default: false | |
| created_at | timestamp | auto | |
| updated_at | timestamp | auto | |

**Relationships**: belongs to Match

---

### Offer (existing)

| Field | Type | Validation | Notes |
|-------|------|------------|-------|
| id | UUID | auto-generated | Primary key |
| match_id | UUID | required, must exist | Match reference |
| salary_amount | integer | required, > 0 | In cents (e.g., 150000 = $1500) |
| start_date | date | required | Proposed start date |
| expires_at | timestamp | required | Response deadline |
| status | enum | default: pending | pending, accepted, declined, expired |
| created_at | timestamp | auto | |
| updated_at | timestamp | auto | |

**Relationships**: belongs to Match

---

## State Transitions

### Match Status Flow

```
pending → expressed_interest → mutual_interest → negotiating → offer_sent → accepted/declined
                                    ↓
                            interview_scheduled → confirmed/cancelled
```

### Interview Status Flow

```
pending → confirmed (both parties accept) | cancelled (either party declines)
```

### Offer Status Flow

```
pending → accepted (candidate accepts) | declined (candidate declines) | expired (deadline passed)
```

---

## Validation Rules

- Job: title required, description required, structured.required[] must have at least 1 item
- Interview: scheduled_at must be in the future when creating
- Offer: expires_at must be after created_at + 24h minimum, salary_amount > 0
- All timestamps in UTC, stored as timestamptz in PostgreSQL
