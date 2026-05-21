# PRD: Session 管理与记忆持久化

## Problem Statement

目前的 JobLinker A2A 对话系统存在以下问题：

1. **纯内存运行**：`SeekerAgent` 和 `RecruiterAgent` 各自持有 `AgentMemory`，所有 `[]*schema.Message` 只存在于内存中，进程重启后全部丢失。
2. **状态与存储耦合**：`DeepRecruiter` 既管状态转换、又持 memory 引用、又管消息路由——三个职责混在一起，难以独立测试。
3. **无 Session 概念**：match 只能进行一次对话，没有「重开」「恢复」「版本管理」能力。matchID 同时承担了业务实体和运行时上下文的角色。
4. **无 Checkpoint 持久化**：Eino `CheckPointStore` 接口有 Redis 实现，但 Redis 有 TTL 限制且非项目共识方案；StateGraph 的真正价值（state snapshots）未被利用。
5. **消息膨胀无压缩**：对话可以无限进行，没有 token 量触发压缩的机制，迟早会撑爆 context window。

## Solution

引入三层架构的 Session & Memory 系统：

```
业务层 (Business Layer)      框架层 (Framework Layer)       存储层 (Storage Layer)
┌──────────────┐          ┌─────────────────┐          ┌─────────────────────┐
│ SessionService │─────▶│  StateGraph.Run  │─────▶│  PostgreSQL          │
│ - CreateSession │      │  - Seeker Node  │      │  ├─ session_messages  │
│ - AppendMessages│      │  - Recruiter Node│      │  ├─ session_checkpoints│
│ - Load          │      │  - Shared State  │      │  └─ session_meta      │
│ - Compress      │      │  - CheckPointStore │      │                      │
│ - List/Delete   │      └─────────────────┘      │  JSONL (dev only)     │
└──────────────┘                                  └─────────────────────┘
```

转换目标：
- Agent 不再持有 `*memory.AgentMemory`，只收 `[]*schema.Message`
- `DeepRecruiter` 简化为 stateless 状态机，state 由 SessionService 管理
- 引入 `SessionID = matchID + "@v" + version`，支持同一个 match 多次 session
- 每次 turn 后，业务层将新消息追加到 PostgreSQL
- Token 计数超过阈值时，业务层触发压缩并保存 summary 到 checkpoint
- CheckPointStore 后端切换到 PostgreSQL，存 StateGraph 快照（MatchState + Messages + Summary）

## User Stories

1. 作为一个求职者，我希望我的对话在页面刷新后不会丢失，以便我能继续上次的谈判。
2. 作为一个招聘方，我希望 match 结束后可以「重开」新 session，以便在变化的条件基础上重新谈判。
3. 作为一个系统管理员，我希望能够查询某个 match 的所有 session 历史，以便审计整个谈判过程。
4. 作为一个开发者，我希望 Agent 的输入输出只有 `[]*schema.Message`，以便我能独立测试 Agent 而不依赖存储层。
5. 作为一个开发者，我希望 SessionService 使用 `Load → RunTurn → AppendMessages` 的显式流程，以便我能轻松调试每一步。
6. 作为一个系统管理员，我希望对话压缩可以自动触发（基于 token 计数），以便控制存储和 LLM 成本。
7. 作为一个开发者，我希望 CheckPointStore 使用 PostgreSQL 而不是 Redis，以便与项目已有基础设施一致。
8. 作为一个开发者，我希望在开发环境中可以使用 JSONL 文件来替代 PostgreSQL 存储对话，以便快速原型而不依赖数据库。
9. 作为一个求职者，当一个 session 被标记为「concluded」后，我希望系统拒绝写入新消息，以便保证数据完整性。
10. 作为一个开发者，我希望 Session 包含版本号，以便跟踪同一个 match 的多次对话生命周期。

## Implementation Decisions

### Decision 1: Session 身份边界

```
SessionID = matchID.String() + "@v" + strconv.Itoa(sessionVersion)
```

