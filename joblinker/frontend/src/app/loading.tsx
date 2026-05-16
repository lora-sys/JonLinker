'use client'

import Image from 'next/image'
import React, { useEffect, useState } from 'react'

function AnimatedLogo() {
  return (
    <div className="relative">
      {/* Outer ring */}
      <div className="absolute inset-0 rounded-full border-2 border-blue-200 animate-ping opacity-20" />
      {/* Logo image */}
      <div className="w-20 h-20 rounded-full bg-gradient-to-br from-blue-600 via-blue-500 to-sky-400 flex items-center justify-center shadow-xl shadow-blue-500/30 overflow-hidden">
        <div className="relative w-12 h-12">
          <Image
            src="/logo.png"
            alt="JobLinker"
            fill
            className="object-contain"
          />
        </div>
      </div>
      {/* Spinning arc */}
      <svg className="absolute inset-0 w-20 h-20 animate-spin" viewBox="0 0 80 80" style={{ animationDuration: '3s' }}>
        <circle cx="40" cy="40" r="38" fill="none" stroke="url(#gradient)" strokeWidth="3" strokeLinecap="round" strokeDasharray="60 180" />
        <defs>
          <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#3B82F6" />
            <stop offset="100%" stopColor="#06B6D4" />
          </linearGradient>
        </defs>
      </svg>
    </div>
  )
}

function SkeletonCard() {
  return (
    <div className="bg-white/70 backdrop-blur-xl border border-white/50 rounded-2xl p-5 space-y-3">
      <div className="flex items-center gap-3">
        <div className="w-12 h-12 rounded-xl bg-slate-200/60 animate-pulse" />
        <div className="space-y-2 flex-1">
          <div className="h-4 rounded-full bg-slate-200/60 animate-pulse w-3/4" />
          <div className="h-3 rounded-full bg-slate-200/40 animate-pulse w-1/2" />
        </div>
      </div>
      <div className="h-3 rounded-full bg-slate-200/40 animate-pulse" />
      <div className="flex gap-2">
        <div className="h-6 w-16 rounded-full bg-slate-200/40 animate-pulse" />
        <div className="h-6 w-20 rounded-full bg-slate-200/40 animate-pulse" />
      </div>
    </div>
  )
}

function FloatingOrb({ className = '' }: { className?: string }) {
  return <div className={`absolute rounded-full bg-gradient-to-br from-blue-400/20 to-sky-300/10 blur-2xl ${className}`} />
}

export default function LoadingPage() {
  const [progress, setProgress] = useState(0)
  const [phase, setPhase] = useState(0)

  useEffect(() => {
    const progressInterval = setInterval(() => {
      setProgress(p => (p >= 100 ? 0 : p + Math.random() * 20 + 10))
    }, 400)

    const phaseInterval = setInterval(() => {
      setPhase(p => (p + 1) % 3)
    }, 1200)

    return () => {
      clearInterval(progressInterval)
      clearInterval(phaseInterval)
    }
  }, [])

  const phases = ['Connecting', 'Loading', 'Ready']

  return (
    <div className="min-h-screen w-full bg-gradient-to-br from-slate-50 via-blue-50/40 to-white flex flex-col items-center justify-center relative overflow-hidden">
      {/* Background orbs */}
      <FloatingOrb className="w-96 h-96 -top-48 -left-48 animate-pulse" />
      <FloatingOrb className="w-80 h-80 top-1/2 -right-40 animate-pulse" />
      <FloatingOrb className="w-64 h-64 -bottom-32 left-1/3 animate-pulse" />

      {/* Main loading card */}
      <div className="relative z-10 flex flex-col items-center text-center px-6 py-10">
        {/* Logo */}
        <div className="mb-8">
          <AnimatedLogo />
        </div>

        {/* Brand */}
        <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-sky-500 bg-clip-text text-transparent mb-2">
          JobLinker
        </h1>
        <p className="text-slate-500 text-sm mb-10">
          Intelligent Recruitment Platform
        </p>

        {/* Progress bar */}
        <div className="w-64 space-y-3 mb-8">
          <div className="relative h-1.5 bg-slate-200/60 rounded-full overflow-hidden">
            <div
              className="absolute inset-y-0 left-0 bg-gradient-to-r from-blue-500 to-sky-400 rounded-full transition-all duration-300 ease-out"
              style={{ width: `${Math.min(progress, 100)}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-slate-400">
            <span className="flex items-center gap-1.5">
              <span className={`inline-block w-1.5 h-1.5 rounded-full ${progress < 100 ? 'bg-blue-500 animate-pulse' : 'bg-green-500'}`} />
              {phases[phase % 3]}
            </span>
            <span>
              {Math.round(Math.min(progress, 100))}
              %
            </span>
          </div>
        </div>

        {/* Skeleton cards preview */}
        <div className="w-full max-w-sm space-y-3 opacity-70">
          <div className="transform -rotate-1">
            <SkeletonCard />
          </div>
          <div className="transform rotate-0.5">
            <SkeletonCard />
          </div>
          <div className="transform -rotate-0.5">
            <SkeletonCard />
          </div>
        </div>
      </div>

      {/* Bottom status */}
      <div className="absolute bottom-8 left-0 right-0 text-center">
        <div className="flex items-center justify-center gap-2 text-slate-400 text-xs">
          <span className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />
            System Online
          </span>
          <span className="mx-2 text-slate-300">•</span>
          <span>v1.0.0</span>
        </div>
      </div>

    </div>
  )
}
