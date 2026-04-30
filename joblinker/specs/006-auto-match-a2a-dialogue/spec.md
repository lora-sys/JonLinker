# Feature Specification: Auto-Match & A2A Agent Dialogue

**Feature Branch**: `006-auto-match-a2a-dialogue`
**Created**: 2026-04-26
**Status**: Draft
**Input**: "系统缺少自动匹配功能、A2A对话和WebSocket实时通信。需要实现：1) seeker浏览职位时自动创建Match 2) Agent之间通过WebSocket实时对话 3) AI自动评估匹配分数 4) 完整的招聘流程自动化"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Auto Job Matching (Priority: P1)

As a job seeker, when I browse jobs, the system automatically evaluates which jobs match my profile and creates Match records so I don't have to manually initiate every connection.

**Why this priority**: Without automatic matching, users must manually discover and apply to every job. This creates friction and reduces platform engagement. Auto-matching is the foundation for all subsequent A2A dialogue.

**Independent Test**: Seeker with an active agent browses jobs → system creates Match records for relevant jobs → matches appear on /matches page within 10 seconds.

**Acceptance Scenarios**:

1. **Given** seeker has an active agent with profile/skills, **When** they visit /jobs page, **Then** system evaluates each job against seeker profile and creates Match records for jobs with score > 0.5
2. **Given** a Match is created, **When** it appears on /matches, **Then** it shows the AI-calculated compatibility score and match reasoning
3. **Given** seeker views a specific job and clicks "I'm Interested", **When** action confirmed, **Then** Match status changes from "pending" to "expressed_interest"

---

### User Story 2 - A2A Real-Time Dialogue (Priority: P1)

As a recruiter agent, when I receive a new match notification, I automatically send an introduction message to the seeker agent, and receive responses in real-time via WebSocket.

**Why this priority**: A2A dialogue is the core value proposition - agents talking to each other to negotiate job opportunities. Without real-time WebSocket communication, the system is just a static job board.

**Independent Test**: Match created → recruiter agent sends message → seeker agent receives it via WebSocket within 5 seconds → conversation thread visible on /messages page.

**Acceptance Scenarios**:

1. **Given** a new Match is created with status "pending", **When** Match is confirmed by seeker, **Then** recruiter agent automatically sends introduction message via WebSocket
2. **Given** recruiter agent sends a message, **When** it is sent, **Then** seeker agent receives it via WebSocket and can reply in real-time
3. **Given** user visits /messages page, **When** page loads, **Then** it shows all conversation threads sorted by most recent activity

---

### User Story 3 - AI-Powered Match Evaluation (Priority: P1)

As the system, when evaluating a seeker-agent to job match, I use AI to analyze skills alignment, experience requirements, and other factors to calculate a compatibility score with reasoning.

**Why this priority**: Without AI evaluation, matches are either random or based on simple keyword matching. AI-powered scoring ensures quality matches and provides transparency to users about why they were matched.

**Independent Test**: Seeker agent profile + Job description → AI evaluates match → Match created with score between 0-1 and written reasoning.

**Acceptance Scenarios**:

1. **Given** seeker agent has skills ["Go", "Kubernetes"] and job requires ["Go", "Docker"], **When** AI evaluates match, **Then** score reflects 2/3 skills matched with reasoning like "Strong Go alignment, missing Docker experience"
2. **Given** AI evaluation takes longer than 3 seconds, **When** timeout occurs, **Then** match is created with default score 0.5 and reasoning "AI evaluation timed out"
3. **Given** AI API is unavailable, **When** match evaluation is requested, **Then** system uses rule-based fallback scoring

---

### User Story 4 - Agent AI Response Generation (Priority: P2)

As a seeker agent, when I receive a message from a recruiter agent, I use AI to generate an intelligent response based on my profile, the job context, and the message content.

**Why this priority**: Agents must be able to autonomously communicate. Manual responses defeat the purpose of an A2A recruitment platform. AI response generation makes agents feel alive and responsive.

**Independent Test**: Recruiter sends "Hi, are you interested in this Go developer role?" → Seeker agent generates AI response → Response appears in conversation within 10 seconds.

**Acceptance Scenarios**:

1. **Given** seeker agent receives message about a job opportunity, **When** message is received, **Then** AI generates contextual response within 10 seconds
2. **Given** recruiter agent receives response, **When** response is generated, **Then** it reflects seeker's profile/skills and the specific job details
3. **Given** AI response generation fails, **When** error occurs, **Then** system shows "Agent is thinking..." placeholder and retries up to 3 times

---

### User Story 5 - WebSocket Connection Management (Priority: P1)

As a user, my WebSocket connection remains stable and automatically reconnects if disconnected, so I don't miss any messages.

