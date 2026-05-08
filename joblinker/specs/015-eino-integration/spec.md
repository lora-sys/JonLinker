# Spec: Eino Framework Integration for JobLinker

## Overview

Integrate the Eino framework (by ByteDance) into the JobLinker A2A recruitment system to replace the custom-built AI orchestration layer while preserving existing communication infrastructure (WebSocket, RabbitMQ, API Gateway, Protobuf).

## System定位

The system operates in **全自动 A2A 自治模式**:
- Agent 为业务主体，人类仅做最终关键结果确认
- 双 Agent 完全自主对话，无人值守自动化运行
- Agent 自动从对话中提炼信息写入向量知识库
- Agent 拥有完全自主工具调用权限

## User Stories

### 作为系统，我想要：
1. **替换手工编排层** - 使用 Eino ADK 替代手写 FSM 状态机
2. **标准化工具注册** - 所有工具通过 Eino Tool 接口注册
3. **集中提示词管理** - 使用 Eino ChatTemplate 替代散乱的 prompt 字符串
4. **内置上下文管理** - 利用 Eino Memory 解决 token 爆炸问题
5. **多 Agent 编排** - 使用 Eino DeepAgent 实现 Seeker/Recruiter 协调

### 用户价值：
- 更高的代码可维护性（Eino 是成熟框架）
- 自动上下文精简（解决 token 溢出风险）
- 更灵活的 Agent 编排（DeepAgent 支持复杂协调）
- 更好的可扩展性（Eino 组件生态）

## Technical Requirements

### 保留不变：
- WebSocket 处理器 (`handler/message.go`)
- RabbitMQ 队列配置
- PostgreSQL + pgvector / Chroma 存储
- Protobuf 内容协商
- API Gateway 中间件

### 需要替换：
- `agent/fsm.go` 中的 switch 状态转换 → Eino Graph
- `agent/function_definitions.go` → Eino Tool 接口
- `agent/tool_executor.go` → Eino Tools
- `service/prompts/*.go` → Eino ChatTemplate
- `service/message_queue_service.go` 中的编排逻辑 → Eino Runner

### 新增组件：
- Eino ChatModel 组件（包装现有 AI client）
- Eino Agent (ChatModelAgent, DeepAgent)
- Eino Memory (短期记忆)
- Eino Retriever (长期向量记忆)

## Scope

### In Scope:
- Phase 1: Eino 基础设施搭建
- Phase 2: 工具迁移到 Eino Tool 接口
- Phase 3: 提示词迁移到 Eino ChatTemplate
- Phase 4: FSM 迁移到 Eino Graph + DeepAgent

### Out of Scope:
- 通信层改造（保持不变）
- 数据库 schema 变更
- Frontend 改造

## Success Criteria

1. **构建成功**: `go build ./cmd/server/...` 通过
2. **功能验证**: E2E 测试通过 - A2A 消息通过 Eino Agent 处理
3. **回归验证**: 现有 WebSocket、RabbitMQ、Protobuf 流程不变
4. **Token 控制**: 验证 Eino Memory 自动精简上下文

## Dependencies

- Go 1.18+
- github.com/cloudwego/eino (最新版本)
- github.com/cloudwego/eino-ext (最新版本)

## Risks

1. **Eino 版本兼容性**: 需要跟踪上游更新
2. **迁移复杂度**: 涉及多个核心组件，需要充分测试
3. **性能影响**: Eino Runner 引入额外抽象层，需性能测试