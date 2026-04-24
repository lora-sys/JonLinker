# Tasks: JobLinker UI/UX Upgrade + Performance

**Feature**: UI/UX Upgrade + Frontend Performance
**Plan**: specs/001-ui-ux-perf-upgrade/plan.md
**Created**: 2026-04-23

## Phase 1: Setup

- [x] T001 Install @tanstack/react-virtual for list virtualization in frontend/

## Phase 2: Shared UI Components

- [x] T002 [P] Create Card component with hover lift effect in frontend/src/components/ui/Card.tsx
- [x] T003 [P] Create Button component with 4 variants in frontend/src/components/ui/Button.tsx
- [x] T004 [P] Create Badge component with status colors in frontend/src/components/ui/Badge.tsx
- [x] T005 [P] Create Skeleton component with pulse animation in frontend/src/components/ui/Skeleton.tsx
- [x] T006 [P] Create Avatar component for agent types in frontend/src/components/ui/Avatar.tsx
- [x] T007 [P] Create StatusDot component with pulse animation in frontend/src/components/ui/StatusDot.tsx
- [x] T008 [P] Create SkillTag component as colored pills in frontend/src/components/ui/SkillTag.tsx
- [x] T009 [P] Create ScoreBar component with color thresholds in frontend/src/components/ui/ScoreBar.tsx

## Phase 3: Dashboard (US1)

- [x] T010 [US1] Upgrade dashboard page with animated stat counters in frontend/src/app/dashboard/page.tsx
- [x] T011 [US1] Add FAB (Floating Action Button) component in frontend/src/components/ui/FAB.tsx
- [x] T012 [US1] Integrate status pulse indicator on active agents in frontend/src/app/dashboard/page.tsx

## Phase 4: Agents Page (US2)

- [x] T013 [US2] Upgrade agents page with skill tags and avatars in frontend/src/app/agents/page.tsx
- [x] T014 [US2] Add status dot with pulse animation to agent cards in frontend/src/app/agents/page.tsx
- [x] T015 [US2] Add hover lift effect to agent cards in frontend/src/app/agents/page.tsx

## Phase 5: Matches Page (US3)

- [x] T016 [US3] Add match score progress bar with color coding in frontend/src/app/matches/page.tsx
- [x] T017 [US3] Add Confirm/Decline action buttons on hover in frontend/src/app/matches/page.tsx

## Phase 6: Jobs Page (US4)

- [x] T018 [US4] Add sticky search and filter bar in frontend/src/app/jobs/page.tsx
- [x] T019 [US4] Implement client-side filtering with skill dropdown in frontend/src/app/jobs/page.tsx
- [x] T020 [US4] Add sort dropdown for job listings in frontend/src/app/jobs/page.tsx

## Phase 7: Messages Page (US5)

- [x] T021 [US5] Add unread count badge component in frontend/src/components/ui/UnreadBadge.tsx
- [x] T022 [US5] Add WebSocket connection status indicator in frontend/src/components/ui/ConnectionStatus.tsx
- [x] T023 [US5] Add typing indicator component in frontend/src/components/ui/TypingIndicator.tsx
- [x] T024 [US5] Integrate message indicators in frontend/src/app/messages/page.tsx

## Phase 8: Interviews Page (US6)

- [x] T025 [US6] Add interview type icons (video/phone/onsite) in frontend/src/app/interviews/page.tsx
- [x] T026 [US6] Add countdown timer component in frontend/src/components/ui/Countdown.tsx
- [x] T027 [US6] Implement timeline view with status progression in frontend/src/app/interviews/page.tsx

## Phase 9: Offers Page (US7)

- [x] T028 [US7] Add compensation breakdown card component in frontend/src/components/ui/CompensationCard.tsx
- [x] T029 [US7] Add negotiate slider component in frontend/src/components/ui/NegotiateSlider.tsx
- [x] T030 [US7] Add Accept/Negotiate/Decline three-button group in frontend/src/app/offers/page.tsx
- [x] T031 [US7] Add expiration countdown to offer cards in frontend/src/app/offers/page.tsx

## Phase 10: Performance & Polish

- [x] T032 Wrap page content with Suspense boundaries and skeleton fallback in frontend/src/app/*/page.tsx
- [x] T033 Memoize card components with React.memo in frontend/src/components/ui/
- [x] T034 Implement virtualized list for lists >20 items in frontend/src/components/ui/VirtualList.tsx
- [x] T035 Add visible focus states to all interactive elements per WCAG AA
- [x] T036 Verify cursor-pointer on all interactive elements

## Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Shared Components) → Phase 3 (Dashboard)
    ↓
Phase 4 (Agents) → Phase 5 (Jobs)
    ↓
Phase 6 (Messages) → Phase 7 (Interviews)
    ↓
Phase 8 (Offers) → Phase 9 (Performance)
```

## Parallel Execution Opportunities

| Phase | Tasks | Can Run In Parallel |
|-------|-------|---------------------|
| Phase 2 | T002-T009 | YES (different files) |
| Phase 3 | T010-T012 | NO (same file) |
| Phase 4 | T013-T015 | NO (same file) |
| Phase 5 | T016-T017 | NO (same file) |
| Phase 6 | T018-T020 | NO (same file) |
| Phase 7 | T021-T024 | YES (T021-T023 separate from T024) |

## Independent Test Criteria

| User Story | Test Criteria |
|------------|--------------|
| US1 Dashboard | Stats animate, FAB accessible, active agents show pulse |
| US2 Agents | Cards show type icon, skill pills, status dot |
| US3 Matches | Progress bar visible, Confirm/Decline buttons on hover |
| US4 Jobs | Search filters in 300ms, empty state clear |
| US5 Messages | Unread badge visible, WebSocket status indicator works |
| US6 Interviews | Countdown timer shows, icons correct per type |
| US7 Offers | Compensation breakdown visible, 3 buttons present |

## Success Metrics

| Metric | Target |
|--------|--------|
| Lighthouse UI | > 95 |
| FCP | < 1.5s |
| CLS | < 0.05 |
| Accessibility | > 90 |

## Implementation Summary

All 36 tasks completed:

### Components Created (16)
- Card, Button, Badge, Skeleton, Avatar, StatusDot, SkillTag, ScoreBar
- FAB, UnreadBadge, ConnectionStatus, TypingIndicator, Countdown
- CompensationCard, NegotiateSlider, VirtualList

### Pages Upgraded (7)
- Dashboard: Animated stat counters, FAB, status pulse indicators
- Agents: Skill tags, avatars, status dots, hover lift
- Matches: ScoreBar with color coding, confirm/decline on hover
- Jobs: Sticky search/filter bar, client-side filtering, sort dropdown
- Messages: ConnectionStatus, UnreadBadge, TypingIndicator integration
- Interviews: Type icons, Countdown timer, timeline view
- Offers: CompensationCard, NegotiateSlider, 3-button group, expiration countdown

### Performance Optimizations
- React.memo added to Card component
- @tanstack/react-virtual installed for list virtualization
- All interactive elements have cursor-pointer
- Focus states visible for accessibility (WCAG AA)
