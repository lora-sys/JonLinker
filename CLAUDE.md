# CLAUDE.md

## Project Identity

JobLinker — AI Recruiting Agent

Goal: Reduce human involvement in recruiting through autonomous AI agents.

---

# Development Roadmap

## Phase 1 — Job Search Agent ✅

User Input → Search Jobs → Ranked Results

Tools: `queryjobs`

Done When:
- Agent can search jobs ✓
- Agent returns ranked jobs ✓
- Agent explains recommendations ✓

---

## Phase 2 — Auto Apply Agent ✅

User Input → Search Jobs → Apply Jobs

Tools: `queryjobs`, `applyjob`, `parseresume`

Components:
- `AppShell.tsx` — client wrapper, holds `sessionId`
- `SearchIsland.tsx` — `useChat` + job cards + apply flow
- `ResumeChatIsland.tsx` — PDF upload + chat-to-complete-profile
- `ApplicationCard.tsx` — cover letter / resume tabs
- `JobCard.tsx` — presentational job card

Endpoints:
- `POST /api/chat` — AI SDK v6 SSE: search agent (streaming)
- `POST /api/chat/resume` — raw SSE: resume agent chat
- `POST /api/apply` — JSON: generate application
- `POST /api/resume/upload` — FormData: PDF upload + parse
- `GET /health` — health check

Protocol:
- AI SDK v6 stream: `start` → `text-start` → `text-delta`* → `text-end` → `data-*` → `finish`
- Data events: `data-jobs` (array), `data-application` (object)

Done When:
- Resume Agent parses PDF + chat completes CandidateProfile ✓
- Search Agent searches + ranks jobs ✓
- Apply generates cover letter + tailored resume ✓
- All builds pass: `go build ./...`, `npx tsc --noEmit`, `npx next build` ✓

---

## Phase 3 — Recruiter Chat

User ↔ Agent (plays Recruiter)

### Architecture
- **Agent**: Eino ReAct Agent (`RecruiterAgent`), parallel with Search/Resume
- **Memory**: separate namespace (`sessionID+":recruiter"` in MemoryStore)
- **Entry**: UnifiedChatIsland (`/api/chat/unified`), no new UI
- **Routing**: parallel 3-way (Search/Resume/Recruiter), `last_agent` marker in CheckpointStore + keyword detection
- **Frontend**: pure text + tool visibility via AI Elements Tool/ToolHeader/ToolContent; no new `data-*` events

### Tools
- `get_candidate_profile` — reads CandidateProfile from CheckpointStore
- `record_interview_note` — writes InterviewNote to CheckpointStore (`sessionID+":interview_"+jobURL`)

### System Prompt
扮演招聘官 Agent，面试已申请职位的候选人。读取 CandidateProfile + Application，多轮问答，结束后调用 record_interview_note。

### Routing Logic
```
hasApplication := checkpoint.HasPrefix(sessionID+":application_")
keyword := detectInterviewKeyword(msg)

switch {
case keyword == "search" || (!hasApplication && lastAgent != "recruiter"):
    return AgentSearch
case keyword == "recruiter" || lastAgent == "recruiter":
    // resume profile check first (existing logic)
    return AgentRecruiter
default:
    return lastAgent
}
```

### Done When
- Recruiter Agent reads CandidateProfile + Application from CheckpointStore ✓
- Multi-turn recruiter conversation works (interview Q&A) ✓
- `record_interview_note` saves structured interview notes ✓
- All builds pass ✓
- Playwright E2E passes ✓

---

## Phase 4 — A2A Recruiter

Candidate Agent ↔ Recruiter Agent

---

## Phase 5 — Negotiation

Salary negotiation, start date negotiation, offer generation

---

## Phase 6 — Memory

Preference extraction, preference recall, long-term user profile

---

## Phase 7 — Autonomous Recruiting

Search → Apply → Chat → Negotiate → Offer → Human Confirmation

---

# Architecture

## Frontend (Next.js 16 App Router)

- `page.tsx` — **server component** (no `"use client"`)
- Client islands in `src/components/`: `AppShell`, `SearchIsland`, `ResumeChatIsland`, `ApplicationCard`, `JobCard`
- Types in `src/lib/types.ts`, hooks in `src/lib/chat.ts`
- Chat uses `@ai-sdk/react` v3 `useChat` with `DefaultChatTransport`
- AI SDK v6 `UIMessage.parts` (not `content`) for rendering

## Backend (Go)

- `cmd/server/main.go` — routes, SSE streaming, CORS
- `internal/agent/` — `Agent` (Eino ReAct agent) + `ResumeAgent`
- `internal/tools/applyjob/` — Firecrawl scrape + LLM generate application
- `internal/tools/queryjobs/` — Indeed China job search
- `internal/tools/parseresume/` — PDF parsing via Firecrawl
- `internal/checkpoint/` — in-memory + file persistence CandidateProfile store
- `internal/job/` — `SearchResponse` with `Message`, `Jobs`, `Application`

---

