# API Contract: Interviews & Offers

**Feature**: specs/007-name-ui-completion
**Date**: 2026-04-28

## GET /api/interviews/:matchId

Get interview for a match.

### Response (200 OK)

```json
{
  "id": "880e8400-e29b-41d4-a716-446655440008",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "scheduled_at": "2026-05-01T14:00:00Z",
  "format": "video",
  "location": "https://meet.google.com/abc-defg-hij",
  "status": "pending",
  "reminder_sent": false,
  "created_at": "2026-04-28T10:00:00Z",
  "updated_at": "2026-04-28T10:00:00Z"
}
```

### Error Responses

| Status | Condition |
|--------|----------|
| 404 | No interview for this match |

---

## POST /api/interviews/:matchId/confirm

Confirm an interview invitation.

### Response (200 OK)

```json
{
  "id": "880e8400-e29b-41d4-a716-446655440008",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "scheduled_at": "2026-05-01T14:00:00Z",
  "format": "video",
  "location": "https://meet.google.com/abc-defg-hij",
  "status": "confirmed",
  "reminder_sent": false,
  "created_at": "2026-04-28T10:00:00Z",
  "updated_at": "2026-04-28T10:05:00Z"
}
```

---

## POST /api/interviews/:matchId/cancel

Cancel an interview.

### Response (200 OK)

```json
{
  "id": "880e8400-e29b-41d4-a716-446655440008",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "cancelled",
  "updated_at": "2026-04-28T10:10:00Z"
}
```

---

## GET /api/offers/:matchId

Get offer for a match.

### Response (200 OK)

```json
{
  "id": "990e8400-e29b-41d4-a716-446655440009",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "salary_amount": 150000,
  "start_date": "2026-06-01",
  "expires_at": "2026-05-05T23:59:59Z",
  "status": "pending",
  "created_at": "2026-04-28T10:00:00Z",
  "updated_at": "2026-04-28T10:00:00Z"
}
```

> **Note**: `salary_amount` is in cents (150000 = $1,500.00 per year)

### Error Responses

| Status | Condition |
|--------|----------|
| 404 | No offer for this match |

---

## POST /api/offers/:matchId/accept

Accept a job offer.

### Response (200 OK)

```json
{
  "id": "990e8400-e29b-41d4-a716-446655440009",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "salary_amount": 150000,
  "start_date": "2026-06-01",
  "status": "accepted",
  "updated_at": "2026-04-28T10:15:00Z"
}
```

---

## POST /api/offers/:matchId/decline

Decline a job offer.

### Response (200 OK)

```json
{
  "id": "990e8400-e29b-41d4-a716-446655440009",
  "match_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "declined",
  "updated_at": "2026-04-28T10:15:00Z"
}
```

---

## Interview Card UI Display

| Field | Display |
|-------|---------|
| `scheduled_at` | "May 1, 2026 at 2:00 PM" |
| `format` | Icon + label: 🎥 Video / 📍 Onsite / 📞 Phone |
| `location` | Clickable link or address |
| `status: pending` | Yellow "Pending Response" badge |
| `status: confirmed` | Green "Confirmed" badge |
| `status: cancelled` | Red "Cancelled" badge, dimmed |

Actions: [Confirm] [Decline] buttons when `status: pending`

---

## Offer Card UI Display

| Field | Display |
|-------|---------|
| `salary_amount` | "$150,000/year" (converted from cents) |
| `start_date` | "Start: June 1, 2026" |
| `expires_at` | "Expires: May 5, 2026 at 11:59 PM" (with urgency if <48h) |
| `status: pending` | Amber "Pending Your Response" |
| `status: accepted` | Green "Accepted" with checkmark |
| `status: declined` | Red "Declined" with X |
| `status: expired` | Gray "Expired" badge |

Actions: [Accept Offer] [Decline] buttons when `status: pending`

---

## State Transitions

### Interview

```
pending → confirmed  (user clicks Confirm)
pending → cancelled  (user clicks Decline)
```

### Offer

```
pending → accepted   (user clicks Accept)
pending → declined   (user clicks Decline)
pending → expired    (expires_at passes)
```
