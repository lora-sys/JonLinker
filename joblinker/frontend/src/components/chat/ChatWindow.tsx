'use client';

import { useState } from 'react';
import { useChat } from '@/hooks/useChat';

interface ChatWindowProps {
  matchId: string;
  onConfirmMilestone?: () => void;
}

export function ChatWindow({ matchId, onConfirmMilestone }: ChatWindowProps) {
  const { messages, isConnected, isLoading, error, sendMessageREST } = useChat({ matchId });
  const [inputValue, setInputValue] = useState('');
  const [isSending, setIsSending] = useState(false);

  const handleSend = async () => {
    if (!inputValue.trim() || isSending) return;

    setIsSending(true);
    try {
      const intentType = detectIntent(inputValue);
      await sendMessageREST(inputValue, intentType);
      setInputValue('');
    } catch (err) {
      console.error('Failed to send message:', err);
    } finally {
      setIsSending(false);
    }
  };

  const detectIntent = (text: string): string => {
    const lower = text.toLowerCase();
    if (lower.includes('accept') || lower.includes('yes')) return 'ACCEPT';
    if (lower.includes('decline') || lower.includes('no')) return 'DECLINE';
    if (lower.includes('offer')) return 'OFFER';
    if (lower.includes('salary') || lower.includes('compensat')) return 'NEGOTIATION';
    if (lower.includes('schedule') || lower.includes('interview')) return 'SCHEDULE';
    return 'INQUIRY';
  };

  const parseMessageContent = (contentXml: string): string => {
    try {
      // Simple XML parsing for display
      if (contentXml.includes('<message>')) {
        const intentMatch = contentXml.match(/intent="([^"]+)"/);
        const textMatch = contentXml.match(/>([^<]+)</);
        if (intentMatch) {
          return `[${intentMatch[1]}] ${textMatch?.[1] || ''}`;
        }
      }
      return contentXml;
    } catch {
      return contentXml;
    }
  };

  return (
    <div className="flex flex-col h-full bg-white rounded-xl shadow-sm border border-gray-200">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200">
        <h3 className="font-semibold text-gray-900">Conversation</h3>
        <span className={`flex items-center gap-1.5 px-2 py-1 rounded-full text-xs ${
          isConnected
            ? 'bg-green-100 text-green-700'
            : 'bg-gray-100 text-gray-500'
        }`}>
          <span className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-400'}`} />
          {isConnected ? 'Connected' : 'Disconnected'}
        </span>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-gray-500">Loading messages...</div>
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-500">
            <svg className="w-12 h-12 mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <p>No messages yet</p>
            <p className="text-sm">Start the conversation below</p>
          </div>
        ) : (
          messages.map((msg) => (
            <MessageBubble
              key={msg.id}
              content={parseMessageContent(msg.content_xml)}
              intent={msg.intent_type}
              timestamp={msg.created_at}
            />
          ))
        )}
      </div>

      {/* Input */}
      <div className="p-4 border-t border-gray-200">
        {error && (
          <div className="mb-3 p-2 bg-red-50 text-red-700 text-sm rounded-lg">
            {error}
          </div>
        )}
        <div className="flex gap-2">
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSend()}
            placeholder="Type a message..."
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            disabled={isSending}
          />
          <button
            onClick={handleSend}
            disabled={!inputValue.trim() || isSending}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
          >
            {isSending ? 'Sending...' : 'Send'}
          </button>
        </div>

        {/* Milestone confirmation button */}
        {onConfirmMilestone && messages.some(m =>
          ['NEGOTIATION', 'OFFER'].includes(m.intent_type)
        ) && (
          <button
            onClick={onConfirmMilestone}
            className="mt-3 w-full px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 font-medium"
          >
            Confirm Milestone
          </button>
        )}
      </div>
    </div>
  );
}

interface MessageBubbleProps {
  content: string;
  intent: string;
  timestamp: string;
}

function MessageBubble({ content, intent, timestamp }: MessageBubbleProps) {
  const formatTime = (ts: string) => {
    try {
      return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  const intentColors: Record<string, string> = {
    INTRODUCTION: 'bg-blue-50 border-blue-200',
    INTEREST: 'bg-green-50 border-green-200',
    NEGOTIATION: 'bg-yellow-50 border-yellow-200',
    OFFER: 'bg-purple-50 border-purple-200',
    ACCEPT: 'bg-green-100 border-green-300',
    DECLINE: 'bg-red-50 border-red-200',
    SCHEDULE: 'bg-blue-50 border-blue-200',
    CONFIRM: 'bg-green-100 border-green-300',
  };

  return (
    <div className={`p-3 rounded-lg border ${intentColors[intent] || 'bg-gray-50 border-gray-200'}`}>
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1">
          <p className="text-gray-900">{content}</p>
          {intent && (
            <span className="inline-block mt-1 px-2 py-0.5 text-xs font-medium bg-white/50 rounded">
              {intent}
            </span>
          )}
        </div>
        <span className="text-xs text-gray-400">{formatTime(timestamp)}</span>
      </div>
    </div>
  );
}