# Implementation Plan: Backend Refactor — Fix 22 Issues

**Branch**: `019-backend-refactor-22fixes` | **Date**: 2026-05-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/019-backend-refactor-22fixes/spec.md`

## Summary

Refactor the JobLinker backend to eliminate hardcoded values, wire missing AI integrations, fix Chroma collection resolution, complete the database schema, remove dead code, and establish trustworthy integration tests. 7 phases, 25 functional requirements, TDD-driven (Red-Green-Refactor for every change).

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: Gin (HTTP), GORM (ORM), cloudwego/eino (AI agent framework), gorilla/websocket, protobuf
**Storage**: PostgreSQL 15+ (primary), Chroma (vector embeddings)
**Testing**: `go test` with `//go:build integration` tag for integration tests
**Target Platform**: Linux server (containerized)
**Project Type**: Web service (REST API + WebSocket)
**Performance Goals**: Standard web service (< 200ms p95 for reads)
**Constraints**: Must compile after each phase; no breaking changes to existing API contracts
**Scale/Scope**: ~30 Go source files modified/created, 7 integration test files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | Every phase follows Red-Green-Refactor. Integration tests require live services. |
| II. Code Quality Gates | PASS | No magic strings (constants centralized), single responsibility maintained, static typing enforced |
| III. User Experience Consistency | N/A | Backend-only changes, no UI impact |
| IV. Performance Requirements | N/A | No performance-critical paths changed |
| V. Observability & Debugging | PASS | Structured logging maintained; API key length log removed (security improvement) |
| Security Requirements | PASS | API key logging removed; no secrets in code |
| AI Model Configuration | PASS | All AI calls use env vars (AI_API_KEY, AI_BASE_URL, AI_MODEL) |
| Data Management | PASS | AutoMigrate additions are backward-compatible |

**Gate Result**: PASS — all applicable principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/019-backend-refactor-22fixes/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
joblinker/backend/
├── cmd/server/main.go                          # Phase 5: AutoMigrate additions, Phase 7: route fixes
├── pkg/
│   ├── ai/client.go                            # Phase 7: remove GenerateEmbedding
│   └── chroma/client.go                        # Phase 1: EnsureCollection + name→UUID mapping
├── internal/
│   ├── agent/memory.go                         # Phase 7: DELETE
│   ├── handler/message.go                      # Phase 7: Protobuf timestamp fix
│   ├── service/
│   │   ├── conversation_service.go             # Phase 2: fix extraction functions
│   │   ├── dual_agent_negotiation_service.go   # Phase 3: AI integration
│   │   ├── reasoning_service.go                # Phase 3: AI integration
│   │   └── message_queue_service.go            # Phase 4: remove hardcoded values, Phase 7: remove API key log
│   └── model/                                  # Phase 5: new models (if not already defined)
└── tests/
    ├── testutil/                               # Phase 6: helpers (already exists, extend if needed)
    └── integration/                            # Phase 6: new/updated integration tests
```

**Structure Decision**: Existing Go backend structure. All changes are in-place modifications or deletions. No new top-level directories.

## Complexity Tracking

No constitution violations to justify. All changes follow existing patterns.
