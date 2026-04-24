# Implementation Plan: JobLinker UI/UX Upgrade + Performance

**Branch**: `001-ui-ux-perf-upgrade`
**Date**: 2026-04-23
**Spec**: specs/001-ui-ux-perf-upgrade/spec.md
**Type**: Frontend-only enhancement

## Summary

Upgrade JobLinker frontend UI/UX with polished flat design, animated interactions, and performance optimizations targeting Lighthouse UI >95, FCP <1.5s, CLS <0.05.

## Technical Context

| Aspect | Value |
|--------|-------|
| **Language** | TypeScript (Next.js 16) |
| **Framework** | Next.js 16 + React 19 |
| **Styling** | Tailwind CSS v4 |
| **State** | Zustand v5 |
| **Icons** | Heroicons (inline SVG) |
| **Virtualization** | @tanstack/react-virtual |
| **Performance** | FCP <1.5s, CLS <0.05, FID <100ms |
| **Accessibility** | WCAG AA, Lighthouse >90 |

### Design System
- **Primary**: #0369A1 (blue-600)
- **Background**: #F0F9FF (sky-50)
- **Surface**: #FFFFFF
- **Border**: #E2E8F0 (slate-200)
- **Text**: #0C4A6E (blue-800)
- **Font**: Plus Jakarta Sans

## Phase 1: Shared UI Components (P1)

### New Files: `frontend/src/components/ui/`

| Component | File | Features |
|-----------|------|----------|
| Card | Card.tsx | Hover lift (translate-y-0.5, shadow-md), active scale, cursor-pointer |
| Button | Button.tsx | Variants: primary, secondary, ghost, destructive |
| Badge | Badge.tsx | Status: active(green), paused(gray), pending(yellow), negotiating(blue), hired |
| Skeleton | Skeleton.tsx | Pulse animation for loading |
| Avatar | Avatar.tsx | Agent type icon (bot/person) |
| StatusDot | StatusDot.tsx | Animated pulse for active status |
| SkillTag | SkillTag.tsx | Colored skill pills |
| ScoreBar | ScoreBar.tsx | Progress bar with color thresholds (>80% green, 50-80% yellow, <50% red) |

## Phase 2: Dashboard (P1)

**File**: `frontend/src/app/dashboard/page.tsx`

- Animated stat counters (CSS animation, 500ms ease-out)
- FAB (Floating Action Button) for quick actions
- Status pulse indicator for active agents

## Phase 3: Agents Page (P1)

**File**: `frontend/src/app/agents/page.tsx`

- Skill tags as colored pills
- Agent type avatars
- Status dot with pulse animation
- Hover lift effect on cards

## Phase 4: Matches Page (P1)

**File**: `frontend/src/app/matches/page.tsx`

- Match score progress bar with color coding
- Confirm/Decline action buttons on hover

## Phase 5: Jobs Page (P2)

**File**: `frontend/src/app/jobs/page.tsx`

- Sticky search + filter bar
- Client-side filtering (no API call)
- Sort dropdown

## Phase 6: Messages Page (P2)

**File**: `frontend/src/app/messages/page.tsx`

- Unread count badge
- WebSocket connection status indicator
- Typing indicator

## Phase 7: Interviews Page (P2)

**File**: `frontend/src/app/interviews/page.tsx`

- Interview type icons (video/phone/onsite)
- Countdown timer for upcoming
- Timeline view with status progression

## Phase 8: Offers Page (P2)

**File**: `frontend/src/app/offers/page.tsx`

- Compensation breakdown card
- Accept/Negotiate/Decline three-button group
- Expiration countdown

## Phase 9: Performance (P1)

| Optimization | Implementation |
|--------------|----------------|
| Suspense | Wrap page content with skeleton fallback |
| React.memo | Memoize card components |
| Virtualization | @tanstack/react-virtual for lists >20 items |
| Dynamic imports | `dynamic()` for heavy components |

## Implementation Order

```
Phase 1 (Shared Components) → Phase 2 (Dashboard) → Phase 3 (Agents)
    → Phase 4 (Matches) → Phase 5 (Jobs) → Phase 6 (Messages)
    → Phase 7 (Interviews) → Phase 8 (Offers) → Phase 9 (Performance)
```

## Files to Create

```
frontend/src/components/ui/
├── Card.tsx
├── Button.tsx
├── Badge.tsx
├── Skeleton.tsx
├── Avatar.tsx
├── StatusDot.tsx
├── SkillTag.tsx
├── ScoreBar.tsx
├── FAB.tsx
├── UnreadBadge.tsx
├── ConnectionStatus.tsx
├── TypingIndicator.tsx
├── Countdown.tsx
├── CompensationCard.tsx
├── NegotiateSlider.tsx
└── VirtualList.tsx
```

## Files to Modify

```
frontend/src/app/
├── dashboard/page.tsx
├── agents/page.tsx
├── matches/page.tsx
├── jobs/page.tsx
├── messages/page.tsx
├── interviews/page.tsx
└── offers/page.tsx
```

## Dependencies

```json
{
  "@tanstack/react-virtual": "^3.x"
}
```

## Success Metrics

| Metric | Target |
|--------|--------|
| Lighthouse UI | > 95 |
| FCP | < 1.5s |
| CLS | < 0.05 |
| FID | < 100ms |
| Accessibility | > 90 |
| cursor-pointer | 100% of interactive elements |
| Focus states | Visible on all inputs |
