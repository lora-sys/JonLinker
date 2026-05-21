'use client'

import { ChevronDown, ChevronRight, FileText } from 'lucide-react'
import { useEffect, useState } from 'react'

import { apiClient } from '@/lib/api_client'

interface SummaryPanelProps {
  matchId: string
}

export function SummaryPanel({ matchId }: SummaryPanelProps) {
  const [open, setOpen] = useState(false)
  const [summary, setSummary] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!open || summary !== null || loading)
      return
    setLoading(true)
    apiClient.get<{ summary_text?: string; summary?: string }>(`/api/sessions/${matchId}/summary`)
      .then((data) => {
        setSummary(data.summary_text || data.summary || null)
      })
      .catch(() => {
        setSummary(null)
      })
      .finally(() => setLoading(false))
  }, [open, matchId, summary, loading])

  return (
    <div className="border border-gray-200 rounded-lg">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 w-full px-3 py-2 text-left text-sm font-medium text-gray-600 hover:bg-gray-50 rounded-lg transition-colors"
      >
        {open ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
        <FileText className="w-4 h-4" />
        Session Summary
      </button>
      {open && (
        <div className="px-3 pb-3">
          <div className="border-t border-gray-100 pt-2">
            {loading
              ? (
                  <p className="text-sm text-gray-400">Loading...</p>
                )
              : summary
                ? (
                    <p className="text-sm text-gray-700 whitespace-pre-wrap">{summary}</p>
                  )
                : (
                    <p className="text-sm text-gray-400 italic">No summary available</p>
                  )}
          </div>
        </div>
      )}
    </div>
  )
}
