# Feature Specification: Production Hardening

**Feature Branch**: `020-production-hardening`

**Created**: 2026-05-14

**Status**: Draft

**Input**: User description: "生产就绪重构：安全加固、多租户隔离、A2A 核心修复、Eino 硬化、WebSocket 路由拆分、可观测性、Playwright CLI E2E"

## User Scenarios & Testing

### User Story 1 - 安全加固与多租户隔离 (Priority: P1)

开源项目要让社区放心部署，安全是门槛。当前 JWT 密钥硬编码在仓库里、auth cookie 设了 `httpOnly: false`、多租户只录 header 不做 SQL 过滤——这些是部署即违规的级别。

**Why this priority**: 没有安全，其他都白搭。数据和认证泄露会让项目刚上线就死。

**Independent Test**: 可独立验证：部署后用 A 用户注册，尝试用 B 用户的 token 访问 A 的数据，应返回 403。

**Acceptance Scenarios**:

1. **Given** `.env` 中不设 `JWT_SECRET`, **When** 后端启动, **Then** log.Fatalf 退出
2. **Given** 两个不同 tenant 的用户 A 和 B, **When** A 请求 B 的资源, **Then** 返回 403
3. **Given** 浏览器收到 auth cookie, **When** 检查 cookie 属性, **Then** `httpOnly=true`, `secure=true`, `sameSite=strict`
4. **Given** WebSocket 连接, **When** token 通过 query param 传递, **Then** token 不出现在日志和 Referer 中
5. **Given** 未认证请求, **When** 访问受保护端点, **Then** 返回 401

---

### User Story 2 - A2A 双 Agent 真正对话 (Priority: P1)

当前 `DualAgentNegotiationService` 只有 seeker 侧生成消息，recruiter 侧根本没执行。协议检测只用 "agree"+"terms" 字符串匹配——AI 返回"I accept your offer"无法识别。

**Why this priority**: A2A 自治对话是产品的核心价值主张。如果两个 Agent 不能真正对话，整个产品概念不成立。

**Independent Test**: 创建 match 后观察消息链：必须出现 INTRODUCTION → INTEREST → NEGOTIATION → OFFER → CONFIRM 完整链路。

**Acceptance Scenarios**:

1. **Given** 一个 match 已创建, **When** 触发 A2A 对话, **Then** seeker Agent 和 recruiter Agent 交替发送消息
2. **Given** AI 返回结构化 JSON（含 intent、salary、terms）, **When** 协议检测, **Then** 正确解析并推进 FSM
3. **Given** AI 返回任意文本, **When** 不包含可识别的 agreement 信号, **Then** 继续协商而非僵死
4. **Given** 对话超过 20 轮未达成, **When** 达到上限, **Then** 状态标记为 "failed" 并通知人工介入

---

### User Story 3 - 架构硬化与代码债务清理 (Priority: P2)

MessageQueueService 712 行处理 10+ 职责、WebSocket 一根路由处理人类和 Agent 两种协议、三套状态机并存、配置散落在 5 个文件里。

**Why this priority**: 长期维护成本。不改的话每加一个功能都像在雷区上走路。

**Independent Test**: 每个拆出来的服务都有清晰的接口和测试，原来的集成测试全部通过。

**Acceptance Scenarios**:

1. **Given** MessageQueueService, **When** 重构后, **Then** 拆为 QueueProcessor + AIOrchestrator + FSMEventHandler 三个独立服务
2. **Given** WebSocket 路由, **When** 重构后, **Then** `/api/messages/:matchId/ws` 仅处理人类消息, `/api/a2a/:matchId/ws` 仅处理 Agent 对话
3. **Given** 三套状态机, **When** 重构后, **Then** 合并为 `Match.Status` 单一状态来源
4. **Given** 三套 SalaryNegotiator, **When** 重构后, **Then** 只保留 `internal/agent/negotiator.go` 一个实现
5. **Given** 配置管理, **When** 重构后, **Then** 统一从 `internal/config/config.go` 加载

---

### User Story 4 - 可观测性与基础设施 (Priority: P2)

Health check 永远返回 "healthy"、没有结构化日志、没有 RED 指标、rate limiter 完全禁用。

**Why this priority**: 没有可观测性，生产环境出了问题只能靠猜。

**Independent Test**: 关掉 RabbitMQ 后 `/health` 应返回 degraded 状态。

**Acceptance Scenarios**:

1. **Given** 数据库断开, **When** 访问 `/health`, **Then** 返回 `status: "unhealthy"` 并注明 db 不可达
2. **Given** 请求经过系统, **When** 处理后, **Then** 有结构化日志（slog），包含 requestID、耗时、状态码
3. **Given** API 被调用, **When** rate limiter 启用, **Then** 超过阈值后返回 429

