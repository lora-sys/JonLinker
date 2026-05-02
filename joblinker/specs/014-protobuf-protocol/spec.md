# Feature Specification: Protobuf Communication Protocol

**Feature Branch**: `014-protobuf-protocol`  
**Created**: 2026-05-02  
**Status**: Draft  
**Input**: User description: "在现有 A2A 招聘系统中，新增 Protobuf 通信协议，用于减少高频通信带宽、提升系统性能"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Internal Service Communication Optimization (Priority: P1)

System operators want internal services (API Gateway to backend, backend to RabbitMQ, backend to vector database) to communicate using Protobuf instead of JSON, reducing bandwidth usage and serialization overhead for high-frequency operations.

**Why this priority**: Internal service communication carries the highest message volume and offers immediate performance gains with no user-facing changes required.

**Independent Test**: Measure bandwidth reduction between API Gateway and backend services when processing identical message loads, comparing JSON vs Protobuf modes via network monitoring.

**Acceptance Scenarios**:

1. **Given** a Protobuf-enabled internal service channel, **When** a match scoring request is sent, **Then** the request is serialized as Protobuf and correctly deserialized by the receiver with all fields intact
2. **Given** both JSON and Protobuf endpoints available, **When** a request includes `Accept: application/x-protobuf` header, **Then** the system responds with Protobuf binary format
3. **Given** a RabbitMQ task queue, **When** a job matching task is enqueued with Protobuf mode enabled, **Then** the message body uses Protobuf serialization and consumers parse it correctly

---

### User Story 2 - WebSocket Agent Dialogue with Cache Optimization (Priority: P2)

Dual agents (Seeker and Recruiter) exchange messages via WebSocket using Protobuf format, with tool call results transmitted as cache keys and summaries instead of full payloads.

**Why this priority**: Agent-to-agent WebSocket traffic is high-volume (negotiation rounds) and the primary source of context window bloat. Combining Protobuf with cache key optimization maximizes token reduction.

**Independent Test**: Monitor WebSocket message sizes during a full negotiation session, verify tool results are replaced with cache keys, confirm agents retrieve full data on demand via cache lookup.

**Acceptance Scenarios**:

1. **Given** two agents in an active negotiation, **When** the Seeker agent makes a `query_jobs` tool call, **Then** the result is stored in cache and only `cache_key` + summary is transmitted via WebSocket
2. **Given** an agent receives a message with a cache key reference, **When** the agent needs full tool result data, **Then** it retrieves it via cache lookup using the key
3. **Given** a WebSocket message encoded as Protobuf, **When** the receiving agent parses it, **Then** message content, intent type, and tool references are correctly extracted

---

### User Story 3 - FSM State Change Notifications (Priority: P3)

Match state machine transitions (e.g., negotiating → interviewing → offered) are broadcast via Protobuf messages to all relevant agents and frontend clients.

**Why this priority**: State change notifications are frequent and benefit from Protobuf's compact size, but the functional impact is lower than core message traffic.

**Independent Test**: Trigger a state transition and verify all subscribed listeners receive a correctly formatted Protobuf notification with correct old_state, new_state, and match_id fields.

**Acceptance Scenarios**:

1. **Given** a match in `negotiating` state, **When** both agents agree on terms, **Then** all subscribed parties receive a Protobuf state change notification to `offer_received`
2. **Given** a memory update triggered by preference extraction, **When** a user preference is stored in the vector database, **Then** the memory service sends a Protobuf notification of the update

---

### User Story 4 - Gradual Migration with Backward Compatibility (Priority: P4)

Development team can deploy Protobuf alongside existing JSON interfaces with zero disruption, using content negotiation to select format per request.

**Why this priority**: Production safety requires gradual rollout; this enables teams to test Protobuf on staging before full production cutover.

**Independent Test**: Configure a client to request JSON while the server supports Protobuf, verify JSON responses remain byte-for-byte identical to pre-Protobuf deployment.

**Acceptance Scenarios**:

1. **Given** a server supporting both protocols, **When** a client sends no `Accept` header, **Then** the server responds with JSON (backward compatible default)
2. **Given** a server supporting both protocols, **When** a client sends `Accept: application/x-protobuf`, **Then** the server responds with Protobuf binary
3. **Given** an invalid Protobuf byte sequence, **When** the server attempts to parse it, **Then** it returns a clear error message indicating the parse failure with the byte offset

### Edge Cases

- What happens when a client sends a Protobuf message with missing required fields? → Server returns a validation error listing the missing field names.
- How does the system handle version mismatch between Protobuf definitions? → Include a `schema_version` field in every top-level message; reject messages with incompatible versions and return the server's supported version.
- What happens when a cache key references non-existent or expired cached data? → System returns a cache-miss error; the caller re-fetches data via the original API endpoint and may re-populate the cache.
- How does fallback work when Protobuf parsing fails mid-WebSocket-stream? → Close the WebSocket connection gracefully with a protocol error frame, log the error, and let the client reconnect with JSON fallback.
- What happens during the migration window when some services speak Protobuf and others still speak JSON? → The API Gateway handles conversion; internal services advertise their protocol capability, and the gateway transcodes when necessary.

## Requirements *(mandatory)*

### Functional Requirements

**Protocol Schema**:

