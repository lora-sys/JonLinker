# Phase 2.5: 组装 main.go + 删除旧目录

> Parent: 大重构计划

## What to build

重写 `cmd/server/main.go` 按四层架构顺序初始化，然后删除旧目录 `internal/{service,agent,handler,model,repository,cache,config}` 和精简 `pkg/`。

## Acceptance criteria

- [ ] `cmd/server/main.go` 按 core → adapters → engine → transport 顺序初始化
- [ ] 删除旧 import，只保留新架构的 import
- [ ] 启动 server 无错误
- [ ] 删除 `internal/{service,agent,handler,model,repository,cache,config}`
- [ ] 精简 `pkg/` 只保留 `pkg/shared/`
- [ ] `go build ./...` 通过
- [ ] 旧目录全部消失

## Blocked by

- #06 Phase 2.4: 重写 transport 层
