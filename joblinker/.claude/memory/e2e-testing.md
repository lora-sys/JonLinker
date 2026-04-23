---
name: e2e-testing-preference
description: E2E tests use agent-browser and playwright MCP per project constitution
type: reference
---

**File**: `.specify/memory/constitution.md` (Appendix: E2E Testing)

**Rule**: All e2e testing for JobLinker MUST use:
1. **agent-browser** - Primary browser automation
2. **playwright MCP** - Next.js DevTools integration

**Why**: No X server/xvfb in this environment.

**Quick test**:
```bash
agent-browser open http://localhost:3000
agent-browser snapshot -i
# Navigate, fill forms, click buttons
agent-browser click @e3
agent-browser wait --url "**/dashboard"
```