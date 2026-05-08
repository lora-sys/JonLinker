import { AgentsContent } from './AgentsContent';
import { getServerToken } from '@/lib/auth-utils';
import type { Agent } from '@/types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function getAgents(): Promise<{ agents: Agent[]; error: string | null }> {
  const token = await getServerToken();

  if (!token) {
    return { agents: [], error: 'Not authenticated' };
  }

  try {
    const response = await fetch(`${API_BASE}/api/agents`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      cache: 'no-store',
    });

    if (!response.ok) {
      return { agents: [], error: `Failed to fetch: ${response.status}` };
    }

    const data = await response.json();
    return { agents: Array.isArray(data) ? data : data.agents || [], error: null };
  } catch (error) {
    console.error('Failed to fetch agents:', error);
    return { agents: [], error: 'Failed to connect to backend' };
  }
}

export default async function AgentsPage() {
  const { agents, error } = await getAgents();

  return <AgentsContent initialAgents={agents} isLoading={false} initialError={error} />;
}