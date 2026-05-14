#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-1: Dual User Registration ==="

# Seeker registration
SEEKER_EMAIL=$(random_email)
echo "Registering seeker: $SEEKER_EMAIL"

playwright-cli -s=seeker open http://localhost:3000/register 2>/dev/null
playwright-cli -s=seeker snapshot 2>/dev/null
# Fill registration form (adjust element refs from snapshot)
playwright-cli -s=seeker fill e5 "$SEEKER_EMAIL" 2>/dev/null
playwright-cli -s=seeker fill e7 "Test1234!" 2>/dev/null
playwright-cli -s=seeker click e9 2>/dev/null  # "I'm a Job Seeker"
playwright-cli -s=seeker click e11 2>/dev/null # Submit
sleep 2
playwright-cli -s=seeker state-save seeker.json 2>/dev/null

SEEKER_PATH=$(playwright-cli -s=seeker eval "window.location.pathname" 2>/dev/null)
if echo "$SEEKER_PATH" | grep -q "/dashboard"; then
  echo "✅ Seeker registered and redirected to dashboard"
else
  echo "❌ Seeker registration failed, path: $SEEKER_PATH"
  exit 1
fi

# Recruiter registration
RECRUITER_EMAIL=$(random_email)
echo "Registering recruiter: $RECRUITER_EMAIL"

playwright-cli -s=recruiter open http://localhost:3000/register 2>/dev/null
playwright-cli -s=recruiter fill e5 "$RECRUITER_EMAIL" 2>/dev/null
playwright-cli -s=recruiter fill e7 "Test1234!" 2>/dev/null
playwright-cli -s=recruiter click e10 2>/dev/null  # "I'm a Recruiter"
playwright-cli -s=recruiter click e11 2>/dev/null
sleep 2
playwright-cli -s=recruiter state-save recruiter.json 2>/dev/null

RECR_PATH=$(playwright-cli -s=recruiter eval "window.location.pathname" 2>/dev/null)
if echo "$RECR_PATH" | grep -q "/dashboard"; then
  echo "✅ Recruiter registered and redirected to dashboard"
else
  echo "❌ Recruiter registration failed"
  exit 1
fi

# Verify tokens are different
SEEKER_TOKEN=$(playwright-cli -s=seeker --raw localstorage-get joblinker-auth 2>/dev/null | jq -r '.state.token')
RECR_TOKEN=$(playwright-cli -s=recruiter --raw localstorage-get joblinker-auth 2>/dev/null | jq -r '.state.token')

assert_not_equal "$SEEKER_TOKEN" "$RECR_TOKEN" "Seeker vs Recruiter tokens"
echo "✅ Tokens are different (tenant isolation works)"

# Save tokens for subsequent tests
echo "$SEEKER_TOKEN" > /tmp/e2e-seeker-token.txt
echo "$RECR_TOKEN" > /tmp/e2e-recruiter-token.txt
echo "$SEEKER_EMAIL" > /tmp/e2e-seeker-email.txt
echo "$RECRUITER_EMAIL" > /tmp/e2e-recruiter-email.txt

echo "=== TC-1 PASSED ==="
