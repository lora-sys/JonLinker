# JonLinker 大重构设计文档 v2

> 日期: 2026-05-24
> 仓库: https://github.com/lora-sys/JonLinker
> 定位: Agent-driven A2A 招聘平台，开源项目
> 当前状态: 功能已大量实现，但 42 个具体 bug（1 P0 + 21 P1 + 20 P2），代码杂而无序、拼凑感强、尘余代码多

---

## 1. 问题诊断

### 1.1 根因

不是"功能不够"，是三层问题叠加：

```
问题层 3: 架构混乱 — service/ 15+ 文件职责不清，agent/ 与 service/ 循环依赖
问题层 2: 尘余代码多 — Legacy 降级链、空 stub、未引用 model、重复逻辑
问题层 1: 边界 bug — P0 除零 panic、FSM 死锁、消息静默丢失、并发无保护
```

### 1.2 重构原则

| 原则 | 说明 |
|------|------|
| 不重写业务逻辑 | 只移动、重组、修 bug，核心算法不动 |
| 先删后移 | 先清尘余，再迁移，最后修 bug |
| 渐进验证 | 每步执行后 `go build` 确保不引入编译错误 |
| 保持克制 | 不新增功能，不动 DB schema，不引入新依赖 |

---

## 2. 重构目标

| 优先级 | 目标 | 衡量标准 |
|--------|------|----------|
| P0 | 删除所有尘余代码 | Legacy 路径、空 stub、未引用 model、无用文件全部删除 |
| P0 | 消除所有 P0/P1 bug | 除零 panic、FSM 死锁、消息丢失全部修复 |
| P1 | 架构重组为 core/engine/adapters/transport | 每个包职责单一，依赖方向清晰 |
| P1 | Eino 成为唯一对话编排入口 | 删除 ADK/Legacy fallback 降级链 |
| P2 | 前端按 feature 组织 | 组件/hooks/stores 归入对应功能模块 |
| P2 | 开源友好 | 清晰目录，新人 10 分钟看懂架构 |

---

## 3. Phase 0: 精确诊断（必须先做）

> 目标: 精确掌握尘余代码分布，不是凭猜测删

### 3.1 诊断任务清单

| # | 诊断项 | 方法 | 产出 |
|---|--------|------|------|
| 1 | 统计所有 .go 文件引用关系 | `go vet` + `staticcheck` 检测未使用代码 | 未引用/未使用文件列表 |
| 2 | 检查 Eino 真实集成程度 | 搜索 `eino` 全仓引用，查看 `eino/` 目录所有文件内容 | Eino 使用程度报告 |
| 3 | 识别 Legacy 路径 | 搜索 `legacy`、`fallback`、`mock`、`generateEinoResponse` | Legacy 代码位置清单 |
| 4 | 检查前端组件引用 | 搜索 `import.*from.*components` 统计每个组件的引用次数 | 未使用组件列表 |
| 5 | 检查测试文件 | `find . -name "*_test.go" -o -name "*.test.*" -o -name "*.spec.*"` | 测试覆盖率 |
| 6 | 检查 model 引用 | 对 `model/` 下每个文件做 `grep -r "类型名" backend/` | 每个 model 的引用次数 |
| 7 | 检查 repository 引用 | 对 `repository/` 下每个文件做引用统计 | 每个 repo 的引用次数 |
| 8 | 检查 pkg/ 子目录引用 | 对 `pkg/` 下每个子包做引用统计 | pkg 使用情况 |

### 3.2 诊断后的决策点

诊断完成后，根据实际数据确认：

- **删除清单**: 哪些确实可以删（引用次数 = 0）
- **迁移清单**: 哪些需要移动但有大量引用（需同步更新 import）
- **保留清单**: 哪些虽看起来冗余但实际在用

---

## 4. Phase 1: 删除尘余代码

> 目标: 安全删除确定不用的代码，为架构重组扫清障碍
> 前提: Phase 0 诊断完成后执行

### 4.1 删除策略（保守优先）

