# CLAUDE.md

Goal: Reduce human involvement in recruiting via autonomous AI agents.

---

## Phase 1 — Job Search Agent ✅
`queryjobs` — search + rank + explain jobs

## Phase 2 — Auto Apply Agent ✅
`queryjobs` + `applyjob` + `parseresume`

Components: `AppShell` | `UnifiedChatIsland` | `ApplicationCard` | `JobCard`

Endpoints:
- `POST /api/chat` — AI SDK v6 SSE search agent
- `POST /api/apply` — JSON generate application
- `POST /api/resume/upload` — FormData PDF upload
- `GET /health`

Protocol: `start` → `text-start` → `text-delta*` → `text-end` → `data-*` → `finish`
Data events: `data-jobs` (array), `data-application` (object)

## Phase 3 — Recruiter Chat
https://www.a2a-registry.org/about/a2a
https://a2a-protocol.org/latest/specification/#8-agent-discovery-the-agent-card
https://modelcontextprotocol.io/introduction
https://github.com/a2aproject/A2A
https://adk.dev/a2a/intro/
Agent Card → capability negotiation → A2A communication

## Phase 4-7
A2A Recruiter → Negotiation → Memory → Autonomous (Search→Apply→Chat→Negotiate→Offer→Confirm)

---

## Architecture

### Frontend (Next.js 16 App Router)
- `page.tsx` — server component
- Client islands: `AppShell`, `UnifiedChatIsland`, `ApplicationCard`, `JobCard`
- Types: `src/lib/types.ts`, helpers: `src/lib/ai-utils.ts`
- Chat: `@ai-sdk/react` `useChat` + `DefaultChatTransport`
- Messages: AI SDK v6 `UIMessage.parts`

### Backend (Go)
`cmd/server/main.go` — routes, SSE, CORS
`internal/agent/` — Eino ReAct Agent + ResumeAgent
`internal/tools/applyjob/` — Firecrawl + LLM generate application
`internal/tools/queryjobs/` — Indeed China search
`internal/tools/parseresume/` — Firecrawl PDF parse
`internal/checkpoint/` — CandidateProfile store

---

## Build
```bash
cd /home/lora/repos/joblinker && go build ./... && go vet ./... && go test ./...
cd frontend && npx tsc --noEmit && npx next build
```

---

## Wheel Ban — Use Libraries, Don't Build

Golden Rule: If lib exists, use it. No custom versions.

### Backend
| Don't build | Use |
|---|---|
| Agent framework | `github.com/cloudwego/eino` ReAct |
| Memory/persistence | `eino MemoryStore` + `compose.CheckPointStore` |
| SSE streaming | `io.Writer` + `json.NewEncoder` |
| Tool system | Eino `tool.Tool` |

### Frontend
| Don't build | Use |
|---|---|
| Custom `useChat` | `@ai-sdk/react` `useChat()` + `DefaultChatTransport` |
| Chat UI (Conversation/Message/Input) | **`ai-elements` from `elements.ai-sdk.dev`** — 禁止手写 |
| Tool call card | `ai-elements` `<Tool>` `<ToolHeader>` `<ToolContent>` `<ToolInput>` `<ToolOutput>` |
| Reasoning panel | `ai-elements` `<Reasoning>` `<ReasoningTrigger>` `<ReasoningContent>` |
| Markdown renderer | `streamdown` + `@streamdown/*` |
| SSE reader | AI SDK `DefaultChatTransport` |
| Job card state | `message.parts` `data-jobs` via `useMemo` |

**AI Elements 铁律**: 所有 `src/components/ai-elements/` 下文件必须来自 `npx ai-elements@latest add`。禁止手写。需要新组件用 CLI 安装。

### Enforcement
1. Every new file under `src/components/` or `internal/`: "does lib already do this?"
2. 🔴 State duplication: data in `UIMessage.parts` must NOT copy to separate `useState`
3. 🔴 SDK bypass: avoid `setMessages()` — use `sendMessage()` / `append()` instead

