# Phase 0 诊断报告 — 三份清单

> 生成日期: 2026-05-17
> 扫描范围: 全仓 157 个 Go 文件 + 前端 src/

---

## 一、删除清单（引用次数 = 0 或可安全删除）

### 后端死代码

| 文件 | 原因 |
|------|------|
| `internal/service/ai_orchestrator.go` | `generateEinoResponse` 始终返回 nil，降级链已由 message_queue_service.go 接管 |
| `internal/agent/function_definitions.go` | 旧 ADK agent 函数定义，已被 eino/tools/ 替代 |
| `internal/agent/xml_protocol.go` | 旧 XML 协议解析，已被 eino/stategraph 替代 |
| `internal/agent/skill_dispatch.go` | 旧技能调度，已被 eino/agent/supervisor* 替代 |
| `internal/agent/requirements.go` | 旧需求匹配逻辑，已被 eino/tools/job_tools.go 替代 |
| `internal/agent/offer_generator.go` | 旧 Offer 生成，已被 eino/workflow 替代 |
| `internal/agent/scheduler.go` | 旧调度器，已被 eino/runner 替代 |
| `internal/service/privacy_service.go` | 隐私服务，可合并到 handler |
| `internal/service/security_service.go` | 安全服务，可合并 |
| `internal/cache/context_optimizer.go` | 缓存优化器已由 eino/hooks/context_compressor 替代 |
| `internal/cache/tool_cache.go` | 工具缓存引用少，已由 eino/hooks 替代 |
| `internal/config/agent_tools.go` | 旧工具配置 |
| `internal/config/proto_config.go` | 旧 proto 配置 |
| `internal/middleware/gateway.go` | 旧 API 网关中间件 |
| `internal/middleware/rate_limit.go` | 旧限流中间件，已由 eino/hooks/rate_limiter 替代 |
| `internal/middleware/tracing.go` | 旧追踪中间件 |
| `pkg/redis/client.go` | 仅 seed 引用 1 次，可内联 |

### 后端可合并的文件

| 文件集 | 建议 |
|--------|------|
| `internal/repository/*` (16 个文件) | 合并为 `adapters/postgres.go` (计划中) |
| `internal/model/*` (13 个文件) | 收归 `core/models.go` (计划中) |
| `internal/handler/*` (14 个文件) | 迁入 `transport/rest/` + `transport/ws/` (计划中) |
| `internal/service/*` (15 个文件) | 分散到 engine/ 和 core/ |

---

## 二、迁移清单（需要移动但有大量引用的文件）

### 后端迁移

| 源路径 | 目标路径 | 引用数 | 迁移原因 |
|--------|----------|--------|----------|
| `internal/agent/fsm.go` | `internal/core/fsm.go` | 高 | 核心领域逻辑 |
| `internal/agent/negotiator.go` | `internal/core/negotiation.go` | 高 | 核心领域逻辑 |
| `internal/agent/decision.go` | `internal/engine/nodes/ai_node.go` | 中 | Eino Graph 节点 |
| `internal/agent/tool_executor.go` | `internal/engine/nodes/tool_node.go` | 高 | Eino Graph 节点 |
| `internal/eino/agent/*` | `internal/engine/nodes/*` | 高 | Eino Graph 节点 |
| `internal/eino/stategraph/*` | `internal/engine/nodes/*` | 中 | Eino Graph 节点 |
| `internal/eino/runner/*` | `internal/engine/runner.go` | 高 | Eino 编排 |
| `internal/eino/workflow/*` | `internal/engine/nodes/*` | 中 | Eino 编排 |
| `internal/eino/tools/*` | `internal/engine/nodes/*` | 中 | Eino 工具 |
| `internal/eino/hooks/*` | `internal/engine/hooks/*` | 高 | Eino 钩子 |
| `internal/eino/memory/*` | `internal/engine/memory/*` | 中 | Eino 记忆 |
| `internal/eino/prompt/*` | `internal/engine/prompt/*` | 中 | Eino 提示词 |
| `internal/eino/sessionstore/*` | `internal/engine/session/*` | 高 | Eino 会话存储 |
| `internal/eino/chatmodel/*` | `internal/engine/chatmodel/*` | 高 | Eino 聊天模型 |
| `internal/handler/*` | `internal/transport/rest/*` + `ws/*` | 极高 | 传输层 |
| `pkg/ai/client.go` | `internal/adapters/ai_client.go` | 14 | Adapter |
| `pkg/rabbitmq/rabbitmq.go` | `internal/adapters/rabbitmq.go` | 7 | Adapter |
| `pkg/parser/resume.go` | `internal/adapters/parser.go` | 1 | Adapter |
| `pkg/proto/*` | `internal/adapters/proto/*` | 5 | Adapter |
| `pkg/crypto/encryption.go` | `internal/adapters/crypto.go` | 1 | Adapter |
| `internal/middleware/auth.go` | `internal/transport/middleware/auth.go` | 高 |
| `internal/service/message_queue_service.go` | `internal/engine/queue_processor.go` | 中 | 引擎层 |

### 前端迁移

