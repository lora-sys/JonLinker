'use client';

import Link from 'next/link';
import { useAgent } from '@/hooks/useAgent';
import Card from '@/components/ui/Card';
import Avatar from '@/components/ui/Avatar';
import StatusDot from '@/components/ui/StatusDot';
import Badge from '@/components/ui/Badge';
import SkillTag from '@/components/ui/SkillTag';

export default function AgentsPage() {
  const { agents, isLoading } = useAgent();

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-slate-900">Your Agents</h1>
            <p className="text-slate-600 mt-1">Manage your AI recruitment agents</p>
          </div>
          <Link href="/agents/create" className="px-5 py-2.5 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-semibold text-sm transition-all duration-200 shadow-lg shadow-blue-600/25 cursor-pointer">
            + Create Agent
          </Link>
        </div>

        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[1, 2].map((i) => (
              <Card key={i} className="p-5 animate-pulse">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 bg-slate-200 rounded-xl"></div>
                    <div className="w-24 h-5 bg-slate-200 rounded"></div>
                  </div>
                  <div className="w-16 h-6 bg-slate-200 rounded-full"></div>
                </div>
              </Card>
            ))}
          </div>
        ) : agents.length === 0 ? (
          <Card className="p-10 text-center bg-white/80 backdrop-blur-xl border border-white/20">
            <div className="w-16 h-16 bg-gradient-to-br from-blue-100 to-sky-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <h2 className="text-lg font-semibold text-slate-900 mb-2">No agents yet</h2>
            <p className="text-slate-500 mb-5">Create your first agent to get started</p>
            <Link href="/agents/create" className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-semibold text-sm transition-all duration-200 shadow-lg shadow-blue-600/25 cursor-pointer">
              Create Agent
            </Link>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {agents.map((agent) => (
              <Card key={agent.id} hover className="p-5 bg-white/80 backdrop-blur-xl border border-white/20">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <Avatar type={agent.type === 'seeker' ? 'seeker' : 'recruiter'} size="lg" />
                    <div>
                      <p className="font-semibold text-slate-900 capitalize">{agent.type}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <StatusDot status={agent.status === 'active' ? 'active' : 'idle'} size="sm" />
                        <span className="text-sm text-slate-500">{agent.status}</span>
                      </div>
                    </div>
                  </div>
                  <Badge status={agent.status === 'active' ? 'active' : agent.status === 'paused' ? 'paused' : 'pending'}>
                    {agent.status}
                  </Badge>
                </div>
                {(agent as any).skills && (
                  <div className="flex flex-wrap gap-1.5 mt-3">
                    {(agent as any).skills.slice(0, 4).map((skill: string) => (
                      <SkillTag key={skill}>{skill}</SkillTag>
                    ))}
                  </div>
                )}
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
