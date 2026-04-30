# Quickstart: Auto-Match & A2A Agent Dialogue

## Prerequisites

- Backend running on port 8080
- Frontend running on port 3000
- PostgreSQL database with migrated schema
- AI API credentials in `.env` (AI_API_KEY, AI_BASE_URL)

## Test Scenarios

### Scenario 1: Auto-Match on Job Browse

**Steps**:
1. Login as seeker user with active agent
2. Visit `/jobs` page
3. System automatically calls `POST /api/matches/auto` with job IDs
4. Backend creates Match records for jobs with score > 0.5
5. Visit `/matches` page to see generated matches

**Expected Result**: Matches appear with AI-calculated scores and reasoning within 10 seconds

### Scenario 2: A2A Real-Time Dialogue

**Steps**:
1. Login as seeker, confirm a match (status: `expressed_interest`)
2. Recruiter agent auto-sends introduction via WebSocket
3. Visit `/messages` page - conversation thread visible
4. Send reply via `/messages/:matchId` endpoint
5. AI generates contextual response within 10 seconds

**Expected Result**: Messages delivered in real-time via WebSocket, AI responses generated

### Scenario 3: Interview Auto-Scheduling

**Steps**:
1. Match status reaches `mutual_interest`
2. Agents exchange 3+ messages
3. Backend auto-schedules interview
4. Visit `/interviews` page to see scheduled interview

**Expected Result**: Interview appears on page with datetime and format

### Scenario 4: Offer Generation

**Steps**:
1. Interview is confirmed by both parties
2. Backend generates offer based on job salary range
3. Visit `/offers` page to see offer with countdown timer
4. Accept or decline offer

**Expected Result**: Offer appears with compensation details and countdown timer

### Scenario 5: WebSocket Reconnection

**Steps**:
1. User connected to WebSocket on `/messages` page
2. Simulate network disconnect (devtools network tab)
3. Connection auto-recovers within 30 seconds
4. No messages are lost during disconnection

**Expected Result**: "Reconnecting..." indicator shown, connection restored automatically

## API Quick Reference

### Auto-Create Matches

```bash
POST /api/matches/auto
Authorization: Bearer <token>
Content-Type: application/json

{"job_ids": ["uuid1", "uuid2"]}
```

### List Offers

```bash
GET /api/offers
Authorization: Bearer <token>
```

### List Conversations

```bash
GET /api/messages
Authorization: Bearer <token>
```

### Send Message

```bash
POST /api/messages/:matchId
Authorization: Bearer <token>
Content-Type: application/json

{"content_xml": "<message>...</message>", "intent_type": "INTRODUCTION"}
```

### WebSocket Connection

```
ws://localhost:8080/api/messages/ws?token=<jwt>
```
