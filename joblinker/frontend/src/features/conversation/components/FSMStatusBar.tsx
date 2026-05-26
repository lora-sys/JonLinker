'use client'

import { CheckCircle, Circle, Loader2 } from 'lucide-react'

import type { FSMStage } from '@/features/conversation/hooks/useAIChat'

const STAGES: { key: FSMStage, label: string }[] = [
  { key: 'INTRODUCTION', label: 'Introduction' },
  { key: 'NEGOTIATION', label: 'Negotiation' },
  { key: 'INTERVIEW', label: 'Interview' },
  { key: 'OFFER', label: 'Offer' },
  { key: 'COMPLETED', label: 'Completed' },
]

const STAGE_ORDER = STAGES.map(s => s.key)

interface FSMStatusBarProps {
  currentStage: FSMStage
  className?: string
}

export function FSMStatusBar({ currentStage, className = '' }: FSMStatusBarProps) {
  const currentIdx = STAGE_ORDER.indexOf(currentStage)

  return (
    <div className={`px-4 py-2 bg-white border-b border-gray-200 ${className}`}>
      <div className="flex items-center gap-0 max-w-2xl mx-auto">
        {STAGES.map((stage, idx) => {
          const isCompleted = idx < currentIdx
          const isCurrent = idx === currentIdx
          const isFuture = idx > currentIdx

          return (
            <div key={stage.key} className="flex-1 flex items-center">
              <div className="flex items-center gap-1.5">
                {isCompleted && <CheckCircle className="w-4 h-4 text-green-500 shrink-0" />}
                {isCurrent && <Loader2 className="w-4 h-4 text-blue-500 animate-spin shrink-0" />}
                {isFuture && <Circle className="w-4 h-4 text-gray-300 shrink-0" />}
                <span className={`text-[11px] leading-tight hidden sm:inline ${
                  isCompleted ? 'text-green-700' : isCurrent ? 'text-blue-700 font-medium' : 'text-gray-400'
                }`}
                >
                  {stage.label}
                </span>
              </div>
              {idx < STAGES.length - 1 && (
                <div className={`flex-1 h-0.5 mx-2 ${
                  isCompleted ? 'bg-green-400' : isCurrent ? 'bg-blue-200' : 'bg-gray-200'
                }`}
                />
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
