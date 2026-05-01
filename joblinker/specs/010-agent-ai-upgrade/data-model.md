# Data Model: AI Agent Capability Upgrade

**Feature**: specs/010-agent-ai-upgrade
**Date**: 2026-05-01

## Entities

### AgentPrompt

Template for agent system prompts with role-specific variations.

```go
type AgentPrompt struct {
    ID          string    // UUID
    Role        AgentRole // seeker | recruiter
    Scenario    string    // greeting | negotiation | salary | interview | offer | decline
    MustDo      string    // List of required behaviors
    MustNotDo   string    // List of prohibitions
    Behavior    string    // Communication style and rules
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type AgentRole string

const (
    RoleSeeker   AgentRole = "seeker"
    RoleRecruiter AgentRole = "recruiter"
)
```

### PromptScenario

Scenario-specific prompt fragments for dynamic composition.

```go
type PromptScenario struct {
    ID          string    // UUID
    Scenario    string    // greeting | negotiation | salary | interview | offer | decline | termination
    Description string    // Human-readable description
    TriggerConditions string // When to activate this scenario
    PromptFragment string  // The actual prompt fragment to insert
    Priority    int       // Execution order (lower = earlier)
}
```

### ConversationSummary

Auto-generated summary for context compression.

```go
type ConversationSummary struct {
    ID            string    // UUID
    MatchID       string    // Foreign key to match
    SummaryText   string    // The compressed summary
    KeyFacts      []KeyFact // Preserved critical information
    TokenCount    int       // Original token count before compression
    CreatedAt     time.Time
}

type KeyFact struct {
    FactType string    // salary_expectation | preferred_location | job_type | rejected_titles | skills
    Value    string    // The actual fact value
    Priority int       // Higher = more important to preserve
}
```

### UserPreferenceVector

Vectorized user preferences stored in PostgreSQL (pgvector).

```go
type UserPreferenceVector struct {
    ID              string    // UUID
    UserID          string    // Foreign key to user
    PreferenceType  string    // salary | location | job_type | skills
    Embedding       []float32 // Vector embedding (1536 dims for OpenAI)
    RawValue         string    // Original text value (e.g., "London", "$100k+")
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### AgentToolCall

Record of API calls made during agent reasoning.

```go
type AgentToolCall struct {
    ID          string    // UUID
    MatchID     string    // Foreign key to match
    ToolName    string    // query_jobs | get_candidate | create_offer | etc
    Arguments   string    // JSON string of arguments
    Result      string    // JSON string of result
    Status      ToolStatus // pending | success | failed
    ExecutedAt  time.Time
}

type ToolStatus string

const (
    ToolStatusPending ToolStatus = "pending"
    ToolStatusSuccess ToolStatus = "success"
    ToolStatusFailed  ToolStatus = "failed"
)
```

### AgentMemory

Short-term and long-term memory entries for agent context.

```go
type AgentMemory struct {
    ID          string       // UUID
    MatchID     string       // Conversation thread ID
    MemoryType  MemoryType   // short_term | long_term
    Content     string       // The memory content
    Embedding   []float32    // For long-term recall
    RecallCount int          // Times this memory was recalled
    LastRecalledAt *time.Time // Last time this was used
    CreatedAt   time.Time
}

type MemoryType string

const (
    MemoryTypeShortTerm MemoryType = "short_term"
    MemoryTypeLongTerm  MemoryType = "long_term"
)
```

## State Transitions

### Conversation → Summary

```
Active Conversation (tokens > 80% limit)
    ↓
Summarization triggered
    ↓
Extract KeyFacts (salary, location, job_type, skills)
    ↓
Generate SummaryText
    ↓
Replace conversation history with ConversationSummary
```

### UserPreferenceVector Creation

```
User message contains preference
    ↓
Extract preference (type + value)
    ↓
Generate embedding via AI model
    ↓
Store in PostgreSQL with pgvector
    ↓
Available for cross-session recall
```

## Validation Rules

| Field | Rule |
|-------|------|
| AgentPrompt.MustDo | Required, min 50 chars |
| AgentPrompt.MustNotDo | Required, min 30 chars |
| AgentPrompt.Behavior | Required, min 40 chars |
| UserPreferenceVector.Embedding | Must be 1536-dimensional (OpenAI ada) |
| ConversationSummary.KeyFacts | Min 3 facts preserved |
| AgentToolCall.Result | Valid JSON or error message |