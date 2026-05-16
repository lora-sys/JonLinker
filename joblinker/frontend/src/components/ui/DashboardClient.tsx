'use client'

import { AnimatePresence, motion } from 'framer-motion'
import {
  ArrowLeftRight,
  Briefcase,
  Building2,
  Calendar,
  CheckCircle2,
  FileText,
  LayoutGrid,
  MessageSquare,
  Plus,
  Sparkles,
  UserPlus,
  Users,
} from 'lucide-react'
import Link from 'next/link'
import { useEffect, useRef, useState } from 'react'

import { AnimatedTabs } from '@/components/ui/AnimatedTabs'
import { BentoCard, BentoGrid } from '@/components/ui/BentoGrid'
import FAB from '@/components/ui/FAB'
import StatusDot from '@/components/ui/StatusDot'
import { apiClient } from '@/lib/api_client'
import { useAgentStore } from '@/stores/agent'
import { useAuthStore } from '@/stores/auth'
import { useMatchStore } from '@/stores/match'

function RevealCard({ children, delay = 0, className = '' }: { children: React.ReactNode, delay?: number, className?: string }) {
  const [visible, setVisible] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting)
          setVisible(true)
      },
      { threshold: 0.1 },
    )
    if (ref.current)
      observer.observe(ref.current)
    return () => observer.disconnect()
  }, [])

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={visible ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5, delay: delay * 0.1 }}
      className={className}
    >
      {children}
    </motion.div>
  )
}

function AnimatedNumber({ value, suffix = '' }: { value: number, suffix?: string }) {
  const [display, setDisplay] = useState(0)

  useEffect(() => {
    const end = value
    const duration = 1000
    const startTime = Date.now()

    const animate = () => {
      const elapsed = Date.now() - startTime
      const progress = Math.min(elapsed / duration, 1)
      const eased = 1 - (1 - progress) ** 3
      setDisplay(Math.floor(eased * end))

      if (progress < 1)
        requestAnimationFrame(animate)
    }

    requestAnimationFrame(animate)
  }, [value])

  return (
    <span>
      {display}
      {suffix}
    </span>
  )
}

function StatCard({ delay, icon, label, value, suffix = '', color }: { delay: number, icon: React.ReactNode, label: string, value: number, suffix?: string, color: string }) {
  return (
    <RevealCard delay={delay}>
      <BentoCard className="group hover:shadow-lg transition-all duration-300 hover:-translate-y-1">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">{label}</p>
            <p className="text-4xl font-bold text-slate-900 mt-2">
              <AnimatedNumber value={value} suffix={suffix} />
            </p>
          </div>
          <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${color} flex items-center justify-center group-hover:scale-110 transition-transform duration-300`}>
            {icon}
          </div>
        </div>
      </BentoCard>
    </RevealCard>
  )
}

function NavCard({ delay, icon, label, href, description, color }: { delay: number, icon: React.ReactNode, label: string, href: string, description: string, color: string }) {
  return (
    <RevealCard delay={delay}>
      <Link href={href}>
        <BentoCard className="group hover:shadow-lg transition-all duration-300 hover:-translate-y-1 cursor-pointer">
          <div className="flex items-center gap-4">
            <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${color} flex items-center justify-center group-hover:scale-110 transition-transform duration-300`}>
              {icon}
            </div>
            <div className="flex-1">
              <p className="font-semibold text-slate-900">{label}</p>
              <p className="text-sm text-slate-500">{description}</p>
            </div>
          </div>
        </BentoCard>
      </Link>
    </RevealCard>
  )
}

function AgentCard({ agent, index }: { agent: any, index: number }) {
  return (
    <RevealCard delay={index}>
      <motion.div
        className="bg-slate-50 rounded-xl p-4 hover:bg-slate-100 transition-all duration-200 cursor-pointer border border-transparent hover:border-blue-200"
        whileHover={{ scale: 1.02 }}
      >
        <div className="flex items-center gap-3">
          <StatusDot status={agent.status === 'active' ? 'active' : 'idle'} size="md" />
          <div>
            <p className="font-medium text-slate-900 capitalize">{agent.type}</p>
            <p className="text-sm text-slate-500">{agent.status}</p>
          </div>
        </div>
      </motion.div>
    </RevealCard>
  )
}