| 源路径 | 目标路径 |
|--------|----------|
| `hooks/useAIChat.ts` | `features/conversation/hooks/` |
| `hooks/useAgent.ts` | `features/agent/hooks/` |
| `hooks/useAuth.ts` | `shared/hooks/` |
| `hooks/useWebSocket.ts` | `shared/hooks/` |
| `hooks/useSSE.ts` | `shared/hooks/` |
| `hooks/useIntersectionReveal.ts` | `shared/hooks/` |
| `hooks/useRole.ts` | `shared/hooks/` |
| `stores/agent.ts` | `features/agent/stores/` |
| `stores/match.ts` | `features/match/stores/` |
| `stores/auth.ts` | `shared/stores/` |
| `stores/ui.ts` | `shared/stores/` |
| `components/ai-elements/*` | `features/conversation/components/` |
| `components/conversation/*` | `features/conversation/components/` |
| `components/FSMStatusBar.tsx` | `features/conversation/components/` |
| `components/HumanConfirmModal.tsx` | `features/confirmation/components/` |
| `components/interview/*` | `features/confirmation/components/` |
| `components/offer/*` | `features/confirmation/components/` |
| `components/feature/JobCard.tsx` | `features/match/components/` |
| `components/layout/*` | `shared/ui/` |
| `components/ui/*` | `shared/ui/` |

---

## 三、保留清单（看似冗余但实际在用）

| 文件 | 看似 | 真实用途 |
|------|------|----------|
| `internal/eino/runner/adk_runner.go` | "Legacy ADK" | ADK 是 Eino 的核心 runtime，非 Legacy |
| `internal/agent/fsm.go` | "旧 agent 代码" | FSM 核心逻辑需保留并迁移到 core/ |
| `internal/agent/negotiator.go` | "旧 agent 代码" | 薪资协商核心逻辑需保留并迁移到 core/ |
| `internal/agent/tool_executor.go` | "旧 tool 执行" | 仍被 message_queue_service.go 调用 |
| `pkg/shared/errors.go` | "仅 1 引用" | 公共错误类型，被多处隐式使用 |
| `pkg/shared/logger.go` | "仅 1 引用" | 公共日志工具 |
| `internal/service/fsm_integration.go` | "旧集成层" | 桥接 FSM 到 message queue，仍在使用 |
| `internal/service/observability_service.go` | "旧可观测性" | 实际在用的指标收集 |
| `internal/eino/agent/deep_agent.go` | "legacy" | 注释写 deprecated 但 ADK 依赖它 |

---

## 四、ADK 使用报告

**结论：ADK 不是 Legacy！不要删除！** ADK 是 Eino 生态的核心运行时，被大量使用。

使用 ADK 的文件：
- `internal/eino/runner/adk_runner.go` - ADK Runner 核心
- `internal/eino/agent/` 中 6 个 agent 文件全部基于 ADK
- `internal/eino/hooks/` 3 个 hook 全部基于 ADK
- `internal/eino/a2ui/streamer.go` - A2A 流式输出基于 ADK
- `cmd/server/main.go` - 注入 ADK Runner
- `internal/service/message_queue_service.go` - 调用 ADK Runner
- `internal/handler/a2a_sse.go` - SSE handler 基于 ADK

---

## 五、Legacy 降级链（需删除的路径）

message_queue_service.go 中的降级链：

```
1. ADK Runner → 成功则返回 ✓ (保留)
2. Eino streaming (generateEinoResponseStream) → 成功则返回 ✓ (保留)
3. Eino sync (generateEinoResponse) → 总是返回 nil ✗ (删除)
4. Legacy AI (ai_orchestrator.go) → 始终返回 nil ✗ (删除)
```

**只需保留步骤 1-2，删除 3-4 的死代码路径。**

---

## 六、前端结构总览

| 源目录 | 文件数 | 分类 |
|---------|--------|------|
| `components/ai-elements/` | 10 | → `features/conversation/components/` |
| `components/conversation/` | 3 | → `features/conversation/components/` |
| `components/feature/` | 1 | → `features/match/components/` |
| `components/interview/` | 1 | → `features/confirmation/components/` |
| `components/layout/` | 4 | → `shared/ui/` |
| `components/offer/` | 1 | → `features/confirmation/components/` |
| `components/ui/` | 56 (含测试) | → `shared/ui/` |
| `components/FSMStatusBar.tsx` | 1 | → `features/conversation/components/` |
| `components/HumanConfirmModal.tsx` | 1 | → `features/confirmation/components/` |
| `hooks/` | 7 | → `features/*/hooks/` + `shared/hooks/` |
| `stores/` | 4 | → `features/*/stores/` + `shared/stores/` |
| `lib/` | 10 + 2 子目录 | → `shared/api/` |
| `types/` | 3 | → `shared/types/` |

---

## 七、测试文件清单 (24 个)

| 类型 | 数量 | 路径 |
|------|------|------|
| 单元测试 | 5 | `tests/unit/` |
| 集成测试 | 7 | `tests/integration/` |
| 内联测试 | 8 | `internal/*/**/*_test.go` |
| 包测试 | 3 | `pkg/*/*_test.go` |
| 测试工具 | 2 | `tests/testutil/` |

**编译状态:** `go vet ./...` ✅ 通过
