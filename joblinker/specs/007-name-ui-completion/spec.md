# Feature Specification: Frontend UI Completion

**Feature Branch**: `007-name-ui-completion`
**Created**: 2026-04-28
**Status**: Draft
**Input**: Frontend UI completion with Aceternity UI components: Job posting page, WebSocket testing, Agent auto-dialogue, Interview/Offer cards

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Recruiter Posts Job (Priority: P1)

As a recruiter, I need a dedicated page to post new jobs so that job listings are created with proper structured data without using API tools.

**Why this priority**: Job posting is a fundamental recruiter action. Without a UI, recruiters cannot independently add jobs — they must use curl or Postman, which is not user-friendly.

**Independent Test**: Recruiter visits `/jobs/new`, fills the form, submits → job appears in the job board within 5 seconds.

**Acceptance Scenarios**:

1. **Given** user is logged in as recruiter, **When** they visit `/jobs/new`, **Then** the job posting form is displayed with all required fields
2. **Given** recruiter fills all required fields (title, description, location, salary, requirements), **When** they submit, **Then** the job is created and they are redirected to the job board
3. **Given** recruiter submits with missing required fields, **When** validation fails, **Then** inline error messages appear on the missing fields
4. **Given** recruiter enters structured requirements as tags or a list, **When** submitted, **Then** the structured JSON is stored and searchable

---

### User Story 2 - WebSocket Real-Time Communication (Priority: P1)

As a user, I need to verify that WebSocket connections work correctly in the browser so that real-time messaging is functional.

**Why this priority**: WebSocket is the backbone of real-time A2A dialogue. Without verified WS communication, agent auto-chat cannot function.

**Independent Test**: User opens browser WS test page → connects to WS endpoint → sends a test message → receives echo response within 1 second.

**Acceptance Scenarios**:

1. **Given** user visits `/ws-test` page, **When** page loads, **Then** a WebSocket connection is established to `/api/messages/ws` with the auth token
2. **Given** WS connection is open, **When** user sends a test message, **Then** the message appears in the debug panel and the server echoes it back
3. **Given** WS connection drops, **When** disconnection is detected, **Then** the page shows "Disconnected" status and attempts reconnection every 5 seconds
4. **Given** WS is reconnecting, **When** reconnection succeeds, **Then** status changes to "Connected" and pending messages are flushed

---

### User Story 3 - Agent Auto-Dialogue (Priority: P2)

As the system, when a match reaches "expressed_interest" status, the agents automatically exchange messages via WebSocket through RabbitMQ without user intervention.

**Why this priority**: The core A2A value proposition is agents talking to each other autonomously. Auto-dialogue is what makes the platform intelligent rather than a simple job board.

**Independent Test**: Match status changes to `expressed_interest` → within 10 seconds, agent messages appear in the conversation thread → chat continues for at least 3 exchanges.

**Acceptance Scenarios**:

1. **Given** a match has status "expressed_interest", **When** the condition is met, **Then** recruiter agent is triggered via RabbitMQ to send an introduction message
2. **Given** recruiter agent sends a message, **When** it is published to RabbitMQ, **Then** the message appears in the conversation thread via WebSocket within 5 seconds
3. **Given** seeker agent receives a message, **When** received, **Then** AI generates a contextual response and sends it back via RabbitMQ within 10 seconds
4. **Given** user visits `/messages` or `/conversation/:matchId`, **When** page loads, **Then** all exchanged messages are displayed in chronological order

---

### User Story 4 - Interview & Offer Cards (Priority: P2)

As a user, when my match progresses to interview or offer stage, I see visual cards in the conversation thread so I can quickly understand and respond to interview invitations and job offers.

**Why this priority**: Interview and offer are critical decision points. Cards make these high-value actions visually prominent and easy to act on, improving conversion rates.

**Independent Test**: Match status changes to `interview_scheduled` → interview card appears in conversation → user clicks Accept → status updates.

**Acceptance Scenarios**:

