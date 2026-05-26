# Phase 2.4: 重写 transport 层 — REST + WebSocket

> Parent: 大重构计划

## What to build

创建 `internal/transport/rest/` 和 `internal/transport/ws/`，将 handler 迁入薄层，WebSocket 添加补推支持。

## Acceptance criteria

- [ ] `handler/a2a.go` → `transport/rest/agent_handler.go`
- [ ] `handler/match.go` → `transport/rest/match_handler.go`
- [ ] `handler/message.go` + `handler/a2a_sse.go` → `transport/rest/message_handler.go`
- [ ] `handler/` 中 WS 部分 → `transport/ws/websocket_handler.go`（添加 `?since=<timestamp>` 补推）
- [ ] `handler/health.go` → `transport/rest/health_handler.go`
- [ ] `handler/auth.go` → `transport/rest/auth_handler.go`
- [ ] `go build ./...` 通过

## Blocked by

- #05 Phase 2.3: 构建 engine/nodes 层
