# Feature Specification: A2A Business Logic Fixes

**Feature Branch**: `021-a2a-business-logic-fixes`

**Created**: 2026-05-14

**Status**: Draft

**Input**: User description: "修复 A2A 业务逻辑缺陷：响应路由、匹配分数、对话意图演进"

## User Scenarios & Testing

### User Story 1 - A2A 响应交替路由 (Priority: P1)

当前系统收到用户消息后，AI 生成的响应消息的 `sender_agent_id` 与原始消息相同（都是 seeker Agent 发给自己）。正确的行为应该是 seeker 发送 → recruiter 响应，形成真正的 A2A 交替对话。

**Why this priority**: 没有交替就不叫 A2A 对话，核心产品逻辑错误。

**Independent Test**: 发送 INQUIRY 消息后，响应消息的 sender_agent_id 应与原始消息不同。

**Acceptance Scenarios**:

1. **Given** 用户从 seeker 发送 INQUIRY, **When** 系统生成 auto-response, **Then** 响应消息的 sender_agent_id 应为 recruiter agent
2. **Given** recruiter 响应后继续对话, **When** 下一轮生成, **Then** sender 应交替变化
3. **Given** match 有 seeker 和 recruiter 两个 agent, **When** 生成响应, **Then** 自动选择对方 agent 作为发送者

---

### User Story 2 - Match 分数动态计算 (Priority: P1)

当前 `/api/matches/auto` 返回的 score 始终为 1（最大可能值），而不是基于技能、地点、经验等因素的动态计算。

**Why this priority**: 分数无区分度，匹配功能无法真正使用。

**Independent Test**: 创建两个不同匹配，分数应不同且反映匹配质量。

**Acceptance Scenarios**:

1. **Given** 完全匹配的 seeker 和 job, **When** 创建 match, **Then** score≈1.0
2. **Given** 完全不匹配的 seeker 和 job, **When** 创建 match, **Then** score≈0.0
3. **Given** 部分匹配的 seeker 和 job, **When** 创建 match, **Then** 0 < score < 1
4. **Given** 分数计算结果, **When** 验证, **Then** 不是硬编码的 0.9 或 1.0

---

### User Story 3 - 对话意图演进 (Priority: P2)

当前 AI 响应后 intent 始终停留在 INQUIRY，没有按照 INQUIRY → INTRODUCTION → INTEREST → NEGOTIATION → OFFER 的链路推进。

**Why this priority**: 对话不推进则无法完成完整的招聘流程。

**Independent Test**: 发送 INQUIRY 后轮询消息，intent 应从 INQUIRY 演进到 INTRODUCTION。

**Acceptance Scenarios**:

1. **Given** 发送 INQUIRY, **When** AI 响应, **Then** 响应 intent 应为 INTRODUCTION
2. **Given** 多轮对话后, **When** 检查消息链, **Then** 应包含至少 3 种不同 intent

---

### User Story 4 - Eino 响应路径修复 (Priority: P2)

`generateEinoResponse` 在 `ai_orchestrator.go` 中始终返回 nil，导致所有 AI 请求都走 legacy 回退路径。

**Why this priority**: Eino 集成半残废，投入了框架却没有产生价值。

**Independent Test**: Eino 路径生成响应时不应返回 nil。

**Acceptance Scenarios**:

1. **Given** Eino Runner 已配置, **When** 生成响应, **Then** 不走 legacy 回退路径
2. **Given** Eino 生成的响应, **When** 存入数据库, **Then** content_xml 是有效 XML

---

### Edge Cases

- 当没有 recruiter agent 时，响应应优雅降级而非崩溃
- 当 match 分数计算时，缺少技能数据应给出中等分数而非满分
- 多轮对话中某一方 agent 不可用时，系统不应死循环

## Requirements

### Functional Requirements

- **FR-001**: A2A 响应消息的 sender_agent_id 必须对应当前对话的另一方
- **FR-002**: Match score 必须基于 seeker 技能和 job 要求的匹配度动态计算
- **FR-003**: AI 响应 intent 必须根据对话上下文演进，不得停留在初始 intent
- **FR-004**: Eino `generateEinoResponse` 必须实现并返回有效响应
- **FR-005**: 所有分数和响应内容不得包含硬编码占位符值

### Key Entities

- **AutoResponse (修复)**: sender_agent_id 改为对端 agent
- **Match scoring (实现)**: 基于技能关键词匹配的加权算法
- **Intent chain (演进)**: INQUIRY → INTRODUCTION → INTEREST → NEGOTIATION → OFFER

## Success Criteria

- **SC-001**: INQUIRY 的 auto-response sender_agent_id != 原始 sender
- **SC-002**: 两个不同匹配的 score 不同
- **SC-003**: E2E 测试中至少出现 3 种不同 intent
- **SC-004**: Eino 路径不再返回 nil

## Assumptions

- 当前使用 `agent.Type` 判断 seeker/recruiter
- Match 分数暂时按简单的技能关键词匹配（不引入 NLP）
- Intent 演进在 `message_queue_service.go` 的 `handleAgentMessage` 中控制
