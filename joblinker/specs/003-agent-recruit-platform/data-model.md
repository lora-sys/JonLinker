# Data Model: A2A Agent Recruitment Platform

## Entity Relationship Diagram

```
User ─────────┐
  │           │
  │           │ 1
  │           ▼
  │    Organization
  │           │
  │           │ 1
  │           ▼
  └──────────►Agent◄─────────────┐
        │           │             │
        │ 1         │ 1           │
        │           │             │
        ▼           ▼             │
     Resume       Job             │
        │           │             │
        │           │ 1           │
        │           ▼             │
        │         Match◄──────────┘
        │           │       N:1 (to SeekerAgent)
        │           │
        │           │ 1
        │           ▼
        │        Message
        │           │
        │           │
        │     ┌─────┴─────┐
        │     │           │
        │     ▼           ▼
        │  Interview    Offer
        │
        └──────────────►SecurityEvent
```

## Entities

### User (USERS)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email address |
| password_hash | VARCHAR(255) | NOT NULL | Argon2 hashed password |
| role | ENUM | NOT NULL | 'seeker', 'recruiter', 'admin' |
| organization_id | UUID | FK, NULLABLE | Associated organization |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

### Organization (ORGANIZATIONS)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| name | VARCHAR(255) | NOT NULL | Company name |
| admin_user_id | UUID | FK | Admin user reference |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |

### Agent (AGENTS)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| user_id | UUID | FK, NOT NULL | Owner user reference |
| type | ENUM | NOT NULL | 'seeker', 'recruiter' |
| status | ENUM | NOT NULL | 'active', 'paused' |
| config_json | JSONB | | FSM config, preferences |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

**Agent Status Transitions**:
```
seeker: IDLE → PROFILE_BUILT → MATCHING → NEGOTIATING → INTERVIEW_SCHEDULED → OFFER_RECEIVED → HIRED/REJECTED
                     ↓              ↓
                 PAUSED         DEADLOCKED

recruiter: IDLE → POSTING_ACTIVE → MATCHING → NEGOTIATING → INTERVIEW_SCHEDULED → OFFER_EXTENDED → FILLED/REJECTED
                      ↓              ↓
                  PAUSED         DEADLOCKED
```

### Resume (RESUMES)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| agent_id | UUID | FK, UNIQUE | Associated seeker agent |
| structured_json | JSONB | NOT NULL | Skills, experience, education |
| vector_id | VARCHAR(255) | | Chroma vector reference |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

**Note**: Original resume file stored client-side in IndexedDB, encrypted with AES-256. Only structured_json (anonymized) and vector_id sent to server.

### Job (JOBS)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| agent_id | UUID | FK, NOT NULL | Associated recruiter agent |
| structured_json | JSONB | NOT NULL | Requirements, responsibilities |
| vector_id | VARCHAR(255) | | Chroma vector reference |
| status | ENUM | NOT NULL | 'draft', 'active', 'paused', 'filled', 'closed' |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

### Match (MATCHES)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| seeker_agent_id | UUID | FK, NOT NULL | Seeker agent reference |
| job_id | UUID | FK, NOT NULL | Job reference |
| score | DECIMAL(5,4) | NOT NULL | Compatibility score (0-1) |
| status | ENUM | NOT NULL | 'pending', 'mutual_interest', 'negotiating', 'offered', 'hired', 'rejected' |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

**Match Status Transitions**:
```
PENDING → MUTUAL_INTEREST → NEGOTIATING → OFFERED → HIRED
    ↓            ↓              ↓
  REJECTED    REJECTED      REJECTED
```

### Message (MESSAGES)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| match_id | UUID | FK, NOT NULL | Parent match reference |
| sender_agent_id | UUID | FK, NOT NULL | Sending agent reference |
| content_xml | TEXT | NOT NULL | A2A XML message content |
| intent_type | VARCHAR(50) | | Detected intent (salary/schedule/requirements) |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |

### Interview (INTERVIEWS)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| match_id | UUID | FK, UNIQUE, NOT NULL | Associated match |
| scheduled_at | TIMESTAMP | NOT NULL | Interview time |
| format | ENUM | NOT NULL | 'video', 'phone', 'onsite' |
| location | VARCHAR(500) | | Link or address |
| status | ENUM | NOT NULL | 'scheduled', 'completed', 'cancelled', 'rescheduled' |
| feedback | JSONB | | Ratings, notes |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

### Offer (OFFERS)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| match_id | UUID | FK, UNIQUE, NOT NULL | Associated match |
| compensation_json | JSONB | NOT NULL | Salary, benefits, equity |
| start_date | DATE | NOT NULL | Proposed start date |
| status | ENUM | NOT NULL | 'pending', 'accepted', 'declined', 'negotiating', 'withdrawn' |
| responded_at | TIMESTAMP | | Response timestamp |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

### SecurityEvent (SECURITY_EVENTS)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK | Unique identifier |
| user_id | UUID | FK, NOT NULL | User reference |
| action_type | VARCHAR(50) | NOT NULL | Event type |
| details_json | JSONB | | Additional context |
| ip_address | VARCHAR(45) | | Client IP |
| created_at | TIMESTAMP | NOT NULL | Event timestamp |

**Event Types**: login, logout, password_change, profile_update, data_export, data_delete, agent_create, agent_pause, agent_resume, match_created, interview_scheduled, offer_sent, offer_accepted, offer_declined

## Validation Rules

| Entity | Rule |
|--------|------|
| User.email | Valid email format, max 255 chars |
| User.password_hash | Argon2 hash output, min 32 chars |
| Agent.type | Must be 'seeker' or 'recruiter' |
| Agent.status | Must be 'active' or 'paused' |
| Match.score | Decimal between 0.0000 and 1.0000 |
| Interview.scheduled_at | Must be future timestamp |
| Offer.start_date | Must be future date |
| SecurityEvent.action_type | Must be from defined event types |

## Indexes

- `idx_agents_user_id` on AGENTS(user_id)
- `idx_agents_status` on AGENTS(status)
- `idx_jobs_agent_id` on JOBS(agent_id)
- `idx_jobs_status` on JOBS(status)
- `idx_matches_seeker` on MATCHES(seeker_agent_id)
- `idx_matches_job` on MATCHES(job_id)
- `idx_matches_status` on MATCHES(status)
- `idx_messages_match` on MESSAGES(match_id)
- `idx_security_user` on SECURITY_EVENTS(user_id)
- `idx_security_created` on SECURITY_EVENTS(created_at)

## Constraints

- Unique constraint: One resume per seeker agent
- Unique constraint: One interview per match
- Unique constraint: One offer per match
- Foreign key: Agent.user_id → User.id (cascade delete)
- Foreign key: Resume.agent_id → Agent.id (cascade delete)
- Foreign key: Job.agent_id → Agent.id (cascade delete)