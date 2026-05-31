#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
  echo ""
  echo "Shutting down..."
  [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null || true
  [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null || true
  wait 2>/dev/null || true
  echo "Done."
}
trap cleanup EXIT INT TERM

echo "=== JobLinker Dev Environment ==="

echo "Starting Go backend..."
(cd "$ROOT_DIR" && go run ./cmd/server/main.go) &
BACKEND_PID=$!

echo "Starting Next.js frontend..."
(cd "$ROOT_DIR/frontend" && npm run dev) &
FRONTEND_PID=$!

echo "Waiting for backend (port 8080)..."
for i in $(seq 1 30); do
  if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ 2>/dev/null | grep -q "404\|405\|200"; then
    echo "Backend ready"
    break
  fi
  if [ "$i" -eq 30 ]; then echo "Backend failed to start"; exit 1; fi
  sleep 1
done

echo "Waiting for frontend (port 3000)..."
for i in $(seq 1 60); do
  if curl -s -o /dev/null http://localhost:3000 2>/dev/null; then
    echo "Frontend ready"
    break
  fi
  if [ "$i" -eq 60 ]; then echo "Frontend failed to start"; exit 1; fi
  sleep 1
done

echo ""
echo "=== Both servers running ==="
echo "  Frontend: http://localhost:3000"
echo "  Backend:  http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop both servers."

wait
