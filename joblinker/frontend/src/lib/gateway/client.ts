// Unified Gateway Client - sole entry point for all API calls
// Used by: Frontend, AI Agents, MCP tools, Function Calling

import type { GatewayConfig, ApiResponse, RequestConfig } from './types';
import { ServiceRoutes, type HttpMethod } from './routes';
import { RateLimiter, CircuitBreaker, getTenantContext } from './middleware';

export { ServiceRoutes } from './routes';
export type { GatewayConfig, ApiResponse } from './types';
export type { HttpMethod } from './routes';

export interface GatewayClientOptions {
  config: GatewayConfig;
  rateLimit?: { maxRequests: number; windowMs: number };
  circuitBreaker?: { failureThreshold: number; resetTimeoutMs: number };
}

export class GatewayClient {
  private baseURL: string;
  private _headers: Record<string, string>;
  private rateLimiter?: RateLimiter;
  private circuitBreaker?: CircuitBreaker;
  private token: string | null = null;

  constructor(options: GatewayClientOptions) {
    this.baseURL = options.config.baseURL;
    this._headers = this.buildHeaders(options.config);

    if (options.rateLimit) {
      this.rateLimiter = new RateLimiter(options.rateLimit);
    }

    if (options.circuitBreaker) {
      this.circuitBreaker = new CircuitBreaker(options.circuitBreaker);
    }
  }

  private buildHeaders(config: GatewayConfig): Record<string, string> {
    const tenantCtx = getTenantContext(config);
    return {
      'Content-Type': 'application/json',
      'X-User-ID': tenantCtx.userId,
      'X-Agent-ID': tenantCtx.agentId || '',
      'X-Tenant-ID': tenantCtx.tenantId,
    };
  }

  setToken(token: string | null) {
    this.token = token;
  }

  setUserContext(userId: string, agentId?: string) {
    this._headers['X-User-ID'] = userId;
    if (agentId) {
      this._headers['X-Agent-ID'] = agentId;
    }
  }

  setAgentId(agentId: string) {
    this._headers['X-Agent-ID'] = agentId;
  }

  // Direct header access for ApiClient wrapper
  get headers(): Record<string, string> {
    return this._headers;
  }

  // Low-level HTTP request
  private async doRequest<R = unknown>(
    method: string,
    path: string,
    data?: unknown
  ): Promise<ApiResponse<R>> {
    // Rate limiting check
    if (this.rateLimiter) {
      const rateCheck = await this.rateLimiter.check();
      if (!rateCheck.allowed) {
        return {
          success: false,
          error: {
            code: 'RATE_LIMITED',
            message: `Rate limit exceeded. Retry after ${rateCheck.resetAt.toISOString()}`,
          },
        };
      }
    }

    // Build request config
    const url = `${this.baseURL}${path}`;
    const config: RequestConfig = {
      method: method as HttpMethod,
      url,
      headers: { ...this._headers },
      body: data ? JSON.stringify(data) : undefined,
    };

    // Add auth header if token exists
    if (this.token) {
      config.headers['Authorization'] = `Bearer ${this.token}`;
    }

    // Execute request (with circuit breaker if enabled)
    const executeRequest = async (): Promise<ApiResponse<R>> => {
      try {
        const response = await fetch(config.url, {
          method: config.method,
          headers: config.headers,
          body: config.body,
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({
            error: `HTTP ${response.status}`,
          }));

          return {
            success: false,
            error: {
              code: `HTTP_${response.status}`,
              message: errorData.error || `Request failed with status ${response.status}`,
            },
          };
        }

        const responseData = await response.json();
        return { success: true, data: responseData };
      } catch (error) {
        return {
          success: false,
          error: {
            code: 'NETWORK_ERROR',
            message: error instanceof Error ? error.message : 'Network request failed',
          },
        };
      }
    };

    if (this.circuitBreaker) {
      return this.circuitBreaker.execute(executeRequest);
    }

    return executeRequest();
  }