- **FR-001**: System MUST define Protobuf schema files in a `proto/` directory with the following files: `agent.proto`, `tools.proto`, `websocket.proto`, `queue.proto`
- **FR-002**: `agent.proto` MUST define AgentRole enum (SEEKER, RECRUITER), AgentState enum (IDLE, SEARCHING, NEGOTIATING, INTERVIEWING, OFFER_RECEIVED, HIRED, REJECTED), and message types: TextMessage, ToolCall, ToolResult, StateChangeEvent, MemoryUpdate
- **FR-003**: `tools.proto` MUST define JobDetail, ResumeDetail, InterviewInvitation, OfferDetail message structures corresponding to the tool schemas in the existing function calling system
- **FR-004**: `websocket.proto` MUST define a WebSocketFrame envelope with fields: message_type (enum), payload (bytes), sequence_num (uint64), timestamp (int64), cache_key (string, optional)
- **FR-005**: `queue.proto` MUST define a QueueTask message for RabbitMQ with fields: task_type (string), priority (enum: LOW, NORMAL, HIGH, CRITICAL), payload (bytes), retry_count (uint32), max_retries (uint32)

**Code Generation**:

- **FR-006**: System MUST provide a script (`scripts/generate-proto.sh`) that generates TypeScript code from .proto files using protoc with ts-proto plugin, and Go code using protoc-gen-go
- **FR-007**: Generated TypeScript code MUST be placed in `frontend/src/lib/proto/` directory
- **FR-008**: Generated Go code MUST be placed in `backend/pkg/proto/` directory
- **FR-009**: The generation script MUST validate that protoc and required plugins (protoc-gen-go, protoc-gen-ts) are installed before execution, and print clear installation instructions if missing

**Integration**:

- **FR-010**: API Gateway MUST support `Content-Type: application/x-protobuf` and `Accept: application/x-protobuf` headers to negotiate Protobuf encoding per-request
- **FR-011**: API Gateway MUST default to JSON format (when no Accept header is present) to maintain full backward compatibility
- **FR-012**: RabbitMQ message publishing MUST support optional Protobuf serialization controlled by environment variable `QUEUE_PROTOBUF_ENABLED` (default: false)
- **FR-013**: WebSocket handler MUST detect message format by inspecting the first byte (0x0A indicates Protobuf varint length-delimited encoding) and route to the appropriate parser
- **FR-014**: Tool execution results MUST be stored in the existing cache layer; transmitted WebSocket messages MUST include only `cache_key` and a human-readable summary, not the full payload
- **FR-015**: FSM state change events MUST be broadcast using the Protobuf StateChangeEvent message when Protobuf mode is active
- **FR-016**: Memory update notifications from the vector memory service MUST use the Protobuf MemoryUpdate message format

**Configuration**:

- **FR-017**: System MUST support environment variable `PROTOBUF_ENABLED` with values: "all", "internal", "websocket", "none" to control which communication channels use Protobuf
- **FR-018**: Development environment MUST default `PROTOBUF_ENABLED` to "none" (JSON only) to facilitate debugging with standard tools
- **FR-019**: Production environment SHOULD default `PROTOBUF_ENABLED` to "all" after successful staging validation

### Key Entities

- **Protobuf Schema**: Formal .proto definition files that specify message structures, field types, enumerations, and nested types for all system communication
- **AgentMessage**: Core conversation unit with sender role, message content, intent type, timestamp, and optional cache_key reference
- **ToolCall**: Function invocation request containing tool_name, serialized arguments, and a unique call_id for tracking
- **ToolResult**: Function execution result with success flag, serialized result payload, and cache_key for large results stored in the cache layer
- **StateChangeEvent**: FSM transition notification carrying old_state, new_state, match_id, timestamp, and human-readable reason
- **MemoryUpdate**: Notification of a vector memory change with memory_type, content summary, and the affected user or match identifiers
- **WebSocketFrame**: Transport envelope for all WebSocket messages with a type discriminator, binary payload, monotonic sequence number, and timestamp
- **QueueTask**: RabbitMQ task envelope with task type identifier, priority level, serialized payload, and retry configuration

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Protobuf-encoded messages achieve at least 60% smaller payload size compared to equivalent JSON messages for typical agent dialogue exchanges
- **SC-002**: Tool call results returning 20 or more job listings see 90% or greater reduction in transmitted bytes when using cache_key + summary versus full payload
- **SC-003**: All existing system functionality (dual-agent negotiation, WebSocket chat, job matching, interview scheduling, offer creation, vector memory) operates identically regardless of whether JSON or Protobuf encoding is used
- **SC-004**: Migration enables zero-downtime deployment — existing clients continue receiving JSON responses without interruption while servers roll out Protobuf support
- **SC-005**: WebSocket negotiation roundtrip latency decreases measurably due to smaller message sizes and faster binary serialization compared to JSON
- **SC-006**: Each of the four migration phases (dual-protocol, internal services, WebSocket, full switch) can be enabled or disabled independently without breaking any other phase

## Assumptions

- The tool result caching layer (from the prior optimization phase) is deployed and provides reliable get/set operations with configurable TTL
- The backend is written in Go, which has mature protoc-gen-go support for generating idiomatic Go structs from .proto definitions
- The frontend uses TypeScript; ts-proto generates type-safe interfaces compatible with the existing codebase
- RabbitMQ natively supports binary message payloads without special configuration
- Communication with the vector database uses REST and can be migrated to Protobuf in phase 2
- The FSM (finite state machine) exposes event hooks that can be intercepted to broadcast notifications without modifying core FSM logic
- Agent logic is transport-agnostic — agents receive already-deserialized message objects regardless of wire format
- Schema versioning uses a single version initially; forward-compatible evolution (adding optional fields) will be addressed in a future iteration
