# Phase 4.2: E2E 测试 + 最终验证

> Parent: 大重构计划

## What to build

新增完整的 A2A 流程 Playwright E2E 测试，覆盖用户注册 → 创建 Agent → 匹配 → 对话 → Offer → 确认全过程。`docker-compose up` 一键启动通过，后端日志无 panic/error。

## E2E 测试流程

文件：`frontend/tests/e2e/a2a-full-flow.spec.ts`

```
流程: 用户注册 → 创建 Seeker Agent → 创建 Recruiter Agent → 发起匹配
     → A2A 对话 → Offer 生成 → 人类确认
断言: 每一步的 UI 状态正确，后端日志无 panic/error
```

## Acceptance criteria

- [ ] `frontend/tests/e2e/a2a-full-flow.spec.ts` 创建并覆盖完整 A2A 流程
- [ ] `docker-compose up` 一键启动通过
- [ ] `npx playwright test tests/e2e/a2a-full-flow.spec.ts` 通过
- [ ] 后端日志无 panic / error
- [ ] README 架构图更新，新人 5 分钟理解

## Blocked by

- #08 Phase 3: Frontend 重组
- #09 Phase 4.1: 修复 11 个关键 bug
