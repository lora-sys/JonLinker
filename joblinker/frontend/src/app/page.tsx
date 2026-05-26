'use client'

import { motion } from 'framer-motion'
import Image from 'next/image'
import Link from 'next/link'
import { useEffect, useState } from 'react'

import { useIntersectionReveal } from '@/shared/hooks/useIntersectionReveal'

function AgentNode({ type, label, position, delay = 0 }: { type: 'seeker' | 'recruiter', label: string, position: string, delay?: number }) {
  const [ref, isVisible] = useIntersectionReveal<HTMLDivElement>({ threshold: 0.1 })

  return (
    <div
      ref={ref}
      className={`absolute ${position} transition-all duration-700 ${isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'}`}
      style={{ transitionDelay: `${delay}ms` }}
    >
      <div className={`w-20 h-20 rounded-2xl backdrop-blur-xl border-2 flex flex-col items-center justify-center gap-1 ${
        type === 'seeker'
          ? 'bg-blue-500/20 border-blue-400/50'
          : 'bg-emerald-500/20 border-emerald-400/50'
      }`}
      >
        <div className={`w-8 h-8 rounded-lg ${type === 'seeker' ? 'bg-blue-500' : 'bg-emerald-500'} flex items-center justify-center`}>
          <span className="text-white text-xs font-bold">{type === 'seeker' ? 'S' : 'R'}</span>
        </div>
        <span className="text-[10px] font-medium text-slate-300">{label}</span>
      </div>
    </div>
  )
}

function MatchingPulse({ active }: { active: boolean }) {
  if (!active)
    return null
  return (
    <div className="absolute inset-0 flex items-center justify-center">
      {[0, 200, 400].map(delay => (
        <div
          key={delay}
          className="absolute w-4 h-4 rounded-full bg-gradient-to-r from-blue-500 to-emerald-400 animate-ping opacity-60"
          style={{ animationDelay: `${delay}ms`, animationDuration: '1.5s' }}
        />
      ))}
    </div>
  )
}

function A2AMatchVisualizer() {
  const [active, setActive] = useState(false)
  const [ref, isVisible] = useIntersectionReveal<HTMLDivElement>({ threshold: 0.3 })

  if (isVisible && !active) {
    setActive(true)
  }

  return (
    <div ref={ref} className="relative h-48 w-full max-w-md mx-auto">
      <MatchingPulse active={active} />

      {/* Central Hub */}
      <div className={`absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 transition-all duration-1000 ${active ? 'opacity-100 scale-100' : 'opacity-0 scale-50'}`}>
        <div className="w-16 h-16 rounded-full bg-gradient-to-br from-slate-800 to-slate-900 flex items-center justify-center shadow-2xl border border-slate-700">
          <svg className="w-8 h-8 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
      </div>

      {/* Seekers */}
      <AgentNode type="seeker" label="React Dev" position="top-4 left-8" delay={0} />
      <AgentNode type="seeker" label="Go Expert" position="top-4 right-8" delay={100} />
      <AgentNode type="seeker" label="DevOps" position="top-16 left-1/4" delay={200} />

      {/* Recruiters */}
      <AgentNode type="recruiter" label="TechCorp" position="bottom-4 left-8" delay={300} />
      <AgentNode type="recruiter" label="StartupX" position="bottom-4 right-8" delay={400} />
      <AgentNode type="recruiter" label="CloudBase" position="bottom-16 right-1/4" delay={500} />

      {/* Connection lines */}
      {active && (
        <svg className="absolute inset-0 w-full h-full pointer-events-none" style={{ zIndex: -1 }}>
          <line x1="15%" y1="20%" x2="50%" y2="50%" stroke="url(#lineGrad)" strokeWidth="1" opacity="0.3" className="animate-pulse" />
          <line x1="85%" y1="20%" x2="50%" y2="50%" stroke="url(#lineGrad)" strokeWidth="1" opacity="0.3" className="animate-pulse" style={{ animationDelay: '200ms' }} />
          <line x1="25%" y1="35%" x2="50%" y2="50%" stroke="url(#lineGrad)" strokeWidth="1" opacity="0.3" className="animate-pulse" style={{ animationDelay: '400ms' }} />
          <line x1="15%" y1="80%" x2="50%" y2="50%" stroke="url(#lineGrad2)" strokeWidth="1" opacity="0.3" className="animate-pulse" style={{ animationDelay: '600ms' }} />
          <line x1="85%" y1="80%" x2="50%" y2="50%" stroke="url(#lineGrad2)" strokeWidth="1" opacity="0.3" className="animate-pulse" style={{ animationDelay: '800ms' }} />
          <line x1="75%" y1="65%" x2="50%" y2="50%" stroke="url(#lineGrad2)" strokeWidth="1" opacity="0.3" className="animate-pulse" style={{ animationDelay: '1000ms' }} />
          <defs>
            <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="#3B82F6" />
              <stop offset="100%" stopColor="#10B981" />
            </linearGradient>
            <linearGradient id="lineGrad2" x1="100%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" stopColor="#10B981" />
              <stop offset="100%" stopColor="#3B82F6" />
            </linearGradient>
          </defs>
        </svg>
      )}
    </div>
  )
}

