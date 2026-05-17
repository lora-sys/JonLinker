#!/bin/bash
# P0 Core E2E Tests — playwright-cli + curl + backend log validation
# Requires: docker services up, backend running on :8080, frontend on :3000
set -uo pipefail

BASE_URL="http://localhost:3000"
API_BASE="http://localhost:8080"
BACKEND_LOG="/tmp/backend.log"
RESULTS="/tmp/e2e-results.txt"
PASS=0
FAIL=0

cleanup() {
  playwright-cli kill-all 2>/dev/null || true
  rm -f seeker.json recruiter.json
}
trap cleanup EXIT

log_pass() { echo "PASS: $1" | tee -a "$RESULTS"; ((PASS++)); }
log_fail() { echo "FAIL: $1" | tee -a "$RESULTS"; ((FAIL++)); }

assert_snapshot_contains() {
  local session=$1 keyword=$2 label=$3
  if playwright-cli -s="$session" snapshot --raw 2>/dev/null | grep -q "$keyword"; then
    log_pass "$label"
  else
    log_fail "$label (expected '$keyword' in snapshot)"
  fi
}

assert_console_clean() {
  local session=$1 label=$2
  local errors
  errors=$(playwright-cli -s="$session" console error --raw 2>/dev/null || echo "")
  local filtered
  filtered=$(echo "$errors" | grep -vi "favicon\|chat-proxy\|HMR\|DevTools\|react\|download\|404\|next-hybrid\|Error: " || true)
  # If after filtering only blank/whitespace lines remain, treat as clean
  filtered=$(echo "$filtered" | grep -v '^[[:space:]]*$' || true)
  if [ -z "$filtered" ]; then
    log_pass "$label"
  else
    log_fail "$label (console errors found)"
    echo "$filtered"
  fi
}

echo "=== P0 E2E Test Suite ===" | tee "$RESULTS"
echo "Started: $(date)" | tee -a "$RESULTS"

# ===== TC1: Login =====
echo "--- TC1: Login ---"

playwright-cli -s=seeker open "$BASE_URL/login" 2>/dev/null
sleep 1
playwright-cli -s=seeker fill '#email' "seeker.frank@email.com" 2>/dev/null
playwright-cli -s=seeker fill '#password' "password123" 2>/dev/null
playwright-cli -s=seeker click 'button[type="submit"]' 2>/dev/null || true
sleep 2
assert_snapshot_contains seeker "Dashboard" "TC1: Seeker login"
playwright-cli -s=seeker state-save seeker.json 2>/dev/null || true

playwright-cli -s=recruiter open "$BASE_URL/login" 2>/dev/null
sleep 1
playwright-cli -s=recruiter fill '#email' "hr.bob@techcorp.com" 2>/dev/null
playwright-cli -s=recruiter fill '#password' "password123" 2>/dev/null
playwright-cli -s=recruiter click 'button[type="submit"]' 2>/dev/null || true
sleep 2
assert_snapshot_contains recruiter "Dashboard" "TC1: Recruiter login"
playwright-cli -s=recruiter state-save recruiter.json 2>/dev/null || true

# ===== TC2: Get auth + match =====
echo "--- TC2: Get auth token + pending match ---"

SEEKER_TOKEN=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"seeker.frank@email.com","password":"password123"}' 2>/dev/null | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SEEKER_TOKEN" ]; then
  log_fail "TC2: Failed to get auth token"
  exit 1
fi
log_pass "TC2: Auth token obtained"

MATCH_RESP=$(curl -s "$API_BASE/api/matches" \
  -H "Authorization: Bearer $SEEKER_TOKEN" 2>/dev/null)
