# AGENT.md

## Mission

JobLinker exists to build an autonomous AI recruiting system.

The final vision is:

Candidate Agent
→ Search Jobs
→ Apply Jobs
→ Communicate
→ Negotiate
→ Receive Offer
→ Human Confirmation

Human involvement should continuously decrease as the system evolves.

---

## Core Principle

DO NOT BUILD THE FINAL SYSTEM FIRST.

Every feature must belong to a specific phase.

If a feature is not required by the current phase,
DO NOT IMPLEMENT IT.

---

## Product Vision

We are NOT building:

* Agent Framework
* Workflow Engine
* MCP Platform
* Multi-Agent Operating System
* Generic AI Infrastructure

We ARE building:

AI Recruiter

An autonomous recruiting agent that helps candidates find jobs, apply, communicate, negotiate and receive offers.

---

## Development Rules

### Rule 1

Current phase completion is more important than future architecture.

### Rule 2

Every feature must map to a user value.

Bad:

* RabbitMQ
* Gateway
* Vector DB

Good:

* Find Jobs
* Apply Jobs
* Recruiter Chat
* Offer Negotiation

### Rule 3

No future-proof engineering.

Build only what is needed now.

### Rule 4

No new infrastructure unless current phase requires it.

Examples:

DO NOT introduce:

* RabbitMQ
* Kafka
* Event Bus
* Workflow Engine
* Distributed System

until required by a phase.

### Rule 5

Every phase must have a demo.

If a feature cannot be demonstrated to a user,
it is not complete.

---

## Success Metric

The success of JobLinker is measured by:

How much recruiting work can be delegated to AI.

Not by:

* Number of services
* Number of agents
* Number of databases
* Number of abstractions
* Number of architecture diagrams
