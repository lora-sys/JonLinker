export type Job = {
  title: string; company: string; location: string; salary: string;
  url: string; description: string; tags: string[]; source: string;
};
export type RankedJob = Job & { match_score: number; summary: string; highlights: string[] };
export type Application = {
  job_title: string; company: string; cover_letter: string;
  resume_md: string; highlights: string[]; generated_at: string;
};

export interface CandidateProfile {
  name: string;
  title: string;
  skills: string[];
  experience: Experience[];
  education: Education[];
  phone: string;
  email: string;
  summary: string;
  hobbies: string[];
}

export interface Experience {
  company: string;
  title: string;
  duration: string;
  description: string;
}

export interface Education {
  school: string;
  degree: string;
  major: string;
  duration: string;
}

export interface ResumeState {
  complete: boolean;
  message?: string;
  profile?: CandidateProfile;
}

export interface ChatDataTypes extends Record<string, unknown> {
  jobs: RankedJob[];
  application: Application;
  resume: ResumeState;
}