- `matchID` = 业务实体（谁和谁 match），不变
- `sessionVersion` 从 1 开始，每次「重开」递增
- CheckPointStore 的 KV key：`"session:{sessionID}:{checkPointID}"`

### Decision 2: 单 StateGraph 共享

整个 A2A 对话使用 **一个** StateGraph，内部有 Seeker 和 Recruiter 两个节点，共享一个 `RecruitmentState`。

不拆成双 Graph——保持与现有 `DeepRecruiter` coordinator 模式一致，降低迁移成本。

### Decision 3: Checkpoint 存储内容

Checkpoint blob 序列化为 JSON，包含：

```json
{
  "sessionID": "abc-def@v1",
  "matchID": "abc-def",
  "version": 1,
  "state": "seeker_turn",
  "messageCount": 42,
  "compressedRounds": 3,
  "summary": "Summary text...",
  "createdAt": "...",
  "updatedAt": "..."
}
```

- 不存全量 Messages 到 checkpoint——Messages 存在 `session_messages` 表
- Checkpoint 只存 **状态快照 + 游标信息**（消息总条数）+ 摘要
- 每次 checkpoint 前触发一次压缩瘦身

### Decision 4: 存储后端首选 PostgreSQL

| 用途 | 表名 | 说明 |
|------|------|------|
| 消息存储 | `session_messages` | 每条消息一行，JSONB 存储 role/content/sender/seq |
| Session 元数据 | `session_meta` | sessionID, matchID, version, state, createdAt |
| Checkpoint | `session_checkpoints` | sessionID, checkPointID, blob (JSON), createdAt |
| 摘要 | `session_summaries` | sessionID, round, summary text, token_count |

开发环境可选 JSONL 实现（每条消息一行 JSON，追加写入）。

### Decision 5: 三层职责分离

```
业务层 (SessionService):
  1. sessionStore.Create(ctx, sessionID, matchID)
  2. sessionStore.Load(ctx, sessionID) → Session{Messages, State}
  3. runner.RunTurn(ctx, messages, sender, content) → (reply, newState)
  4. sessionStore.AppendMessages(ctx, sessionID, [userMsg, aiReply])
  5. if tokenCount > threshold { compress(ctx, sessionID) }

框架层 (StateGraph):
  1. 接收 []*schema.Message + 新输入
  2. 内部路由给 Seeker/Recruiter Agent
  3. 返回 reply + 新 state

存储层 (SessionStore):
  1. interface: Create/Load/Append/List/Delete
  2. implementation: PostgreSQL OR JSONL
```

### Decision 6: Agent-Memory 解耦

当前 `Agent` 持有 `*memory.AgentMemory` → 改为只接收 `[]*schema.Message`。

变更点：
- `SeekerAgent.ChatWithMemory(ctx, msg)` → `SeekerAgent.Chat(ctx, history, msg)`
- `RecruiterAgent.ChatWithMemory(ctx, msg)` → `RecruiterAgent.Chat(ctx, history, msg)`
- 删除 `AddMemoryMessage` / `GetMemoryMessages` 方法
- `memory.AgentMemory` 可退役（由 `SessionStore` 替代）

### Decision 7: 自动压缩

业务层在每次 `AppendMessages` 后检查 token 计数：

- `estimateTokenCount(messages) > 4000` 时触发压缩
- 压缩后生成 `ConversationSummary`，存入 `session_summaries`
- 压缩后的首轮消息可以用 summary 替换（或者保留全量但标记为已压缩）

### Decision 8: Match 版本升级策略

#### Conclude 有子状态

Conclude 有两种语义：

| 子状态 | 含义 | 可 Reopen? |
|--------|------|:---------:|
| `concluded_naturally` | 双方达成协议/offer 确认/hired/rejected | ❌ |
| `concluded_paused` | 用户主动结束/超时/max rounds | ✅ |

自然终局不可 reopen；暂停/超时/手动结束的可 reopen。

`session_meta` 增加 `conclusion_reason` 列。

#### Reopen 的数据继承

Reopen 只继承 **Summary + State**，**不复制 Messages**。

