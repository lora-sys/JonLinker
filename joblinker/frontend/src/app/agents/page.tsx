import { fetchServer } from '@/lib/api-server';
import { AgentsContent } from './AgentsContent';
import type { Agent } from '@/types';

export default async function AgentsPage() {
  const agents = await fetchServer<Agent[]>('/api/agents');
  return <AgentsContent initialAgents={agents || []} />;
}