# Feature Specification: JobLinker UI/UX Consistency Fix

**Feature Branch**: `004-homepage-ui-fix`
**Created**: 2026-04-24
**Status**: Draft
**Input**: Fix UI/UX issues across ALL pages - layout, typography, buttons, cards, consistency

## Scope

This fix covers ALL pages in the JobLinker application:
- Homepage (`/`)
- Login (`/login`)
- Register (`/register`)
- Dashboard (`/dashboard`)
- Jobs (`/jobs`)
- Agents (`/agents`)
- Matches (`/matches`)
- Offers (`/offers`)
- Interviews (`/interviews`)
- Messages (`/messages`)
- Settings (`/settings`)
- Admin (`/admin`)

## User Scenarios & Testing

### User Story 1 - Visual Layout & Spacing (Priority: P1)

As a visitor, I want to see all page content properly centered with consistent spacing so I can comfortably read and navigate.

**Why this priority**: First impression matters - cramped or misaligned content drives visitors away immediately

**Independent Test**: Load each page at 1440px width, verify content is centered and readable without horizontal scroll

**Acceptance Scenarios**:

1. **Given** visitor views any page at desktop width, **When** content is rendered, **Then** all content is horizontally centered within max-width container (1280px/max-w-7xl)
2. **Given** visitor views any page at tablet/mobile, **When** content is rendered, **Then** layout adapts responsively without horizontal overflow
3. **Given** visitor scrolls through page, **When** content sections appear, **Then** vertical spacing between sections is consistent (multiples of 8px baseline: py-8, py-12, py-16)
4. **Given** visitor views interior page (not homepage), **When** page loads, **Then** page has consistent gradient background wrapper

---

### User Story 2 - Typography Hierarchy (Priority: P1)

As a visitor, I want clear visual hierarchy between headings, subheadings, and body text so I can quickly scan and understand the page content.

**Why this priority**: Poor typography makes content hard to read and reduces credibility

**Independent Test**: Visually inspect typography scale - each level (H1, H2, H3, body, caption) must be clearly distinguishable

**Acceptance Scenarios**:

1. **Given** visitor views page section, **When** main heading is displayed, **Then** font size, weight, and line-height clearly distinguish it from body text
2. **Given** visitor reads section content, **When** paragraph text is displayed, **Then** line-height is minimum 1.5 for comfortable reading
3. **Given** visitor scans page, **When** headings appear, **Then** clear visual progression from H1 → H2 → H3 with decreasing prominence
4. **Given** visitor reads text across pages, **When** color is observed, **Then** all pages use slate color palette (text-slate-*) not gray (text-gray-*)

---

### User Story 3 - Button Consistency (Priority: P1)

As a visitor, I want consistently styled buttons with clear hover feedback so I can confidently click interactive elements.

**Why this priority**: Inconsistent or unresponsive buttons create doubt and reduce conversion

**Independent Test**: Tab through all buttons - each must have visible focus state and smooth hover transition

**Acceptance Scenarios**:

1. **Given** visitor hovers over primary button, **When** mouse enters button area, **Then** background color transitions smoothly within 200ms
2. **Given** visitor hovers over secondary/ghost button, **When** mouse enters button area, **Then** border or background changes to indicate interactivity
3. **Given** visitor tabs to button, **When** button receives focus, **Then** visible focus ring appears (outline or shadow)
4. **Given** visitor compares buttons across pages, **When** both are displayed, **Then** they follow consistent button styling rules (same padding, same transition timing)

---

### User Story 4 - Card Component Uniformity (Priority: P2)

As a visitor, I want uniformly styled cards with consistent spacing and hover effects so the page feels polished and professional.

**Why this priority**: Inconsistent card styling makes the page appear amateur and reduces trust

**Independent Test**: Compare all card components - they should have identical border-radius, padding, shadow, and hover behavior

**Acceptance Scenarios**:

1. **Given** visitor views any page with cards, **When** cards are displayed, **Then** all cards have identical padding (p-5/20px), border-radius (rounded-xl/12px), and shadow
2. **Given** visitor hovers over any card, **When** mouse enters card area, **Then** card lifts with subtle shadow increase (hover:-translate-y-0.5 hover:shadow-lg transition)
3. **Given** visitor views empty state cards, **When** cards appear, **Then** they have glass morphism styling (bg-white/80 backdrop-blur-xl border border-white/20)

---

### User Story 5 - Brand Consistency (Priority: P2)

As a visitor, I want consistent branding throughout the site so the site feels cohesive and trustworthy.

**Why this priority**: Inconsistent branding (different blues, misaligned logos) undermines professional credibility

**Independent Test**: Compare logo, header, footer, and button colors - all should use the same blue palette

**Acceptance Scenarios**:

