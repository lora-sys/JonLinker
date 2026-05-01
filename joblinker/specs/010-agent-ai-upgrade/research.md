# Research: AI Agent Capability Upgrade

**Feature**: specs/010-agent-ai-upgrade
**Date**: 2026-05-01
**Research Methods**: Firecrawl web search and scraping

---

## 1. Three-Part Prompt Structure

**Decision**: Adopt "Persona + Scope + Constraints" pattern for agent prompts

**Source**: Oracle AI Agent Studio best practices (2026-02-18)

### Prompt Anatomy (5 layers):
1. **Persona** — Who the agent is, professional identity
2. **Scope** — What topics the agent can address
3. **Tools** — Available APIs and capabilities
4. **Constraints** — What the agent MUST NOT do
5. **Behavior Rules** — How to respond in specific scenarios

### Three-Part Structure Alignment:
| Requirement | Implementation |
|-------------|----------------|
| Must Do | Tools + Scope |
| Must Not Do | Constraints |
| Behavior Rules | Persona + Scenario-specific topics |

### Key Finding from Oracle Blog:
> "Never generate facts. All answers should rely on tool call response."
> — This validates FR-011: Agent must verify via API before responding.

---

## 2. Function Calling Patterns

**Decision**: Use OpenAI-style function calling schema with typed parameters

**Source**: Prompt Engineering Guide (promptingguide.ai)

### Function Calling Flow:
```
User Question → LLM detects function need → Returns JSON arguments
→ Execute external API → Return result to LLM → Generate final response
```

### Best Practices:
- Define functions with clear `description` for LLM understanding
- Use `enum` for constrained parameters (e.g., job type: remote/hybrid/onsite)
- Mark required fields in `parameters.required`
- Handle `tool_calls` in response object

### For JobLinker: Required Functions
| Function | Purpose | Required Params |
|----------|---------|-----------------|
| `query_jobs` | Search jobs | location, skills, salary_range |
| `get_candidate` | Get candidate profile | candidate_id |
| `create_offer` | Generate offer | match_id, salary, terms |
| `schedule_interview` | Book interview | match_id, datetime |
| `search_candidates` | Vector similarity | skills, location |

---

## 3. Long-Term Memory with Vector Databases

**Decision**: Use PostgreSQL pgvector for user preferences, Chroma optional for additional context

**Source**: Medium article on ChromaDB as AI memory (2026)

### Memory Architecture:
| Type | Storage | Recall Method | Use Case |
|------|---------|---------------|----------|
| Short-term | Conversation context | Token window | Current session |
| Long-term | pgvector (PostgreSQL) | Similarity search | Cross-session preferences |

### Key Insights:
- User preferences (salary, location, job type) → Store as embeddings in pgvector
- Recall at session start: `SELECT * FROM preferences WHERE embedding <-> query_embedding < threshold`
- FR-014 requirement aligns with this architecture

### Search vs. pgvector:
| Aspect | Chroma | pgvector |
|--------|--------|----------|
| Deployment | Separate service | Embedded in PostgreSQL |
| Consistency | CP (consistency partition) | CA (consistency availability) |
| Integration | Requires extra connection | Use existing DB connection |

**Recommendation**: pgvector is sufficient for JobLinker. No need for Chroma.

---

## 4. Short-Term Memory & Context Management

**Decision**: Implement context window tracking with automatic summarization trigger at 80% capacity

**Source**: Multi-agent system research + JobLinker existing implementation

### Context Lifecycle:
```
New Message → Count tokens → If >80% limit:
  → Generate summary (preserve key facts: salary, location, skills)
  → Replace conversation with summary + recent messages
```

### Key Facts to Preserve:
- User's stated salary expectations
- Preferred work location
- Job type preference (remote/hybrid/onsite)
- Specific skills mentioned as important
- Rejected job titles (to avoid showing similar)

### Implementation from spec:
- FR-004: Context compression at 80% token limit
- FR-005: Summarization preserves 90% of key information (SC-003: 40% reduction)

---

## 5. Long-Chain Reasoning (ReAct Pattern)

**Decision**: Implement understand → query → analyze → respond flow with explicit tool calls

