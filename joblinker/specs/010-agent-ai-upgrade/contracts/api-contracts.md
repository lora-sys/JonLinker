# API Contracts: AI Agent Capability Upgrade

**Feature**: specs/010-agent-ai-upgrade
**Date**: 2026-05-01

## Function Calling Schemas

### query_jobs

Search for jobs matching candidate preferences.

```json
{
  "name": "query_jobs",
  "description": "Search for jobs based on location, skills, salary range, and job type. Returns matching job listings with details.",
  "parameters": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "Desired work location (city, region, or 'remote')"
      },
      "skills": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Required skills for the job (e.g., ['python', 'django'])"
      },
      "salary_min": {
        "type": "integer",
        "description": "Minimum annual salary in USD"
      },
      "job_type": {
        "type": "string",
        "enum": ["remote", "hybrid", "onsite"],
        "description": "Preferred job type"
      },
      "limit": {
        "type": "integer",
        "default": 10,
        "description": "Maximum number of results to return"
      }
    },
    "required": ["location"]
  }
}
```

**Response Schema**:
```json
{
  "jobs": [
    {
      "id": "string",
      "title": "string",
      "company": "string",
      "location": "string",
      "salary_min": 80000,
      "salary_max": 120000,
      "job_type": "remote",
      "skills": ["python", "django"],
      "posted_at": "2026-04-15"
    }
  ]
}
```

---

### get_candidate

Retrieve candidate profile and qualifications.

```json
{
  "name": "get_candidate",
  "description": "Get detailed candidate profile including skills, experience, and preferences.",
  "parameters": {
    "type": "object",
    "properties": {
      "candidate_id": {
        "type": "string",
        "description": "Unique candidate identifier"
      }
    },
    "required": ["candidate_id"]
  }
}
```

**Response Schema**:
```json
{
  "candidate": {
    "id": "string",
    "name": "string",
    "skills": ["python", "django", "postgresql"],
    "experience_years": 5,
    "location": "London",
    "preferred_job_type": "hybrid",
    "salary_expectation": 95000
  }
}
```

---

### create_offer

Generate and send job offer to candidate.

```json
{
  "name": "create_offer",
  "description": "Create a job offer for a candidate match. Sends offer details to candidate.",
  "parameters": {
    "type": "object",
    "properties": {
      "match_id": {
        "type": "string",
        "description": "The match identifier linking candidate and job"
      },
      "salary": {
        "type": "integer",
        "description": "Annual salary offer in USD"
      },
      "start_date": {
        "type": "string",
        "description": "Proposed start date (YYYY-MM-DD)"
      },
      "notes": {
        "type": "string",
        "description": "Additional notes for the candidate"
      }
    },
    "required": ["match_id", "salary"]
  }
}
```

**Response Schema**:
```json
{
  "offer": {
    "id": "string",
    "match_id": "string",
    "status": "sent",
    "created_at": "2026-05-01T10:00:00Z"
  }
}
```

---

### schedule_interview

Book interview slot between candidate and recruiter.

```json
{
  "name": "schedule_interview",
  "description": "Schedule an interview between the candidate and recruiter.",
  "parameters": {
    "type": "object",
    "properties": {
      "match_id": {
        "type": "string",
        "description": "The match identifier"
      },
      "datetime": {
        "type": "string",
        "description": "Interview date and time (ISO 8601 format)"
      },
      "duration_minutes": {
        "type": "integer",
        "default": 60,
        "description": "Interview duration in minutes"
      },
      "interview_type": {
        "type": "string",
        "enum": ["video", "phone", "onsite"],
        "description": "Type of interview"
      }
    },
    "required": ["match_id", "datetime"]
  }
}
```

**Response Schema**:
```json
{
  "interview": {
    "id": "string",
    "match_id": "string",
    "scheduled_at": "2026-05-10T14:00:00Z",
    "duration_minutes": 60,
    "type": "video",
    "status": "confirmed"
  }
}
```

---

### search_candidates

Vector similarity search for candidates matching job requirements.

```json
{
  "name": "search_candidates",
  "description": "Find candidates matching job requirements using vector similarity search.",
  "parameters": {
    "type": "object",
    "properties": {
      "skills": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Required skills"
      },
      "location": {
        "type": "string",
        "description": "Preferred location"
      },
      "experience_min": {
        "type": "integer",
        "description": "Minimum years of experience"
      },
      "limit": {
        "type": "integer",
        "default": 10,
        "description": "Maximum candidates to return"
      }
    },
    "required": ["skills"]
  }
}
```

**Response Schema**:
```json
{
  "candidates": [
    {
      "id": "string",
      "name": "string",
      "skills": ["python", "django"],
      "location": "London",
      "experience_years": 5,
      "similarity_score": 0.89
    }
  ]
}
```

---

## Memory Service Contracts

### recall_preferences

Recall user preferences from long-term vector store.

**Request**:
```json
{
  "user_id": "string",
  "preference_types": ["salary", "location", "job_type"],
  "limit": 5
}
```

**Response**:
```json
{
  "preferences": [
    {
      "type": "location",
      "value": "London",
      "confidence": 0.92
    },
    {
      "type": "job_type",
      "value": "remote",
      "confidence": 0.88
    }
  ]
}
```

---

### store_preference

Store user preference as vector embedding.

**Request**:
```json
{
  "user_id": "string",
  "preference_type": "salary",
  "raw_value": "$100k+",
  "embedding": [0.123, -0.456, ...]
}
```

**Response**:
```json
{
  "id": "string",
  "stored_at": "2026-05-01T10:00:00Z"
}
```

---

## Context Management Contract

### compress_conversation

Summarize conversation and preserve key facts.

**Request**:
```json
{
  "match_id": "string",
  "messages": [
    { "role": "user", "content": "I want remote work" },
    { "role": "assistant", "content": "I'll search for remote jobs" }
  ],
  "token_count": 45000,
  "token_limit": 50000
}
```

**Response**:
```json
{
  "summary": {
    "summary_text": "User prefers remote work, Python development, $80k+ salary...",
    "key_facts": [
      { "fact_type": "job_type", "value": "remote", "priority": 1 },
      { "fact_type": "skills", "value": "python", "priority": 2 },
      { "fact_type": "salary", "value": "$80k+", "priority": 1 }
    ],
    "token_count_after": 27000,
    "compression_ratio": 0.4
  }
}
```