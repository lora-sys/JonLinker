'use client';

import { useEffect, useState, useCallback, useRef } from 'react';
import { use } from 'react';
import { X, PanelRightClose, PanelRightOpen } from 'lucide-react';
import { useAIChat } from '@/hooks/useAIChat';
import { ConversationHeader } from '@/components/conversation/ConversationHeader';
import { AgentMessageBubble, type AgentRole } from '@/components/conversation/AgentMessageBubble';
import { TextareaInput } from '@/components/conversation/TextareaInput';
import { FlowPanel, type FSMStage } from '@/components/conversation/FlowPanel';
import { InterviewCard } from '@/components/interview/InterviewCard';
import { OfferCard } from '@/components/offer/OfferCard';
import { getAuthToken } from '@/lib/api-utils';
import type { Match, Interview, Offer } from '@/types';

interface PageProps {
  params: Promise<{ matchId: string }>;
}

export default function ConversationPage({ params }: PageProps) {
  const { matchId } = use(params);
  const [match, setMatch] = useState<Match | null>(null);
  const [interview, setInterview] = useState<Interview | null>(null);
  const [offer, setOffer] = useState<Offer | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [notification, setNotification] = useState<string | null>(null);
  const [fsmStage, setFsmStage] = useState<FSMStage>('INTRODUCTION');
  const [showFlowPanel, setShowFlowPanel] = useState(true);
  const scrollRef = useRef<HTMLDivElement>(null);

  const {
    messages,
    input,
    setInput,
    status,
    isConnected,
    error,
    handleSubmit,
    stop,
    reload,
  } = useAIChat({ matchId });

  const isStreaming = status === 'streaming';

  useEffect(() => {
    const fetchData = async () => {
      try {
        const token = getAuthToken();
        const matchHeaders: HeadersInit = token ? { Authorization: `Bearer ${token}` } : {};
        const matchRes = await fetch(`/api/matches/${matchId}`, { headers: matchHeaders });
        if (matchRes.ok) {
          const matchData = await matchRes.json();
          setMatch(matchData);

          if (matchData.status === 'interview_scheduled') {
            try {
              const interviewRes = await fetch(`/api/interviews/${matchId}`, { headers: matchHeaders });
              if (interviewRes.ok) setInterview(await interviewRes.json());
            } catch {}
          }

          if (matchData.status === 'offer_sent' || matchData.status === 'offered') {
            try {
              const offerRes = await fetch(`/api/offers/${matchId}`, { headers: matchHeaders });
              if (offerRes.ok) setOffer(await offerRes.json());
            } catch {}
          }
        }
      } catch (err) {
        console.error('Failed to fetch conversation data:', err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [matchId]);

  useEffect(() => {
    if (messages.length > 0) {
      scrollRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
    }
  }, [messages]);

  const handleInterviewConfirm = async () => {
    try {
      const token = getAuthToken();
      await fetch(`/api/interviews/${matchId}/confirm`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      const res = await fetch(`/api/interviews/${matchId}`);
      if (res.ok) setInterview(await res.json());
    } catch (err) {
      console.error('Failed to confirm interview:', err);
    }
  };

  const handleInterviewCancel = async () => {
    try {
      const token = getAuthToken();
      await fetch(`/api/interviews/${matchId}/cancel`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      const res = await fetch(`/api/interviews/${matchId}`);
      if (res.ok) setInterview(await res.json());
    } catch (err) {
      console.error('Failed to cancel interview:', err);
    }
  };

  const handleOfferAccept = async () => {
    try {
      const token = getAuthToken();
      await fetch(`/api/offers/${matchId}/accept`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      const res = await fetch(`/api/offers/${matchId}`);
      if (res.ok) setOffer(await res.json());
    } catch (err) {
      console.error('Failed to accept offer:', err);
    }
  };

  const handleMilestoneConfirm = async () => {
    const token = getAuthToken();
    if (!token) return;
    try {
      await fetch(`/api/matches/${matchId}/confirm`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      setNotification('Milestone confirmed!');
      setTimeout(() => setNotification(null), 3000);
    } catch (err) {
      console.error('Failed to confirm milestone:', err);
    }
  };

  const determineAgentRole = useCallback((msg: { role: string; id?: string; sender_agent_id?: string }): AgentRole => {
    if (msg.role === 'user') return 'seeker';
    if (msg.role === 'system') return 'system';
    // Use match recruiter/seeker agent IDs for accurate role detection
    if (match) {
      const senderId = msg.sender_agent_id || msg.id;
      if (senderId === match.recruiter_agent_id) return 'recruiter';
      if (senderId === match.seeker_agent_id) return 'seeker';
    }
    return 'recruiter';
  }, [match]);

  const getMessageContent = useCallback((msg: { parts?: Array<{ type: string; text?: string }>; role: string; created_at?: string }): string => {
    if (msg.parts) {
      return msg.parts
        .filter((p) => p.type === 'text')
        .map((p) => p.text || '')
        .join('');
    }
    return '';
  }, []);

  const getMessageTimestamp = useCallback((msg: { created_at?: string; id?: string }): Date => {
    if (msg.created_at) {
      return new Date(msg.created_at);
    }
    // Fallback: derive timestamp from message id (contains time component)
    if (msg.id) {
      const timePart = msg.id.split('-').pop();
      if (timePart && timePart.length >= 10) {
        const ts = parseInt(timePart.slice(0, 10), 10);
        if (!isNaN(ts) && ts > 1000000000) {
          return new Date(ts * 1000);
        }
      }
    }
    return new Date();
  }, []);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-500">Loading conversation...</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)]">
      {/* Toast notification */}
      {notification && (
        <div className="fixed top-4 right-4 z-50 flex items-center gap-2 px-4 py-3 bg-green-600 text-white rounded-lg shadow-lg">
          <span className="text-sm font-medium">{notification}</span>
          <button onClick={() => setNotification(null)} className="p-0.5 hover:bg-green-700 rounded">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Header */}
      <ConversationHeader
        matchId={matchId}
        fsmStage={fsmStage}
        isConnected={isConnected}
      />

      {/* Main content - dual column */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left: Messages */}
        <div className={`flex flex-col border-r border-gray-200 transition-all duration-300 ${
          showFlowPanel ? 'w-[65%] lg:w-[65%]' : 'w-full'
        }`}>
          {/* Message list */}
          <div className="flex-1 overflow-y-auto p-4 space-y-1">
            {messages.length === 0 && (
              <div className="flex flex-col items-center justify-center h-full text-gray-400">
                <p className="text-lg font-medium">No messages yet</p>
                <p className="text-sm">Start the conversation below</p>
              </div>
            )}

            {messages.map((msg) => (
              <AgentMessageBubble
                key={msg.id}
                content={getMessageContent(msg)}
                agentRole={determineAgentRole(msg)}
                timestamp={getMessageTimestamp(msg)}
              />
            ))}

            {isStreaming && (
              <div className="flex justify-start">
                <div className="bg-blue-50 border border-blue-200 rounded-2xl px-4 py-3">
                  <div className="flex items-center gap-2">
                    <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                    <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                    <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                    <span className="text-sm text-blue-600">AI is thinking...</span>
                  </div>
                </div>
              </div>
            )}

            <div ref={scrollRef} />
          </div>

          {/* Interview/Offer cards above input */}
          {(interview || offer) && (
            <div className="px-4 py-2 border-t border-gray-100 space-y-2">
              {interview && (
                <InterviewCard
                  interview={interview}
                  onConfirm={handleInterviewConfirm}
                  onCancel={handleInterviewCancel}
                />
              )}
              {offer && (
                <OfferCard
                  offer={offer}
                  onRespond={handleOfferAccept}
                />
              )}
            </div>
          )}

          {/* Input */}
          <TextareaInput
            value={input}
            onChange={setInput}
            onSubmit={handleSubmit}
            onStop={stop}
            onRetry={reload}
            disabled={false}
            isStreaming={isStreaming}
            error={error}
          />
        </div>

        {/* Right: Flow Panel */}
        {showFlowPanel ? (
          <div className="w-[35%] lg:w-[35%] overflow-y-auto p-4 bg-gray-50 hidden lg:block">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Process Flow</h2>
              <button
                onClick={() => setShowFlowPanel(false)}
                className="p-1 text-gray-400 hover:text-gray-600 rounded"
              >
                <PanelRightClose className="w-4 h-4" />
              </button>
            </div>

            <FlowPanel currentStage={fsmStage} />

            {/* Cards in flow panel on desktop */}
            {(interview || offer) && (
              <div className="mt-6 space-y-4">
                {interview && (
                  <InterviewCard
                    interview={interview}
                    onConfirm={handleInterviewConfirm}
                    onCancel={handleInterviewCancel}
                  />
                )}
                {offer && (
                  <OfferCard
                    offer={offer}
                    onRespond={handleOfferAccept}
                  />
                )}
              </div>
            )}
          </div>
        ) : (
          <button
            onClick={() => setShowFlowPanel(true)}
            className="hidden lg:flex items-center gap-1 px-2 py-1 text-gray-400 hover:text-gray-600 border-l border-gray-200"
          >
            <PanelRightOpen className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Mobile flow panel - bottom drawer */}
      <div className="lg:hidden">
        {showFlowPanel && (
          <div className="fixed inset-0 z-40 bg-black/50" onClick={() => setShowFlowPanel(false)}>
            <div
              className="absolute bottom-0 left-0 right-0 max-h-[70vh] bg-white rounded-t-2xl shadow-xl overflow-y-auto p-4"
              onClick={(e) => e.stopPropagation()}
            >
              {/* Drag handle */}
              <div className="flex justify-center mb-2">
                <div className="w-10 h-1 bg-gray-300 rounded-full" />
              </div>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">Process Flow</h2>
                <button
                  onClick={() => setShowFlowPanel(false)}
                  className="p-1 text-gray-400 hover:text-gray-600 rounded"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
              <FlowPanel currentStage={fsmStage} />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
