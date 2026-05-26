# Phase 2.2: 构建 adapters 层 — 外部依赖薄封装

> Parent: 大重构计划

## What to build

创建 `internal/adapters/` 目录，将外部依赖（AI、RabbitMQ、Postgres、Parser、Embedding）封装到 adapters 层。所有 Repository 合并为一个 `postgres.go`。

## Acceptance criteria

- [ ] `pkg/ai/client.go` → `adapters/ai_client.go`（修复：mockRound 并发、流式错误信号）
- [ ] `pkg/rabbitmq/` → `adapters/rabbitmq.go`（添加：发布重试 3 次 + DLQ）
- [ ] `repository/` 所有文件 → `adapters/postgres.go`（合并为一个文件）
- [ ] `pkg/parser/` → `adapters/parser.go`
- [ ] Jina Embedding 客户端 → `adapters/embedding.go`
- [ ] `go build ./...` 通过

## Blocked by

- #02 Phase 1: 删除尘余代码
