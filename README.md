# JobLinker

AI Recruiting Agent — autonomous job search and application system.

---

# Current Status

✅ Phase 1 — Job Search Agent (complete)
✅ Phase 2 — Auto Apply Agent (complete)

---

# Architecture

```
frontend/              # Next.js 16 (App Router)
├── src/app/
│   ├── page.tsx       # Server component shell (no "use client")
│   └── layout.tsx     # Root layout
├── src/components/
│   ├── AppShell.tsx         # Client island: session state + grid
│   ├── SearchIsland.tsx     # Client: useChat + job cards + apply
│   ├── ResumeChatIsland.tsx # Client: PDF upload + resume chat
│   ├── ApplicationCard.tsx  # Client: cover letter + resume tabs
│   └── JobCard.tsx          # Presentational job card
├── src/lib/
│   ├── types.ts      # Job, RankedJob, Application types
│   └── chat.ts       # useSSEChat hook + helpers
└── src/components/ai-elements/   # Chat UI primitives (Streamdown, etc.)

cmd/server/main.go     # Go HTTP server
internal/
├── agent/             # Search agent (Eino ReAct) + Resume agent
├── tools/
│   ├── applyjob/      # Generate cover letter + tailored resume
│   ├── queryjobs/     # Firecrawl job search (Indeed China)
│   └── parseresume/   # PDF resume parsing via Firecrawl
├── checkpoint/        # Candidate profile store
├── job/               # Shared types (SearchResponse, etc.)
└── config/            # Environment configuration
```

---

# Frontend Architecture

Follows **Next.js App Router** pattern:
- `page.tsx` is a **server component** (no `"use client"`) — static shell
- Interactive islands are **client components** with `"use client"`
- Chat uses **AI SDK v6** with `useChat` + `DefaultChatTransport`
- Resume chat uses raw SSE → custom `useSSEChat` hook
- Structured data (jobs, applications) sent via AI SDK `data-*` events

---

# API Endpoints

| Method | Path | Protocol | Description |
|--------|------|----------|-------------|
| POST | `/api/chat` | AI SDK v6 SSE | Search agent (streaming) |
| POST | `/api/chat/resume` | Raw SSE | Resume agent chat |
| POST | `/api/apply` | JSON REST | Generate application |
| POST | `/api/resume/upload` | FormData | Upload + parse PDF |
| GET  | `/health` | JSON REST | Health check |

---

# Setup

```bash
# Backend
cp .env.example .env   # Configure OPENAI_API_KEY, FIRECRAWL_API_KEY, etc.
go build ./cmd/server/
./server

# Frontend
cd frontend
npm install
npm run dev
```

---

# Roadmap

## Phase 1 — Job Search Agent ✅
Query, rank, and explain job recommendations.

## Phase 2 — Auto Apply Agent ✅
Parse resumes, generate cover letters and tailored resumes.

## Phase 3 — Recruiter Chat
Multi-turn conversations with recruiters.

## Phase 4 — A2A Recruiter
Agent-to-agent recruiting communication.

## Phase 5 — Negotiation
Autonomous offer negotiation.

## Phase 6 — Memory
Long-term candidate understanding.

## Phase 7 — Autonomous Recruiting
End-to-end autonomous workflow.
