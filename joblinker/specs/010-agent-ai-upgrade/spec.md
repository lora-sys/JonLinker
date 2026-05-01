# Feature Specification: AI Agent Capability Upgrade

**Feature Branch**: `010-agent-ai-upgrade`
**Created**: 2026-04-30
**Updated**: 2026-05-01
**Status**: Draft
**Input**: AI Agent 整体能力全面升级 + 全自动 A2A 自治模式

## System Core Positioning: Full Auto A2A Autonomous Mode

**CRITICAL UPDATE**: This system operates as a fully autonomous A2A recruitment platform where AI Agents are the primary business actors. Human involvement is limited to final key result confirmation only.

### Core Principles

1. **Agent-as-Business-Entity**: The system treats AI Agents as the primary business actors. Humans do NOT participate in or干预 intermediate dialogue and negotiation processes. The entire链路 communication and process progression between求职Agent and招聘Agent proceeds autonomously.

2. **Dual-Agent Fully Autonomous Dialogue**: The求职Agent and 招聘Agent engage in completely self-initiated conversations. They automatically发起会话, automatically承接对话线程上下文, and autonomously negotiate岗位、薪资、入职条件. The entire process runs unmanned without any human attendance.

3. **Fully Autonomous Memory & Vector Store**: Agents automatically extract user preferences, salary底线, city preferences, job type倾向, and other key information from conversations. They automatically vectorize and write to the vector knowledge base. Subsequent reasoning automatically recalls historical memory. The entire process of填充 and retrieval is fully automated without human intervention.

4. **Full Autonomous Tool Calling**: Agents have complete autonomous tool calling permissions. They self-determine timing, proactively call MCP, system business interfaces, and vector search tools. They autonomously search jobs, search resumes, initiate interviews, generate offers, and perform secondary job-person matching. The entire process is self-directed and self-executed.

## User Scenarios & Testing

### User Story 1 - Three-Part System Prompt Architecture (Priority: P1)

As an AI Agent, I need structured system prompts so my behavior is predictable, professional, and consistent.

**Why this priority**: Foundation determines all downstream AI behavior quality.

**Independent Test**: Can be tested by prompting the agent with various scenarios and verifying consistent, appropriate responses.

**Acceptance Scenarios**:

1. **Given** a new conversation, **When** the agent responds, **Then** the response follows the three-part prompt structure (must do, must not do, behavior rules)
2. **Given** a seeker agent, **When** it responds, **Then** it uses the seeker-specific prompt with recruitment industry context
3. **Given** a recruiter agent, **When** it responds, **Then** it uses the recruiter-specific prompt with candidate evaluation context
4. **Given** a salary negotiation scenario, **When** the agent responds, **Then** it uses the salary negotiation sub-prompt with professional tone
5. **Given** a long conversation, **When** context exceeds token limits, **Then** the system auto-compresses context while preserving key information

---

### User Story 2 - Function Calling Tool Integration (Priority: P1)

As an AI Agent, I need to call internal APIs to get real data so I never provide fabricated information.

**Why this priority**: Core requirement - agents must query real data before responding.

**Independent Test**: Can be tested by asking agents about jobs/candidates and verifying they call the correct APIs.

**Acceptance Scenarios**:

1. **Given** a user asks about job details, **When** the agent responds, **Then** it first calls the job query API and includes actual job data in response
2. **Given** a user asks about candidate skills, **When** the agent responds, **Then** it calls the candidate profile API and references real skills
3. **Given** an offer stage is reached, **When** the agent needs to create an offer, **Then** it calls the offer generation API with correct parameters
4. **Given** the agent needs to schedule an interview, **When** it processes the request, **Then** it calls the interview invitation API
5. **Given** the agent needs to find matching candidates, **When** it searches, **Then** it calls the vector similarity search API

---

### User Story 3 - Short-Term Conversation Memory (Priority: P1)

As an AI Agent, I need to maintain conversation context so I remember what was discussed without repeating myself.

**Why this priority**: Essential for natural conversation flow and professional interaction.

**Independent Test**: Can be tested by having multi-turn conversations and verifying context retention.

**Acceptance Scenarios**:

1. **Given** a conversation with 10+ messages, **When** the agent responds, **Then** it references earlier context appropriately without re-asking for information already provided
2. **Given** a user corrects information mid-conversation, **When** the agent continues, **Then** it uses the corrected information
3. **Given** redundant information is provided, **When** the agent responds, **Then** it acknowledges without duplicating back the same information
4. **Given** context needs truncation, **When** token limits approach, **Then** the system generates a summary and preserves critical information

