'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import { useChat as useWSChat } from './useChat';
import { useStream } from './useStream';
import { loadThread, saveThread } from '@/lib/ai/thread-storage';
import type { ChatMessage, MessageRole, MessageStatus } from '@/types/ai';

interface UseAIChatOptions {
  matchId: string;
  enabled?: boolean;
}

interface UseAIChatReturn {
  messages: ChatMessage[];
  input: string;
  setInput: (value: string) => void;
  status: MessageStatus;
  isConnected: boolean;
  error: string | null;
  append: (message: { role: MessageRole; content: string }) => Promise<void>;
  handleSubmit: (e?: React.FormEvent) => Promise<void>;
  stop: () => void;
  reload: () => Promise<void>;
  loadMoreMessages: () => void;
  hasMoreMessages: boolean;
}

const MESSAGES_PER_PAGE = 50;

export function useAIChat({ matchId, enabled = true }: UseAIChatOptions): UseAIChatReturn {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [allMessages, setAllMessages] = useState<ChatMessage[]>([]);
  const [displayCount, setDisplayCount] = useState(MESSAGES_PER_PAGE);
  const [status, setStatus] = useState<MessageStatus>('done');
  const [error, setError] = useState<string | null>(null);
  
  const streamingMsgIdRef = useRef<string | null>(null);
  const messagesRef = useRef<ChatMessage[]>([]);
  
  const wsChat = useWSChat({ matchId, enabled });
  const stream = useStream({
    throttleMs: 33, // ~30 chars/sec
    onComplete: () => {
      setStatus('done');
      streamingMsgIdRef.current = null;
    },
    onError: (err) => {
      setStatus('error');
      setError(err);
      streamingMsgIdRef.current = null;
    },
  });

  // Keep ref in sync
  useEffect(() => {
    messagesRef.current = allMessages;
  }, [allMessages]);

  // Load thread on mount: localStorage first, then WS fallback
  useEffect(() => {
    const cached = loadThread(matchId);
    if (cached && cached.messages.length > 0) {
      setAllMessages(cached.messages);
      setMessages(cached.messages.slice(-displayCount));
    } else if (wsChat.messages.length > 0) {
      const converted = wsChat.messages.map((msg, idx): ChatMessage => {
        const content = parseXmlContent(msg.content_xml);
        return {
          id: msg.id || `msg-${idx}`,
          matchId,
          role: detectRole(msg.sender_id),
          content,
          createdAt: new Date(msg.created_at),
          status: 'done',
        };
      });
      setAllMessages(converted);
      setMessages(converted.slice(-displayCount));
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [matchId]);

  // Sync visible messages when displayCount changes
  useEffect(() => {
    setMessages(allMessages.slice(-displayCount));
  }, [allMessages, displayCount]);

  // Listen for new WS messages and animate them
  useEffect(() => {
    const wsMessages = wsChat.messages;
    const aiMessages = messagesRef.current;

    if (wsMessages.length > aiMessages.length) {
      const newWsMsgs = wsMessages.slice(aiMessages.length);

      newWsMsgs.forEach((msg, idx) => {
        const content = parseXmlContent(msg.content_xml);
        const role = detectRole(msg.sender_id);
        const msgId = msg.id || `ws-msg-${Date.now()}-${idx}`;

        const newMessage: ChatMessage = {
          id: msgId,
          matchId,
          role,
          content,
          createdAt: new Date(msg.created_at),
          status: 'done',
        };

        setAllMessages(prev => [...prev, newMessage]);

        // Animate assistant messages
        if (role === 'assistant') {
          setStatus('streaming');
          streamingMsgIdRef.current = msgId;
          stream.startStreaming(msgId, content);
        }
      });
    }
  }, [wsChat.messages.length, matchId, stream]);

  // Persist thread to localStorage
  useEffect(() => {
    if (allMessages.length > 0) {
      saveThread({
        id: matchId,
        messages: allMessages,
        createdAt: allMessages[0]?.createdAt ?? new Date(),
        updatedAt: new Date(),
        status: 'active',
      });
    }
  }, [allMessages, matchId]);

  const detectRole = (senderId: string): MessageRole => {
    if (senderId === 'user' || senderId === 'self') return 'user';
    if (senderId === 'system') return 'system';
    return 'assistant';
  };

  const parseXmlContent = (contentXml: string): string => {
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
  };

  const append = useCallback(async (message: { role: MessageRole; content: string }) => {
    const msgId = `msg-${Date.now()}`;
    const newMessage: ChatMessage = {
      id: msgId,
      matchId,
      role: message.role,
      content: message.content,
      createdAt: new Date(),
      status: 'done',
    };

    setAllMessages(prev => [...prev, newMessage]);

    if (message.role === 'user') {
      setStatus('pending');
      setError(null);

      try {
        const intentType = detectIntent(message.content);
        const contentXml = `<message><header><message_id>${msgId}</message_id></header><payload><intent>${intentType}</intent><message_text>${escapeXml(message.content)}</message_text></payload></message>`;
        await wsChat.sendMessageREST(contentXml, intentType);
        setInput('');
      } catch (err) {
        const errorMsg = err instanceof Error ? err.message : 'Failed to send message';
        setError(errorMsg);
        setStatus('error');
      }
    }
  }, [matchId, wsChat]);

  const handleSubmit = useCallback(async (e?: React.FormEvent) => {
    e?.preventDefault();
    if (!input.trim() || status === 'streaming' || status === 'pending') return;

    await append({ role: 'user', content: input.trim() });
  }, [input, status, append]);

  const stop = useCallback(() => {
    stream.complete();
    setStatus('done');
    streamingMsgIdRef.current = null;
  }, [stream]);

  const reload = useCallback(async () => {
    // Retry the last user message
    const lastUserMsg = [...messagesRef.current].reverse().find(m => m.role === 'user');
    if (lastUserMsg) {
      setStatus('pending');
      setError(null);
      try {
        const intentType = detectIntent(lastUserMsg.content);
        const contentXml = `<message><header><message_id>${lastUserMsg.id}</message_id></header><payload><intent>${intentType}</intent><message_text>${escapeXml(lastUserMsg.content)}</message_text></payload></message>`;
        await wsChat.sendMessageREST(contentXml, intentType);
      } catch (err) {
        const errorMsg = err instanceof Error ? err.message : 'Failed to resend message';
        setError(errorMsg);
        setStatus('error');
      }
    }
  }, [wsChat]);

  const loadMoreMessages = useCallback(() => {
    setDisplayCount(prev => prev + MESSAGES_PER_PAGE);
  }, []);

  const hasMoreMessages = allMessages.length > displayCount;

  return {
    messages,
    input,
    setInput,
    status,
    isConnected: wsChat.isConnected,
    error,
    append,
    handleSubmit,
    stop,
    reload,
    loadMoreMessages,
    hasMoreMessages,
  };
}

function detectIntent(text: string): string {
  const lower = text.toLowerCase();
  if (lower.includes('accept') || lower.includes('yes')) return 'ACCEPT';
  if (lower.includes('decline') || lower.includes('no')) return 'DECLINE';
  if (lower.includes('offer')) return 'OFFER';
  if (lower.includes('salary') || lower.includes('compensat')) return 'NEGOTIATION';
  if (lower.includes('schedule') || lower.includes('interview')) return 'SCHEDULE';
  return 'INQUIRY';
}

function escapeXml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}