1. **Given** match status changes to "interview_scheduled", **When** the conversation loads, **Then** an interview card is displayed with date, time, format, and location
2. **Given** interview card is shown, **When** user clicks "Confirm" or "Decline", **Then** the action is sent to the server and the card updates to reflect the response
3. **Given** match status changes to "offer_sent", **When** the conversation loads, **Then** an offer card is displayed with salary, start date, and response deadline
4. **Given** offer card is shown, **When** user clicks "Accept" or "Decline", **Then** the action is persisted and the match status transitions accordingly
5. **Given** an interview or offer card is pending user response, **When** the deadline passes, **Then** the card shows an "Expired" state with appropriate styling

---

### Edge Cases

- What happens when WebSocket fails to connect? → Show "Connection failed" with manual retry button
- What happens when AI response generation times out? → Show "Agent is thinking..." with animated indicator
- What happens when recruiter submits duplicate job for the same position? → Allow duplicates (no blocking)
- What happens when user tries to post a job but is not a recruiter? → Show 403 forbidden with redirect to dashboard
- What happens when interview/offer response fails to submit? → Show error toast with retry option
- What happens when WS receives a message for an unknown match? → Log error, discard message, do not crash

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `/jobs/new` page MUST be accessible only to recruiter role; seeker users see 403
- **FR-002**: Job posting form MUST include: title, description, location, employment type, salary range (min/max), structured requirements (as tag input)
- **FR-003**: Job posting form submission MUST call `POST /api/jobs` with properly structured JSON
- **FR-004**: After successful job creation, user MUST be redirected to `/jobs` with a success toast
- **FR-005**: `/ws-test` page MUST establish WebSocket connection to `/api/messages/ws?token=<jwt>`
- **FR-006**: WS test page MUST display connection status (Connecting/Connected/Disconnected/Reconnecting)
- **FR-007**: WS test page MUST have a message input and debug log showing sent/received messages
- **FR-008**: WS test page MUST auto-reconnect with exponential backoff (5s, 10s, 20s, max 60s) on disconnect
- **FR-009**: Agent auto-dialogue MUST be triggered when match status transitions to `expressed_interest`
- **FR-010**: Agent messages MUST be queued via RabbitMQ (`job_seeker_queue`, `recruiter_queue`) and relayed via WebSocket
- **FR-011**: Conversation page MUST display messages in chronological order with sender identification
- **FR-012**: Interview cards MUST display: scheduled datetime, format (video/onsite/phone), location/link, participant names
- **FR-013**: Offer cards MUST display: salary amount, start date, response deadline, and accept/decline buttons
- **FR-014**: All async actions (form submit, card responses) MUST show loading states and error handling
- **FR-015**: Use Aceternity UI components for form (signup-form), cards (expandable-cards/focus-cards), background (background-beams)

### Key Entities

- **Job**: Title, description, location, type (full-time/part-time/contract), salary range (min/max), structured requirements (JSON), recruiter agent reference
- **Conversation**: Match-linked thread, participants, last message preview, unread count
- **Interview**: Match reference, scheduled datetime, format, location, status (pending/confirmed/cancelled), reminder flag
- **Offer**: Match reference, salary amount, start date, expiration datetime, status (pending/accepted/declined/expired)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Recruiter can post a job via UI and see it appear in job board within 5 seconds
- **SC-002**: WebSocket connection establishes within 3 seconds of page load
- **SC-003**: Test messages sent via WS appear in debug log within 1 second (round-trip)
- **SC-004**: Agent auto-dialogue produces at least 3 message exchanges after match reaches `expressed_interest`
- **SC-005**: Interview card appears in conversation within 10 seconds of status change
- **SC-006**: Offer card appears in conversation within 10 seconds of status change
- **SC-007**: All form submissions show loading state during request and success/error feedback after
- **SC-008**: WS reconnection succeeds within 30 seconds on disconnect

## Assumptions

- Backend API endpoints already exist: `POST /api/jobs`, WebSocket endpoints at `/api/messages/ws` and `/api/messages/:matchId/ws`
- RabbitMQ queues (`job_seeker_queue`, `recruiter_queue`) are already configured and consumers are registered
- Agent auto-dialogue handlers exist in backend (MessageHandler with AI GenerateAgentResponse)
- Match status transitions are implemented: pending → expressed_interest → mutual_interest → negotiating → offer_sent → accepted/declined
- Auth system (JWT) works for WebSocket authentication via query param
- Aceternity UI components are available via shadcn integration: signup-form, expandable-cards, focus-cards, background-beams, typewriter-effect
