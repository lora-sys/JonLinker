import { DollarSign, MapPin } from 'lucide-react'

import type { Job } from '@/shared/types'

interface JobCardProps {
  job: Job
  onClick?: () => void
}

export function JobCard({ job, onClick }: JobCardProps) {
  const { structured } = job
  const salaryRange = structured?.salary_range

  const formatSalary = (range: { min: number, max: number, currency: string } | undefined) => {
    if (!range)
      return null
    const min = range.min / 1000
    const max = range.max / 1000
    return `$${min}k - $${max}k`
  }

  const workTypeLabels: Record<string, string> = {
    remote: 'Remote',
    hybrid: 'Hybrid',
    onsite: 'On-site',
  }

  return (
    <div
      className="bg-white rounded-xl border border-slate-200 p-5 hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <h3 className="text-lg font-semibold text-slate-900 truncate">
            {structured?.title || 'Untitled Job'}
          </h3>
          <p className="text-sm text-slate-500 mt-1 line-clamp-2">
            {structured?.description || 'No description'}
          </p>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3 mt-3">
        {structured?.work_type && (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700 border border-blue-200">
            {workTypeLabels[structured.work_type] || structured.work_type}
          </span>
        )}
        {structured?.location && (
          <span className="flex items-center gap-1 text-xs text-slate-500">
            <MapPin className="w-3 h-3" />
            {structured.location}
          </span>
        )}
        {salaryRange && (
          <span className="flex items-center gap-1 text-xs text-slate-500">
            <DollarSign className="w-3 h-3" />
            {formatSalary(salaryRange)}
          </span>
        )}
      </div>

      {structured?.requirements && structured.requirements.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mt-4">
          {structured.requirements.slice(0, 4).map((req, i) => (
            <span
              key={i}
              className="px-2 py-0.5 bg-slate-100 text-slate-600 text-xs rounded-full"
            >
              {req}
            </span>
          ))}
          {structured.requirements.length > 4 && (
            <span className="px-2 py-0.5 bg-slate-100 text-slate-500 text-xs rounded-full">
              +
              {structured.requirements.length - 4}
            </span>
          )}
        </div>
      )}
    </div>
  )
}

export default JobCard