```
第一轮: 删确定无引用的（引用次数 = 0）
  - 未引用 model 文件
  - 未引用 repository 文件
  - 未使用前端组件/hooks/stores
  - 空文件（只有 package 声明）

第二轮: 删死代码（有引用但逻辑为空）
  - generateEinoResponse 返回 nil 的 stub
  - 空函数体
  - 未被调用的常量/变量

第三轮: 删 Legacy 路径（确认 Eino 已替代后）
  - ADK runner 调用
  - Legacy fallback 降级链
  - 旧的 AI 编排逻辑
```

### 4.2 每轮删除后验证

```bash
cd backend && go build ./...
cd frontend && npm run build
```

编译不过 → 回滚 → 检查遗漏的引用 → 修复后再删。

---

## 5. Phase 2: Backend 架构重组

> 目标: 将保留的代码迁移到新架构，不改业务逻辑
> 原则: 文件移动 + 重命名 + import 更新，不动核心算法

### 5.1 新架构

```
backend/
├── cmd/server/main.go              # 入口，组装依赖 (DI)
├── internal/
│   ├── core/                       # 纯领域逻辑，零外部依赖
│   │   ├── fsm.go                  # FSM 状态机（从 agent/fsm.go 提取，修复 CanHandle 不一致）
│   │   ├── negotiation.go          # 薪资协商算法（从 agent/negotiator.go 提取，修复除零 panic）
│   │   └── models.go              # 核心数据模型（从 model/ 精简后保留有用的）
│   │
│   ├── engine/                     # Eino 编排引擎 — 唯一对话入口
│   │   ├── graph.go               # Eino Graph 定义
│   │   ├── runner.go              # Graph 执行器（RabbitMQ 驱动）
│   │   └── nodes/
│   │       ├── fsm_node.go        # 状态转换 + 并发保护
│   │       ├── ai_node.go         # AI 调用（从 agent/decision.go + pkg/ai 合并）
│   │       ├── tool_node.go       # 工具执行 + function_definitions（从 agent/tool_executor 迁移）
│   │       ├── memory_node.go     # 向量记忆 + 偏好提取（合并 agent_memory + preference_extraction）
│   │       ├── confirm_node.go    # 人类确认（合并 confirmation + interview）
│   │       └── negotiation_node.go # 薪资协商节点（调用 core/negotiation）
│   │
│   ├── adapters/                   # 外部依赖适配器 — 薄封装层
│   │   ├── ai_client.go           # LongCat AI 客户端（从 pkg/ai 精简）
│   │   ├── rabbitmq.go            # RabbitMQ 客户端（从 pkg/rabbitmq 精简，加重试逻辑）
│   │   ├── postgres.go            # 数据库 + Repository（从 repository/ 合并）
│   │   ├── embedding.go           # Jina Embedding
│   │   └── parser.go              # 简历解析器（从 pkg/parser）
│   │
│   ├── transport/                  # 传输层 — 仅做验证 + 转发
│   │   ├── rest/
│   │   │   ├── agent_handler.go   # Agent CRUD（薄层，调用 adapters）
│   │   │   ├── match_handler.go   # Match 管理
│   │   │   └── message_handler.go # 消息入口（触发 engine/runner）
│   │   └── ws/
│   │       └── websocket_handler.go # WebSocket（加重连补推）
│   │
│   └── middleware/                 # JWT, CORS, 限流
│
├── migrations/                     # 数据库迁移（保留不变）
├── tests/
│   ├── unit/
│   └── integration/
├── go.mod / go.sum
└── Dockerfile
```

### 5.2 依赖方向（不可违反）

```text
transport → engine → core
transport → adapters
engine → core
engine → adapters

禁止: core → adapters（core 零外部依赖）
禁止: engine → transport
禁止: adapters → engine
```

### 5.3 迁移步骤

