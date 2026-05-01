import { Activity, AlertTriangle, CheckCircle, Users, Shield, Zap } from 'lucide-react';
import { redirect } from 'next/navigation';
import { cookies } from 'next/headers';

interface MetricsSummary {
  active_seekers: number;
  active_recruiters: number;
  active_conversations: number;
  total_errors_unresolved: number;
}

interface AuditLog {
  id: string;
  agent_id: string;
  match_id: string;
  event_type: string;
  event_data: Record<string, unknown>;
  timestamp: string;
}

interface ErrorEvent {
  id: string;
  error_type: string;
  error_message: string;
  stack_trace?: string;
  context?: Record<string, unknown>;
  resolved: boolean;
  timestamp: string;
}

interface AgentMetric {
  id: string;
  agent_id: string;
  agent_type: string;
  messages_processed: number;
  errors_count: number;
  last_heartbeat: string;
}

interface DashboardData {
  metrics: MetricsSummary;
  auditLogs: AuditLog[];
  errors: ErrorEvent[];
  agentMetrics: AgentMetric[];
}

async function getAdminData(): Promise<DashboardData | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get('joblinker-auth')?.value;

  if (!token) {
    return null;
  }

  try {
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    const [metricsRes, auditRes, errorsRes, agentMetricsRes] = await Promise.all([
      fetch(`${backendUrl}/api/admin/metrics`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      }),
      fetch(`${backendUrl}/api/admin/audit?limit=50`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      }),
      fetch(`${backendUrl}/api/admin/errors?limit=50`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      }),
      fetch(`${backendUrl}/api/admin/agent-metrics`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      }),
    ]);

    if (!metricsRes.ok) return null;

    const [metrics, auditLogs, errors, agentMetrics] = await Promise.all([
      metricsRes.json(),
      auditRes.json().catch(() => []),
      errorsRes.json().catch(() => []),
      agentMetricsRes.json().catch(() => []),
    ]);

    return { metrics, auditLogs, errors, agentMetrics };
  } catch {
    return null;
  }
}

