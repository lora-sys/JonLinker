# Phase 2.1: 提取 core 层 — FSM/Negotiator/Models

> Parent: 大重构计划

## What to build

创建 `internal/core/` 目录，将零外部依赖的纯领域逻辑从旧位置迁移过来。同时修复已知的 FSM 和 Negotiator bug。

## Acceptance criteria

- [ ] 创建 `internal/core/` 目录
- [ ] `agent/fsm.go` → `core/fsm.go`（修复：合并 CanHandle/Handle，补充 StatePaused 转换表，加 sync.Mutex 并发保护）
- [ ] `agent/negotiator.go` → `core/negotiation.go`（修复：除零 panic，min==target 时 early return）
- [ ] `model/` 中的核心模型 → `core/models.go`
- [ ] `go build ./...` 通过
- [ ] FSM 和 Negotiator 单元测试通过

## Blocked by

- #02 Phase 1: 删除尘余代码
