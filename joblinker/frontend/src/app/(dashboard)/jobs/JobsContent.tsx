'use client'

import { Briefcase, Plus, Search, SortAsc } from 'lucide-react'
import Link from 'next/link'
import { useCallback, useMemo, useState } from 'react'

import type { Job } from '@/types'

import { JobCard } from '@/components/feature/JobCard'
import { Card, EmptyState, ErrorState, LoadingSkeleton } from '@/components/ui'
import { useRole } from '@/hooks/useRole'
import { apiClient } from '@/lib/api_client'
import { useAuthStore } from '@/stores/auth'

interface JobFiltersProps {
  jobs: Job[]
}

function JobFilters({ jobs }: JobFiltersProps) {
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState<'title' | 'company'>('title')

  const filteredJobs = useMemo(() => {
    return jobs
      .filter((job) => {
        const matchesSearch
          = job.structured?.title?.toLowerCase().includes(search.toLowerCase())
            || job.structured?.description?.toLowerCase().includes(search.toLowerCase())
        return matchesSearch
      })
      .sort((a, b) => {
        const aVal = a.structured?.title || ''
        const bVal = b.structured?.title || ''
        return aVal.localeCompare(bVal)
      })
  }, [jobs, search])

  return (
    <>
      {/* Sticky Search & Filter Bar */}
      <Card className="p-4 mb-6 sticky top-4 z-sticky bg-white/80 backdrop-blur-xl border border-white/20">
        <div className="flex flex-col md:flex-row gap-3">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              type="text"
              placeholder="Search jobs..."
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 bg-slate-50/50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent cursor-text transition-all duration-200"
            />
          </div>

          <div className="relative">
            <SortAsc className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
            <select
              value={sortBy}
              onChange={e => setSortBy(e.target.value as 'title' | 'company')}
              className="w-full md:w-auto pl-10 pr-8 py-2.5 bg-slate-50/50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer transition-all duration-200 appearance-none"
            >
              <option value="title">Sort by Title</option>
              <option value="company">Sort by Company</option>
            </select>
          </div>
        </div>
      </Card>

      {filteredJobs.length === 0
        ? (
            <EmptyState
              icon={<Briefcase className="w-10 h-10 text-blue-600" />}
              title="No jobs found"
              description={search ? 'Try adjusting your search criteria' : 'Check back later for new opportunities'}
            />
          )
        : (
            <div className="space-y-4">
              {filteredJobs.map(job => (
                <JobCard key={job.id} job={job} />
              ))}
            </div>
          )}
    </>
  )
}

interface JobsContentProps {
  initialJobs: Job[]
}

export function JobsContent({ initialJobs }: JobsContentProps) {
  const [jobs, setJobs] = useState<Job[]>(initialJobs)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { user, isAuthenticated } = useAuthStore()
  const role = useRole()

  const fetchJobs = useCallback(async () => {
    if (!isAuthenticated || !user) {
      setIsLoading(false)
      return
    }
    try {
      setIsLoading(true)
      setError(null)
      const data = await apiClient.get<Job[]>('/api/jobs')
      const parsedJobs = data.map(job => ({
        ...job,
        structured: typeof job.structured === 'string'
          ? JSON.parse(job.structured)
          : job.structured,
      }))
      setJobs(parsedJobs)
    }
    catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load jobs'
      if (errorMessage.includes('Authorization') || errorMessage.includes('401')) {
        setError(null)
        setJobs([])
      }
      else {
        setError(errorMessage)
      }
    }
    finally {
      setIsLoading(false)
    }
  }, [isAuthenticated, user])

  const isRecruiter = role === 'recruiter'

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3">
              <Briefcase className="w-8 h-8 text-blue-600" />
              Job Board
            </h1>
            <p className="text-slate-600 mt-1">Discover opportunities that match your skills</p>
          </div>
          {isRecruiter && (
            <Link href="/jobs/new">
              <button className="flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 text-sm font-medium shadow-lg shadow-blue-600/25 transition-all hover:-translate-y-0.5">
                <Plus className="w-4 h-4" />
                Post a Job
              </button>
            </Link>
          )}
        </div>

        {isLoading
          ? (
              <LoadingSkeleton count={3} variant="card" />
            )
          : error
            ? (
                <ErrorState message={error} onRetry={fetchJobs} />
              )
            : (
                <JobFilters jobs={jobs} />
              )}
      </div>
    </div>
  )
}
