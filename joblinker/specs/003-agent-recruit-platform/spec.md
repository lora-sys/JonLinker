# Feature Specification: A2A Agent Recruitment Platform

**Feature Branch**: `003-agent-recruit-platform`
**Created**: 2026-04-22
**Status**: Draft
**Input**: Feature plan from `../plan.md`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agent Creation and Registration (Priority: P1)

As a job seeker (Seeker), I want to create a personalized job search Agent that represents me so that it can autonomously find and match with suitable positions on my behalf.

**Why this priority**: Without an Agent, there is no A2A platform. The Agent is the core abstraction that automates all recruitment interactions.

**Independent Test**: Can be fully tested by a user creating an Agent and verifying it appears in the system, is active, and can receive/match jobs without further manual input from the user.

**Acceptance Scenarios**:

1. **Given** a user has registered an account, **When** they create a job seeker Agent, **Then** the system creates an Agent associated with their account and sets it to active status
2. **Given** a user creates an Agent, **When** they provide information via manual entry, AI generation, or file upload, **Then** the system structures, encrypts, and stores the information locally
3. **Given** a user wants to pause their job search, **When** they pause their Agent, **Then** the Agent stops receiving new matches and notifications

---

### User Story 2 - Automated Matching (Priority: P1)

As a job seeker Agent, I want to be automatically matched with suitable job positions based on my skills, experience, and preferences so that relevant opportunities surface without manual searching.

**Why this priority**: Matching is the core value proposition. Automated vector-based matching replaces hours of manual job hunting and resume posting.

**Independent Test**: Can be fully tested by creating a Seeker Agent with specific qualifications and a Recruiter Agent with matching job requirements, then verifying matches are generated automatically.

**Acceptance Scenarios**:

1. **Given** a Seeker Agent is active with complete profile information, **When** a matching job is posted, **Then** the system generates a match record with a compatibility score
2. **Given** a match is generated, **When** both Agents review the match, **Then** the match status progresses to "Mutual Interest" if both agree
3. **Given** a match has "Mutual Interest" status, **When** the human users confirm, **Then** the system initiates the professional dialogue phase

---

### User Story 3 - A2A Professional Dialogue (Priority: P1)

As a Seeker Agent and Recruiter Agent pair, I want to automatically conduct professional negotiations about salary, job requirements, and interview timing so that human users only need to confirm the outcomes.

**Why this priority**: The A2A dialogue is what transforms this from a job board into an autonomous recruitment system. It eliminates the back-and-forth that delays hiring.

**Independent Test**: Can be fully tested by simulating two Agents with aligned interests, then verifying they complete a professional dialogue covering salary, requirements, and interview scheduling without human intervention.

**Acceptance Scenarios**:

1. **Given** two Agents have mutual interest, **When** the dialogue begins, **Then** messages are exchanged via XML protocol and real-time updates are pushed to both human users
2. **Given** the Seeker Agent proposes a salary range and the Recruiter Agent counter-offers, **When** they negotiate, **Then** the dialogue produces an agreed compensation range that both Agents accept
3. **Given** the Agents negotiate interview time, **When** they reach consensus, **Then** the system generates an interview invitation for human confirmation

---

### User Story 4 - Interview Scheduling (Priority: P2)

As a Seeker Agent, I want to coordinate interview times with the Recruiter Agent so that human users receive a confirmed interview slot without scheduling friction.

**Why this priority**: Scheduling is the biggest friction point in recruitment. Automating it accelerates the hiring timeline significantly.

**Independent Test**: Can be fully tested by two Agents exchanging availability and generating a confirmed interview without human intervention in the scheduling phase.

**Acceptance Scenarios**:

1. **Given** Agents have completed initial negotiation, **When** they begin scheduling, **Then** the Seeker Agent proposes available time slots based on user preferences
2. **Given** the Recruiter Agent receives time proposals, **When** it finds a mutually available slot, **Then** it generates an interview invitation with date, time, and format
3. **Given** the human user confirms the interview invitation, **When** the confirmation is received, **Then** calendar invitations are sent to all parties with reminders 24 hours before

