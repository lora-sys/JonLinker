# Feature Specification: Production Observability & Agent Monitoring

**Feature Branch**: `012-production-observability`
**Created**: 2026-05-01
**Status**: Draft

## User Scenarios & Testing

### User Story 1 - Agent Health Dashboard (Priority: P1)

Operations team monitors health of all running agents in real-time. Dashboard shows: active conversations, messages processed, AI response times, error rates.

**Acceptance Scenarios**:

1. **Given** admin opens dashboard, **When** page loads, **Then** shows count of active seeker agents, recruiter agents, and ongoing conversations
2. **Given** agent is failing, **When** error rate exceeds threshold, **Then** dashboard highlights agent in red with error count
3. **Given** agent processes message, **When** AI responds, **Then** response time is logged and displayed

---

### User Story 2 - Conversation Audit Trail (Priority: P1)

Every agent decision (tool call, response generation, negotiation move) is logged for compliance and debugging. Logs are queryable by match_id, agent_id, time range.

**Acceptance Scenarios**:

1. **Given** compliance officer queries audit trail, **When** searching by match_id, **Then** returns all messages, decisions, tool calls for that conversation
2. **Given** debugging a failed negotiation, **When** reviewing audit, **Then** can see each negotiation round with agent reasoning

---

### User Story 3 - Error Tracking & Alerting (Priority: P2)

System automatically tracks errors (AI failures, RabbitMQ issues, database timeouts) and alerts ops team. Errors are categorized and deduplicated.

**Acceptance Scenarios**:

1. **Given** AI API returns error, **When** error occurs, **Then** error is logged with full context (match_id, agent_id, intent, error message)
2. **Given** error rate spikes, **When** threshold exceeded (10 errors/min), **Then** alert sent to configured webhook

---

### User Story 4 - Agent Performance Metrics (Priority: P2)

Collect metrics on agent performance: average response time, tool call success rate, conversation completion rate, negotiation success rate.

**Acceptance Scenarios**:

1. **Given** reviewing agent performance, **When** viewing metrics, **Then** sees average AI response time per agent type
2. **Given** comparing recruiter vs seeker agents, **When** viewing dashboard, **Then** sees completion rate for each

---

### User Story 5 - Multi-Agent Coordination View (Priority: P3)

Visualize agent-to-agent conversations as a graph. See which agents are talking to which, message flow, and negotiation states.

**Acceptance Scenarios**:

1. **Given** ops monitoring conversation, **When** viewing graph, **Then** sees nodes (agents) connected by edges (messages) with direction
2. **Given** agent stuck in negotiation, **When** viewing detail, **Then** sees negotiation state and last 5 messages

---

## Requirements

### Functional Requirements

- **FR-001**: System MUST provide admin dashboard at `/admin` showing agent health metrics
- **FR-002**: System MUST log all agent decisions with timestamp, agent_id, match_id, decision_type, reasoning
- **FR-003**: System MUST track error events with full context and stack traces
- **FR-004**: System MUST support querying audit logs by match_id, agent_id, time range
- **FR-005**: System MUST calculate and display agent performance metrics
- **FR-006**: System MUST support webhook alerting for error spikes
- **FR-007**: System MUST provide WebSocket-based real-time dashboard updates

### Data Model

#### AgentMetrics
- `agent_id`: UUID
- `agent_type`: seeker | recruiter
- `messages_processed`: int
- `errors_count`: int
- `avg_response_time_ms`: int
- `conversations_active`: int
- `conversations_completed`: int
- `last_heartbeat`: timestamp

#### AuditLog
- `id`: UUID
- `agent_id`: UUID
- `match_id`: UUID
- `event_type`: message_sent | tool_call | decision | error
- `event_data`: JSONB (full context)
- `timestamp`: timestamp

#### ErrorEvent
- `id`: UUID
- `error_type`: ai_failure | rabbitmq_failure | db_timeout | tool_error
- `error_message`: string
- `stack_trace`: text
- `context`: JSONB (match_id, agent_id, intent)
- `resolved`: boolean
- `timestamp`: timestamp

## Success Criteria

- **SC-001**: Admin dashboard loads in < 2 seconds with all metrics visible
- **SC-002**: Audit log query by match_id returns results in < 500ms
- **SC-003**: Error events are logged within 100ms of occurrence
- **SC-004**: WebSocket updates appear on dashboard within 1 second of event
- **SC-005**: All 7 functional requirements implemented and testable

## Edge Cases

- What if AI API is down? Show "degraded" status, queue messages for retry
- What if too many audit logs? Implement retention policy (30 days default)
- What if dashboard has many concurrent viewers? Use efficient broadcast via WebSocket

## Assumptions

- PostgreSQL used for metrics and audit log storage
- Redis not available - use in-memory caching with TTL
- Webhook URLs configured via environment variable
- Retention: audit logs kept 30 days, errors kept 90 days
