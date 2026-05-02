# Implementation Plan: Protobuf Communication Protocol

**Branch**: `014-protobuf-protocol` | **Date**: 2026-05-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/014-protobuf-protocol/spec.md`

## Summary

Add Protobuf serialization to JobLinker's communication channels to reduce bandwidth usage and improve performance. The feature delivers 60-90% payload size reduction through binary encoding and cache-key optimization for tool results. Implementation follows a 4-phase rollout: dual-protocol support, internal services first, WebSocket agent dialogue second, then full production switch.

## Technical Context

**Language/Version**: Go 1.21+ (backend), TypeScript 5.7+ (frontend Next.js 16)  
**Primary Dependencies**: protoc (v25+), protoc-gen-go (v1.31+), protoc-gen-ts (ts-proto v1.146+), existing cache layer (in-memory LRU)  
**Storage**: N/A (protocol change only, no persistence changes)  
**Testing**: go test ./... (backend), vitest/jest (frontend), agent-browser for E2E  
**Target Platform**: Linux server (backend), modern browsers (frontend)  
**Project Type**: Web application (Next.js + Go backend)  
**Performance Goals**: 60%+ bandwidth reduction for all Protobuf channels, 90%+ reduction for tool result messages  
**Constraints**: Zero downtime migration, backward compatibility with JSON, no functional regression  
**Scale/Scope**: 4 .proto files, 2 language targets, ~15 integration points

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| Test-First Development | ✅ PASS | Integration tests for protocol negotiation; unit tests for serialization |
| Code Quality Gates | ✅ PASS | Generated code is type-safe; no `any` types in TS output |
| User Experience Consistency | ✅ PASS | Protocol change is invisible to users |
| Performance Requirements | ✅ PASS | Direct performance improvement (60-90% reduction) |
| Observability & Debugging | ✅ PASS | Add metrics for protobuf vs JSON message counts/sizes |
| Security Requirements | ✅ PASS | No secrets; binary serialization doesn't introduce new vectors |
| AI Model Configuration | N/A | No AI changes |

**Gate Result**: PASS — proceed to Phase 0

## Project Structure

### Documentation (this feature)


### Source Code (repository root)


**Structure Decision**: Protobuf schemas placed in top-level `proto/` directory (shared between frontend/backend). Generated code lives in respective language-specific directories. Existing code modified minimally to add protocol negotiation.

## Complexity Tracking

> No constitution violations — this feature reduces complexity (smaller messages, cleaner cache pattern) rather than adding it.
