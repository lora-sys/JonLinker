# JobLinker 系统架构

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 前端 | Next.js 16 + React 19 | App Router, Turbopack |
| 状态 | Zustand + persist middleware | 本地存储持久化 |
| 样式 | Tailwind CSS | 组件库 |
| 后端 | Go + Gin + GORM | REST API |
| 数据库 | PostgreSQL + pgvector | 向量存储 |
| 消息队列 | RabbitMQ | 异步任务 |
| AI | LongCat Flash API | OpenAI 兼容 |

## 系统架构图

```
┌──────────────────────────────────────────────────────────────────┐
│                         用户浏览器                                │
│                    (http://localhost:3000)                        │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │    Next.js Frontend   │
                    │  ┌────────────────┐   │
                    │  │ GatewayClient  │   │
                    │  │  (统一入口)    │   │
                    │  └───────┬────────┘   │
                    │          │           │
                    │  ┌───────▼────────┐  │
                    │  │ ServiceRoutes   │  │
                    │  │ (路由配置)      │  │
                    │  └────────────────┘   │
                    └──────────┬───────────┘
                               │
                 HTTP/WebSocket
                               │
                    ┌──────────▼──────────┐
                    │    Go Backend        │
                    │  ┌────────────────┐  │
                    │  │ JWT Middleware │  │
                    │  │ X-User-ID      │  │
                    │  │ X-Agent-ID    │  │
                    │  │ X-Tenant-ID   │  │
                    │  └───────┬────────┘  │
                    │          │           │
                    │  ┌───────▼────────┐  │
                    │  │   Handlers     │  │
                    │  │   Services     │  │
                    │  │   Repositories │  │
                    │  └───────┬────────┘  │
                    └──────────┼───────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼────────┐ ┌─────▼──────┐  ┌──────▼──────┐
     │   PostgreSQL    │ │  RabbitMQ  │  │  LongCat    │
     │   + pgvector   │ │  (Queue)   │  │  AI API     │
     └─────────────────┘ └───────────┘  └─────────────┘
```

## API Gateway 统一入口

所有前端请求、Agent 工具调用、MCP 工具调用都通过 `GatewayClient` 统一入口：

```
┌─────────────────────────────────────────────────────────────────┐
│                     GatewayClient                                │
│                                                                  │
│  import { GatewayClient, ServiceRoutes } from '@/lib/gateway'   │
│                                                                  │
│  const gateway = new GatewayClient({                            │
│    baseURL: 'http://localhost:8080',                            │
│    userId: 'user-123',                                          │
│    agentId: 'agent-456',                                        │
│    tenantId: 'acme-corp',                                       │
│  })                                                             │
│                                                                  │
│  // 使用路由映射                                                   │
│  const agents = await gateway.request(ServiceRoutes.agents.list)│
│                                                                  │
│  // WebSocket 连接                                                │
│  const ws = gateway.connectWebSocket(ServiceRoutes.messages.ws,│
│    { matchId: 'match-789' })                                    │
└─────────────────────────────────────────────────────────────────┘
```

### Header 注入

每个请求自动携带：
- `X-User-ID`: 用户身份
- `X-Agent-ID`: AI Agent 身份
- `X-Tenant-ID`: 多租户隔离

### 路由映射

```typescript
// 路由定义 (routes.ts)
export const ServiceRoutes = {
  agents: { list, create, get, update, delete },
  jobs: { list, create, get, update },
  matches: { list, get, autoCreate, confirm },
  interviews: { list, create, update, getByMatch, confirm, cancel },
  offers: { getByMatch, create, accept, decline },
  messages: { list, send, websocket },
  privacy: { export, delete },
}
```

## A2A 自动对话流程

```
┌─────────────────────────────────────────────────────────────────┐
│                   A2A Autonomous Negotiation                     │
│                                                                  │
│  ┌─────────────┐                              ┌─────────────┐   │
│  │ Seeker Agent │ ◄────── XML Messages ──────►│ Recruiter  │   │
│  │  (求职方)   │                              │   Agent    │   │
│  │             │                              │  (招聘方)   │   │
│  └──────┬──────┘                              └──────┬─────┘   │
│         │                                            │          │
│         └──────────────────┬─────────────────────────┘          │
│                            ▼                                    │
│              ┌─────────────────────────────┐                    │
│              │   MessageQueueService        │                    │
│              │   - AI Response (LongCat)    │                    │
│              │   - Tool Executor           │                    │
│              │     • schedule_interview    │                    │
│              │     • create_offer          │                    │
│              │     • query_jobs            │                    │
│              │   - Memory Service          │                    │
│              │     • preference_vector     │                    │
│              │     • conversation_summary  │                    │
│              └───────────────┬─────────────┘                    │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ▼                    ▼                    ▼             │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐     │
│  │ PostgreSQL  │      │  RabbitMQ   │      │ LongCat API  │     │
│  │ + pgvector  │      │  (async)    │      │             │     │
│  └─────────────┘      └─────────────┘      └─────────────┘     │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              Human Final Confirmation ONLY                  │ │
│  │   - Agent reaches agreement autonomously                   │ │
│  │   - Human receives only final offer/interview confirmation  │ │
│  │   - No mid-negotiation human intervention                  │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 对话状态机

```
INQUIRY → INTRODUCTION → INTEREST → NEGOTIATION → OFFER → ACCEPT → CONFIRM
  │              │             │            │          │        │
  │              │             │            │          │        └── 结束
  │              │             │            │          └── 接受 offer
  │              │             │            └── 谈判薪资/入职条件
  │              │             └── 表示有兴趣
  │              └── 介绍职位详情 (title, salary, location, skills)
  └── 询问职位信息
