# Tasks: JobLinker UI/UX Consistency Fix

**Feature**: 004-homepage-ui-fix | **Total Tasks**: 16
**Generated**: 2026-04-24

## Overview

Fix UI/UX consistency issues across ALL 12 JobLinker pages. Issues include: gray→slate color palette, max-width inconsistencies, missing gradient backgrounds, button styling, and logo sizing.

## Phase 1: Setup (No user story - foundational)

- [x] T001 Fix settings/page.tsx - gray→slate palette, max-w-2xl→max-w-7xl, add gradient wrapper
- [x] T002 Fix admin/page.tsx - gray→slate palette throughout

## Phase 2: Interior Pages (Gradient + Max-width fixes)

- [x] T003 [P] Fix jobs/page.tsx - add gradient wrapper, max-w-4xl→max-w-7xl
- [x] T004 [P] Fix agents/page.tsx - add gradient wrapper, max-w-4xl→max-w-7xl
- [x] T005 [P] Fix matches/page.tsx - add gradient wrapper, max-w-4xl→max-w-7xl
- [x] T006 [P] Fix offers/page.tsx - add gradient wrapper, max-w-4xl→max-w-7xl
- [x] T007 [P] Fix interviews/page.tsx - add gradient wrapper, max-w-4xl→max-w-7xl
- [x] T008 [P] Fix messages/page.tsx - add gradient wrapper, max-w-4xl→max-w-7xl

## Phase 3: Homepage Refinement

- [x] T009 Fix homepage layout - max-w-6xl→max-w-7xl, section padding consistency
- [x] T010 Fix homepage typography - H1 text-7xl→text-4xl lg:text-5xl
- [x] T011 Fix homepage footer logo - w-6 h-6 text-xs→w-10 h-10 text-lg to match header

## Phase 4: Cross-Cutting Polish

- [x] T012 Verify all pages use slate palette - no text-gray-* or bg-gray-* remain
- [x] T013 Verify all buttons have consistent hover/focus states
- [x] T014 Browser test - 1440px viewport, all pages centered, no horizontal scroll
- [x] T015 Lighthouse accessibility audit on homepage
- [x] T016 Update SPEC.md status to "Ready for Testing"

---

## Task Details

### T001 - Fix settings/page.tsx

**File**: `frontend/src/app/settings/page.tsx`
**Story**: N/A (foundational)

**Changes**:
- Replace `max-w-2xl mx-auto` container with `min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white`
- Add inner container: `max-w-7xl mx-auto px-4 py-8`
- Replace `text-gray-900` → `text-slate-900`
- Replace `text-gray-500` → `text-slate-500`
- Replace `text-gray-600` → `text-slate-600`
- Replace `bg-gray-200` → `bg-slate-200`
- Add glass morphism to card: `bg-white/80 backdrop-blur-xl border border-white/20`

**Verification**: Page loads at 1440px, content centered, slate colors used

---

### T002 - Fix admin/page.tsx

**File**: `frontend/src/app/admin/page.tsx`
**Story**: N/A (foundational)

**Changes**:
- Replace `bg-gray-50` → `bg-slate-50`
- Replace `text-gray-900` → `text-slate-900`
- Replace `text-gray-700` → `text-slate-700`
- Replace `border-gray-200` → `border-slate-200`
- Update MetricsPanel, JobManagement, AgentManagement to use slate palette

**Verification**: Admin dashboard shows slate colors throughout

---

### T003 - Fix jobs/page.tsx

**File**: `frontend/src/app/jobs/page.tsx`
**Story**: User Story 1

**Changes**:
- Wrap content in `min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white`
- Change `max-w-4xl` → `max-w-7xl`
- Add glass morphism to empty state card

**Verification**: jobs page centered, consistent with other interior pages

---

### T004-T008 - Fix remaining interior pages

**Files**: agents, matches, offers, interviews, messages - each page.tsx
**Story**: User Story 1

**Pattern** (apply to all):
```tsx
<div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
  <div className="max-w-7xl mx-auto px-4 py-8">
    {/* existing content */ }
  </div>
</div>
```

**Changes per page**:
- jobs/page.tsx: Add wrapper, max-w-7xl
- agents/page.tsx: Add wrapper, max-w-7xl
- matches/page.tsx: Add wrapper, max-w-7xl, ensure hover class on Card
- offers/page.tsx: Add wrapper, max-w-7xl
- interviews/page.tsx: Add wrapper, max-w-7xl
- messages/page.tsx: Add wrapper, max-w-7xl

---

### T009-T011 - Homepage refinements

**File**: `frontend/src/app/page.tsx`
**Story**: User Story 1, 2, 5

**T009 Changes**:
- Hero section: `max-w-4xl mx-auto` → `max-w-7xl mx-auto`
- Section containers: `max-w-6xl` → `max-w-7xl`
- Section padding: `py-20 py-24` → consistent `py-16`

**T010 Changes**:
- H1: `text-5xl md:text-7xl` → `text-4xl lg:text-5xl`

**T011 Changes** (footer logo):
- Change footer logo container from `w-6 h-6` → `w-10 h-10`
- Change footer logo text from `text-xs` → `text-lg`
- Apply same gradient and rounded styling as header logo

---

### T012-T015 - Verification tasks

**T012**: Run grep across frontend/src/app for `text-gray-`, `bg-gray-`, `border-gray-` - should return zero matches (except admin purely-functional areas)

**T013**: Inspect button components in DashboardClient.tsx and page.tsx files - verify consistent `transition-all duration-200` and `focus:ring-*`

**T014**: Open each page at 1440px in browser, verify no horizontal scrollbar

**T015**: Run Lighthouse accessibility audit on homepage, target score ≥ 90

---

## Dependency Graph

```
T001 (settings) ─┐
T002 (admin) ─────┴── T012 (palette verification)
                         │
T003-T008 (interior pages, parallel) ──┬── T014 (browser test)
                                      │
T009-T011 (homepage) ─────────────────┴── T015 (Lighthouse)
                                          │
                                      T016 (complete)
```

## Independent Test Criteria

| User Story | Test |
|------------|------|
| US1 (Layout) | Load homepage at 1440px, verify content centered, no horizontal scroll |
| US2 (Typography) | Visual inspection - H1/H2/H3 clearly distinguishable |
| US3 (Buttons) | Tab through buttons, visible focus ring, smooth hover transition |
| US4 (Cards) | Compare How It Works and Privacy cards - identical styling |
| US5 (Brand) | Header and footer logos identical size |

## Suggested MVP Scope

**Phase 1 only (T001-T002)**: Fix settings and admin gray→slate palette issues. This is the highest impact change since these pages are most-visited after homepage.

## Implementation Strategy

1. **First**: Fix gray→slate (T001, T002) - immediate visual consistency
2. **Second**: Add gradient wrappers to interior pages (T003-T008) - unified look
3. **Third**: Homepage polish (T009-T011) - final refinement
4. **Fourth**: Verification (T012-T016) - ensure quality