---

### User Story 4 - Long-Term Vector Memory (Priority: P2)

As an AI Agent, I need to remember user preferences over time so users don't repeat themselves across sessions.

**Why this priority**: Improves user experience and reduces friction in recurring conversations.

**Independent Test**: Can be tested by creating preferences in one session and verifying recall in a new session.

**Acceptance Scenarios**:

1. **Given** a user specifies salary expectations, **When** they start a new conversation later, **Then** the agent recalls and applies those salary preferences
2. **Given** a user mentions preferred work location, **When** the agent matches jobs, **Then** it prioritizes jobs matching the user's location preference
3. **Given** a user has a preferred job type (remote/hybrid/onsite), **When** job results are returned, **Then** they are filtered/ranked by this preference

---

### User Story 5 - Long-Chain Reasoning & Task Planning (Priority: P2)

As an AI Agent, I need to follow a structured reasoning process so my responses are professional and actionable.

**Why this priority**: Ensures agents provide high-quality, well-reasoned responses aligned with business processes.

**Independent Test**: Can be tested by presenting complex scenarios and verifying the agent follows correct reasoning steps.

**Acceptance Scenarios**:

1. **Given** a candidate asks about a job, **When** the agent responds, **Then** it follows the process: understand request → query relevant data → analyze fit → generate response
2. **Given** a request requires an action (schedule interview, send offer), **When** the agent processes it, **Then** it executes the action through the appropriate API
3. **Given** an ambiguous request, **When** the agent receives it, **Then** it asks clarifying questions before taking action
4. **Given** a request is off-topic or outside the agent's scope, **When** the agent receives it, **Then** it redirects to relevant topics without engaging in unauthorized conversations

---

### User Story 6 - Multi-Agent Parallel Scheduling (Priority: P3)

As a system, I need to support multiple agents running concurrently so the platform scales to many simultaneous users.

**Why this priority**: Enables multi-tenant/multi-user scenarios with isolated agent contexts.

**Independent Test**: Can be tested by running multiple agent conversations in parallel and verifying isolation.

**Acceptance Scenarios**:

1. **Given** 10 concurrent user conversations, **When** agents respond, **Then** each conversation is processed independently without cross-contamination
2. **Given** multiple recruiter agents are active, **When** they process messages, **Then** each agent only accesses its own job postings and matches
3. **Given** a message is being processed, **When** another message for the same conversation arrives, **Then** the system handles ordering correctly without duplicate responses

---

### User Story 7 - Fully Autonomous Dual-Agent Negotiation (Priority: P1)

As a system, I need dual agents (求职Agent and 招聘Agent) to negotiate completely autonomously without any human intervention in the middle of the process.

**Why this priority**: Core system positioning - the entire recruitment flow depends on autonomous agent negotiation.

**Independent Test**: Can be verified by observing complete negotiation flow between two agents with no human input.

**Acceptance Scenarios**:

1. **Given** a job seeker initiates interest in a position, **When** the negotiation begins, **Then** the求职Agent and 招聘Agent automatically engage in dialogue without human prompting
2. **Given** the negotiation is in progress, **When** salary, job type, or start date are discussed, **Then** both agents autonomously reach agreement through their own dialogue
3. **Given** the negotiation completes successfully, **When** the final terms are agreed, **Then** the system automatically proceeds to offer generation without human confirmation
4. **Given** the negotiation fails or reaches impasse, **When** agents cannot agree, **Then** the system logs the outcome and does not proceed further

---

### User Story 8 - Autonomous Memory Extraction & Vector Storage (Priority: P1)

As an AI Agent, I need to automatically extract user preferences from conversation and store them in vector database without human intervention.

**Why this priority**: Enables fully autonomous memory management - critical for A2A autonomous mode.

**Independent Test**: Can be tested by reviewing vector store after conversations and verifying preferences were automatically extracted and stored.

**Acceptance Scenarios**:

1. **Given** a user mentions salary expectations, **When** the conversation ends, **Then** the agent automatically extracts, vectorizes, and stores the preference in pgvector
2. **Given** a user mentions city/location preference, **When** the conversation ends, **Then** the agent automatically extracts and stores the location preference
3. **Given** a user mentions job type preference (remote/hybrid/onsite), **When** the conversation ends, **Then** the agent automatically stores the preference
4. **Given** a new session starts, **When** the agent needs to recall preferences, **Then** it automatically retrieves relevant vectors from the store
5. **Given** the agent recalls preferences, **When** it generates responses, **Then** it automatically applies those preferences to filter/rank results

