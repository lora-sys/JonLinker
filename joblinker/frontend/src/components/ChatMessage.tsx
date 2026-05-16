'use client'

import type { TextUIPart, ToolUIPart, UIMessage } from 'ai'

import { Bot, User } from 'lucide-react'

export type AgentRole = 'recruiter' | 'seeker' | 'system'

interface ChatMessageProps {
  message: UIMessage
  agentRole?: AgentRole
  agentName?: string
  timestamp?: Date
  isLoading?: boolean
}

const roleStyles: Record<AgentRole, { bg: string, border: string, align: string, accent: string }> = {
  recruiter: {
    bg: 'bg-emerald-50',
    border: 'border-emerald-200',
    align: 'self-end',
    accent: 'text-emerald-700',
  },
  seeker: {
    bg: 'bg-blue-50',
    border: 'border-blue-200',
    align: 'self-start',
    accent: 'text-blue-700',
  },
  system: {
    bg: 'bg-gray-50',
    border: 'border-gray-200',
    align: 'self-center',
    accent: 'text-gray-600',
  },
}

function formatTime(date: Date) {
  try {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  catch {
    return ''
  }
}

function getTextContent(msg: UIMessage): string {
  if (!msg.parts)
    return ''
  return msg.parts
    .filter((p): p is TextUIPart => p.type === 'text')
    .map(p => p.text)
    .join('')
}

function getToolCalls(msg: UIMessage): Array<{ id: string, name: string }> {
  if (!msg.parts)
    return []
  return msg.parts
    .filter((p): p is ToolUIPart => p.type === 'tool-call' || p.type === 'tool-result')
    .map(p => ({
      id: 'toolCallId' in p ? (p as { toolCallId: string }).toolCallId : String(msg.id),
      name: 'toolName' in p ? String((p as { toolName: string }).toolName) : p.type,
    }))
}

export function ChatMessage({ message, agentRole = 'recruiter', agentName, timestamp, isLoading }: ChatMessageProps) {
  const role = message.role === 'user' ? 'seeker' : agentRole
  const style = roleStyles[role] || roleStyles.recruiter
  const textContent = getTextContent(message)
  const toolCalls = getToolCalls(message)

  if (message.role === 'system') {
    return (
      <div className="flex justify-center my-2">
        <div className="px-3 py-1.5 bg-gray-100 rounded-full text-xs text-gray-500 max-w-[90%] text-center">
          {textContent || `[System: ${message.id}]`}
        </div>
      </div>
    )
  }

  return (
    <div className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'} mb-3`}>
      <div className={`max-w-[75%] flex gap-2 ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'}`}>
        <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${
          role === 'recruiter' ? 'bg-emerald-200' : 'bg-blue-200'
        }`}
        >
          {role === 'recruiter'
            ? (
                <User className={`w-4 h-4 ${style.accent}`} />
              )
            : (
                <Bot className={`w-4 h-4 ${style.accent}`} />
              )}
        </div>

        <div className={`px-4 py-2.5 rounded-2xl border ${style.bg} ${style.border}`}>
          <div className="flex items-center gap-2 mb-1">
            <span className={`text-xs font-semibold ${style.accent}`}>
              {agentName || (role === 'recruiter' ? 'Recruiter Agent' : message.role === 'user' ? 'You' : 'Seeker Agent')}
            </span>
            {timestamp && (
              <span className="text-xs text-gray-400">{formatTime(timestamp)}</span>
            )}
          </div>

          <p className="text-sm text-gray-800 whitespace-pre-wrap">
            {textContent}
            {isLoading && <span className="inline-block w-1.5 h-4 ml-0.5 bg-gray-400 animate-pulse" />}
          </p>

          {toolCalls.map((tc, i) => (
            <div key={tc.id || i} className="mt-2 p-2 bg-gray-100 rounded text-xs text-gray-600">
              <span className="font-medium">Tool:</span>
              {' '}
              {tc.name}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