---

### User Story 5 - Offer Generation and Confirmation (Priority: P2)

As a Recruiter Agent, I want to generate and send an Offer after successful interview, and as a Seeker Agent, I want to receive and respond to the Offer so that the hiring process concludes efficiently.

**Why this priority**: The offer stage is where deals close. Automating it ensures prompt, professional responses and reduces time-to-fill.

**Independent Test**: Can be fully tested by an employer sending an offer through the Recruiter Agent, then verifying the Seeker Agent receives it clearly and can respond accept/decline/negotiate.

**Acceptance Scenarios**:

1. **Given** an employer decides to make an offer, **When** they initiate via Recruiter Agent, **Then** the system generates an Offer with compensation, start date, and key terms
2. **Given** a Seeker Agent receives an Offer, **When** the human user reviews it, **Then** they can accept, decline, or request negotiation through the Agent
3. **Given** a Seeker accepts an Offer, **When** confirmation is sent, **Then** both Agents mark the position as filled and the process closes

---

### User Story 6 - Privacy Protection (Priority: P1)

As a job seeker, I want my sensitive information (original resume, contact details) to be stored locally and encrypted so that my privacy is protected even if the platform is compromised.

**Why this priority**: Privacy concerns are the primary reason candidates hesitate to use recruitment platforms. Without strong privacy guarantees, users withhold information that reduces match quality.

**Independent Test**: Can be fully tested by verifying that sensitive fields are encrypted at rest, never transmitted in plaintext, and can be exported or permanently deleted on user request.

**Acceptance Scenarios**:

1. **Given** a user uploads their original resume, **When** the system processes it, **Then** the raw file is stored encrypted locally and the plaintext is never sent to external systems
2. **Given** a user requests their data be exported, **When** the system processes the request, **Then** all user data is bundled and delivered within 24 hours
3. **Given** a user requests account deletion, **When** the system confirms the request, **Then** all personal data is permanently deleted within 30 days and cannot be recovered

---

### User Story 7 - Enterprise Administration (Priority: P3)

As an enterprise administrator, I want to view recruitment progress, manage job postings, and configure recruiting Agents so that I have visibility and control over the organization's hiring pipeline.

**Why this priority**: Enterprise admins need oversight of all recruiting activity. Without admin controls, organizations cannot govern their recruitment operations.

**Independent Test**: Can be fully tested by an admin viewing all active job postings, seeing match statistics, and adjusting Agent configurations without affecting other users.

**Acceptance Scenarios**:

1. **Given** an admin views the enterprise dashboard, **When** they access hiring metrics, **Then** they see active positions, match rates, interview progress, and offer status
2. **Given** an admin wants to pause all recruiting, **When** they deactivate all company Agents, **Then** all active job postings are paused and no new matches are generated
3. **Given** an admin wants to modify job requirements, **When** they update a job posting, **Then** the Recruiter Agent reflects the changes in future negotiations

---

### Edge Cases

