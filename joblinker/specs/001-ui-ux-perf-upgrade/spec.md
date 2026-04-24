# Feature Specification: JobLinker UI/UX Upgrade + Performance

**Feature Branch**: `001-ui-ux-perf-upgrade`
**Created**: 2026-04-23
**Status**: Draft
**Input**: Upgrade UI/UX and frontend performance based on plan.md

## User Scenarios & Testing

### User Story 1 - Dashboard Experience (Priority: P1)

As a user, I want my dashboard to load quickly with animated statistics so I can immediately see my recruitment activity at a glance.

**Why this priority**: First impression matters - dashboard is the landing page after login

**Independent Test**: Load dashboard page and verify stats animate, timeline loads, FAB is accessible

**Acceptance Scenarios**:

1. **Given** user is logged in, **When** dashboard loads, **Then** stat counters animate from 0 to actual values within 500ms
2. **Given** user has active agents, **When** viewing dashboard, **Then** active agents show green pulse indicator
3. **Given** user hovers over stat card, **When** mouse enters card, **Then** card lifts with shadow effect (200ms transition)
4. **Given** user clicks FAB, **When** FAB is clicked, **Then** quick action menu slides up

---

### User Story 2 - Agent Management UI (Priority: P1)

As a recruiter, I want to see all my agents with clear status indicators and skill tags so I can quickly assess their activity.

**Why this priority**: Core daily workflow - users manage agents frequently

**Independent Test**: Navigate to /agents, verify cards show type icon, skill pills, status dot, timestamp

**Acceptance Scenarios**:

1. **Given** user has created agents, **When** viewing agents page, **Then** each card shows agent type icon (bot/person avatar)
2. **Given** user has agents with skills, **When** viewing agent card, **Then** skills display as colored pills
3. **Given** agent is active, **When** card is displayed, **Then** status dot pulses green
4. **Given** agent is paused, **When** card is displayed, **Then** status dot shows static gray
5. **Given** user hovers agent card, **When** mouse enters, **Then** card lifts 2px with shadow

---

### User Story 3 - Match Visualization (Priority: P1)

As a recruiter, I want to see match scores as visual progress bars with action buttons so I can quickly decide to confirm or decline matches.

**Why this priority**: Match decision is frequent action - needs clear visual hierarchy

**Independent Test**: Navigate to /matches with existing matches, verify progress bars and action buttons visible

**Acceptance Scenarios**:

1. **Given** match has score, **When** viewing match card, **Then** score displays as progress bar (0-100%)
2. **Given** match score > 80%, **When** bar renders, **Then** bar color is green
3. **Given** match score 50-80%, **When** bar renders, **Then** bar color is yellow
4. **Given** match score < 50%, **When** bar renders, **Then** bar color is red
5. **Given** user hovers match card, **When** mouse enters, **Then** card shows "Confirm" (green) and "Decline" (gray) buttons

---

### User Story 4 - Job Search & Filter (Priority: P2)

As a recruiter, I want to search and filter jobs by skills and location so I can find relevant positions quickly.

**Why this priority**: Improves daily efficiency when managing multiple job postings

**Independent Test**: Navigate to /jobs, use search box and filters, verify results update in real-time

**Acceptance Scenarios**:

1. **Given** user types in search box, **When** text changes, **Then** job list filters within 300ms
2. **Given** user selects skill filter, **When** filter applied, **Then** only jobs with that skill appear
3. **Given** user clicks sort dropdown, **When** option selected, **Then** jobs reorder accordingly
4. **Given** no jobs match filter, **When** filter applied, **Then** empty state shows with clear message

---

### User Story 5 - Real-time Message Indicators (Priority: P2)

As a recruiter, I want to see typing indicators and unread badges so I know when candidates are actively communicating.

**Why this priority**: Real-time communication is core value proposition

**Independent Test**: Open messages page with WebSocket connected, send message, verify indicators

**Acceptance Scenarios**:

1. **Given** unread messages exist, **When** viewing message list, **Then** unread count badge shows on conversation
2. **Given** other party is typing, **When** WebSocket receives typing event, **Then** "typing..." indicator appears
3. **Given** WebSocket disconnects, **When** connection lost, **Then** yellow dot indicator shows "reconnecting"
4. **Given** message delivered, **When** recipient reads it, **Then** single checkmark changes to double checkmark

---

### User Story 6 - Interview Timeline (Priority: P2)

As a recruiter, I want to see interviews in a calendar timeline view with countdown timers so I never miss an interview.

**Why this priority**: Interview scheduling is time-sensitive - visual timeline aids planning

**Independent Test**: Navigate to /interviews with scheduled interviews, verify timeline and countdowns

**Acceptance Scenarios**:

1. **Given** interview is scheduled, **When** viewing list, **Then** interview shows with countdown timer
2. **Given** interview is video format, **When** card displays, **Then** video camera icon shows
3. **Given** interview is phone format, **When** card displays, **Then** phone icon shows
4. **Given** interview passes scheduled time, **When** status updates, **Then** status changes to "Completed"
5. **Given** user cancels interview, **When** cancel confirmed, **Then** status shows "Cancelled" with strikethrough

---

### User Story 7 - Offer Negotiation Flow (Priority: P2)

As a recruiter, I want to see offer details with accept/decline/negotiate buttons so I can respond to offers efficiently.

**Why this priority**: Offer response is critical path - delays cost candidates

**Independent Test**: Navigate to /offers with pending offers, verify action buttons and compensation display

