# Feature Specification: Frontend-Backend Integration & Navigation

**Feature Branch**: `005-frontend-backend-integration`
**Created**: 2026-04-25
**Status**: Draft
**Input**: User report: "大部分功能没实现，dashboard没有链接到别的页面，mock数据太多"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Unified Dashboard Navigation Hub (Priority: P1)

As a user, I want Dashboard to be the central entry point that links to all major features so that I can navigate the entire platform from one place.

**Why this priority**: Dashboard is the landing page after login but currently missing links to Jobs, Offers, Interviews. Users cannot discover core features.

**Independent Test**: User lands on /dashboard, sees cards/links for Jobs, Offers, Interviews, Messages, Agents, Matches — clicking each navigates to correct page.

**Acceptance Scenarios**:

1. **Given** user is logged in, **When** they visit /dashboard, **Then** they see cards linking to Jobs, Offers, Interviews, Messages, Agents, Matches pages
2. **Given** user clicks "Jobs" on dashboard, **When** navigation completes, **Then** /jobs page loads with real data from API
3. **Given** user clicks "Offers" on dashboard, **When** navigation completes, **Then** /offers page loads with real offer data (not mock)

---

### User Story 2 - Jobs Page with Real API Data (Priority: P1)

As a job seeker, I want to browse real job listings from the backend so that I see actual opportunities, not hardcoded fake data.

**Why this priority**: Jobs page currently shows mock data. Real job listings must come from backend via /api/jobs endpoint.

**Independent Test**: POST a job via API → verify it appears on /jobs page.

**Acceptance Scenarios**:

1. **Given** jobs exist in backend, **When** user visits /jobs, **Then** page fetches GET /api/jobs and displays real job cards
2. **Given** user is HR role, **When** they visit /jobs, **Then** they see "Post Job" button that links to job creation flow
3. **Given** job listings are loading, **When** page renders, **Then** loading skeleton is shown until data arrives

---

### User Story 3 - Offers Page with Real API Data (Priority: P1)

As a candidate, I want to see my real offer offers from the backend so that I can take action on actual opportunities.

**Why this priority**: Offers page currently shows MOCK_OFFERS array. Real offers must come from backend.

**Independent Test**: Backend creates offer → verify it appears on /offers with countdown timer showing real expiration.

**Acceptance Scenarios**:

1. **Given** offers exist in backend for current user, **When** user visits /offers, **Then** GET /api/offers returns real offers
2. **Given** offer has expiration date, **When** page renders, **Then** Countdown component shows time remaining using real dates
3. **Given** user accepts an offer, **When** they click Accept, **Then** PATCH /api/offers/:id updates status to accepted

---

### User Story 4 - Interviews Page with Real API Data (Priority: P1)

As a user, I want to see my real interview invitations from the backend.

**Why this priority**: Interviews page likely shows mock data. Must connect to GET /api/interviews.

**Independent Test**: Backend creates interview → verify it appears on /interviews page.

**Acceptance Scenarios**:

1. **Given** interviews exist for current user, **When** user visits /interviews, **Then** page fetches GET /api/interviews
2. **Given** interview is in pending state, **When** user views it, **Then** they can Accept or Decline (PATCH /api/interviews/:id)
3. **Given** interview has notes, **When** user views details, **Then** notes are displayed correctly

---

### User Story 5 - Messages & Conversation Real-Time (Priority: P1)

As a user, I want to see real A2A conversation messages and receive new messages via WebSocket without refreshing.

**Why this priority**: Messages page likely shows mock data. Real-time WebSocket updates must be connected.

**Independent Test**: Agent sends message → user receives it in <5 seconds via WebSocket on /messages page.

**Acceptance Scenarios**:

1. **Given** user visits /messages, **When** page loads, **Then** GET /api/messages returns real conversation history
2. **Given** new message arrives, **When** WebSocket push received, **Then** message appears in UI without page refresh
3. **Given** user clicks conversation, **When** they visit /conversation/:matchId, **Then** full dialogue history loads

---

### User Story 6 - Auth Flow End-to-End (Priority: P1)

As a user, I want to register, login, and logout with session persistence so that unauthorized users cannot access protected pages.

**Why this priority**: Auth must work end-to-end. Unauthenticated requests to /jobs, /offers, /dashboard should redirect to login.

**Independent Test**: Register → logout → login → access /dashboard. All steps succeed.

**Acceptance Scenarios**:

1. **Given** user submits registration form, **When** data is valid, **Then** POST /api/auth/register creates account and returns session
2. **Given** user submits login form, **When** credentials correct, **Then** POST /api/auth/login returns session, cookie is set, redirect to /dashboard
3. **Given** user clicks logout, **When** logout completes, **Then** session cleared, redirect to /login
4. **Given** unauthenticated user tries to access /dashboard, **When** request made, **Then** redirect to /login

---

### User Story 7 - Agent Management Connected to Backend (Priority: P2)

As a user, I want to create and manage my Agents with real state from the backend.

**Why this priority**: /agents page must show real agent status, not mock state from frontend stores.

**Independent Test**: Create agent via UI → verify it persists in backend → appears in agent list.

**Acceptance Scenarios**:

1. **Given** user visits /agents, **When** page loads, **Then** GET /api/agents returns user's real agents
2. **Given** user creates new agent, **When** form submitted, **Then** POST /api/agents creates agent, agent appears in list with "idle" status
3. **Given** user pauses agent, **When** pause action confirmed, **Then** PATCH /api/agents/:id sets status to paused

