'use client'

import { motion } from 'framer-motion'
import { Building2, Calendar, CheckCircle2, Clock, Phone, Video, X } from 'lucide-react'
import { useCallback, useState } from 'react'

import type { Interview } from '@/types'

import { EmptyState, ErrorState, LoadingSkeleton, RevealSection } from '@/components/ui'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import { apiClient } from '@/lib/api_client'

const typeIcons: Record<string, React.ReactNode> = {
  video: <Video className="w-5 h-5" />,
  phone: <Phone className="w-5 h-5" />,
  onsite: <Building2 className="w-5 h-5" />,
}

const defaultColors = { bg: 'bg-blue-500', border: 'border-blue-300', icon: 'text-blue-600' }

const typeColors: Record<string, { bg: string, border: string, icon: string }> = {
  video: { bg: 'bg-blue-500', border: 'border-blue-300', icon: 'text-blue-600' },
  phone: { bg: 'bg-green-500', border: 'border-green-300', icon: 'text-green-600' },
  onsite: { bg: 'bg-purple-500', border: 'border-purple-300', icon: 'text-purple-600' },
}

function InterviewCardInner({ interview, index, onUpdate }: { interview: Interview, index: number, onUpdate: () => void }) {
  const [isUpdating, setIsUpdating] = useState(false)
  const colors: { bg: string, border: string, icon: string } = typeColors[interview.format] ?? defaultColors

  const handleConfirm = async () => {
    try {
      setIsUpdating(true)
      await apiClient.patch(`/api/interviews/${interview.id}`, { status: 'confirmed' })
      onUpdate()
    }
    catch (err) {
      console.error('Failed to confirm interview:', err)
    }
    finally {
      setIsUpdating(false)
    }
  }

  const handleCancel = async () => {
    try {
      setIsUpdating(true)
      await apiClient.patch(`/api/interviews/${interview.id}`, { status: 'cancelled' })
      onUpdate()
    }
    catch (err) {
      console.error('Failed to cancel interview:', err)
    }
    finally {
      setIsUpdating(false)
    }
  }

  const scheduledDate = new Date(interview.scheduled_at)

  return (
    <div className="relative flex gap-4">
      {index < 1 && (
        <div className="absolute left-6 top-14 bottom-0 w-0.5 bg-gradient-to-b from-slate-200 to-slate-100" />
      )}

      <motion.div
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ duration: 0.3, delay: index * 0.1 }}
        className={`relative z-10 w-12 h-12 rounded-full flex items-center justify-center ${colors.bg} shadow-lg`}
      >
        <div className="text-white">
          {typeIcons[interview.format]}
        </div>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ duration: 0.4, delay: index * 0.15 }}
        className="flex-1"
      >
        <Card hover className="p-5 bg-white/80 backdrop-blur-xl border border-white/20 hover:shadow-lg transition-all duration-300">
          <div className="flex items-start justify-between mb-3">
            <div>
              <h3 className="font-semibold text-slate-900 text-lg">{interview.match?.job?.structured?.title || 'Interview'}</h3>
              <p className="text-sm text-slate-600 mt-0.5">{interview.match?.job?.structured?.title || 'Position'}</p>
            </div>
            <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${
              interview.status === 'scheduled'
                ? 'bg-blue-100 text-blue-700 border-blue-200'
                : interview.status === 'completed'
                  ? 'bg-green-100 text-green-700 border-green-200'
                  : 'bg-slate-100 text-slate-600 border-slate-200'
            }`}
            >
              {interview.status}
            </span>
          </div>

          <div className="flex items-center gap-2 text-sm text-slate-500 mb-4 p-3 bg-slate-50/80 rounded-xl">
            <Clock className="w-4 h-4" />
            <span className="font-medium">Scheduled for</span>
            <span>
              {scheduledDate.toLocaleDateString()}
              {' '}
              at
              {' '}
              {scheduledDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          </div>

          <div className="flex items-center justify-between">
            <span className={`inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium rounded-full bg-slate-100 ${colors.icon}`}>
              {typeIcons[interview.format]}
              <span className="capitalize">{interview.format}</span>
              Interview
            </span>

            {interview.status === 'scheduled' && (
              <div className="flex gap-2">
                <Button variant="secondary" size="sm" onClick={handleCancel} disabled={isUpdating}>
                  <X className="w-4 h-4 mr-1" />
                  Decline
                </Button>
                <Button variant="primary" size="sm" onClick={handleConfirm} disabled={isUpdating}>
                  <CheckCircle2 className="w-4 h-4 mr-1" />
                  Accept
                </Button>
              </div>
            )}
          </div>
        </Card>
      </motion.div>
    </div>
  )
}

export function InterviewsContent({ initialInterviews }: { initialInterviews: Interview[] }) {
  const [interviews, setInterviews] = useState<Interview[]>(initialInterviews)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchInterviews = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await apiClient.get<{ interviews: Interview[] }>('/api/interviews')
      setInterviews(data?.interviews || [])
    }
    catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load interviews')
    }
    finally {
      setIsLoading(false)
    }
  }, [])

  const scheduledInterviews = interviews.filter(i => i.status === 'scheduled')
  const completedInterviews = interviews.filter(i => i.status === 'completed')

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <RevealSection>
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-4xl font-bold text-slate-900 flex items-center gap-3">
                <Calendar className="w-9 h-9 text-blue-600" />
                Interviews
              </h1>
              <p className="text-slate-600 mt-1">Track your upcoming interviews</p>
            </div>
            <div className="flex items-center gap-3">
              <div className="px-4 py-2 bg-blue-50 rounded-xl border border-blue-100">
                <span className="text-sm text-blue-600 font-medium">
                  {scheduledInterviews.length}
                  {' '}
                  upcoming
                </span>
              </div>
            </div>
          </div>
        </RevealSection>

        {isLoading
          ? (
              <LoadingSkeleton count={3} variant="card" />
            )
          : error
            ? (
                <ErrorState message={error} onRetry={fetchInterviews} />
              )
            : scheduledInterviews.length === 0 && completedInterviews.length === 0
              ? (
                  <EmptyState
                    icon={<Calendar className="w-10 h-10 text-blue-600" />}
                    title="No interviews scheduled"
                    description="Interviews will appear here once matches progress to that stage"
                  />
                )
              : (
                  <>
                    {scheduledInterviews.length > 0 && (
                      <div className="space-y-6">
                        {scheduledInterviews.map((interview, index) => (
                          <InterviewCardInner key={interview.id} interview={interview} index={index} onUpdate={fetchInterviews} />
                        ))}
                      </div>
                    )}

                    {completedInterviews.length > 0 && (
                      <RevealSection delay={3}>
                        <div className="mt-16">
                          <h2 className="text-xl font-semibold text-slate-900 mb-6 flex items-center gap-2">
                            <CheckCircle2 className="w-5 h-5 text-green-600" />
                            Completed Interviews
                          </h2>
                          <div className="space-y-4">
                            {completedInterviews.map((interview, index) => (
                              <motion.div
                                key={interview.id}
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                transition={{ delay: index * 0.1 }}
                                className="flex items-center gap-4 p-4 bg-white/60 backdrop-blur rounded-xl border border-white/50 opacity-60"
                              >
                                <div className={`w-10 h-10 rounded-full flex items-center justify-center ${typeColors[interview.format]?.bg || 'bg-slate-400'} opacity-50`}>
                                  <div className="text-white opacity-50">
                                    {typeIcons[interview.format]}
                                  </div>
                                </div>
                                <div className="flex-1">
                                  <p className="font-medium text-slate-400">{interview.match?.job?.structured?.title || 'Interview'}</p>
                                  <p className="text-sm text-slate-400">{interview.match?.job?.structured?.title || 'Position'}</p>
                                </div>
                                <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-green-100 text-green-700 border border-green-200">
                                  Completed
                                </span>
                              </motion.div>
                            ))}
                          </div>
                        </div>
                      </RevealSection>
                    )}
                  </>
                )}
      </div>
    </div>
  )
}