**Acceptance Scenarios**:

1. **Given** offer has compensation details, **When** viewing offer, **Then** breakdown shows base salary, bonus, equity
2. **Given** offer is pending response, **When** viewing card, **Then** three buttons show: Accept (green), Negotiate (blue), Decline (gray)
3. **Given** offer has expiration, **When** viewing card, **Then** countdown shows time remaining
4. **Given** user clicks Negotiate, **When** slider moves, **Then** counter-offer amount updates live

---

### User Story 8 - Performance Targets (Priority: P1)

As a user, I want pages to load fast without layout shifts so my experience feels smooth and professional.

**Why this priority**: Performance directly impacts user perception and task completion rate

**Independent Test**: Run Lighthouse audit, verify Core Web Vitals meet targets

**Acceptance Scenarios**:

1. **Given** user navigates to any page, **When** page loads, **Then** First Contentful Paint under 1.5 seconds
2. **Given** user interacts with page, **When** clicking buttons, **Then** response under 100ms (FID)
3. **Given** page contains async content, **When** content loads, **Then** no layout shift occurs (CLS < 0.05)
4. **Given** user scrolls through list, **When** list is long (>20 items), **Then** list virtualizes and scrolls smoothly at 60fps

---

### User Story 9 - Accessibility Compliance (Priority: P1)

As a user with accessibility needs, I want all interactive elements to be keyboard navigable so I can use the app without a mouse.

**Why this priority**: WCAG compliance - ensures app is usable by everyone

**Independent Test**: Navigate using only Tab key, verify all focus states visible

**Acceptance Scenarios**:

1. **Given** user presses Tab, **When** focus moves, **Then** focus ring visible on active element
2. **Given** user tabs to input field, **When** field receives focus, **Then** label is clearly associated
3. **Given** user tabs to button, **When** button receives focus, **Then** color contrast meets 4.5:1 ratio
4. **Given** page has images, **When** image is meaningful, **Then** alt text is present

---

## Requirements

### Functional Requirements

- **FR-001**: All card components MUST implement hover lift effect with shadow and translate
- **FR-002**: All interactive card elements MUST have cursor-pointer style
- **FR-003**: All hover transitions MUST complete within 200ms
- **FR-004**: Status badges MUST use color-coded styles per status type (active/paused/pending/negotiating/hired)
- **FR-005**: Match cards MUST display score as colored progress bar (green >80%, yellow 50-80%, red <50%)
- **FR-006**: Match cards MUST show Confirm/Decline action buttons on hover
- **FR-007**: Job listings MUST have sticky search and filter bar
- **FR-008**: Message list MUST show unread count badge on conversations with unread messages
- **FR-009**: Messages page MUST show WebSocket connection status indicator
- **FR-010**: Interview cards MUST show countdown timer for upcoming interviews
- **FR-011**: Offer cards MUST show compensation breakdown (base, bonus, equity)
- **FR-012**: Offer cards MUST show Accept/Negotiate/Decline three-button group
- **FR-013**: Dashboard MUST show animated stat counters on page load
- **FR-014**: Agent cards MUST show skill tags as colored pills
- **FR-015**: Agent cards MUST show status indicator dot with pulse animation for active agents

### Non-Functional Requirements

- **NFR-001**: First Contentful Paint MUST be under 1.5 seconds
- **NFR-002**: First Input Delay MUST be under 100ms
- **NFR-003**: Cumulative Layout Shift MUST be under 0.05
- **NFR-004**: All color contrasts MUST meet WCAG AA 4.5:1 ratio
- **NFR-005**: All interactive elements MUST show visible focus state
- **NFR-006**: Lists over 20 items MUST be virtualized for performance
- **NFR-007**: Lighthouse Accessibility score MUST be above 90

### Design System Requirements

- **DS-001**: Primary color MUST be #0369A1
- **DS-002**: Background color MUST be #F0F9FF
- **DS-003**: Font family MUST be Plus Jakarta Sans
- **DS-004**: Cards MUST use border-radius: rounded-xl (12px)
- **DS-005**: Card padding MUST be p-5 (20px)
- **DS-006**: Button transitions MUST use duration-200 (200ms)

---

## Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| Dashboard load with animations | < 1.5s FCP | Lighthouse |
| Time to Interactive | < 3s | Lighthouse |
| Layout Shift (CLS) | < 0.05 | Lighthouse |
| Interactive Elements cursor | 100% | Manual inspection |
| Color Contrast WCAG AA | 4.5:1 | axe-core audit |
| Focus States Visible | 100% | Keyboard navigation test |
| Match Score Progress Bar | Visible on all match cards | Visual inspection |
| Skill Tags on Agents | Visible on all agent cards | Visual inspection |
| WebSocket Status Indicator | Visible when connected/disconnected | Visual inspection |
| Lighthouse Accessibility | > 90 | Lighthouse audit |

---

## Assumptions

1. **Animation library**: Using CSS animations via Tailwind (no external animation library needed)
2. **Icon set**: Using Heroicons (SVG inline)
3. **State management**: Existing Zustand stores remain unchanged
4. **Backend API**: No changes required to API contracts
5. **Browser support**: Modern browsers (Chrome, Firefox, Safari, Edge - last 2 versions)
6. **Mobile support**: iOS Safari and Chrome Android with responsive breakpoints defined

---

## Out of Scope

- Dark mode implementation
- RTL (right-to-left) language support
- PDF/report generation
- Email notification UI
- Mobile native app
