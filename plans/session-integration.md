# Plan: Session 集成 — 接入主线程

> Source PRD: `joblinker/docs/prd-session-memory.md`
> Issues: #7 (S1), #8 (S2), #9 (S6), #10 (S3), #11 (S4), #12 (S5)

## 架构决策

| 维度 | 决策 | 理由 |
|------|------|------|
| **Session ID 格式** | `{matchUUID}@v{version}` | PRD Decision 1 |
| **存储后端** | JSONL（开发） / PG（生产），通过 `Store` 接口切换 | PRD Decision 4 |
| **后端状态管理** | 不用手写状态机。用 Eino `agent.Chat(ctx, history, msg)` 替代 `DeepRecruiter` | S2 要求 |
| **消息持久化** | `model.Message` 旧表 + SessionStore 双写入过渡 | 向后兼容 |
| **前端消息状态** | `useChat` from `@ai-sdk/react` | 禁止手写 useState 管消息 |
| **前端 session 元数据** | `useChat.data` 字段 | 禁止手写 useState 管 fsmStage/pendingConfirm/sessionVersion |
| **传输层** | REST 加载初始消息 + WebSocket 实时推送 | `useWebSocket` hook 保留（传输层，非状态管理） |
| **编排层** | `SessionService` 保留（业务编排，非手写状态机） | PRD Decision 5 |

## 禁止手写清单（红线）

| 内容 | 状态 | 目标 |
|------|:----:|------|
| `DeepRecruiter` 状态机 + `sync.RWMutex` | ❌ 手写 | Phase 3a 消除 |
| `AgentRunner.deepPool` | ❌ 手写池 | Phase 3a 消除 |
| `memory.AgentMemory` | ❌ 手写内存 | Phase 5+6 删除 |
| `useState` 管 `fsmStage` / `pendingConfirm` | ❌ 手写 | Phase 5+6 收进 `useChat.data` |
| `useChat` 管 `messages` / `input` | ✅ 框架 | 保留 |
| `Store` 接口 + 实现 | ✅ 数据抽象层 | 保留 |
| `SessionService` | ✅ 业务编排层 | 保留 |

---

## Phase 1: Session Store 可读 — 修 bug + 前端切数据源

**用户故事**: US1（刷新不丢）、US3（查询历史）

### 后端
1. 修 7 编译 bug:
   - `mqsvc:100`: `SetSessionStore(store *sessionstore.Store)` → `sessionstore.Store`
   - `mqsvc:112`: `sessionStore.CreateSession(matchID)` 改用 `LatestVersion→SessionID→Create`
   - `mqsvc:1042`: `AppendMessage(sender, msg, phase)` → `AppendMessages(ctx, sid, []Message{...})`
   - `mqsvc:1076`: `ConcludeSession(sid)` → `Conclude(ctx, sid, reason)`
   - `handler:86,109,148`: `ListByMatchID(ctx, matchID)` matchID 加 `uuid.Parse`
   - `handler:123,166`: `s.ID` → `s.SessionID`
   - `handler:133`: `GetMessages()` → `Load(ctx, sid).Messages`
2. `go build ./...` 通过

### 前端
1. `useAIChat` 初始加载从 `/api/sessions/:matchId/messages` 取数据
2. `ConversationHeader` 显示 `session_version` 标签

### 可验证
浏览器打开 conversation 页 → 消息正常显示，右上角有 "v1" 标签

---

## Phase 2: Session Store 可写 — StateGraph 路径

**用户故事**: US1、US9（concluded 拒绝写入）

### 后端
1. `runStateGraphForMatch` 正确调用 `Store.Create → AppendMessages → Conclude`
2. 新增 WebSocket 事件：`session_created`、`session_concluded`
3. 修复 `handler:133` 的 `GetSessionMessages` 返回正确数据

### 可验证
触发 StateGraph → 等待完成 → `GET /sessions/:matchId` 返回 version=1, status=concluded

---

## Phase 3a: 消除 DeepRecruiter（手写状态机）

**用户故事**: US4（Agent stateless）、US5（显式流程）

### 后端
1. `generateEinoResponse` 不再调 `DeepRecruiter.Process*Message`（传 nil history）
2. 改为从 SessionStore Load history → `agent.Chat(ctx, history, msg)`
3. 删除 `AgentRunner` 的 `deepPool` 和 `GetDeepRecruiter` / `ReleaseDeepRecruiter`
4. 删除或标记 `DeepRecruiter` struct 为 deprecated

### 可验证
Eino 路径的 AI 回复基于对话历史生成（不再零记忆）

---

## Phase 3b: 正常消息流写 SessionStore

**用户故事**: US1、US9

### 后端
1. 新增 `SessionService.RecordTurn(ctx, matchID, role, content) error`:
   - 内部自动：检查 session 是否存在 → 不存在则 `LatestVersion→SessionID→Create`
   - `AppendMessages` → `maybeCompress`
   - handleAgentMessage 零感知 session 细节
2. `handleAgentMessage` ADK/Eino/Legacy 流在 `messageRepo.Create()` 后调 `RecordTurn`
3. `runStateGraphForMatch` 的 onMessage 也用 `RecordTurn`
4. WebSocket 新增：`session_created`（首次消息时发出）

### 可验证
发消息 → 等回复 → `GET /sessions/:matchId/messages` 能看到双向消息

---

## Phase 4: Reopen + 压缩摘要

**用户故事**: US2（重开）、US6（自动压缩）

### 后端
1. `POST /sessions/:matchId/reopen` — PRD Decision 8：
   - 验证 session 是 `concluded_paused`
   - 创建 v2，注入 v1 摘要到 system prompt
   - 返回新 sessionID
2. `GET /sessions/:matchId/summary` — 用类型断言替代反射（JSONL 兼容）
3. `GET /conversation/:matchId` 增强：返回 `session_version`、`session_status`、`session_id`
4. WebSocket 新增：`session_reopened`

### 可验证
手动 Conclude → Reopen → GET 看到 v1(concluded) + v2(active)

---

## Phase 5+6: 前端 Session UI + main.go + 清理

**用户故事**: 全部

### 后端 — main.go 连线
1. 将 `seekerAgent`/`recruiterAgent` 从 ADK 块提取到外层
2. `sessionstore.NewSessionService(store, seeker, recruiter)`（ADK 失败时用 fallback）
3. `mqSvc.SetSessionService(sessionSvc)` 替代 `SetSessionStore`
4. `messageHandler.SetSessionService(sessionSvc)` 替代裸 Store

### 前端 — 收编手写状态
1. `useChat.data` 承载所有 session 元数据：
   - `fsmStage` / `sessionVersion` / `sessionStatus` / `sessionSummary` / `pendingConfirm`
   - 不再有独立的 `useState` 管这些
2. 组件：
   - `SessionVersionBadge` — ConversationHeader 显示 "v1"/"v2"
   - `ReopenButton` — sessionStatus === 'paused' 时显示
   - `SummaryPanel` — 折叠面板展示压缩摘要
3. WebSocket 事件处理（`onMessage` 内）：
   - `session_created` → `setData({ sessionVersion, sessionStatus: 'active' })`
   - `session_concluded` → `setData({ sessionStatus: 'concluded' })`
   - `session_reopened` → 清消息 + `setData({ sessionVersion, sessionStatus: 'active' })`

### 清理
1. 删除 `memory.AgentMemory`（如无人引用）
2. 标记 `DeepRecruiter` deprecated
3. 全量 `go build ./...` + `npm run build`

### 可验证
浏览器看到版本号、paused 状态有 reopen 按钮、摘要面板、所有 Phase 1-4 功能正常
