'use client'

import { PanelRightClose, PanelRightOpen, X } from 'lucide-react'
import { use, useCallback, useEffect, useState } from 'react'

import {
  Conversation,
  ConversationContent,
  ConversationEmptyState,
  ConversationScrollButton,
} from '@/components/ai-elements/conversation'
import {
  Message,
  MessageContent,
  MessageResponse,
} from '@/components/ai-elements/message'
import {
  PromptInput,
  PromptInputTextarea,
  PromptInputSubmit,
  PromptInputProvider,
} from '@/components/ai-elements/prompt-input'
import type { Interview, Offer } from '@/types'

import { ConversationHeader } from '@/components/conversation/ConversationHeader'
import { FlowPanel } from '@/components/conversation/FlowPanel'
import { SummaryPanel } from '@/components/conversation/SummaryPanel'
import { FSMStatusBar } from '@/components/FSMStatusBar'
import { HumanConfirmModal } from '@/components/HumanConfirmModal'
import { InterviewCard } from '@/components/interview/InterviewCard'
import { OfferCard } from '@/components/offer/OfferCard'
import { useAIChat } from '@/hooks/useAIChat'
import { apiClient } from '@/lib/api_client'

interface PageProps {
  params: Promise<{ matchId: string }>
}

export default function ConversationPage({ params }: PageProps) {
  const { matchId } = use(params)
  const [interview, setInterview] = useState<Interview | null>(null)
  const [offer, setOffer] = useState<Offer | null>(null)
  const [notification, setNotification] = useState<string | null>(null)
  const [showFlowPanel, setShowFlowPanel] = useState(true)

  const {
    messages,
    setMessages,
    input,
    setInput,
    handleSubmit,
    isLoading,
    error,
    stop,
    reload,
    wsStatus,
    fsmStage,
    sessionVersion,
    sessionStatus,
    pendingConfirm,
    handleHumanConfirm,
    handleReopen,
  } = useAIChat({ matchId })

  // Load interview/offer cards
  useEffect(() => {
    const fetchCards = async () => {
      try {
        const matchData = await apiClient.get<{ status: string }>(`/api/matches/${matchId}`)
        if (matchData.status === 'interview_scheduled' || matchData.status === 'interviewing') {
          try {
            const data = await apiClient.get<Interview>(`/api/interviews/${matchId}`)
            setInterview(data)
          }
          catch {}
        }
        if (matchData.status === 'offer_sent' || matchData.status === 'offered' || matchData.status === 'hired') {
          try {
            const data = await apiClient.get<Offer>(`/api/offers/${matchId}`)
            setOffer(data)
          }
          catch {}
        }
      }
      catch {}
    }
    fetchCards()
  }, [matchId])

  const handleInterviewConfirm = async () => {
    try {
      await apiClient.post(`/api/interviews/${matchId}/confirm`)
      setInterview(await apiClient.get<Interview>(`/api/interviews/${matchId}`))
    }
    catch (err) {
      console.error('Failed to confirm interview:', err)
    }
  }

  const handleInterviewCancel = async () => {
    try {
      await apiClient.post(`/api/interviews/${matchId}/cancel`)
      setInterview(await apiClient.get<Interview>(`/api/interviews/${matchId}`))
    }
    catch (err) {
      console.error('Failed to cancel interview:', err)
    }
  }

  const handleOfferAccept = async () => {
    try {
      await apiClient.post(`/api/offers/${matchId}/accept`)
      setOffer(await apiClient.get<Offer>(`/api/offers/${matchId}`))
    }
    catch (err) {
      console.error('Failed to accept offer:', err)
    }
  }

  const handlePromptSubmit = useCallback(
    ({ text }: { text: string }) => {
      if (!text.trim() || isLoading) return

      setMessages(prev => [
        ...prev,
        {
          id: crypto.randomUUID(),
          role: 'user' as const,
          parts: [{ type: 'text' as const, text }],
          createdAt: new Date(),
        },
      ])

      const xml = `<message><payload><intent>INQUIRY</intent><parameters>{"message":"${text.replace(/"/g, '\\"')}"}</parameters></payload></message>`
      apiClient.post(`/api/messages/${matchId}`, {
        content_xml: xml,
        intent_type: 'INQUIRY',
      }).catch(console.error)
    },
    [isLoading, matchId, setMessages],
  )

  return (
    <PromptInputProvider>
      <div className="flex flex-col h-[calc(100vh-4rem)]">
        {notification && (
          <div className="fixed top-4 right-4 z-50 flex items-center gap-2 px-4 py-3 bg-green-600 text-white rounded-lg shadow-lg">
            <span className="text-sm font-medium">{notification}</span>
            <button onClick={() => setNotification(null)} className="p-0.5 hover:bg-green-700 rounded">
              <X className="w-4 h-4" />
            </button>
          </div>
        )}

        <ConversationHeader
          matchId={matchId}
          fsmStage={fsmStage}
          wsStatus={wsStatus}
          sessionVersion={sessionVersion}
          sessionStatus={sessionStatus}
          onReopen={handleReopen}
        />

        <FSMStatusBar currentStage={fsmStage} />

        <div className="flex flex-1 overflow-hidden">
          <div className={`flex flex-col border-r border-gray-200 transition-all duration-300 ${
            showFlowPanel ? 'w-full lg:w-[65%]' : 'w-full'
          }`}
          >
            <Conversation>
              <ConversationContent>
                {messages.length === 0 && (
                  <ConversationEmptyState />
                )}

                {messages.map(msg => (
                  <Message key={msg.id} from={msg.role}>
                    <MessageContent>
                      {msg.parts?.map((part, i) => {
                        if (part.type === 'text') {
                          return (
                            <MessageResponse key={i}>
                              {part.text}
                            </MessageResponse>
                          )
                        }
                        if (part.type === 'reasoning') {
                          return null
                        }
                        if (part.type === 'tool-invocation') {
                          return null
                        }
                        return null
                      })}
                    </MessageContent>
                  </Message>
                ))}
              </ConversationContent>
              <ConversationScrollButton />
            </Conversation>

            {(interview || offer) && (
              <div className="px-4 py-2 border-t border-gray-100 space-y-2">
                {interview && (
                  <InterviewCard interview={interview} onConfirm={handleInterviewConfirm} onCancel={handleInterviewCancel} />
                )}
                {offer && (
                  <OfferCard offer={offer} onRespond={handleOfferAccept} />
                )}
              </div>
            )}

            <PromptInput
              onSubmit={handlePromptSubmit}
            >
              <PromptInputTextarea autoFocus placeholder="Type a message..." />
              <PromptInputSubmit onStop={stop} />
            </PromptInput>
          </div>

          {showFlowPanel
            ? (
                <div className="w-[35%] overflow-y-auto p-4 bg-gray-50 hidden lg:block">
                  <div className="flex items-center justify-between mb-4">
                    <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Process Flow</h2>
                    <button onClick={() => setShowFlowPanel(false)} className="p-1 text-gray-400 hover:text-gray-600 rounded">
                      <PanelRightClose className="w-4 h-4" />
                    </button>
                  </div>
                  <FlowPanel currentStage={fsmStage} />
                  <div className="mt-4">
                    <SummaryPanel matchId={matchId} />
                  </div>
                  {(interview || offer) && (
                    <div className="mt-6 space-y-4">
                      {interview && <InterviewCard interview={interview} onConfirm={handleInterviewConfirm} onCancel={handleInterviewCancel} />}
                      {offer && <OfferCard offer={offer} onRespond={handleOfferAccept} />}
                    </div>
                  )}
                </div>
              )
            : (
                <button
                  onClick={() => setShowFlowPanel(true)}
                  className="hidden lg:flex items-center gap-1 px-2 py-1 text-gray-400 hover:text-gray-600 border-l border-gray-200"
                >
                  <PanelRightOpen className="w-4 h-4" />
                </button>
              )}
        </div>

        <div className="lg:hidden">
          {showFlowPanel && (
            <div className="fixed inset-0 z-40 bg-black/50" onClick={() => setShowFlowPanel(false)}>
              <div className="absolute bottom-0 left-0 right-0 max-h-[50vh] bg-white rounded-t-2xl shadow-xl overflow-y-auto p-4 pb-8" onClick={e => e.stopPropagation()}>
                <div className="flex justify-center mb-2">
                  <div className="w-10 h-1 bg-gray-300 rounded-full" />
                </div>
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Process Flow</h2>
                  <button onClick={() => setShowFlowPanel(false)} className="p-1 text-gray-400 hover:text-gray-600 rounded">
                    <X className="w-4 h-4" />
                  </button>
                </div>
                <FlowPanel currentStage={fsmStage} />
                <div className="mt-4">
                  <SummaryPanel matchId={matchId} />
                </div>
              </div>
            </div>
          )}
        </div>

        {pendingConfirm && (
          <HumanConfirmModal
            pendingConfirm={pendingConfirm}
            onConfirm={handleHumanConfirm}
            onDismiss={() => {}}
          />
        )}
      </div>
    </PromptInputProvider>
  )
}
