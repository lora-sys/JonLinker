# CLAUDE.md

## Project Identity

JobLinker

AI Recruiting Agent

Goal:

Reduce human involvement in recruiting through autonomous AI agents.

---

# Development Roadmap

## Phase 1

Job Search Agent

User Input
→ Search Jobs
→ Ranked Results

Tools:

* query_jobs

Done When:

* Agent can search jobs
* Agent returns ranked jobs
* Agent explains recommendations
* 95% success rate

---

## Phase 2

Auto Apply Agent

User Input
→ Search Jobs
→ Apply Jobs

Tools:

* query_jobs
* apply_job

Done When:

* Agent writes applications
* Agent submits applications
* User does not manually write content

---

## Phase 3

Recruiter Chat

User
↔ Agent
↔ Recruiter

Capabilities:

* Multi-turn conversation
* Context tracking
* Job discussion

Done When:

* 20-turn conversations succeed
* Context remains consistent

---

## Phase 4

A2A Recruiter

Candidate Agent
↔ Recruiter Agent

Done When:

* Agents communicate autonomously
* No human required during conversation
* Conversation state is trackable

---

## Phase 5

Negotiation

Candidate Agent
↔ Recruiter Agent

Capabilities:

* Salary negotiation
* Start date negotiation
* Offer generation

Done When:

* Offer generated automatically
* Human only confirms final result

---

## Phase 6

Memory

Capabilities:

* Preference extraction
* Preference recall
* Long-term user profile

Done When:

* Previous preferences affect future decisions

---

## Phase 7

Autonomous Recruiting

Search
→ Apply
→ Chat
→ Negotiate
→ Offer
→ Human Confirmation

Done When:

* End-to-end recruiting workflow is autonomous

---

# Forbidden During Early Phases

Until Phase 4:

DO NOT BUILD

* RabbitMQ
* Kafka
* Workflow Engine
* DAG Engine
* Multi-Agent Framework
* Plugin Marketplace
* MCP Infrastructure
* Vector Memory

unless explicitly required by roadmap.

---

# Engineering Philosophy

Simple > Clever

Working > Scalable

Delivered > Designed

Validated > Imagined

# Must

use rtk  example rtk git status , git commit -m   
every phase test must use playwright cli  skills to open browser
and test the agent's capabilities in a real browser environment.
frontend test ui element  and backend test api response and agent's decision making process.
every phase run check issues  git checkout main -> new checkout branch-> commit changes -> push to origin -> create PR
remember to commit run eslint ,build ,type check  fontend and backend code before push to origin

# Must Not
forbid hardcode and call ai api mock data
