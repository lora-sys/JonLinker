'use client'

import { Ban, Check, X } from 'lucide-react'

import type { PendingConfirm } from '@/hooks/useAIChat'

interface HumanConfirmModalProps {
  pendingConfirm: PendingConfirm
  onConfirm: (approved: boolean) => void
  onDismiss?: () => void
}

export function HumanConfirmModal({ pendingConfirm, onConfirm, onDismiss }: HumanConfirmModalProps) {
  const intentLabel = pendingConfirm.intent === 'OFFER' ? 'Offer' : 'Schedule'

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40">
      <div className="bg-white rounded-2xl shadow-xl max-w-md w-full p-6 relative animate-in fade-in zoom-in-95 duration-200">
        {onDismiss && (
          <button
            onClick={onDismiss}
            className="absolute top-3 right-3 p-1 text-gray-400 hover:text-gray-600 rounded"
          >
            <X className="w-4 h-4" />
          </button>
        )}

        <div className="text-center mb-6">
          <div className="w-12 h-12 mx-auto mb-3 rounded-full bg-amber-100 flex items-center justify-center">
            <span className="text-xl">🤖</span>
          </div>
          <h2 className="text-lg font-semibold text-gray-900">Human Confirmation Required</h2>
          <p className="text-sm text-gray-500 mt-1">
            The agent has generated a
            {' '}
            <span className="font-medium text-gray-700">{intentLabel}</span>
            . Please review and confirm.
          </p>
        </div>

        {pendingConfirm.content_xml && (
          <div className="mb-4 p-3 bg-gray-50 rounded-lg text-sm text-gray-600 max-h-24 overflow-y-auto">
            {pendingConfirm.content_xml}
          </div>
        )}

        <div className="flex gap-3">
          <button
            onClick={() => onConfirm(false)}
            className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 border border-gray-300 text-gray-700 rounded-xl hover:bg-gray-50 font-medium transition-colors"
          >
            <Ban className="w-4 h-4" />
            Reject
          </button>
          <button
            onClick={() => onConfirm(true)}
            className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-600 text-white rounded-xl hover:bg-blue-700 font-medium transition-colors"
          >
            <Check className="w-4 h-4" />
            Approve
          </button>
        </div>
      </div>
    </div>
  )
}
