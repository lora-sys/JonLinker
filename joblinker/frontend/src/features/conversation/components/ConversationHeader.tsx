'use client'

import { ArrowLeft, Loader, RefreshCw, Wifi, WifiOff } from 'lucide-react'
import Link from 'next/link'

import type { WSConnectionStatus } from '@/shared/hooks/useWebSocket'

import type { FSMStage } from './FlowPanel'

interface ConversationHeaderProps {
  matchId: string
  recruiterName?: string
  seekerName?: string
  fsmStage?: FSMStage
  wsStatus?: WSConnectionStatus
  sessionVersion?: number
  sessionStatus?: 'active' | 'concluded'
  onReopen?: () => void
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
  sessionVersion,
  sessionStatus,
  onReopen,
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

      {/* Session Version Badge */}
      {sessionVersion != null && (
        <span className="px-2 py-1 rounded text-xs font-mono bg-gray-100 text-gray-500">
          v
          {sessionVersion}
        </span>
      )}

      {/* Session Status */}
      {sessionStatus === 'concluded' && (
        <span className="px-2 py-1 rounded text-xs font-mono bg-orange-100 text-orange-600">
          concluded
        </span>
      )}

      {/* Reopen Button */}
      {sessionStatus === 'concluded' && onReopen && (
        <button
          onClick={onReopen}
          className="flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-blue-50 text-blue-600 hover:bg-blue-100 transition-colors"
        >
          <RefreshCw className="w-3 h-3" />
          Reopen
        </button>
      )}

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
