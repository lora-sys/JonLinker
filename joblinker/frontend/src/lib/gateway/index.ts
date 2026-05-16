// API Gateway Unified Layer
// Sole entry point for all API calls from frontend, AI agents, MCP tools, Function Calling

export { GatewayClient, gatewayClient, getGateway, initGateway, ServiceRoutes } from './client'
export type { ApiResponse, GatewayClientOptions, GatewayConfig, HttpMethod } from './client'
export { CircuitBreaker, getTenantContext, RateLimiter } from './middleware'
export type { TenantContext } from './middleware'
export { ServiceRoutes as Routes } from './routes'
export type { HttpMethod as RouteMethod } from './routes'
export type { RequestInterceptor, RequestLog, ResponseInterceptor } from './types'
