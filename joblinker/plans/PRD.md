# JonLinker 大重构 PRD

> **PRD 日期**: 2025-06-24
> **仓库**: https://github.com/lora-sys/JonLinker
> **定位**: Agent-driven A2A 招聘平台，开源项目
> **基于**: [重构计划](joblinker/plans/refactor.md)

---

## Problem Statement

JonLinker 已经实现了大量功能，但代码质量严重阻碍了后续开发。当前有 **42 个已知 bug**（1 P0 + 21 P1 + 20 P2），其中包含生产级别的除零 panic 和消息静默丢失问题。代码结构混乱：`service/` 目录 15+ 个文件职责不清，`agent/` 与 `service/` 存在循环依赖，Legacy 降级链、空 stub、未引用 model 等尘余代码大量堆积。Eino 编排引擎的引入不完整，仍存在多条对话处理路径。

项目面临的不再是"功能不够"的问题，而是三层问题叠加：
1. **边界 bug**（除零 panic、FSM 死锁、消息静默丢失、并发无保护）
2. **尘余代码多**（Legacy 降级链、空 stub、未引用 model、重复逻辑）
3. **架构混乱**（职责不清、循环依赖、多路径并存）

> **不解决问题 1 的后果**: 生产环境继续出现 panic，用户对话丢失
> **不解决问题 2 的后果**: 后续开发人员无法判断哪些代码在用、哪些已废弃
> **不解决问题 3 的后果**: 新人上手成本高，bug 修复周期长，新功能引入风险大

---

## Solution

对现有代码进行一次**克制的架构重构**，遵循"先删后移、不重写业务逻辑、渐进验证、保持克制"四个原则：

### 四阶段执行

| Phase | 内容 | 目标 | 预估 |
|-------|------|------|------|
| Phase 0 | 精确诊断 | 用数据驱动删除决策，不凭猜测删代码 | 1-2 天 |
| Phase 1 | 删除尘余代码 | 安全删除确定不用的代码，为重组清场 | 1 天 |
| Phase 2 | Backend 架构重组 | 重组为 `core/engine/adapters/transport` 四层架构 | 2-3 天 |
| Phase 3 | Frontend 重组 | 按 feature 组织前端代码 | 1 天 |
| Phase 4 | 修 bug + 验证 | 修复 11 个关键 bug 并验证 | 1 天 |

### 新后端架构

```
backend/
├── cmd/server/main.go              # 入口，组装依赖 (DI)
├── internal/
│   ├── core/                       # 纯领域逻辑，零外部依赖
│   ├── engine/                     # Eino 编排引擎 — 唯一对话入口
│   ├── adapters/                   # 外部依赖适配器 — 薄封装层
│   └── transport/                  # 传输层 — 仅做验证 + 转发
```

依赖方向不可违反：`transport → engine → core`, `transport → adapters`, `engine → core`, `engine → adapters`

### 新前端结构

```
frontend/src/
├── app/                        # Next.js 页面路由（不变）
├── features/                   # 按功能组织
│   ├── conversation/
│   ├── agent/
│   ├── match/
│   └── confirmation/
├── shared/
│   ├── ui/                     # Button, Input, Modal, Card
│   ├── hooks/                  # useAuth, useSSE, useWebSocket
│   ├── stores/                 # auth store, ui store
│   ├── api/                    # client.ts + routes.ts
│   └── types/                  # 全局类型
└── lib/                        # 第三方工具
```

---

## User Stories

