#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-2: Agent + Job Creation ==="

SEEKER_TOKEN=$(cat /tmp/e2e-seeker-token.txt)
RECR_TOKEN=$(cat /tmp/e2e-recruiter-token.txt)

# Create seeker agent
SEEKER_AGENT=$(curl -s -X POST http://localhost:8080/api/agents \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"seeker","config":{"name":"E2E Seeker","skills":["Go","React","PostgreSQL"]}}')
SEEKER_AGENT_ID=$(echo "$SEEKER_AGENT" | jq -r '.id')
FSM_STATE=$(echo "$SEEKER_AGENT" | jq -r '.fsm_state')

if [ "$FSM_STATE" != "idle" ]; then
  echo "❌ Seeker agent fsm_state should be 'idle', got: $FSM_STATE"
  exit 1
fi
echo "✅ Seeker agent created: $SEEKER_AGENT_ID (fsm_state=$FSM_STATE)"

# Create recruiter agent
RECR_AGENT=$(curl -s -X POST http://localhost:8080/api/agents \
  -H "Authorization: Bearer $RECR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"recruiter","config":{"name":"E2E Recruiter","skills":["Go","Python"]}}')
RECR_AGENT_ID=$(echo "$RECR_AGENT" | jq -r '.id')
echo "✅ Recruiter agent created: $RECR_AGENT_ID"

# Create job
JOB_RESP=$(curl -s -X POST http://localhost:8080/api/jobs \
  -H "Authorization: Bearer $RECR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Senior Go Developer","location":"Remote","salary_min":80000,"salary_max":150000,"description":"Build the future of recruitment AI"}')
JOB_ID=$(echo "$JOB_RESP" | jq -r '.id')
echo "✅ Job created: $JOB_ID"

# Verify job salary is not hardcoded
JOB_SALARY=$(echo "$JOB_RESP" | jq -r '.salary_min // empty')
assert_no_hardcoded "$JOB_SALARY" "Job salary"

echo "$SEEKER_AGENT_ID" > /tmp/e2e-seeker-agent-id.txt
echo "$RECR_AGENT_ID" > /tmp/e2e-recruiter-agent-id.txt
echo "$JOB_ID" > /tmp/e2e-job-id.txt

echo "=== TC-2 PASSED ==="
