# Tasks: Production Hardening

**Input**: Design documents from `specs/020-production-hardening/`

**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Tests**: 所有功能修改必须在集成测试中验证, 不能有 t.Skip, 不能有硬编码断言.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Security & Multi-Tenancy (US-1) 🔴 最高优先级

**Purpose**: 修复安全漏洞, 实现多租户隔离. 这是部署的前提条件.

### P0-1: JWT 密钥从环境变量加载

- [ ] T001 [P] [US1] 修改 `internal/middleware/auth.go`: 删除全局 `var jwtSecret`, 改为每次请求从环境变量读取 (`os.Getenv("JWT_SECRET")`), 启动时若未设则 `log.Fatalf`
- [ ] T002 [P] [US1] 将 `backend/.env` 从 git 中移除 (`git rm --cached backend/.env`), 创建 `.env.example` 作为模板
- [ ] T003 [P] [US1] 在 `backend/.gitignore` 中添加 `.env` 确保未来不会被提交

### P0-2: Auth Cookie 修复

- [ ] T004 [P] [US1] 修改 `frontend/src/app/api/auth/login/route.ts`: 设置 cookie 时 `httpOnly: true`, `secure: true`, `sameSite: 'strict'`
- [ ] T005 [P] [US1] 删除 `frontend/src/app/login/page.tsx` 中手动写 `document.cookie` 的代码 (约第 40 行), 仅依赖 Zustand persist
- [ ] T006 [P] [US1] 修改 `frontend/src/app/api/auth/register/route.ts`: 同样修复 cookie 属性 (与 login 一致)

### P0-3: 多租户隔离

- [ ] T007 [P] [US1] 在所有核心 model 中添加 `TenantID string` 字段: `User`, `Agent`, `Job`, `Match`, `Message`, `Offer`, `Interview`
- [ ] T008 [P] [US1] 修改 `internal/model/user.go`: 注册时自动生成 tenant_id (可为 UUID 或 email 哈希)
- [ ] T009 [P] [US1] 修改所有 `internal/repository/*.go`: 每个查询方法加 `WHERE tenant_id = ?` 过滤, 值从 context 中的 `X-Tenant-ID` 提取
- [ ] T010 [P] [US1] 修改 `internal/middleware/gateway.go`: `GetTenantID()` 不能默认返回 "default"——对于没有 tenant 上下文的请求直接返回 403
- [ ] T011 [US1] 修改 `internal/middleware/auth.go`: JWT claims 中包含 `tenant_id`, 并与 `X-Tenant-ID` header 交叉校验, 不一致则 403

### P0-4: WebSocket 认证修复

- [ ] T012 [P] [US1] 修改 `internal/handler/message.go`: `HandleWebSocket` 改为从 `Sec-WebSocket-Protocol` header 取 token, 而非 query param
- [ ] T013 [P] [US1] 修改 `frontend/src/lib/gateway/client.ts`: `connectWebSocket` 方法将 token 放在 `Sec-WebSocket-Protocol` header 而非 URL query
- [ ] T014 [P] [US1] 删除 query param token 的可疑日志输出 (避免 token 泄漏到日志中)

### P0-5: CORS + 安全头 + Image Hostname