```
v1 (concluded_paused)：
  ├─ Messages:    45 条 (全量存 PG)
  ├─ State:       { stage: "negotiating", salary_agreed: 190000 }
  └─ Summary:     "双方确认基本薪资 190K...\n待定事项：入职日期\n..."

Reopen → v2 (active)：
  ├─ Messages:    []  ← 空，新对话
  └─ SystemPrompt: (自动注入，包含 v1 Summary + State)
```

Summary 必须结构化（四段格式）：
1. **已达成项**：key-value 对（salary=190K, bonus=30K, ...）
2. **待定事项**：未解决的议题列表
3. **放弃项**：讨论后被否决的选项
4. **对话概览**：一句话概述 what happened

Conclude 时使用专门的 summary prompt 保证摘要质量，而非简单拼接。

#### Match 状态自动更新

Session 推进 → 自动更新 `Match.Status`：

| Session 推进 | 原 Match 状态 | 新 Match 状态 | 触发时机 |
|-------------|:-----------:|:-----------:|---------|
| Reopen 后首次轮换 | negotiating | (不变) | 暂无 |
| 完成薪资谈判 | negotiating | interviewing | recruiter_turn 切换时检测 |
| 排好面试 | interviewing | offered | interview 记录创建 |
| offer 确认 | offered | hired | offer accepted |
| 用户放弃 | any | rejected | 显式放弃 |
| 24h 无回复 | any | paused | 定时 Job |

自动更新规则代码化在 `SessionService`，不在 Agent 中。

### Decision 9: Session 过期与清理（延后实现）

- Session 过期时间判定：暂不自动清理，先支持手动 `Delete`
- 后续可以加 TTL 列 + 定时清理 job
- 会话搜索 / 导出 / 导入：暂不做，列入 Out of Scope

## Testing Decisions

### 测试哲学

- 只测外部行为，不测实现细节
- 存储层：用接口 mock 测试业务层逻辑
- 框架层：用 mock AI server 测试 graph 节点行为（已有先例 `graph_test.go`）

### 覆盖模块

| 模块 | 测试策略 | 先例 |
|------|---------|------|
| `SessionStore (PG)` | 集成测试（用 testcontainers 或 in-memory PG） | 项目已有 repository 测试模式 |
| `SessionStore (JSONL)` | 单元测试（临时目录 + 读写校验） | 轻量，无需 DB |
| `SessionService` | 单元测试（mock SessionStore + mock Runner） | 标准 Go mock 模式 |
| `DeepRecruiter` (stateless) | 纯逻辑测试（给定 messages → 预期 state 转换） | `TestParseMessage` 等纯函数测试 |
| Checkpoint 序列化 | 单元测试（JSON marshal/unmarshal 往返） | 标准 Go 测试 |
| 压缩触发 | 单元测试（给 10 条长消息 → 确认调用 compress） | 项目无先例，需新写 |

## Out of Scope

- 会话搜索（全文检索 messages）
- 会话导出 / 导入（JSON export）
- 会话分享（跨 match 引用）
- 自动清理 / 定时删除过期 session
- 向量嵌入和 RAG 检索（现有 `vector_store.go` 已覆盖）
- UI 层面的 session 管理界面
- StateGraph 本身的 checkpoint 恢复机制（Eino 官方 `WithCheckPointStore` 的集成，在 Phase 2）

## Further Notes

- 本 PRD 与已有的 `stategraph` (Phase 1) 和 `DeepRecruiter` 共存：StateGraph 用于全自动 A2A 招聘流程 `stategraph/graph.go`（4 阶段），DeepRecruiter 用于用户交互式 A2A 对话 `agent/deep_recruiter.go`。两路径都需要 Session 系统。
- CheckPointStore 的 PG 实现需要兼容 Eino 的 `compose.CheckPointStore` 接口（`Get/Set` 方法签名）。
- `message.go` 的 `model.Message` 与 `schema.Message` 是两种结构。Session 存储层存 `schema.Message`（`session_messages` 表 JSONB），业务层存 `model.Message`（PG 关系表）。需要明确映射关系。