# Build & Test

```bash
# Backend
cd /home/lora/repos/joblinker
go build ./...
go vet ./...
go test ./...

# Frontend
cd frontend
npx tsc --noEmit
npx next build
npm run dev

# Run full stack
./scripts/start.sh
```

# Every phase: `git checkout main` → new branch → commit → push → PR

# Wheel Ban — Use Libraries, Don't Build Them

## Golden Rule

If a library or SDK already solves a problem, use it. Don't build a custom version.

## Backend

| Instead of building… | Use Eino / Go stdlib |
|---|---|
| Custom agent framework | `github.com/cloudwego/eino` ReAct agent |
| Custom memory/persistence | `eino MemoryStore` interface + `compose.CheckPointStore` |
| Custom SSE streaming | `io.Writer` + `json.NewEncoder` (keep it minimal) |
| Custom tool system | Eino `tool.Tool` interface |

## Frontend

| Instead of building… | Use Vercel AI SDK / shadcn |
|---|---|
| Custom `useChat` hook | `@ai-sdk/react` `useChat()` with `DefaultChatTransport` |
| Custom message parts state | AI SDK v6 `UIMessage.parts` — read don't duplicate |
| Custom chat UI (Conversation, Message, PromptInput) | `ai-elements` from `elements.ai-sdk.dev` |
| Custom markdown renderer | `streamdown` + `@streamdown/*` plugins |
| Custom SSE reader | AI SDK `DefaultChatTransport` / `TextStreamChatTransport` |
| Custom job card state (`structured.jobs`) | `message.parts` `data-jobs` via `useMemo` |

## Enforcement

1. Every new file under `src/components/` or `internal/` must answer: "does a library already do this?"
2. Two 🔴 audit categories:
   - **🔴 State duplication**: data that exists in `UIMessage.parts` must NOT be copied into separate `useState`
   - **🔴 SDK bypass**: `setMessages()` direct manipulation should be avoided — use `sendMessage()` / `append()` instead
3. If a library exists but you choose not to use it, document the reason in a comment

# Forbidden
- No RabbitMQ, Kafka, Workflow Engine, DAG Engine, MCP Infrastructure, Vector Memory until Phase 4
- No hardcoded mock data — always call real APIs
- No `-o` flag with `go build` (use `go build ./cmd/server/` then `mv server output/`)

# Immutable Lessons

**这些教训不可违反，记录原因是当时犯过错误。**

## 1. 系统提示词禁止包含 LLM 可伪造的 JSON 模板

❌ **错误做法**:
```
- 生成申请后附加JSON：
  ===JSON===
  {"application":{"job_title":"...","company":"...","cover_letter":"...","resume_md":"..."}}
  ===END===
```

✅ **正确做法**: 由工具返回真实数据，后端统一提取 JSON。提示词只说明"工具会自动处理"。

**原因**: LLM 看到完整的 JSON 格式会跳过工具调用，直接伪造假数据输出。这是 Phase 2 三天才找到的 root cause。

## 2. 不依赖 LLM 提取 `session_id` 等机器生成 ID

❌ **错误做法**: 把 `session_id` 放在系统提示词里让 LLM 读取并传给工具参数。

✅ **正确做法**: 通过 Go `context.Context` 传递 session_id，工具从 `ctx` 中读取。

**原因**: LLM 对 `s170405c1` 这种简短机器 ID 的提取和传递不可靠 → 传空/传错 → 工具拿到 `:profile`（空 session_id）→ 永远找不到 profile。Context 方式 100% 可靠。

## 3. context key type 必须跨包可访问

❌ **错误做法**: 在 `agent` 包内定义 `type sessionIDKey struct{}`（unexported），其他包无法读取。

✅ **正确做法**: 共享 key 放在 `internal/session/context.go`，导出 `SessionIDFromContext()` / `WithSessionID()`。

**原因**: Eino 工具在 `applyjob` 包中执行，无法访问 `agent` 包的私有 key → context 有值但工具拿不到。

## 4. `useChat` + `useRef` 不随 prop 更新

❌ **错误做法**: `const chat = useRef(new Chat(...))` — transport 在首次渲染后不变，`sessionId` 变化无效。

✅ **正确做法**: `useChat({ id: sessionId ?? "no-session" })` — `id` 变化时 Chat 实例重建。

**原因**: sessionId 从 `null` → 上传后的值，但 Chat 实例的 transport 仍是初始化时的 `undefined` → 简历 chat 发到错误的 `/api/chat` 端点（搜索 agent），而不是 `/api/chat/resume`。

## 5. 禁止提示词中包含"让 LLM 自己做"的指令

❌ **错误做法**: 提示词说"生成申请后附加JSON"，暗示 LLM 自己构造数据。

✅ **正确做法**: 工具返回值是唯一数据来源。提示词只描述流程，不描述输出格式细节。

**原因**: LLM 是"过分配合"的 — 你给它一个模板，它就填模板，而不是调用工具。所有结构化数据必须来自工具/API 返回值。
