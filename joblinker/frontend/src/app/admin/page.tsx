import { MetricsPanel } from '@/components/admin/MetricsPanel';
import { JobManagement } from '@/components/admin/JobManagement';
import { AgentManagement } from '@/components/admin/AgentManagement';

const dashboardData = {
  metrics: {
    totalAgents: 24,
    activeJobs: 8,
    pendingMatches: 12,
    completedOffers: 45,
    totalInterviews: 67,
  },
  jobs: [] as Array<{
    id: string;
    title: string;
    status: string;
    created_at: string;
  }>,
  agents: [] as Array<{
    id: string;
    type: string;
    status: string;
    created_at: string;
    user?: { email?: string };
  }>,
};

export default function AdminPage() {

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
      <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <header className="bg-white/80 backdrop-blur-xl border-b border-white/20 shadow-sm rounded-xl mb-6">
          <div className="px-6 py-4">
            <h1 className="text-2xl font-bold text-slate-900">Admin Dashboard</h1>
          </div>
        </header>

        <main className="px-4 sm:px-6 lg:px-8">
          <section className="mb-8">
            <h2 className="text-lg font-semibold text-slate-900 mb-4">Overview</h2>
            <MetricsPanel metrics={dashboardData.metrics} />
          </section>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <section>
              <h2 className="text-lg font-semibold text-slate-900 mb-4">Jobs</h2>
              <JobManagement jobs={dashboardData.jobs} />
            </section>

            <section>
              <h2 className="text-lg font-semibold text-slate-900 mb-4">Agents</h2>
              <AgentManagement agents={dashboardData.agents} />
            </section>
          </div>
        </main>
      </div>
    </div>
  );
}