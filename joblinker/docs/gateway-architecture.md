# API Gateway Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (Next.js)                       │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐             │
│  │ Agents  │  │  Jobs   │  │ Matches │  │Messages │             │
│  │  Page   │  │  Page   │  │  Page   │  │  Page   │             │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘             │
│       └──────────┬┴───────────┬┴───────────┬┘                   │
│                  ▼            ▼            ▼                     │
│            ┌─────────────────────────────────┐                  │
│            │      GatewayClient (index.ts)     │                  │
│            │  - Route mapping                  │                  │
│            │  - Header injection               │                  │
│            │  - Rate limiting                 │                  │
│            │  - Circuit breaker               │                  │
│            └──────────────┬──────────────────┘                  │
└──────────────────────────┼──────────────────────────────────────┘
                           │ HTTP/WS
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Backend (Go + Gin)                           │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    JWT Auth Middleware                   │    │
│  │              X-User-ID, X-Agent-ID, X-Tenant-ID         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ▼                    ▼                    ▼             │
│  ┌────────────┐      ┌────────────┐      ┌────────────┐         │
│  │  Handlers  │      │  Services  │      │ Repositories│        │
│  │  /api/*    │      │   Logic    │      │    Data     │        │
│  └────────────┘      └────────────┘      └────────────┘         │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ▼                    ▼                    ▼             │
│  ┌────────────┐      ┌────────────┐      ┌────────────┐         │
│  │ PostgreSQL │      │  RabbitMQ  │      │   AI API    │         │
│  │ (pgvector) │      │  (Queue)   │      │ (LongCat)   │         │
│  └────────────┘      └────────────┘      └────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

## Gateway Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Gateway Client                              │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    ServiceRoutes                         │    │
│  │  agents.* │ jobs.* │ matches.* │ interviews.* │ offers.* │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ▼                    ▼                    ▼             │
│  ┌────────────┐      ┌────────────┐      ┌────────────┐         │
│  │ RateLimiter│      │CircuitBreaker│    │  Headers   │         │
│  │ (100/min)  │      │ (5 failures) │   │ X-User-ID  │         │
│  └────────────┘      └────────────┘      │ X-Agent-ID│         │
│                                           │ X-Tenant-ID│         │
│                                           └────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
                    ┌──────────────────┐
                    │   fetch() HTTP   │
                    │  or WebSocket    │
                    └──────────────────┘
```

## Multi-Agent A2A Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    A2A Autonomous Dialogue                       │
│                                                                  │
│  ┌─────────────────┐           ┌─────────────────┐             │
│  │   Seeker Agent   │◄─────────►│  Recruiter Agent│             │
│  │   (Job Seeker)   │   A2A     │   (Company)     │             │
│  └────────┬────────┘  XML       └────────┬────────┘             │
│           │       Protocol              │                        │
│           │                              │                        │
│           ▼                              ▼                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    MessageQueueService                    │    │
│  │  - AI Response Generation (LongCat API)                  │    │
│  │  - Tool Execution (schedule_interview, create_offer)     │    │
│  │  - Memory Storage (pgvector embeddings)                 │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                   │
│           ┌──────────────────┼──────────────────┐               │
│           ▼                  ▼                  ▼                │
│  ┌────────────┐      ┌────────────┐      ┌────────────┐         │
│  │ PostgreSQL │      │  RabbitMQ  │      │   AI API   │         │
│  │ + pgvector │      │            │      │ (LongCat)  │         │
│  └────────────┘      └────────────┘      └────────────┘         │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                  Human Confirmation Only                 │    │
│  │     (Agent receives final offer/schedule, human confirms)│    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow (Gateway Headers)

```
Frontend                     Gateway                    Backend
   │                           │                          │
   │  ┌──────────────────┐     │                          │
   │  │ apiClient.post() │     │                          │
   │  │ ServiceRoutes    │     │                          │
   │  │ .offers.create   │     │                          │
   │  └────────┬─────────┘     │                          │
   │           │              │                          │
   │           ▼              │                          │
   │  ┌──────────────────┐     │                          │
   │  │ X-User-ID: uid   │     │                          │
   │  │ X-Agent-ID: aid  │────►│  HTTP Request              │
   │  │ X-Tenant-ID: tid │     │  + Headers                │
   │  │ Authorization    │     │  + Body                   │
   │  └──────────────────┘     │                          │
   │                           │                          │
   │                           ▼                          │
   │                    ┌──────────────┐                  │
   │                    │ JWT Validate │                  │
   │                    │ Extract user │                  │
   │                    └──────────────┘                  │
   │                           │                          │
   │                           ▼                          │
   │                    ┌──────────────┐                  │
   │                    │   Handler    │                  │
   │                    │   Service    │                  │
   │                    │   Repository │                  │
   │                    └──────────────┘                  │
```

## Directory Structure

```
frontend/src/lib/gateway/
├── index.ts        # Main export (GatewayClient, ServiceRoutes)
├── client.ts       # GatewayClient class
│   ├── request()   # Route-based HTTP calls
│   ├── doRequest() # Low-level HTTP execution
│   ├── get/post/patch/delete()  # Convenience methods
│   └── connectWebSocket() # WebSocket with token query
├── routes.ts      # Service route mapping (agents, jobs, matches, etc.)
├── types.ts       # TypeScript interfaces (GatewayConfig, ApiResponse)
└── middleware.ts  # RateLimiter, CircuitBreaker, TenantContext
```

## Route Mapping

```typescript
ServiceRoutes.agents.list    // GET  /api/agents
ServiceRoutes.agents.create  // POST /api/agents
ServiceRoutes.jobs.get       // GET  /api/jobs/:id
ServiceRoutes.offers.accept  // POST /api/offers/:matchId/accept
ServiceRoutes.messages.ws    // WS   /api/messages/:matchId/ws
```

## Extension Points

| Feature | Implementation | Config |
|---------|--------------|--------|
| Rate Limiting | `RateLimiter` class | `maxRequests: 100, windowMs: 60000` |
| Circuit Breaker | `CircuitBreaker` class | `failureThreshold: 5, resetTimeoutMs: 30000` |
| Request Logging | `createRequestLogger()` | `onLog: (log) => console.log(log)` |
| Tenant Isolation | Header injection | `X-Tenant-ID: <tenant>` |