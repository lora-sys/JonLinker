# Plan: Agent Tool Calling

> Source PRD: Conversation 2026-05-22 — Eino ADK agent tool calling extension

## Architectural decisions

- **WebSocket events**: `adk_tool_call_start` (LLM decided to call → show pending/running) + `adk_tool_call` (execution complete → show result/error)
- **Tool definition**: Eino `utils.NewTool()`, name/desc/params self-documenting via `schema.ToolInfo`
- **Tool registration**: Single function `NewRealTools(repos...)`, ADK auto-binds via `compose.WithChatModelOption(model.WithTools(bc.toolInfos))`
- **Frontend state**: `toolCalls: ToolCall[]` independent state (not embedded in UIMessage.parts), page layer passes to both message area and FlowPanel
- **System prompt**: Single-sentence tool list ("You have access to: ..."), no parameter duplication
- **DeepAgent**: Phase 3 introduces `deep.New()` + `localbk.NewBackend` for filesystem tools

---

## Phase 1: Tool Call Visibility (Existing 5 Tools)

**User stories**: ADK agent calls query_jobs / search_candidates / get_candidate / create_offer / schedule_interview → frontend shows tool call in message area and FlowPanel

### What to build

Fix `generateWithADK` event loop to properly detect assistant messages with tool_calls (emit `adk_tool_call_start`) and tool role result messages (emit `adk_tool_call` instead of treating as final response). Frontend handles both events, stores in `toolCalls` state, renders `<Tool>` collapsible cards in message list and passes to `FlowPanel`.

### Acceptance criteria

- [ ] Backend emits `adk_tool_call_start` when LLM decides to call a tool (tool_name, call_id, arguments, sender_agent_id, sender_label)
- [ ] Backend emits `adk_tool_call` when tool completes (tool_name, call_id, result, status, error, sender_agent_id, sender_label, cached)
- [ ] Backend no longer treats Tool role events as final response
- [ ] Frontend `useAIChat` processes both events into `toolCalls` state
- [ ] Message area shows collapsible `<Tool>` card per tool call (input arguments + output result)
- [ ] FlowPanel shows tool call list with status icons
- [ ] E2E: create match → send message → agent calls query_jobs → tool card visible in chat + sidebar

---

## Phase 2: Custom Database Tools

**User stories**: Agent can autonomously query match progress, interview details, offer status, user profiles

### What to build

Add 4 new Eino tools as `utils.NewTool()` executors backed by real repositories:

- `get_match_progress(match_id)` → match stage, status, version
- `get_interview_details(match_id)` → interview datetime, format, status
- `get_offer_details(match_id)` → offer salary, compensation, status
- `get_user_profile(agent_id)` → agent config, skills, experience

Register in `NewRealTools()`. Append one line to system prompts listing available tools.

### Acceptance criteria

- [ ] 4 new tools defined with `schema.ToolInfo` and typed input/output structs
- [ ] Tools registered in `NewRealTools()` with real repository executors
- [ ] System prompts updated with tool list
- [ ] LLM autonomously calls new tools in conversation
- [ ] Tool results display in frontend (Phase 1 infrastructure)

---

## Phase 3: DeepAgent + Filesystem Tools

**User stories**: Agent can read/write files, execute commands

### What to build

Upgrade DeepRecruiterAgent to use `deep.New()` with `localbk.NewBackend` configuration. Auto-registers `read_file`, `write_file`, `edit_file`, `glob`, `grep`, `execute` tools. Existing custom tools remain available via `ToolsConfig`.

### Acceptance criteria

- [ ] `deep.New()` creates DeepAgent with Backend and StreamingShell
- [ ] Filesystem tools auto-registered and available to LLM
- [ ] Existing custom tools still work alongside filesystem tools
- [ ] Agent can read project files and display content in chat
