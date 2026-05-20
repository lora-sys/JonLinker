'use client'

import { ArrowLeft, Loader, Wifi, WifiOff } from 'lucide-react'
import Link from 'next/link'

import type { WSConnectionStatus } from '@/hooks/useWebSocket'

import type { FSMStage } from './FlowPanel'

interface ConversationHeaderProps {
  matchId: string
  recruiterName?: string
  seekerName?: string
  fsmStage?: FSMStage
  wsStatus?: WSConnectionStatus
}

const stageColors: Record<string, string> = {
  INTRODUCTION: 'bg-blue-100 text-blue-700',
  NEGOTIATION: 'bg-amber-100 text-amber-700',
  INTERVIEW: 'bg-purple-100 text-purple-700',
  OFFER: 'bg-green-100 text-green-700',
  COMPLETED: 'bg-gray-100 text-gray-700',
}

export function ConversationHeader({
  recruiterName = 'Recruiter Agent',
  seekerName = 'Seeker Agent',
  fsmStage = 'INTRODUCTION',
  wsStatus = 'Disconnected',
}: ConversationHeaderProps) {
  return (
    <header className="h-14 border-b border-gray-200 bg-white flex items-center px-4 gap-4 shrink-0">
      {/* Back button */}
      <Link
        href="/matches"
        className="p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
      >
        <ArrowLeft className="w-5 h-5" />
      </Link>

      {/* Agent names */}
      <div className="flex items-center gap-2 flex-1 min-w-0">
        <span className="text-sm font-medium text-emerald-700 truncate">{recruiterName}</span>
        <span className="text-gray-300">&mdash;</span>
        <span className="text-sm font-medium text-blue-700 truncate">{seekerName}</span>
      </div>

      {/* FSM Badge */}
      <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
        stageColors[fsmStage] || 'bg-gray-100 text-gray-700'
      }`}
      >
        {fsmStage.replace(/_/g, ' ')}
      </span>

      {/* WS Status */}
      <div className="flex items-center gap-1.5">
        {wsStatus === 'Connected'
          ? (
              <Wifi className="w-4 h-4 text-green-500" />
            )
          : wsStatus === 'Reconnecting' || wsStatus === 'Connecting'
            ? (
                <Loader className="w-4 h-4 text-amber-500 animate-spin" />
              )
            : (
                <WifiOff className="w-4 h-4 text-gray-400" />
              )}
        <span className={`text-xs ${
          wsStatus === 'Connected'
            ? 'text-green-600'
            : wsStatus === 'Reconnecting' || wsStatus === 'Connecting'
              ? 'text-amber-600'
              : 'text-gray-400'
        }`}
        >
          {wsStatus === 'Connected'
            ? 'Live'
            : wsStatus === 'Reconnecting' || wsStatus === 'Connecting'
              ? 'Reconnecting'
              : 'Offline'}
        </span>
      </div>
    </header>
  )
}
