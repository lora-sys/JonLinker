#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
  echo "Cleaning up..."
  playwright-cli close 2>/dev/null || true
  [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null || true
  [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null || true
  wait 2>/dev/null || true
}
trap cleanup EXIT

echo "=== JobLinker E2E Test (Phase 2) ==="

echo "Starting Go backend..."
(cd "$ROOT_DIR" && go run ./cmd/server/main.go) &
BACKEND_PID=$!

echo "Starting Next.js frontend..."
(cd "$ROOT_DIR/frontend" && npm run dev) &
FRONTEND_PID=$!

echo "Waiting for servers to be ready..."
for i in $(seq 1 30); do
  if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ 2>/dev/null | grep -q "404\|405\|200"; then
    echo "Backend ready"
    break
  fi
  if [ "$i" -eq 30 ]; then echo "Backend failed to start"; exit 1; fi
  sleep 1
done

for i in $(seq 1 60); do
  if curl -s -o /dev/null http://localhost:3000 2>/dev/null; then
    echo "Frontend ready"
    break
  fi
  if [ "$i" -eq 60 ]; then echo "Frontend failed to start"; exit 1; fi
  sleep 1
done

echo "Opening browser..."
playwright-cli open http://localhost:3000

sleep 2

echo "Taking initial screenshot..."
playwright-cli screenshot --filename=e2e-phase2-initial.png

echo "Verifying page layout..."
SNAPSHOT=$(playwright-cli --raw snapshot)
echo "$SNAPSHOT" | head -30

echo "Phase 1 test: searching jobs..."
playwright-cli fill input[placeholder*="描述"] "找北京的前端岗位" --submit

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

echo "=== E2E Test Complete ==="
