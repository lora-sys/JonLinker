'use client'

import { MessageCircle, RefreshCw, WifiOff } from 'lucide-react'
import Link from 'next/link'
import { useCallback, useEffect, useRef, useState } from 'react'

import type { Conversation, Message } from '@/types'

import { EmptyState, ErrorState, LoadingSkeleton } from '@/components/ui'
import Card from '@/components/ui/Card'
import ConnectionStatus from '@/components/ui/ConnectionStatus'
import UnreadBadge from '@/components/ui/UnreadBadge'
import { apiClient } from '@/lib/api_client'

// Helper to parse A2A XML message content (aligned with useAIChat.ts)
function parseA2AMessage(xml: string): { intent: string, text: string, senderId: string } {
  try {
    const senderMatch = xml.match(/<sender_id>(.*?)<\/sender_id>/)
    const intentMatch = xml.match(/<intent>(.*?)<\/intent>/)
    const paramsMatch = xml.match(/<parameters>([^<]+)<\/parameters>/)
    let text = ''
    if (paramsMatch?.[1]) {
      try {
        const params = JSON.parse(paramsMatch[1])
        text = params.message || paramsMatch[1]
      }
      catch {
        text = paramsMatch[1]
      }
    }
    else {
      const textMatch = xml.match(/<text>([^<]*)<\/text>/)
      text = textMatch?.[1] || xml
    }
    return {
      senderId: senderMatch?.[1] ?? '',
      intent: intentMatch?.[1] ?? '',
      text,
    }
  }
  catch {
    return { senderId: '', intent: '', text: xml }
  }
}

const RECONNECT_DELAYS = [5000, 10000, 20000, 60000] // 5s, 10s, 20s, max 60s

type ConnectionState = 'connected' | 'disconnected' | 'reconnecting' | 'failed'

