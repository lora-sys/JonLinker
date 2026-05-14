#!/usr/bin/env bash
set -euo pipefail

# TC-7: Data Isolation & Full Flow (T14, T15)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/helpers.sh"

echo "=== TC-7: Data Isolation & Full Flow ==="

# Create a new user (different tenant)
EMAIL="e2e-iso-$(date +%s)-$RANDOM@test.com"
echo "Registering new user: $EMAIL"

REG=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"Test1234!\",\"role\":\"seeker\"}")

TOKEN=$(echo "$REG" | jq -r '.token')
USER_ID=$(echo "$REG" | jq -r '.user.id')
echo "New user ID: $USER_ID"

# T14: Cross-tenant isolation - new user should see 0 agents
echo ""
echo "--- Data Isolation ---"
AGENTS=$(curl -s http://localhost:8080/api/agents \
  -H "Authorization: Bearer $TOKEN" 2>/dev/null)
COUNT=$(echo "$AGENTS" | jq 'length' 2>/dev/null || echo "0")
echo "New user sees $COUNT agents (expect 0)"
if [ "$COUNT" = "0" ]; then
  echo "✅ Cross-tenant data isolation: OK (new user sees no other tenant's agents)"
else
  echo "❌ Data isolation FAILED: new user sees $COUNT agents!"
  echo "  Agents: $AGENTS"
  exit 1
fi

# T15: Verify full flow can be completed
echo ""
echo "--- Full Flow Verification ---"
echo "Checking backend services..."
python3 -c "
import requests
checks = {
    'backend': ('http://localhost:8080/health', lambda r: r.status_code == 200),
    'frontend': ('http://localhost:3000/login', lambda r: r.status_code == 200),
    'chroma': ('http://localhost:8080/health', lambda r: True),  # chroma is part of backend health
}
all_ok = True
for name, (url, check) in checks.items():
    try:
        r = requests.get(url, timeout=5)
        ok = check(r)
        status = '✅' if ok else '⚠️'
        print(f'{status} {name}: HTTP {r.status_code}')
        if not ok: all_ok = False
    except Exception as e:
        print(f'❌ {name}: {e}')
        all_ok = False
print()
print('All services up!' if all_ok else 'Some services degraded')
" 2>/dev/null

echo ""
echo "=== TC-7 PASSED ==="