- What happens when a Seeker Agent and Recruiter Agent cannot reach mutual interest after maximum negotiation rounds?
- How does the system handle salary expectations that are completely misaligned with job budget?
- What occurs when a user deletes their account while an interview is scheduled?
- How does the system behave when an offer is extended but the candidate becomes unresponsive?
- What happens when two Seeker Agents match with the same Recruiter Agent for the same position?
- How does the system handle vector matching when profile or job data is sparse (few skills listed)?
- What occurs when a Recruiter Agent receives interview feedback but the human recruiter does not provide ratings?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to create job seeker Agents and recruiter Agents associated with their account
- **FR-002**: System MUST support three information input modes: manual entry, AI one-sentence generation, file parsing (PDF/Word)
- **FR-003**: System MUST structure user information into standardized format after input
- **FR-004**: System MUST encrypt and store sensitive data locally with user-controlled keys
- **FR-005**: System MUST generate vector representations of user profiles and job postings for similarity matching
- **FR-006**: System MUST automatically match active Seeker Agents with active job postings based on vector similarity
- **FR-007**: System MUST calculate and present match compatibility scores for each candidate-position pair
- **FR-008**: System MUST support A2A communication between Seeker and Recruiter Agents via structured protocol
- **FR-009**: System MUST push real-time dialogue updates to human users during A2A conversations
- **FR-010**: System MUST generate interview invitations when Agents reach consensus on time slot
- **FR-011**: System MUST send calendar invitations and reminders 24 hours before scheduled interviews
- **FR-012**: System MUST generate job Offers containing compensation, start date, and key terms
- **FR-013**: System MUST track Offer status and send notifications when status changes
- **FR-014**: System MUST allow users to export all their personal data on request
- **FR-015**: System MUST permanently delete user data within 30 days of account deletion request
- **FR-016**: System MUST support Agent pause/activate to control matching activity
- **FR-017**: System MUST manage Agent state via finite state machine (FSM) for consistent behavior
- **FR-018**: System MUST maintain conversation context to ensure dialogue coherence across multiple exchanges
- **FR-019**: System MUST provide enterprise admins visibility into recruitment metrics and Agent status
- **FR-020**: System MUST audit all Agent actions for compliance and dispute resolution

### Key Entities

- **User (USERS)**: Platform user with account, role (Seeker/Recruiter/Admin), and organization association
- **Agent (AGENTS)**: Autonomous representation of user; has type (job_seeker/recruiter), status, configuration, and associated user
- **Resume (RESUMES)**: Structured candidate data including skills, experience, education, preferences; vector representation for matching
- **Job (JOBS)**: Position requirements including responsibilities, qualifications, compensation range, status
- **Match (MATCHES)**: Record of candidate-position pairing with compatibility score and status (pending/mutual_interest/negotiating/offered/hired/rejected)
- **Message (MESSAGES)**: A2A dialogue content in structured format between Agents
- **Interview (INTERVIEWS)**: Scheduled interview with participants, time, format, location/link, status
- **Offer (OFFERS)**: Formal job offer with compensation details, start date, terms, status (pending/accepted/declined/negotiating)
- **Security Event (SECURITY_EVENTS)**: Audit log entry for compliance; includes action type, user, timestamp, IP, details

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Match generation occurs within 4 hours of new job posting going live
- **SC-002**: 85% of matched Seeker Agents progress to "Mutual Interest" within 7 days when match score exceeds threshold
- **SC-003**: A2A negotiation completes to interview scheduling within 72 hours of mutual interest confirmation
- **SC-004**: Interview scheduling completes within 48 hours of employer initiating scheduling process
- **SC-005**: Offer response time averages under 72 hours from delivery to final decision
- **SC-006**: Users receive real-time dialogue updates within 5 seconds of Agent message generation
- **SC-007**: Data export requests are fulfilled within 24 hours
- **SC-008**: Data deletion is complete within 30 days of confirmed deletion request
- **SC-009**: System maintains audit log for 100% of Agent actions
- **SC-010**: 95% of users report confidence that their private data is protected
- **SC-011**: Enterprise admins can view recruitment metrics within 2 seconds of dashboard access

## Assumptions

- Users access the platform via web browsers on desktop and mobile devices
- Seeker Agents operate within user-defined boundaries (salary range, location preferences, job types)
- Human users set initial Agent parameters and confirm key decisions (mutual interest, interview, offer)
- AI one-sentence generation uses locally-run LLM with user consent for data processing
- File parsing supports PDF and Word formats commonly used for resumes
- Calendar integration uses standard iCalendar format supporting major email providers
- Platform operates in jurisdictions where automated recruitment matching is permitted
- Vector matching uses industry-standard embedding algorithms for skills and requirements
- A2A communication protocol uses XML format for structured message exchange
- Agents use finite state machines to manage consistent state transitions
- Conversation context is maintained per match for coherent multi-turn negotiation
- Audit logs are retained for minimum 2 years for compliance purposes