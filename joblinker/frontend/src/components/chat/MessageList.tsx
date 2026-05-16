'use client'

import { useEffect, useRef } from 'react'

import type { ChatMessage, MessageStatus } from '@/types/ai'

import { StreamingText } from './StreamingText'
import { ToolCallIndicator } from './ToolCallIndicator'

interface MessageListProps {
  messages: ChatMessage[]
  status: MessageStatus
  isLoading?: boolean
  emptyMessage?: string
  onLoadMore?: () => void
  hasMore?: boolean
}

export function MessageList({ messages, status, isLoading, emptyMessage = 'No messages yet', onLoadMore, hasMore }: MessageListProps) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const lastMsgCount = useRef(messages.length)

  useEffect(() => {
    if (messages.length !== lastMsgCount.current || status === 'streaming') {
      scrollRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' })
      lastMsgCount.current = messages.length
    }
  }, [messages, status])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-500">Loading messages...</div>
      </div>
    )
  }

  if (messages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-gray-500">
        <svg className="w-12 h-12 mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
        <p>{emptyMessage}</p>
        <p className="text-sm">Start the conversation below</p>
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {hasMore && onLoadMore && (
        <div className="flex justify-center py-2">
          <button
            onClick={onLoadMore}
            className="px-4 py-1.5 text-sm text-blue-600 bg-blue-50 rounded-full hover:bg-blue-100 font-medium"
          >
            Load older messages
          </button>
        </div>
      )}
      {messages.map(msg => (
        <MessageItem key={msg.id} message={msg} isStreaming={status === 'streaming' && msg.role === 'assistant'} />
      ))}
      {status === 'pending' && <PendingIndicator />}
      <div ref={scrollRef} />
    </div>
  )
}

function MessageItem({ message, isStreaming }: { message: ChatMessage, isStreaming: boolean }) {
  const isUser = message.role === 'user'
  const isSystem = message.role === 'system'

  return (
    <div
      className={`flex ${isUser ? 'justify-end' : isSystem ? 'justify-center' : 'justify-start'}`}
      tabIndex={0}
      role="listitem"
      aria-label={`${message.role} message`}
    >
      <div
        className={`max-w-[80%] rounded-2xl px-4 py-3 ${
          isUser
            ? 'bg-blue-600 text-white rounded-br-md'
            : isSystem
              ? 'bg-gray-100 text-gray-600 text-sm rounded-full'
              : 'bg-white border border-gray-200 text-gray-900 rounded-bl-md shadow-sm'
        }`}
      >
        {isStreaming && message.status !== 'done'
          ? (
              <StreamingText content={message.content} speed={30} />
            )
          : (
              <p className="whitespace-pre-wrap break-words">{message.content}</p>
            )}

        {message.toolInvocations && message.toolInvocations.length > 0 && (
          <div className="mt-2 space-y-1">
            {message.toolInvocations.map(tool => (
              <ToolCallIndicator key={tool.id} tool={tool} />
            ))}
          </div>
        )}

        <span className={`text-xs mt-1 block ${isUser ? 'text-blue-200' : 'text-gray-400'}`}>
          {message.createdAt.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
        </span>
      </div>
    </div>
  )
}

function PendingIndicator() {
  return (
    <div className="flex justify-start">
      <div className="bg-white border border-gray-200 rounded-2xl rounded-bl-md px-4 py-3 shadow-sm">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
          <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
          <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
          <span className="text-sm text-gray-500">AI is thinking...</span>
        </div>
      </div>
    </div>
  )
}
