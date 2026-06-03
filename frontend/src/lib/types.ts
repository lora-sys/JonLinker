export type Job = {
  title: string; company: string; location: string; salary: string;
  url: string; description: string; tags: string[]; source: string;
};
export type RankedJob = Job & { match_score: number; summary: string; highlights: string[] };
export type Application = {
  job_title: string; company: string; cover_letter: string;
  resume_md: string; highlights: string[]; generated_at: string;
};
