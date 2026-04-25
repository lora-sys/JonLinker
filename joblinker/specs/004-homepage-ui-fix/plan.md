# Implementation Plan: JobLinker UI/UX Consistency Fix

**Branch**: `004-homepage-ui-fix` | **Date**: 2026-04-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-homepage-ui-fix/spec.md`

## Summary

Fix UI/UX consistency issues across ALL JobLinker pages including: layout centering, typography hierarchy, button consistency, card uniformity, and brand consistency (slate color palette). Currently pages use inconsistent max-widths, gray colors instead of slate, and missing gradient backgrounds.

## Technical Context

**Language/Version**: TypeScript 5.x, Next.js 16, React 19
**Primary Dependencies**: Tailwind CSS v4, Framer Motion, Heroicons
**Storage**: N/A (pure frontend styling fix)
**Testing**: Playwright E2E, Lighthouse Accessibility
**Target Platform**: Web (responsive: 320px to 1920px+)
**Project Type**: Web application - Next.js frontend (12 pages total)
**Performance Goals**: FCP < 1.5s, CLS < 0.1, Lighthouse Accessibility ≥ 90
**Constraints**: Must use existing Tailwind utilities, maintain Plus Jakarta Sans font
**Scale/Scope**: 12 pages requiring consistent styling updates

## Constitution Check

*Gate: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| III. User Experience Consistency | ⚠️ VIOLATED | Pages use inconsistent colors (gray vs slate), widths (max-w-4xl vs max-w-7xl), missing gradient backgrounds |
| IV. Performance Requirements | ⚠️ UNVERIFIED | FCP/CLS targets stated but not measured pre/post fix |

**GATE RESULT**: Can proceed with UI-only changes. Performance measurement should be done post-fix to establish baseline.

## Project Structure

### Documentation (this feature)

```text
specs/004-homepage-ui-fix/
├── plan.md              # This file
├── research.md          # Phase 0 output (UI best practices research)
├── data-model.md        # N/A - no data entities in pure UI fix
├── quickstart.md        # N/A - no integration scenarios
├── contracts/           # N/A - no API contracts
└── tasks.md             # Phase 2 output
```

### Source Code (12 pages to fix)

```text
frontend/src/app/
├── page.tsx                    # Homepage (FR-001, FR-003, FR-010, FR-011, FR-012)
├── layout.tsx                  # Root layout (header/footer logo consistency)
├── login/page.tsx             # Login (add gradient wrapper)
├── register/page.tsx          # Register (add gradient wrapper)
├── dashboard/page.tsx         # Dashboard (verify gradient, consistency)
├── jobs/page.tsx              # Jobs (max-w-4xl→max-w-7xl, add gradient)
├── agents/page.tsx           # Agents (max-w-4xl→max-w-7xl, add gradient)
├── matches/page.tsx          # Matches (max-w-4xl→max-w-7xl, add gradient)
├── offers/page.tsx           # Offers (max-w-4xl→max-w-7xl, add gradient)
├── interviews/page.tsx       # Interviews (max-w-4xl→max-w-7xl, add gradient)
├── messages/page.tsx         # Messages (max-w-4xl→max-w-7xl, add gradient)
├── settings/page.tsx         # Settings (max-w-2xl→max-w-7xl, gray→slate)
└── admin/page.tsx            # Admin (gray→slate palette)
```

**Structure Decision**: Pure frontend styling changes across 12 pages. Target files listed above.

## Design Decisions

### Layout Container (FR-001, FR-002)

**Choice**: All interior pages use identical structure:
```tsx
<div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
  <div className="max-w-7xl mx-auto px-4 py-8">
    {/* content */}
  </div>
</div>
```

**Pages needing gradient wrapper update**:
- /jobs, /agents, /matches, /offers, /interviews, /messages - currently missing gradient
- /login, /register - have BackgroundBeams but need consistent inner container

**Pages needing max-width update**:
- `max-w-4xl` → `max-w-7xl`: jobs, agents, matches, offers, interviews, messages
- `max-w-2xl` → `max-w-7xl`: settings (too narrow)

### Typography (FR-003, FR-004)

**Choice**:
- H1: `text-3xl lg:text-4xl` (page titles - was text-5xl on homepage)
- H2: `text-2xl lg:text-3xl` (section headings)
- H3: `text-lg lg:text-xl` (card headings)
- Body: `text-base text-slate-600 leading-relaxed`
- ALL text MUST use slate palette NOT gray

**Pages using gray palette (MUST FIX)**:
- settings/page.tsx: `text-gray-900`, `text-gray-500`, `bg-gray-200` → slate equivalents
- admin/page.tsx: `bg-gray-50`, `text-gray-900`, `border-gray-200` → slate equivalents

### Buttons (FR-005, FR-006, FR-007)

**Choice**: Consistent padding `px-6 py-3`, 200ms transitions, visible focus rings
```tsx
className="px-6 py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl
         hover:from-blue-700 hover:to-sky-600
         focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
         transition-all duration-200"
```

### Cards (FR-008, FR-009)

**Choice**: All cards use `rounded-xl p-5` with `hover:-translate-y-0.5 hover:shadow-lg`

Current inconsistencies:
- Empty state cards use `p-10` - should be `p-8` or consistent with other cards
- Some cards missing glass morphism (bg-white/80 backdrop-blur-xl)

### Logo Consistency (FR-012)

**Choice**: Both header and footer use identical logo: `w-10 h-10` container, `text-lg` text

Current footer logo is `w-6 h-6 text-xs` - needs update to match header `w-10 h-10 text-lg`

### Brand Colors (FR-011, FR-013)

**Choice**: Use slate palette consistently
- ✅ `text-slate-900`, `bg-slate-50`, `border-slate-200`
- ❌ `text-gray-900`, `bg-gray-50`, `border-gray-200`

Pages with gray palette violations:
- settings/page.tsx
- admin/page.tsx

## Complexity Tracking

> No complexity violations. This is a straightforward styling fix across 12 pages.

## Implementation Notes

1. **Approach**: Fix one page type at a time (interior pages first, then homepage)
2. **Testing**: Browser screenshot at 1440px to verify centering per page
3. **Verification**: Lighthouse accessibility audit post-fix
4. **Order**: Settings/Admin (gray→slate) → Interior pages (gradient+max-width) → Homepage (final polish)
