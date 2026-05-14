// API Client - wraps GatewayClient for backwards compatibility
// All API calls now route through the unified gateway

import type { ApiError } from '@/types';
import { getAuthToken } from '@/lib/api-utils';
import { GatewayClient, initGateway, ServiceRoutes, type GatewayConfig } from '@/lib/gateway';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>;
}

// Internal gateway client instance
let internalGateway: GatewayClient | null = null;

function getGateway(): GatewayClient {
  if (!internalGateway) {
    const token = getAuthToken();
    const config: GatewayConfig = {
      baseURL: API_BASE,
      userId: '', // Will be set from auth store
      tenantId: 'default',
    };
    internalGateway = initGateway(config);
    if (token) {
      internalGateway.setToken(token);
    }
  }
  return internalGateway;
}

function setUserContext(userId: string, agentId?: string) {
  const gateway = getGateway();
  gateway.headers['X-User-ID'] = userId;
  if (agentId) {
    gateway.headers['X-Agent-ID'] = agentId;
  }
}

class ApiClient {
  private token: string | null = null;
  private userId: string | null = null;

  constructor() {
    this.loadAuth();
  }

  private loadAuth() {
    if (typeof document === 'undefined') return;
    const match = document.cookie.split('; ').find(row => row.startsWith('joblinker-auth='));
    if (match) {
      try {
        const auth = JSON.parse(decodeURIComponent(match.split('=')[1]));
        if (auth.token) {
          this.token = auth.token;
          const gateway = getGateway();
          gateway.setToken(auth.token);
        }
        if (auth.userId) {
          this.userId = auth.userId;
          setUserContext(auth.userId);
        }
      } catch {}
    }
  }

  setToken(token: string | null) {
    this.token = token;
    const gateway = getGateway();
    gateway.setToken(token);
  }

  setUserId(userId: string, agentId?: string) {
    setUserContext(userId, agentId);
  }

  private getToken(): string | null {
    return this.token || getAuthToken();
  }

  private async request<T>(
    endpoint: string,
    options: FetchOptions = {}
  ): Promise<T> {
    const { params, ...fetchOptions } = options;

    // Build URL with params
    let url = `${API_BASE}${endpoint}`;
    if (params) {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.append(key, String(value));
        }
      });
      const query = searchParams.toString();
      if (query) url += `?${query}`;
    }

    const token = this.getToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    // For backwards compatibility, use direct fetch but with gateway headers
    const gateway = getGateway();
    const response = await fetch(url, {
      ...fetchOptions,
      headers: { ...gateway.headers, ...headers, ...(fetchOptions.headers as Record<string, string> || {}) },
    });

    if (!response.ok) {
      let error: ApiError = { error: 'Unknown error' };
      try {
        error = await response.json();
      } catch {
        // Response was not JSON, use status text
        error = { error: response.statusText || 'Request failed' };
      }
      throw error;
    }

    try {
      return await response.json();
    } catch {
      throw { error: 'Invalid JSON response from server' };
    }
  }

  get<T>(endpoint: string, options?: FetchOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' });
  }

  post<T>(endpoint: string, data?: unknown, options?: FetchOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  put<T>(endpoint: string, data?: unknown, options?: FetchOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  patch<T>(endpoint: string, data?: unknown, options?: FetchOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  delete<T>(endpoint: string, options?: FetchOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

export const apiClient = new ApiClient();
export { ServiceRoutes, GatewayClient, initGateway };