1. **Given** visitor views header and footer (on homepage), **When** both are displayed, **Then** logo appears identical in both locations (w-10 h-10 container, text-lg font)
2. **Given** visitor compares blue colors across page, **When** observing buttons and accents, **Then** all blues are within same hue family (blue-600 family: from-blue-600 to-sky-500)
3. **Given** visitor scrolls to footer, **When** footer content appears, **Then** clear visual separation from main content with appropriate spacing
4. **Given** visitor views any interior page, **When** page loads, **Then** it has consistent gradient background (bg-gradient-to-br from-slate-50 via-blue-50/30 to-white)

---

## Pages Inventory & Issues

### Interior Pages (Need gradient wrapper + consistent max-width)
| Page | Current max-width | Issue |
|------|------------------|-------|
| /login | N/A | Missing gradient background |
| /register | N/A | Missing gradient background |
| /dashboard | N/A | Has gradient, needs review |
| /jobs | max-w-4xl | Wrong width, missing gradient wrapper |
| /agents | max-w-4xl | Wrong width, missing gradient wrapper |
| /matches | max-w-4xl | Wrong width, missing gradient wrapper |
| /offers | max-w-4xl | Wrong width, missing gradient wrapper |
| /interviews | max-w-4xl | Wrong width, missing gradient wrapper |
| /messages | max-w-4xl | Wrong width, missing gradient wrapper |
| /settings | max-w-2xl | Too narrow, uses gray-* colors |
| /admin | max-w-7xl | Uses gray-* colors instead of slate |

### Color Palette Issue
- **Correct**: `text-slate-900`, `bg-slate-50`, `text-slate-600`
- **Incorrect**: `text-gray-900`, `bg-gray-50`, `text-gray-600`
- Pages using gray palette: settings, admin

## Requirements

### Functional Requirements

- **FR-001**: All page content MUST be contained within a centered container with max-width of 1280px (max-w-7xl)
- **FR-002**: Interior pages (non-homepage) MUST have gradient background wrapper (bg-gradient-to-br from-slate-50 via-blue-50/30 to-white)
- **FR-003**: Typography hierarchy MUST use distinct size/weight combinations: H1 (40-48px bold), H2 (28-32px semibold), H3 (20-24px semibold), Body (16px regular), Caption (14px)
- **FR-004**: Body text line-height MUST be minimum 1.5 for readability (leading-relaxed or leading-7)
- **FR-005**: All buttons MUST have consistent padding (minimum 12px vertical, 24px horizontal)
- **FR-006**: All button hover transitions MUST complete within 200ms using ease-out timing (transition-all duration-200)
- **FR-007**: All interactive buttons MUST show visible focus state for keyboard navigation (focus:ring-2 focus:ring-blue-500)
- **FR-008**: All card components MUST have identical border-radius (12px/rounded-xl) and padding (20px/p-5)
- **FR-009**: All cards with hover effect MUST use identical lift animation (translateY(-2px) + shadow increase via hover:-translate-y-0.5 hover:shadow-lg)
- **FR-010**: Hero section heading MUST be responsive: 48px at desktop, 32px at tablet, 24px at mobile
- **FR-011**: Primary color blue MUST be consistent (#0369A1 or equivalent blue-600 family) across all components
- **FR-012**: Logo MUST appear identically in header and footer (same size w-10 h-10, same styling text-lg)
- **FR-013**: All pages MUST use slate color palette (text-slate-*, bg-slate-*, border-slate-*) NOT gray palette
- **FR-014**: All interior pages MUST use same container structure: gradient wrapper + centered content container

## Success Criteria

### Measurable Outcomes

- **SC-001**: Content is horizontally centered at all viewport widths ≥ 320px on all pages
- **SC-002**: No horizontal scrollbar appears at any viewport width from 320px to 1920px on all pages
- **SC-003**: All heading levels (H1-H3) are visually distinguishable without counting pixels
- **SC-004**: All buttons respond to hover with smooth transition (no instant color change)
- **SC-005**: All card components have identical visual styling (verified by side-by-side comparison)
- **SC-006**: Lighthouse Accessibility score ≥ 90
- **SC-007**: All pages use consistent slate color palette (zero instances of text-gray-* or bg-gray-* in content areas)

## Assumptions

- Existing color palette (#0369A1 primary) is acceptable and should be standardized across components
- Using Tailwind CSS utility classes for all styling (no custom CSS unless Tailwind cannot achieve the effect)
- Heroicons are used for all icons (already in use)
- Plus Jakarta Sans font is acceptable for body text (already in use)
- Mobile-first responsive breakpoints: 640px (sm), 768px (md), 1024px (lg), 1280px (xl)
- No dark mode required for this fix (out of scope)
- Admin page is separate (has its own admin layout) but should still use slate palette
