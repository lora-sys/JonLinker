# Plan: JonLinker 大重构

> Source PRD: [PRD.md](./PRD.md)

## Architectural decisions（贯穿所有 Phase 的不变决策）

### 后端架构
- **四层架构**: `core/`（纯领域）→ `engine/`（Eino 编排）→ `adapters/`（外部依赖）→ `transport/`（传输层）
- **依赖方向**: `transport → engine → core`, `transport → adapters`, `engine → core`, `engine → adapters`
- **禁止反向依赖**: `core` 不可引用 `adapters`, `engine` 不可引用 `transport`, `adapters` 不可引用 `engine`
- **对话编排**: Eino 为唯一入口（删除 ADK/Legacy fallback 降级链）
- **DI**: `cmd/server/main.go` 手动组装依赖
- **消息队列**: RabbitMQ（保留，添加发布重试 + DLQ）
- **数据访问**: 所有 Repository 合并进 `adapters/postgres.go`

### 前端架构
- **按 feature 组织**: `features/{conversation,agent,match,confirmation}/`
- **共享代码**: `shared/ui/`, `shared/hooks/`, `shared/stores/`, `shared/api/`, `shared/types/`
- **页面路由**: 保持不变（Next.js App Router）
- **删除**: `components/`、`hooks/`、`stores/` 根目录

### 不变的内容
- DB schema 不变
- 前端页面路由不变
- docker-compose 不变
- 不引入新依赖
- 不新增业务功能

---

## Phase 0: 精确诊断

**User stories**: N/A（技能获取，非用户可见）

### What to build

在执行任何删除或移动之前，使用工具扫描代码库，生成数据驱动的决策依据。

### Tracer bullet 路径

| 诊断工具 | 扫描范围 | 产出 |
|----------|----------|------|
| `go vet` + `staticcheck` | 所有 `.go` 文件 | 未使用代码列表 |
| `grep -r "eino" backend/` | 全仓 | Eino 使用程度报告 |
| `grep -r "legacy\|fallback\|mock\|generateEinoResponse"` | backend/ | Legacy 代码位置清单 |
| 统计 `import.*from.*components` | frontend/src/ | 未使用组件列表 |
| `find . -name "*_test.go"` | 全仓 | 测试文件清单 |
| `grep -r "类型名" backend/` 对 `model/` 每个类型 | backend/internal/model/ | 每个 model 引用次数 |
| `grep -r "引用" backend/` 对 `repository/` 每个文件 | backend/internal/repository/ | 每个 repo 引用次数 |

### Acceptance criteria

- [ ] 生成"删除清单"（引用次数 = 0 的文件）
- [ ] 生成"迁移清单"（需要移动但有大量引用的文件）
- [ ] 生成"保留清单"（看起来冗余但实际在用的文件）
- [ ] 以上三份清单记录在文档中，供 Phase 1 使用

---

## Phase 1: 删除尘余代码

**User stories**: 2, 4, 5, 6, 7, 8, 9, 10

### What to build

根据 Phase 0 的诊断结果，安全删除确定不用的代码。

### Tracer bullet 路径

**第一轮（安全删除）**：删除引用次数 = 0 的文件
- 前端：删除未引用的组件、hooks、stores
- 后端：删除未引用的 model、repository、空文件

**第二轮（死代码）**：删除有引用但逻辑为空的代码
- `generateEinoResponse` 返回 nil 的 stub
- 空函数体
- 未被调用的常量/变量

**第三轮（Legacy 路径）**：确认 Eino 已替代后，删除
- ADK runner 调用
- Legacy fallback 降级链
- 旧的 AI 编排逻辑

每轮操作后运行：
```bash
cd backend && go build ./...
cd frontend && npm run build
```

### Acceptance criteria

- [ ] 第一轮：引用次数 = 0 的文件全部删除
- [ ] 第二轮：空 stub 和死代码全部删除
- [ ] 第三轮：Legacy 降级链全部删除
- [ ] 每轮删除后 `go build` 和 `npm run build` 通过
- [ ] 编译不过时回滚并检查遗漏引用

---

## Phase 2: Backend 架构重组

**User stories**: 1, 2, 3, 4, 5, 6, 7, 8

### What to build

将保留的后端代码迁移到 `core/engine/adapters/transport` 四层架构，不修改业务逻辑。每个迁移步骤都是一条垂直切片——从一个旧位置提取代码放入新位置，更新所有 import，编译验证。

