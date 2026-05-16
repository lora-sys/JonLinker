import type {
  Agent,
  AuthResponse,
  Interview,
  Job,
  Match,
  Message,
  Offer,
  Organization,
  PaginatedResponse,
  Resume,
  SecurityEvent,
  User,
} from './index'

// Auth
export interface RegisterRequest {
  email: string
  password: string
  role: 'seeker' | 'recruiter'
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RefreshTokenRequest {
  refresh_token: string
}

// Agent
export interface CreateAgentRequest {
  type: 'seeker' | 'recruiter'
  config?: Record<string, unknown>
}

export interface UpdateAgentRequest {
  status?: 'active' | 'paused'
  config?: Record<string, unknown>
}

// Job
export interface CreateJobRequest {
  structured: {
    title: string
    description: string
    requirements: string[]
    location?: string
    salary_range?: {
      min: number
      max: number
      currency: string
    }
    benefits?: string[]
    work_type?: 'remote' | 'hybrid' | 'onsite'
    experience_level?: 'entry' | 'mid' | 'senior' | 'lead'
  }
}

export interface UpdateJobRequest {
  status?: 'draft' | 'active' | 'paused' | 'filled' | 'closed'
  structured?: CreateJobRequest['structured']
}

export interface ListJobsQuery {
  status?: string
  limit?: number
  offset?: number
}

// Resume
export interface CreateResumeRequest {
  structured: {
    summary?: string
    experience: Array<{
      company: string
      title: string
      start_date: string
      end_date?: string
      current?: boolean
      description?: string
    }>
    education: Array<{
      institution: string
      degree: string
      field: string
      graduation_date?: string
    }>
    skills: string[]
    certifications?: string[]
    languages?: string[]
  }
}

// Match
export interface ListMatchesQuery {
  status?: string
  limit?: number
  offset?: number
}

export interface ConfirmMatchRequest {
  interest_level: 'high' | 'medium' | 'low'
}

// Interview
export interface CreateInterviewRequest {
  match_id: string
  scheduled_at: string
  format: 'video' | 'phone' | 'onsite'
  location?: string
}

export interface UpdateInterviewRequest {
  status?: 'scheduled' | 'completed' | 'cancelled' | 'rescheduled'
  scheduled_at?: string
  location?: string
  feedback?: {
    rating?: number
    notes?: string
    recommendation?: 'strong_hire' | 'hire' | 'no_hire' | 'strong_no_hire'
  }
}

// Offer
export interface CreateOfferRequest {
  match_id: string
  compensation: {
    base_salary: number
    currency: string
    bonus?: {
      amount: number
      description: string
    }
    equity?: {
      shares: number
      vesting_period: string
    }
    benefits?: string[]
  }
  start_date: string
}

export interface RespondToOfferRequest {
  response: 'accept' | 'decline' | 'negotiate'
}

// Message
export interface SendMessageRequest {
  match_id: string
  content_xml: string
  intent_type?: string
}

// Privacy
export interface ExportDataResponse {
  user: User
  agent?: Agent
  organization?: Organization
  matches: Match[]
  messages: Message[]
  interviews: Interview[]
  offers: Offer[]
  export_date: string
}

// API Response types
export type AuthResponseOk = AuthResponse
export type AgentResponseOk = Agent
export type JobResponseOk = Job
export type ResumeResponseOk = Resume
export type MatchResponseOk = Match
export type InterviewResponseOk = Interview
export type OfferResponseOk = Offer
export type SecurityEventResponseOk = SecurityEvent
export type OrganizationResponseOk = Organization

export type AgentsResponseOk = PaginatedResponse<Agent>
export type JobsResponseOk = PaginatedResponse<Job>
export type ResumesResponseOk = PaginatedResponse<Resume>
export type MatchesResponseOk = PaginatedResponse<Match>
export type InterviewsResponseOk = PaginatedResponse<Interview>
export type OffersResponseOk = PaginatedResponse<Offer>
export type OrganizationsResponseOk = PaginatedResponse<Organization>
