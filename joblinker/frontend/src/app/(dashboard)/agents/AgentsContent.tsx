'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Users, Plus } from 'lucide-react';
import Card from '@/components/ui/Card';
import Avatar from '@/components/ui/Avatar';
import StatusDot from '@/components/ui/StatusDot';
import Badge from '@/components/ui/Badge';
import SkillTag from '@/components/ui/SkillTag';
import { LoadingSkeleton, ErrorState, EmptyState } from '@/components/ui';
import { ErrorBoundary } from '@/components/ui/ErrorBoundary';
import { motion } from 'framer-motion';
import type { Agent } from '@/types';

interface AgentWithSkills extends Agent {
  config?: {
    skills?: string[];
    [key: string]: unknown;
  };
}

function AgentsLoadingSkeleton() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      <LoadingSkeleton count={2} variant="card" />
    </div>
  );
}

function AgentCard({ agent, index }: { agent: AgentWithSkills; index: number }) {
  const skills = agent.config?.skills || [];

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.1 }}
    >
      <Card hover className="p-5 bg-white/80 backdrop-blur-xl border border-white/20 cursor-pointer">
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
        {skills.length > 0 && (
          <div className="flex flex-wrap gap-1.5 mt-3">
            {skills.slice(0, 4).map((skill) => (
              <SkillTag key={skill}>{skill}</SkillTag>
            ))}
            {skills.length > 4 && (
              <span className="px-2 py-1 text-xs bg-slate-100 text-slate-500 rounded-full">
                +{skills.length - 4}
              </span>
            )}
          </div>
        )}
      </Card>
    </motion.div>
  );
}

function AgentsEmptyState() {
  return (
    <EmptyState
      icon={<Users className="w-10 h-10 text-blue-600" />}
      title="No agents yet"
      description="Create your first agent to get started"
      actionLabel="Create Agent"
      href="/agents/create"
    />
  );
}

export function AgentsContent({ initialAgents, isLoading = false, initialError = null }: { initialAgents: AgentWithSkills[]; isLoading?: boolean; initialError?: string | null }) {
  const [agents, setAgents] = useState<AgentWithSkills[]>(initialAgents);
  const [error, setError] = useState<string | null>(initialError);

  // Update when initialAgents changes (after API fetch completes)
  useEffect(() => {
    setAgents(initialAgents);
    if (initialError) {
      setError(initialError);
    }
  }, [initialAgents, initialError]);

  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
        <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-3xl font-bold text-slate-900 flex items-center gap-3">
                <Users className="w-8 h-8 text-blue-600" />
                Your Agents
              </h1>
              <p className="text-slate-600 mt-1">Manage your AI recruitment agents</p>
            </div>
            <Link href="/agents/create" className="flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl hover:from-blue-700 hover:to-sky-600 font-semibold text-sm transition-all duration-200 shadow-lg shadow-blue-600/25 cursor-pointer">
              <Plus className="w-4 h-4" />
              Create Agent
            </Link>
          </div>

          {isLoading ? (
            <AgentsLoadingSkeleton />
          ) : error ? (
            <ErrorState message={error} onRetry={() => setError(null)} />
          ) : agents.length === 0 ? (
            <AgentsEmptyState />
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {agents.map((agent, index) => (
                <AgentCard key={agent.id} agent={agent} index={index} />
              ))}
            </div>
          )}
        </div>
      </div>
    </ErrorBoundary>
  );
}