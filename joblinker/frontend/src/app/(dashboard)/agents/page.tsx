import { AgentsContent } from './AgentsContent';
import { fetchServerWithResult } from '@/lib/api-server';
import type { Agent } from '@/types';

export default async function AgentsPage() {
  const { data: agents, error } = await fetchServerWithResult<Agent[]>('/api/agents');

  return (
    <AgentsContent
      initialAgents={Array.isArray(agents) ? agents : []}
      isLoading={false}
      initialError={error}
    />
  );
}