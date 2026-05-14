'use client';

import { useEffect, useRef, useCallback, useState } from 'react';

export type WSConnectionStatus = 'Connecting' | 'Connected' | 'Disconnected' | 'Reconnecting';

export interface WSMessage {
  type: string;
  [key: string]: unknown;
}

interface UseWebSocketOptions {
  url: string;
  token?: string;
  onMessage?: (msg: WSMessage) => void;
  onStatusChange?: (status: WSConnectionStatus) => void;
  autoConnect?: boolean;
  maxRetries?: number;
}

const BACKOFF_DELAYS = [5000, 10000, 20000, 60000]; // 5s, 10s, 20s, 60s max
const PING_INTERVAL = 30000; // 30s
const MAX_PENDING = 100;

export function useWebSocket({
  url,
  token,
  onMessage,
  onStatusChange,
  autoConnect = true,
  maxRetries = 5,
}: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const retryCountRef = useRef(0);
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const pendingQueueRef = useRef<WSMessage[]>([]);

  const [status, setStatus] = useState<WSConnectionStatus>('Disconnected');

  const setStatusWithNotify = useCallback((s: WSConnectionStatus) => {
    setStatus(s);
    onStatusChange?.(s);
  }, [onStatusChange]);

  const clearTimers = useCallback(() => {
    if (retryTimeoutRef.current) clearTimeout(retryTimeoutRef.current);
    if (pingIntervalRef.current) clearInterval(pingIntervalRef.current);
  }, []);

  const flushPendingQueue = useCallback(() => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    while (pendingQueueRef.current.length > 0) {
      const msg = pendingQueueRef.current.shift()!;
      ws.send(JSON.stringify(msg));
    }
  }, []);

  const startPing = useCallback(() => {
    if (pingIntervalRef.current) clearInterval(pingIntervalRef.current);
    pingIntervalRef.current = setInterval(() => {
      const ws = wsRef.current;
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, PING_INTERVAL);
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;
    if (retryCountRef.current >= maxRetries) {
      setStatusWithNotify('Disconnected');
      return;
    }

    setStatusWithNotify(retryCountRef.current === 0 ? 'Connecting' : 'Reconnecting');

    try {
      const protocols = token ? [token] : [];
      const ws = new WebSocket(url, protocols);
      wsRef.current = ws;

      ws.onopen = () => {
        retryCountRef.current = 0;
        setStatusWithNotify('Connected');
        startPing();
        flushPendingQueue();
      };

      ws.onclose = () => {
        clearTimers();
        if (pingIntervalRef.current) clearInterval(pingIntervalRef.current);

        if (retryCountRef.current < maxRetries) {
          const delay = BACKOFF_DELAYS[Math.min(retryCountRef.current, BACKOFF_DELAYS.length - 1)];
          setStatusWithNotify('Reconnecting');
          retryTimeoutRef.current = setTimeout(() => {
            retryCountRef.current++;
            connect();
          }, delay);
        } else {
          setStatusWithNotify('Disconnected');
        }
      };

      ws.onerror = () => {
        // error event always precedes close, let onclose handle retry logic
      };

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data);

          if (msg.type === 'pong') return; // ignore pong, just keep-alive

          onMessage?.(msg);
        } catch {
          // non-JSON message
          onMessage?.({ type: 'raw', data: event.data });
        }
      };
    } catch {
      if (retryCountRef.current < maxRetries) {
        const delay = BACKOFF_DELAYS[Math.min(retryCountRef.current, BACKOFF_DELAYS.length - 1)];
        retryTimeoutRef.current = setTimeout(() => {
          retryCountRef.current++;
          connect();
        }, delay);
      } else {
        setStatusWithNotify('Disconnected');
      }
    }
  }, [url, onMessage, setStatusWithNotify, clearTimers, startPing, flushPendingQueue, maxRetries]);

  const disconnect = useCallback(() => {
    clearTimers();
    retryCountRef.current = maxRetries; // prevent auto-reconnect
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setStatusWithNotify('Disconnected');
  }, [clearTimers, setStatusWithNotify, maxRetries]);

  const reconnect = useCallback(() => {
    retryCountRef.current = 0;
    clearTimers();
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    connect();
  }, [clearTimers, connect]);

  const send = useCallback((msg: WSMessage) => {
    const ws = wsRef.current;
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(msg));
    } else {
      // queue message, flush on reconnect
      if (pendingQueueRef.current.length < MAX_PENDING) {
        pendingQueueRef.current.push(msg);
      }
    }
  }, []);

  useEffect(() => {
    if (autoConnect) connect();
    return () => {
      clearTimers();
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [autoConnect, connect, clearTimers]);

  return { status, send, disconnect, reconnect };
}
