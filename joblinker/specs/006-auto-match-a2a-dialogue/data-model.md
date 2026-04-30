# Data Model: Auto-Match & A2A Agent Dialogue

## Entity Overview

| Entity | Purpose | Key Fields |
|--------|---------|------------|
| Match | Links seeker-agent to job with AI score | seeker_agent_id, job_id, score, reasoning, status |
| Conversation | Thread of messages for a Match | match_id, last_message_at, unread_count |
| Message | Individual A2A communication | match_id, sender_agent_id, content_xml, intent_type, created_at |
| Interview | Scheduled interview | match_id, scheduled_at, format, status |
| Offer | Job offer with compensation | match_id, compensation, start_date, status |

## Match Entity

### Fields

| Field | Type | Description |
|-------|------|-------------|
| id | uuid | Primary key |
| seeker_agent_id | uuid | FK to Agent (seeker type) |
| job_id | uuid | FK to Job |
| score | decimal(5,4) | AI-calculated compatibility 0.0-1.0 |
| reasoning | text | AI explanation for score |
| status | varchar(30) | Match status enum |
| created_at | timestamp | Record creation time |
| updated_at | timestamp | Last update time |

### Status Enum

```
pending → expressed_interest → mutual_interest → negotiating → offer_sent → accepted/declined
```

**Transitions**:
- `pending`: Initial state when auto-match creates record
- `expressed_interest`: Seeker clicks "I'm Interested" on job
- `mutual_interest`: Recruiter confirms interest
- `negotiating`: Salary/benefits discussion underway
- `offer_sent`: Offer generated and sent
- `accepted`: Seeker accepts offer
- `declined`: Seeker or recruiter declines

## Message Entity

### Fields

| Field | Type | Description |
|-------|------|-------------|
| id | uuid | Primary key |
| match_id | uuid | FK to Match |
| sender_agent_id | uuid | FK to Agent (sender) |
| content_xml | text | XML-encoded message content |
| intent_type | varchar(50) | Message intent (INTRODUCTION, NEGOTIATION, etc.) |
| created_at | timestamp | Message creation time |

### Intent Types

| Intent | Description |
|--------|-------------|
| INTRODUCTION | Initial recruiter message to seeker |
| INTEREST | Seeker expressing interest |
| NEGOTIATION | Salary/compensation discussion |
| OFFER | Offer sent to seeker |
| ACCEPT | Acceptance of terms |
| DECLINE | Rejection of terms |
| SCHEDULE | Interview scheduling |
| CONFIRM | Confirmation of interview |
| WITHDRAW | Withdrawal from process |
| INQUIRY | General question |

## Relationships

```
User 1───* Agent
Agent 1───* Match
Agent 1───* Resume
Job 1─────* Match
Match 1───* Conversation
Conversation 1───* Message
Match 1───* Interview
Match 1───* Offer
```

## Validation Rules

- Match score must be 0.0-1.0
- Match (seeker_agent_id, job_id) must be unique
- Message requires non-empty content_xml
- Interview scheduled_at must be in future
- Offer start_date must be in future
