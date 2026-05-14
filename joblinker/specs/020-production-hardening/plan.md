# Implementation Plan: Production Hardening

**Branch**: `020-production-hardening` | **Date**: 2026-05-14 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/020-production-hardening/spec.md`

## Summary

对 JobLinker 项目进行生产就绪级加固：修复安全漏洞（JWT 密钥硬编码、httpOnly cookie、多租户不隔离）、修复 A2A 双 Agent 对话核心逻辑（当前只有独白）、硬化 Eino 集成（工具返回 mock 数据）、拆分 WebSocket 路由、分解 God Object（MessageQueueService 712 行）、建立可观测性，最后用 Playwright CLI 做全流程验收测试。

## Technical Context

**Language/Version**: Go 1.23 (backend), TypeScript 5.7 / Next.js 16 / React 19 (frontend)

**Primary Dependencies**: Gin v1.10, GORM v1.25, RabbitMQ amqp091-go v1.11, Gorilla WebSocket v1.5, CloudWeGo Eino v0.8.13, JWT v5.2, PostgreSQL via pgx v5.6, Chroma HTTP client (自定义), slog (Go 标准库)

**Storage**: PostgreSQL (pgvector) + Chroma 向量库 + RabbitMQ 消息队列

**Testing**: Go 标准 testing (后端), Jest (前端单元), Playwright Test (前端 E2E), Playwright CLI (验收测试)

**Target Platform**: Linux 服务器, Docker 容器化部署

**Project Type**: Web 应用 (Go 后端 + Next.js 前端 + PostgreSQL + RabbitMQ)

**Performance Goals**: 100 并发 Agent 对话, P95 响应 < 2s, WebSocket 消息延迟 < 500ms

**Constraints**: 
- 多租户数据严格隔离，不能通过修改 header 越权
- A2A 对话必须在无人工干预的情况下完成完整协商链路
- RabbitMQ 断连后重试，不允许丢消息
- AI API 不可达时系统降级而非崩溃

**Scale/Scope**: 开源项目初期目标 100 用户, 支持多租户架构

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| 测试优先: 所有 bug fix 前先写失效测试 | ✅ PASS | 已有 integration test 框架和 anti-cheat |
| 禁止 t.Skip: 缺失服务必须 t.Fatal | ✅ PASS | 已有 `verify_tests.sh` 检测 |
| 禁止硬编码: 所有数据运行时生成 | ✅ PASS | 已有 `anti_cheat.go` 强制执行 |
| 安全不可协商: 密钥/认证/cookie 修复必须做 | ✅ PASS | Phase 0 最高优先级 |
| 核心逻辑必须真实: A2A 不能是 mock/独白 | ✅ PASS | Phase 1 修复双向对话 |
| Context 链完整: 禁止 production 路径使用 context.Background() | ✅ PASS | Phase 3 修复 |

## Project Structure

### Documentation (this feature)

```text
specs/020-production-hardening/
├── spec.md              # 功能规格 (已完成)
├── plan.md              # 本文件
├── research.md          # 技术调研 (当前对话已完成)
├── data-model.md        # Phase 1 输出
├── quickstart.md        # 验证场景
├── contracts/           # API 契约变更
├── checklists/
│   └── requirements.md  # 规格质量检查
└── tasks.md             # Phase 2 输出 (/speckit.tasks)
```

### Source Code (变更范围)

```text
backend/
├── cmd/server/main.go                  # ─── [P0-1/P0-3/P0-5/P3-1] 路由注册、启动逻辑
├── internal/
│   ├── agent/
│   │   ├── fsm.go                      # [P2-3] 合并状态机
│   │   ├── negotiator.go               # [P2-4] 唯一 SalaryNegotiator
│   │   └── tool_executor.go            # [P1-2] 保持 (不变)
│   ├── config/
│   │   └── config.go                   # [P2-5] 新建: 集中配置
│   ├── eino/
│   │   ├── tools/job_tools.go          # [P1-2] 连接真实 repository
│   │   ├── agent/seeker_agent.go       # 不变
│   │   ├── agent/recruiter_agent.go    # 不变
│   │   └── memory/vector.go            # [P1-3] 接入真实向量检索
│   ├── handler/
│   │   ├── a2a.go                      # [P2-1] 新建: A2A WebSocket handler
│   │   ├── message.go                  # [P2-1] 剥离 Agent 代码, [P0-4] WebSocket 认证修复
│   │   └── health.go                   # [P3-1] 重写: 真实健康检查
│   ├── middleware/
│   │   ├── auth.go                     # [P0-1] JWT 密钥从环境变量加载
│   │   ├── cors.go                     # [P0-5] 删除通配符 origin
│   │   ├── rate_limit.go               # [P0-6/P3-4] 启用并改为内存版
│   │   └── error_handler.go            # [P2-4] 删除, 统一用 pkg/shared/errors.go
│   ├── model/user.go                   # [P0-3] Match 合并状态机字段
│   ├── repository/
│   │   ├── agent.go                    # [P0-3] 加 tenant 过滤, [P2-6] 构造注入 db
│   │   ├── job.go                      # [P0-3] 加 tenant 过滤
│   │   ├── match.go                    # [P0-3] 加 tenant 过滤
│   │   └── *.go                        # [P2-6] 所有 repository 构造注入 db
│   └── service/
│       ├── message_queue_service.go    # [P2-2] 拆分为协调层 (~80行)
│       ├── queue_processor.go          # [P2-2] 新建: RabbitMQ 消费/发布
│       ├── ai_orchestrator.go          # [P2-2] 新建: AI 响应编排
│       ├── fsm_event_handler.go        # [P2-2/P2-3] 新建: FSM 事件处理
│       ├── dual_agent_negotiation_service.go # [P1-1] 重写: 真正双向对话
│       └── observability_service.go    # [P3-2/P3-3] 结构化日志 + RED 指标
├── pkg/
│   ├── ai/client.go                    # [P1-2] 删除 SalaryNegotiatorSimple
│   ├── rabbitmq/rabbitmq.go            # [P0-4] 不暴露 token
│   └── shared/errors.go                # [P2-4] 唯一错误定义
├── .env                                # [P0-1] 从 git 移除, 加入 .gitignore
├── .gitignore                          # [P0-1] 添加 .env

