'use client'

import type { UIMessage } from 'ai'

import { useChat } from '@ai-sdk/react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import type { MatchStatus } from '@/types'

import { apiClient } from '@/lib/api_client'
import { buildMatchWSUrl } from '@/lib/websocket'
import { useAuthStore } from '@/stores/auth'

import { useWebSocket } from './useWebSocket'

export type FSMStage
  = | 'INTRODUCTION'
    | 'NEGOTIATION'
    | 'INTERVIEW'
    | 'OFFER'
    | 'COMPLETED'

export interface PendingConfirm {
  match_id: string
  intent: string
  message_id: string
  content_xml?: string
}

interface UseAIChatOptions {
  matchId: string
  enabled?: boolean
}

function matchStatusToStage(status: string): FSMStage {
  switch (status) {
    case 'mutual_interest': case 'negotiating': return 'NEGOTIATION'
    case 'interview_scheduled': case 'interviewing': return 'INTERVIEW'
    case 'offer_sent': case 'offered': return 'OFFER'
    case 'hired': case 'rejected': case 'completed': return 'COMPLETED'
    default: return 'INTRODUCTION'
  }
}

export function useAIChat({ matchId, enabled = true }: UseAIChatOptions) {
  const [input, setInput] = useState('')
  const checkedConfirmRef = useRef(false)
  const agentIdsRef = useRef({ seeker: '', recruiter: '', current: '' })

  // useChat is a pure UI state container — no transport, no api endpoint.
  // Messages are populated exclusively via setMessages() from WebSocket onMessage or initial fetch.
  const {
    messages: aiMessages,
    stop,
    error,
    setMessages,
    status,
  } = useChat({
    id: matchId,
  })

  // Session metadata (not available on useChat.data in @ai-sdk/react v3)
  const [fsmStage, setFsmStage] = useState<FSMStage>('INTRODUCTION')
  const [sessionVersion, setSessionVersion] = useState<number | undefined>(undefined)
  const [sessionStatus, setSessionStatus] = useState<'active' | 'concluded' | undefined>(undefined)
  const [sessionSummary] = useState<string | undefined>(undefined)
  const [pendingConfirm, setPendingConfirm] = useState<PendingConfirm | null>(null)

  const loadInitialMessagesRef = useRef(false)
  const initialLoadAttemptedRef = useRef(false)

  useEffect(() => {
    // Skip during SSR — no localStorage available
    if (typeof window === 'undefined')
      return
    if (!enabled || loadInitialMessagesRef.current)
      return
    loadInitialMessagesRef.current = true
    initialLoadAttemptedRef.current = false

    // Step 1: fetch match data + session meta in parallel
    Promise.all([
      apiClient.get<Record<string, unknown>>(`/api/matches/${matchId}`),
      apiClient.get<Record<string, unknown>>(`/api/sessions/${matchId}`).catch(() => null),
      apiClient.get<ConversationResponse>(`/api/conversation/${matchId}`),
    ])
      .then(([matchData, sessionData, convData]) => {
        const seeker = String(matchData.seeker_agent_id || '')
        const recruiter = String(matchData.recruiter_agent_id || '')
        agentIdsRef.current = { seeker, recruiter, current: seeker }

        if (matchData.status) {
          setFsmStage(matchStatusToStage(String(matchData.status)))
        }
        if (sessionData && sessionData.version != null) {
          setSessionVersion(Number(sessionData.version))
        }
        if (sessionData && sessionData.status) {
          setSessionStatus(String(sessionData.status) as 'active' | 'concluded')
        }
        if (convData.session_version != null) {
          setSessionVersion(Number(convData.session_version))
        }
        if (convData.session_status) {
          setSessionStatus(String(convData.session_status) as 'active' | 'concluded')
        }
        if (matchData.status === 'offered' && !checkedConfirmRef.current) {
          checkedConfirmRef.current = true
          setPendingConfirm({
            match_id: matchId,
            intent: 'OFFER',
            message_id: '',
          })
        }

        // Load messages
        const msgsData = convData.messages || []
        if (msgsData.length === 0)
          return
        const msgs: UIMessage[] = msgsData
          .filter(m => extractTextFromXML(m.content_xml || ''))
          .map((m) => {
            const text = extractTextFromXML(m.content_xml || '')
            const isSeeker = m.sender_agent_id && m.sender_agent_id === seeker
            return {
              id: m.id || crypto.randomUUID(),
              role: isSeeker ? 'user' : 'assistant',
              parts: [{ type: 'text', text }] as UIMessage['parts'],
              createdAt: m.created_at ? new Date(m.created_at) : new Date(),
            } as UIMessage
          })
        setMessages((prev) => {
          if (prev.length > 0)
            return prev
          return msgs
        })
      })
      .catch(() => {})
    return () => {
      // Reset guard ref on unmount so Strict Mode double-invoke re-fetches
      loadInitialMessagesRef.current = false
    }
  }, [matchId, enabled, setMessages])

  const isLoading = status === 'submitted' || status === 'streaming'

  useEffect(() => {
    if (!enabled)
      return
    apiClient.get<Record<string, unknown>>(`/api/matches/${matchId}`)
      .then((data) => {
        setFsmStage(matchStatusToStage(String(data.status || '')))
        if (data.seeker_agent_id || data.recruiter_agent_id) {
          const seeker = String(data.seeker_agent_id || '')
          const recruiter = String(data.recruiter_agent_id || '')
          const current = String(data.seeker_agent_id || '')
          agentIdsRef.current = { seeker, recruiter, current }
        }
        if (data.status === 'offered' && !checkedConfirmRef.current) {
          checkedConfirmRef.current = true
          setPendingConfirm({
            match_id: matchId,
            intent: 'OFFER',
            message_id: '',
          })
        }
      })
      .catch(() => {})
  }, [matchId, enabled])

  useEffect(() => {
    if (!enabled)
      return
    const id = setInterval(async () => {
      try {
        const data = await apiClient.get<{ status: MatchStatus }>(`/api/matches/${matchId}`)
        setFsmStage(matchStatusToStage(data.status))
      }
      catch {}
    }, 10000)
    return () => clearInterval(id)
  }, [matchId, enabled])

  const wsToken = useMemo(() => getToken(), [])
  const user = useAuthStore(s => s.user)
  const wsUrl = useMemo(
    () => {
      if (!enabled)
        return ''
      return buildMatchWSUrl(matchId, wsToken, {
        userId: user?.id,
        agentId: agentIdsRef.current.current || undefined,
        tenantId: undefined,
      })
    },
    [enabled, matchId, wsToken, user?.id],
  )

  const { status: wsStatus } = useWebSocket({
    url: wsUrl,
    token: wsToken,
    autoConnect: enabled,
    maxRetries: 3,
    onMessage: useCallback(
      (data: Record<string, unknown>) => {
        switch (data.type) {
          case 'agent_response': {
            const text = String(data.text || data.content_xml || '')
            if (!text || text === '{}')
              break
            const { seeker } = agentIdsRef.current
            const isSeeker = data.sender_agent_id && String(data.sender_agent_id) === seeker
            setMessages(prev => [
              ...prev,
              {
                id: String(data.message_id || crypto.randomUUID()),
                role: isSeeker ? 'user' : 'assistant',
                parts: [{ type: 'text', text }] as UIMessage['parts'],
                createdAt: data.created_at ? new Date(String(data.created_at)) : new Date(),
              } as UIMessage,
            ])
            break
          }
          case 'fsm_state_change':
            setFsmStage(matchStatusToStage(String(data.new_state || '')))
            break
          case 'state_graph_message': {
            const rawText = String(data.text || '')
            if (!rawText)
              break
            const text = extractTextFromXML(rawText)
            const { seeker } = agentIdsRef.current
            const isSeeker = data.sender_agent_id && String(data.sender_agent_id) === seeker
            setMessages(prev => [
              ...prev,
              {
                id: crypto.randomUUID(),
                role: isSeeker ? 'user' : 'assistant',
                parts: [{ type: 'text', text }] as UIMessage['parts'],
                createdAt: data.timestamp ? new Date(String(data.timestamp)) : new Date(),
              } as UIMessage,
            ])
            break
          }
          case 'state_graph_complete':
            setFsmStage('COMPLETED')
            break
          case 'adk_stream_chunk': {
            const chunk = String(data.accumulated || data.chunk || '')
            if (!chunk)
              break
            const streamId = `stream-${matchId}`
            setMessages((prev) => {
              const last = prev[prev.length - 1]
              if (last && last.id === streamId) {
                return [
                  ...prev.slice(0, -1),
                  { ...last, parts: [{ type: 'text', text: chunk }] as UIMessage['parts'] },
                ]
              }
              return [
                ...prev,
                {
                  id: streamId,
                  role: 'assistant' as const,
                  parts: [{ type: 'text', text: chunk }] as UIMessage['parts'],
                  createdAt: new Date(),
                } as UIMessage,
              ]
            })
            break
          }
          case 'adk_tool_call':
            break
          case 'adk_error':
            break
          case 'human_rejected':
            setMessages(prev => [
              ...prev,
              {
                id: crypto.randomUUID(),
                role: 'system',
                parts: [{ type: 'text', text: `Human rejected ${String(data.confirm_type || 'request')}.` }],
              } as UIMessage,
            ])
            setPendingConfirm(null)
            break
          case 'session_created':
            if (data.version != null) {
              setSessionVersion(Number(data.version))
            }
            setSessionStatus('active')
            break
          case 'session_concluded':
            setSessionStatus('concluded')
            break
          case 'session_reopened':
            setMessages([])
            if (data.version != null) {
              setSessionVersion(Number(data.version))
            }
            setSessionStatus('active')
            break
        }
      },
      [matchId, setMessages],
    ),
  })

  const handleSubmit = useCallback(
    async (e?: React.FormEvent) => {
      e?.preventDefault()
      const text = input.trim()
      if (!text || isLoading)
        return

      setMessages(prev => [
        ...prev,
        {
          id: crypto.randomUUID(),
          role: 'user',
          parts: [{ type: 'text', text }] as UIMessage['parts'],
          createdAt: new Date(),
        } as UIMessage,
      ])

      const xml = `<message><payload><intent>INQUIRY</intent><parameters>{"message":"${text.replace(/"/g, '\\"')}"}</parameters></payload></message>`
      apiClient.post(`/api/messages/${matchId}`, {
        content_xml: xml,
        intent_type: 'INQUIRY',
      }).catch(console.error)

      setInput('')
    },
    [input, isLoading, matchId, setMessages],
  )

  const handleHumanConfirm = useCallback(
    async (approved: boolean) => {
      if (!pendingConfirm)
        return
      try {
        await apiClient.post(`/api/matches/${matchId}/human-confirm`, {
          approved,
          feedback: approved ? '' : 'User rejected',
        })
        setMessages(prev => [
          ...prev,
          {
            id: crypto.randomUUID(),
            role: 'system',
            parts: [{ type: 'text', text: approved
              ? `You accepted the ${pendingConfirm.intent}.`
              : `You rejected the ${pendingConfirm.intent}.` }],
          } as UIMessage,
        ])
        setPendingConfirm(null)
      }
      catch (err) {
        console.error('Confirm failed:', err)
      }
    },
    [pendingConfirm, matchId, setMessages],
  )

  const handleReopen = useCallback(async () => {
    try {
      const res = await apiClient.post<{ session_id: string }>(`/api/sessions/${matchId}/reopen`)
      if (res.session_id) {
        setMessages([])
        setSessionStatus('active')
      }
    }
    catch (err) {
      console.error('Reopen failed:', err)
    }
  }, [matchId, setMessages])

  return {
    messages: aiMessages as UIMessage[],
    setMessages,
    input,
    setInput,
    handleSubmit,
    isLoading,
    status,
    error: error?.message || null,
    stop,
    reload: () => {
      // WebSocket mode: messages are streamed via WS, not useChat API.
      // Errors are cleared on the next WS message — no regenerate needed.
    },
    wsStatus,
    fsmStage,
    sessionVersion,
    sessionStatus,
    sessionSummary,
    pendingConfirm,
    handleHumanConfirm,
    handleReopen,
  }
}