---

### User Story 9 - Autonomous Tool Execution & Decision Making (Priority: P1)

As an AI Agent, I need to autonomously decide when to call tools and execute them without human confirmation for each action.

**Why this priority**: Enables fully autonomous operation - agents must be able to self-direct throughout the entire process.

**Independent Test**: Can be tested by observing agent tool calling behavior during complex scenarios.

**Acceptance Scenarios**:

1. **Given** a求职Agent needs job recommendations, **When** it determines it's time to search, **Then** it autonomously calls the job search tool and incorporates results
2. **Given** a招聘Agent needs candidate matches, **When** it decides to search, **Then** it autonomously calls vector similarity search and presents matches
3. **Given** an interview stage is reached, **When** agents agree it's time, **Then** the agent autonomously calls schedule_interview and confirms the booking
4. **Given** an offer stage is reached, **When** terms are finalized, **Then** the agent autonomously calls create_offer and generates the offer document
5. **Given** the agent encounters an error in tool execution, **When** the error occurs, **Then** it autonomously handles the failure gracefully and informs the other agent

---

### User Story 10 - Human Final Confirmation Only (Priority: P1)

As a system, I need to limit human involvement to only final key result confirmation, with all intermediate processes running autonomously.

**Why this priority**: Defines the human role in the autonomous A2A system - confirmation only, not participation.

**Independent Test**: Can be tested by verifying human receives only final confirmation requests and does not receive mid-negotiation messages.

**Acceptance Scenarios**:

1. **Given** agents complete an offer negotiation, **When** terms are finalized, **Then** only then is human asked to confirm the final offer
2. **Given** agents complete an interview scheduling, **When** both parties agree, **Then** only then is human asked to confirm the interview time
3. **Given** a human receives a confirmation request, **When** they respond, **Then** the system proceeds or cancels based on that single confirmation
4. **Given** the human rejects a confirmation, **When** they provide feedback, **Then** agents autonomously incorporate the feedback and renegotiate

---

### Edge Cases

- What happens when the AI returns malformed JSON in a function call response?
- How does the system handle very long messages (over 1000 characters)?
- What happens when the vector similarity search returns no results?
- How does the agent handle rate limiting from internal API calls?
- What happens when PostgreSQL vector store is unavailable?
- How does the system behave when RabbitMQ consumer is overloaded?
- What happens when context summarization fails?
- What happens when dual agents cannot reach agreement after N rounds?
- How does the system handle conflicting preferences extracted from different conversations?
- What happens when human rejects final confirmation?

## Requirements

### Functional Requirements

- **FR-001**: System MUST use three-part prompt structure (must do + must not do + behavior rules) for all agent types
- **FR-002**: System MUST maintain separate prompt templates for seeker agents and recruiter agents
- **FR-003**: System MUST provide scenario-specific sub-prompts for: initial greeting, job negotiation, salary discussion, interview invitation, offer communication, decline/termination
- **FR-004**: System MUST implement automatic context compression when conversation token count exceeds 80% of model limit
- **FR-005**: System MUST implement context summarization that preserves key facts while reducing token usage
- **FR-006**: Agent MUST call job query API before providing job details
- **FR-007**: Agent MUST call candidate profile API before discussing candidate qualifications
- **FR-008**: Agent MUST call interview invitation API when scheduling interviews
- **FR-009**: Agent MUST call offer generation API when extending offers
- **FR-010**: Agent MUST call vector similarity API for job-candidate matching
- **FR-011**: Agent MUST NOT respond with information it has not verified through an API call
- **FR-012**: System MUST implement conversation short-term memory with thread isolation
- **FR-013**: System MUST implement automatic conversation summarization for long threads
- **FR-014**: System MUST store long-term user preferences (salary, location, job type) as vectors in PostgreSQL
- **FR-015**: System MUST recall relevant memories before each AI response generation
- **FR-016**: Agent MUST follow standard reasoning flow: understand → query → analyze → respond
- **FR-017**: Agent MUST refuse off-topic conversations and redirect to recruitment scope
- **FR-018**: System MUST support concurrent agent execution via RabbitMQ with session isolation
- **FR-019**: System MUST ensure agent memory isolation (no cross-conversation context leakage)
- **FR-020**: System MUST prevent duplicate message processing via idempotency checks
- **FR-021**: All existing core functionality MUST remain intact: RabbitMQ messaging, WebSocket communication, PostgreSQL vector matching, chat UI, conversation threads, interview and offer flows
- **FR-022**: System MUST enable fully autonomous dual-agent negotiation with no human intervention in intermediate processes
- **FR-023**: System MUST support automatic session initiation between求职Agent and 招聘Agent when match is created
- **FR-024**: System MUST enable agents to autonomously negotiate岗位、薪资、入职条件 through natural dialogue
- **FR-025**: System MUST automatically extract user preferences (salary底线, city preference, job type倾向) from conversation context
- **FR-026**: System MUST automatically vectorize extracted preferences and store in pgvector without human intervention
- **FR-027**: System MUST automatically recall and apply user preferences from vector store in subsequent reasoning
- **FR-028**: System MUST grant agents complete autonomous tool calling permissions for MCP, business APIs, and vector retrieval
- **FR-029**: System MUST allow agents to autonomously decide timing and execute tool calls (search jobs, schedule interview, generate offer)
- **FR-030**: System MUST limit human involvement to final key result confirmation only (no mid-negotiation participation)
- **FR-031**: System MUST notify human only when agents reach final agreement requiring confirmation
- **FR-032**: System MUST enable agents to autonomously incorporate human rejection feedback and renegotiate

