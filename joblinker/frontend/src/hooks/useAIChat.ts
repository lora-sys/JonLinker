'use client';

import { useEffect, useCallback, useRef, useState } from 'react';
import { useChat } from '@ai-sdk/react';
import { DefaultChatTransport } from 'ai';
import { useWebSocket } from './useWebSocket';
import { buildMatchWSUrl, type WSGatewayParams } from '@/lib/websocket';
import { loadThread, saveThread } from '@/lib/ai/thread-storage';
import type { ChatMessage } from '@/types/ai';

interface UseAIChatOptions {
  matchId: string;
  enabled?: boolean;
}

interface WSEvent {
  type: string;
  data?: Record<string, unknown>;
  content?: string;
  sender_id?: string;
  message_id?: string;
  created_at?: string;
}

interface StoredMessage {
  id: string;
  sender_agent_id: string;
  content_xml: string;
  intent_type: string;
  created_at: string;
}

export function useAIChat({ matchId, enabled = true }: UseAIChatOptions) {
  const [input, setInput] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [initialMessagesLoaded, setInitialMessagesLoaded] = useState(false);
  const wsRef = useRef<ReturnType<typeof useWebSocket> | null>(null);
  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const {
    messages: aiMessages,
    status: aiStatus,
    stop,
    regenerate,
    setMessages,
    sendMessage,
    error: aiError,
  } = useChat({
    transport: new DefaultChatTransport({
      api: '/api/chat',
      body: { matchId },
    }),
  });

  const isLoading = aiStatus === 'streaming' || aiStatus === 'submitted';

  // Poll REST API for message history (fallback when WebSocket fails)
  const loadMessagesFromREST = useCallback(async () => {
    try {
      const raw = typeof window !== 'undefined' ? localStorage.getItem('joblinker-auth') : null;
      let token = '';
      if (raw) {
        try {
          const parsed = JSON.parse(raw);
          token = parsed.state?.token || parsed.token || '';
        } catch {
          token = raw;
        }
      }

      const res = await fetch(`/api/messages/${matchId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const storedMessages: StoredMessage[] = await res.json();
        if (storedMessages.length > 0 && !initialMessagesLoaded) {
          const loadedMsgs = storedMessages.map(msg => {
            const content = parseXmlContent(msg.content_xml);
            const role = msg.sender_agent_id === 'system' ? 'system' : 'assistant';
            return {
              id: msg.id,
              role: role as 'user' | 'assistant' | 'system',
              parts: [{ type: 'text' as const, text: content }],
            };
          });
          setMessages(loadedMsgs);
          setInitialMessagesLoaded(true);
        }
      }
    } catch (err) {
      console.error('Failed to load messages from REST:', err);
    }
  }, [matchId, initialMessagesLoaded, setMessages]);

  // Load cached thread
  useEffect(() => {
    const cached = loadThread(matchId);
    if (cached && cached.messages.length > 0) {
      setMessages(cached.messages.map(m => ({
        id: m.id,
        role: m.role as 'user' | 'assistant' | 'system',
        parts: [{ type: 'text' as const, text: m.content }],
      })));
      setInitialMessagesLoaded(true);
    }
  }, [matchId, setMessages]);

  // Initial REST load and periodic polling as fallback
  useEffect(() => {
    loadMessagesFromREST();
    // Poll every 5 seconds as fallback
    pollIntervalRef.current = setInterval(loadMessagesFromREST, 5000);
    return () => {
      if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
    };
  }, [loadMessagesFromREST]);

  const ws = useWebSocket({
    url: enabled ? buildWSUrl(matchId) : '',
    token: getAuthToken(),
    autoConnect: enabled,
    onMessage: (event: WSEvent) => {
      switch (event.type) {
        case 'message':
        case 'ai_response_sent':
          if (event.content && event.sender_id !== 'user') {
            sendMessage({
              role: 'assistant',
              parts: [{ type: 'text' as const, text: parseXmlContent(event.content) }],
            });
          }
          break;
        case 'connected':
          setIsConnected(true);
          break;
      }
    },
  });

  useEffect(() => {
    wsRef.current = ws;
    setIsConnected(ws.status === 'Connected');
  }, [ws]);

  const handleSubmit = useCallback((e?: React.FormEvent) => {
    e?.preventDefault();
    if (!input.trim() || isLoading) return;
    sendMessage({ role: 'user', parts: [{ type: 'text' as const, text: input.trim() }] });
    setInput('');
  }, [input, isLoading, sendMessage]);

  const status = isLoading ? 'streaming' as const : 'done' as const;

  return {
    messages: aiMessages,
    input,
    setInput,
    status,
    isConnected,
    error: aiError?.message || null,
    append: (msg: { role: 'user' | 'assistant' | 'system'; content: string }) => {
      sendMessage({
        role: msg.role,
        parts: [{ type: 'text' as const, text: msg.content }],
      });
    },
    handleSubmit,
    stop,
    reload: () => regenerate(),
  };
}

function buildWSUrl(matchId: string): string {
  const token = getAuthToken();
  const gatewayParams: WSGatewayParams = { tenantId: 'default' };
  try {
    const raw = typeof window !== 'undefined' ? localStorage.getItem('joblinker-auth') : null;
    if (raw) {
      const parsed = JSON.parse(raw);
      const userId = parsed.state?.user?.id || parsed.userId;
      if (userId) gatewayParams.userId = userId;
    }
  } catch {}
  return buildMatchWSUrl(matchId, token, gatewayParams);
}

function getAuthToken(): string {
  if (typeof window === 'undefined') return '';
  const raw = localStorage.getItem('joblinker-auth');
  if (!raw) return '';
  try {
    const parsed = JSON.parse(raw);
    return parsed.state?.token || parsed.token || '';
  } catch {
    return raw;
  }
}

function parseXmlContent(contentXml: string): string {
  try {
    if (contentXml.includes('<message>')) {
      const paramsMatch = contentXml.match(/<parameters>([^<]+)<\/parameters>/);
      if (paramsMatch?.[1]) {
        try {
          const params = JSON.parse(paramsMatch[1]);
          if (params.message) return params.message;
          if (params.title) {
            return `${params.title} - ${params.location || ''} $${params.salary_min || 0}-${params.salary_max || 0}`;
          }
          return paramsMatch[1];
        } catch {
          return paramsMatch[1];
        }
      }
      const intentMatch = contentXml.match(/intent="([^"]+)"/);
      const intentTagMatch = contentXml.match(/<intent>([^<]+)<\/intent>/);
      const intent = intentMatch?.[1] || intentTagMatch?.[1] || 'UNKNOWN';
      return `[${intent}]`;
    }
    return contentXml;
  } catch {
    return contentXml;
  }
}
