<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:

specs/024-compile-fix-e2e/plan.md

Current feature scope: TypeScript Compile Fixes & Playwright CLI E2E Coverage.
Phase 1: Fix all compile errors (api_client.ts, MatchStatus, useChat types)
Phase 2: Playwright CLI pure-browser E2E tests T01-T16
All tests use playwright-cli commands only — no curl, python, or bash scripts

## Key Design Decisions

1. API calls from client components use `apiClient` from `lib/api_client.ts` (wraps GatewayClient) instead of raw fetch()
2. Server Components fetch data server-side with server auth token, skipping GatewayClient headers
3. E2E tests use playwright-cli for browser automation + curl/python for reliable API verification
4. GatewayClient singleton initialized lazily via auth store on login; X-User-ID/X-Tenant-ID headers set from auth state
5. Next.js API routes act as BFF, converting auth cookie to Bearer token; client components bypass them via apiClient
<!-- SPECKIT END -->
