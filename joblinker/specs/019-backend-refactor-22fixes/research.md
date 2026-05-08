# Research: Backend Refactor — Fix 22 Issues

**Date**: 2026-05-08
**Plan**: [plan.md](plan.md)

## R1: Chroma Collection Name Resolution

**Decision**: Add a `collections` map to the Chroma Client struct that caches name→UUID mappings. Populate via `GetOrCreateCollection` on first access.

**Rationale**: The current Chroma client passes collection names directly to API endpoints, but Chroma internally requires UUIDs. The `GetOrCreateCollection` method already exists and returns the UUID — we just need to cache it.

**Alternatives considered**:
- Pass UUIDs from callers: Rejected — callers shouldn't know about Chroma internals
- Call `GetOrCreateCollection` on every request: Rejected — unnecessary latency
- Use a persistent cache (Redis/DB): Rejected — overkill for 2 collections; in-memory map suffices

## R2: Extraction Function Design

**Decision**: Use regex-based extraction with curated keyword lists. Fallback returns a substring of the original content (first N chars).

**Rationale**: The existing `extractSalaryValue` already uses regex patterns successfully. Extending this pattern to location and skills is consistent and testable.

**Alternatives considered**:
- Call AI for extraction: Rejected — too expensive for every message; keyword matching is sufficient for structured fields
- NLP library: Rejected — adds dependency; regex + keyword list handles 90% of cases
- Return empty string on failure: Rejected — loses information; substring fallback preserves context

**Keyword lists**:
- Location: remote, onsite, hybrid, and top 20 US/international tech cities
- Skills: 50+ common programming languages, frameworks, tools (Go, Python, Java, JavaScript, TypeScript, React, Node, Docker, Kubernetes, AWS, etc.)

## R3: AI Integration for Negotiation/Reasoning

**Decision**: Add `aiClient *ai.Client` field to `DualAgentNegotiationService` and `ReasoningService`. Call `aiClient.Chat(systemPrompt, userPrompt)` with context-aware prompts. Fallback to generic text messages.

**Rationale**: The `ai.Client` already exists with a working `Chat()` method. Adding it as a dependency is straightforward constructor injection.

**Alternatives considered**:
- Use Eino agents directly: Rejected — adds complexity; simple Chat() suffices for proposal generation
- Use `ChatWithTools`: Rejected — negotiation proposals don't need tool calls
- No AI, just better templates: Rejected — doesn't meet the requirement for "real" AI responses

**Fallback strategy**:
- `generateSeekerProposal`: "I'd like to discuss the compensation package further." (no round number)
- `understandPhase`: Keyword matching (existing behavior)
- `respondPhase`: "I'll review the details and get back to you."

## R4: Hardcoded Value Removal Strategy

**Decision**: Replace hardcoded salary numbers with text-only messages in fallback paths. For salary resolution failures, return an error instead of a magic number.

**Rationale**: Fallback responses should never contain specific business values. If the system can't determine a real salary, it should say so rather than invent one.

**Alternatives considered**:
- Use environment variables for defaults: Rejected — still hardcoded, just in a different place
- Use industry averages: Rejected — too variable; better to not guess
- Log warning and continue with hardcoded: Rejected — the whole point is to eliminate hardcoded values

## R5: Route Conflict Resolution

**Decision**: Rename `/interviews/:matchId` to `/interviews/match/:matchId` to disambiguate from `/interviews/:id`.

**Rationale**: Gin's router treats `:id` and `:matchId` as the same pattern (both match any path segment). The `/interviews/match/:matchId` prefix makes the intent explicit.

**Alternatives considered**:
- Rename `:id` to `/interviews/detail/:id`: Also valid but less intuitive
- Use query parameters: Rejected — changes API contract
- Use different HTTP methods: Rejected — both are GET endpoints

## R6: Dead Code Identification

**Decision**: Delete `internal/agent/memory.go` (replaced by Eino memory), `GenerateEmbedding` (replaced by Chroma), and the API key length log line.

**Rationale**: These are confirmed dead code:
- `agent/memory.go`: The Eino `memory/agent_memory.go` and `service/agent_memory_service.go` supersede it
- `GenerateEmbedding`: Chroma handles all embeddings via its own embedding function
- API key length log: Security risk (leaks key length info)

**Verification**: `grep -rn` confirms no imports or usages of the deleted code paths.
