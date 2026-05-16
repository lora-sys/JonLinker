// Gateway middleware for request/response processing

import type { RequestInterceptor, RequestLog, ResponseInterceptor } from './types'

// Request logging
export function createRequestLogger(onLog?: (log: RequestLog) => void): RequestInterceptor {
  return async (config) => {
    const startTime = Date.now();

    // Log on response completion via wrapper
    (config as any)._startTime = startTime;
    (config as any)._onLog = onLog

    return config
  }
}

// Response logging (wraps response)
export function createResponseLogger(): ResponseInterceptor {
  return async (response) => {
    // Response logging handled by request wrapper
    return response
  }
}

// Rate limiter implementation
export class RateLimiter {
  private requests: number[] = []
  private maxRequests: number
  private windowMs: number

  constructor(config: { maxRequests: number, windowMs: number }) {
    this.maxRequests = config.maxRequests
    this.windowMs = config.windowMs
  }

  async check(): Promise<{ allowed: boolean, remaining: number, resetAt: Date }> {
    const now = Date.now()
    const windowStart = now - this.windowMs

    // Remove old requests outside window
    this.requests = this.requests.filter(ts => ts > windowStart)

    if (this.requests.length >= this.maxRequests) {
      const resetAt = new Date((this.requests[0] ?? Date.now()) + this.windowMs)
      return { allowed: false, remaining: 0, resetAt }
    }

    this.requests.push(now)
    return {
      allowed: true,
      remaining: this.maxRequests - this.requests.length,
      resetAt: new Date(now + this.windowMs),
    }
  }
}

// Circuit breaker implementation
export class CircuitBreaker {
  private failures = 0
  private lastFailureTime = 0
  private state: 'closed' | 'open' | 'half-open' = 'closed'
  private failureThreshold: number
  private resetTimeoutMs: number

  constructor(config: { failureThreshold: number, resetTimeoutMs: number }) {
    this.failureThreshold = config.failureThreshold
    this.resetTimeoutMs = config.resetTimeoutMs
  }

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === 'open') {
      const timeSinceLastFailure = Date.now() - this.lastFailureTime
      if (timeSinceLastFailure >= this.resetTimeoutMs) {
        this.state = 'half-open'
      }
      else {
        throw new Error('Circuit breaker is open')
      }
    }

    try {
      const result = await fn()
      this.onSuccess()
      return result
    }
    catch (error) {
      this.onFailure()
      throw error
    }
  }

  private onSuccess() {
    this.failures = 0
    this.state = 'closed'
  }

  private onFailure() {
    this.failures++
    this.lastFailureTime = Date.now()
    if (this.failures >= this.failureThreshold) {
      this.state = 'open'
    }
  }

  getState() {
    return this.state
  }
}

// Tenant context for multi-tenant isolation
export interface TenantContext {
  userId: string
  agentId?: string
  tenantId: string
}

export function getTenantContext(config: {
  userId: string
  agentId?: string
  tenantId?: string
}): TenantContext {
  return {
    userId: config.userId,
    agentId: config.agentId,
    tenantId: config.tenantId || 'default',
  }
}
