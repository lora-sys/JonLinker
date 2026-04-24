'use client';

import { useAuthStore } from '@/stores/auth';
import { useAgentStore } from '@/stores/agent';
import { useMatchStore } from '@/stores/match';
import Link from 'next/link';
import Card from '@/components/ui/Card';
import FAB from '@/components/ui/FAB';
import StatusDot from '@/components/ui/StatusDot';

export default function DashboardPage() {
  const { user } = useAuthStore();
  const { agents, currentAgent } = useAgentStore();
  const { matches } = useMatchStore();

  const activeAgents = agents.filter((a) => a.status === 'active').length;
  const pendingMatches = matches.filter((m) => m.status === 'pending').length;

  return (
    <div className="max-w-6xl mx-auto px-4 py-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900" style={{ fontFamily: 'Plus Jakarta Sans, sans-serif' }}>
          Welcome back, {user?.email?.split('@')[0] || 'User'}
        </h1>
        <p className="text-slate-600 mt-2 text-base">Here&apos;s what&apos;s happening with your recruitment activity</p>
      </div>

      {/* Stats Grid - Animated Counters */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mb-8">
        <Card className="p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm font-medium text-slate-500">Active Agents</p>
              <p className="text-3xl font-bold text-slate-900 mt-2 animate-[counter]_500ms_ease-out">{activeAgents}</p>
            </div>
            <div className="w-10 h-10 rounded-lg flex items-center justify-center bg-blue-100">
              <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm font-medium text-slate-500">Pending Matches</p>
              <p className="text-3xl font-bold text-slate-900 mt-2">{pendingMatches}</p>
            </div>
            <div className="w-10 h-10 rounded-lg flex items-center justify-center bg-sky-100">
              <svg className="w-5 h-5 text-sky-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
            </div>
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm font-medium text-slate-500">Total Matches</p>
              <p className="text-3xl font-bold text-slate-900 mt-2">{matches.length}</p>
            </div>
            <div className="w-10 h-10 rounded-lg flex items-center justify-center bg-green-100">
              <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
        </Card>
      </div>

      {/* Cards Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* Agents Card */}
        <Card className="p-6">
          <div className="flex items-center justify-between mb-5">
            <h2 className="text-lg font-semibold text-slate-900" style={{ fontFamily: 'Plus Jakarta Sans, sans-serif' }}>Your Agents</h2>
            <Link href="/agents/create" className="text-sm text-[#0369A1] hover:text-[#0284C7] font-semibold transition-colors duration-200 cursor-pointer">
              + Create Agent
            </Link>
          </div>
          {agents.length === 0 ? (
            <div className="text-center py-10">
              <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-7 h-7 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
              <p className="text-slate-500 mb-4">No agents created yet</p>
              <Link
                href="/agents/create"
                className="inline-flex items-center gap-2 px-5 py-2.5 bg-[#0369A1] text-white rounded-lg hover:bg-[#0284C7] text-sm font-semibold transition-colors duration-200 cursor-pointer"
              >
                Create Your First Agent
              </Link>
            </div>
          ) : (
            <div className="space-y-3">
              {agents.slice(0, 3).map((agent) => (
                <Card key={agent.id} hover className="p-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <StatusDot status={agent.status === 'active' ? 'active' : 'idle'} size="md" />
                    <div>
                      <p className="font-semibold text-slate-900 capitalize">{agent.type}</p>
                      <p className="text-sm text-slate-500">Status: {agent.status}</p>
                    </div>
                  </div>
                  <span className={`px-3 py-1 text-xs font-semibold rounded-full transition-colors duration-200 ${
                    agent.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-slate-200 text-slate-600'
                  }`}>
                    {agent.status}
                  </span>
                </Card>
              ))}
            </div>
          )}
        </Card>

        {/* Matches Card */}
        <Card className="p-6">
          <div className="flex items-center justify-between mb-5">
            <h2 className="text-lg font-semibold text-slate-900" style={{ fontFamily: 'Plus Jakarta Sans, sans-serif' }}>Recent Matches</h2>
            <Link href="/matches" className="text-sm text-[#0369A1] hover:text-[#0284C7] font-semibold transition-colors duration-200 cursor-pointer">
              View All
            </Link>
          </div>
          {matches.length === 0 ? (
            <div className="text-center py-10">
              <div className="w-14 h-14 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-7 h-7 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                </svg>
              </div>
              <p className="text-slate-500">No matches yet</p>
              <p className="text-sm text-slate-400 mt-1">Create an agent to start matching</p>
            </div>
          ) : (
            <div className="space-y-3">
              {matches.slice(0, 3).map((match) => (
                <Card key={match.id} hover className="p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-semibold text-slate-900">Match #{match.id.slice(0, 8)}</p>
                      <p className="text-sm text-slate-500 mt-0.5">Score: {(match.score * 100).toFixed(1)}%</p>
                    </div>
                    <span className={`px-3 py-1 text-xs font-semibold rounded-full transition-colors duration-200 ${
                      match.status === 'mutual_interest' ? 'bg-green-100 text-green-700' :
                      match.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                      'bg-slate-200 text-slate-600'
                    }`}>
                      {match.status.replace('_', ' ')}
                    </span>
                  </div>
                </Card>
              ))}
            </div>
          )}
        </Card>
      </div>

      {/* FAB */}
      <FAB label="Quick Actions" />
    </div>
  );
}
