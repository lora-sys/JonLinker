# Quickstart: Performance & Stability Optimization

## Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- RabbitMQ 3.13+

## Development Setup

### 1. Start Infrastructure
```bash
docker-compose up -d
```

### 2. Start Backend
```bash
cd backend
go run cmd/server/main.go
```

### 3. Start Frontend
```bash
cd frontend
npm run dev
```

## Testing the Flow

### 1. Create User & Agent
```bash
# Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","role":"seeker"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Create agent (use token from login)
curl -X POST http://localhost:8080/api/agents \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"type":"seeker"}'
```

### 2. Create Job & Match
```bash
# Create job (as recruiter)
curl -X POST http://localhost:8080/api/jobs \
  -H "Authorization: Bearer {recruiter_token}" \
  -H "Content-Type: application/json" \
  -d '{"structured":"{\"title\":\"Senior Engineer\",\"skills\":[\"Go\",\"Python\"]}"}'

# Auto-match
curl -X POST http://localhost:8080/api/matches/auto \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"seeker_agent_id":"{agent_id}","job_ids":["{job_id}"]}'
```

### 3. Test WebSocket Conversation
```javascript
// In browser console or WebSocket client
const ws = new WebSocket('ws://localhost:8080/api/messages/{matchId}/ws?token={token}');

ws.onopen = () => console.log('Connected');
ws.onmessage = (e) => console.log('Received:', e.data);

// Send message
ws.send('<message><header><message_id>123</message_id></header><payload><intent>INQUIRY</intent><message_text>Hello</message_text></payload></message>');
```

## Performance Testing

### Load Test
```bash
# Using hey or ab
hey -n 1000 -c 10 http://localhost:8080/api/matches
```

### WebSocket Stress Test
```javascript
// Open multiple connections
for (let i = 0; i < 100; i++) {
  new WebSocket(`ws://localhost:8080/api/messages/{matchId}/ws?token={token}`);
}
```

## Monitoring

### Check RabbitMQ Queues
```bash
docker exec joblinker-rabbitmq rabbitmqctl list_queues
```

### Check PostgreSQL Connections
```sql
SELECT count(*), state FROM pg_stat_activity GROUP BY state;
```

### View Error Logs
```bash
# Backend logs include correlation IDs
tail -f backend.log | grep correlation_id
```
