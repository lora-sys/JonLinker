#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
  local exit_code=$?
  echo ""
  echo "Shutting down..."
  [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null && echo "Backend stopped" || true
  [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null && echo "Frontend stopped" || true
  wait 2>/dev/null || true
  echo "Done."
  exit "$exit_code"
}
trap cleanup EXIT INT TERM

cd "$ROOT_DIR"

echo "=== JobLinker Dev Environment ==="
echo ""

# ---- .env check ----
if [ ! -f .env ]; then
  echo "Error: .env not found. Copy .env.example to .env and configure your API keys."
  echo "  cp .env.example .env"
  exit 1
fi
export $(grep -v '^#' .env | xargs)

# ---- Backend ----
echo "Building Go backend..."
go build -o /tmp/joblinker-server ./cmd/server/

echo "Starting Go backend on :$BACKEND_PORT..."
/tmp/joblinker-server &
BACKEND_PID=$!

echo -n "Waiting for backend"
for i in $(seq 1 30); do
  if curl -sf http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
    echo " ready"
    break
  fi
  echo -n "."
  sleep 1
done
if ! curl -sf http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
  echo ""
  echo "Error: Backend failed to start within 30s. Check logs:"
  echo "  Run manually: /tmp/joblinker-server"
  exit 1
fi

# ---- Frontend ----
echo "Checking frontend dependencies..."
if [ ! -d "$ROOT_DIR/frontend/node_modules" ]; then
  echo "Installing frontend dependencies..."
  (cd "$ROOT_DIR/frontend" && npm install)
fi

echo "Starting Next.js frontend on :$FRONTEND_PORT..."
(cd "$ROOT_DIR/frontend" && npm run dev) &
FRONTEND_PID=$!

echo -n "Waiting for frontend"
for i in $(seq 1 60); do
  if curl -sf http://localhost:$FRONTEND_PORT > /dev/null 2>&1; then
    echo " ready"
    break
  fi
  echo -n "."
  sleep 1
done
if ! curl -sf http://localhost:$FRONTEND_PORT > /dev/null 2>&1; then
  echo ""
  echo "Warning: Frontend may still be starting. Check: http://localhost:$FRONTEND_PORT"
fi

echo ""
echo "=== Both servers running ==="
echo "  Frontend: http://localhost:$FRONTEND_PORT"
echo "  Backend:  http://localhost:$BACKEND_PORT"
echo "  Health:   http://localhost:$BACKEND_PORT/health"
echo ""
echo "Press Ctrl+C to stop both."
echo ""

wait
