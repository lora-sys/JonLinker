import { fetchServer } from '@/lib/api-server';
import { InterviewsContent } from './InterviewsContent';
import type { Interview } from '@/types';

export default async function InterviewsPage() {
  const interviews = await fetchServer<Interview[]>('/api/interviews');

  return <InterviewsContent initialInterviews={interviews || []} />;
}