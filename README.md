# JobLinker - A2A Agent Recruitment Platform

智能 A2A (Agent-to-Agent) 招聘平台，通过自主 AI Agent 进行招聘方与求职方的智能匹配和自动对话。

**核心定位**: 全自动 A2A 自治模式 — Agent 为业务主体，人类仅做最终关键结果确认。

## 核心特性

- **A2A 智能对话**: AI Agent 之间自动进行招聘对话 ( INTRODUCTION → INTEREST → NEGOTIATION → OFFER → CONFIRM )
- **真实 AI 驱动**: 基于 LongCat Flash API 的 AI 响应，非硬编码规则
- **三段式 Prompt**: MUST DO / MUST NOT DO / BEHAVIOR 架构
- **函数调用**: Agent 自主调用工具获取真实数据 (query_jobs, create_offer, schedule_interview)
- **短期记忆**: 10+ 消息会话上下文
- **长期向量记忆**: pgvector 存储用户偏好，自动召回
- **双 Agent 自治协商**: 求职 Agent 与招聘 Agent 自动协商薪资、入职条件
- **人类最终确认**: 仅在协商完成后的人类确认，中途无干预
- **实时消息推送**: WebSocket 支持即时消息收发
- **异步任务处理**: RabbitMQ 消息队列解耦 AI 响应

## 技术栈

- **后端**: Go + Gin + GORM + PostgreSQL + pgvector + RabbitMQ
- **前端**: Next.js 16 + React 19 + Tailwind CSS + Zustand
- **AI**: LongCat Flash API (OpenAI compatible)
- **消息队列**: RabbitMQ
- **数据库**: PostgreSQL (jsonb 存储 + pgvector 向量)
- **API Gateway**: 统一入口，Header 注入 (X-User-ID, X-Agent-ID, X-Tenant-ID)

## 快速启动

### 1. 启动基础设施 (Docker)

```bash
cd /home/lora/repos/joblinker/joblinker
docker-compose up -d
```

### 2. 启动后端

```bash
cd backend && go run cmd/server/main.go
```

后端地址: http://localhost:8080

### 3. 启动前端

```bash
cd frontend && npm run dev
```

前端地址: http://localhost:3000

## 服务地址

| 服务 | 地址 | 说明 |
|------|------|------|
| Frontend | http://localhost:3000 | Next.js |
| Backend API | http://localhost:8080 | Go API |
| PostgreSQL | localhost:5432 | 数据库 + pgvector |
| RabbitMQ | localhost:5672 | 消息队列 (guest/guest) |

## API Gateway 统一入口

所有前端请求、Agent 工具调用都通过 `GatewayClient` 统一入口：

```typescript
import { GatewayClient, ServiceRoutes } from '@/lib/gateway'

const gateway = new GatewayClient({
  baseURL: 'http://localhost:8080',
  userId: userId,
  agentId: agentId,
  tenantId: 'default',
})

// 使用路由映射
const agents = await gateway.request(ServiceRoutes.agents.list)
const offer = await gateway.post(ServiceRoutes.offers.create, {}, { salary: 150000 })

// WebSocket
const ws = gateway.connectWebSocket(ServiceRoutes.messages.websocket, { matchId: 'xxx' })
```

### 路由映射

```typescript
ServiceRoutes.agents.list    // GET  /api/agents
ServiceRoutes.agents.create  // POST /api/agents
ServiceRoutes.jobs.get       // GET  /api/jobs/:id
ServiceRoutes.offers.accept  // POST /api/offers/:matchId/accept
ServiceRoutes.messages.ws    // WS   /api/messages/:matchId/ws
```

### 自动 Header 注入

每个请求自动携带：
- `X-User-ID`: 用户身份
- `X-Agent-ID`: AI Agent 身份
- `X-Tenant-ID`: 多租户隔离

## A2A 对话流程

```
用户 (seeker agent)                 AI recruiter agent
      |                                    |
      |---- INQUIRY (感兴趣) ------------->|
      |<--- INTRODUCTION (职位介绍) --------|
      |                                    |
      |---- INTEREST (表示兴趣) ---------->|
      |<--- NEGOTIATION (薪资谈判) --------|
      |                                    |
      |---- NEGOTIATION (还价) ----------->|
      |<--- OFFER (发 offer) --------------|
      |                                    |
      |---- ACCEPT (接受) ---------------->|
      |<--- CONFIRM (确认) ----------------|
```

