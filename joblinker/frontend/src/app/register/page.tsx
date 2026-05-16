'use client'

import { Briefcase, Eye, EyeOff, Lock, Mail, User, UserPlus } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { Suspense, useState } from 'react'

import { BackgroundBeams } from '@/components/ui/BackgroundBeams'
import { useAuthStore } from '@/stores/auth'

function RegisterForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const setAuth = useAuthStore(state => state.setAuth)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [role, setRole] = useState<'seeker' | 'recruiter'>(
    (searchParams.get('role') as 'seeker' | 'recruiter') || 'seeker',
  )
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }

    setIsLoading(true)

    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, role }),
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Registration failed')
      }

      const data = await response.json()
      // API route sets cookie server-side - just redirect
      setAuth(data.user, data.token)
      router.push('/dashboard')
    }
    catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    }
    finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-2xl shadow-xl p-8">
      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm flex items-center gap-2">
          <Lock className="w-4 h-4" />
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label htmlFor="email" className="block text-sm font-medium text-slate-700 mb-1.5">
            Email address
          </label>
          <div className="relative">
            <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              id="email"
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              className="w-full pl-12 pr-4 py-3 bg-slate-50/50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200 cursor-text"
              placeholder="you@example.com"
              required
            />
          </div>
        </div>

        <div>
          <label htmlFor="password" className="block text-sm font-medium text-slate-700 mb-1.5">
            Password
          </label>
          <div className="relative">
            <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={e => setPassword(e.target.value)}
              className="w-full pl-12 pr-12 py-3 bg-slate-50/50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200 cursor-text"
              placeholder="••••••••"
              required
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 transition-colors cursor-pointer"
            >
              {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
            </button>
          </div>
        </div>

        <div>
          <label htmlFor="confirmPassword" className="block text-sm font-medium text-slate-700 mb-1.5">
            Confirm password
          </label>
          <div className="relative">
            <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              id="confirmPassword"
              type={showPassword ? 'text' : 'password'}
              value={confirmPassword}
              onChange={e => setConfirmPassword(e.target.value)}
              className="w-full pl-12 pr-4 py-3 bg-slate-50/50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all duration-200 cursor-text"
              placeholder="••••••••"
              required
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-2">
            I am a...
          </label>
          <div className="grid grid-cols-2 gap-3">
            <button
              type="button"
              onClick={() => setRole('seeker')}
              className={`p-4 rounded-xl border-2 transition-all duration-200 cursor-pointer ${
                role === 'seeker'
                  ? 'border-blue-600 bg-blue-50 text-blue-700'
                  : 'border-slate-200 text-slate-600 hover:border-slate-300 hover:bg-slate-50'
              }`}
            >
              <User className="w-6 h-6 mb-2 mx-auto" />
              <div className="font-medium">Job Seeker</div>
              <div className="text-sm opacity-75">Looking for opportunities</div>
            </button>
            <button
              type="button"
              onClick={() => setRole('recruiter')}
              className={`p-4 rounded-xl border-2 transition-all duration-200 cursor-pointer ${
                role === 'recruiter'
                  ? 'border-blue-600 bg-blue-50 text-blue-700'
                  : 'border-slate-200 text-slate-600 hover:border-slate-300 hover:bg-slate-50'
              }`}
            >
              <Briefcase className="w-6 h-6 mb-2 mx-auto" />
              <div className="font-medium">Recruiter</div>
              <div className="text-sm opacity-75">Hiring talent</div>
            </button>
          </div>
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className="w-full py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl font-semibold hover:from-blue-700 hover:to-sky-600 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-lg shadow-blue-600/25 hover:shadow-blue-600/35 cursor-pointer flex items-center justify-center gap-2"
        >
          <UserPlus className="w-5 h-5" />
          {isLoading ? 'Creating account...' : 'Create account'}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-slate-600">
        Already have an account?
        {' '}
        <Link href="/login" className="text-blue-600 hover:text-blue-700 font-medium transition-colors cursor-pointer">
          Sign in
        </Link>
      </p>
    </div>
  )
}

function RegisterFallback() {
  return (
    <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-2xl shadow-xl p-8 animate-pulse">
      <div className="space-y-5">
        <div className="h-12 bg-slate-100 rounded-xl"></div>
        <div className="h-12 bg-slate-100 rounded-xl"></div>
        <div className="h-12 bg-slate-100 rounded-xl"></div>
        <div className="h-24 bg-slate-100 rounded-xl"></div>
        <div className="h-12 bg-slate-100 rounded-xl"></div>
      </div>
    </div>
  )
}

export default function RegisterPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-white relative overflow-hidden">
      <BackgroundBeams />

      {/* Header */}
      <header className="relative z-10 w-full max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        <nav className="flex items-center justify-between">
          <Link href="/" className="flex items-center gap-2 cursor-pointer">
            <div className="relative w-10 h-10">
              <Image
                src="/logo.png"
                alt="JobLinker"
                fill
                className="object-contain"
              />
            </div>
            <span className="font-bold text-xl text-slate-900">JobLinker</span>
          </Link>
        </nav>
      </header>

      {/* Register Form */}
      <main className="relative z-10 w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full mx-auto">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-slate-900">Create your account</h1>
            <p className="text-slate-600 mt-2">Start your A2A recruitment journey</p>
          </div>
          <Suspense fallback={<RegisterFallback />}>
            <RegisterForm />
          </Suspense>
        </div>
      </main>
    </div>
  )
}
