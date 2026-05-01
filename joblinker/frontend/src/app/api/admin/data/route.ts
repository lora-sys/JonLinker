import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

export async function GET() {
  const cookieStore = await cookies();
  const authCookie = cookieStore.get('joblinker-auth');

  if (!authCookie?.value) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  try {
    const [metricsRes, auditRes, errorsRes, agentMetricsRes] = await Promise.all([
      fetch(`${backendUrl}/api/admin/metrics`, {
        headers: { Authorization: `Bearer ${authCookie.value}` },
        cache: 'no-store',
      }),
      fetch(`${backendUrl}/api/admin/audit?limit=50`, {
        headers: { Authorization: `Bearer ${authCookie.value}` },
        cache: 'no-store',
      }),
      fetch(`${backendUrl}/api/admin/errors?limit=50`, {
        headers: { Authorization: `Bearer ${authCookie.value}` },
        cache: 'no-store',
      }),
      fetch(`${backendUrl}/api/admin/agent-metrics`, {
        headers: { Authorization: `Bearer ${authCookie.value}` },
        cache: 'no-store',
      }),
    ]);

    if (!metricsRes.ok) return NextResponse.json({ error: 'Backend error' }, { status: 502 });

    const [metrics, auditLogs, errors, agentMetrics] = await Promise.all([
      metricsRes.json().catch(() => ({ active_seekers: 0, active_recruiters: 0, active_conversations: 0, total_errors_unresolved: 0 })),
      auditRes.ok ? auditRes.json().catch(() => []) : Promise.resolve([]),
      errorsRes.ok ? errorsRes.json().catch(() => []) : Promise.resolve([]),
      agentMetricsRes.ok ? agentMetricsRes.json().catch(() => []) : Promise.resolve([]),
    ]);

    return NextResponse.json({ metrics, auditLogs, errors, agentMetrics });
  } catch (err) {
    return NextResponse.json({ error: 'Failed to fetch data' }, { status: 500 });
  }
}
