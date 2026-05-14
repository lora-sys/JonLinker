'use client';

import { useEffect, useCallback, useRef, useState } from 'react';
import { useChat } from '@ai-sdk/react';
import { useWebSocket } from './useWebSocket';
import { buildMatchWSUrl, type WSGatewayParams } from '@/lib/websocket';
import { apiClient } from '@/lib/api_client';
import type { ChatMessage } from '@/types/ai';

interface UseAIChatOptions {
  matchId: string;
  enabled?: boolean;
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
  const loadedIdsRef = useRef<Set<string>>(new Set());

  // useChat as UI container only — no HTTP requests
  const {
    messages: aiMessages,
    status: aiStatus,
    stop,
    setMessages,
    error: aiError,
  } = useChat();

  const isLoading = aiStatus === 'streaming' || aiStatus === 'submitted';

  // Load initial messages + poll every 3s
  useEffect(() => {
    let active = true;
    const poll = async () => {
      try {
        const data = await apiClient.get<StoredMessage[]>(`/api/conversation/${matchId}`);
        if (!active || !Array.isArray(data)) return;

        const newMsgs = data
          .filter(m => !loadedIdsRef.current.has(m.id))
          .map(m => {
            loadedIdsRef.current.add(m.id);
            const content = parseXmlContent(m.content_xml);
            const role = m.sender_agent_id === 'system' ? 'system' as const : 'assistant' as const;
            return {
              id: m.id,
              role,
              parts: [{ type: 'text' as const, text: content }],
            };
          });

        if (newMsgs.length > 0) {
          setMessages(prev => [...prev, ...newMsgs]);
        }
      } catch {
        // silent — poll will retry
      }
    };
    poll();
    const id = setInterval(poll, 3000);
    return () => { active = false; clearInterval(id); };
  }, [matchId, setMessages]);

  // WebSocket for FSM state changes only (optional)
  const { status: wsStatus } = useWebSocket({
    url: enabled ? buildMatchWSUrl(matchId, getToken(), { tenantId: 'default' }) : '',
    token: getToken(),
    autoConnect: enabled,
    maxRetries: 2,
    onMessage: useCallback(() => {
      // FSM state notifications handled elsewhere
    }, []),
  });

  useEffect(() => {
    setIsConnected(wsStatus === 'Connected');
  }, [wsStatus]);

  // Submit: direct REST POST through Gateway, useChat.sendMessage NOT called
  const handleSubmit = useCallback(async (e?: React.FormEvent) => {
    e?.preventDefault();
    const text = input.trim();
    if (!text || isLoading) return;

    // Add user message to UI immediately
    const tempId = crypto.randomUUID();
    setMessages(prev => [...prev, {
      id: tempId,
      role: 'user',
      parts: [{ type: 'text' as const, text }],
    }]);
    setInput('');

    // Build agent XML and POST through Gateway
    const xml = `<message><payload><intent>INQUIRY</intent><parameters>{"message":"${text.replace(/"/g, '\\"')}"}</parameters></payload></message>`;
    try {
      await apiClient.post(`/api/conversation/${matchId}`, {
        content_xml: xml,
        intent_type: 'INQUIRY',
      });
    } catch (err) {
      console.error('Failed to send message:', err);
    }
  }, [input, isLoading, matchId, setMessages]);

  return {
    messages: aiMessages,
    input,
    setInput,
    status: aiStatus,
    isConnected,
    wsStatus,
    error: aiError?.message || null,
    handleSubmit,
    stop,
    reload: () => { loadedIdsRef.current.clear(); },
  };
}

function getToken(): string {
  if (typeof window === 'undefined') return '';
  try {
    const raw = localStorage.getItem('joblinker-auth');
    if (!raw) return '';
    const parsed = JSON.parse(raw);
    return parsed.state?.token || parsed.token || '';
  } catch { return ''; }
}

function parseXmlContent(contentXml: string): string {
  try {
    if (contentXml.includes('<message>')) {
      // <parameters>{"message":"Hello"}</parameters> (AI-generated)
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
      // <content><text>Hello</text></content> (seed data format)
      const textMatch = contentXml.match(/<text>([^<]*)<\/text>/);
      if (textMatch?.[1]) {
        return textMatch[1];
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
