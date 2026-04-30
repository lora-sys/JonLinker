import { fetchServer } from '@/lib/api-server';
import { JobsContent } from './JobsContent';
import type { Job } from '@/types';

function parseJobStructured(job: Job): Job {
  return {
    ...job,
    structured: typeof job.structured === 'string'
      ? JSON.parse(job.structured)
      : job.structured,
  };
}

export default async function JobsPage() {
  const jobs = await fetchServer<Job[]>('/api/jobs');
  const parsedJobs = (jobs || []).map(parseJobStructured);

  return <JobsContent initialJobs={parsedJobs} />;
}