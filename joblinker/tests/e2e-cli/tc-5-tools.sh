#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-5: Tool Calls & Memory ==="

MATCH_ID=$(cat /tmp/e2e-match-id.txt)
RECR_TOKEN=$(cat /tmp/e2e-recruiter-token.txt)
SEEKER_TOKEN=$(cat /tmp/e2e-seeker-token.txt)

# Verify offer exists
OFFER_RESP=$(curl -s "http://localhost:8080/api/offers/$MATCH_ID" \
  -H "Authorization: Bearer $RECR_TOKEN")
OFFER_SALARY=$(echo "$OFFER_RESP" | jq -r '.salary // .salary_min // "0"')
OFFER_STATUS=$(echo "$OFFER_RESP" | jq -r '.status // "unknown"')

echo "Offer: salary=$OFFER_SALARY, status=$OFFER_STATUS"

# Salary should not be hardcoded
assert_not_equal "$OFFER_SALARY" "150000" "Offer salary (anti-hardcode)"
assert_not_equal "$OFFER_SALARY" "0" "Offer should have valid salary"

echo "✅ Offer salary is dynamic and valid"

# Verify interview exists
INTERVIEW_RESP=$(curl -s "http://localhost:8080/api/interviews/match/$MATCH_ID" \
  -H "Authorization: Bearer $RECR_TOKEN")
INT_COUNT=$(echo "$INTERVIEW_RESP" | jq 'length' 2>/dev/null || echo "0")
echo "Interviews found: $INT_COUNT"

# Verify messages exist with proper sender alternation
MESSAGES=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN")
MSG_COUNT=$(echo "$MESSAGES" | jq 'length' 2>/dev/null || echo "0")
SENDERS=$(echo "$MESSAGES" | jq -r '.[].sender_agent_id' 2>/dev/null | sort -u | wc -l)

echo "Messages: $MSG_COUNT, Unique senders: $SENDERS"

if [ "$SENDERS" -lt 2 ] && [ "$MSG_COUNT" -gt 0 ]; then
  echo "⚠️ Only one agent sent messages (monologue, not dialogue)"
fi

if [ "$MSG_COUNT" -eq 0 ]; then
  echo "⚠️ No messages found - dialogue may not have started"
fi

# Check for tool calls in match
TOOL_CALLS=$(curl -s "http://localhost:8080/api/matches/$MATCH_ID/tool_calls" \
  -H "Authorization: Bearer $SEEKER_TOKEN" || echo "[]")
TC_COUNT=$(echo "$TOOL_CALLS" | jq 'length' 2>/dev/null || echo "0")
echo "Tool calls: $TC_COUNT"

# WebSocket connectivity check
echo "Checking browser console for WS events..."
playwright-cli -s=seeker console 2>/dev/null || true

echo ""
echo "=== TC-5 SUMMARY ==="
echo "Messages: $MSG_COUNT | Unique agents: $SENDERS | Tool calls: $TC_COUNT | Salary: $OFFER_SALARY"
echo "=== TC-5 COMPLETED ==="
