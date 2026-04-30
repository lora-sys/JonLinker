# API Contracts: Performance & Stability Optimization

## WebSocket: Real-time Messaging

### Connect
```
WS /api/messages/{matchId}/ws?token={jwt}
```

**Connection Lifecycle**:
1. Client connects with JWT token in query param
2. Server validates token and upgrades connection
3. Server sends `{"type": "connected", "payload": {"match_id": "..."}}`
4. Bidirectional message exchange begins

### Message Format (Client → Server)
```xml
<message>
  <header>
    <message_id>uuid</message_id>
  </header>
  <payload>
    <intent>INQUIRY|INTRODUCTION|INTEREST|NEGOTIATION|OFFER|ACCEPT|DECLINE</intent>
    <message_text>User message content</message_text>
  </payload>
</message>
```

### Message Format (Server → Client)
```xml
<message>
  <header>
    <message_id>uuid</message_id>
    <timestamp>ISO8601</timestamp>
    <sender_id>agent_id|system</sender_id>
  </header>
  <payload>
    <intent>INQUIRY|INTRODUCTION|INTEREST|NEGOTIATION|OFFER|ACCEPT|DECLINE|CONFIRM</intent>
    <parameters>{...json with message, title, salary_min, salary_max, location...}</parameters>
  </payload>
</message>
```

### Rate Limiting Response
When user exceeds 10 messages/minute:
```json
{
  "type": "error",
  "payload": {
    "code": "RATE_LIMITED",
    "message": "Too many messages. Please wait before sending more."
  }
}
```

---

## REST: Message API

### Send Message
```
POST /api/messages/{matchId}
Authorization: Bearer {jwt}
Content-Type: application/json

{
  "content_xml": "<message>...</message>",
  "intent_type": "INQUIRY"
}
```

**Responses**:
- `201 Created`: Message stored and queued for AI response
- `401 Unauthorized`: Invalid or missing token
- `429 Too Many Requests`: Rate limit exceeded

### Get Conversation History
```
GET /api/messages/{matchId}
Authorization: Bearer {jwt}
```

**Response**:
```json
[
  {
    "id": "uuid",
    "match_id": "uuid",
    "sender_agent_id": "uuid",
    "content_xml": "<message>...</message>",
    "intent_type": "INTRODUCTION",
    "created_at": "2026-04-30T12:00:00Z"
  }
]
```

---

## REST: Match API

### Create Match (Auto-Match)
```
POST /api/matches/auto
Authorization: Bearer {jwt}
Content-Type: application/json

{
  "seeker_agent_id": "uuid",
  "job_ids": ["uuid1", "uuid2"]
}
```

**Response**:
```json
{
  "matches": [
    {
      "id": "uuid",
      "seeker_agent_id": "uuid",
      "job_id": "uuid",
      "score": 0.85,
      "status": "pending"
    }
  ],
  "created_count": 1,
  "skipped_count": 0
}
```

### Get Match with Score Breakdown
```
GET /api/matches/{matchId}
Authorization: Bearer {jwt}
```

**Response**:
```json
{
  "id": "uuid",
  "seeker_agent_id": "uuid",
  "job_id": "uuid",
  "score": 0.85,
  "status": "pending",
  "seeker_agent": {
    "id": "uuid",
    "type": "seeker"
  },
  "job": {
    "id": "uuid",
    "structured": "{\"title\": \"Senior Engineer\", \"skills\": [\"Go\", \"Python\"]}"
  }
}
```

---

## Error Response Format

All API errors follow this structure:
```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "correlation_id": "uuid-for-log-tracing"
}
```

### Error Codes
- `UNAUTHORIZED`: Missing or invalid authentication
- `FORBIDDEN`: User lacks permission for this resource
- `NOT_FOUND`: Resource does not exist
- `RATE_LIMITED`: Too many requests
- `VALIDATION_ERROR`: Request body validation failed
- `INTERNAL_ERROR`: Unexpected server error
