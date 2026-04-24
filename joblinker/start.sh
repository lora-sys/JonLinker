#!/bin/bash
# JobLinker 一键启动脚本

set -e

echo "=== JobLinker 启动脚本 ==="

# 1. 检查 Docker 基础设施
echo "[1/4] 检查 Docker 基础设施..."
if docker ps | grep -q joblinker-postgres; then
    echo "  - Docker 基础设施已运行"
else
    echo "  - 启动 Docker 基础设施..."
    docker-compose up -d
    sleep 5
fi

# 2. 停止旧服务
echo "[2/4] 停止旧服务..."
pkill -f "next" 2>/dev/null || true
pkill -f "server" 2>/dev/null || true
sleep 2

# 3. 启动后端
echo "[3/4] 启动后端..."
cd /home/lora/repos/joblinker/joblinker/backend/cmd/server
go run . &>/tmp/backend.log &
BACKEND_PID=$!
echo "  - 后端 PID: $BACKEND_PID"
sleep 3

# 4. 启动前端
echo "[4/4] 启动前端..."
cd /home/lora/repos/joblinker/joblinker/frontend
npm run dev &>/tmp/frontend.log &
FRONTEND_PID=$!
echo "  - 前端 PID: $FRONTEND_PID"

# 等待服务就绪
sleep 5

# 检查状态
echo ""
echo "=== 服务状态 ==="
curl -s http://localhost:8080/health && echo " [后端 OK]" || echo " [后端 FAIL]"
curl -s -o /dev/null -w "Frontend: %{http_code}\n" http://localhost:3000 || echo " [前端 FAIL]"

echo ""
echo "=== 启动完成 ==="
echo "Frontend: http://localhost:3000"
echo "Backend:  http://localhost:8080"
echo ""
echo "测试账号:"
echo "Email:    test_e2e@example.com"
echo "Password: testpass123"