frontend/
├── src/
│   ├── lib/
│   │   ├── api_client.ts              # [P3-5] 废弃, 删除
│   │   └── gateway/client.ts          # [P0-4] WebSocket 连接方式修复
│   ├── hooks/
│   │   ├── useAgent.ts                # [P3-5] 改用 GatewayClient
│   │   └── useAIChat.ts              # [P3-6] 修复 hook 嵌套
│   ├── middleware.ts                   # 不变
│   └── app/
│       ├── login/page.tsx             # [P0-2] 删除手动写 document.cookie
│       └── api/auth/login/route.ts    # [P0-2] cookie 改为 httpOnly
├── package.json                        # [P3-7] 清理未用依赖
└── next.config.js                      # [P0-5] 移除通配符 image hostname

tests/e2e-cli/
└── *.sh                               # [P4] Playwright CLI 验收测试
```

**Structure Decision**: 保持现有项目结构不变, 在现有文件上做修改而非重构目录. 新增文件: `handler/a2a.go`, `service/queue_processor.go`, `service/ai_orchestrator.go`, `service/fsm_event_handler.go`, `config/config.go`. 删除文件: `middleware/error_handler.go` (合并), `api_client.ts` (废弃).

## Execution Phases

### Phase 0: Security & Multi-Tenancy (P0)
**Story**: US-1 (安全加固与多租户隔离)
**估算**: 6-8h
**并行度**: 大部分任务可并行

| Task | File(s) | 并行 |
|------|---------|------|
| P0-1 JWT 密钥从 env 加载 + .env 移出版本控制 | `auth.go`, `.env`, `.gitignore` | [P] |
| P0-2 Auth cookie httpOnly/secure/sameSite 修复 | `login/route.ts`, `login/page.tsx` | [P] |
| P0-3 多租户隔离 (所有表加 tenant_id + repository 过滤) | `model/*.go`, `repository/*.go`, `middleware/gateway.go` | — |
| P0-4 WebSocket 认证 token 移到 Sec-WebSocket-Protocol | `handler/message.go`, `gateway/client.ts` | [P] |
| P0-5 CORS 安全头 + 删除通配符 + 限制 image hostname | `middleware/cors.go`, `next.config.js` | [P] |
| P0-6 Rate limiter 改为内存版并启用 | `middleware/rate_limit.go`, `main.go` | [P] |
| P0-7 更新 golang.org/x/net 修复 CVE | `go.mod` | [P] |

### Phase 1: A2A Core Fix + Eino Hardening (P1)
**Story**: US-2 (A2A 双 Agent 真正对话)
**估算**: 4-6h

| Task | File(s) | 并行 |
|------|---------|------|
| P1-1 重写 DualAgentNegotiationService (双向循环 + 结构化 JSON 协议检测) | `dual_agent_negotiation_service.go` | — |
| P1-2 Eino 工具注入真实 repository (删除 mock) | `eino/tools/job_tools.go` | [P] |
| P1-3 Eino VectorRetriever 接入 pgvector/Chroma | `eino/memory/vector.go` | [P] |

### Phase 2: Architecture Hardening (P2)
**Story**: US-3 (架构硬化)
**估算**: 6-8h

| Task | File(s) | 并行 |
|------|---------|------|
| P2-1 拆分 WebSocket 路由 (人类 vs A2A) | `handler/a2a.go` (新建), `handler/message.go`, `main.go` | — |
| P2-2 拆分 MessageQueueService (3 个新服务) | `service/message_queue_service.go`, `service/queue_processor.go`, `service/ai_orchestrator.go`, `service/fsm_event_handler.go` | — |
| P2-3 合并三套状态机 | `model/user.go`, `agent/fsm.go`, `service/fsm_integration.go` | [P] |
| P2-4 统一 Error 系统 | `middleware/error_handler.go` (删除), `pkg/shared/errors.go` | [P] |
| P2-5 配置集中化 | `config/config.go` (新建), 各 `getEnv` 调用点 | [P] |
| P2-6 Repository 构造注入 db | 所有 `repository/*.go` | [P] |
| P2-7 合并三套 SalaryNegotiator | `pkg/ai/client.go`, `handler/message.go`, `agent/negotiator.go` | [P] |

### Phase 3: Observability (P3)
**Story**: US-4 (可观测性)
**估算**: 4-6h

| Task | File(s) | 并行 |
|------|---------|------|
| P3-1 重写 Health Check (验证 DB/RabbitMQ/AI/Chroma) | `handler/health.go` | — |
| P3-2 结构化日志 (slog 替代 log.Printf) | 全局替换 (可渐进) | [P] |
| P3-3 RED 指标收集 | `service/observability_service.go` | [P] |
| P3-4 启用 Rate Limiter | `middleware/rate_limit.go`, `main.go` | [P] |
| P3-5 前端 API 统一 (废弃 api_client.ts) | `api_client.ts`, `useAgent.ts`, `stores/auth.ts` | — |
| P3-6 修复 useAIChat hook 嵌套 | `hooks/useAIChat.ts` | [P] |
| P3-7 清理未用依赖 + 死代码 | `package.json`, `useAIChat.ts` | [P] |

### Phase 4: Playwright CLI E2E Acceptance (P4)
**Story**: US-5 (Playwright CLI 测试)
**估算**: 4-5h
**依赖**: P0-1 到 P3-7 全部完成

| Task | File(s) |
|------|---------|
| P4-1 创建测试基础结构 (helpers.sh, random email/token 管理) | `tests/e2e-cli/helpers.sh` |
| P4-2 TC-1: 双用户注册 + 认证隔离 | `tests/e2e-cli/tc-1-register.sh` |
| P4-3 TC-2: Agent + Job 创建 (验证 fsm_state) | `tests/e2e-cli/tc-2-agents.sh` |
| P4-4 TC-3: 自动匹配 + 抗硬编码校验 | `tests/e2e-cli/tc-3-match.sh` |
| P4-5 TC-4: A2A 双 Agent 对话链路轮询 | `tests/e2e-cli/tc-4-conversation.sh` |
| P4-6 TC-5: 工具调用 + Offer/Interview 验证 | `tests/e2e-cli/tc-5-tools.sh` |
| P4-7 创建 run-all.sh 主入口 | `tests/e2e-cli/run-all.sh` |

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 保留 Eino v0.8.13 (大量间接依赖) | 用户选择用 Eino 框架稳定系统 | 删除 Eino 可减少 30% 依赖树, 但已做技术决策 |
| 三阶段拆分 MQ Service (P2-2 耗时大) | 当前 712 行 god object 必须分解 | 渐进式拆分的边际成本更高 |

## Quick Start (Verification)

```bash
# 1. 启动基础设施
cd joblinker && docker-compose up -d

# 2. 运行后端 (需设置 JWT_SECRET)
cd backend && JWT_SECRET=$(openssl rand -hex 32) go run cmd/server/main.go

# 3. 运行前端
cd frontend && npm run dev

# 4. 运行验收测试
cd joblinker && bash tests/e2e-cli/run-all.sh
```

## Success Criteria Verification

| SC-ID | Description | Verification Command |
|-------|-------------|---------------------|
| SC-001 | JWT 密钥不在仓库中 | `grep -r "JWT_SECRET" .env` → 空 |
| SC-002 | 多租户隔离 | 创建两个 tenant 用户, 互访 403 |
| SC-003 | A2A 完整对话链 | `tc-4-conversation.sh` 输出完整 intent 序列 |
| SC-004 | Health check 真实 | DB 断开返回 degraded |
| SC-005 | MQ Service 行数 < 80 | `wc -l message_queue_service.go` |
| SC-006 | 无硬编码 | `grep -r "extracted_from_conversation"` → 0 |
| SC-007 | 测试全过 | `run-all.sh` exit code 0 |
