# Quickstart: Next.js Architecture Refactor Testing Guide

**Date**: 2026-04-28
**Feature**: specs/008-nextjs-architecture-refactor

## Prerequisites

Ensure services are running before testing:
```bash
# Frontend (Next.js)
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000  # Should return 200

# Backend (Go)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health  # Should return 200
```

## Verification Checklist

### 1. Build Verification (MUST PASS)

```bash
cd frontend
npm run build
```

**Expected**: Build succeeds with no TypeScript errors, no lint warnings in modified files.

### 2. Bundle Size Check

Before refactor, note the client JS bundle size from `npm run build` output.
After refactor, bundle should be at least 20% smaller.

```bash
# Check bundle size
ls -la .next/static/chunks/*.js | wc -l
# Compare count before/after
```

### 3. SSR Verification

Verify that conversation page renders content server-side:

```bash
# Login first to get session cookie
curl -s -c /tmp/cookies.txt -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"etest@test.com","password":"password123"}'

# Fetch conversation page and check for server-rendered content
curl -s -b /tmp/cookies.txt http://localhost:3000/conversation/some-valid-match-id | \
  grep -o '<div class=".*message.*">' | head -5

# Should find message content in HTML, not just loading spinner
```

### 4. Component Classification Audit

Verify that only client-necessary components have 'use client':

```bash
# Count 'use client' directives before and after
grep -r "'use client'" src/app --include="*.tsx" | wc -l  # Should decrease
grep -r "'use client'" src/components --include="*.tsx" | wc -l  # Should decrease

# Verify no unnecessary 'use client' in static pages
grep -r "'use client'" src/app/page.tsx  # Should be empty or justified
```

### 5. Shared Component Usage

Verify pages use shared components instead of inline markup:

```bash
# Check that Card component is used in job listings
grep -r "components/ui/Card" src/app/jobs --include="*.tsx"

# Check that EmptyState is used
grep -r "EmptyState" src/app --include="*.tsx" | head -10
```

## E2E Smoke Tests (agent-browser)

### Test 1: Login Flow

```bash
agent-browser open --session smoke1 http://localhost:3000/login
agent-browser snapshot -i
# Verify login form visible
agent-browser fill e7 testuser@test.com
agent-browser fill e8 password123
agent-browser click e5  # Sign in
sleep 5
agent-browser snapshot -i
# Should redirect to dashboard
```

### Test 2: Job Posting (if recruiter)

```bash
agent-browser open --session smoke2 http://localhost:3000/login
# Login as recruiter
agent-browser fill e7 etest@test.com
agent-browser fill e8 password123
agent-browser click e5
sleep 5

# Navigate to jobs/new
agent-browser open http://localhost:3000/jobs/new
sleep 5
agent-browser snapshot -i
# Should see job form, not just "Loading..."
```

### Test 3: Conversation Page

```bash
# Login first
agent-browser open --session smoke3 http://localhost:3000/login
agent-browser fill e7 testuser@test.com
agent-browser fill e8 password123
agent-browser click e5
sleep 5

# Navigate to a conversation
agent-browser open http://localhost:3000/conversation/some-valid-match-id
sleep 5
agent-browser snapshot -i
# Should see message thread, not just loading
```

## Common Issues & Fixes

### Issue: Page stuck on "Loading..."

**Symptom**: /jobs/new or other protected pages show loading spinner forever.

**Cause**: Usually zustand hydration issue — `_hasRehydrated` flag not being set correctly.

**Fix**: Check `stores/auth.ts` persist configuration. Ensure `onRehydrateStorage` sets `_hasRehydrated: true`.

### Issue: Build fails after refactor

**Symptom**: `npm run build` fails with TypeScript errors.

**Cause**: Removed 'use client' from a component that still uses hooks.

**Fix**: Run `npm run build` in frontend directory to see exact error. Add 'use client' back to component that uses hooks.

### Issue: WebSocket not connecting

**Symptom**: Real-time messages don't appear in conversation.

**Cause**: WebSocket hook or page not marked 'use client'.

**Fix**: Ensure `use-websocket.ts` and any component using WebSocket has 'use client'.

### Issue: Auth redirect loop

**Symptom**: Browser redirects between login and protected page repeatedly.

**Cause**: zustand store not rehydrating before auth check runs.

**Fix**: The `_hasRehydrated` pattern should prevent this. Check that auth store is persisting correctly.

## Success Criteria

| Criterion | Threshold | How to Verify |
|-----------|-----------|---------------|
| Bundle size | <150KB gzipped (was ~200KB) | `npm run build` output |
| Initial load | <3s on 3G | Browser devtools network tab |
| Build | No errors | `npm run build` exit code 0 |
| TypeScript | No errors | `npm run build` TypeScript output |
| Login works | Redirects to dashboard | agent-browser smoke test |
| Job form loads | Form visible, not loading | agent-browser smoke test |
| Conversation loads | Messages in HTML | `curl` + grep |
