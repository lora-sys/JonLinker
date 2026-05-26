# Phase 0: 精确诊断 — 生成三份清单

> Parent: 大重构计划

## What to build

在执行任何删除或移动之前，使用工具扫描代码库，生成数据驱动的决策依据。产出三份清单供后续 Phase 使用。

## Acceptance criteria

- [ ] `go vet` + `staticcheck` 扫描所有 `.go` 文件，输出未使用代码列表
- [ ] `grep -r "eino" backend/` 生成 Eino 使用程度报告
- [ ] `grep -r "legacy\|fallback\|mock\|generateEinoResponse" backend/` 生成 Legacy 代码位置清单
- [ ] 统计 `import.*from.*components` 生成前端未使用组件列表
- [ ] `find . -name "*_test.go"` 生成测试文件清单
- [ ] 对 `internal/model/` 每个类型，统计引用次数
- [ ] 对 `internal/repository/` 每个文件，统计引用次数
- [ ] **生成三份清单**：删除清单（引用=0）、迁移清单（需移动）、保留清单（看似冗余实际在用）
- [ ] 清单记录在文档中供 Phase 1 使用

## Blocked by

None - can start immediately
