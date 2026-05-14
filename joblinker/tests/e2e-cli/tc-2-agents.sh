#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-2: Agent + Job Creation ==="

SEEKER_TOKEN=$(cat /tmp/e2e-seeker-token.txt)
RECR_TOKEN=$(cat /tmp/e2e-recruiter-token.txt)

# Create seeker agent (config must be JSON-encoded string)
SEEKER_AGENT=$(curl -s -X POST http://localhost:8080/api/agents \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"seeker","config":"{\"name\":\"E2E Seeker\",\"skills\":[\"Go\",\"React\",\"PostgreSQL\"],\"location\":\"Remote\"}"}')
SEEKER_AGENT_ID=$(echo "$SEEKER_AGENT" | jq -r '.id')
FSM_STATE=$(echo "$SEEKER_AGENT" | jq -r '.fsm_state')

if [ "$FSM_STATE" != "idle" ]; then
  echo "❌ Seeker agent fsm_state should be 'idle', got: $FSM_STATE"
  echo "  Response: $SEEKER_AGENT"
  exit 1
fi
echo "✅ Seeker agent created: $SEEKER_AGENT_ID (fsm_state=$FSM_STATE)"

# Create recruiter agent
RECR_AGENT=$(curl -s -X POST http://localhost:8080/api/agents \
  -H "Authorization: Bearer $RECR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"recruiter","config":"{\"name\":\"E2E Recruiter\",\"skills\":[\"Go\",\"Python\"]}"}')
RECR_AGENT_ID=$(echo "$RECR_AGENT" | jq -r '.id')
if [ "$RECR_AGENT_ID" = "null" ] || [ -z "$RECR_AGENT_ID" ]; then
  echo "❌ Recruiter agent creation failed: $RECR_AGENT"
  exit 1
fi
echo "✅ Recruiter agent created: $RECR_AGENT_ID"

# Create job (structured must be JSON-encoded string)
JOB_RESP=$(python3 -c "
import requests, json
rtok = open('/tmp/e2e-recruiter-token.txt').read().strip()
inner = json.dumps({'title':'Senior Go Developer','location':'Remote','salary_min':80000,'salary_max':150000,'skills':['Go','React','PostgreSQL']})
r = requests.post('http://localhost:8080/api/jobs', headers={'Authorization': f'Bearer {rtok}', 'Content-Type': 'application/json'}, json={'structured': inner})
data = r.json()
print(json.dumps(data))
")
JOB_ID=$(echo "$JOB_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null)
if [ -z "$JOB_ID" ] || [ "$JOB_ID" = "null" ]; then
  echo "❌ Job creation failed: $JOB_RESP"
  exit 1
fi
echo "✅ Job created: $JOB_ID"

echo "$SEEKER_AGENT_ID" > /tmp/e2e-seeker-agent-id.txt
echo "$RECR_AGENT_ID" > /tmp/e2e-recruiter-agent-id.txt
echo "$JOB_ID" > /tmp/e2e-job-id.txt

echo "=== TC-2 PASSED ==="
