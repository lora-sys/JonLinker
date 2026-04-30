# API Contract: WebSocket

**Feature**: specs/007-name-ui-completion
**Date**: 2026-04-28

## WebSocket /api/messages/ws

Real-time bidirectional messaging for A2A conversation.

### Connection

```
ws://localhost:8080/api/messages/ws?token=<jwt>
```

Authentication: JWT token passed as query parameter. Token contains user ID and role.

### Connection State

| State | Description |
|-------|-------------|
| `Connecting` | Initial connection attempt |
| `Connected` | WS connection open, authenticated |
| `Disconnected` | Connection lost, not reconnecting |
| `Reconnecting` | Attempting to reconnect with backoff |

### Client → Server Messages

#### Ping (keepalive)

```json
{
  "type": "ping"
}
```

#### Send Message

```json
{
  "type": "message",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "content_xml": "<A2AMessage><Payload><Intent>express_interest</Intent><Content>I'm interested in this position</Content><Context><TurnNumber>1</TurnNumber></Context></Payload></A2AMessage>",
  "intent_type": "express_interest"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | string | Yes | Always "message" |
| match_id | UUID | Yes | Match to send to |
| content_xml | string | Yes | A2A XML message format |
| intent_type | string | Yes | Intent classification |

### Server → Client Messages

#### Pong (keepalive response)

```json
{
  "type": "pong"
}
```

#### Incoming Message

```json
{
  "type": "message",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Hello! Thank you for your interest in the Senior Go Developer position...",
  "sender": "recruiter",
  "timestamp": "2026-04-28T10:05:00Z"
}
```

| Field | Type | Description |
|-------|------|-------------|
| type | string | Always "message" |
| match_id | UUID | Match thread |
| content | string | Human-readable message content |
| sender | string | "seeker" or "recruiter" |
| timestamp | ISO8601 | When message was sent |

#### Error

```json
{
  "type": "error",
  "message": "invalid match_id"
}
```

### Reconnection Strategy

Exponential backoff: 5s → 10s → 20s → 60s (max)

On reconnect, server sends any messages that arrived while disconnected (up to 100 messages).

### Status Indicator Display

```
Connecting:    Yellow badge "Connecting..."
Connected:     Green badge "Connected"
Disconnected:  Red badge "Disconnected" + Retry button
Reconnecting:  Yellow badge "Reconnecting (attempt 2/5)..."
```

### Debug Log Format

Each entry shows direction, timestamp, and content:

```
[10:05:00] → ping
[10:05:00] ← pong
[10:05:01] → message (express_interest)
[10:05:02] ← message from recruiter
```