interface ConversationResponse {
  messages: MessageItem[]
  session_version?: number
  session_status?: string
}

interface MessageItem {
  id: string
  match_id?: string
  sender_agent_id?: string
  content_xml?: string
  intent_type?: string
  created_at?: string
}

function extractTextFromXML(xml: string): string {
  if (!xml || xml.startsWith('{'))
    return ''

  // Try <message><text>...</text></message> format (StateGraph messages)
  const textMatch = xml.match(/<text>([\s\S]*?)<\/text>/)
  if (textMatch && textMatch[1] && textMatch[1].trim())
    return textMatch[1].trim()

  // Try <parameters>{"message":"..."} format (Agent/ADK)
  const match = xml.match(/<parameters>(\{.*?\})<\/parameters>/)
  if (match && match[1]) {
    try {
      const parsed = JSON.parse(match[1])
      return parsed.message || parsed.text || ''
    }
    catch {
      return xml
    }
  }

  // Fallback: try any <text> tag
  const anyText = xml.match(/<text>([\s\S]*?)<\/text>/)
  if (anyText && anyText[1])
    return anyText[1].trim()

  return xml
}

function getToken(): string {
  if (typeof window === 'undefined')
    return ''
  try {
    const raw = localStorage.getItem('joblinker-auth')
    if (!raw)
      return ''
    const parsed = JSON.parse(raw)
    return parsed.state?.token || parsed.token || ''
  }
  catch {
    return ''
  }
}
