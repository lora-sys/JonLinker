import { Building2, Calendar, Clock, ExternalLink, Phone, Video } from 'lucide-react'

import type { Interview } from '@/types'

interface InterviewCardProps {
  interview: Interview
  onConfirm?: () => void
  onCancel?: () => void
}

export function InterviewCard({ interview, onConfirm, onCancel }: InterviewCardProps) {
  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleDateString('en-US', {
        month: 'long',
        day: 'numeric',
        year: 'numeric',
      })
    }
    catch {
      return dateStr
    }
  }

  const formatTime = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleTimeString('en-US', {
        hour: 'numeric',
        minute: '2-digit',
      })
    }
    catch {
      return ''
    }
  }

  const typeIcons: Record<string, React.ReactNode> = {
    video: <Video className="w-4 h-4" />,
    phone: <Phone className="w-4 h-4" />,
    onsite: <Building2 className="w-4 h-4" />,
  }

  const statusColors: Record<string, string> = {
    scheduled: 'bg-blue-100 text-blue-700 border-blue-200',
    confirmed: 'bg-green-100 text-green-700 border-green-200',
    completed: 'bg-slate-100 text-slate-600 border-slate-200',
    cancelled: 'bg-red-100 text-red-700 border-red-200',
  }

  const statusLabels: Record<string, string> = {
    scheduled: 'Scheduled',
    confirmed: 'Confirmed',
    completed: 'Completed',
    cancelled: 'Cancelled',
  }

  return (
    <div className="bg-white rounded-xl border border-slate-200 p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-3">
            <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium border ${statusColors[interview.status] || 'bg-slate-100 text-slate-600 border-slate-200'}`}>
              {statusLabels[interview.status] || interview.status}
            </span>
            <span className="flex items-center gap-1.5 text-sm text-slate-500">
              {typeIcons[interview.format] || <Video className="w-4 h-4" />}
              <span className="capitalize">{interview.format}</span>
            </span>
          </div>

          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <Calendar className="w-4 h-4 text-slate-400" />
              <span className="text-slate-700">{formatDate(interview.scheduled_at)}</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Clock className="w-4 h-4 text-slate-400" />
              <span className="text-slate-700">{formatTime(interview.scheduled_at)}</span>
            </div>
          </div>

          {interview.location && (
            <div className="mt-3">
              {interview.format === 'video'
                ? (
                    <a
                      href={interview.location}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1.5 text-sm text-blue-600 hover:text-blue-700 font-medium"
                    >
                      <ExternalLink className="w-4 h-4" />
                      Join Meeting
                    </a>
                  )
                : (
                    <p className="text-sm text-slate-500">{interview.location}</p>
                  )}
            </div>
          )}
        </div>
      </div>

      {interview.status === 'scheduled' && (
        <div className="flex gap-2 mt-4 pt-4 border-t border-slate-100">
          {onConfirm && (
            <button
              onClick={onConfirm}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
            >
              Confirm
            </button>
          )}
          {onCancel && (
            <button
              onClick={onCancel}
              className="flex-1 px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 transition-colors text-sm font-medium"
            >
              Cancel
            </button>
          )}
        </div>
      )}
    </div>
  )
}

export default InterviewCard