function MatchCard({ match, index }: { match: any, index: number }) {
  const statusColors = {
    mutual_interest: 'bg-green-100 text-green-700',
    pending: 'bg-yellow-100 text-yellow-700',
    declined: 'bg-slate-200 text-slate-600',
  }

  return (
    <RevealCard delay={index}>
      <motion.div
        className="bg-slate-50 rounded-xl p-4 hover:bg-slate-100 transition-all duration-200 cursor-pointer border border-transparent hover:border-blue-200"
        whileHover={{ scale: 1.02 }}
      >
        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium text-slate-900">
              Match #
              {match.id.slice(0, 8)}
            </p>
            <p className="text-sm text-slate-500">
              Score:
              {(match.score * 100).toFixed(1)}
              %
            </p>
          </div>
          <span className={`px-3 py-1 text-xs font-medium rounded-full ${statusColors[match.status as keyof typeof statusColors] || 'bg-slate-200 text-slate-600'}`}>
            {match.status.replace('_', ' ')}
          </span>
        </div>
      </motion.div>
    </RevealCard>
  )
}

function EmptyState({ icon, title, actionLabel, href }: { icon: React.ReactNode, title: string, description: string, actionLabel: string, href: string }) {
  return (
    <div className="text-center py-12">
      <div className="w-20 h-20 bg-gradient-to-br from-blue-100 to-sky-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
        {icon}
      </div>
      <p className="text-slate-500 mb-4">{title}</p>
      <Link
        href={href}
        className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 text-sm font-medium shadow-lg shadow-blue-600/25 transition-all hover:-translate-y-0.5 cursor-pointer"
      >
        <Plus className="w-4 h-4" />
        {actionLabel}
      </Link>
    </div>
  )
}

