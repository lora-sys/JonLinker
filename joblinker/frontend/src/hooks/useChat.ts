'use client';

import { useState, useEffect, useRef } from 'react';

interface Message {
  id: string;
  sender_id: string;
  content_xml: string;
  intent_type: string;
  created_at: string;
}

interface UseChatOptions {
  matchId: string;
  enabled?: boolean;
}

export function useChat({ matchId, enabled = true }: UseChatOptions) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Fetch conversation history
  const fetchMessages = async () => {
    if (!matchId) return;
    setIsLoading(true);
    try {
      const token = localStorage.getItem('joblinker-auth');
      if (!token) return;
      const parsed = JSON.parse(token);
      const authToken = parsed.state?.token || parsed.token;

      const response = await fetch(`/api/messages/${matchId}`, {
        headers: { 'Authorization': `Bearer ${authToken}` }
      });
      if (response.ok) {
        const data = await response.json();
        setMessages(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      setError('Failed to fetch messages');
    } finally {
      setIsLoading(false);
    }
  };

  // Connect to WebSocket
  const connect = () => {
    if (!enabled || !matchId) return;

    const token = localStorage.getItem('joblinker-auth');
    if (!token) return;
    const parsed = JSON.parse(token);
    const authToken = parsed.state?.token || parsed.token;

    const wsHost = process.env.NEXT_PUBLIC_API_WS_URL || 'ws://localhost:8080';
    const ws = new WebSocket(`${wsHost}/api/messages/${matchId}/ws?token=${authToken}`);

    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
    };

    ws.onmessage = (event) => {
      try {
        // Try to parse as JSON first
        const data = JSON.parse(event.data);
        if (data.type === 'connected') {
          setIsConnected(true);
        } else if (data.type === 'message' || data.payload?.content_xml) {
          // Handle incoming XML message
          const xmlMsg = data.payload || data;
          setMessages(prev => [...prev, {
            id: xmlMsg.header?.message_id || Date.now().toString(),
            sender_id: xmlMsg.header?.sender_id || 'remote',
            content_xml: typeof xmlMsg === 'string' ? xmlMsg : JSON.stringify(xmlMsg),
            intent_type: xmlMsg.payload?.intent || 'INQUIRY',
            created_at: xmlMsg.header?.timestamp || new Date().toISOString(),
          }]);
        }
      } catch {
        // Not JSON - treat as XML message from WebSocket
        const xmlData = event.data;
        const intentMatch = xmlData.match(/<intent>([^<]+)<\/intent>/);
        const msgIdMatch = xmlData.match(/<message_id>([^<]+)<\/message_id>/);
        const senderMatch = xmlData.match(/<sender_id>([^<]+)<\/sender_id>/);
        const timestampMatch = xmlData.match(/<timestamp>([^<]+)<\/timestamp>/);

        setMessages(prev => [...prev, {
          id: msgIdMatch?.[1] || Date.now().toString(),
          sender_id: senderMatch?.[1] || 'remote',
          content_xml: xmlData,
          intent_type: intentMatch?.[1] || 'UNKNOWN',
          created_at: timestampMatch?.[1] || new Date().toISOString(),
        }]);
      }
    };

    ws.onerror = () => {
      setError('WebSocket connection failed');
      setIsConnected(false);
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Auto reconnect after 3 seconds
      if (enabled) {
        reconnectTimeoutRef.current = setTimeout(connect, 3000);
      }
    };

    wsRef.current = ws;
  };

  // Send message
  const sendMessage = async (contentXml: string, intentType: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    wsRef.current.send(contentXml);
  };

  // Send via REST API (fallback)
  const sendMessageREST = async (contentXml: string, intentType: string) => {
    const token = localStorage.getItem('joblinker-auth');
    if (!token) throw new Error('Not authenticated');
    const parsed = JSON.parse(token);
    const authToken = parsed.state?.token || parsed.token;

    const response = await fetch(`/api/messages/${matchId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`
      },
      body: JSON.stringify({ content_xml: contentXml, intent_type: intentType })
    });

    if (!response.ok) {
      throw new Error('Failed to send message');
    }

    const message = await response.json();
    setMessages(prev => [...prev, message]);
    return message;
  };

  // Disconnect
  const disconnect = () => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  };

  // Initial fetch and connect
  useEffect(() => {
    if (enabled && matchId) {
      fetchMessages();
      connect();
    }
    return () => {
      disconnect();
    };
  }, [matchId, enabled]);

  return {
    messages,
    isConnected,
    isLoading,
    error,
    sendMessage,
    sendMessageREST,
    connect,
    disconnect,
    fetchMessages,
  };
}