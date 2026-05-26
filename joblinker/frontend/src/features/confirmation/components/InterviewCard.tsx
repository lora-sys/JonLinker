'use client'

import type { Interview as DomainInterview } from '@/shared/types'

import { Card } from '@/shared/ui/Card'

interface InterviewCardProps {
  interview: DomainInterview
  onReschedule?: (id: string) => void
  onCancel?: (id: string) => void
  onConfirm?: (id: string) => void
}

export function InterviewCard({
  interview,
  onReschedule,
  onCancel,
  onConfirm,
}: InterviewCardProps) {
  const formatDate = (dateStr: string) => {
    try {
      const date = new Date(dateStr)
      return date.toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    }
    catch {
      return dateStr
    }
  }

  const formatIcon = () => {
    switch (interview.format) {
      case 'video':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        )
      case 'phone':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
          </svg>
        )
      case 'onsite':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        )
      default:
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        )
    }
  }

  const statusColors: Record<string, string> = {
    scheduled: 'bg-blue-100 text-blue-800',
    rescheduled: 'bg-yellow-100 text-yellow-800',
    completed: 'bg-green-100 text-green-800',
    cancelled: 'bg-red-100 text-red-800',
  }

  return (
    <Card variant="outlined" padding="md" className="hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <div className="p-2 bg-gray-100 rounded-lg">
            {formatIcon()}
          </div>
          <div>
            <h3 className="font-semibold text-gray-900">
              {interview.match?.job?.structured?.title || 'Interview'}
            </h3>
            <p className="text-sm text-gray-500 mt-1">
              {formatDate(interview.scheduled_at)}
            </p>
            <div className="flex items-center gap-2 mt-2">
              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                {interview.format}
              </span>
              {interview.location && (
                <span className="text-xs text-gray-500">
                  {interview.location}
                </span>
              )}
            </div>
            {/* Participant names */}
            <div className="flex items-center gap-4 mt-3 pt-3 border-t border-gray-100">
              <div className="flex items-center gap-1.5">
                <svg className="w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                <span className="text-xs text-gray-600">
                  <span className="font-medium">Seeker:</span>
                  {' '}
                  {interview.match?.seeker_agent?.user?.email || 'Unknown'}
                </span>
              </div>
              <div className="flex items-center gap-1.5">
                <svg className="w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
                <span className="text-xs text-gray-600">
                  <span className="font-medium">Recruiter:</span>
                  {' '}
                  {interview.match?.job?.agent?.user?.email || 'Unknown'}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div className="flex flex-col items-end gap-2">
          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[interview.status] || 'bg-gray-100 text-gray-800'}`}>
            {interview.status}
          </span>
        </div>
      </div>

      {(interview.status === 'scheduled' || interview.status === 'rescheduled') && (
        <div className="flex items-center gap-2 mt-4 pt-4 border-t border-gray-100">
          {onConfirm && (
            <button
              onClick={() => onConfirm(interview.id)}
              className="px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 transition-colors"
            >
              Confirm
            </button>
          )}
          {onReschedule && (
            <button
              onClick={() => onReschedule(interview.id)}
              className="px-3 py-1.5 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
            >
              Reschedule
            </button>
          )}
          {onCancel && (
            <button
              onClick={() => onCancel(interview.id)}
              className="px-3 py-1.5 text-sm font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 transition-colors"
            >
              Cancel
            </button>
          )}
        </div>
      )}
    </Card>
  )
}
