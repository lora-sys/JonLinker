import type { Interview } from '@/shared/types'

import { fetchServer } from '@/lib/api-server'

import { InterviewsContent } from './InterviewsContent'

export default async function InterviewsPage() {
  const data = await fetchServer<{ interviews: Interview[] }>('/api/interviews')
  const interviews = data?.interviews || []

  return <InterviewsContent initialInterviews={interviews} />
}
