# A2A Communication Protocol

## Overview

Seeker Agents and Recruiter Agents communicate via structured XML messages over WebSocket. This protocol enables autonomous negotiation of salary, scheduling, and requirements without human intervention.

## Message Envelope

All A2A messages follow this envelope structure:

```xml
<a2a-message
  type="negotiate | schedule | offer | system"
  id="msg-UNIQUE-ID"
  timestamp="ISO-8601-DATETIME"
  ref-id="msg-REFERENCE-ID"
>
  <sender agent-id="agent-UUID" type="seeker | recruiter"/>
  <recipient agent-id="agent-UUID" type="seeker | recruiter"/>
  <payload>
    <!-- Type-specific content -->
  </payload>
</a2a-message>
```

### Envelope Attributes

| Attribute | Required | Description |
|-----------|----------|-------------|
| type | Yes | Message category |
| id | Yes | Globally unique message ID (UUID format recommended) |
| timestamp | Yes | ISO 8601 UTC timestamp |
| ref-id | No | ID of message being replied to (for responses) |

### Message Types

| Type | Direction | Purpose |
|------|----------|---------|
| negotiate | Bidirectional | Salary, requirements negotiation |
| schedule | Bidirectional | Interview time coordination |
| offer | Recruiter→Seeker | Formal job offer |
| system | System→Both | Error, timeout, escalation notifications |

## Intent Structure

Each message payload contains an `<intent>` element identifying the communication goal:

```xml
<intent type="salary | schedule | requirements | general" confidence="0.00-1.00">
  <proposal value="..." />
  <!-- or -->
  <counter value="..." />
  <!-- or -->
  <question value="..." />
</intent>
```

### Intent Types

| Type | Description | Typical Flow |
|------|-------------|--------------|
| salary | Compensation negotiation | propose → counter → accept |
| schedule | Interview timing | propose → accept/modify |
| requirements | Job qualifications discussion | clarify → confirm |
| general | Other professional dialogue | query → response |

## Negotiation Patterns

### Salary Negotiation

```xml
<!-- Seeker proposes salary range -->
<a2a-message type="negotiate" id="msg-001" timestamp="2026-04-22T10:00:00Z">
  <sender agent-id="seeker-001" type="seeker"/>
  <recipient agent-id="recruiter-001" type="recruiter"/>
  <payload>
    <intent type="salary" confidence="0.95">
      <proposal range="$80k-$95k" currency="USD" />
      <reasoning>Based on market research and experience level</reasoning>
    </intent>
  </payload>
</a2a-message>

<!-- Recruiter counter-offers -->
<a2a-message type="negotiate" id="msg-002" timestamp="2026-04-22T10:05:00Z" ref-id="msg-001">
  <sender agent-id="recruiter-001" type="recruiter"/>
  <recipient agent-id="seeker-001" type="seeker"/>
  <payload>
    <intent type="salary" confidence="0.88">
      <counter range="$85k-$100k" currency="USD" />
      <reasoning>Budget allocation for this role</reasoning>
    </intent>
  </payload>
</a2a-message>

<!-- Seeker accepts -->
<a2a-message type="negotiate" id="msg-003" timestamp="2026-04-22T10:10:00Z" ref-id="msg-002">
  <sender agent-id="seeker-001" type="seeker"/>
  <recipient agent-id="recruiter-001" type="recruiter"/>
  <payload>
    <intent type="salary" confidence="1.00">
      <accept range="$85k-$100k" currency="USD" />
    </intent>
  </payload>
</a2a-message>
```

### Schedule Negotiation

```xml
<!-- Recruiter proposes interview times -->
<a2a-message type="schedule" id="msg-010" timestamp="2026-04-22T11:00:00Z">
  <sender agent-id="recruiter-001" type="recruiter"/>
  <recipient agent-id="seeker-001" type="seeker"/>
  <payload>
    <intent type="schedule" confidence="0.90">
      <proposal>
        <option time="2026-04-25T14:00:00Z" timezone="America/New_York" format="video"/>
        <option time="2026-04-26T10:00:00Z" timezone="America/New_York" format="video"/>
      </proposal>
    </intent>
  </payload>
</a2a-message>

<!-- Seeker selects option -->
<a2a-message type="schedule" id="msg-011" timestamp="2026-04-22T11:15:00Z" ref-id="msg-010">
  <sender agent-id="seeker-001" type="seeker"/>
  <recipient agent-id="recruiter-001" type="recruiter"/>
  <payload>
    <intent type="schedule" confidence="1.00">
      <accept time="2026-04-25T14:00:00Z" timezone="America/New_York" format="video"/>
    </intent>
  </payload>
</a2a-message>
```

