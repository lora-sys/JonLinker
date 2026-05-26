// Route mapping configuration for all backend service endpoints
// This is the single source of truth for API routes

export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE' | 'WS'

export interface RouteDef {
  method: HttpMethod
  path: string
}

export interface RouteGroup {
  [key: string]: RouteDef
}

// Service routes - all backend API endpoints
export const ServiceRoutes = {
  // Agent management
  agents: {
    list: { method: 'GET' as const, path: '/api/agents' },
    create: { method: 'POST' as const, path: '/api/agents' },
    get: { method: 'GET' as const, path: '/api/agents/:id' },
    update: { method: 'PATCH' as const, path: '/api/agents/:id' },
    delete: { method: 'DELETE' as const, path: '/api/agents/:id' },
  },

  // Job management
  jobs: {
    list: { method: 'GET' as const, path: '/api/jobs' },
    create: { method: 'POST' as const, path: '/api/jobs' },
    get: { method: 'GET' as const, path: '/api/jobs/:id' },
    update: { method: 'PATCH' as const, path: '/api/jobs/:id' },
  },

  // Match management
  matches: {
    list: { method: 'GET' as const, path: '/api/matches' },
    get: { method: 'GET' as const, path: '/api/matches/:id' },
    autoCreate: { method: 'POST' as const, path: '/api/matches/auto' },
    confirm: { method: 'POST' as const, path: '/api/matches/:id/confirm' },
  },

  // Interview management
  interviews: {
    list: { method: 'GET' as const, path: '/api/interviews' },
    create: { method: 'POST' as const, path: '/api/interviews' },
    update: { method: 'PATCH' as const, path: '/api/interviews/:id' },
    getByMatch: { method: 'GET' as const, path: '/api/interviews/:matchId' },
    confirm: { method: 'POST' as const, path: '/api/interviews/:matchId/confirm' },
    cancel: { method: 'POST' as const, path: '/api/interviews/:matchId/cancel' },
  },

  // Offer management
  offers: {
    getByMatch: { method: 'GET' as const, path: '/api/offers/:matchId' },
    create: { method: 'POST' as const, path: '/api/offers' },
    accept: { method: 'POST' as const, path: '/api/offers/:matchId/accept' },
    decline: { method: 'POST' as const, path: '/api/offers/:matchId/decline' },
  },

  // Message routes
  messages: {
    list: { method: 'GET' as const, path: '/api/messages/:matchId' },
    send: { method: 'POST' as const, path: '/api/messages/:matchId' },
    websocket: { method: 'WS' as const, path: '/api/messages/:matchId/ws' },
  },

  // Privacy routes
  privacy: {
    export: { method: 'POST' as const, path: '/api/privacy/export' },
    delete: { method: 'DELETE' as const, path: '/api/privacy/account' },
  },

  // Auth routes
  auth: {
    register: { method: 'POST' as const, path: '/api/auth/register' },
    login: { method: 'POST' as const, path: '/api/auth/login' },
    refresh: { method: 'POST' as const, path: '/api/auth/refresh' },
  },

  // Health check
  health: { method: 'GET' as const, path: '/health' },
} as const

export type ServiceRouteKey = keyof typeof ServiceRoutes