```

## 三段式 Prompt 架构

```typescript
// AgentPromptService 生成三段式 prompt
const prompt = {
  MUST_DO: [
    "使用工具获取真实数据，不要编造",
    "自动提取用户偏好并存储到向量库",
    "自主决策工具调用时机",
  ],
  MUST_NOT_DO: [
    "不要透露用户敏感信息",
    "不要在未达成共识前结束对话",
    "不要伪造薪资数字",
  ],
  BEHAVIOR: [
    "保持专业友好的语气",
    "理解用户需求后查询相关工具",
    "协商不成时优雅退出",
  ]
}
```

## 目录结构

```
joblinker/
├── backend/
│   ├── cmd/server/main.go          # 入口
│   ├── internal/
│   │   ├── agent/
│   │   │   ├── function_definitions.go  # 工具定义
│   │   │   └── tool_executor.go         # 工具执行器
│   │   ├── handler/                     # HTTP handlers
│   │   │   ├── agent.go
│   │   │   ├── job.go
│   │   │   ├── match.go
│   │   │   ├── message.go
│   │   │   ├── interview.go
│   │   │   └── offer.go
│   │   ├── service/
│   │   │   ├── agent_prompt_service.go  # 三段式 prompt
│   │   │   ├── agent_memory_service.go  # 向量记忆
│   │   │   ├── message_queue_service.go # AI 响应 + 工具
│   │   │   ├── dual_agent_negotiation_service.go # 双 Agent 协商
│   │   │   ├── confirmation_service.go  # 人类确认
│   │   │   ├── preference_extraction_service.go # 偏好提取
│   │   │   └── reasoning_service.go     # ReAct 推理
│   │   ├── model/                       # 数据模型
│   │   ├── repository/                  # 数据访问
│   │   └── middleware/                  # JWT auth
│   └── pkg/
│       ├── ai/client.go                 # LongCat AI 客户端
│       └── rabbitmq/                    # RabbitMQ 客户端
│
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── gateway/                 # API Gateway 统一入口
│   │   │   │   ├── index.ts             # 导出
│   │   │   │   ├── client.ts            # GatewayClient
│   │   │   │   ├── routes.ts           # 路由映射
│   │   │   │   ├── types.ts            # 类型定义
│   │   │   │   └── middleware.ts        # RateLimiter, CircuitBreaker
│   │   │   ├── api_client.ts           # 兼容层
│   │   │   └── websocket.ts            # WebSocket hook
│   │   ├── hooks/
│   │   │   ├── useWebSocket.ts         # WS 连接管理
│   │   │   └── useAuth.ts              # 认证状态
│   │   ├── stores/
│   │   │   └── auth.ts                 # Zustand store
│   │   ├── app/                        # Next.js pages
│   │   └── components/                 # UI 组件
│   └── package.json
│
└── specs/
    ├── 010-agent-ai-upgrade/          # AI Agent 升级
    └── 011-api-gateway/               # API Gateway
```

## 数据模型

### Agent
- `id`: UUID
- `user_id`: 所属用户
- `type`: seeker | recruiter
- `name`: Agent 名称
- `config_json`: 配置 (skills, preferences)

### Match
- `id`: UUID
- `seeker_agent_id`: 求职 Agent
- `recruiter_agent_id`: 招聘 Agent
- `job_id`: 匹配职位
- `status`: pending | confirmed | rejected

### Message
- `id`: UUID
- `match_id`: 所属对话
- `sender_agent_id`: 发送者
- `content_xml`: XML 格式消息
- `intent_type`: INQUIRY | INTRODUCTION | INTEREST | ...

### Interview
- `id`: UUID
- `match_id`: 所属匹配
- `datetime`: 面试时间
- `interview_type`: video | phone | onsite
- `status`: scheduled | confirmed | cancelled

### Offer
- `id`: UUID
- `match_id`: 所属匹配
- `salary`: 薪资
- `start_date`: 入职日期
- `status`: pending | accepted | declined

## 环境变量

### Backend (.env)
```
DATABASE_URL=postgres://joblinker:joblinker_dev@localhost:5432/joblinker
RABBITMQ_URL=amqp://joblinker:joblinker_dev@localhost:5672/
AI_API_KEY=ak_xxxxx
AI_BASE_URL=https://api.longcat.chat/openai
AI_MODEL=LongCat-Flash-Lite
JWT_SECRET=your-secret
PORT=8080
```

### Frontend (.env.local)
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 快速启动

```bash
# 1. 启动基础设施
cd /home/lora/repos/joblinker/joblinker
docker-compose up -d

# 2. 启动后端
cd backend && go run cmd/server/main.go

# 3. 启动前端
cd frontend && npm run dev
```