### Tracer bullet 路径

**Step 1**: 创建新目录结构
```bash
mkdir -p internal/{core,engine/nodes,adapters,transport/{rest,ws}}
```

**Step 2**: 提取 `core/`（零外部依赖的纯逻辑）
- `agent/fsm.go` → `core/fsm.go`（修复：合并 CanHandle/Handle、补充 StatePaused 转换表）
- `agent/negotiator.go` → `core/negotiation.go`（修复：除零 panic，min==target 时 early return）
- `model/user.go` + 精简后的核心模型 → `core/models.go`

**Step 3**: 构建 `adapters/`（外部依赖薄封装层）
- `pkg/ai/client.go` → `adapters/ai_client.go`（修复：mockRound 并发、流式错误信号）
- `pkg/rabbitmq/` → `adapters/rabbitmq.go`（添加：发布重试 3 次 + DLQ）
- `repository/` 所有文件 → `adapters/postgres.go`（合并为一个文件）
- `pkg/parser/` → `adapters/parser.go`
- Jina Embedding 客户端 → `adapters/embedding.go`

**Step 4**: 构建 `engine/nodes/`（Eino Graph 节点）
- `agent/decision.go` + `pkg/ai` → `nodes/ai_node.go`
- `agent/tool_executor` + `function_definitions` → `nodes/tool_node.go`
- `agent_memory` + `preference_extraction` → `nodes/memory_node.go`
- `confirmation_service` + `interview_service` → `nodes/confirm_node.go`
- `eino/` 目录现有内容 → `engine/graph.go` + `engine/runner.go`
- 新建 `engine/runner.go`（RabbitMQ 驱动的 Graph 执行器）

**Step 5**: 重写 `transport/`（薄层 handlers）
- `handler/a2a.go` → `transport/rest/agent_handler.go`
- `handler/match.go` → `transport/rest/match_handler.go`
- `handler/message.go` + `handler/a2a_sse.go` → `transport/rest/message_handler.go`
- `handler/` 中 WS 部分 → `transport/ws/websocket_handler.go`（添加 `?since=<timestamp>` 补推）
- 保留：`handler/health.go`, `handler/auth.go`（移动到 `transport/rest/`）

**Step 6**: 组装 `cmd/server/main.go`
- 删除旧 import
- 按顺序初始化各层：core → adapters → engine → transport
- 启动 server

**Step 7**: 删除旧目录
```bash
rm -rf internal/{service,agent,handler,model,repository,cache,config}
```
精简 `pkg/`，只保留 `pkg/shared/`

### Acceptance criteria

- [ ] Step 1: 新目录结构创建
- [ ] Step 2: core/ 编译通过，FSM/negotiator 修复集成
- [ ] Step 3: adapters/ 编译通过，RabbitMQ 带重试+DLQ
- [ ] Step 4: engine/ 编译通过，Eino Graph 可执行
- [ ] Step 5: transport/ 编译通过，WS 带补推
- [ ] Step 6: main.go 启动成功
- [ ] Step 7: 旧目录全部删除
- [ ] 每步后 `go build ./...` 通过

---

## Phase 3: Frontend 重组

**User stories**: 9, 10

### What to build

将前端代码按 feature 重组，只做文件移动 + import 路径更新，不改组件逻辑。

### Tracer bullet 路径

**移动点 1**: `conversation` feature
- `components/ai-elements/*` + `components/conversation/*` + `components/FSMStatusBar.tsx` → `features/conversation/components/`
- `hooks/useAIChat.ts` → `features/conversation/hooks/`
- 创建 `features/conversation/stores/`（如有需要）

**移动点 2**: `agent` feature
- `components/AgentCard.tsx`, `components/AgentForm.tsx` → `features/agent/components/`
- `hooks/useAgent.ts` → `features/agent/hooks/`
- `stores/agent.ts` → `features/agent/stores/`

**移动点 3**: `match` feature
- `components/MatchList.tsx`, `components/MatchCard.tsx` → `features/match/components/`
- `hooks/useMatch.ts` → `features/match/hooks/`
- `stores/match.ts` → `features/match/stores/`

**移动点 4**: `confirmation` feature
- `components/interview/*` + `components/offer/*` + `components/HumanConfirmModal.tsx` → `features/confirmation/components/`
- `hooks/useConfirmation.ts` → `features/confirmation/hooks/`

