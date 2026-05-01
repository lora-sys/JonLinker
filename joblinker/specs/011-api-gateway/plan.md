# Plan: API Gateway Unified Layer

**Feature Branch**: `011-api-gateway`  
**Created**: 2026-05-01  
**Status**: Draft  
**Input**: User description for lightweight API Gateway unified layer

## Technical Context

### Tech Stack
- **Frontend**: Next.js 16 with TypeScript, React hooks
- **Backend**: Go (Gin framework), PostgreSQL, RabbitMQ
- **AI**: LongCat API with function calling support
- **Architecture**: A2A multi-agent with autonomous negotiation

### Project Structure
```
joblinker/
├── frontend/src/
│   ├── lib/
│   │   └── api_client.ts     # Current direct fetch calls
│   ├── hooks/
│   │   └── useWebSocket.ts   # WebSocket connections
│   └── app/                  # Next.js pages
├── backend/
│   ├── internal/
│   │   ├── handler/          # HTTP handlers
│   │   ├── service/         # Business logic
│   │   └── repository/      # Data access
│   ├── pkg/
│   │   └── ai/              # AI client
│   └── cmd/server/main.go   # Server entry
└── specs/
    └── 011-api-gateway/     # This feature
```

### Key Constraints
- **Preserve existing business logic**: No changes to handlers, services, repositories
- **Non-disruptive**: Must work with current RabbitMQ, WebSocket, vector DB, dual-agent features
- **A2A compatibility**: Each Agent's tool calls must route through gateway

### Unknowns (NEEDS CLARIFICATION)
- Route configuration format: JSON vs YAML vs TypeScript constants?
- Gateway location: frontend/lib/ vs shared/pkg/ for reuse across Agent?

## Implementation Plan

### Phase 1: Gateway Core (Infrastructure)

#### T001 [P] Create gateway directory structure
```
frontend/src/lib/gateway/
├── index.ts              # Main export
├── client.ts             # Gateway client class
├── routes.ts             # Route mapping config
├── types.ts              # TypeScript types
└── middleware.ts         # Header/tenant middleware
```

#### T002 [P] Define route mapping configuration
Create `routes.ts` with service route mappings:
```typescript
export const ServiceRoutes = {
  // Agent routes
  agents: {
    list:   { method: 'GET',    path: '/api/agents' },
    create: { method: 'POST',   path: '/api/agents' },
    get:    { method: 'GET',    path: '/api/agents/:id' },
    update: { method: 'PATCH',  path: '/api/agents/:id' },
    delete: { method: 'DELETE', path: '/api/agents/:id' },
  },
  // Job routes
  jobs: {
    list:   { method: 'GET',    path: '/api/jobs' },
    create: { method: 'POST',   path: '/api/jobs' },
    get:    { method: 'GET',    path: '/api/jobs/:id' },
    update: { method: 'PATCH',  path: '/api/jobs/:id' },
  },
  // Match routes
  matches: {
    list:      { method: 'GET',    path: '/api/matches' },
    get:       { method: 'GET',    path: '/api/matches/:id' },
    autoCreate:{ method: 'POST',   path: '/api/matches/auto' },
    confirm:   { method: 'POST',   path: '/api/matches/:id/confirm' },
  },
  // Interview routes
  interviews: {
    list:     { method: 'GET',    path: '/api/interviews' },
    create:   { method: 'POST',   path: '/api/interviews' },
    update:   { method: 'PATCH',   path: '/api/interviews/:id' },
    getByMatch:{ method: 'GET',   path: '/api/interviews/:matchId' },
    confirm:  { method: 'POST',   path: '/api/interviews/:matchId/confirm' },
    cancel:   { method: 'POST',   path: '/api/interviews/:matchId/cancel' },
  },
  // Offer routes
  offers: {
    getByMatch:  { method: 'GET',    path: '/api/offers/:matchId' },
    create:      { method: 'POST',   path: '/api/offers' },
    accept:      { method: 'POST',   path: '/api/offers/:matchId/accept' },
    decline:     { method: 'POST',   path: '/api/offers/:matchId/decline' },
  },
  // Message routes
  messages: {
    list:         { method: 'GET',    path: '/api/messages/:matchId' },
    send:         { method: 'POST',   path: '/api/messages/:matchId' },
    websocket:    { method: 'WS',    path: '/api/messages/:matchId/ws' },
  },
  // Privacy routes
  privacy: {
    export:  { method: 'POST',   path: '/api/privacy/export' },
    delete:  { method: 'DELETE', path: '/api/privacy/account' },
  },
} as const;
```

