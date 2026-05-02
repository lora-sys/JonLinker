# Data Model: Protobuf Communication Protocol

**Feature**: 014-protobuf-protocol
**Date**: 2026-05-02

## Domain Entities (as Protobuf Messages)

### agent.proto

**AgentRole** (enum):
- SEEKER = 0
- RECRUITER = 1

**AgentState** (enum):
- IDLE = 0
- SEARCHING = 1
- NEGOTIATING = 2
- INTERVIEWING = 3
- OFFER_RECEIVED = 4
- HIRED = 5
- REJECTED = 6

**TextMessage**:
- `id` (string, required) - unique identifier
- `role` (AgentRole, required) - sender role
- `content` (string, required) - message text
- `intent_type` (string, optional) - e.g., "negotiation", "offer", "schedule"
- `timestamp` (int64, required) - Unix timestamp in milliseconds

**ToolCall**:
- `call_id` (string, required) - unique identifier for tracking
- `tool_name` (string, required) - e.g., "query_jobs", "create_offer"
- `arguments` (bytes, required) - JSON or structured arguments
- `cache_key` (string, optional) - reference to cached result

**ToolResult**:
- `call_id` (string, required) - matches ToolCall.call_id
- `success` (bool, required)
- `payload` (bytes, optional) - result data (may be omitted if cache_key provided)
- `cache_key` (string, optional) - reference to cached full result
- `summary` (string, optional) - human-readable summary
- `error_message` (string, optional) - if success=false

**StateChangeEvent**:
- `match_id` (string, required) - UUID of affected match
- `old_state` (AgentState, required)
- `new_state` (AgentState, required)
- `reason` (string, optional) - human-readable reason
- `timestamp` (int64, required)

**MemoryUpdate**:
- `memory_id` (string, required) - unique identifier
- `memory_type` (string, required) - e.g., "preference", "fact"
- `user_id` (string, required)
- `content` (string, required) - summary or full content
- `embedding` (repeated double, optional) - vector embedding
- `timestamp` (int64, required)

### tools.proto

**JobDetail**:
- `job_id` (string, required)
- `title` (string, required)
- `description` (string, required)
- `salary_min` (int64, optional)
- `salary_max` (int64, optional)
- `currency` (string, required) - e.g., "USD", "CNY"
- `location` (string, optional)
- `skills` (repeated string, optional)
- `experience_years_min` (int32, optional)
- `experience_years_max` (int32, optional)
- `posted_at` (int64, required) - timestamp
- `status` (string, required) - e.g., "open", "closed", "filled"

**ResumeDetail**:
- `resume_id` (string, required)
- `candidate_name` (string, required)
- `email` (string, optional)
- `phone` (string, optional)
- `skills` (repeated string, optional)
- `experience_years` (int32, optional)
- `salary_expectation_min` (int64, optional)
- `salary_expectation_max` (int64, optional)
- `location_preference` (string, optional)
- `summary` (string, optional)
- `uploaded_at` (int64, required)

**InterviewInvitation**:
- `interview_id` (string, required)
- `match_id` (string, required)
- `scheduled_at` (int64, required) - timestamp
- `duration_minutes` (int32, required)
- `format` (enum) - options: VIDEO, PHONE, ONSITE
- `location_or_url` (string, optional)
- `status` (enum) - options: PENDING, CONFIRMED, CANCELLED, COMPLETED
- `notes` (string, optional)

**OfferDetail**:
- `offer_id` (string, required)
- `match_id` (string, required)
- `salary_amount` (int64, required)
- `currency` (string, required)
- `start_date` (int64, required) - timestamp
- `benefits` (repeated string, optional)
- `expires_at` (int64, required) - timestamp
- `status` (enum) - options: PENDING_ACCEPTANCE, ACCEPTED, DECLINED, EXPIRED
- `terms` (string, optional) - additional terms as text

### websocket.proto

**MessageType** (enum):
- TEXT = 0
- TOOL_CALL = 1
- TOOL_RESULT = 2
- STATE_CHANGE = 3
- MEMORY_UPDATE = 4
- HEARTBEAT = 5
- ERROR = 6

**WebSocketFrame**:
- `message_type` (MessageType, required)
- `payload` (bytes, required) - serialized message (TextMessage, ToolCall, etc.)
- `sequence_num` (uint64, required) - monotonic increasing for ordering
- `timestamp` (int64, required) - Unix timestamp in milliseconds
- `cache_key` (string, optional) - for tool results using cache pattern
- `correlation_id` (string, optional) - for request-response matching

### queue.proto

**TaskPriority** (enum):
- LOW = 0
- NORMAL = 1
- HIGH = 2
- CRITICAL = 3

**QueueTask**:
- `task_id` (string, required)
- `task_type` (string, required) - e.g., "job.match", "resume.parse"
- `priority` (TaskPriority, required)
- `payload` (bytes, required) - serialized task data
- `retry_count` (uint32, optional) - current retry attempt
- `max_retries` (uint32, optional) - default 3
- `created_at` (int64, required)
- `scheduled_at` (int64, optional) - for delayed execution

## Validation Rules

From spec.md requirements:

- **FR-002**: AgentRole, AgentState enums must contain all specified values
- **FR-003**: ToolDetail messages must match existing function calling schemas (see `function_definitions.go`)
- **FR-004**: WebSocketFrame must include sequence_num (uint64) and timestamp (int64)
- **FR-005**: QueueTask must support task_type string and priority enum with specified values

## Versioning

All top-level messages should include a `schema_version` field (uint32, default 1) for future evolution. Added in v2 if needed; not included in v1 for minimal overhead.

## Relationships

- `WebSocketFrame.payload` contains one of: TextMessage, ToolCall, ToolResult, StateChangeEvent, MemoryUpdate
- `ToolCall.call_id` matches `ToolResult.call_id` for request tracking
- `ToolResult.cache_key` references entry in existing `tool_cache` (in-memory LRU)
- `QueueTask.payload` contains serialized task-specific message (e.g., JobDetail for matching task)