**移动点 5**: `shared/` 目录
- `components/ui/*` → `shared/ui/`
- `hooks/useAuth.ts` + `useWebSocket.ts` + `useSSE.ts` → `shared/hooks/`
- `stores/auth.ts` + `stores/ui.ts` → `shared/stores/`
- `lib/gateway/` → `shared/api/`（删 RateLimiter/CircuitBreaker）

**移动点 6**: 删除旧目录
- `components/feature/` → 删除（分散到各 feature）
- 删除空的 `components/`、`hooks/`、`stores/` 根目录

### Acceptance criteria

- [ ] features/conversation/ 移动完成
- [ ] features/agent/ 移动完成
- [ ] features/match/ 移动完成
- [ ] features/confirmation/ 移动完成
- [ ] shared/ 重组完成
- [ ] 旧根目录全部删除
- [ ] `npm run build` 通过

---

## Phase 4: 修 bug + 验证

**User stories**: 11, 12, 13, 14, 15, 16, 17, 18, 19, 20

### What to build

修复 11 个关键 bug 并完整验证。这是重构的收尾阶段——确保系统可用。

### Tracer bullet 路径（按 bug 修复）

| # | Bug | 位置 | 修复方式 |
|---|-----|------|----------|
| 1 | negotiator 除零 panic | `core/negotiation.go` | min==target 时 early return |
| 2 | FSM CanHandle/Handle 不一致 | `core/fsm.go` | 合并为单一 Handle 方法 |
| 3 | StatePaused 无转换表 | `core/fsm.go` | 补充 Resume/Rejected/Timeout 转换 |
| 4 | FSM 零并发保护 | `core/fsm.go` | 加 sync.Mutex |
| 5 | 消息发布失败静默丢失 | `adapters/rabbitmq.go` | 指数退避重试 3 次 + DLQ |
| 6 | runningGraphs 泄漏 | `engine/runner.go` | defer delete |
| 7 | WebSocket 重连无补推 | `transport/ws/websocket_handler.go` | 支持 `since=<timestamp>` 参数 |
| 8 | 工具空结果 vs 错误混淆 | `engine/nodes/tool_node.go` | 返回不同 status |
| 9 | JSON Unmarshal 忽略 | 全局 | 检查所有 JSON error 返回值 |
| 10 | confirm 超时无处理 | `engine/nodes/confirm_node.go` | 24h 自动 reject + 通知 |
| 11 | AI 流式响应无错误信号 | `adapters/ai_client.go` | 使用带 error 的结果 channel |

### 验证

```bash
# 后端编译验证
cd backend && go build ./... && go vet ./...

# 前端编译验证
cd frontend && npm run build

# 启动验证
docker-compose up -d

# E2E 完整流程
cd frontend && npx playwright test tests/e2e/a2a-full-flow.spec.ts

# 观察日志
docker-compose logs -f backend
```

**新增 E2E 测试**：`frontend/tests/e2e/a2a-full-flow.spec.ts`
```
流程: 用户注册 → 创建 Seeker Agent → 创建 Recruiter Agent → 发起匹配
     → A2A 对话 → Offer 生成 → 人类确认
断言: 每一步的 UI 状态正确，后端日志无 panic/error
```

### Acceptance criteria

- [ ] 11 个关键 bug 全部修复
- [ ] `go build ./...` + `go vet ./...` 通过
- [ ] `npm run build` 通过
- [ ] `docker-compose up` 一键启动
- [ ] Playwright E2E 完整 A2A 流程通过
- [ ] 后端日志无 panic / error
- [ ] README 架构图更新，新人 5 分钟理解

---

## 附录：完成标准检查清单

- [ ] `internal/service/` 目录消失
- [ ] `internal/agent/` 目录消失
- [ ] `internal/handler/` 目录消失
- [ ] `internal/model/` 目录消失（模型收归 core/models.go）
- [ ] `internal/repository/` 目录消失（合并到 adapters）
- [ ] `frontend/src/components/` 根目录消失
- [ ] `frontend/src/stores/` 根目录消失
- [ ] `frontend/src/hooks/` 根目录消失
- [ ] 11 个关键 bug 全部修复
- [ ] Legacy AI 降级链全部删除，Eino 为唯一入口
- [ ] `docker-compose up` 一键启动
- [ ] 完整 A2A 对话流程可走通
- [ ] README 架构图更新
