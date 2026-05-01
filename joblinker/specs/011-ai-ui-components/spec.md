# Feature Specification: AI UI Components Upgrade

**Feature Branch**: `011-ai-ui-components`
**Created**: 2026-04-30
**Status**: Draft
**Input**: User description: "优化前端AI组件，使用Vercel AI SDK (ai-sdk.dev) 重构聊天界面，实现流式对话、工具调用显示、消息状态管理、Thread管理"

## User Scenarios & Testing

### User Story 1 - Streamlined Chat Interface (Priority: P1)

As a user, I want to chat with AI agents through a modern streaming interface so conversations feel responsive and natural.

**Why this priority**: Core UI - all user interactions depend on this.

**Independent Test**: Can be tested by opening conversation page and sending messages.

**Acceptance Scenarios**:

1. **Given** I am on a conversation page, **When** I send a message, **Then** I see streaming text appear character-by-character
2. **Given** an AI is processing my message, **When** it calls a tool, **Then** I see a "calling tool: X" indicator
3. **Given** the AI sends multiple messages in sequence, **When** they arrive, **Then** they appear in order without duplication
4. **Given** I send a message with attachments, **When** the AI processes it, **Then** attachments are properly displayed

---

### User Story 2 - Message State Management (Priority: P1)

As a developer, I need consistent message types and state so the UI correctly displays all conversation stages.

**Why this priority**: Foundation for all AI UI features.

**Independent Test**: Can be tested by verifying all message types render correctly.

**Acceptance Scenarios**:

1. **Given** a user message arrives, **When** it displays, **Then** it shows user role and content clearly
2. **Given** an assistant message arrives, **When** it displays, **Then** it shows assistant role with streaming animation
3. **Given** a tool call is in progress, **When** the message displays, **Then** it shows tool name and pending state
4. **Given** a tool result returns, **When** the message displays, **Then** it shows the result inline

---

### User Story 3 - Tool Call Display (Priority: P2)

As a user, I want to see when the AI is using tools so I understand what actions are being taken.

**Why this priority**: Transparency builds trust in AI actions.

**Independent Test**: Can be tested by triggering tool-calling scenarios.

**Acceptance Scenarios**:

1. **Given** the AI needs to look up a job, **When** it calls the job query tool, **Then** I see "Searching for jobs..." indicator
2. **Given** the AI is creating an offer, **When** it calls the offer API, **Then** I see "Creating offer..." indicator
3. **Given** tool execution completes, **When** the AI responds, **Then** the tool result is incorporated into the response

---

### User Story 4 - Conversation Thread Management (Priority: P2)

As a user, I need organized conversation threads so I can track multiple job applications separately.

**Why this priority**: Essential for managing multiple concurrent recruitment conversations.

**Independent Test**: Can be tested by creating multiple conversations and switching between them.

**Acceptance Scenarios**:

1. **Given** I have multiple matches, **When** I select one, **Then** I see only that conversation's messages
2. **Given** I switch between conversations, **When** I return, **Then** the original conversation state is preserved
3. **Given** a conversation has many messages, **When** I scroll up, **Then** older messages load progressively

---

### User Story 5 - AI SDK Integration with Backend (Priority: P2)

As a developer, I need to integrate Vercel AI SDK with existing WebSocket backend so AI responses flow through the established infrastructure.

**Why this priority**: Must use existing RabbitMQ/WebSocket infrastructure, not bypass it.

**Independent Test**: Can be tested by verifying AI responses flow through existing message APIs.

**Acceptance Scenarios**:

1. **Given** a user sends a message, **When** the AI responds, **Then** the response is stored via existing message API
2. **Given** an AI needs to call a tool, **When** it executes, **Then** the tool call goes through the existing internal APIs
3. **Given** WebSocket connection drops, **When** it reconnects, **Then** message continuity is maintained

---

### User Story 6 - E2E AI-Driven Testing (Priority: P1)

As a QA engineer, I need automated tests that verify AI-driven flows so I can validate the system end-to-end.

**Why this priority**: Ensures AI behavior is tested automatically, not just manually.

**Independent Test**: Can be tested by running the test suite.

**Acceptance Scenarios**:

1. **Given** the AI conversation flow, **When** tests run, **Then** they verify the AI calls tools correctly
2. **Given** the memory recall feature, **When** tests run, **Then** they verify preferences are remembered across sessions
3. **Given** the reasoning process, **When** tests run, **Then** they verify the AI follows the understand→query→analyze→respond flow

---

### Edge Cases

- What happens when WebSocket disconnects mid-stream?
- How does the UI handle malformed AI responses?
- What happens when tool call fails?
- How does the system handle very long messages (>1000 tokens)?
- What happens when concurrent messages arrive out of order?

## Requirements

### Functional Requirements

- **FR-001**: System MUST use Vercel AI SDK `useChat` hook for chat interface
- **FR-002**: System MUST implement streaming text display with character-by-character animation
- **FR-003**: System MUST display message roles (user/assistant/system) with distinct styling
- **FR-004**: System MUST show tool call indicators when AI invokes tools
- **FR-005**: System MUST display tool results inline after tool execution
- **FR-006**: System MUST support file attachments (images, PDFs) in messages
- **FR-007**: System MUST maintain conversation thread isolation (no cross-talk)
- **FR-008**: System MUST preserve conversation state across page navigation
- **FR-009**: System MUST integrate with existing WebSocket infrastructure for real-time updates
- **FR-010**: System MUST route AI tool calls through existing internal APIs (job query, offer creation, etc.)
- **FR-011**: System MUST implement proper error handling for failed tool calls
- **FR-012**: System MUST support progressive loading of message history
- **FR-013**: All existing chat UI functionality MUST remain intact (message display, send button, input field)
- **FR-014**: System MUST include E2E tests for AI-driven flows using Playwright

### Key Entities

- **ChatMessage**: Message with id, role, content, toolInvocations, attachments
- **ToolCall**: Tool invocation state with name, args, result
- **ConversationThread**: Isolated message history for a single match/conversation
- **AIStreamState**: Streaming text state with partial content

## Success Criteria

### Measurable Outcomes

- **SC-001**: Chat interface loads and displays messages within 1 second
- **SC-002**: Streaming text renders at minimum 30 characters per second
- **SC-003**: Tool call indicators appear within 500ms of tool invocation
- **SC-004**: Zero duplicate messages in conversation history
- **SC-005**: Conversation switch restores correct state in under 500ms
- **SC-006**: All E2E AI tests pass (tool calls, memory recall, reasoning flow)
- **SC-007**: Backward compatibility: existing manual conversation tests pass without modification

## Assumptions

- Vercel AI SDK is compatible with existing Next.js 16 setup
- WebSocket infrastructure can handle increased message volume from streaming
- Existing internal APIs (job query, offer creation) support tool calling pattern
- AI model (configured in backend) supports tool calling / function calling
- Playwright is available for E2E testing
- Token limits are sufficient for conversation context
