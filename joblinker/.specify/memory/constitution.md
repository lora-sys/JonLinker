# JobLinker Constitution

## Core Principles

### I. Test-First Development (NON-NEGOTIABLE)

All features MUST be implemented test-first. The Red-Green-Refactor cycle is mandatory:
1. Write a failing test that describes the desired behavior
2. Implement the minimum code to make the test pass
3. Refactor for clarity and performance only after tests pass

- Unit tests REQUIRED for all business logic
- Integration tests REQUIRED for API contracts and service boundaries
- Test coverage MUST exceed 80% for all new code
- No feature is complete until tests pass in CI

**Rationale**: Prevents regression, enforces clear contracts, enables confident refactoring.

### II. Code Quality Gates

- **Type Safety**: All code MUST be statically typed; no `any` or untyped interfaces
- **Linting**: Code MUST pass all lint rules before merge; `npm run lint` (or equivalent) must be green
- **No Magic Strings**: Shared constants, enum values, and config keys MUST be centralized
- **Single Responsibility**: Every module/function has one job; no God classes or 200-line functions
- **Dependency Direction**: High-level modules depend on abstractions, not implementations

**Rationale**: Technical debt compounds; quality gates prevent accumulation.

### III. User Experience Consistency

All user-facing interfaces MUST follow these rules:
- **Responsive**: Works on mobile (320px+), tablet, and desktop breakpoints
- **Accessible**: WCAG 2.1 AA compliance; keyboard navigable, screen-reader friendly
- **Consistent Patterns**: Same component used same way everywhere; no behavior surprises
- **Loading States**: Every async operation shows loading indicator; no silent failures
- **Error Messages**: User-friendly messages, never raw stack traces in UI

**Rationale**: Inconsistent UX erodes user trust; accessibility is not optional.

### IV. Performance Requirements

Performance is a feature. All implementations MUST meet these targets:
- **Time to Interactive (TTI)**: Under 3 seconds on 3G connections
- **Core Web Vitals**: LCP < 2.5s, FID < 100ms, CLS < 0.1
- **API Response**: 95th percentile response under 200ms for read operations
- **Bundle Size**: Initial JS payload under 150KB gzipped
- **Memory**: No memory leaks; GC pauses under 50ms

**Rationale**: Performance directly impacts user retention and conversion.

### V. Observability & Debugging

All services MUST emit structured observability signals:
- **Structured Logging**: JSON logs with correlation IDs, user context, operation timing
- **Metrics**: Key business metrics exposed (request count, error rate, latency p50/p95/p99)
- **Tracing**: Distributed trace propagation across service boundaries
- **Health Checks**: `/health` endpoint returning service status and dependency health

**Rationale**: Systems without observability cannot be diagnosed; debugging production without logs is unacceptable.

## Additional Constraints

### Security Requirements

- No secrets in code or version control; use environment variables or secret management
- All user input MUST be validated and sanitized before processing
- Authentication tokens MUST expire; refresh tokens rotated on each use
- Dependencies audited for known vulnerabilities before each release

### AI Model Configuration

All AI model credentials and endpoints MUST be configured via environment variables:
- **API Keys**: Stored in `AI_API_KEY` env var, never in code or version control
- **Base URL**: Stored in `AI_BASE_URL` env var for endpoint configuration
- **Model Selection**: Model name/id via `AI_MODEL` env var
- **Temperature/Parameters**: Via `AI_TEMPERATURE`, `AI_MAX_TOKENS` env vars if needed

**Rationale**: AI provider credentials are sensitive. Centralizing in `.env` files ensures no leaks to version control and enables runtime configuration per environment.

### Data Management

- Database migrations are version-controlled and reversible
- Schema changes are backward-compatible or accompanied by migration scripts
- Sensitive data encrypted at rest and in transit
- Data retention policies documented and enforced

### Deployment & Release

- All deployments are repeatable via automation (no manual server touching)
- Feature flags control rollouts; emergencies can disable features instantly
- Semantic versioning for all public APIs and packages
- Release notes REQUIRED for any breaking change

## Development Workflow

### Required Process

1. **Research First**: Search existing patterns before writing new code
2. **Specification**: All features documented in spec.md before implementation begins
3. **Tests Before Code**: Unit and integration tests written until they fail, then implemented
4. **Code Review**: All changes reviewed before merge; reviewer MUST verify quality gates pass
5. **Continuous Integration**: Pipeline runs tests, lint, type-check on every PR

### Review Requirements

- Code review MUST verify: tests exist, lint passes, types correct, no regressions
- Security-sensitive changes require security specialist review
- UX changes require visual regression check (or explicit design sign-off)

## Governance

This constitution supersedes all other development practices. Amendments require:
1. Proposal documented with rationale and migration impact
2. Team review period (minimum 3 business days)
3. Explicit approval from project lead
4. Version increment (MAJOR for removals, MINOR for additions, PATCH for clarifications)

All PRs and issues MUST reference constitution compliance verification.

**Version**: 1.1.0 | **Ratified**: 2026-04-22 | **Last Amended**: 2026-04-23

## Appendix: E2E Testing

### E2E Testing Tools (REQUIRED)

All end-to-end testing for this project MUST use:

1. **agent-browser** - Primary browser automation (native Rust CLI, headless by default)
   - No X server/xvfb dependency required
   - Fast accessibility-tree snapshots with compact `@eN` refs
   - Usage: `agent-browser open <url>` → `agent-browser snapshot -i` → `agent-browser click @e3`

2. **playwright MCP** - For Next.js DevTools integration when available
   - Use for Next.js app debugging and component inspection

**Why not xvfb + playwright CLI**: Environment lacks X server, xvfb not installable without sudo.

**Running E2E Tests**:
```bash
# Verify services are running
curl http://localhost:3000   # Frontend
curl http://localhost:8080/health  # Backend

# Open browser for testing
agent-browser open http://localhost:3000
agent-browser snapshot -i

# Example: Test login flow
agent-browser fill @e5 "user@test.com"
agent-browser fill @e6 "password123"
agent-browser click @e3  # Sign in button
agent-browser wait --url "**/dashboard"
```

**Context for this project**:
- Frontend: Next.js 16 on port 3000
- Backend: Go/Gin on port 8080
- Services: Postgres (5432), Redis (6379), RabbitMQ (5672), Chroma (8000)