export default function MessagesPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected')
  const [, setIsTyping] = useState(false)
  const [, setReconnectAttempt] = useState(0)
  const [disconnectedTime, setDisconnectedTime] = useState<number | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const messageQueueRef = useRef<Message[]>([])
  const reconnectAttemptRef = useRef(0)

  const [conversations, setConversations] = useState<Conversation[]>([])

  const fetchMessages = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await apiClient.get<{ conversations: Conversation[] }>('/api/messages')
      const convs = data?.conversations || []
      setConversations(convs)
      const msgs: Message[] = convs
        .filter(c => c.LastMessage)
        .map(c => c.LastMessage!)
      setMessages(msgs)
    }
    catch {
      setError(null)
      setMessages([])
    }
    finally {
      setIsLoading(false)
    }
  }, [])

  const processMessageQueueRef = useRef<() => void>(() => {})
  processMessageQueueRef.current = () => {
    if (messageQueueRef.current.length > 0 && connectionState === 'connected') {
      setMessages(prev => [...prev, ...messageQueueRef.current])
      messageQueueRef.current = []
    }
  }

  const connectWebSocket = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN)
      return

    const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const token = localStorage.getItem('joblinker-auth')
    const apiUrl = new URL(API_BASE)
    let wsUrl = `${protocol}//${apiUrl.host}/api/messages/ws`
    if (token) {
      try {
        const parsed = JSON.parse(token)
        const actualToken = parsed.state?.token || parsed.token
        if (actualToken) {
          wsUrl += `?token=${encodeURIComponent(actualToken)}`
        }
      }
      catch {
        // ignore
      }
    }
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      setConnectionState('connected')
      setReconnectAttempt(0)
      reconnectAttemptRef.current = 0
      setDisconnectedTime(null)
      processMessageQueueRef.current()
    }

    ws.onclose = () => {
      setConnectionState('disconnected')
      setIsTyping(false)
      if (wsRef.current) {
        wsRef.current = null
        const attempt = reconnectAttemptRef.current
        if (attempt < RECONNECT_DELAYS.length) {
          setConnectionState('reconnecting')
          setDisconnectedTime(Date.now())
          const delay = RECONNECT_DELAYS[attempt]
          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttemptRef.current++
            setReconnectAttempt(attempt + 1)
            connectWebSocket()
          }, delay)
        }
        else {
          setConnectionState('failed')
        }
      }
    }

    ws.onerror = () => {
      setConnectionState('disconnected')
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'message' || data.type === 'ai_response_sent') {
          const msg = data.payload
          if (connectionState === 'connected') {
            setMessages(prev => [...prev, msg])
          }
          else {
            messageQueueRef.current.push(msg)
          }
        }
        else if (data.type === 'typing') {
          setIsTyping(data.payload.isTyping)
        }
        else if (data.type === 'ai_thinking') {
          setIsTyping(true)
        }
        else if (data.type === 'connected') {
          ws.send(JSON.stringify({ type: 'join_match', match_id: 'all' }))
        }
      }
      catch {
        // ignore parse errors
      }
    }
  }, []) // stable — no deps that change on connectionState

  const manualReconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    reconnectAttemptRef.current = 0
    setConnectionState('reconnecting')
    setReconnectAttempt(0)
    connectWebSocket()
  }, [connectWebSocket])

  const mountedRef = useRef(false)

  useEffect(() => {
    fetchMessages()
    mountedRef.current = true
    return () => {
      mountedRef.current = false
    }
  }, [fetchMessages])

  useEffect(() => {
    connectWebSocket()
    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
    }
  }, [connectWebSocket])

  // Check if disconnected for more than 5 seconds for "Reconnecting..." indicator
  useEffect(() => {
    if (connectionState === 'reconnecting' && disconnectedTime) {
      const checkDisconnectedTime = setInterval(() => {
        // Connection time check handled in render
      }, 1000)
      return () => clearInterval(checkDisconnectedTime)
    }
    return undefined
  }, [connectionState, disconnectedTime])

  const unreadCount = conversations.reduce((sum, c) => sum + (c.UnreadCount || 0), 0)

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3">
            <MessageCircle className="w-8 h-8 text-blue-600" />
            Messages
          </h1>
          <div className="flex items-center gap-3">
            <ConnectionStatus state={connectionState} />
            {connectionState === 'failed' && (
              <button
                onClick={manualReconnect}
                className="flex items-center gap-2 px-3 py-1.5 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors text-sm font-medium"
              >
                <RefreshCw className="w-4 h-4" />
                Retry
              </button>
            )}
            {connectionState === 'reconnecting' && disconnectedTime && (Date.now() - disconnectedTime > 5000) && (
              <span className="text-sm text-amber-600 flex items-center gap-1">
                <RefreshCw className="w-4 h-4 animate-spin" />
                Reconnecting...
              </span>
            )}
            <UnreadBadge count={unreadCount} />
          </div>
        </div>

        {isLoading
          ? (
              <LoadingSkeleton count={3} variant="card" />
            )
          : error
            ? (
                <ErrorState message={error} onRetry={fetchMessages} />
              )
            : messages.length === 0 && connectionState !== 'failed'
              ? (
                  <EmptyState
                    icon={<MessageCircle className="w-10 h-10 text-blue-600" />}
                    title="No messages yet"
                    description="Start a conversation with a match to see messages here"
                  />
                )
              : messages.length === 0 && connectionState === 'failed'
                ? (
                    <div className="flex flex-col items-center justify-center py-12">
                      <WifiOff className="w-12 h-12 text-slate-400 mb-4" />
                      <h3 className="text-lg font-medium text-slate-900 mb-2">Connection Failed</h3>
                      <p className="text-slate-500 mb-4">Unable to connect to the message server</p>
                      <button
                        onClick={manualReconnect}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                      >
                        <RefreshCw className="w-4 h-4" />
                        Reconnect
                      </button>
                    </div>
                  )
                : (
                    <div className="space-y-4">
                      {messages.map((message) => {
                        const parsed = parseA2AMessage(message.content_xml)
                        return (
                          <Link key={message.id} href={`/conversation/${message.match_id}`}>
                            <Card hover className="p-4 bg-white/80 backdrop-blur-xl border border-white/20 cursor-pointer">
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-3">
                                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-100 to-sky-100 flex items-center justify-center">
                                    <MessageCircle className="w-5 h-5 text-blue-600" />
                                  </div>
                                  <div>
                                    <div className="flex items-center gap-2">
                                      <p className="font-medium text-slate-900">
                                        {parsed.intent || 'Message'}
                                      </p>
                                      <span className="px-2 py-0.5 bg-blue-100 text-blue-700 text-xs rounded-full">
                                        {parsed.senderId.slice(0, 8)}
                                        ...
                                      </span>
                                    </div>
                                    <p className="text-sm text-slate-500 line-clamp-1">
                                      {parsed.text}
                                    </p>
                                  </div>
                                </div>
                                <span className="text-xs text-slate-400">
                                  {new Date(message.created_at).toLocaleDateString()}
                                </span>
                              </div>
                            </Card>
                          </Link>
                        )
                      })}
                    </div>
                  )}
      </div>
    </div>
  )
}
