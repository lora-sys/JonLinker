# API Contract: Jobs

**Feature**: specs/007-name-ui-completion
**Date**: 2026-04-28

## POST /api/jobs

Create a new job posting.

### Request

```json
{
  "title": "Senior Go Developer",
  "description": "We are looking for a senior Go developer to join our team...",
  "location": "San Francisco, CA (Remote OK)",
  "type": "full-time",
  "salary_range": "120k-180k",
  "structured": {
    "requirements": ["Go", "Kubernetes", "PostgreSQL"],
    "nice_to_have": ["Rust", "AWS", "Terraform"],
    "location": "San Francisco, CA",
    "work_type": "remote",
    "salary_range": { "min": 120000, "max": 180000 }
  }
}
```

### Response (201 Created)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "agent_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Senior Go Developer",
  "description": "We are looking for a senior Go developer...",
  "location": "San Francisco, CA (Remote OK)",
  "type": "full-time",
  "salary_range": "120k-180k",
  "structured": {
    "requirements": ["Go", "Kubernetes", "PostgreSQL"],
    "nice_to_have": ["Rust", "AWS", "Terraform"],
    "location": "San Francisco, CA",
    "work_type": "remote",
    "salary_range": { "min": 120000, "max": 180000 }
  },
  "vector_id": "770e8400-e29b-41d4-a716-446655440002",
  "status": "active",
  "created_at": "2026-04-28T10:00:00Z",
  "updated_at": "2026-04-28T10:00:00Z"
}
```

### Error Responses

| Status | Condition | Response |
|--------|----------|----------|
| 400 | Missing required fields | `{"error": "title is required"}` |
| 400 | Invalid type | `{"error": "type must be full-time, part-time, or contract"}` |
| 400 | Empty requirements | `{"error": "structured.requirements must have at least 1 item"}` |
| 403 | Not a recruiter | `{"error": "only recruiters can post jobs"}` |

---

## GET /api/jobs

List all active jobs.

### Response (200 OK)

```json
{
  "jobs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "agent_id": "660e8400-e29b-41d4-a716-446655440001",
      "title": "Senior Go Developer",
      "location": "San Francisco, CA (Remote OK)",
      "type": "full-time",
      "salary_range": "120k-180k",
      "status": "active",
      "created_at": "2026-04-28T10:00:00Z"
    }
  ]
}
```

---

## GET /api/jobs/:id

Get job details.

### Response (200 OK)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "agent_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Senior Go Developer",
  "description": "We are looking for a senior Go developer...",
  "location": "San Francisco, CA (Remote OK)",
  "type": "full-time",
  "salary_range": "120k-180k",
  "structured": {
    "requirements": ["Go", "Kubernetes", "PostgreSQL"],
    "nice_to_have": ["Rust", "AWS", "Terraform"],
    "location": "San Francisco, CA",
    "work_type": "remote",
    "salary_range": { "min": 120000, "max": 180000 }
  },
  "status": "active",
  "created_at": "2026-04-28T10:00:00Z",
  "updated_at": "2026-04-28T10:00:00Z"
}
```

### Error Responses

| Status | Condition |
|--------|----------|
| 404 | Job not found |
