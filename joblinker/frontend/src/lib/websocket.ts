// WebSocket URL builder for the JobLinker real-time messaging system

const WS_PROTOCOL = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const API_WS_PATH = '/api/messages/ws';

/**
 * Build the WebSocket URL for the general messaging endpoint.
 * @param token - JWT auth token (passed as query param)
 * @returns Full WebSocket URL string
 */
export function buildWSUrl(token: string): string {
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost:8080';
  return `${WS_PROTOCOL}//${host}${API_WS_PATH}?token=${encodeURIComponent(token)}`;
}

/**
 * Build a match-specific WebSocket subscription URL.
 * @param matchId - Match UUID to subscribe to
 * @param token - JWT auth token
 */
export function buildMatchWSUrl(matchId: string, token: string): string {
  return `${buildWSUrl(token)}`; // Same URL, use join_match after connect
}

export type { WSMessage } from '@/hooks/useWebSocket';