1. 作为开发者，我希望 backend 代码按 `core/engine/adapters/transport` 四层组织，以便新人 10 分钟理解架构
2. 作为开发者，我希望 FSM 状态机在 `core/` 包中且 CanHandle/Handle 一致，以避免运行时死锁
3. 作为开发者，我希望 Eino 成为唯一对话编排入口，以消除 ADK/Legacy fallback 降级链的维护负担
4. 作为后端维护者，我希望 `internal/service/` 目录消失，以消除职责不清和循环依赖
5. 作为后端维护者，我希望 `internal/agent/` 目录消失（核心逻辑收归 `core/`，编排逻辑收归 `engine/nodes/`）
6. 作为后端维护者，我希望 `internal/handler/` 目录消失（handler 重组为 `transport/rest/` + `transport/ws/`）
7. 作为后端维护者，我希望 `internal/model/` 目录消失（核心模型统一进 `core/models.go`）
8. 作为后端维护者，我希望 `internal/repository/` 目录消失（合并进 `adapters/postgres.go`）
9. 作为前端开发者，我希望 `frontend/src/components/` 根目录消失，组件按 feature 分组
10. 作为前端开发者，我希望 `frontend/src/stores/` 和 `frontend/src/hooks/` 根目录消失（迁移至 feature/shared）
11. 作为平台用户，我希望 negiotiator 不会因除零 panic 导致服务崩溃（P0 bug 修复）
12. 作为平台用户，我希望消息不会在发布失败时静默丢失（P1 bug 修复）
13. 作为 AI 调度维护者，我希望 `runningGraphs` 不会永久泄漏（P1 bug 修复）
14. 作为平台用户，我希望 WebSocket 断线重连能收到补推消息（P1 bug 修复）
15. 作为平台用户，我希望 AI 流式响应断开时能看到错误信号（P1 bug 修复）
16. 作为平台用户，我希望 confirmation 超时 24h 后自动 reject（P1 bug 修复）
17. 作为测试人员，我希望重构过程中每步都能通过 `go build` 编译验证，以确保不引入新问题
18. 作为运维人员，我希望重构后 `docker-compose up` 能一键启动完整系统
19. 作为 QA，我希望完整 A2A 对话流程（注册 → 创建 Agent → 匹配 → 对话 → Offer → 确认）可通过 Playwright E2E 测试验证
20. 作为开源贡献者，我希望 README 架构图更新后能在 5 分钟内理解项目结构

---

## 重构原则 / Implementation Decisions

### 核心原则

- **不重写业务逻辑**：只移动、重组、修 bug，核心算法不动
- **先删后移**：先清尘余，再迁移，最后修 bug
- **渐进验证**：每步执行后 `go build` 确保不引入编译错误
- **保持克制**：不新增功能，不动 DB schema，不引入新依赖

### 后端架构决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 架构模式 | 四层架构（core/engine/adapters/transport） | 职责单一、依赖方向清晰、可测试 |
| 对话编排 | Eino 为唯一入口 | 消除 ADK/Legacy 多路径问题 |
| DI 方式 | main.go 手动组装 | 避免引入 DI 框架，保持简单 |
| 消息队列 | RabbitMQ（保留并加重试+DLQ） | 已有基础设施 |
| AI 客户端 | LongCat AI（从 pkg/ai 精简） | 已有集成，仅精简 |

### 前端架构决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 组织方式 | 按 feature 分目录 | 契合 Next.js App Router 模式 |
| 共享代码 | `shared/ui/`, `shared/hooks/`, `shared/stores/` | 避免跨 feature 耦合 |
| 页面路由 | 保持不变 | 已合理，仅重组内部 |
| API 客户端 | `shared/api/`（删 RateLimiter/CircuitBreaker） | 简化，保留 HTTP client |

### 边界条件处理

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

## Phase 拆解

### Phase 0: 精确诊断（必须先做）

用数据驱动删除决策：

| # | 诊断项 | 方法 |
|---|--------|------|
| 1 | 统计所有 .go 文件引用关系 | `go vet` + `staticcheck` |
| 2 | 检查 Eino 真实集成程度 | 搜索 `eino` 全仓引用 |
| 3 | 识别 Legacy 路径 | 搜索 legacy/fallback/mock |
| 4 | 检查前端组件引用 | 统计 import 次数 |
| 5 | 检查测试文件 | 搜索 `_test.go` / `.spec.` |
| 6 | 检查 model 引用 | 每个 model 文件的引用计数 |
| 7 | 检查 repository 引用 | 每个 repo 文件的引用计数 |
| 8 | 检查 pkg/ 引用 | 每个子包的引用计数 |

