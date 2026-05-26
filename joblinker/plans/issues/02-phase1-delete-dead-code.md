# Phase 1: 删除尘余代码 — 三轮删除

> Parent: 大重构计划

## What to build

根据 Phase 0 的诊断结果，安全删除确定不用的代码。分三轮执行，每轮后编译验证。

## Acceptance criteria

- [ ] **第一轮（安全删除）**：引用次数 = 0 的文件全部删除
- [ ] **第二轮（死代码）**：`generateEinoResponse` 等空 stub、空函数体、未调用常量/变量删除
- [ ] **第三轮（Legacy 路径）**：ADK runner、Legacy fallback 降级链、旧 AI 编排逻辑删除
- [ ] 每轮删除后 `cd backend && go build ./...` 通过
- [ ] 每轮删除后 `cd frontend && npm run build` 通过
- [ ] 编译不过时回滚并检查遗漏引用

## Blocked by

- #01 Phase 0: 精确诊断
