# Research: A2A Agent Recruitment Platform

## Technical Decisions Summary

| Decision | Choice | Rationale | Alternatives |
|----------|--------|-----------|--------------|
| Frontend Framework | Next.js 16 | App Router, RSC, fast builds, strong TypeScript | Plain React, Remix |
| Frontend State | Zustand | Minimal, TypeScript-native, simple API | Redux, Jotai |
| Styling | Tailwind CSS v5 + native CSS | Lightweight, utility-first | Styled-components |
| UI Components | Aceternity UI (minimal) | Lightweight, modern look | Radix, Headless UI |
| Backend Language | Go 1.23+ | Performance, strong typing, mature HTTP | Rust/Axia, Node.js |
| Backend Framework | Gin | Fast, middleware ecosystem | Echo, Fiber |
| ORM | GORM | Mature Go ORM, migrations | SQLx, ent |
| Business Database | PostgreSQL 17 | ACID, JSON support, mature | MySQL, CockroachDB |
| Vector Database | Chroma 0.6+ | Lightweight, embedded option, good for MVP | Pinecone, Weaviate |
| Cache | Redis 7.4+ | Session cache, pub/sub, speed | Memcached |
| Task Queue | RabbitMQ | Reliable delivery, routing | Kafka, SQS |
| Local Storage | IndexedDB + AES-256 | Client-side privacy, encryption | localForage |
| E2E Testing | Playwright | Next.js support, good DX | Cypress, Selenium |
| Container | Docker + Docker Compose | Reproducible dev/deploy | Kubernetes (future) |

## Architecture Research

### Privacy-First Design

Core constraint: Raw user data (resumes, contact info) never leaves the client.

**Implementation**:
- Client-side: IndexedDB stores encrypted resume data with AES-256-GCM
- Encryption key derived from user password (PBKDF2)
- Only structured anonymized data + vector embeddings sent to server
- Vector embeddings are one-way (cannot reverse-engineer original data)

**Server receives**:
- User account (email, hashed password)
- Agent configuration (preferences, boundaries)
- Structured JSON (skills list, experience summary — no names/contact)
- Vector embedding (float array representing skills/experience)

### Agent FSM States

```
Seeker Agent States:
  IDLE → PROFILE_BUILT → MATCHING → NEGOTIATING → INTERVIEW_SCHEDULED → OFFER_RECEIVED → HIRED/REJECTED
                     ↓              ↓
                 PAUSED         DEADLOCKED (max rounds)

Recruiter Agent States:
  IDLE → POSTING_ACTIVE → MATCHING → NEGOTIATING → INTERVIEW_SCHEDULED → OFFER_EXTENDED → FILLED/REJECTED
                     ↓              ↓
                 PAUSED         DEADLOCKED
```

### A2A XML Protocol

Based on request/response pattern with structured intent:

```xml
<!-- Salary negotiation -->
<a2a-message type="negotiate" id="msg-001" timestamp="2026-04-22T10:00:00Z">
  <sender agent-id="seeker-001" type="seeker"/>
  <intent type="salary" confidence="0.92">
    <proposal range="$80k-$90k" currency="USD"/>
  </intent>
</a2a-message>

<a2a-message type="negotiate" id="msg-002" timestamp="2026-04-22T10:01:00Z" ref-id="msg-001">
  <sender agent-id="recruiter-001" type="recruiter"/>
  <intent type="salary" confidence="0.88">
    <counter range="$85k-$95k" currency="USD"/>
  </intent>
</a2a-message>
```

## Performance Analysis

From PRD constraints:
- TTI ≤ 2s (more stringent than constitution's 3s)
- Agent response ≤ 1s
- Vector match ≤ 3s
- API p95 ≤ 200ms

**Optimization strategies**:
- SSR for initial page loads (Next.js)
- Lazy load non-critical routes
- WebSocket for real-time dialogue (not polling)
- Chroma approximate nearest neighbor (ANN) for fast vector search
- Redis cache for hot match data
- RabbitMQ for async negotiation processing

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Vector matching quality low | Medium | High | User feedback loop; adjustable thresholds |
| Agent negotiation deadlocks | Medium | Medium | Max round limit + human escalation |
| Privacy key loss | Low | Critical | Key derivation backup + recovery flow |
| Scale beyond MVP | Low | Medium | PostgreSQL + Chroma scale patterns established |
| Calendar integration failures | Medium | Low | Fallback to email-only scheduling |

## Conclusion

PRD provides sufficient technical direction. No additional research needed. Implementation can proceed with documented choices.