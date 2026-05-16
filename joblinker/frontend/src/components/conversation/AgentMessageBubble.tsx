'use client'

import { Bot, User } from 'lucide-react'

export type AgentRole = 'recruiter' | 'seeker' | 'system'

interface AgentMessageBubbleProps {
  content: string
  agentRole: AgentRole
  agentName?: string
  timestamp?: Date | string
  intent?: string
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

const intentColors: Record<string, string> = {
  INTRODUCTION: 'bg-blue-100 text-blue-700',
  INTEREST: 'bg-green-100 text-green-700',
  NEGOTIATION: 'bg-amber-100 text-amber-700',
  OFFER: 'bg-purple-100 text-purple-700',
  ACCEPT: 'bg-green-100 text-green-700',
  DECLINE: 'bg-red-100 text-red-700',
  SCHEDULE: 'bg-blue-100 text-blue-700',
  INQUIRY: 'bg-gray-100 text-gray-600',
}

export function AgentMessageBubble({
  content,
  agentRole,
  agentName,
  timestamp,
  intent,
}: AgentMessageBubbleProps) {
  const style = roleStyles[agentRole]
  const isSystem = agentRole === 'system'

  const formatTime = (ts: Date | string) => {
    try {
      const d = typeof ts === 'string' ? new Date(ts) : ts
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
    catch {
      return ''
    }
  }

  if (isSystem) {
    return (
      <div className="flex justify-center my-2">
        <div className="px-3 py-1.5 bg-gray-100 rounded-full text-xs text-gray-500">
          {content}
        </div>
      </div>
    )
  }

  return (
    <div className={`flex ${style.align === 'self-end' ? 'justify-end' : 'justify-start'} mb-3`}>
      <div className={`max-w-[75%] flex gap-2 ${style.align === 'self-end' ? 'flex-row-reverse' : 'flex-row'}`}>
        {/* Avatar */}
        <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${
          agentRole === 'recruiter' ? 'bg-emerald-200' : 'bg-blue-200'
        }`}
        >
          {agentRole === 'recruiter'
            ? (
                <User className={`w-4 h-4 ${style.accent}`} />
              )
            : (
                <Bot className={`w-4 h-4 ${style.accent}`} />
              )}
        </div>

        {/* Bubble */}
        <div className={`px-4 py-2.5 rounded-2xl border ${style.bg} ${style.border}`}>
          {/* Header */}
          <div className="flex items-center gap-2 mb-1">
            <span className={`text-xs font-semibold ${style.accent}`}>
              {agentName || (agentRole === 'recruiter' ? 'Recruiter Agent' : 'Seeker Agent')}
            </span>
            {timestamp && (
              <span className="text-xs text-gray-400">{formatTime(timestamp)}</span>
            )}
          </div>

          {/* Content */}
          <p className="text-sm text-gray-800 whitespace-pre-wrap">{content}</p>

          {/* Intent badge */}
          {intent && (
            <div className="mt-1.5">
              <span className={`inline-block px-2 py-0.5 text-xs font-medium rounded ${
                intentColors[intent] || 'bg-gray-100 text-gray-600'
              }`}
              >
                {intent}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
