// WebSocket URL builder for the JobLinker real-time messaging system

const WS_PROTOCOL = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const API_WS_PATH = '/api/messages';

export interface WSGatewayParams {
  userId?: string;
  agentId?: string;
  tenantId?: string;
}

/**
 * Build the WebSocket URL for the general messaging endpoint.
 * @param token - JWT auth token (passed as query param)
 * @param gatewayParams - Gateway identity params (user_id, agent_id, tenant_id)
 * @returns Full WebSocket URL string
 */
export function buildWSUrl(token: string, gatewayParams?: WSGatewayParams): string {
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost:8080';
  const params = new URLSearchParams();
  params.set('token', token);
  if (gatewayParams?.userId) params.set('user_id', gatewayParams.userId);
  if (gatewayParams?.agentId) params.set('agent_id', gatewayParams.agentId);
  params.set('tenant_id', gatewayParams?.tenantId || 'default');
  return `${WS_PROTOCOL}//${host}${API_WS_PATH}/ws?${params.toString()}`;
}

/**
 * Build a match-specific WebSocket subscription URL.
 * @param matchId - Match UUID to subscribe to
 * @param token - JWT auth token
 * @param gatewayParams - Gateway identity params
 */
export function buildMatchWSUrl(matchId: string, token: string, gatewayParams?: WSGatewayParams): string {
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost:8080';
  const params = new URLSearchParams();
  params.set('token', token);
  if (gatewayParams?.userId) params.set('user_id', gatewayParams.userId);
  if (gatewayParams?.agentId) params.set('agent_id', gatewayParams.agentId);
  params.set('tenant_id', gatewayParams?.tenantId || 'default');
  return `${WS_PROTOCOL}//${host}${API_WS_PATH}/${matchId}/ws?${params.toString()}`;
}

export type { WSMessage } from '@/hooks/useWebSocket';
