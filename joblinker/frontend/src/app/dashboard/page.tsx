'use client';

import { useAuthStore } from '@/stores/auth';
import { useAgentStore } from '@/stores/agent';
import { useMatchStore } from '@/stores/match';
import Link from 'next/link';

export default function DashboardPage() {
  const { user } = useAuthStore();
  const { agents, currentAgent } = useAgentStore();
  const { matches } = useMatchStore();

  const stats = [
    { label: 'Active Agents', value: agents.filter((a) => a.status === 'active').length },
    { label: 'Pending Matches', value: matches.filter((m) => m.status === 'pending').length },
    { label: 'Total Matches', value: matches.length },
  ];

  return (
    <div className="max-w-6xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">
          Welcome back, {user?.email?.split('@')[0] || 'User'}
        </h1>
        <p className="text-gray-600 mt-1">Here&apos;s what&apos;s happening with your recruitment activity</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {stats.map((stat) => (
          <div key={stat.label} className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
            <p className="text-sm text-gray-500 font-medium">{stat.label}</p>
            <p className="text-3xl font-bold text-gray-900 mt-2">{stat.value}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Your Agents</h2>
            <Link href="/agents/create" className="text-sm text-blue-600 hover:text-blue-700 font-medium">
              + Create Agent
            </Link>
          </div>
          {agents.length === 0 ? (
            <div className="text-center py-8">
              <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
              <p className="text-gray-500 mb-4">No agents created yet</p>
              <Link
                href="/agents/create"
                className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
              >
                Create Your First Agent
              </Link>
            </div>
          ) : (
            <div className="space-y-3">
              {agents.slice(0, 3).map((agent) => (
                <div key={agent.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div>
                    <p className="font-medium text-gray-900 capitalize">{agent.type}</p>
                    <p className="text-sm text-gray-500">Status: {agent.status}</p>
                  </div>
                  <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                    agent.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
                  }`}>
                    {agent.status}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Recent Matches</h2>
            <Link href="/matches" className="text-sm text-blue-600 hover:text-blue-700 font-medium">
              View All
            </Link>
          </div>
          {matches.length === 0 ? (
            <div className="text-center py-8">
              <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                </svg>
              </div>
              <p className="text-gray-500">No matches yet</p>
              <p className="text-sm text-gray-400 mt-1">Create an agent to start matching</p>
            </div>
          ) : (
            <div className="space-y-3">
              {matches.slice(0, 3).map((match) => (
                <div key={match.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div>
                    <p className="font-medium text-gray-900">Match #{match.id.slice(0, 8)}</p>
                    <p className="text-sm text-gray-500">Score: {(match.score * 100).toFixed(1)}%</p>
                  </div>
                  <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                    match.status === 'mutual_interest' ? 'bg-green-100 text-green-700' :
                    match.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                    'bg-gray-100 text-gray-700'
                  }`}>
                    {match.status.replace('_', ' ')}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
