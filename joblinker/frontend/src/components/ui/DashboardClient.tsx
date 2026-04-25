'use client';

import { useAuthStore } from '@/stores/auth';
import { useAgentStore } from '@/stores/agent';
import { useMatchStore } from '@/stores/match';
import Link from 'next/link';
import { BentoGrid, BentoCard } from '@/components/ui/BentoGrid';
import { AnimatedTabs } from '@/components/ui/AnimatedTabs';
import FAB from '@/components/ui/FAB';
import StatusDot from '@/components/ui/StatusDot';

function DashboardContent() {
  const { user } = useAuthStore();
  const { agents } = useAgentStore();
  const { matches } = useMatchStore();

  const activeAgents = agents.filter((a) => a.status === 'active').length;
  const pendingMatches = matches.filter((m) => m.status === 'pending').length;

  const statsTabs = [
    {
      id: 'overview',
      label: 'Overview',
      icon: (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
        </svg>
      ),
      content: (
        <BentoGrid>
          <BentoCard className="p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm font-medium text-slate-500">Active Agents</p>
                <p className="text-4xl font-bold text-slate-900 mt-2">{activeAgents}</p>
              </div>
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-100 to-sky-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
            </div>
          </BentoCard>

          <BentoCard className="p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm font-medium text-slate-500">Pending Matches</p>
                <p className="text-4xl font-bold text-slate-900 mt-2">{pendingMatches}</p>
              </div>
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-sky-100 to-cyan-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-sky-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                </svg>
              </div>
            </div>
          </BentoCard>

          <BentoCard className="p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm font-medium text-slate-500">Total Matches</p>
                <p className="text-4xl font-bold text-slate-900 mt-2">{matches.length}</p>
              </div>
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-green-100 to-emerald-100 flex items-center justify-center">
                <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
          </BentoCard>

          <BentoCard span="wide" className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-slate-900">Your Agents</h3>
              <Link href="/agents/create" className="text-sm text-blue-600 hover:text-blue-700 font-medium transition-colors">
                + Create Agent
              </Link>
            </div>
            {agents.length === 0 ? (
              <div className="text-center py-8">
                <div className="w-16 h-16 bg-gradient-to-br from-blue-100 to-sky-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </div>
                <p className="text-slate-500 mb-4">No agents created yet</p>
                <Link href="/agents/create" className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 text-sm font-medium shadow-lg shadow-blue-600/25 transition-all cursor-pointer">
                  Create Your First Agent
                </Link>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {agents.slice(0, 3).map((agent) => (
                  <div key={agent.id} className="bg-slate-50 rounded-xl p-4 hover:bg-slate-100 transition-colors cursor-pointer">
                    <div className="flex items-center gap-3">
                      <StatusDot status={agent.status === 'active' ? 'active' : 'idle'} size="md" />
                      <div>
                        <p className="font-medium text-slate-900 capitalize">{agent.type}</p>
                        <p className="text-sm text-slate-500">{agent.status}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </BentoCard>

          <BentoCard span="wide" className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-slate-900">Recent Matches</h3>
              <Link href="/matches" className="text-sm text-blue-600 hover:text-blue-700 font-medium transition-colors">
                View All
              </Link>
            </div>
            {matches.length === 0 ? (
              <div className="text-center py-8">
                <div className="w-16 h-16 bg-gradient-to-br from-sky-100 to-cyan-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-sky-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                  </svg>
                </div>
                <p className="text-slate-500">No matches yet</p>
              </div>
            ) : (
              <div className="space-y-3">
                {matches.slice(0, 3).map((match) => (
                  <div key={match.id} className="bg-slate-50 rounded-xl p-4 hover:bg-slate-100 transition-colors cursor-pointer">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-slate-900">Match #{match.id.slice(0, 8)}</p>
                        <p className="text-sm text-slate-500">Score: {(match.score * 100).toFixed(1)}%</p>
                      </div>
                      <span className={`px-3 py-1 text-xs font-medium rounded-full ${
                        match.status === 'mutual_interest' ? 'bg-green-100 text-green-700' :
                        match.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                        'bg-slate-200 text-slate-600'
                      }`}>
                        {match.status.replace('_', ' ')}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </BentoCard>
        </BentoGrid>
      ),
    },
    {
      id: 'agents',
      label: 'Agents',
      icon: (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      ),
      content: (
        <div className="text-center py-12">
          <p className="text-slate-500">View all your agents here</p>
          <Link href="/agents" className="inline-flex items-center gap-2 mt-4 px-5 py-2.5 bg-blue-600 text-white rounded-xl hover:bg-blue-700 text-sm font-medium transition-colors">
            Go to Agents
          </Link>
        </div>
      ),
    },
    {
      id: 'matches',
      label: 'Matches',
      icon: (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
        </svg>
      ),
      content: (
        <div className="text-center py-12">
          <p className="text-slate-500">View all your matches here</p>
          <Link href="/matches" className="inline-flex items-center gap-2 mt-4 px-5 py-2.5 bg-blue-600 text-white rounded-xl hover:bg-blue-700 text-sm font-medium transition-colors">
            Go to Matches
          </Link>
        </div>
      ),
    },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900">
          Welcome back, {user?.email?.split('@')[0] || 'User'}
        </h1>
        <p className="text-slate-600 mt-2">Here&apos;s what&apos;s happening with your recruitment activity</p>
      </div>

      <AnimatedTabs tabs={statsTabs} defaultTab="overview" />
      <FAB label="Quick Actions" />
    </div>
  );
}

export default DashboardContent;