### Offer Delivery

```xml
<!-- Recruiter sends formal offer -->
<a2a-message type="offer" id="msg-020" timestamp="2026-04-22T16:00:00Z">
  <sender agent-id="recruiter-001" type="recruiter"/>
  <recipient agent-id="seeker-001" type="seeker"/>
  <payload>
    <offer>
      <compensation>
        <base salary="$90,000" currency="USD"/>
        <bonus target="10%" />
        <equity units="1000" vesting="4 years" />
      </compensation>
      <start_date>2026-05-15</start_date>
      <benefits>Health, dental, 401k matching</benefits>
      <response_deadline>2026-04-24T23:59:59Z</response_deadline>
    </offer>
  </payload>
</a2a-message>

<!-- Seeker accepts -->
<a2a-message type="offer" id="msg-021" timestamp="2026-04-23T09:00:00Z" ref-id="msg-020">
  <sender agent-id="seeker-001" type="seeker"/>
  <recipient agent-id="recruiter-001" type="recruiter"/>
  <payload>
    <intent type="offer" confidence="1.00">
      <accept />
      <confirmation_message>Excited to join the team!</confirmation_message>
    </intent>
  </payload>
</a2a-message>
```

### System Notifications

```xml
<!-- Negotiation timeout -->
<a2a-message type="system" id="msg-sys-001" timestamp="2026-04-22T18:00:00Z">
  <recipient agent-id="seeker-001" type="seeker"/>
  <recipient agent-id="recruiter-001" type="recruiter"/>
  <payload>
    <notification type="deadlock">
      <reason>Maximum negotiation rounds (5) reached without agreement</reason>
      <escalation>Human review required</escalation>
    </notification>
  </payload>
</a2a-message>
```

## Error Handling

| Error Code | Meaning | Action |
|------------|---------|--------|
| E001 | Invalid message format | Reject with schema validation errors |
| E002 | Unknown recipient agent | Return error, do not retry |
| E003 | Stale message (ref-id not found) | Return error with current state |
| E004 | Unauthorized sender | Reject, log security event |
| E005 | System error during processing | Retry with exponential backoff |

```xml
<a2a-message type="system" id="msg-err-001" timestamp="2026-04-22T10:00:00Z" ref-id="msg-001">
  <payload>
    <error code="E001">
      <reason>Invalid salary format</reason>
      <detail>Expected format: "$XXk-$YYk"</detail>
    </error>
  </payload>
</a2a-message>
```

## Human Confirmation Triggers

Agents request human confirmation at these decision points:

| Stage | Trigger | Confirmation Type |
|-------|---------|-------------------|
| Mutual Interest | Match score above threshold | User confirms match participation |
| Negotiation Complete | Agreement reached on terms | User approves negotiated values |
| Interview Scheduled | Time slot selected | User confirms interview |
| Offer Extended | Compensation finalized | User reviews and approves offer |
| Offer Accepted | Seeker accepts offer | Both parties notified |

```xml
<a2a-message type="system" id="msg-confirm-001" timestamp="2026-04-22T10:15:00Z">
  <recipient agent-id="seeker-001" type="seeker"/>
  <payload>
    <confirmation_required type="salary_agreement">
      <negotiated_salary range="$85k-$100k" currency="USD" />
      <user_action>Required: APPROVE | REJECT | COUNTER</user_action>
    </confirmation_required>
  </payload>
</a2a-message>
```

## Protocol Version

Current version: 1.0

Version negotiation: Client sends highest supported version in initial connection; server responds with version to use (minimum of client highest and server highest supported).