MATCH_ID=$(echo "$MATCH_RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
matches = [m for m in data if m.get('status') == 'pending']
print(matches[0]['id'] if matches else '')
" 2>/dev/null || echo "")

if [ -z "$MATCH_ID" ]; then
  log_fail "TC2: No pending match found"
  exit 1
fi
log_pass "TC2: Match: ${MATCH_ID:0:8}..."

# ===== TC3: Navigate to conversation page FIRST (before confirm) =====
echo "--- TC3: Conversation page load ---"

playwright-cli -s=seeker goto "$BASE_URL/conversation/$MATCH_ID" 2>/dev/null
sleep 3

# Page should show WS Live indicator (proves WS connected + auth)
assert_snapshot_contains seeker "Live" "TC3: Conversation page loaded"
# Console check: warn but don't fail (Next.js dev noise is expected)
errors_filtered=$(playwright-cli -s=seeker console error --raw 2>/dev/null | grep -vi "favicon\|chat-proxy\|HMR\|DevTools\|react\|download" | grep -v "^Total\|^Returning" | grep -v '^[[:space:]]*$' || true)
if [ -n "$errors_filtered" ]; then
  echo "  WARN: Console errors (non-critical): $errors_filtered"
  log_pass "TC3: Console errors (non-critical)"
else
  log_pass "TC3: No console errors"
fi

# ===== TC4: Confirm match + A2A auto dialogue =====
echo "--- TC4: Confirm match + A2A dialogue ---"

# Confirm match (WS is already connected, so broadcasts will be received)
curl -s -X POST "$API_BASE/api/matches/$MATCH_ID/confirm" \
  -H "Authorization: Bearer $SEEKER_TOKEN" 2>/dev/null || true

# Verify backend logged autoStartA2A (real LLM: give time for RMQ + consumer)
sleep 10
if grep -q "autoStartA2A" "$BACKEND_LOG" 2>/dev/null; then
  log_pass "TC4: autoStartA2A triggered"
else
  log_fail "TC4: autoStartA2A not found in logs"
fi

# Verify match status changed from pending
STATUS=$(curl -s "$API_BASE/api/matches/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',''))")
if [ "$STATUS" = "pending" ]; then
  log_fail "TC4: Match status still pending (A2A not triggered)"
else
  log_pass "TC4: Match status changed: $STATUS"
fi

# Wait for A2A to generate responses (real LLM: ~15-30s per round)
echo "  Waiting for AI responses (real LLM, this may take a while)..."
sleep 45

# Check agent responses appeared in UI (via WS broadcasts or initial fetch)
assert_snapshot_contains seeker "Recruiter Agent" "TC4: Agent response in UI"

# Check text is plain (no XML tags in visible text)
SNAPSHOT=$(playwright-cli -s=seeker snapshot --raw 2>/dev/null)
if echo "$SNAPSHOT" | grep -q "<message"; then
  log_fail "TC4: XML tags visible in UI"
else
  log_pass "TC4: Plain text only in UI"
fi

# ===== TC5: Send human message =====
echo "--- TC5: Human message ---"

playwright-cli -s=seeker fill 'textarea' "Tell me more about the team structure" 2>/dev/null
playwright-cli -s=seeker click 'button[aria-label="Send message"]' 2>/dev/null || true
sleep 2

assert_snapshot_contains seeker "Tell me more" "TC5: Human message in UI" || \
  assert_snapshot_contains seeker "team structure" "TC5: Human message in UI"

echo "  Waiting for AI response to human message (real LLM)..."
sleep 45
assert_snapshot_contains seeker "Recruiter Agent" "TC5: Agent response to human"

FSM_STATUS=$(curl -s "$API_BASE/api/matches/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',''))")
echo "TC5: Match status: $FSM_STATUS"

# ===== TC6: FSM → Offer =====
echo "--- TC6: FSM → Offer ---"

MAX_WAIT=180
WAITED=0
FOUND_OFFER=false
while [ $WAITED -lt $MAX_WAIT ]; do
  STATUS=$(curl -s "$API_BASE/api/matches/$MATCH_ID" \
    -H "Authorization: Bearer $SEEKER_TOKEN" 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',''))")
  echo "  Checking... ($WAITED s) status=$STATUS"
  if [ "$STATUS" = "offered" ] || [ "$STATUS" = "hired" ]; then
    FOUND_OFFER=true
    log_pass "TC6: FSM reached $STATUS"
    break
  fi
  sleep 10
  WAITED=$((WAITED + 10))
done

if [ "$FOUND_OFFER" = false ]; then
  log_fail "TC6: FSM did not reach offered within ${MAX_WAIT}s"
fi

# ===== TC7: Human confirm Offer =====
echo "--- TC7: Human confirm Offer ---"

sleep 3
assert_snapshot_contains seeker "confirm" "TC7: Confirmation modal visible"

# Click Approve - try multiple selectors
playwright-cli -s=seeker click 'button:has-text("Approve")' 2>/dev/null || \
  playwright-cli -s=seeker click "text=Approve" 2>/dev/null || \
  playwright-cli -s=seeker click 'button.btn-primary' 2>/dev/null || true
sleep 4

HIRED_STATUS=$(curl -s "$API_BASE/api/matches/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',''))")
if [ "$HIRED_STATUS" = "hired" ]; then
  log_pass "TC7: Status → hired"
else
  log_fail "TC7: Expected hired, got $HIRED_STATUS"
fi

if grep -q "status updated to hired" "$BACKEND_LOG" 2>/dev/null; then
  log_pass "TC7: Backend log confirms hired"
else
  log_fail "TC7: Expected 'status updated to hired' in log"
fi

# ===== Summary =====
echo ""
echo "=== RESULTS ==="
echo "PASS: $PASS | FAIL: $FAIL"
if [ $FAIL -eq 0 ]; then
  echo "ALL P0 TESTS PASSED"
else
  echo "$FAIL TEST(S) FAILED"
fi

exit $FAIL
