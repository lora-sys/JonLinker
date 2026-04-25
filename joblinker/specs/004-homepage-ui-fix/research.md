# Research: JobLinker UI/UX Consistency Fix

**Date**: 2026-04-24
**Feature**: 004-homepage-ui-fix

## Decision: Unified Page Container Pattern

**Choice**: All interior pages use identical gradient wrapper + centered container:

```tsx
<div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-white">
  <div className="max-w-7xl mx-auto px-4 py-8">
    {/* page content */}
  </div>
</div>
```

**Rationale**: Consistent user experience across all pages. The gradient creates visual continuity.

**Pages Requiring Container Update**:
| Page | Current Pattern | Issue |
|------|-----------------|-------|
| /jobs | `max-w-4xl mx-auto px-4 py-6` | Missing gradient, narrow width |
| /agents | `max-w-4xl mx-auto px-4 py-6` | Missing gradient, narrow width |
| /matches | `max-w-4xl mx-auto px-4 py-6` | Missing gradient, narrow width |
| /offers | `max-w-4xl mx-auto px-4 py-6` | Missing gradient, narrow width |
| /interviews | `max-w-4xl mx-auto px-4 py-6` | Missing gradient, narrow width |
| /messages | `max-w-4xl mx-auto px-4 py-6` | Missing gradient, narrow width |
| /settings | `max-w-2xl mx-auto px-4` | Missing gradient, too narrow, gray palette |
| /admin | Uses different structure | Uses gray palette instead of slate |

## Decision: Color Palette (Slate not Gray)

**Choice**: All pages MUST use slate color palette

```tsx
// CORRECT
text-slate-900    // headings
text-slate-600    // body text
text-slate-500    // muted text
bg-slate-50       // subtle backgrounds
border-slate-200  // borders

// INCORRECT (must fix)
text-gray-900
text-gray-600
bg-gray-50
border-gray-200
```

**Pages with Gray Palette Violations**:
- `frontend/src/app/settings/page.tsx` - uses gray-500, gray-900
- `frontend/src/app/admin/page.tsx` - uses gray-50, gray-200, gray-900

**Rationale**: The design system uses slate as the neutral color. Gray is only for true grayscale elements (like monochrome images). This ensures visual consistency.

## Decision: Typography Scale

**Choice**:
- H1 (Page titles): `text-3xl lg:text-4xl` (30px tablet, 36px desktop)
- H2 (Section titles): `text-2xl lg:text-3xl` (24px tablet, 30px desktop)
- H3 (Card titles): `text-lg lg:text-xl` (18px tablet, 20px desktop)
- Body: `text-base text-slate-600 leading-relaxed`

**Rationale**: Maintains readability hierarchy. Homepage hero is exception (larger for impact).

## Decision: Button Styling Standardization

**Choice**: Primary button pattern across all pages:
```tsx
className="px-6 py-3 bg-gradient-to-r from-blue-600 to-sky-500 text-white rounded-xl
         hover:from-blue-700 hover:to-sky-600
         focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
         transition-all duration-200"
```

**Secondary/Ghost button pattern**:
```tsx
className="px-6 py-3 bg-white text-blue-600 border-2 border-blue-200 rounded-xl
         hover:border-blue-400 hover:bg-blue-50
         focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
         transition-all duration-200"
```

## Decision: Card Uniformity

**Choice**: All cards use identical structure:
```tsx
// Standard card
<Card className="p-5 bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl">
  {/* content */}
</Card>

// Card with hover
<Card hover className="p-5 bg-white/80 backdrop-blur-xl border border-white/20 rounded-xl">
  {/* hover:-translate-y-0.5 hover:shadow-lg applied via hover prop */}
</Card>
```

**Current Inconsistencies Found**:
- Empty state cards use `p-10` (40px padding) instead of `p-5` (20px)
- Some cards missing glass morphism styling

## Decision: Logo Sizing

**Choice**: Logo appears identically in header and footer
- Container: `w-10 h-10` (both locations)
- Text: `text-lg` font size (both locations)

**Current Issue**: Footer uses `w-6 h-6 text-xs` which is different from header

## Accessibility Checklist

Per Constitution III (User Experience Consistency):
- [x] Responsive: Mobile-first breakpoints (sm:640px, md:768px, lg:1024px)
- [x] Accessible: Focus states on all interactive elements
- [x] Consistent: Same component patterns used throughout
- [ ] Loading States: Per-component (not all pages have async)
- [ ] Error Messages: Per-component (forms have error states)

## Performance Considerations

- No JavaScript changes - pure CSS/Tailwind modifications
- No bundle size impact expected
- CLS may improve with more predictable content heights
- FCP should remain under 1.5s (no blocking resources added)

## Verification Plan

1. **Browser Test**: Open each page at 1440px, verify:
   - Content centered with `max-w-7xl mx-auto`
   - No horizontal scroll
   - Gradient background visible

2. **Lighthouse**: Run accessibility audit on each page, target score ≥ 90

3. **Responsive Test**: Check at 375px, 768px, 1024px, 1440px breakpoints

4. **Visual Comparison**: Before/after screenshots at each breakpoint

5. **Color Audit**: Verify zero instances of `text-gray-*`, `bg-gray-*` in content areas (except admin purely-functional elements)
