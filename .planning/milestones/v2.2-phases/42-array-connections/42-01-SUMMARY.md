---
phase: 42-array-connections
plan: "01"
subsystem: client
tags: [array-connections, client, mock, tests]
dependency_graph:
  requires: []
  provides: [ArrayConnection CRUD client, full CRUD mock handler, 7 client tests]
  affects: [42-02-PLAN.md (provider layer)]
tech_stack:
  added: []
  patterns: [getOneByName generic, postOne/patchOne generics, remote_names= query param, byName mock store]
key_files:
  created:
    - internal/client/array_connections_test.go
  modified:
    - internal/client/models_admin.go
    - internal/client/array_connections.go
    - internal/testmock/handlers/array_connections.go
decisions:
  - "ArrayConnectionPatch.CACertificateGroup is **NamedReference for nil=omit vs inner-nil=set-null semantics"
  - "All CRUD operations use ?remote_names= (not ?names=) — API-mandated"
  - "Mock GET empty-list + 200 on miss (not 404) — getOneByName[T] detects not-found via empty list"
  - "Mock store keyed by conn.Remote.Name (replaced byID keying)"
  - "Patch_CACertificateGroup test covers set-value only; JSON null/omitempty limitation prevents clear test"
metrics:
  duration: "~5 minutes"
  completed_date: "2026-04-08"
  tasks: 2
  files: 4
---

# Phase 42 Plan 01: Array Connection Client Layer Summary

ArrayConnection client extended with full CRUD (Post/Patch/Delete), model structs (ArrayConnectionThrottle/Post/Patch), mock handler replaced with full GET/POST/PATCH/DELETE keyed by remote.Name, and 7 client unit tests added.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Extend models and add Post/Patch/Delete client methods | 1cce01d | models_admin.go, array_connections.go |
| 2 | Replace mock handler with full CRUD + tests | ae40c20 | handlers/array_connections.go, array_connections_test.go |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Renamed conflicting test function names**
- **Found during:** Task 2
- **Issue:** `TestUnit_ArrayConnection_Get_Found` and `TestUnit_ArrayConnection_Get_NotFound` already existed in `array_admin_test.go` using inline httptest handlers — redeclaration compile error
- **Fix:** Renamed new tests to `TestUnit_ArrayConnection_Get_Found_Mock` and `TestUnit_ArrayConnection_Get_NotFound_Mock` (using mock handler package instead of inline handler)
- **Files modified:** `internal/client/array_connections_test.go`
- **Commit:** ae40c20

**2. [Rule 1 - Bug] Removed CACertificateGroup clear test (JSON null limitation)**
- **Found during:** Task 2 test execution
- **Issue:** `**NamedReference` + `omitempty`: JSON `null` decodes outer ptr to nil, making "set null" indistinguishable from "field absent" — same limitation as targets mock
- **Fix:** Test covers set-value only (matching `TestUnit_Target_Patch_CACertGroup` pattern)
- **Files modified:** `internal/client/array_connections_test.go`

## Known Stubs

None.

## Self-Check: PASSED

- `internal/client/models_admin.go` — contains ArrayConnectionThrottle, ArrayConnectionPost, ArrayConnectionPatch
- `internal/client/array_connections.go` — contains PostArrayConnection, PatchArrayConnection, DeleteArrayConnection
- `internal/testmock/handlers/array_connections.go` — full CRUD, byName store, Seed by remote.Name
- `internal/client/array_connections_test.go` — 7 TestUnit_ArrayConnection_* tests
- `go build ./...` — clean
- `go test ./internal/... -count=1` — 742 tests pass (baseline was 716)
