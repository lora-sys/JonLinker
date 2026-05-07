# Tasks: JobLinker Bug Fixes (Island Architecture)

## Architecture Context

**Island Architecture** - Next.js Server Components + Client Components
- Server Components: fetch data directly, no client-side state needed
- Client Components: 'use client' only where interaction required
- Auth: cookie-based, server reads directly from cookie

## Phase 1: Critical Issues

### T001: Agent Persistence - Convert agents/page to Server Component

**Problem**: Client-side Zustand persist causes race conditions on rehydration

**Solution**:
- Convert `frontend/src/app/agents/page.tsx` to Server Component
- Fetch agents server-side using cookie for auth
- Pass data to Client Component via props

**Files**:
- `frontend/src/app/agents/page.tsx` - Convert to Server Component ✓
- `frontend/src/app/agents/AgentsContent.tsx` - Keep as Client Component for interactivity ✓

**Test**: Refresh page → agent list still shows

**Status**: ✅ COMPLETED

---

### T002: Match Creation - Add POST Handler

**Problem**: Frontend `/api/matches/route.ts` only has GET, missing POST proxy

**Files**:
- `frontend/src/app/api/matches/route.ts` - Add POST handler ✓

**Test**: Can create match via API

**Status**: ✅ COMPLETED

---

### T003: Verify Match Auto-Create Endpoint

**Files**:
- `backend/internal/handler/match.go` - Check AutoCreate handler ✓
- Test with curl

**Status**: ✅ COMPLETED (endpoint exists)

---

## Phase 2: High Priority Issues

### T004: Sign Up Navigation - Fix Cookie + Redirect

**Problem**: `document.cookie` set client-side not visible to middleware in same request

**Solution**:
- Use `NextResponse.redirect()` with cookie set server-side
- Or use `window.location.href` for full page reload

**Files**:
- `frontend/src/app/register/page.tsx` - Fix redirect logic ✓
- `frontend/src/app/api/auth/register/route.ts` - Cookie set server-side ✓

**Status**: ✅ COMPLETED

---

### T005: Create Agent Form - Add Input Fields

**Problem**: Form only sends `{type: agentType}`

**Files**:
- `frontend/src/app/agents/create/page.tsx` - Add name, skills, preferences fields ✓

**Status**: ✅ COMPLETED

---

### T006: JSON Parsing Error Handling

**Problem**: `api_client.ts` does direct `response.json()` without try/catch

**Files**:
- `frontend/src/lib/api_client.ts` - Add error handling for malformed JSON ✓

**Status**: ✅ COMPLETED

---

## Phase 3: Medium Priority

### T007: Create Logout Endpoint

**Files**:
- `frontend/src/app/api/auth/logout/route.ts` - Create logout route ✓

**Status**: ✅ COMPLETED

---

### T008: Add Logout Button to Sidebar

**Files**:
- `frontend/src/components/layout/Sidebar.tsx` - Add logout button ✓

**Status**: ✅ COMPLETED

---

### T009: Server-Side Validation

**Files**:
- `backend/internal/handler/auth.go` - Add validation for role selection ✓

**Status**: ✅ COMPLETED (validation already exists via binding)

---

## Implementation Order

| Phase | Tasks | Approach |
|-------|-------|----------|
| 1 | T001-T003 | Server Component + API fix |
| 2 | T004-T006 | Navigation + Form + Error handling |
| 3 | T007-T009 | Logout + Validation |

**Parallel**: Tasks within same phase marked [P] can run together

**Total**: 9 tasks
