# API Contracts: Route Mapping Configuration

## Overview

This document defines the route mapping configuration used by the GatewayClient to route API requests to backend services. All routes are relative to the base URL configured in GatewayClient.

## Route Groups

### Agent Routes (`agents`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `agents.list` | GET | `/api/agents` | List all agents for user |
| `agents.create` | POST | `/api/agents` | Create new agent |
| `agents.get` | GET | `/api/agents/:id` | Get agent by ID |
| `agents.update` | PATCH | `/api/agents/:id` | Update agent |
| `agents.delete` | DELETE | `/api/agents/:id` | Delete agent |

### Job Routes (`jobs`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `jobs.list` | GET | `/api/jobs` | List all jobs |
| `jobs.create` | POST | `/api/jobs` | Create new job |
| `jobs.get` | GET | `/api/jobs/:id` | Get job by ID |
| `jobs.update` | PATCH | `/api/jobs/:id` | Update job |

### Match Routes (`matches`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `matches.list` | GET | `/api/matches` | List all matches |
| `matches.get` | GET | `/api/matches/:id` | Get match by ID |
| `matches.autoCreate` | POST | `/api/matches/auto` | Auto-create match |
| `matches.confirm` | POST | `/api/matches/:id/confirm` | Confirm match |

### Interview Routes (`interviews`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `interviews.list` | GET | `/api/interviews` | List all interviews |
| `interviews.create` | POST | `/api/interviews` | Create interview |
| `interviews.update` | PATCH | `/api/interviews/:id` | Update interview |
| `interviews.getByMatch` | GET | `/api/interviews/:matchId` | Get interview by match ID |
| `interviews.confirm` | POST | `/api/interviews/:matchId/confirm` | Confirm interview |
| `interviews.cancel` | POST | `/api/interviews/:matchId/cancel` | Cancel interview |

### Offer Routes (`offers`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `offers.getByMatch` | GET | `/api/offers/:matchId` | Get offer by match ID |
| `offers.create` | POST | `/api/offers` | Create offer |
| `offers.accept` | POST | `/api/offers/:matchId/accept` | Accept offer |
| `offers.decline` | POST | `/api/offers/:matchId/decline` | Decline offer |

### Message Routes (`messages`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `messages.list` | GET | `/api/messages/:matchId` | Get message history |
| `messages.send` | POST | `/api/messages/:matchId` | Send message |
| `messages.websocket` | WS | `/api/messages/:matchId/ws` | WebSocket connection |

### Privacy Routes (`privacy`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `privacy.export` | POST | `/api/privacy/export` | Export user data |
| `privacy.delete` | DELETE | `/api/privacy/account` | Delete account |

### Auth Routes (`auth`)

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `auth.register` | POST | `/api/auth/register` | Register user |
| `auth.login` | POST | `/api/auth/login` | Login user |
| `auth.refresh` | POST | `/api/auth/refresh` | Refresh token |

### Health Routes

| Key | Method | Path | Description |
|-----|--------|------|-------------|
| `health` | GET | `/health` | Health check |

## TypeScript Type Definitions

```typescript
export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE' | 'WS';

export interface RouteDef {
  method: HttpMethod;
  path: string;
}

export interface RouteGroup {
  [key: string]: RouteDef;
}

export interface ServiceRoutes {
  agents: RouteGroup;
  jobs: RouteGroup;
  matches: RouteGroup;
  interviews: RouteGroup;
  offers: RouteGroup;
  messages: RouteGroup;
  privacy: RouteGroup;
  auth: RouteGroup;
  health: RouteGroup;
}
```

## Example Usage

```typescript
import { GatewayClient, ServiceRoutes } from '@/lib/gateway';

// Configure gateway once at app init
const gateway = new GatewayClient({
  baseURL: 'http://localhost:8080',
  userId: 'user-123',
  agentId: 'agent-456',
  tenantId: 'tenant-789',
});

// Call API using route key
const agents = await gateway.request(ServiceRoutes.agents.list);

// Call with path parameters
const agent = await gateway.request(ServiceRoutes.agents.get, { id: 'agent-456' });

// Create something
await gateway.request(ServiceRoutes.jobs.create, {}, {
  title: 'Software Engineer',
  location: 'Remote',
});

// WebSocket connection
gateway.connectWebSocket(ServiceRoutes.messages.websocket, { matchId: 'match-123' });
```

## Request Headers

All requests through the gateway include:

| Header | Source | Description |
|--------|--------|-------------|
| `Content-Type` | Always `application/json` | Request content type |
| `X-User-ID` | GatewayClient.config.userId | Current user identifier |
| `X-Agent-ID` | GatewayClient.config.agentId | Current agent identifier |
| `X-Tenant-ID` | GatewayClient.config.tenantId | Tenant for multi-tenant isolation |

## WebSocket Special Case

WebSocket connections cannot include headers during the handshake. Therefore:

- Token is passed as query parameter `?token=<jwt_token>`
- Same GatewayClient config applies but headers become query params
