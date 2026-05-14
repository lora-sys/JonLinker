// User & Auth
export type UserRole = 'seeker' | 'recruiter' | 'admin';

export interface User {
  id: string;
  email: string;
  role: UserRole;
  organization_id?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Agent
export type AgentType = 'seeker' | 'recruiter';
export type AgentStatus = 'active' | 'paused';

export interface Agent {
  id: string;
  user_id: string;
  type: AgentType;
  status: AgentStatus;
  config?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  user?: User;
}

// Organization
export interface Organization {
  id: string;
  name: string;
  admin_user_id: string;
  created_at: string;
}

// Job
export type JobStatus = 'draft' | 'active' | 'paused' | 'filled' | 'closed';

export interface Job {
  id: string;
  agent_id: string;
  structured: JobStructured;
  vector_id?: string;
  status: JobStatus;
  created_at: string;
  updated_at: string;
  agent?: Agent;
}

export interface JobStructured {
  title: string;
  description: string;
  requirements: string[];
  location?: string;
  salary_range?: {
    min: number;
    max: number;
    currency: string;
  };
  benefits?: string[];
  work_type?: 'remote' | 'hybrid' | 'onsite';
  experience_level?: 'entry' | 'mid' | 'senior' | 'lead';
}

// Resume
export interface Resume {
  id: string;
  agent_id: string;
  structured: ResumeStructured;
  vector_id?: string;
  created_at: string;
  updated_at: string;
  agent?: Agent;
}

export interface ResumeStructured {
  summary?: string;
  experience: WorkExperience[];
  education: Education[];
  skills: string[];
  certifications?: string[];
  languages?: string[];
}

export interface WorkExperience {
  company: string;
  title: string;
  start_date: string;
  end_date?: string;
  current: boolean;
  description?: string;
}

export interface Education {
  institution: string;
  degree: string;
  field: string;
  graduation_date?: string;
}

// Match
export type MatchStatus =
  | 'pending'
  | 'mutual_interest'
  | 'negotiating'
  | 'interview_scheduled'
  | 'offer_sent'
  | 'offered'
  | 'hired'
  | 'rejected';

export interface Match {
  id: string;
  seeker_agent_id: string;
  recruiter_agent_id: string;
  job_id: string;
  score: number;
  status: MatchStatus;
  created_at: string;
  updated_at: string;
  seeker_agent?: Agent;
  job?: Job;
}

// Message (A2A XML Protocol)
export interface Message {
  id: string;
  match_id: string;
  sender_agent_id: string;
  content_xml: string;
  intent_type?: string;
  created_at: string;
  match?: Match;
}

// Conversation thread for messages page
export interface Conversation {
  MatchID: string;
  JobTitle: string;
  LastMessage?: Message;
  UnreadCount: number;
  UpdatedAt: string;
}

// Interview
export type InterviewFormat = 'video' | 'phone' | 'onsite';
export type InterviewStatus = 'scheduled' | 'completed' | 'cancelled' | 'rescheduled';

export interface Interview {
  id: string;
  match_id: string;
  scheduled_at: string;
  format: InterviewFormat;
  location?: string;
  status: InterviewStatus;
  feedback?: InterviewFeedback;
  created_at: string;
  updated_at: string;
  match?: Match;
  // Participant names for display
  seeker_name?: string;
  recruiter_name?: string;
}

export interface InterviewFeedback {
  rating?: number;
  notes?: string;
  recommendation?: 'strong_hire' | 'hire' | 'no_hire' | 'strong_no_hire';
}

// Offer
export type OfferStatus = 'pending' | 'accepted' | 'declined' | 'negotiating' | 'withdrawn' | 'expired';

export interface Offer {
  id: string;
  match_id: string;
  compensation: Compensation;
  start_date: string;
  status: OfferStatus;
  expires_at?: string;
  responded_at?: string;
  created_at: string;
  updated_at: string;
  match?: Match;
}

export interface Compensation {
  base_salary: number;
  currency: string;
  bonus?: {
    amount: number;
    description: string;
  };
  equity?: {
    shares: number;
    vesting_period: string;
  };
  benefits?: string[];
}

// Security
export interface SecurityEvent {
  id: string;
  user_id: string;
  action_type: string;
  details?: Record<string, unknown>;
  ip_address?: string;
  created_at: string;
}

// API Response wrappers
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface ApiError {
  error: string;
  details?: Record<string, unknown>;
}

// A2A Protocol Types
export interface A2AMessage {
  header: {
    message_id: string;
    timestamp: string;
    sender_id: string;
    receiver_id: string;
  };
  payload: A2APayload;
}

export interface A2APayload {
  intent: string;
  parameters?: Record<string, unknown>;
  negotiation?: {
    round: number;
    offers?: Offer[];
  };
}