### Phase 1: 删除尘余代码

三轮删除：
1. **确定无引用的**（引用次数 = 0）：未引用 model/repository/组件/stores/空文件
2. **死代码**：`generateEinoResponse` stub、空函数体、未调用常量
3. **Legacy 路径**：ADK runner、Legacy fallback、旧 AI 编排逻辑

### Phase 2: Backend 架构重组

7 步迁移：
1. 创建新目录结构
2. 提取 core（FSM、negotiator、核心模型）
3. 构建 adapters（AI 客户端、RabbitMQ、Postgres、Embedding、Parser）
4. 构建 engine nodes（Eino Graph 节点）
5. 重写 transport（薄层 REST + WebSocket handlers）
6. 组装 main.go
7. 删除旧目录

### Phase 3: Frontend 重组

按迁移清单移动文件并更新 import 路径。

### Phase 4: 修 bug + 验证

修复 11 个关键 bug，通过编译验证 + 启动验证 + E2E 验证。

---

## Testing Decisions

### 验证策略（非完整测试套件）

重构阶段的验证策略是**编译通过 + 关键路径 E2E**，不在此阶段追求测试覆盖率：

| 验证层级 | 方式 | 触发时机 |
|----------|------|----------|
| 编译验证 | `go build ./...` + `go vet ./...` | 每步迁移后 |
| 前端构建 | `npm run build` | 每步移动后 |
| 启动验证 | `docker-compose up -d` | Phase 4 |
| E2E 验证 | Playwright 完整 A2A 流程 | Phase 4 |

### 现有 E2E 测试覆盖

已有测试文件（将保留并更新 import 路径）：
- `tests/e2e/auth.spec.ts` — 注册/登录
- `tests/e2e/agent.spec.ts` — Agent 创建/列表
- `tests/e2e/ai-chat.spec.ts` — AI 对话/流式/线程隔离/持久化
- `tests/e2e/chat.spec.ts` — 聊天页面
- `tests/e2e/full-flow.spec.ts` — 完整 E2E 流程
- `tests/e2e/interview.spec.ts` — 面试页面
- `tests/e2e/matching.spec.ts` — 匹配页面
- `tests/e2e/offer.spec.ts` — Offer 页面
- `tests/e2e/resume.spec.ts` — 简历上传/生成
- `tests/e2e/admin.spec.ts` — 管理后台
- `tests/e2e/tool-call-ui.spec.ts` — 工具调用 UI
- `tests/e2e/settings-agent-check.spec.ts` — 设置页面验证

### 需新增的 E2E 测试

Phase 4 需要新增 `a2a-full-flow.spec.ts`：
- 用户注册
- 创建 Agent（Seeker + Recruiter）
- 发起匹配
- A2A 对话
- Offer 生成
- 人类确认

---

## Out of Scope

| 不做 | 原因 |
|------|------|
| 新增功能 | 重构目标是对现有代码做减法 |
| 改数据库 schema | 冒险，且不是当前主要矛盾 |
| 改 docker-compose | 部署配置已可用 |
| 引入新依赖 | 保持依赖最小化 |
| 写完整测试套件 | 重构完成后（Phase 5）再补 |
| 动 frontend 页面路由 | 页面结构合理，仅重组内部 |

---

## Success Criteria

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

## Security Considerations

- 重构不涉及安全模型改动
- JWT/CORS 中间件保留并移至新结构
- WebSocket 认证方式在迁移时保持现有规则
- 无新攻击面引入

## Monitoring & Observability

- 保留现有日志框架（`pkg/shared/logger.go`）
- 重构过程中不新增监控指标
- 验证阶段通过 tail 后端日志观察异常