### Forbidden
- ❌ RabbitMQ, Kafka, Workflow Engine, DAG Engine, MCP, Vector Memory (until Phase 4)
- ❌ Hardcoded mock data — always call real APIs
- ❌ `-o` flag with `go build`
- 🔴 **Custom chat UI components** — must use `ai-elements` CLI
- 🔴 **`as any` on `UIMessage.parts`** — use `isDataUIPart`, `isTextUIPart`, `isReasoningUIPart` type guards
- 🔴 **`setTimeout` for React state sync** — use `useRef` + `useEffect`
- 🔴 **Error-as-success string** — Go tool errors must `return "", fmt.Errorf(...)`

---

# Immutable Lessons

**不可违反。每条都是真金白银踩出来的坑。**

## 1. 系统提示词禁止包含 LLM 可伪造的 JSON 模板
❌ 提示词给 JSON 模板 → LLM 跳过工具调用伪造假数据。
✅ 工具返回值是唯一数据来源。提示词只描述流程不描述格式。
Root cause: Phase 2 花三天才找到。

## 2. 不依赖 LLM 提取 `session_id`
❌ 把 `session_id` 放提示词让 LLM 传给工具。
✅ Go `context.Context` 传递，工具从 `ctx` 读。
原因: LLM 对短机器 ID 提取不可靠 → 传空 → 找不到 profile。

## 3. Context key type 必须跨包导出
❌ `agent` 包内定义 unexported `type sessionIDKey struct{}`。
✅ 放 `internal/session/context.go`，导出 `SessionIDFromContext()` / `WithSessionID()`。
原因: Eino 工具在其他包执行，无法访问私有 key。

## 4. `useChat` + `useRef` 不随 prop 更新
❌ `useRef(new Chat(...))` — transport 首次渲染后不变。
✅ `useChat({ id: sessionId ?? "no-session" })` — `id` 变化重建实例。
原因: sessionId 变但 transport 仍是老的 → 请求发错端点。

## 5. 禁止提示词"让 LLM 自己做"
❌ "生成申请后附加JSON" → LLM 自己构造数据不调工具。
✅ 提示词只描述流程。所有结构化数据必须来自工具/API 返回值。

## 6. 工具调用必须对用户可见
❌ Agent 调 `query_jobs` / `apply_job`，UI 卡死 30 秒无反馈。
✅ 渲染 `UIMessage.parts` 中 `tool-*` 为 `<Tool>` 组件显示状态。
原因: 无可见性 = 用户以为系统卡死。Agent 产品 vs 聊天机器人的本质区别。

## 7. 禁止 `(part as any)`
❌ `(part as any).data` — 后端改 data 格式零编译保护。
✅
```typescript
if (isDataUIPart(part) && part.type === 'data-jobs') {
  return part.data; // RankedJob[]
}
function isToolPart(part: UIMessagePart): part is Extract<UIMessagePart, { type: `tool-${string}` }> {
  return part.type.startsWith('tool-');
}
```
原因: Phase 2 有 6 处 `as any`，后端改格式全部静默断掉。

## 8. 禁止 `setTimeout` 同步 React 状态
❌ `await new Promise(r => setTimeout(r, 50))` 等状态更新 — 竞态条件。
✅
```typescript
const pendingRef = useRef<string | null>(null);
setSessionId(sid); pendingRef.current = text;
useEffect(() => {
  if (sessionId && pendingRef.current) {
    sendMessage({ text: pendingRef.current }, { body: { session_id: sessionId } });
    pendingRef.current = null;
  }
}, [sessionId, sendMessage]);
```

## 9. Go Error 必须作为 error 返回
❌ `return fmt.Sprintf("失败: %v", err), nil` — LLM 把错误当有效数据解析。
✅ `return "", fmt.Errorf("失败: %w", err)` — Eino ReAct 知道工具调用失败。

## 10. HTTP handler 必须设 context deadline
❌ 直接用 `r.Context()` — LLM/Firecrawl 挂死 handler 也挂死。
✅ `ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)`。

## 11. Session 资源必须清理
❌ `locks.Get(sessionID)` 永远不 Remove；checkpoint.Set() 永远不 Delete。
✅ 加 `defer locks.Remove(sessionID)`；cleanupLoop 清理检查点。
原因: 每次 upload 产生新 session，永久占 map 内存。长期运行 OOM。