**Source**: Prompt Engineering Guide + Oracle Agent Studio

### ReAct Loop:
```
1. PERCEIVE → Read goal and context
2. THINK → Plan which tools to call
3. ACT → Execute tool calls
4. OBSERVE → Review results
5. REPEAT → Until goal achieved
```

### For JobLinker:
```
Candidate: "I want a job paying >$80k in London"
↓
UNDERSTAND: Extract requirements (salary >80k, location London)
↓
QUERY: Call search_candidates(skills, location=London) + query_jobs(location=London)
↓
ANALYZE: Score matches against preferences
↓
RESPOND: Present top matches with reasoning
```

### Validation:
- SC-005: 90% of complex scenarios follow this flow
- FR-016: Must follow understand → query → analyze → respond

---

## 6. Multi-Agent Scheduling

**Decision**: Use RabbitMQ with session isolation via match_id correlation

**Source**: JobLinker existing RabbitMQ implementation

### Architecture:
```
User Message → RabbitMQ (match_id routing key)
→ Agent Consumer (match_id determines context isolation)
→ Process → Response via WebSocket
```

### Key Requirements:
- FR-018: Concurrent execution with session isolation
- FR-019: No cross-conversation context leakage
- FR-020: Idempotency checks prevent duplicate processing

### Idempotency Implementation:
```
message_id = hash(user_id + match_id + timestamp)
IF message_id IN processed_messages:
  SKIP
ELSE:
  PROCESS + ADD to processed_messages
```

---

## 7. Chain-of-Thought Prompting

**Decision**: Use "Let's think step by step" for complex reasoning

**Source**: Prompt Engineering Guide - Chain-of-Thought section

### When to Use:
- Multi-step logic (e.g., job matching with multiple criteria)
- Commonsense reasoning
- Debugging suggestions

### Three Variants:
| Variant | Use Case | Effectiveness |
|---------|----------|---------------|
| Zero-Shot CoT | No examples available | "Let's think step by step" |
| Few-Shot CoT | Have good examples | Show reasoning steps |
| Self-Consistency | High-stakes decisions | Majority voting on multiple runs |

**Recommendation**: Use Zero-Shot CoT for most responses. Few-Shot for salary negotiation scenarios.

---

## 8. Context Window Management

**Decision**: Track token count, compress at 80%, preserve specific keys

**Source**: AI Agent best practices

### Token Counting Strategy:
- Count system prompt + conversation history + tool results
- If total > 80% of model limit: trigger compression
- Preserve: salary_expectations, preferred_location, job_type, rejected_titles

### Compression Algorithm:
```
1. Generate summary: "User prefers remote Python work, $100k+ salary, rejected Java roles"
2. Keep last N messages for context continuity
3. Replace middle messages with summary
4. Token reduction target: 40%+ (SC-003)
```

---

## Key Decisions Summary

| Area | Decision | Rationale |
|------|----------|------------|
| Prompt Structure | Persona + Scope + Constraints + Behavior | Matches three-part requirement |
| Function Calling | OpenAI-style JSON schema | Industry standard, good LLM support |
| Long-term Memory | pgvector in PostgreSQL | Single DB, no extra service |
| Short-term Memory | Token tracking + auto-summarization | Standard context window management |
| Reasoning | ReAct loop with explicit steps | Proven pattern for tool use |
| Scheduling | RabbitMQ + match_id isolation | Existing infrastructure |
| CoT Prompting | "Let's think step by step" | Simplest effective approach |

---

## References

1. [A Practical Guide to Prompt Engineering and AI Agents](https://medium.com/@vprprudhvi/a-practical-guide-to-prompt-engineering-and-ai-agents-004ce4647549) - 2026-02-08
2. [Function Calling with LLMs](https://www.promptingguide.ai/applications/function_calling) - Prompt Engineering Guide
3. [Basics of Prompt Engineering](https://blogs.oracle.com/fusioncoe/basics-of-prompt-engineering) - Oracle AI Agent Studio - 2026-02-18
4. [Using ChromaDB as Long-Term Memory for AI Agents](https://medium.com/@techlatest.net/using-chromadb-as-long-term-memory-for-ai-agents-da96ed843e75)