```
Step 1: 创建新目录结构
  mkdir -p internal/{core,engine/nodes,adapters,transport/{rest,ws}}

Step 2: 提取 core（无外部依赖的纯逻辑）
  从 agent/fsm.go → core/fsm.go（修复 CanHandle/Handle 不一致、StatePaused 转换表）
  从 agent/negotiator.go → core/negotiation.go（修复除零，删 SimulateNegotiation）
  从 model/user.go → core/models.go（精简，只保留实际使用的模型）

Step 3: 构建 adapters（外部依赖封装）
  从 pkg/ai/client.go → adapters/ai_client.go（修复 mockRound 并发、流式错误信号）
  从 pkg/rabbitmq/ → adapters/rabbitmq.go（添加发布重试）
  从 repository/* → adapters/postgres.go（合并数据访问）
  从 pkg/parser/ → adapters/parser.go

Step 4: 构建 engine nodes（Eino Graph 节点）
  合并 agent/decision.go + ai_client → nodes/ai_node.go
  合并 agent/tool_executor + function_definitions → nodes/tool_node.go
  合并 agent_memory_service + preference_extraction → nodes/memory_node.go
  合并 confirmation_service + interview_service → nodes/confirm_node.go
  新建 engine/graph.go + runner.go

Step 5: 重写 transport（薄层 handlers）
  从 handler/* → transport/rest/ + transport/ws/
  修复: WebSocket 重连补推、processMessage 错误传播、认证方式

Step 6: 组装 main.go
  删除旧 import，注入新依赖

Step 7: 删除旧目录
  rm -rf internal/{service,agent,handler,model,repository,cache,config}
  精简 pkg/（只保留 shared/）
```

### 5.4 边界条件处理（在迁移时一并修复）

| 场景 | 处理方式 |
|------|----------|
| negotiator min == target | 返回 early，不进入除法 |
| FSM EventResume 被拦截 | 合并 CanHandle+Handle |
| FSM StatePaused 无出口 | 补充转换表 |
| 消息发布失败 | 指数退避重试 3 次，失败写 DLQ |
| WebSocket 断线重连 | 支持 `?since=<timestamp>` 补推 |
| 工具结果为空 vs 查询失败 | 返回不同 status |
| confirm 超时 24h | 自动 reject + 通知 |
| runningGraphs 泄漏 | defer delete |
| AI 流式响应半截断开 | 使用带 error 的结果 channel |

---

## 6. Phase 3: Frontend 重组

> 目标: 按 feature 组织，不重写组件逻辑
> 原则: 文件移动 + import 路径更新

### 6.1 新结构

```
frontend/src/
├── app/                        # Next.js 页面路由（不变）
├── features/                   # 按功能组织
│   ├── conversation/
│   │   ├── components/        # ChatPane, MessageBubble, FSMStatusBar
│   │   ├── hooks/             # useAIChat
│   │   └── stores/            # conversation store
│   ├── agent/
│   │   ├── components/        # AgentCard, AgentForm
│   │   ├── hooks/             # useAgent
│   │   └── stores/            # agent store
│   ├── match/
│   │   ├── components/        # MatchList, MatchCard
│   │   ├── hooks/             # useMatch
│   │   └── stores/            # match store
│   └── confirmation/
│       ├── components/        # HumanConfirmModal, OfferCard, InterviewCard
│       └── hooks/             # useConfirmation
├── shared/
│   ├── ui/                    # Button, Input, Modal, Card
│   ├── hooks/                 # useAuth, useSSE, useWebSocket
│   ├── stores/                # auth store, ui store
│   ├── api/                   # client.ts + routes.ts
│   └── types/                 # 全局类型
└── lib/                       # 第三方工具
    └── websocket.ts
```

### 6.2 迁移清单

