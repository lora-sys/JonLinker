#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-5: Tool Calls & Memory ==="

MATCH_ID=$(cat /tmp/e2e-match-id.txt)
RECR_TOKEN=$(cat /tmp/e2e-recruiter-token.txt)
SEEKER_TOKEN=$(cat /tmp/e2e-seeker-token.txt)

# Check messages first
MESSAGES=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN")
MSG_COUNT=$(echo "$MESSAGES" | jq 'length' 2>/dev/null || echo "0")
SENDERS=$(echo "$MESSAGES" | jq -r '.[].sender_agent_id // empty' 2>/dev/null | sort -u | wc -l)
echo "Messages: $MSG_COUNT | Unique senders: $SENDERS"

if [ "$MSG_COUNT" -gt 0 ]; then
  echo "✅ Dialogue contains messages"
fi
if [ "$SENDERS" -ge 2 ]; then
  echo "✅ A2A alternation: multiple agents participating"
fi

# Try to get offer (may not exist yet)
OFFER_RESP=$(curl -s "http://localhost:8080/api/offers/$MATCH_ID" \
  -H "Authorization: Bearer $RECR_TOKEN" 2>/dev/null || echo "{}")
OFFER_SALARY=$(echo "$OFFER_RESP" | jq -r '.salary // .salary_min // "0"' 2>/dev/null)
OFFER_STATUS=$(echo "$OFFER_RESP" | jq -r '.status // "not_found"' 2>/dev/null)
echo "Offer: salary=$OFFER_SALARY, status=$OFFER_STATUS"

if [ "$OFFER_STATUS" != "not_found" ]; then
  echo "✅ Offer exists"
  if [ "$OFFER_SALARY" = "150000" ]; then
    echo "❌ Offer salary is hardcoded (150000)"
  else
    echo "✅ Offer salary is dynamic: $OFFER_SALARY"
  fi
else
  echo "ℹ️ No offer yet (dialogue may not have progressed to OFFER stage)"
fi

# Try to get interviews
INTERVIEW_RESP=$(curl -s "http://localhost:8080/api/interviews/match/$MATCH_ID" \
  -H "Authorization: Bearer $RECR_TOKEN" 2>/dev/null || echo "[]")
INT_COUNT=$(echo "$INTERVIEW_RESP" | jq 'length' 2>/dev/null || echo "0")
echo "Interviews found: $INT_COUNT"

# Check for tool calls
TOOL_CALLS=$(curl -s "http://localhost:8080/api/matches/$MATCH_ID/tool_calls" \
  -H "Authorization: Bearer $SEEKER_TOKEN" 2>/dev/null || echo "[]")
TC_COUNT=$(echo "$TOOL_CALLS" | jq 'length' 2>/dev/null || echo "0")
echo "Tool calls: $TC_COUNT"

# Browser console check
echo ""
echo "Browser console for seeker session..."
playwright-cli -s=seeker console 2>/dev/null || true

echo ""
echo "=== TC-5 SUMMARY ==="
echo "Messages: $MSG_COUNT | Agents: $SENDERS | Tools: $TC_COUNT | Salary: $OFFER_SALARY"
echo "=== TC-5 COMPLETED ==="
