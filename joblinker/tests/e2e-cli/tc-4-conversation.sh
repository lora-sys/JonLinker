#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-4: A2A Dialogue Chain ==="

MATCH_ID=$(cat /tmp/e2e-match-id.txt)
SEEKER_TOKEN=$(cat /tmp/e2e-seeker-token.txt)

# Open conversation page in seeker's browser
playwright-cli -s=seeker state-load seeker.json 2>/dev/null
playwright-cli -s=seeker goto "http://localhost:3000/conversation/$MATCH_ID" 2>/dev/null
playwright-cli -s=seeker snapshot --filename=conv-start.yaml 2>/dev/null

# Wait for full A2A chain
echo "Polling for A2A dialogue chain..."
FULL_CHAIN=""
for i in $(seq 1 30); do
  INTENTS=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
    -H "Authorization: Bearer $SEEKER_TOKEN" \
    | jq -r '.[].intent_type' 2>/dev/null | tr '\n' ' → ')
  echo "[$i] $INTENTS"

  # Check for complete chain
  if echo "$INTENTS" | grep -q "OFFER"; then
    FULL_CHAIN="$INTENTS"
    echo "✅ Full A2A chain detected!"
    break
  fi
  sleep 5
done

if [ -z "$FULL_CHAIN" ]; then
  echo "❌ A2A chain incomplete (no OFFER detected)"
  exit 1
fi

# Verify messages have valid XML content
MSG_XML=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  | jq -r '.[0].content_xml' 2>/dev/null)
if echo "$MSG_XML" | grep -q "<message>"; then
  echo "✅ Messages contain valid XML"
else
  echo "❌ Messages missing XML format"
  exit 1
fi

# Verify no hardcoded responses
ALL_RESPONSES=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  | jq -r '.[].content_xml' 2>/dev/null)
assert_no_hardcoded "$ALL_RESPONSES" "AI responses"

echo "=== TC-4 PASSED ==="