export default async function AdminDashboard() {
  const data = await getAdminData();

  if (!data) {
    redirect('/login?redirect=/admin');
  }

  const { metrics, auditLogs, errors, agentMetrics } = data;

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      <div className="max-w-7xl mx-auto p-6 lg:p-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white flex items-center gap-3">
            <div className="p-2 bg-blue-500/20 rounded-xl">
              <Activity className="w-8 h-8 text-blue-400" />
            </div>
            Admin Dashboard
          </h1>
          <p className="text-slate-400 mt-2">Production observability and monitoring</p>
        </div>

        {/* Metrics Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 lg:gap-6 mb-8">
          <MetricCard
            title="Active Seekers"
            value={metrics.active_seekers}
            icon={<Users className="w-6 h-6" />}
            color="blue"
          />
          <MetricCard
            title="Active Recruiters"
            value={metrics.active_recruiters}
            icon={<Shield className="w-6 h-6" />}
            color="emerald"
          />
          <MetricCard
            title="Active Conversations"
            value={metrics.active_conversations}
            icon={<Zap className="w-6 h-6" />}
            color="amber"
          />
          <MetricCard
            title="Unresolved Errors"
            value={metrics.total_errors_unresolved}
            icon={<AlertTriangle className="w-6 h-6" />}
            color="rose"
          />
        </div>

        {/* Two Column Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Audit Logs */}
          <div className="bg-slate-800/50 backdrop-blur-xl rounded-2xl p-6 border border-slate-700">
            <h3 className="text-lg font-semibold text-white mb-4">Recent Audit Logs</h3>
            <div className="space-y-3">
              {auditLogs.length > 0 ? auditLogs.slice(0, 5).map(log => (
                <div key={log.id} className="flex items-start gap-3 text-sm">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${
                    log.event_type === 'error' ? 'bg-rose-500/20 text-rose-400' :
                    log.event_type === 'tool_call' ? 'bg-blue-500/20 text-blue-400' :
                    'bg-emerald-500/20 text-emerald-400'
                  }`}>
                    {log.event_type}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-slate-400 text-xs truncate">
                      {log.event_data?.intent ? `intent: ${log.event_data.intent}` : JSON.stringify(log.event_data)}
                    </p>
                  </div>
                  <span className="text-slate-500 text-xs">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </span>
                </div>
              )) : (
                <p className="text-slate-500 text-sm">No audit logs yet</p>
              )}
            </div>
          </div>

          {/* Errors */}
          <div className="bg-slate-800/50 backdrop-blur-xl rounded-2xl p-6 border border-slate-700">
            <h3 className="text-lg font-semibold text-white mb-4">Error Status</h3>
            <div className="space-y-3">
              {errors.length > 0 ? errors.slice(0, 5).map(err => (
                <div key={err.id} className="flex items-start gap-3 text-sm">
                  <AlertTriangle className="w-4 h-4 text-rose-500 mt-0.5" />
                  <div className="flex-1 min-w-0">
                    <p className="text-slate-300 truncate">{err.error_message}</p>
                    <p className="text-slate-500 text-xs">{err.error_type}</p>
                  </div>
                </div>
              )) : (
                <div className="text-center py-8">
                  <CheckCircle className="w-12 h-12 text-emerald-500 mx-auto mb-2" />
                  <p className="text-slate-400">No unresolved errors</p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Agent Metrics Table */}
        <div className="bg-slate-800/50 backdrop-blur-xl rounded-2xl border border-slate-700 overflow-hidden">
          <div className="p-4 border-b border-slate-700">
            <h3 className="font-semibold text-white">Agent Performance</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-700/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400 uppercase">Agent</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-slate-400 uppercase">Type</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-slate-400 uppercase">Messages</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-slate-400 uppercase">Errors</th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-slate-400 uppercase">Last Heartbeat</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {agentMetrics.length > 0 ? agentMetrics.map(metric => (
                  <tr key={metric.id}>
                    <td className="px-4 py-3 text-sm text-slate-300 font-mono">
                      {metric.agent_id.slice(0, 8)}...
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span className={`px-2 py-0.5 rounded text-xs ${
                        metric.agent_type === 'seeker' ? 'bg-blue-500/20 text-blue-400' : 'bg-emerald-500/20 text-emerald-400'
                      }`}>
                        {metric.agent_type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-300 text-right">{metric.messages_processed}</td>
                    <td className="px-4 py-3 text-sm text-right">
                      <span className={metric.errors_count > 0 ? 'text-rose-400' : 'text-slate-400'}>
                        {metric.errors_count}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-500 text-right">
                      {metric.last_heartbeat ? new Date(metric.last_heartbeat).toLocaleString() : 'N/A'}
                    </td>
                  </tr>
                )) : (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-slate-500">
                      No agent metrics recorded yet
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Full Audit Log */}
        <div className="mt-8 bg-slate-800/50 backdrop-blur-xl rounded-2xl border border-slate-700 overflow-hidden">
          <div className="p-4 border-b border-slate-700">
            <h3 className="font-semibold text-white">All Audit Logs ({auditLogs.length})</h3>
          </div>
          <div className="divide-y divide-slate-700 max-h-96 overflow-y-auto">
            {auditLogs.length > 0 ? auditLogs.map(log => (
              <div key={log.id} className="p-4 flex items-start gap-4">
                <span className={`px-2 py-1 rounded text-xs font-medium shrink-0 ${
                  log.event_type === 'error' ? 'bg-rose-500/20 text-rose-400' :
                  log.event_type === 'tool_call' ? 'bg-blue-500/20 text-blue-400' :
                  log.event_type === 'message_sent' ? 'bg-emerald-500/20 text-emerald-400' :
                  'bg-slate-600 text-slate-300'
                }`}>
                  {log.event_type}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-slate-300 font-mono">
                    {JSON.stringify(log.event_data)}
                  </p>
                  <p className="text-xs text-slate-500 mt-1">
                    Agent: {log.agent_id.slice(0, 8)}... | Match: {log.match_id.slice(0, 8)}...
                  </p>
                </div>
                <span className="text-xs text-slate-500 shrink-0">
                  {new Date(log.timestamp).toLocaleString()}
                </span>
              </div>
            )) : (
              <div className="p-8 text-center">
                <p className="text-slate-500">No audit logs yet</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function MetricCard({ title, value, icon, color }: {
  title: string;
  value: number;
  icon: React.ReactNode;
  color: 'blue' | 'emerald' | 'amber' | 'rose';
}) {
  const colorClasses = {
    blue: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
    emerald: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    amber: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    rose: 'bg-rose-500/10 text-rose-400 border-rose-500/20',
  };

  return (
    <div className={`bg-slate-800/50 backdrop-blur-xl rounded-2xl p-6 border ${colorClasses[color]}`}>
      <div className="flex items-center justify-between mb-4">
        <div className={`p-3 rounded-xl ${colorClasses[color]}`}>
          {icon}
        </div>
        <span className="text-4xl font-bold text-white">{value}</span>
      </div>
      <h3 className="text-sm font-medium text-slate-400">{title}</h3>
    </div>
  );
}