**Why this priority**: Real-time features are only valuable if the connection is reliable. Users cannot be expected to manually refresh or reconnect.

**Independent Test**: User is connected to WebSocket → network disconnects → connection auto-recovers within 30 seconds → no messages are lost.

**Acceptance Scenarios**:

1. **Given** user is connected to WebSocket, **When** connection drops, **Then** system automatically attempts reconnection every 5 seconds up to 5 times
2. **Given** reconnection succeeds, **When** user reconnects, **Then** they receive any messages that arrived during disconnection
3. **Given** WebSocket is disconnected for more than 5 minutes, **When** user visits /messages, **Then** page shows "Reconnecting..." indicator and blocks sending

---

### User Story 6 - Interview & Offer Flow (Priority: P2)

As the system, when agents reach mutual interest, I automatically schedule an interview and generate an offer, with all status changes communicated via WebSocket.

**Why this priority**: The recruitment lifecycle doesn't end at conversation. Interview scheduling and offer management complete the full recruitment loop.

**Independent Test**: Agents exchange 3+ messages with positive sentiment → Interview is auto-scheduled → Offer is generated → Both appear on respective pages.

**Acceptance Scenarios**:

1. **Given** Match status is "mutual_interest", **When** 3+ messages exchanged, **Then** system auto-schedules interview for available slot
2. **Given** interview is scheduled, **When** status confirmed by both parties, **Then** system generates offer with salary based on job range
3. **Given** offer is generated, **When** user visits /offers, **Then** offer shows countdown timer and accept/decline actions work

---

### Edge Cases

- What happens when WebSocket connection fails repeatedly? → Show "Connection failed" with manual retry button
- What happens when AI API returns an error? → Use fallback rule-based scoring, log error
- What happens when two agents have conflicting interests? → Set status to "disputed", require human review
- What happens when Match already exists for same seeker-agent and job? → Return existing Match, don't create duplicate
- What happens when recruiter agent sends inappropriate message? → Content filter blocks message, sender sees warning
- What happens when offer expires? → Status changes to "expired", countdown shows "Expired"

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST automatically create Match records when seeker agent views/browses jobs, evaluating profile-job compatibility
- **FR-002**: Match scoring MUST use AI evaluation (pkg/ai/client.go EvaluateMatch) to calculate 0-1 score with reasoning
- **FR-003**: WebSocket server MUST run at /api/messages/ws, accepting connections and relaying messages in real-time
- **FR-004**: Agent messages MUST trigger AI response generation using pkg/ai/client.go GenerateAgentResponse
- **FR-005**: GET /api/offers endpoint MUST return all offers for the authenticated user
- **FR-006**: GET /api/messages endpoint MUST return all conversation threads for the authenticated user
- **FR-007**: WebSocket connection MUST auto-reconnect on disconnect with exponential backoff (5s, 10s, 20s, max 60s)
- **FR-008**: Match status transitions MUST follow: pending → expressed_interest → mutual_interest → negotiating → offer_sent → accepted/declined
- **FR-009**: AI response generation MUST complete within 10 seconds or show "thinking" placeholder
- **FR-010**: Messages MUST be persisted to database before WebSocket relay (no message loss)
- **FR-011**: Content filtering MUST block messages with profanity or PII before sending
- **FR-012**: System MUST log all AI evaluation inputs and outputs for audit trail

### Key Entities

- **Match**: Links seeker-agent to job, contains AI-calculated score (0-1), reasoning text, and status
- **Conversation**: Thread of messages between two agents for a specific Match, has unread count
- **Message**: Individual message with sender, content, timestamp, intent type (introduction/negotiation/offer)
- **Interview**: Scheduled interview with participants, datetime, status (pending/confirmed/cancelled)
- **Offer**: Job offer with compensation details, expiration, status (pending/accepted/declined/expired)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Match records created within 10 seconds of seeker viewing job (auto-match latency)
- **SC-002**: Messages delivered via WebSocket within 5 seconds of sending (message latency)
- **SC-003**: AI response generation completes within 10 seconds (response latency)
- **SC-004**: WebSocket auto-reconnects within 30 seconds on disconnect (reconnection success)
- **SC-005**: Match score accuracy - AI scores correlate with human assessment in 80%+ of cases
- **SC-006**: Zero message loss during WebSocket reconnection (reliability)
- **SC-007**: Full recruitment flow completes from job view → match → conversation → interview → offer within 5 minutes

## Assumptions

- Backend WebSocket library is available (Gorilla WebSocket or similar)
- AI API (LongCat) is functional and responsive (<3s latency)
- Seeker agents have profile/skills defined for matching evaluation
- Recruiter agents have jobs posted with clear requirements
- Database schema supports Match, Message, Interview, Offer with proper indexes
- Existing auth system (JWT) works for WebSocket authentication
