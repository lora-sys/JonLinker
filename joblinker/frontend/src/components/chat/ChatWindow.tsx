'use client';

import { useCallback } from 'react';
import { Send, Square, RotateCcw } from 'lucide-react';
import { useAIChat } from '@/hooks/useAIChat';
import { MessageList } from './MessageList';

interface ChatWindowProps {
  matchId: string;
  onConfirmMilestone?: () => void;
}

export function ChatWindow({ matchId, onConfirmMilestone }: ChatWindowProps) {
  const {
    messages: rawMessages,
    input,
    setInput,
    status,
    isConnected,
    error,
    handleSubmit,
    stop,
    reload,
    loadMoreMessages,
    hasMoreMessages,
  } = useAIChat({ matchId });

  // Convert AI SDK messages to ChatMessage format for MessageList
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const messages = rawMessages.map((m: any) => {
    const text = m.parts?.filter((p: any) => p.type === 'text').map((p: any) => p.text || '').join('') || '';
    return {
      id: m.id,
      matchId,
      role: m.role as 'user' | 'assistant' | 'system',
      content: text,
      createdAt: new Date(),
      status: 'done' as const,
    };
  });

  const isStreaming = status === 'streaming';
  const isDisabled = isStreaming;

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSubmit();
      }
      if (e.key === 'Escape' && isStreaming) {
        e.preventDefault();
        stop();
      }
    },
    [handleSubmit, isStreaming, stop]
  );

  const showMilestone = onConfirmMilestone && messages.some(
    (m) => m.content.toLowerCase().includes('negotiation') || m.content.toLowerCase().includes('offer')
  );

  return (
    <div className="flex flex-col h-full bg-white rounded-xl shadow-sm border border-gray-200">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200">
        <h3 className="font-semibold text-gray-900">Conversation</h3>
        <div className="flex items-center gap-2">
          {isStreaming && (
            <button
              onClick={stop}
              className="flex items-center gap-1 px-2 py-1 text-xs font-medium text-red-600 bg-red-50 rounded-md hover:bg-red-100"
              aria-label="Stop generating"
            >
              <Square className="w-3 h-3" />
              Stop
            </button>
          )}
          <span
            className={`flex items-center gap-1.5 px-2 py-1 rounded-full text-xs ${
              isConnected
                ? 'bg-green-100 text-green-700'
                : 'bg-gray-100 text-gray-500'
            }`}
          >
            <span
              className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-400'}`}
              aria-hidden="true"
            />
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
      </div>

      {/* Messages */}
      <MessageList
        messages={messages}
        status={status}
        isLoading={messages.length === 0 && isStreaming}
        onLoadMore={loadMoreMessages}
        hasMore={hasMoreMessages}
      />

      {/* Input */}
      <div className="p-4 border-t border-gray-200">
        {error && (
          <div className="mb-3 p-2 bg-red-50 text-red-700 text-sm rounded-lg flex items-center justify-between">
            <span>{error}</span>
            <button
              onClick={reload}
              className="flex items-center gap-1 px-2 py-1 text-xs font-medium text-red-700 bg-red-100 rounded-md hover:bg-red-200"
            >
              <RotateCcw className="w-3 h-3" />
              Retry
            </button>
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a message..."
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50"
            disabled={isDisabled}
            aria-label="Message input"
          />
          <button
            type="submit"
            disabled={!input.trim() || isDisabled}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium flex items-center gap-1.5"
            aria-label="Send message"
          >
            <Send className="w-4 h-4" />
            Send
          </button>
        </form>

        {/* Milestone confirmation */}
        {showMilestone && (
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
