#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-3: Auto Match ==="

SEEKER_TOKEN=$(cat /tmp/e2e-seeker-token.txt)
RECR_TOKEN=$(cat /tmp/e2e-recruiter-token.txt)
JOB_ID=$(cat /tmp/e2e-job-id.txt)
SEEKER_AGENT_ID=$(cat /tmp/e2e-seeker-agent-id.txt)

# Create auto match (returns {matches: [{id,score,...}], created_count, skipped_count})
MATCH_RESP=$(curl -s -X POST http://localhost:8080/api/matches/auto \
  -H "Authorization: Bearer $SEEKER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"job_ids\":[\"$JOB_ID\"]}")
MATCH_ID=$(echo "$MATCH_RESP" | jq -r '.matches[0].id // .id // empty')
MATCH_SCORE=$(echo "$MATCH_RESP" | jq -r '.matches[0].score // .score // "0"')
MATCH_STATUS=$(echo "$MATCH_RESP" | jq -r '.matches[0].status // .status // "unknown"')

echo "Match: $MATCH_ID (score=$MATCH_SCORE, status=$MATCH_STATUS)"

# Verify fields
if [ -z "$MATCH_ID" ] || [ "$MATCH_ID" = "null" ]; then
  echo "❌ No match ID returned"
  exit 1
fi
echo "✅ Match created: $MATCH_ID"

# Score should be dynamic (not 1.0 or hardcoded)
if [ "$MATCH_SCORE" = "1" ] || [ "$MATCH_SCORE" = "0.9" ] || [ "$MATCH_SCORE" = "1.0" ]; then
  echo "⚠️ Score is $MATCH_SCORE — may be hardcoded (seeker skills should partially match job skills)"
fi
assert_not_equal "$MATCH_SCORE" "0.9" "Match score (anti-hardcode)"

# Score should be in valid range
SCORE_INT=$(echo "$MATCH_SCORE" | jq ' . > 0 and . <= 1' 2>/dev/null || echo "false")
if [ "$SCORE_INT" != "true" ]; then
  echo "❌ Score $MATCH_SCORE out of range (0-1)"
  exit 1
fi
echo "✅ Score $MATCH_SCORE is in valid range"

# Verify FSM initial state via API
MATCH_GET=$(curl -s "http://localhost:8080/api/matches/$MATCH_ID" \
  -H "Authorization: Bearer $SEEKER_TOKEN")
CUR_STATUS=$(echo "$MATCH_GET" | jq -r '.status')
echo "Current FSM status: $CUR_STATUS"

echo "$MATCH_ID" > /tmp/e2e-match-id.txt

echo "=== TC-3 PASSED ==="
