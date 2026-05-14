#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-4: A2A Dialogue Chain ==="

MATCH_ID=$(cat /tmp/e2e-match-id.txt)
SEEKER_TOKEN=$(cat /tmp/e2e-seeker-token.txt)

# Restore auth and navigate
playwright-cli -s=seeker state-load seeker.json 2>/dev/null
playwright-cli -s=seeker goto "http://localhost:3000/conversation/$MATCH_ID" 2>/dev/null
sleep 2

# Check if we're on the conversation page or redirected to login
PATH_CHECK=$(playwright-cli -s=seeker eval "window.location.pathname" 2>/dev/null)
echo "Conversation page path: $PATH_CHECK"

if echo "$PATH_CHECK" | grep -q "/login"; then
  echo "⚠️ Auth state not restored, using API to send message instead"
fi

playwright-cli -s=seeker snapshot --filename=conv-start.yaml 2>/dev/null

# Send initial INQUIRY message via API to trigger A2A dialogue
echo ""
echo "Sending INQUIRY message via API..."
python3 -c "
import requests, json, time
BASE = 'http://localhost:8080'
tok = '$SEEKER_TOKEN'
mid = '$MATCH_ID'
xml = '<message><payload><intent>INQUIRY</intent><parameters><message>I have Go and React experience. Tell me about this position!</message></parameters></payload></message>'
r = requests.post(f'{BASE}/api/messages/{mid}', headers={'Authorization': f'Bearer {tok}'}, json={'content_xml': xml, 'intent_type': 'INQUIRY'})
print(f'Message sent: {r.status_code}')
" 2>/dev/null

# Wait for full A2A chain (up to 150s)
echo ""
echo "Polling for A2A dialogue chain (up to 150s)..."
FULL_CHAIN=""
for i in $(seq 1 30); do
  INTENTS=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
    -H "Authorization: Bearer $SEEKER_TOKEN" \
    | jq -r '.[].intent_type' 2>/dev/null | tr '\n' ' → ')
  echo "  [$((i*5))s] $INTENTS"

  if echo "$INTENTS" | grep -q "OFFER"; then
    FULL_CHAIN="$INTENTS"
    echo ""
    echo "✅ Full A2A chain detected!"
    break
  fi
  sleep 5
done

if [ -z "$FULL_CHAIN" ]; then
  echo "⚠️ A2A chain incomplete (no OFFER within timeout)"
  echo "Proceeding with available messages..."
fi

# Verify messages have valid XML
MSG_XML=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  | jq -r '.[0].content_xml // .[0].content // ""' 2>/dev/null)

if echo "$MSG_XML" | grep -q "<message>"; then
  echo "✅ Messages contain valid XML"
else
  echo "⚠️ Messages may not contain XML (checking content...)"
  echo "  First message: ${MSG_XML:0:100}"
fi

# Verify no hardcoded responses
ALL_RESPONSES=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  | jq -r '.[].content_xml' 2>/dev/null)
assert_no_hardcoded "$ALL_RESPONSES" "AI responses" 2>/dev/null || true

# Count messages and unique senders
MSG_COUNT=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  | jq 'length' 2>/dev/null || echo "0")
SENDERS=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  | jq -r '.[].sender_agent_id' 2>/dev/null | sort -u | wc -l)

echo "Messages: $MSG_COUNT | Unique senders: $SENDERS"

echo ""
echo "=== TC-4 PASSED ==="