---

### User Story 5 - Playwright CLI 全流程 E2E (Priority: P3)

当前的 Playwright 测试是 test framework 方式，需要一套 playwright-cli 驱动的验收测试，覆盖安全、A2A 对话、FSM 流转、工具调用等。

**Why this priority**: 质量门禁。所有重构完成后需要一套可复现的验收流程。

**Independent Test**: 直接运行 `tests/e2e-cli/run-all.sh`，不需要手动操作浏览器。

**Acceptance Scenarios**:

1. **Given** 两个 playwright session（seeker + recruiter）, **When** 执行全流程, **Then** 完成从注册到 agent 对话的完整链路
2. **Given** 测试脚本, **When** 运行, **Then** 验证所有消息 intent 类型、FSM 状态、工具调用返回真实数据

---

### Edge Cases

- 当 AI API 不可用时，系统应降级而非崩溃（已有部分 fallback，需要强化）
- 当 RabbitMQ 断连后恢复，队列消息不应丢失（DLQ 重试）
- 当两个 Agent 进入死循环（互相同意又反悔），需要超时熔断
- 当 JWT 密钥轮换时，旧 token 应优雅失效而非直接 500
- 多租户场景下，管理员应能看到所有 tenant 数据，但普通用户严格隔离

## Requirements

### Functional Requirements

- **FR-001**: 系统必须从环境变量加载 JWT 密钥，启动时若未设则 crash
- **FR-002**: 所有数据表必须支持 tenant_id 字段隔离
- **FR-003**: 所有 repository 查询方法必须过滤当前 tenant
- **FR-004**: Auth cookie 必须设置 httpOnly=true, secure=true, sameSite=strict
- **FR-005**: WebSocket 认证 token 不得出现在 URL query param 中
- **FR-006**: 双 Agent 对话必须是真正的双向交替，而非单侧独白
- **FR-007**: 协议检测必须基于结构化 JSON 解析，而非字符串匹配
- **FR-008**: WebSocket 路由必须拆分（人类 /api/messages/:matchId/ws, Agent /api/a2a/:matchId/ws）
- **FR-009**: MessageQueueService 必须拆分为至少 3 个独立服务
- **FR-010**: 状态机必须合并为单一状态来源（Match.Status）
- **FR-011**: Health check 必须验证 DB / RabbitMQ / AI API 连通性
- **FR-012**: Rate limiter 必须启用（先 IP-based，后 user-based）
- **FR-013**: Eino 工具必须连接真实 repository，不得返回 mock 数据
- **FR-014**: Playwright CLI 测试必须覆盖安全验证、A2A 对话、FSM 流转三个核心场景

### Key Entities

- **Tenant**: 多租户隔离单元，每个用户注册时自动生成 UUID 作为 tenant_id（一用户一租户）
- **Match (增强)**: 统一状态机字段，合并 FSMState + Status
- **WebSocket Route**: 区分人类/A2A 两条路由，不同认证和协议
- **Service (重构)**: MessageQueueService 拆分为 QueueProcessor / AIOrchestrator / FSMEventHandler
- **Config**: 集中化配置结构体，替代散落的 os.Getenv

## Success Criteria

### Measurable Outcomes

- **SC-001**: 修复 PR 合并后，`grep -r "JWT_SECRET" .env` 在仓库中返回空
- **SC-002**: 两个不同 tenant 的用户互相访问数据均返回 403
- **SC-003**: A2A 对话链完整包含 INTRODUCTION → INTEREST → NEGOTIATION → OFFER → CONFIRM
- **SC-004**: `/health` 在 DB 断开时返回 `degraded`，在 RabbitMQ 断开时返回 `degraded`
- **SC-005**: MessageQueueService 行数从 712 降至 80 以下
- **SC-006**: 项目中 `grep -r "extracted_from_conversation\|Sample Job\|Mock Candidate" ` 返回 0 次
- **SC-007**: playwright-cli 测试脚本一次性通过所有 5 个场景

## Clarifications

### Session 2026-05-14

- Q: 多租户 tenant_id 生成策略? → A: 每个注册用户独立 UUID tenant_id（一用户一租户）

## Assumptions

- 多租户使用 UUID tenant_id 隔离，每个注册用户独立一个 tenant（一用户一租户）
- Eino 框架继续使用但不扩展，仅修复现有工具使其返回真实数据
- AI API 假设使用 LongCat Flash（兼容 OpenAI），回退策略已有基础
- WebSocket 路由拆分后前端代码不需要改动（前端仍连 `/api/messages/:matchId/ws`）
- Playwright CLI 测试脚本使用 bash 编写，不引入额外测试框架