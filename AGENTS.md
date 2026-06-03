# AGENT.md

## Mission

JobLinker exists to build an autonomous AI recruiting system.

Candidate Agent → Search Jobs → Apply Jobs → Communicate → Negotiate → Receive Offer → Human Confirmation

Human involvement should continuously decrease as the system evolves.

---

## Core Principle

DO NOT BUILD THE FINAL SYSTEM FIRST.

Every feature must belong to a specific phase. If a feature is not required by the current phase, DO NOT IMPLEMENT IT.

---

## Product Vision

We ARE building: **AI Recruiter**

We are NOT building: Agent Framework, Workflow Engine, MCP Platform, Multi-Agent Operating System, Generic AI Infrastructure

---

## Current Architecture

### Frontend (Next.js 16 App Router)
- `page.tsx` — **server component** (no `"use client"`)
- `AppShell.tsx` — client wrapper: `sessionId` state + 2-column grid
- `SearchIsland.tsx` — `useChat` from `@ai-sdk/react` v3, `DefaultChatTransport`
- `ResumeChatIsland.tsx` — custom `useSSEChat` hook for resume agent
- Types in `lib/types.ts`, helpers in `lib/chat.ts`
- AI SDK v6 protocol: `start` → `text-start` → `text-delta*` → `text-end` → `data-*` → `finish`

### Backend (Go)
- `cmd/server/main.go` — routes, handlers, SSE streaming
- `internal/agent/` — `Agent` (Eino ReAct, 15 steps, 120s) + `ResumeAgent`
- `internal/tools/` — `applyjob`, `queryjobs`, `parseresume`
- `internal/checkpoint/` — `compose.CheckPointStore`
- `internal/job/` — shared types (`SearchResponse`, etc.)

---

## Development Rules

1. Current phase completion > future architecture
2. Every feature must map to user value: Find Jobs, Apply Jobs, Recruiter Chat, Offer Negotiation
3. No future-proof engineering — build only what is needed now
4. No new infrastructure (RabbitMQ, Kafka, Event Bus, Workflow Engine, Distributed System) until required
5. Every phase must have a demo

---

## Success Metric

How much recruiting work can be delegated to AI.

Not by: number of services, agents, databases, abstractions, or architecture diagrams.

---

## Build Verification

```bash
# Backend
cd /home/lora/repos/joblinker
go build ./...
go test ./...

# Frontend
cd /home/lora/repos/joblinker/frontend
npx tsc --noEmit
npx next build
```

## Code Review History

- **Architecture**: RSC + client islands, AI SDK v6 protocol, Go service layer
- **Dead code removed**: 6 frontend files, ~55 unused exports, `ClearSession()`, `UserIntent`, custom `itoa`
- **Fixes applied**: `atomic.Int64` for session counter, per-session mutex for TOCTOU, 30min TTL cleanup goroutine, 30s HTTP client timeout, health endpoint, consistent SSE errors
- **Package naming**: `apply_job` → `applyjob`, `parse_resume` → `parseresume`, `query_jobs` → `queryjobs`
