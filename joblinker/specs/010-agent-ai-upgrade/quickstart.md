# Quickstart: AI Agent Capability Upgrade

**Feature**: specs/010-agent-ai-upgrade
**Date**: 2026-05-01

## E2E Test Scenarios

### Scenario 1: Three-Part Prompt Structure

**Prerequisites**: User logged in, AI agent responding

**Steps**:
1. Start new conversation with seeker agent
2. Ask: "Find me Python developer jobs"
3. Inspect AI response structure

**Expected result**: Response follows pattern:
- "I will search for Python jobs" (Must Do acknowledgment)
- No fabricated salary numbers (Must Not Do)
- Professional tone throughout (Behavior rules)

### Scenario 2: Function Calling - Job Query

**Prerequisites**: User logged in with active match

**Steps**:
1. Ask: "What jobs are available in London paying over $80k?"
2. Verify agent calls query_jobs API
3. Verify response includes actual job data from API

**Expected result**:
- Agent logs show tool_call: query_jobs
- Response contains real job titles, salaries from database
- Zero fabricated details (SC-002)

### Scenario 3: Function Calling - Candidate Profile

**Prerequisites**: Recruiter agent active, valid candidate in system

**Steps**:
1. Ask: "Tell me about candidate #123's Python skills"
2. Verify agent calls get_candidate API
3. Verify response references actual skills from profile

**Expected result**:
- Agent logs show tool_call: get_candidate
- Response includes verified skills, not hallucinated qualifications

### Scenario 4: Context Compression (Short-Term Memory)

**Prerequisites**: 20+ message conversation

**Steps**:
1. Send 20+ messages in conversation
2. Ask a question referencing early conversation (e.g., "what did I say about salary?")
3. Verify agent recalls correctly

**Expected result**:
- Context window shows compression occurred
- Agent recalls key facts from early conversation
- Token count reduced by 40%+ (SC-003)

### Scenario 5: Long-Term Memory Recall

**Prerequisites**: User with stored preferences (salary, location)

**Steps**:
1. Set preference in conversation: "I prefer remote work"
2. Start NEW conversation
3. Ask: "Show me jobs matching my preferences"

**Expected result**:
- Agent recalls remote work preference
- Returns only remote/hybrid positions
- 85% recall accuracy (SC-004)

### Scenario 6: Reasoning Flow (Understand → Query → Analyze → Respond)

**Prerequisites**: Complex job matching scenario

**Steps**:
1. Ask: "Find me a job where I can use Python and Django, near London, paying over $90k"
2. Observe agent reasoning steps

**Expected result**:
- Agent extracts: skills (Python, Django), location (London), salary ($90k+)
- Agent queries with all parameters
- Agent analyzes fit scores
- Agent responds with matched jobs and reasoning
- 90% of complex scenarios follow this flow (SC-005)

### Scenario 7: Multi-Agent Isolation

**Prerequisites**: Two concurrent matches

**Steps**:
1. On Match A conversation: "My favorite color is blue"
2. On Match B conversation: "What is my favorite color?"
3. Verify no cross-contamination

**Expected result**:
- Match A context does NOT leak to Match B
- Agent B responds that it doesn't know (different conversation)
- 100 concurrent agents work without leakage (SC-006)

### Scenario 8: Off-Topic Rejection

**Prerequisites**: User tries to go off-topic

**Steps**:
1. Ask: "Tell me a joke" (outside recruitment scope)
2. Verify agent redirects to relevant topics

**Expected result**:
- Agent refuses the off-topic request
- Agent explains its scope is recruitment only
- FR-017: Agent refuses off-topic conversations

### Scenario 9: Duplicate Message Prevention

**Prerequisites**: Network issues causing potential duplicate sends

**Steps**:
1. Send message with slow network
2. Resend same message quickly
3. Verify only one response

**Expected result**:
- Idempotency check prevents duplicate processing
- Only one AI response appears
- FR-020: No duplicate message processing

### Scenario 10: Error Handling - API Failure

**Prerequisites**: Simulate API failure

**Steps**:
1. Mock job API failure
2. Ask agent for job information
3. Verify graceful error handling

**Expected result**:
- Agent reports API unavailable
- Agent does NOT fabricate job data
- User sees friendly error message

## Manual Test Checklist

- [ ] Three-part prompt visible in system logs (Must Do / Must Not Do / Behavior sections)
- [ ] Function call logs show tool name and arguments
- [ ] Conversation summary preserves key facts (salary, location, skills)
- [ ] Long-term preferences persist across sessions
- [ ] Reasoning steps visible in agent logs
- [ ] Off-topic requests rejected politely
- [ ] Duplicate messages handled correctly
- [ ] API failures handled gracefully without fabrication

## Performance Benchmarks

| Metric | Target | Measurement |
|--------|--------|-------------|
| Response time (with function calling) | <5s | SC-007 |
| Memory recall accuracy | 85% | SC-004 |
| Context compression ratio | 40%+ | SC-003 |
| Concurrent conversations | 100 | SC-006 |
| Reasoning flow compliance | 90% | SC-005 |