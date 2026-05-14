<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:

specs/022-frontend-hardening/plan.md

Current feature scope: Frontend Hardening & E2E Integration Coverage.
Phase 1: GatewayClient auth integration + client-side fetch() migration
Phase 2: E2E test enhancement (gateway, auth, isolation coverage)
Phase 3: Full E2E test suite execution and verification

## Key Design Decisions

1. API calls from client components use `apiClient` from `lib/api_client.ts` (wraps GatewayClient) instead of raw fetch()
2. Server Components fetch data server-side with server auth token, skipping GatewayClient headers
3. E2E tests use playwright-cli for browser automation + curl/python for reliable API verification
4. GatewayClient singleton initialized lazily via auth store on login; X-User-ID/X-Tenant-ID headers set from auth state
5. Next.js API routes act as BFF, converting auth cookie to Bearer token; client components bypass them via apiClient
<!-- SPECKIT END -->
