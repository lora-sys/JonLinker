# Spec: Eino Agent Routing Layer

## Overview

在现有 A2A 全自动招聘系统基础上，**只新增、不修改、不破坏**任何现有功能，添加三层路由体系：
1. 后端统一业务路由 (router.js)
2. Agent 专属路由 (/agent/:agentId/*)
3. API Gateway 路由适配

## System定位

The system operates in **全自动 A2A 自治模式** with multi-agent architecture:
- Agent 为业务主体，人类仅做最终关键结果确认
- 双 Agent 完全自主对话，无人值守自动化运行
- 多 Agent 并行运行，需要严格隔离

## User Stories

### 作为系统，我想要：

1. **统一业务路由** - 创建 router.js 入口，按模块分发：agent、chat、tool、fsm、memory、vector
2. **Agent 专属路由空间** - 每个 Agent 拥有独立路由 /agent/:agentId/*，实现隔离
3. **MCP 工具路由** - /tool/fetch-job、/tool/generate-offer 等标准工具调用路由
4. **FSM 事件路由** - /fsm/transition、/fsm/state-sync、/fsm/event 状态机路由
5. **记忆向量路由** - /memory/save、/memory/retrieve、/vector/search

### 用户价值：
- Agent 级别隔离，确保消息/状态/记忆/工具调用不串
- 多 Agent 并行运行不受干扰
- 统一的路由入口便于监控和限流
- 支持 X-User-ID、X-Agent-ID、X-Tenant-ID 隔离

## Technical Requirements

### 保留不变 (不修改现有功能)：
- WebSocket 处理器 (`handler/message.go`)
- RabbitMQ 队列配置
- PostgreSQL + pgvector / Chroma 存储
- Protobuf 内容协商
- API Gateway 中间件
- Vercel AI SDK 集成
- 真实 AI 调用

### 新增路由层 (只新增路由，不破坏现有逻辑)：

#### Layer 1: 后端统一业务路由
```
router.js (统一入口)
  ├── /agent/*    → Agent 路由
  ├── /chat/*    → Chat 路由
  ├── /tool/*    → MCP 工具路由
  ├── /fsm/*     → FSM 事件路由
  ├── /memory/*  → 记忆路由
  └── /vector/*  → 向量路由
```

#### Layer 2: Agent 专属路由
```
/agent/:agentId/chat     - Agent 对话
/agent/:agentId/state     - Agent 状态
/agent/:agentId/tool     - Agent 工具调用
/agent/:agentId/memory    - Agent 记忆
/agent/:agentId/profile   - Agent 配置
```

#### Layer 3: MCP 工具路由
```
/tool/fetch-job          - 获取职位
/tool/fetch-resume       - 获取简历
/tool/interview-invite   - 面试邀请
/tool/generate-offer      - 生成 Offer
/tool/match-vector        - 向量匹配
```

#### Layer 4: FSM 事件路由
```
/fsm/transition          - 状态转换
/fsm/state-sync          - 状态同步
/fsm/event               - 事件处理
```

#### Layer 5: 记忆向量路由
```
/memory/save             - 保存记忆
/memory/retrieve         - 检索记忆
/vector/search           - 向量搜索
```

### 路由要求：

所有路由必须支持以下 Header：
- `X-User-ID` - 用户隔离
- `X-Agent-ID` - Agent 隔离
- `X-Tenant-ID` - 租户隔离

### Header 处理：

| Header | 用途 | 示例 |
|--------|------|------|
| X-User-ID | 用户身份 | `user-123` |
| X-Agent-ID | Agent 身份 | `agent-456` |
| X-Tenant-ID | 租户隔离 | `tenant-789` |

## Scope

### In Scope：
- 创建 router.js 统一路由入口
- 实现 Agent 专属路由 /agent/:agentId/*
- 实现 MCP 工具路由 /tool/*
- 实现 FSM 事件路由 /fsm/*
- 实现记忆向量路由 /memory/*、/vector/*
- 所有路由支持 X-User-ID、X-Agent-ID、X-Tenant-ID

### Out of Scope：
- 修改现有业务逻辑
- 修改数据库 schema
- 修改 Frontend (除非配合路由调整)
- 修改 Vercel AI SDK 集成
- 修改 WebSocket 处理逻辑
- 修改 RabbitMQ 配置

## Success Criteria

1. **路由隔离验证**: 不同 Agent 的 /agent/:agentId/chat 互不干扰
2. **Header 传递验证**: X-User-ID、X-Agent-ID、X-Tenant-ID 在所有路由层正确传递
3. **回归验证**: 现有 WebSocket、RabbitMQ、Protobuf、真实 AI 调用流程不变
4. **多 Agent 并行**: 多个 Agent 同时调用路由不串消息

## 假设

- 使用 Go 标准库 mux 或 Gin 作为路由框架 (与现有后端一致)
- 路由层作为中间件或 wrapper，不修改现有 handler
- Agent ID 在路由参数中传递，与消息中的 Agent ID 校验一致

## Risks

1. **路由与现有 Handler 冲突**: 需要确保新路由不覆盖现有端点
2. **Header 传递丢失**: 中间件可能丢失 Header，需要显式传递
3. **性能影响**: 额外路由层可能增加延迟