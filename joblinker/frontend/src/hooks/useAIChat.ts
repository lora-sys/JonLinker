'use client';

import { useEffect, useCallback, useRef, useState } from 'react';
import { useChat } from '@ai-sdk/react';
import { DefaultChatTransport } from 'ai';
import { useWebSocket } from './useWebSocket';
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

export function useAIChat({ matchId, enabled = true }: UseAIChatOptions) {
  const [input, setInput] = useState('');
  const wsRef = useRef<ReturnType<typeof useWebSocket> | null>(null);

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

  useEffect(() => {
    const cached = loadThread(matchId);
    if (cached && cached.messages.length > 0) {
      setMessages(cached.messages.map(m => ({
        id: m.id,
        role: m.role as 'user' | 'assistant' | 'system',
        parts: [{ type: 'text' as const, text: m.content }],
      })));
    }
  }, [matchId, setMessages]);

  const ws = useWebSocket({
    url: enabled ? buildWSUrl(matchId) : '',
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
      }
    },
  });

  useEffect(() => {
    wsRef.current = ws;
  }, [ws]);

  const handleSubmit = useCallback((e?: React.FormEvent) => {
    e?.preventDefault();
    if (!input.trim() || isLoading) return;
    sendMessage({ role: 'user', parts: [{ type: 'text' as const, text: input.trim() }] });
    setInput('');
  }, [input, isLoading, sendMessage]);

  const isConnected = ws.status === 'Connected';

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
    loadMoreMessages: () => {},
    hasMoreMessages: false,
  };
}

function buildWSUrl(matchId: string): string {
  const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  const url = new URL(apiBase);
  const protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
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
  return `${protocol}//${url.host}/api/messages/${matchId}/ws?token=${encodeURIComponent(token)}`;
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
