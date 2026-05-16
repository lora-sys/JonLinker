'use client'

import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'

interface Job {
  id: string
  title: string
  status: string
  created_at: string
  agent?: {
    user?: {
      email?: string
    }
  }
}

interface JobManagementProps {
  jobs: Job[]
  onPauseJob?: (jobId: string) => void
  onCloseJob?: (jobId: string) => void
  onUpdateRequirements?: (jobId: string) => void
}

export function JobManagement({
  jobs,
  onPauseJob,
  onCloseJob,
  onUpdateRequirements,
}: JobManagementProps) {
  const statusColors: Record<string, string> = {
    draft: 'bg-slate-100 text-slate-800',
    active: 'bg-green-100 text-green-800',
    paused: 'bg-yellow-100 text-yellow-800',
    filled: 'bg-blue-100 text-blue-800',
    closed: 'bg-red-100 text-red-800',
  }

  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
      })
    }
    catch {
      return dateStr
    }
  }

  return (
    <Card variant="outlined" padding="none">
      <div className="p-4 border-b border-slate-200">
        <h3 className="text-lg font-semibold text-slate-900">Job Management</h3>
      </div>
      <div className="divide-y divide-slate-100">
        {jobs.length === 0
          ? (
              <div className="p-8 text-center text-slate-500">
                No jobs found
              </div>
            )
          : (
              jobs.map(job => (
                <div key={job.id} className="p-4 hover:bg-slate-50 transition-colors">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium text-slate-900">{job.title}</h4>
                        <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[job.status] || 'bg-slate-100 text-slate-800'}`}>
                          {job.status}
                        </span>
                      </div>
                      <p className="text-sm text-slate-500 mt-1">
                        Created
                        {' '}
                        {formatDate(job.created_at)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {job.status === 'active' && onPauseJob && (
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => onPauseJob(job.id)}
                        >
                          Pause
                        </Button>
                      )}
                      {job.status === 'paused' && onCloseJob && (
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => onCloseJob(job.id)}
                        >
                          Close
                        </Button>
                      )}
                      {onUpdateRequirements && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onUpdateRequirements(job.id)}
                        >
                          Edit
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
      </div>
    </Card>
  )
}
