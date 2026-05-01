# Data Model: API Gateway

## Entities

### GatewayConfig
Unified gateway configuration for all API clients.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| baseURL | string | Yes | Backend server base URL |
| userId | string | Yes | Current authenticated user ID |
| agentId | string | No | Active agent ID for A2A operations |
| tenantId | string | No | Tenant ID for multi-tenant isolation (default: 'default') |

### RouteDef
Route definition for service endpoint mapping.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| method | 'GET' \| 'POST' \| 'PATCH' \| 'DELETE' \| 'WS' | Yes | HTTP method or WebSocket |
| path | string | Yes | Path with :param placeholders |

### ApiResponse
Standard API response wrapper.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| success | boolean | Yes | Whether request succeeded |
| data | T | No | Response payload |
| error | ErrorDetail | No | Error information |

### ErrorDetail
Error details for failed requests.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| code | string | Yes | Error code for programmatic handling |
| message | string | Yes | Human-readable error message |

### RateLimitConfig
Rate limiting configuration for gateway extensibility.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| maxRequests | number | Yes | Max requests per window |
| windowMs | number | Yes | Time window in milliseconds |

### CircuitBreakerConfig
Circuit breaker configuration for fault tolerance.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| failureThreshold | number | Yes | Failures before opening circuit |
| resetTimeoutMs | number | Yes | Time before attempting reset |

### RequestLog
Request/response logging entry.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| timestamp | Date | Yes | Request timestamp |
| method | string | Yes | HTTP method |
| path | string | Yes | Request path |
| userId | string | Yes | User ID from header |
| agentId | string | No | Agent ID from header |
| status | number | Yes | Response status code |
| durationMs | number | Yes | Request duration |

## Relationships

```
GatewayClient (Frontend/Agent)
  │
  ├── uses ──> RouteService (route mapping)
  │
  ├── sends ─> HTTP Request (with headers)
  │
  └── receives ─> ApiResponse<T>
```