function DashboardContent() {
  const { user } = useAuthStore()
  const { agents, setAgents } = useAgentStore()
  const { matches, setMatches } = useMatchStore()
  const [, setLoading] = useState(true)

  // Fetch agents and matches on mount
  useEffect(() => {
    async function loadData() {
      try {
        const [agentsData, matchesData] = await Promise.all([
          apiClient.get<any[]>('/api/agents'),
          apiClient.get<any[]>('/api/matches'),
        ])
        setAgents(agentsData || [])
        setMatches(matchesData || [])
      }
      catch (err) {
        console.error('Failed to load dashboard data:', err)
      }
      finally {
        setLoading(false)
      }
    }
    loadData()
  }, [setAgents, setMatches])

  const activeAgents = agents.filter(a => a.status === 'active').length
  const pendingMatches = matches.filter(m => m.status === 'pending').length

  const statsTabs = [
    {
      id: 'overview',
      label: 'Overview',
      icon: <LayoutGrid className="w-4 h-4" />,
      content: (
        <AnimatePresence mode="wait">
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
          >
            <BentoGrid>
              <StatCard
                delay={0}
                color="from-blue-100 to-sky-100"
                icon={<Users className="w-6 h-6 text-blue-600" />}
                label="Active Agents"
                value={activeAgents}
              />
              <StatCard
                delay={1}
                color="from-sky-100 to-cyan-100"
                icon={<ArrowLeftRight className="w-6 h-6 text-sky-600" />}
                label="Pending Matches"
                value={pendingMatches}
              />
              <StatCard
                delay={2}
                color="from-green-100 to-emerald-100"
                icon={<CheckCircle2 className="w-6 h-6 text-green-600" />}
                label="Total Matches"
                value={matches.length}
              />

              <NavCard
                delay={3}
                color="from-blue-100 to-indigo-100"
                icon={<Briefcase className="w-6 h-6 text-blue-600" />}
                label="Jobs"
                description="Browse and manage job listings"
                href="/jobs"
              />
              <NavCard
                delay={4}
                color="from-amber-100 to-orange-100"
                icon={<FileText className="w-6 h-6 text-amber-600" />}
                label="Offers"
                description="View and respond to offers"
                href="/offers"
              />
              <NavCard
                delay={5}
                color="from-purple-100 to-violet-100"
                icon={<Calendar className="w-6 h-6 text-purple-600" />}
                label="Interviews"
                description="Schedule and manage interviews"
                href="/interviews"
              />
              <NavCard
                delay={6}
                color="from-cyan-100 to-sky-100"
                icon={<MessageSquare className="w-6 h-6 text-cyan-600" />}
                label="Messages"
                description="View A2A conversations"
                href="/messages"
              />

              {/* Recruiter-specific card */}
              <NavCard
                delay={7}
                color="from-rose-100 to-pink-100"
                icon={<Building2 className="w-6 h-6 text-rose-600" />}
                label="Post a Job"
                description="Create a new job listing"
                href="/jobs/new"
              />

              <BentoCard span="wide" className="p-6">
                <div className="flex items-center justify-between mb-6">
                  <h3 className="text-lg font-semibold text-slate-900 flex items-center gap-2">
                    <Sparkles className="w-5 h-5 text-blue-500" />
                    Your Agents
                  </h3>
                  <Link href="/agents/create" className="text-sm text-blue-600 hover:text-blue-700 font-medium transition-colors cursor-pointer flex items-center gap-1">
                    <UserPlus className="w-4 h-4" />
                    Create Agent
                  </Link>
                </div>
                {agents.length === 0
                  ? (
                      <EmptyState
                        icon={<Users className="w-8 h-8 text-blue-600" />}
                        title="No agents created yet"
                        description="Create your first agent to start matching"
                        actionLabel="Create Your First Agent"
                        href="/agents/create"
                      />
                    )
                  : (
                      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {agents.slice(0, 3).map((agent, i) => (
                          <AgentCard key={agent.id} agent={agent} index={i} />
                        ))}
                      </div>
                    )}
              </BentoCard>

              <BentoCard span="wide" className="p-6">
                <div className="flex items-center justify-between mb-6">
                  <h3 className="text-lg font-semibold text-slate-900 flex items-center gap-2">
                    <ArrowLeftRight className="w-5 h-5 text-sky-500" />
                    Recent Matches
                  </h3>
                  <Link href="/matches" className="text-sm text-blue-600 hover:text-blue-700 font-medium transition-colors cursor-pointer">
                    View All
                  </Link>
                </div>
                {matches.length === 0
                  ? (
                      <EmptyState
                        icon={<ArrowLeftRight className="w-8 h-8 text-sky-600" />}
                        title="No matches yet"
                        description="Matches will appear when agents find opportunities"
                        actionLabel="View Matches"
                        href="/matches"
                      />
                    )
                  : (
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {matches.slice(0, 4).map((match, i) => (
                          <MatchCard key={match.id} match={match} index={i} />
                        ))}
                      </div>
                    )}
              </BentoCard>
            </BentoGrid>
          </motion.div>
        </AnimatePresence>
      ),
    },
    {
      id: 'agents',
      label: 'Agents',
      icon: <Users className="w-4 h-4" />,
      content: (
        <AnimatePresence mode="wait">
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
          >
            <EmptyState
              icon={<Users className="w-8 h-8 text-blue-600" />}
              title="View all your agents here"
              description="Manage and monitor your AI recruitment agents"
              actionLabel="Go to Agents"
              href="/agents"
            />
          </motion.div>
        </AnimatePresence>
      ),
    },
    {
      id: 'matches',
      label: 'Matches',
      icon: <ArrowLeftRight className="w-4 h-4" />,
      content: (
        <AnimatePresence mode="wait">
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
          >
            <EmptyState
              icon={<ArrowLeftRight className="w-8 h-8 text-sky-600" />}
              title="View all your matches here"
              description="Track and manage your job matches"
              actionLabel="Go to Matches"
              href="/matches"
            />
          </motion.div>
        </AnimatePresence>
      ),
    },
  ]

  return (
    <div className="px-4 sm:px-6 lg:px-8 py-6">
      <RevealCard>
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-slate-900 flex items-center gap-3">
            Welcome back,
            <span className="bg-gradient-to-r from-blue-600 to-sky-500 bg-clip-text text-transparent">
              {user?.email?.split('@')[0] || 'User'}
            </span>
          </h1>
          <p className="text-slate-600 mt-2 text-lg">Here&apos;s what&apos;s happening with your recruitment activity</p>
        </div>
      </RevealCard>

      <AnimatedTabs tabs={statsTabs} defaultTab="overview" />
      <FAB label="Quick Actions" />
    </div>
  )
}

export default DashboardContent