### 对话意图说明

- **INQUIRY**: 用户询问职位信息
- **INTRODUCTION**: AI 介绍职位详情
- **INTEREST**: 用户表示兴趣
- **NEGOTIATION**: 薪资谈判
- **OFFER**: 正式 offer
- **ACCEPT**: 用户接受
- **CONFIRM**: AI 确认
- **SCHEDULE**: 安排面试

### Agent 工具调用

Agent 可自主调用以下工具获取真实数据：

| 工具 | 功能 |
|------|------|
| `query_jobs` | 搜索职位 |
| `get_candidate` | 获取候选人信息 |
| `create_offer` | 创建 offer |
| `schedule_interview` | 安排面试 |
| `search_candidates` | 搜索候选人 |

## 项目结构

```
joblinker/
├── backend/
│   ├── cmd/server/           # 入口
│   ├── internal/
│   │   ├── agent/            # 函数定义 + 工具执行器
│   │   ├── handler/          # HTTP handlers
│   │   ├── service/          # 业务逻辑
│   │   │   ├── agent_prompt_service.go    # 三段式 prompt
│   │   │   ├── agent_memory_service.go    # 向量记忆
│   │   │   ├── message_queue_service.go   # AI 响应 + 工具
│   │   │   ├── dual_agent_negotiation_service.go # 双 Agent
│   │   │   ├── confirmation_service.go    # 人类确认
│   │   │   └── preference_extraction_service.go # 偏好提取
│   │   ├── model/            # 数据模型
│   │   ├── repository/       # 数据访问
│   │   └── middleware/       # JWT 认证
│   └── pkg/
│       ├── ai/client.go      # LongCat AI 客户端
│       └── rabbitmq/         # 消息队列
├── frontend/
│   └── src/
│       ├── lib/
│       │   ├── gateway/      # API Gateway 统一入口
│       │   │   ├── index.ts      # 导出
│       │   │   ├── client.ts     # GatewayClient
│       │   │   ├── routes.ts     # 路由映射
│       │   │   ├── types.ts      # 类型定义
│       │   │   └── middleware.ts  # RateLimiter, CircuitBreaker
│       │   ├── api_client.ts     # 兼容层
│       │   └── websocket.ts      # WebSocket hook
│       ├── hooks/             # React hooks
│       ├── stores/            # Zustand stores
│       ├── app/               # Next.js pages
│       └── components/        # UI 组件
├── docs/
│   ├── architecture.md       # 系统架构
│   └── gateway-architecture.md  # Gateway 架构图
└── specs/                    # 设计文档
```

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

## 测试

### 后端测试

```bash
cd backend && go test ./... -v
```

### 端到端测试 (playwright cli)

```bash
# 登录 -> 创建 Agent -> 创建 Job -> 创建 Match -> 对话
```

## 常用命令

```bash
# 重启后端
pkill -f server && go run cmd/server/main.go &

# 重启前端
pkill -f "next" && npm run dev &

# 查看 RabbitMQ 队列
docker exec joblinker-rabbitmq rabbitmqctl list_queues

# 查看数据库
docker exec -it joblinker-postgres psql -U joblinker -d joblinker
```

## 故障排查

### WebSocket 连接失败

1. 检查 `NEXT_PUBLIC_API_URL` 配置
2. 确认后端 WebSocket 路由 `/api/messages/:matchId/ws` 可访问

### AI 无响应

1. 检查 `AI_API_KEY` 和 `AI_BASE_URL`
2. 查看后端日志 `DEBUG: AI response failed`

### RabbitMQ 连接失败

1. 检查 `RABBITMQ_URL` 格式
2. 确认容器运行中 `docker ps | grep rabbitmq`

## 文档

- [系统架构](docs/architecture.md) - 详细技术架构
- [Gateway 架构](docs/gateway-architecture.md) - API Gateway 设计

MIT
