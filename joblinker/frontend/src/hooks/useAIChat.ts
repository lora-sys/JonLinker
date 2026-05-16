'use client'

import type { UIMessage } from 'ai'

import { useChat } from '@ai-sdk/react'
import { DefaultChatTransport } from 'ai'
import { useCallback, useEffect, useMemo, useState } from 'react'

import type { MatchStatus } from '@/types'

import { apiClient } from '@/lib/api_client'
import { buildMatchWSUrl } from '@/lib/websocket'

import { useWebSocket } from './useWebSocket'

export type FSMStage
  = | 'INTRODUCTION'
    | 'JOB_DESCRIPTION'
    | 'SALARY_NEGOTIATION'
    | 'INTERVIEWING'
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
    case 'mutual_interest': return 'JOB_DESCRIPTION'
    case 'negotiating': return 'SALARY_NEGOTIATION'
    case 'interview_scheduled': return 'INTERVIEWING'
    case 'offer_sent': case 'offered': return 'OFFER'
    case 'hired': case 'rejected': case 'completed': return 'COMPLETED'
    default: return 'INTRODUCTION'
  }
}

export function useAIChat({ matchId, enabled = true }: UseAIChatOptions) {
  const [input, setInput] = useState('')

  // useChat manages messages, sendMessage, status, etc. No input state.
  const transport = useMemo(
    () => new DefaultChatTransport({
      api: '/api/chat-proxy',
      credentials: 'include',
    }),
    [],
  )

  const {
    messages: aiMessages,
    sendMessage,
    stop,
    error,
    setMessages,
    regenerate,
    status,
  } = useChat({
    id: matchId,
    transport,
  })

  const isLoading = status === 'submitted' || status === 'streaming'

  const [fsmStage, setFsmStage] = useState<FSMStage>('INTRODUCTION')
  const [pendingConfirm, setPendingConfirm] = useState<PendingConfirm | null>(null)

  useEffect(() => {
    if (!enabled)
      return
    apiClient.get<{ status: MatchStatus }>(`/api/matches/${matchId}`)
      .then(data => setFsmStage(matchStatusToStage(data.status)))
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
  const wsUrl = useMemo(
    () => (enabled ? buildMatchWSUrl(matchId, wsToken) : ''),
    [enabled, matchId, wsToken],
  )

  const { status: wsStatus } = useWebSocket({
    url: wsUrl,
    token: wsToken,
    autoConnect: enabled,
    maxRetries: 3,
    onMessage: useCallback(
      (data: Record<string, unknown>) => {
        switch (data.type) {
          case 'fsm_state_change':
            setFsmStage(matchStatusToStage(String(data.new_state || '')))
            break
          case 'confirmation_needed':
            setPendingConfirm({
              match_id: String(data.match_id || matchId),
              intent: String(data.intent || ''),
              message_id: String(data.message_id || ''),
              content_xml: String(data.content_xml || ''),
            })
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

      const xml = `<message><payload><intent>INQUIRY</intent><parameters>{"message":"${text.replace(/"/g, '\\"')}"}</parameters></payload></message>`
      apiClient.post(`/api/messages/${matchId}`, {
        content_xml: xml,
        intent_type: 'INQUIRY',
      }).catch(console.error)

      sendMessage({ text })
      setInput('')
    },
    [input, isLoading, matchId, sendMessage],
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

  return {
    messages: aiMessages as UIMessage[],
    input,
    setInput,
    handleSubmit,
    isLoading,
    status,
    error: error?.message || null,
    stop,
    reload: () => regenerate({}),
    wsStatus,
    fsmStage,
    pendingConfirm,
    handleHumanConfirm,
  }
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
