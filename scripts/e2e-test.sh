#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
BACKEND_PID=""
FRONTEND_PID=""

# Load .env
if [ -f "$ROOT_DIR/.env" ]; then
  export $(grep -v '^#' "$ROOT_DIR/.env" | xargs)
fi

cleanup() {
  echo "Cleaning up..."
  playwright-cli close 2>/dev/null || true
  [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null || true
  [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null || true
  wait 2>/dev/null || true
}
trap cleanup EXIT

echo "=== JobLinker E2E Test (Phase 2) ==="

echo "Building and starting Go backend..."
(cd "$ROOT_DIR" && go build -o /tmp/joblinker-server ./cmd/server/ && /tmp/joblinker-server) &
BACKEND_PID=$!

echo "Starting Next.js frontend..."
(cd "$ROOT_DIR/frontend" && npm run dev) &
FRONTEND_PID=$!

echo "Waiting for servers..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then echo "Backend ready"; break; fi
  [ "$i" -eq 30 ] && echo "Backend failed to start" && exit 1
  sleep 1
done

for i in $(seq 1 60); do
  if curl -sf http://localhost:$FRONTEND_PORT > /dev/null 2>&1; then echo "Frontend ready"; break; fi
  [ "$i" -eq 60 ] && echo "Frontend failed to start" && exit 1
  sleep 1
done

echo "Opening browser..."
playwright-cli open http://localhost:$FRONTEND_PORT
sleep 2

echo "Taking initial screenshot..."
playwright-cli screenshot --filename=e2e-phase2-initial.png

echo "Verifying page layout..."
SNAPSHOT=$(playwright-cli --raw snapshot)
echo "$SNAPSHOT" | head -30

echo "Phase 2 test: chat search for jobs..."
playwright-cli fill input[placeholder*="输入职位"] "找北京的前端岗位" --submit

echo "Waiting for search results..."
sleep 8

echo "Taking result screenshot..."
playwright-cli screenshot --filename=e2e-phase2-search.png

echo "Checking for results..."
FINAL_SNAPSHOT=$(playwright-cli --raw snapshot)
echo "$FINAL_SNAPSHOT" | head -40

echo "Checking for apply button..."
playwright-cli --raw snapshot | grep -q "生成申请" && echo "Apply button found ✓" || echo "Apply button not found"

echo "Phase 2 test: verify resume upload button exists..."
playwright-cli --raw snapshot | grep -q "上传 PDF 简历" && echo "Upload button found ✓" || echo "Upload button not found"
playwright-cli --raw snapshot | grep -q "简历助手" && echo "Resume chat panel found ✓" || echo "Resume chat panel not found"

echo "=== E2E Test Complete ==="
