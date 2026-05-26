'use client'

import { useRouter } from 'next/navigation'
import { useEffect } from 'react'

import { BackgroundBeams } from '@/shared/ui/background-beams'
import JobForm from '@/shared/ui/job-form'
import { useRole } from '@/shared/hooks/useRole'
import { useAuthStore } from '@/shared/stores/auth'

export default function NewJobPage() {
  const router = useRouter()
  const role = useRole()
  const isAuthenticated = useAuthStore(state => state.isAuthenticated)

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace('/login')
      return
    }
    if (role !== 'recruiter') {
      router.replace('/dashboard')
    }
  }, [isAuthenticated, role, router])

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-zinc-950 flex items-center justify-center">
        <div className="animate-pulse text-cyan-400">Loading...</div>
      </div>
    )
  }

  if (role !== 'recruiter') {
    return null
  }

  return (
    <div className="min-h-screen bg-zinc-950 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-cyan-950/20 via-zinc-950 to-zinc-950 pointer-events-none" />
      <BackgroundBeams />
      <div className="relative z-10 flex items-center justify-center min-h-screen px-4 py-12">
        <JobForm />
      </div>
    </div>
  )
}
