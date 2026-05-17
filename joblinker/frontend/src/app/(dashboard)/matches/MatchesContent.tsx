'use client'

import { ArrowLeftRight, Sparkles } from 'lucide-react'
import Link from 'next/link'
import { useCallback, useState } from 'react'

import type { Match } from '@/types'

import { Card, EmptyState, ErrorState, LoadingSkeleton } from '@/components/ui'
import { apiClient } from '@/lib/api_client'

interface MatchWithScore extends Match {
  score: number
}

function MatchCard({ match, onConfirm, onDecline }: {
  match: MatchWithScore
  onConfirm?: (id: string) => void
  onDecline?: (id: string) => void
}) {
  const scorePercent = Math.round((match.score || 0) * 100)

  const statusColors: Record<string, string> = {
    mutual_interest: 'bg-green-100 text-green-700',
    pending: 'bg-yellow-100 text-yellow-700',
    rejected: 'bg-slate-200 text-slate-600',
  }

  return (
    <Card hover className="p-5 bg-white/80 backdrop-blur-xl border border-white/20 cursor-pointer group">
      <div className="flex items-center justify-between mb-4">
        <div>
          <p className="font-semibold text-slate-900 flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-blue-500" />
            Match #
            {match.id.slice(0, 8)}
          </p>
          <p className="text-sm text-slate-600 mt-1">
            {match.seeker_agent_id?.slice(0, 8) || 'Seeker'}
            {' '}
            • Job #
            {match.job_id?.slice(0, 8) || 'N/A'}
          </p>
        </div>
        <span className={`px-3 py-1 text-xs font-medium rounded-full ${
          statusColors[match.status] || 'bg-slate-200 text-slate-600'
        }`}
        >
          {match.status.replace('_', ' ')}
        </span>
      </div>

      {/* Match Score Bar */}
      <div className="mb-4">
        <div className="flex items-center justify-between text-sm mb-1">
          <span className="text-slate-600 flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-blue-500" />
            Match Score
          </span>
          <span className="font-medium text-slate-900">
            {scorePercent}
            %
          </span>
        </div>
        <div className="w-full h-2 bg-slate-100 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-blue-500 to-sky-400 rounded-full"
            style={{ width: `${scorePercent}%` }}
          />
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex gap-3">
        {match.status === 'pending' && (
          <>
            <button
              onClick={(e) => {
                e.preventDefault()
                onConfirm?.(match.id)
              }}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm font-medium transition-colors"
            >
              Confirm
            </button>
            <button
              onClick={(e) => {
                e.preventDefault()
                onDecline?.(match.id)
              }}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 text-sm font-medium transition-colors"
            >
              Decline
            </button>
          </>
        )}
        <Link href={`/conversation/${match.id}`} className="flex-1">
          <button className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-blue-50 text-blue-700 rounded-lg hover:bg-blue-100 text-sm font-medium transition-colors">
            View Conversation
          </button>
        </Link>
      </div>
    </Card>
  )
}

export function MatchesContent({ initialMatches }: { initialMatches: MatchWithScore[] }) {
  const [matches, setMatches] = useState<MatchWithScore[]>(initialMatches)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchMatches = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await apiClient.get<MatchWithScore[]>('/api/matches')
      setMatches(data || [])
    }
    catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load matches')
    }
    finally {
      setIsLoading(false)
    }
  }, [])

  const handleConfirm = async (matchId: string) => {
    try {
      await apiClient.post(`/api/matches/${matchId}/confirm`)
      fetchMatches()
    }
    catch (err) {
      console.error('Failed to confirm match:', err)
    }
  }

  const handleDecline = async (matchId: string) => {
    try {
      await apiClient.post(`/api/matches/${matchId}/decline`)
      fetchMatches()
    }
    catch (err) {
      console.error('Failed to decline match:', err)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3 mb-6">
          <ArrowLeftRight className="w-8 h-8 text-blue-600" />
          Matches
        </h1>

        {isLoading
          ? (
              <LoadingSkeleton count={3} variant="card" />
            )
          : error
            ? (
                <ErrorState message={error} onRetry={fetchMatches} />
              )
            : matches.length === 0
              ? (
                  <EmptyState
                    icon={<Sparkles className="w-10 h-10 text-blue-600" />}
                    title="No matches yet"
                    description="Create an agent and add jobs or resumes to start matching"
                  />
                )
              : (
                  <div className="space-y-4">
                    {matches.map(match => (
                      <MatchCard
                        key={match.id}
                        match={match}
                        onConfirm={handleConfirm}
                        onDecline={handleDecline}
                      />
                    ))}
                  </div>
                )}
      </div>
    </div>
  )
}