  // Main request method - used for all HTTP methods
  async request<R = unknown>(
    routeDef: { method: HttpMethod; path: string },
    params?: Record<string, string | number | boolean | undefined>,
    data?: unknown
  ): Promise<ApiResponse<R>> {
    // Substitute path parameters
    let path = routeDef.path;
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          path = path.replace(`:${key}`, String(value));
        }
      });
    }

    return this.doRequest<R>(routeDef.method, path, data);
  }

  // Convenience methods for each HTTP verb
  async get<R = unknown>(
    routeDef: { method: HttpMethod; path: string },
    params?: Record<string, string | number | boolean | undefined>
  ): Promise<ApiResponse<R>> {
    return this.request<R>(routeDef, params);
  }

  async post<R = unknown>(
    routeDef: { method: HttpMethod; path: string },
    params?: Record<string, string | number | boolean | undefined>,
    data?: unknown
  ): Promise<ApiResponse<R>> {
    return this.request<R>(routeDef, params, data);
  }

  async patch<R = unknown>(
    routeDef: { method: HttpMethod; path: string },
    params?: Record<string, string | number | boolean | undefined>,
    data?: unknown
  ): Promise<ApiResponse<R>> {
    return this.request<R>(routeDef, params, data);
  }

  async delete<R = unknown>(
    routeDef: { method: HttpMethod; path: string },
    params?: Record<string, string | number | boolean | undefined>
  ): Promise<ApiResponse<R>> {
    return this.request<R>(routeDef, params);
  }

  // WebSocket connection - token passed as query param since headers not supported
  connectWebSocket(
    routeDef: { method: HttpMethod; path: string },
    params?: Record<string, string>
  ): WebSocket {
    let path = routeDef.path;

    // Substitute params and add token as query param
    if (params) {
      const queryParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        path = path.replace(`:${key}`, String(value));
      });
      if (this.token) {
        queryParams.set('token', this.token);
      }
      const query = queryParams.toString();
      if (query) {
        path += `?${query}`;
      }
    }

    const wsUrl = `${this.baseURL.replace(/^http/, 'ws')}${path}`;
    return new WebSocket(wsUrl);
  }
}

// Singleton instance for frontend use
let gatewayInstance: GatewayClient | null = null;

export function initGateway(config: GatewayConfig): GatewayClient {
  gatewayInstance = new GatewayClient({ config });
  return gatewayInstance;
}

export function getGateway(): GatewayClient {
  if (!gatewayInstance) {
    throw new Error('Gateway not initialized. Call initGateway() first.');
  }
  return gatewayInstance;
}

// Backwards compatibility - export a default client that can be configured
export const gatewayClient = {
  request: async <R = unknown>(
    method: HttpMethod,
    path: string,
    data?: unknown
  ): Promise<ApiResponse<R>> => {
    if (!gatewayInstance) {
      return {
        success: false,
        error: { code: 'NOT_INITIALIZED', message: 'Gateway not initialized' },
      };
    }
    return gatewayInstance.request<R>({ method, path }, {}, data);
  },
  get: async <R = unknown>(path: string, _params?: Record<string, any>): Promise<ApiResponse<R>> => {
    if (!gatewayInstance) {
      return { success: false, error: { code: 'NOT_INITIALIZED', message: 'Gateway not initialized' } };
    }
    return gatewayInstance.request<R>({ method: 'GET', path }, {}, undefined);
  },
  post: async <R = unknown>(path: string, data?: unknown): Promise<ApiResponse<R>> => {
    if (!gatewayInstance) {
      return { success: false, error: { code: 'NOT_INITIALIZED', message: 'Gateway not initialized' } };
    }
    return gatewayInstance.request<R>({ method: 'POST', path }, {}, data);
  },
  connectWebSocket(path: string, params?: Record<string, string>): WebSocket {
    if (!gatewayInstance) {
      throw new Error('Gateway not initialized');
    }
    return gatewayInstance.connectWebSocket({ method: 'WS', path }, params);
  },
};