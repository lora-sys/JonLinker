# Research: Protobuf Communication Protocol

**Feature**: 014-protobuf-protocol
**Date**: 2026-05-02

## Research Tasks

### 1. Protobuf Schema Organization

**Decision**: Top-level `proto/` directory with 4 focused files: `agent.proto`, `tools.proto`, `websocket.proto`, `queue.proto`.

**Rationale**: Separating concerns by domain matches existing module boundaries. Agent messages, tool data structures, WebSocket frames, and queue tasks each have distinct versioning and evolution paths.

**Alternatives considered**:
- Single monolithic `joblinker.proto` — harder to manage, forces coordinated version bumps.
- Per-service directories (e.g., `proto/agent/`, `proto/queue/`) — overkill for current scope.

### 2. Code Generation Approach

**Decision**: Use `protoc` with `protoc-gen-go` (Go) and `ts-proto` (TypeScript). Generate into `backend/pkg/proto/` and `frontend/src/lib/proto/`.

**Rationale**:
- Go: `protoc-gen-go` is standard, produces idiomatic structs with proper getters/setters.
- TypeScript: `ts-proto` produces minimal, readable output with no runtime dependency, supports `BigInt` for uint64 fields.
- Generated files are checked into version control (not `.gitignore`'d) to avoid requiring protoc for every build.

**Alternatives considered**:
- `protobuf-es` for TypeScript — larger runtime but better for browsers; ts-proto is simpler.
- Not checking in generated code — would require protoc in CI and for every developer; adds friction.

### 3. Content Negotiation for Dual Protocol

**Decision**: Use `Accept` request header for responses; `Content-Type` request header for requests. Default to JSON when header missing or `Accept: application/json`.

**Rationale**: Standard HTTP content negotiation. No custom headers needed. Works with existing REST clients.

**Implementation pattern**:

**Alternatives considered**:
- Custom `X-Protocol` header  non-standard, more work for clients.
- Query parameter `?format=protobuf` — works but less idiomatic for content negotiation.

### 4. WebSocket Protocol Detection

**Decision**: Detect by peeking first byte. Protobuf messages always start with a varint tag (bits 7-0 of field 1, wire type). Byte values 0x08–0x0F (field 1, varint) or 0x0A (field 1, length-delimited) indicate Protobuf. Fallback to JSON parsing if detection fails.

**Rationale**: Protobuf wire format has deterministic first-byte patterns. No side channel needed.

**Implementation**:

### 5. Cache Key Integration

**Decision**: Tool execution results already cached via `tool_cache.go` (SHA256 key + summary). In Protobuf mode, `WebSocketFrame.cache_key` field carries the key. The full payload is omitted; summary is included in `WebSocketFrame.payload` as a text field.

**Rationale**: Leverages existing cache infrastructure. No changes to cache storage; only message structure.

**Flow**:
1. Agent makes tool call → executor runs → result stored with key.
2. System generates summary (existing method in `tool_cache.go` summaries map).
3. WebSocket frame sent with `cache_key` and `payload` (summary).
4. Receiving agent sees `cache_key` → retrieves full result from cache if needed.

### 6. Environment Configuration

**Decision**: `PROTOBUF_ENABLED` enum with values "none", "internal", "websocket", "all". Default "none" for dev, "all" for prod.

**Rationale**: Granular control per communication channel. Enables staged rollout without code changes.

**Implementation**:

### 7. RabbitMQ Binary Payloads

**Decision**: RabbitMQ supports binary payloads natively. Set `Content-Type: application/x-protobuf` in message properties.

**Rationale**: No RabbitMQ configuration changes needed. Consumers detect by `Content-Type` header.

**Implementation**:

## Unresolved Items

None. All technical decisions made.

## References

- [protobuf documentation](https://protobuf.dev/)
- [ts-proto](https://github.com/stephenh/ts-proto)
- [protoc-gen-go](https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go)
- Existing cache implementation: `backend/internal/cache/tool_cache.go`
- Existing tool definitions: `backend/internal/agent/function_definitions.go`