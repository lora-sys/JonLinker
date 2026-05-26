// TypeScript types for the gateway layer

export interface GatewayConfig {
  baseURL: string
  userId: string
  agentId?: string
  tenantId?: string
}

export interface ApiResponse<T = unknown> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
}

export interface RequestLog {
  timestamp: Date
  method: string
  path: string
  userId: string
  agentId?: string
  status: number
  durationMs: number
}

// Rate limiting config
export interface RateLimitConfig {
  maxRequests: number
  windowMs: number
}

// Circuit breaker config
export interface CircuitBreakerConfig {
  failureThreshold: number
  resetTimeoutMs: number
}

// Request interceptor hook
export type RequestInterceptor = (config: RequestConfig) => RequestConfig | Promise<RequestConfig>

// Response interceptor hook
export type ResponseInterceptor = <T>(response: ApiResponse<T>) => ApiResponse<T> | Promise<ApiResponse<T>>

// Request configuration built by GatewayClient
export interface RequestConfig {
  method: string
  url: string
  headers: Record<string, string>
  body?: string | undefined
}

// Gateway client interface
export interface IGatewayClient {
  request: <R = unknown>(
    routeDef: { method: string, path: string },
    params?: Record<string, string | number | boolean | undefined>,
    data?: unknown,
  ) => Promise<ApiResponse<R>>

  connectWebSocket: (
    routeDef: { method: string, path: string },
    params?: Record<string, string>,
  ) => WebSocket
}
