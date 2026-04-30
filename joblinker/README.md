# JobLinker - A2A Agent Recruitment Platform

智能 A2A (Agent-to-Agent) 招聘平台，通过自主 AI Agent 进行招聘方与求职方的智能匹配和自动对话。

## 核心特性

- **A2A 智能对话**: AI Agent 之间自动进行招聘对话 ( INTRODUCTION → INTEREST → NEGOTIATION → OFFER → CONFIRM )
- **真实 AI 驱动**: 基于 LongCat Flash API 的 AI 响应，非硬编码规则
- **实时消息推送**: WebSocket 支持即时消息收发
- **异步任务处理**: RabbitMQ 消息队列解耦 AI 响应

## 技术栈

- **后端**: Go + Gin + GORM + PostgreSQL + RabbitMQ
- **前端**: Next.js 16 + React 19 + Tailwind CSS + Zustand
- **AI**: LongCat Flash API (OpenAI compatible)
- **消息队列**: RabbitMQ
- **数据库**: PostgreSQL (jsonb 存储)

## 快速启动

### 1. 启动基础设施 (Docker)

```bash
cd /home/lora/repos/joblinker/joblinker
docker-compose up -d
```

### 2. 启动后端

```bash
cd /home/lora/repos/joblinker/joblinker/backend
go run cmd/server/main.go
```

后端地址: http://localhost:8080

### 3. 启动前端

```bash
cd /home/lora/repos/joblinker/joblinker/frontend
npm run dev
```

前端地址: http://localhost:3000

## 服务地址

| 服务 | 地址 | 说明 |
|------|------|------|
| Frontend | http://localhost:3000 | Next.js |
| Backend API | http://localhost:8080 | Go API |
| PostgreSQL | localhost:5432 | 数据库 |
| RabbitMQ | localhost:5672 | 消息队列 (guest/guest) |

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

## API 端点

### 认证
- `POST /api/auth/register` - 注册 (支持 seeker/recruiter)
- `POST /api/auth/login` - 登录

### Agent
- `GET /api/agents` - 列表
- `POST /api/agents` - 创建

### Job
- `POST /api/jobs` - 创建 (需 `structured` 字段)

### Match
- `POST /api/matches/auto` - 自动匹配 (seeker_agent_id + job_ids)

### Message
- `GET /api/messages/:matchId` - 获取消息历史
- `POST /api/messages/:matchId` - 发送消息
- `WS /api/messages/:matchId/ws?token=<jwt>` - WebSocket

## 项目结构

```
joblinker/
├── backend/
│   ├── cmd/server/           # 入口
│   ├── internal/
│   │   ├── handler/          # HTTP handlers
│   │   ├── service/          # 业务逻辑
│   │   │   ├── message_queue_service.go  # AI 响应生成
│   │   │   └── ...
│   │   ├── repository/        # 数据访问
│   │   ├── model/            # 数据模型
│   │   └── middleware/        # JWT 认证
│   └── pkg/rabbitmq/          # 消息队列
├── frontend/
│   ├── src/
│   │   ├── app/               # Next.js 页面
│   │   ├── components/chat/   # 聊天组件
│   │   ├── hooks/useChat.ts   # WebSocket hook
│   │   └── stores/auth.ts     # 状态管理
│   └── .env.local             # 环境配置
└── specs/                      # 设计文档
```

## 环境变量

### Backend (.env)
```
DATABASE_URL=postgres://joblinker:joblinker_dev@localhost:5432/joblinker
RABBITMQ_URL=amqp://joblinker:joblinker_dev@localhost:5672/
OPENAI_API_KEY=your-api-key
AI_BASE_URL=https://api.longcat.chat/openai
JWT_SECRET=your-secret
```

### Frontend (.env.local)
```
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_WS_URL=ws://localhost:8080
```

## 测试

### 后端测试
```bash
cd backend && go test ./... -v
```

### 端到端测试 (agent-browser)
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
1. 检查 `NEXT_PUBLIC_API_WS_URL` 配置
2. 确认后端 WebSocket 路由 `/api/messages/:matchId/ws` 可访问

### AI 无响应
1. 检查 `OPENAI_API_KEY` 和 `AI_BASE_URL`
2. 查看后端日志 `DEBUG: AI response failed`

### RabbitMQ 连接失败
1. 检查 `RABBITMQ_URL` 格式
2. 确认容器运行中 `docker ps | grep rabbitmq`

MIT
