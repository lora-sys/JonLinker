# Feature Specification: API Gateway Unified Layer

**Feature Branch**: `011-api-gateway`  
**Created**: 2026-05-01  
**Status**: Draft  
**Input**: User description: "轻量 API Gateway 统一网关层"

## User Scenarios & Testing

### User Story 1 - Unified API Entry Point (Priority: P1)

A frontend developer or AI Agent makes API calls without knowing the underlying endpoint structure. They import the gateway client, configure it once, and all subsequent calls automatically include proper headers and routing.

**Why this priority**: This is the core value proposition - a single, consistent interface for all API access.

**Independent Test**: Can be tested by importing GatewayClient in isolation and verifying it makes correct requests with headers.

**Acceptance Scenarios**:

1. **Given** a GatewayClient configured with userId "user-123", **When** any API method is called, **Then** the request includes headers `X-User-ID: user-123`, `X-Agent-ID`, `X-Tenant-ID`
2. **Given** GatewayClient configured for route "agents.list", **When** the method is called, **Then** request goes to GET /api/agents
3. **Given** an invalid route name, **When** the method is called, **Then** TypeScript compiler errors at compile time

---

### User Story 2 - Multi-Agent Tool Calling (Priority: P1)

An AI Agent executing tools (schedule_interview, create_offer) uses the gateway to call backend APIs. Each tool call is routed through the gateway with proper agent identification headers.

**Why this priority**: Critical for A2A architecture where agents must have proper tenant/context isolation.

**Independent Test**: Agent can be tested by calling its tool_executor with gateway client and verifying correct API calls with headers.

**Acceptance Scenarios**:

1. **Given** agent "recruiter-1" making a tool call to schedule_interview, **When** the tool executes, **Then** gateway sends POST /api/interviews with `X-Agent-ID: recruiter-1`
2. **Given** agent from tenant "acme-corp" making a query, **When** the request is sent, **Then** `X-Tenant-ID: acme-corp` header is included

---

### User Story 3 - WebSocket Through Gateway (Priority: P2)

The frontend WebSocket connection for real-time messaging routes through the gateway, maintaining consistent header injection for connection authentication.

**Why this priority**: WebSocket is used for A2A real-time dialogue - must have same header context as HTTP calls.

**Independent Test**: WebSocket connection test verifies connection upgrade with proper query parameters.

**Acceptance Scenarios**:

1. **Given** GatewayClient configured with token, **When** connecting to WebSocket for match "match-456", **Then** connection URL includes token as query param
2. **Given** connected WebSocket, **When** a message is received, **Then** it contains proper A2A XML format

---

### User Story 4 - Rate Limiting & Observability (Priority: P3)

Gateway provides extension points for rate limiting (prevent abuse), circuit breaker (fault tolerance), and request logging (debugging/tracing).

**Why this priority**: Production readiness - these are standard gateway features for reliability.

**Independent Test**: Unit tests verify rate limiter blocks excess requests, circuit breaker opens after failures.

**Acceptance Scenarios**:

1. **Given** rate limit of 100 requests/minute, **When** 101 requests are made, **Then** 101st request is rejected with 429
2. **Given** circuit breaker with failure threshold 5, **When** 6 consecutive failures occur, **Then** subsequent requests fail fast with 503
3. **Given** logging enabled, **When** any request is made, **Then** entry is added to request log with method, path, userId, status, duration

---

### Edge Cases

- What happens when the backend server is unreachable? Gateway returns NetworkError with retry guidance.
- How does the gateway handle concurrent requests from multiple agents? Each request carries its own headers - no cross-contamination.
- What if a required header (X-User-ID) is missing? GatewayClient constructor requires userId - compile-time error if missing.
- How does WebSocket reconnection work if connection drops? WebSocket handler implements reconnection with exponential backoff.

## Requirements

### Functional Requirements

- **FR-001**: System MUST provide a single `GatewayClient` class as the sole entry point for all API calls from frontend, AI agents, and tools
- **FR-002**: System MUST automatically inject headers `X-User-ID`, `X-Agent-ID`, `X-Tenant-ID` on every request
- **FR-003**: System MUST support route mapping configuration for all service endpoints: agents, jobs, matches, interviews, offers, messages, privacy
- **FR-004**: System MUST support WebSocket upgrade through the gateway with same header context
- **FR-005**: Gateway MUST provide extension points for rate limiting, circuit breaker, and request logging
- **FR-006**: System MUST preserve existing backend behavior - no changes to handler, service, or repository logic

### Key Entities

- **GatewayClient**: Main API client class that encapsulates all HTTP/WebSocket communication
- **RouteService**: Configuration holding all route definitions mapped to service names
- **RequestInterceptor**: Hook interface for adding custom logic before requests (logging, rate limiting)
- **ResponseInterceptor**: Hook interface for processing responses (error handling, metrics)

## Success Criteria

### Measurable Outcomes

- **SC-001**: All frontend API calls (currently in api_client.ts, page components) route through GatewayClient
- **SC-002**: All AI Agent tool executions route through GatewayClient with proper agent headers
- **SC-003**: Headers X-User-ID, X-Agent-ID, X-Tenant-ID present on 100% of gateway requests
- **SC-004**: Zero changes to backend handler, service, repository code - only request routing changes
- **SC-005**: Existing features (RabbitMQ, WebSocket, vector DB, dual-agent) continue to work without modification
- **SC-006**: WebSocket connections through gateway work for real-time A2A messaging

## Assumptions

- Frontend and Backend run on same server in development (localhost:3000 frontend, localhost:8080 backend)
- Gateway configuration is set once at app initialization and doesn't change during session
- TypeScript is used for frontend code - route definitions provide compile-time safety
- Backend already has proper auth middleware that reads X-User-ID header (no backend changes needed)
- WebSocket uses token query parameter because WebSocket handshake doesn't support headers
