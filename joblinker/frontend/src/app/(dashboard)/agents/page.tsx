import type { Agent } from '@/shared/types'

import { fetchServerWithResult } from '@/lib/api-server'

import { AgentsContent } from './AgentsContent'

export default async function AgentsPage() {
  const { data: agents, error } = await fetchServerWithResult<Agent[]>('/api/agents')

  return (
    <AgentsContent
      initialAgents={Array.isArray(agents) ? agents : []}
      isLoading={false}
      initialError={error}
    />
  )
}