function RevealSection({ children, className = '', delay = 0, immediate = false }: { children: React.ReactNode, className?: string, delay?: number, immediate?: boolean }) {
  const [ref, isVisible] = useIntersectionReveal<HTMLDivElement>({ threshold: 0.1 })
  const [hasAnimated, setHasAnimated] = useState(false)

  useEffect(() => {
    if (immediate && !hasAnimated) {
      setHasAnimated(true)
    }
  }, [immediate, hasAnimated])

  const shouldAnimate = immediate || isVisible

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={shouldAnimate ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.4, delay: shouldAnimate ? delay * 0.03 : 0 }}
      className={className}
    >
      {children}
    </motion.div>
  )
}

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-50 via-blue-50/50 to-white relative overflow-hidden">
      {/* Subtle grid pattern */}
      <div
        className="absolute inset-0 opacity-[0.02]"
        style={{
          backgroundImage: 'linear-gradient(rgba(0,0,0,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(0,0,0,0.1) 1px, transparent 1px)',
          backgroundSize: '64px 64px',
        }}
      />

      {/* Header */}
      <header className="relative z-10 w-full max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        <nav className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link href="/" className="relative w-11 h-11">
              <Image
                src="/logo.png"
                alt="JobLinker"
                fill
                className="object-contain"
              />
            </Link>
            <span className="font-bold text-xl text-slate-900">JobLinker</span>
          </div>
          <div className="flex items-center gap-5">
            <Link
              href="/login"
              className="text-slate-600 hover:text-slate-900 font-medium transition-colors duration-200 cursor-pointer"
            >
              Sign In
            </Link>
            <Link
              href="/register"
              className="px-5 py-2.5 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-semibold text-sm shadow-lg shadow-blue-600/25 transition-all duration-200 cursor-pointer"
            >
              Get Started
            </Link>
          </div>
        </nav>
      </header>

      {/* Main Content */}
      <main className="relative z-10 w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">

        {/* Hero Section */}
        <section className="pt-12 pb-24 text-center">
          <RevealSection delay={0} immediate={true}>
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-blue-50 border border-blue-100 mb-8">
              <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
              <span className="text-sm font-medium text-blue-700">Powered by A2A Protocol</span>
            </div>
          </RevealSection>

          <RevealSection delay={50} immediate={true}>
            <h1 className="text-5xl sm:text-6xl lg:text-7xl font-bold text-slate-900 mb-6 leading-[1.1] tracking-tight">
              <span className="block text-5xl sm:text-6xl lg:text-8xl bg-gradient-to-r from-blue-600 via-blue-500 to-sky-500 bg-clip-text text-transparent">
                A2A
              </span>
              <span className="block mt-2">Recruitment</span>
              <span className="block text-4xl sm:text-5xl lg:text-6xl font-medium text-slate-600 mt-1">Reimagined</span>
            </h1>
          </RevealSection>

          <RevealSection delay={100} immediate={true}>
            <p className="text-lg text-slate-600 mb-12 max-w-xl mx-auto leading-relaxed">
              Autonomous AI agents connect, negotiate, and match opportunities — eliminating friction between seekers and recruiters.
            </p>
          </RevealSection>

          <RevealSection delay={150} immediate={true}>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <Link
                href="/register?role=seeker"
                className="group px-8 py-4 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-2xl hover:from-blue-700 hover:to-sky-600 font-semibold text-lg shadow-xl shadow-blue-600/30 hover:shadow-blue-600/40 transition-all duration-300 hover:-translate-y-1 cursor-pointer"
              >
                <span className="flex items-center gap-3">
                  I&apos;m a Job Seeker
                  <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                </span>
              </Link>
              <Link
                href="/register?role=recruiter"
                className="group px-8 py-4 bg-white text-blue-600 border-2 border-blue-200 rounded-2xl hover:border-blue-400 hover:bg-blue-50 font-semibold text-lg transition-all duration-300 hover:-translate-y-1 cursor-pointer"
              >
                <span className="flex items-center gap-3">
                  I&apos;m a Recruiter
                  <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                </span>
              </Link>
            </div>
          </RevealSection>
        </section>

        {/* A2A Visualization Section */}
        <section className="py-16 -mt-8">
          <RevealSection immediate={true}>
            <div className="bg-white/60 backdrop-blur-xl rounded-3xl border border-white/50 shadow-xl p-8 lg:p-12">
              <div className="text-center mb-8">
                <h2 className="text-3xl lg:text-4xl font-bold text-slate-900 mb-3">
                  Intelligent Matching
                </h2>
                <p className="text-slate-600">Watch agents discover and connect with ideal matches</p>
              </div>
              <A2AMatchVisualizer />
              <div className="flex items-center justify-center gap-8 mt-8">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-blue-500" />
                  <span className="text-sm text-slate-600">Seekers</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-emerald-500" />
                  <span className="text-sm text-slate-600">Recruiters</span>
                </div>
              </div>
            </div>
          </RevealSection>
        </section>

        {/* How It Works */}
        <section className="py-20">
          <RevealSection immediate={true}>
            <h2 className="text-4xl lg:text-5xl font-bold text-center text-slate-900 mb-4">
              How It Works
            </h2>
            <p className="text-center text-lg text-slate-600 mb-16 max-w-2xl mx-auto">
              Three simple steps to smarter hiring
            </p>
          </RevealSection>

          <div className="grid md:grid-cols-3 gap-8">
            {[
              {
                step: '01',
                title: 'Create Your Agent',
                description: 'Register and deploy an AI agent that learns your requirements, preferences, and career goals.',
                icon: (
                  <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                  </svg>
                ),
              },
              {
                step: '02',
                title: 'Define Criteria',
                description: 'Your agent autonomously searches, evaluates, and ranks opportunities based on your unique profile.',
                icon: (
                  <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                ),
              },
              {
                step: '03',
                title: 'Automatic Matching',
                description: 'A2A protocol enables direct agent-to-agent negotiation for optimal placement without manual intervention.',
                icon: (
                  <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                ),
              },
            ].map((item, i) => (
              <RevealSection key={item.step} delay={i * 50} immediate={true}>
                <div className="group relative bg-white/80 backdrop-blur-xl rounded-2xl p-8 shadow-sm hover:shadow-xl transition-all duration-300 hover:-translate-y-1 border border-white/50">
                  <div className="absolute -top-4 -left-4 w-12 h-12 rounded-xl bg-gradient-to-br from-blue-600 to-sky-500 flex items-center justify-center text-white font-bold text-lg shadow-lg">
                    {item.step}
                  </div>
                  <div className="w-16 h-16 bg-gradient-to-br from-blue-50 to-sky-50 rounded-2xl flex items-center justify-center mb-6 text-blue-600 group-hover:from-blue-100 group-hover:to-sky-100 transition-all duration-300">
                    {item.icon}
                  </div>
                  <h3 className="text-xl font-bold text-slate-900 mb-3">{item.title}</h3>
                  <p className="text-slate-600 leading-relaxed">{item.description}</p>
                </div>
              </RevealSection>
            ))}
          </div>
        </section>

        {/* Features Grid */}
        <section className="py-20">
          <RevealSection immediate={true}>
            <h2 className="text-4xl lg:text-5xl font-bold text-center text-slate-900 mb-4">
              Built for the Future
            </h2>
            <p className="text-center text-lg text-slate-600 mb-16 max-w-2xl mx-auto">
              Enterprise-grade infrastructure meets modern design
            </p>
          </RevealSection>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[
              { title: 'Privacy-First', desc: 'AES-256 encryption with local-first storage', color: 'from-emerald-500 to-teal-400' },
              { title: 'A2A Protocol', desc: 'Standardized agent-to-agent communication', color: 'from-blue-500 to-indigo-400' },
              { title: 'Smart Matching', desc: 'ML-powered candidate-job alignment', color: 'from-violet-500 to-purple-400' },
              { title: 'Real-time Sync', desc: 'Live updates across all your devices', color: 'from-orange-500 to-amber-400' },
              { title: 'Interview Tracking', desc: 'Automated scheduling and milestone management', color: 'from-pink-500 to-rose-400' },
              { title: 'Offer Negotiation', desc: 'Agents negotiate terms autonomously', color: 'from-green-500 to-emerald-400' },
            ].map((feature, i) => (
              <RevealSection key={feature.title} delay={i * 30} immediate={true}>
                <motion.div
                  whileHover={{ scale: 1.02 }}
                  className="group bg-white/70 backdrop-blur-xl rounded-2xl p-6 border border-white/50 hover:border-transparent hover:shadow-xl transition-all duration-300 cursor-pointer"
                >
                  <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${feature.color} flex items-center justify-center mb-4 shadow-lg`}>
                    <span className="text-xl">•</span>
                  </div>
                  <h3 className="text-lg font-bold text-slate-900 mb-2">{feature.title}</h3>
                  <p className="text-sm text-slate-600">{feature.desc}</p>
                </motion.div>
              </RevealSection>
            ))}
          </div>
        </section>

        {/* CTA Section */}
        <section className="py-20">
          <RevealSection immediate={true}>
            <div className="bg-gradient-to-br from-blue-600 via-blue-500 to-sky-500 rounded-3xl p-12 lg:p-16 text-center relative overflow-hidden">
              <div className="absolute inset-0 opacity-10">
                <div className="absolute top-0 left-1/4 w-96 h-96 bg-white rounded-full blur-3xl" />
                <div className="absolute bottom-0 right-1/4 w-64 h-64 bg-sky-300 rounded-full blur-3xl" />
              </div>
              <div className="relative z-10">
                <h2 className="text-4xl lg:text-5xl font-bold text-white mb-4">
                  Ready to Transform Hiring?
                </h2>
                <p className="text-xl text-blue-100 mb-10 max-w-2xl mx-auto">
                  Join thousands of agents already finding their perfect match
                </p>
                <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                  <Link
                    href="/register"
                    className="px-8 py-4 bg-white text-blue-600 rounded-2xl font-bold text-lg hover:bg-blue-50 transition-all duration-300 hover:-translate-y-1 shadow-xl cursor-pointer"
                  >
                    Get Started Free
                  </Link>
                  <Link
                    href="/demo"
                    className="px-8 py-4 bg-transparent text-white border-2 border-white/30 rounded-2xl font-semibold text-lg hover:bg-white/10 transition-all duration-300 cursor-pointer"
                  >
                    Watch Demo
                  </Link>
                </div>
              </div>
            </div>
          </RevealSection>
        </section>
      </main>

      {/* Footer */}
      <footer className="py-10 px-4 sm:px-6 lg:px-8 border-t border-slate-200/50 bg-white/50 backdrop-blur relative z-10">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <Link href="/" className="relative w-10 h-10">
              <Image
                src="/logo.png"
                alt="JobLinker"
                fill
                className="object-contain"
              />
            </Link>
            <span className="text-sm text-slate-600">JobLinker 2026 — A2A Recruitment Platform</span>
          </div>
          <div className="flex items-center gap-6 text-sm text-slate-500">
            <a href="/privacy" className="hover:text-slate-700 transition-colors cursor-pointer">Privacy</a>
            <a href="/terms" className="hover:text-slate-700 transition-colors cursor-pointer">Terms</a>
            <a href="/contact" className="hover:text-slate-700 transition-colors cursor-pointer">Contact</a>
          </div>
        </div>
      </footer>
    </div>
  )
}
