# API Contracts: Auto-Match & A2A Dialogue

## 1. POST /api/matches/auto

Auto-create Match records for seeker agent when viewing jobs.

**Request**:
```http
POST /api/matches/auto
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "job_ids": ["550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"]
}
```

**Response (200 OK)**:
```json
{
  "matches": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "job_id": "550e8400-e29b-41d4-a716-446655440001",
      "score": 0.85,
      "reasoning": "Strong skills alignment: Go, Kubernetes match. Missing: Docker experience",
      "status": "pending"
    }
  ],
  "created_count": 1,
  "skipped_count": 1
}
```

**Errors**:
- 401 Unauthorized: Invalid or missing JWT
- 400 Bad Request: Empty job_ids array
- 500 Internal Server Error: AI evaluation failed

---

## 2. GET /api/offers

List all offers for the authenticated user.

**Request**:
```http
GET /api/offers
Authorization: Bearer <jwt>
```

**Response (200 OK)**:
```json
{
  "offers": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440001",
      "match_id": "660e8400-e29b-41d4-a716-446655440001",
      "compensation": {
        "base_salary": 120000,
        "currency": "USD",
        "bonus": { "amount": 10000, "type": "annual" },
        "equity": { "shares": 5000, "type": "RSU" },
        "benefits": ["health", "dental", "401k"]
      },
      "start_date": "2026-06-01",
      "status": "pending",
      "created_at": "2026-04-26T10:00:00Z"
    }
  ]
}
```

**Errors**:
- 401 Unauthorized: Invalid or missing JWT

---

## 3. GET /api/messages

List all conversation threads for the authenticated user.

**Request**:
```http
GET /api/messages
Authorization: Bearer <jwt>
```

**Response (200 OK)**:
```json
{
  "conversations": [
    {
      "match_id": "660e8400-e29b-41d4-a716-446655440001",
      "job_title": "Senior Go Developer",
      "last_message": {
        "content": "We'd like to schedule an interview",
        "sender": "recruiter",
        "created_at": "2026-04-26T10:30:00Z"
      },
      "unread_count": 2,
      "updated_at": "2026-04-26T10:30:00Z"
    }
  ]
}
```

**Errors**:
- 401 Unauthorized: Invalid or missing JWT

---

## 4. WebSocket /api/messages/ws

General WebSocket connection for real-time message delivery.

**Connection**:
```http
ws://localhost:8080/api/messages/ws?token=<jwt>
```

**Server → Client Messages**:

```json
// Connection established
{"type": "connected", "payload": {"match_id": "660e8400-e29b-41d4-a716-446655440001"}}

// New message received
{"type": "message", "payload": {"id": "...", "content_xml": "...", "sender": "recruiter", "created_at": "..."}}

// Typing indicator
{"type": "typing", "payload": {"match_id": "...", "is_typing": true}}

// Agent AI response generating
{"type": "agent_thinking", "payload": {"match_id": "...", "agent": "seeker"}}
```

**Client → Server Messages**:

```json
// Send message
{"type": "send_message", "payload": {"match_id": "...", "content_xml": "...", "intent_type": "INTRODUCTION"}}

// Mark as read
{"type": "mark_read", "payload": {"match_id": "..."}}

// Typing indicator
{"type": "typing", "payload": {"match_id": "...", "is_typing": true}}
```

---

## 5. POST /api/messages/:matchId

Send a message in a conversation (REST alternative to WebSocket).

**Request**:
```http
POST /api/messages/:matchId
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "content_xml": "<message><header><sender_id>seeker-001</sender_id></header><payload><intent>INTRODUCTION</intent></payload></message>",
  "intent_type": "INTRODUCTION"
}
```

**Response (201 Created)**:
```json
{
  "id": "880e8400-e29b-41d4-a716-446655440001",
  "match_id": "660e8400-e29b-41d4-a716-446655440001",
  "sender_agent_id": "...",
  "content_xml": "...",
  "intent_type": "INTRODUCTION",
  "created_at": "2026-04-26T10:30:00Z"
}
```

---

## 6. Match Status Transitions

```
pending → expressed_interest → mutual_interest → negotiating → offer_sent → accepted/declined
```

**Transition Endpoints**:

```http
# Seeker expresses interest
POST /api/matches/:id/express-interest
Response: { "status": "expressed_interest" }

# Recruiter confirms mutual interest
POST /api/matches/:id/confirm-interest
Response: { "status": "mutual_interest" }

# Transition to negotiating
POST /api/matches/:id/negotiate
Response: { "status": "negotiating" }

# Send offer
POST /api/matches/:id/send-offer
Response: { "status": "offer_sent" }
```
