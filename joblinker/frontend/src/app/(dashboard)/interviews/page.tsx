import type { Interview } from '@/types'

import { fetchServer } from '@/lib/api-server'

import { InterviewsContent } from './InterviewsContent'

export default async function InterviewsPage() {
  const interviews = await fetchServer<Interview[]>('/api/interviews')

  return <InterviewsContent initialInterviews={interviews || []} />
}
