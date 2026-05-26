# Phase 2.3: 构建 engine/nodes 层 — Eino Graph 节点

> Parent: 大重构计划

## What to build

创建 `internal/engine/nodes/` 目录，将 Eino Graph 节点从旧位置提取出来。RabbitMQ 驱动的 Graph 执行器放在 `engine/runner.go`。

## Acceptance criteria

- [ ] `agent/decision.go` + `pkg/ai` → `nodes/ai_node.go`
- [ ] `agent/tool_executor` + `function_definitions` → `nodes/tool_node.go`（修复：空结果 vs 错误区分 status）
- [ ] `agent_memory` + `preference_extraction` → `nodes/memory_node.go`
- [ ] `confirmation_service` + `interview_service` → `nodes/confirm_node.go`（修复：24h 自动 reject + 通知）
- [ ] `eino/` 目录内容 → `engine/graph.go` + `engine/runner.go`（修复：runningGraphs defer delete 防泄漏）
- [ ] 新建 `engine/runner.go`（RabbitMQ 驱动）
- [ ] `go build ./...` 通过

## Blocked by

- #03 Phase 2.1: 提取 core 层
- #04 Phase 2.2: 构建 adapters 层