#### T003 [P] Implement GatewayClient class
```typescript
export class GatewayClient {
  private baseURL: string;
  private headers: Record<string, string>;

  constructor(config: GatewayConfig) {
    this.baseURL = config.baseURL;
    this.headers = {
      'Content-Type': 'application/json',
      'X-User-ID': config.userId,
      'X-Agent-ID': config.agentId,
      'X-Tenant-ID': config.tenantId || 'default',
    };
  }

  async request<R = any>(route: RouteDef, params?: Record<string, any>, data?: any): Promise<ApiResponse<R>> {
    // Build URL with path parameters
    const url = this.buildURL(route.path, params);
    // Execute request with headers
    // Handle response and errors
    // Support WebSocket upgrade
  }
}
```

#### T004 Create gateway middleware for headers
- Header injection middleware
- Request logging hook
- Tenant context propagation

### Phase 2: Global Replacement

#### T005 [P] Replace frontend API calls
Files to modify:
- `frontend/src/lib/api_client.ts` → use GatewayClient
- `frontend/src/stores/auth.ts` → use GatewayClient
- `frontend/src/app/agents/page.tsx`
- `frontend/src/app/jobs/page.tsx`
- `frontend/src/app/matches/page.tsx`
- `frontend/src/app/interviews/page.tsx`
- `frontend/src/app/offers/page.tsx`
- `frontend/src/app/messages/page.tsx`
- `frontend/src/app/settings/page.tsx`

#### T006 [P] Replace WebSocket hook
- `frontend/src/hooks/useWebSocket.ts` → route through gateway WebSocket

#### T007 [P] Replace AI Agent tool calls
Backend files (Agent uses gateway for tool execution):
- `backend/internal/agent/tool_executor.go` → use GatewayClient for internal calls
- `backend/internal/service/dual_agent_negotiation_service.go`

### Phase 3: Extension Points

#### T008 Add rate limiting hook
```typescript
interface RateLimitConfig {
  maxRequests: number;
  windowMs: number;
}
```

#### T009 Add circuit breaker
```typescript
interface CircuitBreakerConfig {
  failureThreshold: number;
  resetTimeoutMs: number;
}
```

#### T010 Add request logging
```typescript
interface LoggingConfig {
  logRequests: boolean;
  logResponses: boolean;
  logHeaders: boolean;
}
```

### Phase 4: Verification

#### T011 Verify all existing features work
- [ ] RabbitMQ message queue
- [ ] WebSocket connections
- [ ] Vector DB memory
- [ ] Dual-agent dialogue
- [ ] Preference vector storage

## Data Model

### GatewayConfig
```typescript
interface GatewayConfig {
  baseURL: string;       // e.g., 'http://localhost:8080'
  userId: string;         // Current user ID
  agentId: string;        // Current agent ID (optional)
  tenantId: string;       // Tenant for multi-tenant isolation
}
```

### Route Definition
```typescript
interface RouteDef {
  method: 'GET' | 'POST' | 'PATCH' | 'DELETE' | 'WS';
  path: string;          // Path with :param placeholders
}
```

### API Response
```typescript
interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}
```

## Verification Checklist

- [ ] GatewayClient handles all HTTP methods
- [ ] WebSocket upgrade works through gateway
- [ ] Headers (X-User-ID, X-Agent-ID, X-Tenant-ID) injected on all requests
- [ ] Route parameters substituted correctly
- [ ] Frontend pages work after replacement
- [ ] Agent tool calls route through gateway
- [ ] RabbitMQ consumer still works
- [ ] No breaking changes to existing business logic