---

### User Story 8 - Matches Page with Real Match Data (Priority: P2)

As a user, I want to see real match records with scores and reasons from the backend.

**Why this priority**: Matches page must show real vector similarity scores, not mock percentages.

**Independent Test**: Backend generates match → verify /matches page shows it with correct score.

**Acceptance Scenarios**:

1. **Given** matches exist for user, **When** user visits /matches, **Then** GET /api/matches returns real match records
2. **Given** match has compatibility score, **When** card renders, **Then** score percentage matches backend calculation
3. **Given** match shows reason/justification, **When** user reviews, **Then** matching rationale is displayed

---

### User Story 9 - Settings Page Loads User Profile (Priority: P2)

As a user, I want to view and edit my profile with data from the backend.

**Why this priority**: Settings page should load real user data, not empty form.

**Independent Test**: Update profile → verify change persists via GET /api/auth/me.

**Acceptance Scenarios**:

1. **Given** user visits /settings, **When** page loads, **Then** GET /api/auth/me returns user profile data
2. **Given** user modifies profile fields, **When** they save, **Then** PATCH /api/auth/me updates backend

---

### User Story 10 - Loading States & Error Handling (Priority: P1)

As a user, I want to see loading indicators and error messages so that I know what's happening during data fetches.

**Why this priority**: Currently pages may render blank or show abrupt mock data. Need proper UX.

**Acceptance Scenarios**:

1. **Given** data is being fetched, **When** page renders, **Then** skeleton/spinner is shown
2. **Given** API request fails, **When** error occurs, **Then** error message displayed with retry option
3. **Given** network is offline, **When** user tries to load page, **Then** "No connection" message shown

---

### Edge Cases

- What happens when API returns 401 Unauthorized? → Redirect to /login
- What happens when API returns empty array (no jobs/offers/matches)? → Show empty state with CTA
- What happens when WebSocket connection drops? → Show "Reconnecting..." indicator, auto-retry
- What happens when Countdown timer reaches zero? → Show "Expired" state
- What happens when user role is HR vs Seeker? → Different dashboard cards and navigation items shown
- What happens when user tries to access page they don't have permission for (e.g., HR accessing /offers of another user)? → 403 or redirect

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Dashboard MUST display navigation cards linking to Jobs, Offers, Interviews, Messages, Agents, Matches pages
- **FR-002**: Dashboard MUST show role-appropriate content (Seeker sees applied jobs; HR sees posted jobs)
- **FR-003**: Jobs page MUST fetch real data via GET /api/jobs and render job cards
- **FR-004**: Offers page MUST fetch real offers via GET /api/offers (user's own offers only)
- **FR-005**: Interviews page MUST fetch via GET /api/interviews and display pending/confirmed/cancelled interviews
- **FR-006**: Messages page MUST fetch via GET /api/messages and display conversation threads
- **FR-007**: Conversation page (/conversation/:matchId) MUST fetch messages for specific match
- **FR-008**: WebSocket connection MUST be established on /messages page for real-time updates
- **FR-009**: Agents page MUST fetch via GET /api/agents and display agent cards with status
- **FR-010**: Matches page MUST fetch via GET /api/matches and display match cards with scores
- **FR-011**: Settings page MUST fetch user profile via GET /api/auth/me
- **FR-012**: All forms (login, register, profile update) MUST submit to correct API endpoints
- **FR-013**: Auth-protected pages MUST redirect to /login when unauthenticated
- **FR-014**: Countdown component MUST accept real Date objects and show correct time remaining
- **FR-015**: All pages MUST show loading skeleton while fetching initial data
- **FR-016**: All pages MUST show error state with retry button on API failure
- **FR-017**: Empty states MUST be shown when API returns empty arrays
- **FR-018**: Sidebar navigation MUST show links to all accessible pages based on user role

### Key Entities

- **User**: Authenticated platform user with role (seeker/hr/admin) and profile
- **Job**: Posted position with title, company, requirements, salary range, status
- **Offer**: Job offer with compensation details, expiration, status (pending/accepted/declined)
- **Interview**: Scheduled interview with participants, datetime, status (pending/confirmed/cancelled)
- **Message**: A2A dialogue message with sender, content, timestamp
- **Agent**: User's AI agent with type (seeker/recruiter), status (active/paused/idle)
- **Match**: Candidate-job match with compatibility score and status

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All dashboard navigation links work and lead to correct pages
- **SC-002**: Jobs page shows real jobs from backend within 2 seconds of load
- **SC-003**: Offers page shows real offers with countdown timers showing accurate time remaining
- **SC-004**: Unauthenticated users cannot access /dashboard, /jobs, /offers, /interviews, /messages — redirected to /login
- **SC-005**: Messages page receives new messages via WebSocket within 5 seconds of sending
- **SC-006**: All pages show loading skeleton during data fetch
- **SC-007**: All pages show error state with retry on API failure (not blank screen)
- **SC-008**: Empty states shown when no data exists (no jobs, no offers, etc.)
- **SC-009**: Countdown on offers page shows "Expired" when offer expires

## Assumptions

- Backend API is already implemented with correct endpoints (GET /api/jobs, GET /api/offers, etc.)
- Backend WebSocket endpoint exists at /api/ws or similar for real-time messages
- Auth uses cookie-based session with httpOnly cookies
- User role is returned in auth responses to determine UI visibility
- API base URL is /api (Next.js proxy to backend)
- Recharts or similar used for dashboard metrics (not implemented yet, out of scope)
