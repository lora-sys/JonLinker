// WebSocket URL builder for the JobLinker real-time messaging system

const WS_PROTOCOL = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:'
const API_WS_PATH = '/api/messages'

export interface WSGatewayParams {
  userId?: string
  agentId?: string
  tenantId?: string
}

/**
 * Build the WebSocket URL for the general messaging endpoint.
 * @param token - JWT auth token (passed as query param)
 * @param gatewayParams - Gateway identity params (user_id, agent_id, tenant_id)
 * @returns Full WebSocket URL string
 */
export function buildWSUrl(token: string, gatewayParams?: WSGatewayParams): string {
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost:8080'
  const params = new URLSearchParams()
  params.set('token', token)
  if (gatewayParams?.userId)
    params.set('user_id', gatewayParams.userId)
  if (gatewayParams?.agentId)
    params.set('agent_id', gatewayParams.agentId)
  params.set('tenant_id', gatewayParams?.tenantId || 'default')
  return `${WS_PROTOCOL}//${host}${API_WS_PATH}/ws?${params.toString()}`
}

/**
 * Build a match-specific WebSocket subscription URL.
 * @param matchId - Match UUID to subscribe to
 * @param token - JWT auth token
 * @param gatewayParams - Gateway identity params
 */
export function buildMatchWSUrl(matchId: string, token: string, gatewayParams?: WSGatewayParams): string {
  // Determine WS base host: prefer NEXT_PUBLIC_WS_URL env, fall back to current page host
  let wsHost: string
  const envUrl = process.env.NEXT_PUBLIC_WS_URL
  if (envUrl && envUrl.startsWith('ws')) {
    // Extract host:port from ws://host:port/path
    try {
      const u = new URL(envUrl)
      wsHost = u.host
    }
    catch {
      wsHost = typeof window !== 'undefined' ? window.location.host : 'localhost:8080'
    }
  }
  else {
    wsHost = typeof window !== 'undefined' ? window.location.host : 'localhost:8080'
  }
  const params = new URLSearchParams()
  params.set('token', token)
  if (gatewayParams?.userId)
    params.set('user_id', gatewayParams.userId)
  if (gatewayParams?.agentId)
    params.set('agent_id', gatewayParams.agentId)
  params.set('tenant_id', gatewayParams?.tenantId || 'default')
  return `${WS_PROTOCOL}//${wsHost}${API_WS_PATH}/${matchId}/ws?${params.toString()}`
}

export type { WSMessage } from '@/shared/hooks/useWebSocket'
