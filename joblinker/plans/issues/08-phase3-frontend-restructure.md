# Phase 3: Frontend 按 feature 重组

> Parent: 大重构计划

## What to build

将前端代码按 feature 重组，只做文件移动 + import 路径更新，不改组件逻辑。

## Acceptance criteria

- [ ] `features/conversation/` 创建并移动 AI 对话相关组件/hooks/stores
- [ ] `features/agent/` 创建并移动 Agent 相关组件/hooks/stores
- [ ] `features/match/` 创建并移动匹配相关组件/hooks/stores
- [ ] `features/confirmation/` 创建并移动面试/Offer/确认相关组件/hooks
- [ ] `shared/` 重组：`shared/ui/`, `shared/hooks/`, `shared/stores/`, `shared/api/`, `shared/types/`
- [ ] 删除旧的 `components/`、`hooks/`、`stores/` 根目录
- [ ] `npm run build` 通过

## Blocked by

- #02 Phase 1: 删除尘余代码
