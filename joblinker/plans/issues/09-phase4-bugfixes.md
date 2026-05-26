# Phase 4.1: 修复 11 个关键 bug

> Parent: 大重构计划

## What to build

修复重构过程中发现以及遗留的 11 个关键 bug。

## Bug list

| # | Bug | 位置 | 修复方式 |
|---|-----|------|----------|
| 1 | negotiator 除零 panic | `core/negotiation.go` | min==target 时 early return |
| 2 | FSM CanHandle/Handle 不一致 | `core/fsm.go` | 合并为单一 Handle 方法 |
| 3 | StatePaused 无转换表 | `core/fsm.go` | 补充 Resume/Rejected/Timeout 转换 |
| 4 | FSM 零并发保护 | `core/fsm.go` | 加 sync.Mutex |
| 5 | 消息发布失败静默丢失 | `adapters/rabbitmq.go` | 指数退避重试 3 次 + DLQ |
| 6 | runningGraphs 泄漏 | `engine/runner.go` | defer delete |
| 7 | WebSocket 重连无补推 | `transport/ws/websocket_handler.go` | 支持 `since=<timestamp>` 参数 |
| 8 | 工具空结果 vs 错误混淆 | `engine/nodes/tool_node.go` | 返回不同 status |
| 9 | JSON Unmarshal 忽略 | 全局 | 检查所有 JSON error 返回值 |
| 10 | confirm 超时无处理 | `engine/nodes/confirm_node.go` | 24h 自动 reject + 通知 |
| 11 | AI 流式响应无错误信号 | `adapters/ai_client.go` | 使用带 error 的结果 channel |

## Acceptance criteria

- [ ] 11 个关键 bug 全部修复并验证
- [ ] `go build ./...` + `go vet ./...` 通过

## Blocked by

- #07 Phase 2.5: 组装 main.go + 删除旧目录
