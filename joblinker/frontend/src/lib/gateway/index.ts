// API Gateway Unified Layer
// Sole entry point for all API calls from frontend, AI agents, MCP tools, Function Calling

export { GatewayClient, initGateway, getGateway, gatewayClient, ServiceRoutes } from './client';
export type { GatewayConfig, ApiResponse, GatewayClientOptions, HttpMethod } from './client';
export { ServiceRoutes as Routes } from './routes';
export type { HttpMethod as RouteMethod } from './routes';
export { RateLimiter, CircuitBreaker, getTenantContext } from './middleware';
export type { TenantContext } from './middleware';
export type { RequestInterceptor, ResponseInterceptor, RequestLog } from './types';
