<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan:

specs/025-agent-chat-arch/plan.md

Current feature scope: Agent Chat Architecture — Backend Single Source of Truth.
Phase 1: Rewrite useAIChat — REST poll + submit, useChat as UI only
Phase 2: Delete /api/chat/route.ts (frontend AI bypass)
Phase 3: Register GetConversation backend route
Phase 4: Playwright CLI E2E full flow testing

## Key Design Decisions

1. API calls from client components use `apiClient` from `lib/api_client.ts` (wraps GatewayClient) instead of raw fetch()
2. Server Components fetch data server-side with server auth token, skipping GatewayClient headers
3. E2E tests use playwright-cli for browser automation + curl/python for reliable API verification
4. GatewayClient singleton initialized lazily via auth store on login; X-User-ID/X-Tenant-ID headers set from auth state
5. Next.js API routes act as BFF, converting auth cookie to Bearer token; client components bypass them via apiClient
<!-- SPECKIT END -->