| 原位置 | 新位置 |
|--------|--------|
| `components/ai-elements/` + `components/conversation/` + `components/FSMStatusBar.tsx` | `features/conversation/components/` |
| `components/interview/` + `components/offer/` + `components/HumanConfirmModal.tsx` | `features/confirmation/components/` |
| `components/feature/` | **删除**（分散到各 feature） |
| `components/ui/` | `shared/ui/` |
| `hooks/useAIChat.ts` | `features/conversation/hooks/` |
| `hooks/useAgent.ts` | `features/agent/hooks/` |
| `hooks/useMatch.ts` | `features/match/hooks/` |
| `hooks/useAuth.ts` + `useWebSocket.ts` + `useSSE.ts` | `shared/hooks/` |
| `stores/agent.ts` | `features/agent/stores/` |
| `stores/match.ts` | `features/match/stores/` |
| `stores/auth.ts` + `ui.ts` | `shared/stores/` |
| `lib/gateway/` | `shared/api/`（删 RateLimiter/CircuitBreaker，只保留 HTTP client） |

---

## 7. Phase 4: 修 bug + 验证

### 7.1 修复的 11 个关键 bug

| # | Bug | 严重度 | 修复 |
|---|-----|--------|------|
| 1 | negotiator 除零 panic | P0 | `min==target` 时 early return |
| 2 | FSM CanHandle/Handle 不一致 | P1 | 合并为单一 Handle |
| 3 | StatePaused 无转换表 | P1 | 补充 Resume/Rejected/Timeout |
| 4 | FSM 零并发保护 | P1 | 加 sync.Mutex |
| 5 | 消息发布失败静默丢失 | P1 | 重试+DLQ |
| 6 | runningGraphs 永久泄漏 | P1 | defer delete |
| 7 | WebSocket 重连无补推 | P1 | since 参数 |
| 8 | 工具空结果 vs 错误混淆 | P1 | 不同 status |
| 9 | JSON Unmarshal 忽略 | P1 | 检查 error |
| 10 | confirm 超时无处理 | P1 | 24h 自动 reject |
| 11 | AI 流式响应无错误信号 | P1 | 带 error channel |

### 7.2 验证方式

```bash
# 后端编译验证
cd backend && go build ./... && go vet ./...

# 前端编译验证
cd frontend && npm run build

# 启动验证
docker-compose up -d

# E2E 验证：完整 A2A 对话流程
playwright test tests/e2e/a2a-full-flow.spec.ts
# 检查: 用户注册 → 创建 Agent → 发起匹配 → A2A 对话 → Offer → 人类确认
# 同时 tail 后端日志观察是否有异常
```

---

## 8. 不做的（明确边界）

| 不做 | 原因 |
|------|------|
| 新增功能 | 重构目标是对现有代码做减法 |
| 改数据库 schema | 冒险，且不是当前主要矛盾 |
| 改 docker-compose | 部署配置已可用 |
| 引入新依赖 | 保持依赖最小化 |
| 写完整测试套件 | 重构完成后（Phase 5）再补 |
| 动 frontend 页面路由 | 页面结构合理，仅重组内部 |

---

## 9. 成功标准

重构完成时，以下全部勾选：

- [ ] `internal/service/` 目录消失
- [ ] `internal/agent/` 目录消失
- [ ] `internal/handler/` 目录消失
- [ ] `internal/model/` 目录消失（模型收归 core/models.go）
- [ ] `internal/repository/` 目录消失（合并到 adapters）
- [ ] `frontend/src/components/` 根目录消失
- [ ] `frontend/src/stores/` 根目录消失
- [ ] `frontend/src/hooks/` 根目录消失（迁移到 feature/shared）
- [ ] 11 个关键 bug 全部修复
- [ ] Legacy AI 降级链全部删除，Eino 为唯一入口
- [ ] `docker-compose up` 一键启动
- [ ] 完整 A2A 对话流程可走通（前端 UI + 后端日志同步验证）
- [ ] 新人阅读代码，5 分钟内理解架构（README 架构图更新）

---

## 10. 时间估算

| Phase | 内容 | 预估 |
|-------|------|------|
| Phase 0 | 精确诊断 | 1-2 天 |
| Phase 1 | 删除尘余 | 1 天 |
| Phase 2 | Backend 架构重组 | 2-3 天 |
| Phase 3 | Frontend 重组 | 1 天 |
| Phase 4 | 修 bug + 验证 | 1 天 |
| **合计** | | **6-8 天** |
