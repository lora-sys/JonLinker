'use client';

import { useState } from 'react';
import { MetricsPanel } from '@/components/admin/MetricsPanel';
import { JobManagement } from '@/components/admin/JobManagement';
import { AgentManagement } from '@/components/admin/AgentManagement';

interface DashboardData {
  metrics: {
    totalAgents: number;
    activeJobs: number;
    pendingMatches: number;
    completedOffers: number;
    totalInterviews: number;
  };
  jobs: Array<{
    id: string;
    title: string;
    status: string;
    created_at: string;
  }>;
  agents: Array<{
    id: string;
    type: string;
    status: string;
    created_at: string;
    user?: { email?: string };
  }>;
}

export default function AdminPage() {
  const [data] = useState<DashboardData>({
    metrics: {
      totalAgents: 24,
      activeJobs: 8,
      pendingMatches: 12,
      completedOffers: 45,
      totalInterviews: 67,
    },
    jobs: [],
    agents: [],
  });

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        <section className="mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Overview</h2>
          <MetricsPanel metrics={data.metrics} />
        </section>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <section>
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Jobs</h2>
            <JobManagement jobs={data.jobs} />
          </section>

          <section>
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Agents</h2>
            <AgentManagement agents={data.agents} />
          </section>
        </div>
      </main>
    </div>
  );
}