### Key Entities

- **AgentPrompt**: Template for agent system prompts with role-specific variations (seeker vs recruiter)
- **PromptScenario**: Scenario-specific prompt fragments (greeting, negotiation, salary, etc.)
- **ConversationSummary**: Auto-generated summary of conversation state for context compression
- **UserPreferenceVector**: Vectorized representation of user preferences stored in PostgreSQL
- **AgentToolCall**: Record of API calls made by agent during reasoning
- **AgentMemory**: Short-term and long-term memory entries for agent context
- **NegotiationSession**: Autonomous negotiation state between求职Agent and 招聘Agent
- **PreferenceExtraction**: Extracted preference items (salary, location, job type) auto-extracted from conversation
- **ConfirmationRequest**: Human confirmation request sent only at final agreement stage

## Success Criteria

### Measurable Outcomes

- **SC-001**: Agent responses follow three-part prompt structure in 95% of conversations
- **SC-002**: Zero fabricated information - all specific facts verified via API calls before being communicated
- **SC-003**: Context compression reduces token count by at least 40% while preserving 90% of key information
- **SC-004**: Long-term preference recall accuracy reaches 85% across sessions
- **SC-005**: Agent follows reasoning process (understand → query → analyze → respond) in 90% of complex scenarios
- **SC-006**: System handles 100 concurrent agent conversations without cross-contamination
- **SC-007**: Average response time remains under 5 seconds even with function calling and memory retrieval
- **SC-008**: All existing E2E test scenarios pass without modification (backward compatibility)
- **SC-009**: E2E tests verify AI-driven flows: agent calls APIs, agent recalls memories, agent follows reasoning process
- **SC-010**: Dual-agent autonomous negotiation completes without any human intervention in intermediate processes
- **SC-011**: Agent automatically extracts and stores at least 90% of explicit user preferences (salary, location, job type) to vector store
- **SC-012**: Agent automatically recalls and applies stored preferences in 85% of relevant reasoning scenarios
- **SC-013**: Agent autonomously executes tool calls (job search, interview scheduling, offer generation) without human confirmation
- **SC-014**: Human receives only final confirmation requests (0 mid-negotiation interruptions)

## Assumptions

- Users have stable internet connectivity for API calls
- PostgreSQL vector extension (pgvector) is available for similarity search
- Existing RabbitMQ infrastructure supports the message schemas
- AI model supports function calling / tool use (OpenAI GPT-4 or equivalent)
- Frontend can use Vercel AI SDK for UI state management (ai-sdk.dev)
- Token limits are based on the AI model's context window (e.g., 128k for GPT-4)
- Existing database schema can be extended with new vector storage tables
- Current WebSocket infrastructure handles the increased message volume
- Human accepts "confirmation only" role in recruitment workflow
- Dual-agent conversation can reach agreement within reasonable negotiation rounds (max 20)
- MCP tools and business APIs are available for autonomous agent execution
