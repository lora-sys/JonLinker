#!/usr/bin/env bash
set -euo pipefail

echo "=========================================="
echo "  JobLinker E2E Acceptance Tests"
echo "  $(date)"
echo "=========================================="
echo ""

# Verify prerequisites
command -v playwright-cli >/dev/null 2>&1 || { echo "❌ playwright-cli not found"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "❌ curl not found"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "❌ jq not found"; exit 1; }
echo "✅ Prerequisites met"
echo ""

# Ensure clean state
playwright-cli close-all 2>/dev/null || true
rm -f /tmp/e2e-seeker-token.txt /tmp/e2e-recruiter-token.txt /tmp/e2e-*.txt

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FAILED=0

run_test() {
  local name=$1
  local script=$2
  echo ""
  echo "--- Running $name ---"
  if bash "$script"; then
    echo ""
    echo "✅ $name PASSED"
  else
    echo ""
    echo "❌ $name FAILED"
    FAILED=$((FAILED + 1))
  fi
  echo "------------------------"
}

run_test "TC-1: User Registration" "$SCRIPT_DIR/tc-1-register.sh"
run_test "TC-2: Agent & Job Creation" "$SCRIPT_DIR/tc-2-agents.sh"
run_test "TC-3: Auto Match" "$SCRIPT_DIR/tc-3-match.sh"
run_test "TC-4: A2A Dialogue" "$SCRIPT_DIR/tc-4-conversation.sh"
run_test "TC-5: Tools & Memory" "$SCRIPT_DIR/tc-5-tools.sh"

# Cleanup
playwright-cli close-all 2>/dev/null || true

echo ""
echo "=========================================="
if [ "$FAILED" -eq 0 ]; then
  echo "  ALL TESTS PASSED ✅"
  exit 0
else
  echo "  $FAILED TEST(S) FAILED ❌"
  exit 1
fi
echo "=========================================="
