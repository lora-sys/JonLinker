import { fetchServer } from '@/lib/api-server';
import { MatchesContent } from './MatchesContent';
import type { Match } from '@/types';

export default async function MatchesPage() {
  const matches = await fetchServer<Match[]>('/api/matches');

  return <MatchesContent initialMatches={matches || []} />;
}