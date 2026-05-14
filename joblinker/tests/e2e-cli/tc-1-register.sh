#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-1: Dual User Registration ==="

# Seeker registration
SEEKER_EMAIL=$(random_email)
echo "Registering seeker: $SEEKER_EMAIL"

playwright-cli -s=seeker open http://localhost:3000/register 2>/dev/null
sleep 2
playwright-cli -s=seeker fill "#email" "$SEEKER_EMAIL" 2>/dev/null
playwright-cli -s=seeker fill "#password" "Test1234!" 2>/dev/null
playwright-cli -s=seeker fill "#confirmPassword" "Test1234!" 2>/dev/null
# Click "Job Seeker" button (already default)
playwright-cli -s=seeker click "getByText('Create account')" 2>/dev/null
sleep 3

SEEKER_PATH=$(playwright-cli -s=seeker eval "window.location.pathname" 2>/dev/null)
if echo "$SEEKER_PATH" | grep -q "/dashboard"; then
  echo "✅ Seeker registered and redirected to dashboard"
else
  echo "❌ Seeker registration failed, path: $SEEKER_PATH"
  playwright-cli -s=seeker close 2>/dev/null
  exit 1
fi

# Recruiter registration
RECRUITER_EMAIL=$(random_email)
echo "Registering recruiter: $RECRUITER_EMAIL"

playwright-cli -s=recruiter open http://localhost:3000/register 2>/dev/null
sleep 2
playwright-cli -s=recruiter fill "#email" "$RECRUITER_EMAIL" 2>/dev/null
playwright-cli -s=recruiter fill "#password" "Test1234!" 2>/dev/null
playwright-cli -s=recruiter fill "#confirmPassword" "Test1234!" 2>/dev/null
playwright-cli -s=recruiter click "getByText('Recruiter')" 2>/dev/null
sleep 1
playwright-cli -s=recruiter click "getByText('Create account')" 2>/dev/null
sleep 3

RECR_PATH=$(playwright-cli -s=recruiter eval "window.location.pathname" 2>/dev/null)
if echo "$RECR_PATH" | grep -q "/dashboard"; then
  echo "✅ Recruiter registered and redirected to dashboard"
else
  echo "❌ Recruiter registration failed"
  playwright-cli -s=recruiter close 2>/dev/null
  exit 1
fi

# Verify tokens are different
SEEKER_TOKEN=$(playwright-cli -s=seeker --raw eval "JSON.parse(localStorage.getItem('joblinker-auth')).state.token" 2>/dev/null | tr -d '"')
RECR_TOKEN=$(playwright-cli -s=recruiter --raw eval "JSON.parse(localStorage.getItem('joblinker-auth')).state.token" 2>/dev/null | tr -d '"')

assert_not_equal "$SEEKER_TOKEN" "$RECR_TOKEN" "Seeker vs Recruiter tokens"
echo "✅ Tokens are different (tenant isolation)"

# Save tokens for subsequent tests
echo "$SEEKER_TOKEN" > /tmp/e2e-seeker-token.txt
echo "$RECR_TOKEN" > /tmp/e2e-recruiter-token.txt
echo "$SEEKER_EMAIL" > /tmp/e2e-seeker-email.txt
echo "$RECRUITER_EMAIL" > /tmp/e2e-recruiter-email.txt

# Save browser state for subsequent tests
playwright-cli -s=seeker state-save seeker.json 2>/dev/null
playwright-cli -s=recruiter state-save recruiter.json 2>/dev/null

echo "=== TC-1 PASSED ==="