- [ ] T015 [P] [US1] 修改 `internal/middleware/cors.go`: 删除通配符 origin (`"*"`) 支持, 只允许明确列出的 origin
- [ ] T016 [P] [US1] 在 `internal/middleware/cors.go` 中添加安全头: `Strict-Transport-Security`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`
- [ ] T017 [P] [US1] 修改 `frontend/next.config.js`: 将 `remotePatterns` 中的通配符 `hostname: '**'` 改为明确的白名单

### P0-6: Rate Limiter 修复

- [ ] T018 [P] [US1] 重写 `internal/middleware/rate_limit.go`: 从 DB 版改为内存版 (sync.Map + 滑动窗口), 基于 IP + userID 双 key
- [ ] T019 [US1] 修改 `cmd/server/main.go`: 注册 rate limit middleware 到所有 API 路由 (排除 health check)
- [ ] T020 [P] [US1] 删除 `_ = rateLimitRepo` (main.go 第 120 行) 和不再需要的 `repository/rate_limit.go`

### P0-7: 更新依赖

- [ ] T021 [P] [US1] 更新 `golang.org/x/net` 到 v0.34.0+ (解决已知 CVE)
- [ ] T022 [P] [US1] 运行 `go mod tidy` 清理未用依赖

**Checkpoint P0**: 运行 `go test -tags=integration ./tests/integration/...` 全部通过. 手动验证: 两个不同 tenant 用户互访 403.

---

## Phase 2: A2A Core Fix + Eino Hardening (US-2) 🔴

### P1-1: 重写 DualAgentNegotiationService

- [ ] T023 [P] [US2] 重写 `internal/service/dual_agent_negotiation_service.go`: 实现真正的双向循环——seeker proposal → recruiter response → seeker counter, 交替进行
- [ ] T024 [P] [US2] 在谈判循环中添加 context cancellation 检查, 支持优雅终止
- [ ] T025 [P] [US2] 修复 `isAgreement()`: 从 `strings.Contains("agree","terms")` 改为解析结构化 JSON 响应 (检查 `{ "agreed": true, "terms": {...} }` 格式)
- [ ] T026 [P] [US2] 添加多种 agreement 信号识别: "I accept", "Deal", "Agreed", "Confirmed" 等自然语言变体也纳入匹配
- [ ] T027 [P] [US2] 添加 AI 返回垃圾/空字符串/格式错误的防护: 解析失败时记录 warn 并重试, 最多 3 次后标记对话异常

### P1-2: Eino 工具连接真实数据

- [ ] T028 [P] [US2] 修改 `internal/eino/tools/job_tools.go`: 构造函数注入 `repository.JobRepository`, `QueryJobs` 改为真实 DB 查询, 删除所有 mock 数据
- [ ] T029 [P] [US2] `CreateOffer` 工具注入 `repository.OfferRepository`, 创建真实 Offer 记录
- [ ] T030 [P] [US2] `ScheduleInterview` 工具注入 `repository.InterviewRepository`, 创建真实 Interview 记录
- [ ] T031 [P] [US2] `GetCandidate` / `SearchCandidates` 工具注入 `repository.AgentRepository`, 返回真实 Agent 数据

### P1-3: Eino VectorRetriever 接入真实向量检索

- [ ] T032 [P] [US2] 修改 `internal/eino/memory/vector.go`: `VectorRetriever` 接入 `pkg/chroma` 客户端, 执行真实向量查询
- [ ] T033 [P] [US2] 当前返回空结果, 改为返回 Chroma/pgvector 中匹配的记忆

**Checkpoint P1**: 创建 match 后观察消息链, 必须出现 `INTRODUCTION → INTEREST → NEGOTIATION → OFFER → CONFIRM`. 检查数据库中有真实的 Offer/Interview 记录.

---

## Phase 3: Architecture Hardening (US-3) 🟡

### P2-1: WebSocket 路由拆分

- [ ] T034 [P] [US3] 新建 `internal/handler/a2a.go`: 实现 `HandleA2AWebSocket`——纯 XML 协议, HMAC 内部签名认证, 无 rate limiter
- [ ] T035 [US3] 修改 `internal/handler/message.go`: 精简 `HandleWebSocket`——删除所有 Agent XML 分支, 仅保留人类 JSON 消息处理
- [ ] T036 [US3] 修改 `cmd/server/main.go`: 注册新路由 `r.GET("/api/a2a/:matchId/ws", a2aHandler.HandleA2AWebSocket)` 在 auth middleware 组之前
- [ ] T037 [P] [US3] 修改 `internal/service/message_queue_service.go`: `Bidirectional A2A` 路由逻辑改为走新 A2A WebSocket, 不再混入人类 WS
- [ ] T038 [P] [US3] `message_queue_service.go` 删除 `conversationRounds` map (移到 `ai_orchestrator.go` 或新 A2A handler)

### P2-2: 拆分 MessageQueueService

- [ ] T039 [P] [US3] 新建 `internal/service/queue_processor.go`: `QueueProcessor`——RabbitMQ 消费/发布/重试/DLQ 逻辑, 从 `message_queue_service.go` 提取
- [ ] T040 [P] [US3] 新建 `internal/service/ai_orchestrator.go`: `AIOrchestrator`——AI 响应生成、Eino 回退链、工具执行编排, 从 `message_queue_service.go` 提取
- [ ] T041 [P] [US3] 新建 `internal/service/fsm_event_handler.go`: `FSMEventHandler`——FSM 转换 + 持久化 + WebSocket 广播, 从 `message_queue_service.go` 提取
- [ ] T042 [P] [US3] 重写 `internal/service/message_queue_service.go`: 缩减为协调层 (~80 行), 调用上述 3 个子服务

### P2-3: 合并三套状态机

- [ ] T043 [P] [US3] 修改 `internal/model/user.go` (Match struct): 删除 `FSMState` 字段, 用 `Status` 作为唯一状态来源, 扩展 `MatchStatus` 枚举覆盖所有 FSM 状态
- [ ] T044 [P] [US3] 修改 `internal/agent/fsm.go`: `agent.FSM` 改为读取/写入 Match.Status, 不维护独立 state
- [ ] T045 [P] [US3] 修改 `internal/service/fsm_integration.go`: 去掉桥接逻辑, 直接操作 Match.Status
- [ ] T046 [P] [US3] 更新所有引用 `Match.FSMState` 的地方改为 `Match.Status`

### P2-4: 统一 Error 系统

- [ ] T047 [P] [US3] 删除 `internal/middleware/error_handler.go` 中的 `AppError` 类型及其文件
- [ ] T048 [P] [US3] 确保所有 handler 和 service 统一使用 `pkg/shared/errors.go` 中的 `AppError`

### P2-5: 配置集中化

- [ ] T049 [P] [US3] 新建 `internal/config/config.go`: `Config` 结构体, 包含所有配置项, 启动时从 env 一次性加载并校验
- [ ] T050 [P] [US3] 删除 `cmd/server/main.go`, `internal/middleware/auth.go`, `pkg/ai/client.go`, `pkg/rabbitmq/rabbitmq.go` 中 4 处 `getEnv` 副本, 改用 `config.Config`

### P2-6: Repository 构造注入

- [ ] T051 [P] [US3] 所有 16 个 `internal/repository/*.go`: 构造函数改为 `NewXRepository(db *gorm.DB)`, 删除 `WithDB()`
- [ ] T052 [P] [US3] 修改 `cmd/server/main.go` 中所有 `NewXRepository()` 调用点, 传入 `db` 参数

### P2-7: 合并 SalaryNegotiator

- [ ] T053 [P] [US3] 删除 `internal/handler/message.go` 中 `SalaryNegotiator` (line 575-670), 改为引用 `agent.NewSalaryNegotiator`
- [ ] T054 [P] [US3] 删除 `pkg/ai/client.go` 中 `SalaryNegotiatorSimple` (line 297-325), 改为引用 `agent.NewSalaryNegotiator`, 确认 `backend/tests/unit/agent/agent_test.go` 测试仍通过

**Checkpoint P3**: `go build ./...` 通过. 旧的集成测试全部通过.

---

## Phase 4: Observability (US-4) 🟡

### P3-1: Health Check

- [ ] T055 [P] [US4] 重写 `internal/handler/health.go`: 检查 PostgreSQL 连接 (ping), RabbitMQ 连接, AI API (head 请求), Chroma 心跳
- [ ] T056 [P] [US4] 返回格式: `{"status":"healthy|degraded|unhealthy","checks":{"db":"ok","rabbitmq":"ok","ai_api":"ok","chroma":"ok"}}`
- [ ] T057 [P] [US4] 注册 `/health` 路由 (不走 auth middleware)

### P3-2: 结构化日志

- [ ] T058 [P] [US4] `internal/middleware/logging.go`: 从 `log.Println` 改为 `slog` 结构化日志
- [ ] T059 [P] [US4] 逐步替换 `log.Printf` 调用为 `slog.Info`/`slog.Warn`/`slog.Error` (先替换 handler 和 service 层)

### P3-3: RED 指标

- [ ] T060 [P] [US4] 在 `internal/service/observability_service.go` 中添加请求 Rate (req/s), Error Rate (%), Duration (P50/P95/P99) 三种指标的内存聚合
- [ ] T061 [P] [US4] middleware/logging 层收集每个请求的耗时和状态码

### P3-4: 启用 Rate Limiter

- [ ] T062 [P] [US4] Rate Limiter middleware 正式挂到所有 API 路由上 (依托 P0-6 的 T018 实现)

### P3-5: 前端 API 统一

- [ ] T063 [P] [US4] 修改 `frontend/src/hooks/useAgent.ts`: 所有 `fetch('/api/agents')` 改为 `gatewayClient.request(ServiceRoutes.agents.*)`
- [ ] T064 [P] [US4] 修改 `frontend/src/stores/auth.ts`: `apiClient.setToken()` 改为直接调用 `gatewayClient.setToken()`
- [ ] T065 [P] [US4] 废弃 `frontend/src/lib/api_client.ts`, 确认无引用后删除

### P3-6: Fix useAIChat Hook

- [ ] T066 [P] [US4] 修改 `frontend/src/hooks/useAIChat.ts`: 将 `useWebSocket` 提升到组件层, useAIChat 改为接收 ws 实例参数
- [ ] T067 [P] [US4] `buildWSUrl` 改为从 Zustand store 读 token, 而非直读 localStorage

### P3-7: Cleanup

- [ ] T068 [P] [US4] 删除 `useAIChat.ts:168-169` 的死字段 (`loadMoreMessages: () => {}`, `hasMoreMessages: false`)
- [ ] T069 [P] [US4] 从 `package.json` 删除 `@tanstack/react-query` (已声明未使用), 或其他改为开始使用
- [ ] T070 [P] [US4] 合并 `api-cookies.ts` 和 `auth-utils.ts` 为单一 auth 模块

**Checkpoint P4**: `npm run build` 通过. `go test ./tests/unit/... -count=1` 通过.

---

## Phase 5: Playwright CLI E2E Acceptance (US-5) 🔵

**⚠️ 依赖**: Phase 1-4 全部完成

### P4-1: 测试基础设施

- [ ] T071 [US5] 创建 `tests/e2e-cli/helpers.sh`: 包含 `random_email()`, `wait_for_message()`, `assert_intents()`, `random_name()` 等公用函数
- [ ] T072 [US5] 在 `.gitignore` 中添加 `tests/e2e-cli/*.json` (避免 auth state 文件被提交)

### P4-2: TC-1 双用户注册

- [ ] T073 [US5] 创建 `tests/e2e-cli/tc-1-register.sh`: 使用 `playwright-cli -s=seeker` 和 `playwright-cli -s=recruiter` 分别注册, 验证跳转 `/dashboard`, token 不同

### P4-3: TC-2 Agent + Job 创建

- [ ] T074 [US5] 创建 `tests/e2e-cli/tc-2-agents.sh`: 通过 API (curl) 创建 seeker agent + recruiter agent + job, 验证 `fsm_state="idle"`

### P4-4: TC-3 自动匹配

- [ ] T075 [US5] 创建 `tests/e2e-cli/tc-3-match.sh`: 调用 `/api/matches/auto` 创建 match, 验证 score 动态 (!= 0.9), fsm_state 正确, 抗硬编码

### P4-5: TC-4 A2A 对话

- [ ] T076 [US5] 创建 `tests/e2e-cli/tc-4-conversation.sh`: 轮询 `/api/messages/{matchId}`, 等待完整 intent 链 `INTRODUCTION → INTEREST → NEGOTIATION → OFFER → CONFIRM`, 验证 XML 格式, 抗硬编码

### P4-6: TC-5 工具调用

- [ ] T077 [US5] 创建 `tests/e2e-cli/tc-5-tools.sh`: 验证 Offer/Interview 记录存在, salary 非硬编码, agent 记忆已持久化

### P4-7: 主入口

- [ ] T078 [US5] 创建 `tests/e2e-cli/run-all.sh`: 按序执行 TC-1 到 TC-5, 任一失败则非零退出

**Checkpoint P5**: `bash tests/e2e-cli/run-all.sh` 一次性全部通过, exit code 0.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (US-1)**: 无依赖, 可以立即开始. 大部分任务可并行 [P].
- **Phase 2 (US-2)**: 依赖 Phase 1 的 P0-3 (多租户隔离) 完成, 因为数据查询依赖 tenant 过滤.
- **Phase 3 (US-3)**: 依赖 Phase 2 的 P1-2 (Eino 工具修复), 因为 MQ 拆分涉及 AI 编排路径.
- **Phase 4 (US-4)**: 依赖 Phase 3 的基础架构稳定.
- **Phase 5 (US-5)**: 依赖 Phase 1-4 全部完成 (作为验收门禁).

### Parallel Opportunities

- Phase 1 内部: T001-T022 中标记 [P] 的可以并行执行 (不同文件, 无依赖)
- Phase 2 内部: T023-T024 可并行, T028-T031 (4 个工具修复) 可并行
- Phase 3 内部: P2-3/P2-4/P2-5/P2-6/P2-7 可并行
- Phase 4 内部: P3-1/P3-2/P3-3 可并行, P3-5/P3-6/P3-7 可并行

### Implementation Strategy

```
Week 1                 Week 2                 Week 3
┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
│ Phase 1 (US-1)  │   │ Phase 3 (US-3)  │   │ Phase 5 (US-5)  │
│ 安全+多租户     │   │ 架构硬化        │   │ E2E 验收        │
├─────────────────┤   ├─────────────────┤   ├─────────────────┤
│ Phase 2 (US-2)  │   │ Phase 4 (US-4)  │   │                  │
│ A2A+Eino        │   │ 可观测性        │   │                  │
└─────────────────┘   └─────────────────┘   └─────────────────┘
```

## Summary

| Phase | Tasks | US | 估算 | 并行度 |
|-------|-------|----|------|--------|
| Phase 1 | T001-T022 (22 tasks) | US-1 | 6-8h | 高 (14/22 = 64% 可并行) |
| Phase 2 | T023-T033 (11 tasks) | US-2 | 4-6h | 中 (6/11 = 55% 可并行) |
| Phase 3 | T034-T054 (21 tasks) | US-3 | 6-8h | 高 (13/21 = 62% 可并行) |
| Phase 4 | T055-T070 (16 tasks) | US-4 | 4-6h | 中 (8/16 = 50% 可并行) |
| Phase 5 | T071-T078 (8 tasks) | US-5 | 4-5h | 低 (串行执行) |
| **Total** | **78 tasks** | | **24-33h** | |
