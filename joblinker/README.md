# JobLinker - A2A Agent Recruitment Platform

## 项目概览

智能 A2A (Agent-to-Agent) 招聘平台，通过自主 AI Agent 进行招聘方与求职方的智能匹配。

## 技术栈

- **后端**: Go + Gin + GORM + PostgreSQL + Redis + RabbitMQ + Chroma
- **前端**: Next.js 16 + React 19 + Tailwind CSS + Zustand
- **AI**: LongCat Flash API
- **基础设施**: Docker (Postgres, Redis, RabbitMQ, Chroma)

## 快速启动

### 1. 启动基础设施 (Docker)

```bash
cd /home/lora/repos/joblinker/joblinker
docker-compose up -d
```

检查状态:
```bash
docker-compose ps
```

### 2. 启动后端

```bash
cd /home/lora/repos/joblinker/joblinker/backend/cmd/server
go run .
```

后端地址: http://localhost:8080

### 3. 启动前端 (开发模式)

```bash
cd /home/lora/repos/joblinker/joblinker/frontend
npm run dev
```

前端地址: http://localhost:3000

### 4. (可选) Nginx 反向代理

如果需要通过 80 端口访问:

```bash
docker run -d \
  --name nginx-joblinker \
  --network host \
  -v /tmp/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v /tmp/sites-enabled:/etc/nginx/sites-enabled:ro \
  nginx:alpine
```

代理地址: http://localhost:8888

## 登录凭证

测试账号:
- Email: `test_e2e@example.com`
- Password: `testpass123`

## 服务地址汇总

| 服务 | 地址 | 说明 |
|------|------|------|
| Frontend (开发) | http://localhost:3000 | Next.js 开发服务器 |
| Backend API | http://localhost:8080 | Go API 服务器 |
| PostgreSQL | localhost:5432 | 数据库 |
| Redis | localhost:6379 | 缓存 |
| RabbitMQ | localhost:5672 | 消息队列 |
| Chroma | localhost:8000 | 向量数据库 |
| Nginx (可选) | http://localhost:8888 | 反向代理 |

## 常用命令

### 后端测试
```bash
cd /home/lora/repos/joblinker/joblinker/backend
go test ./... -v
```

### 前端构建
```bash
cd /home/lora/repos/joblinker/joblinker/frontend
npm run build
```

### 前端测试
```bash
cd /home/lora/repos/joblinker/joblinker/frontend
npx playwright test
```

### 查看日志
```bash
# 后端日志
tail -f /tmp/backend.log

# 前端日志
tail -f /tmp/frontend.log
```

### 重启服务
```bash
# 停止所有服务
pkill -f "next" 2>/dev/null
pkill -f "server" 2>/dev/null

# 重新启动后端
cd /home/lora/repos/joblinker/joblinker/backend/cmd/server && go run . &

# 重新启动前端
cd /home/lora/repos/joblinker/joblinker/frontend && npm run dev &
```

## 项目结构

```
joblinker/
├── backend/
│   ├── cmd/server/       # 后端入口
│   ├── internal/        # 业务逻辑
│   │   ├── agent/       # Agent 相关
│   │   ├── handler/     # HTTP 处理
│   │   ├── service/     # 服务层
│   │   └── repository/  # 数据访问
│   └── pkg/             # 公共包
├── frontend/
│   ├── src/
│   │   ├── app/         # Next.js 页面
│   │   ├── components/ # React 组件
│   │   ├── stores/      # Zustand 状态
│   │   └── hooks/       # 自定义 Hooks
│   └── tests/          # 测试
├── specs/               # 规格文档
└── docker-compose.yml  # 基础设施配置
```

## API 端点

### 认证
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录

### Agent
- `GET /api/agents` - 获取 Agent 列表
- `POST /api/agents` - 创建 Agent

### Jobs
- `GET /api/jobs` - 获取职位列表
- `POST /api/jobs` - 创建职位

### Matches
- `GET /api/matches` - 获取匹配列表

### Health Check
- `GET /health` - 服务健康检查

## 故障排查

### 页面只显示 HTML，无样式/脚本

1. 检查 Next.js 是否正常运行:
```bash
curl http://localhost:3000 | grep "_next/static"
```

2. 如果使用 Nginx，确保代理配置正确:
```nginx
location /_next/ {
    proxy_pass http://127.0.0.1:3000;
}
```

### 连接被拒绝

1. 检查端口占用:
```bash
ss -tlnp | grep -E "3000|8080"
```

2. 重启对应服务

### Docker 容器状态异常

```bash
docker-compose logs [服务名]
docker-compose restart [服务名]
```

## 开发指南

### 前端组件

新增 UI 组件放在 `frontend/src/components/ui/`

### 后端新增 API

1. 在 `internal/handler/` 添加 handler
2. 在 `cmd/server/main.go` 注册路由
3. 编写单元测试

### 代码规范

- 前端: ESLint + Prettier
- 后端: gofmt + golint

## 许可证

MIT
