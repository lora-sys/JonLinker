#!/bin/bash
# JobLinker — 一键启动脚本
# Usage: bash start.sh                # from project root
#        bash /path/to/joblinker/start.sh  # from anywhere
set -euo pipefail

# ── 定位项目根目录 ──────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$SCRIPT_DIR"
DOCKER_COMPOSE="$ROOT/docker-compose.yml"
BACKEND_DIR="$ROOT/backend"
FRONTEND_DIR="$ROOT/frontend"

# 如果从 git 仓库根目录调用（Makefile 所在处），往下找一层
if [ ! -f "$BACKEND_DIR/go.mod" ]; then
  BACKEND_DIR="$ROOT/joblinker/backend"
  FRONTEND_DIR="$ROOT/joblinker/frontend"
  DOCKER_COMPOSE="$ROOT/joblinker/docker-compose.yml"
fi
if [ ! -f "$BACKEND_DIR/go.mod" ]; then
  echo "[ERROR] Cannot find backend/go.mod — are you in the project root?"
  exit 1
fi

# ── 颜色 ────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; CYAN='\033[0;36m'; NC='\033[0m'
info()  { echo -e "${CYAN}[INFO]${NC}  $1"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
fail()  { echo -e "${RED}[FAIL]${NC}  $1"; }

# ── 清理旧进程 ──────────────────────────────────────────────────
cleanup_old() {
  info "清理旧服务进程..."
  pkill -f "next dev" 2>/dev/null || true
  pkill -f "joblinker-server" 2>/dev/null || true
  # 等端口释放
  for port in 3000 8080; do
    while ss -tlnp 2>/dev/null | grep -q ":$port "; do sleep 1; done
  done
  ok "旧进程已清理"
}

# ── 启动 Docker 基础设施 ────────────────────────────────────────
start_docker() {
  info "启动 Docker 基础设施 (PostgreSQL + Redis + RabbitMQ)..."
  cd "$(dirname "$DOCKER_COMPOSE")"
  docker compose -f "$DOCKER_COMPOSE" up -d 2>&1

  # 等待 PostgreSQL 就绪
  info "等待 PostgreSQL 就绪..."
  for i in $(seq 1 30); do
    if docker compose -f "$DOCKER_COMPOSE" exec -T postgres pg_isready -U joblinker 2>/dev/null; then
      ok "PostgreSQL 就绪"
      break
    fi
    if [ "$i" -eq 30 ]; then
      warn "PostgreSQL 未在预期内就绪，但后端会重试连接"
    fi
    sleep 1
  done

  # 等待 Redis 就绪
  info "等待 Redis 就绪..."
  for i in $(seq 1 15); do
    if docker compose -f "$DOCKER_COMPOSE" exec -T redis-stack redis-cli ping 2>/dev/null | grep -q PONG; then
      ok "Redis 就绪"
      break
    fi
    if [ "$i" -eq 15 ]; then warn "Redis 未就绪"; fi
    sleep 1
  done

  # RabbitMQ — 非关键依赖，不阻塞
  info "检查 RabbitMQ..."
  if docker compose -f "$DOCKER_COMPOSE" ps rabbitmq 2>/dev/null | grep -q "running"; then
    ok "RabbitMQ 就绪 (PID: $(docker inspect --format='{{.State.Pid}}' joblinker-rabbitmq))"
  else
    warn "RabbitMQ 未运行 — 事件驱动管道不可用，StateGraph 同步模式正常工作"
  fi

  cd "$ROOT"
}

# ── 启动后端 ────────────────────────────────────────────────────
start_backend() {
  info "构建并启动后端..."
  cd "$BACKEND_DIR"

  # 设置 RabbitMQ 凭证（与 docker-compose 保持一致）
  export RABBITMQ_USER="joblinker"
  export RABBITMQ_PASS="joblinker_dev"

  go build -o joblinker-server ./cmd/server 2>&1
  ./joblinker-server &>/tmp/joblinker-backend.log &
  BACKEND_PID=$!
  ok "后端进程 PID: $BACKEND_PID"

  # 等待 HTTP 就绪
  for i in $(seq 1 20); do
    if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
      ok "后端 HTTP :8080 就绪"
      break
    fi
    if [ "$i" -eq 20 ]; then
      fail "后端未能启动 — 详见 /tmp/joblinker-backend.log"
      tail -10 /tmp/joblinker-backend.log
      exit 1
    fi
    sleep 1
  done
  cd "$ROOT"
}

# ── 启动前端 ────────────────────────────────────────────────────
start_frontend() {
  info "启动前端..."
  cd "$FRONTEND_DIR"
  npm run dev &>/tmp/joblinker-frontend.log &
  FRONTEND_PID=$!
  ok "前端进程 PID: $FRONTEND_PID"

  for i in $(seq 1 30); do
    if curl -sf http://localhost:3000 >/dev/null 2>&1; then
      ok "前端 HTTP :3000 就绪"
      break
    fi
    if [ "$i" -eq 30 ]; then
      warn "前端启动超时 — 详见 /tmp/joblinker-frontend.log (npm run dev 可能仍在前台重连)"
    fi
    sleep 1
  done
  cd "$ROOT"
}

# ── 健康报告 ────────────────────────────────────────────────────
health_report() {
  echo ""
  echo -e "${GREEN}========================================${NC}"
  echo -e "${GREEN}       JobLinker 服务状态               ${NC}"
  echo -e "${GREEN}========================================${NC}"
  curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || echo "(health check unavailable)"
  echo ""
  echo -e "${CYAN}Frontend:${NC} http://localhost:3000"
  echo -e "${CYAN}Backend:${NC}  http://localhost:8080"
  echo -e "${CYAN}RabbitMQ:${NC} http://localhost:15672 (guest/guest)"
  echo ""
  echo -e "${YELLOW}测试账号:${NC}"
  echo "  seeker:    seeker_1779285708@test.com / password123"
  echo "  recruiter: recruiter_1779285708@test.com / password123"
  echo ""
  echo -e "${GREEN}使用 make 命令:${NC}"
  echo "  make logs          — 查看 Docker 日志"
  echo "  make restart       — 重启 Docker"
  echo "  make test-backend  — 运行后端测试"
  echo "  make test-frontend — 运行前端测试"
}

# ── main ────────────────────────────────────────────────────────
main() {
  echo -e "${CYAN}╔══════════════════════════════════════╗${NC}"
  echo -e "${CYAN}║        JobLinker 一键启动            ║${NC}"
  echo -e "${CYAN}╚══════════════════════════════════════╝${NC}"
  echo ""

  # 检测 docker 可用性
  if ! command -v docker &>/dev/null; then
    fail "Docker 未安装，请先安装 Docker"
    exit 1
  fi

  cleanup_old
  start_docker
  start_backend
  start_frontend
  health_report

  echo ""
  echo -e "${GREEN}所有服务已启动！按 Ctrl+C 停止所有服务。${NC}"
  echo ""

  # 等待任一子进程退出
  wait
}

# Trap Ctrl+C to clean up
trap 'info "收到中断信号，停止服务..."; cleanup_old; exit 0' INT TERM

main "$@"
