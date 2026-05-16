import type { Match } from '@/types'

import { fetchServer } from '@/lib/api-server'

import { MatchesContent } from './MatchesContent'

export default async function MatchesPage() {
  const matches = await fetchServer<Match[]>('/api/matches')

  return <MatchesContent initialMatches={matches || []} />
}
