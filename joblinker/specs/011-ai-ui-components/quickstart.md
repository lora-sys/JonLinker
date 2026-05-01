# Quickstart: AI UI Components

**Feature**: specs/011-ai-ui-components
**Date**: 2026-04-30

## E2E Test Scenarios

### Scenario 1: Streaming Chat Flow

**Prerequisites**: User logged in, has an active match

**Steps**:
1. Navigate to `/conversation/[matchId]`
2. Type "What jobs are available for Python developers?"
3. Verify streaming text appears character-by-character
4. Verify message appears in message list with user role
5. Verify AI response includes actual job data

**Expected result**: AI response streams at 30+ chars/sec, includes verified job details

### Scenario 2: Tool Call Display

**Prerequisites**: User on conversation page with active match

**Steps**:
1. Ask: "Find me React developer jobs in London"
2. Observe tool call indicator appears
3. Verify indicator shows "Searching for jobs..." or similar
4. Wait for AI to incorporate results into response

**Expected result**: Tool call indicator visible within 500ms of tool invocation

### Scenario 3: Thread Isolation

**Prerequisites**: User has 2+ active matches

**Steps**:
1. On Match A conversation, send "Remember my favorite color is blue"
2. Navigate to Match B conversation
3. Send "What's my favorite color?"
4. Verify AI does NOT know the answer (different thread)

**Expected result**: No context leakage between threads

### Scenario 4: Conversation Persistence

**Prerequisites**: User on conversation page

**Steps**:
1. Send 3-5 messages
2. Navigate away to dashboard
3. Return to conversation page
4. Verify all messages still visible in correct order

**Expected result**: Thread state restored from localStorage/API

### Scenario 5: WebSocket Reconnection

**Prerequisites**: User on conversation page

**Steps**:
1. Start a conversation with the AI
2. Simulate network drop (dev tools → offline)
3. Reconnect network
4. Continue sending messages

**Expected result**: Messages continue without duplication, error recovery works

## Manual Test Checklist

- [ ] Chat input accepts text and sends on Enter
- [ ] Send button works and is disabled during streaming
- [ ] Messages display with correct role styling (user right/assistant left)
- [ ] Streaming text animates smoothly (no flickering)
- [ ] Tool call indicator shows spinner + tool name
- [ ] Tool result appears inline after execution
- [ ] Attachment (image) displays correctly in message
- [ ] Long messages scroll properly in message list
- [ ] Mobile layout works on 375px width
- [ ] Keyboard navigation works (Tab through messages)

## Performance Benchmarks

| Metric | Target | Measurement |
|--------|--------|-------------|
| Time to first streaming char | <500ms | From send to first char |
| Streaming speed | 30+ chars/sec | Measured over 2 second window |
| Thread switch | <500ms | Navigate between matches |
| Message render | <100ms | Per message batch |