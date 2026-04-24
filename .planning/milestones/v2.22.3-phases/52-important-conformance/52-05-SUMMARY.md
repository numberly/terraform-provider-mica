---
phase: 52-important-conformance
plan: 05
subsystem: testing
tags: [conformance, tests, rename, R-010]
requires: []
provides:
  - "TestUnit_ prefix conformance across internal/"
affects:
  - internal/client/transport_internal_test.go
  - internal/provider/object_store_user_resource_test.go
  - internal/provider/object_store_user_policy_resource_test.go
  - internal/provider/object_store_user_data_source_test.go
tech-stack:
  added: []
  patterns:
    - "TestUnit_<Resource>_<Operation>[_<Variant>] mandatory prefix (CONVENTIONS.md §Test Conventions)"
key-files:
  created: []
  modified:
    - internal/client/transport_internal_test.go
    - internal/provider/object_store_user_resource_test.go
    - internal/provider/object_store_user_policy_resource_test.go
    - internal/provider/object_store_user_data_source_test.go
decisions:
  - "D-52-05: pure rename (no body/move) of 6 test functions"
metrics:
  duration: "~2 min"
  completed: 2026-04-20
  tasks_completed: 1
  files_modified: 4
  tests_total: 775
  tests_delta: 0
---

# Phase 52 Plan 05: Test naming conformance (R-010) Summary

Pure rename of 6 tests to the mandatory `TestUnit_<Resource>_<Operation>` prefix — no body changes, no file moves.

## What changed

6 function declarations renamed across 4 files:

| Old | New | File |
|-----|-----|------|
| `TestComputeDelayJitter` | `TestUnit_Transport_ComputeDelayJitter` | `internal/client/transport_internal_test.go` |
| `TestComputeDelayCap` | `TestUnit_Transport_ComputeDelayCap` | `internal/client/transport_internal_test.go` |
| `TestMocked_ObjectStoreUser_Lifecycle` | `TestUnit_ObjectStoreUserResource_Lifecycle` | `internal/provider/object_store_user_resource_test.go` |
| `TestMocked_ObjectStoreUser_FullAccess` | `TestUnit_ObjectStoreUserResource_FullAccess` | `internal/provider/object_store_user_resource_test.go` |
| `TestMocked_ObjectStoreUserPolicy_Lifecycle` | `TestUnit_ObjectStoreUserPolicyResource_Lifecycle` | `internal/provider/object_store_user_policy_resource_test.go` |
| `TestMocked_ObjectStoreUser_DataSource` | `TestUnit_ObjectStoreUserDataSource_Basic` | `internal/provider/object_store_user_data_source_test.go` |

Doc comments immediately above each function were also updated to reference the new name (the only adjacent change permitted by the plan).

## Decisions made

- Applied **D-52-05** as specified: pure rename, zero behavioral change.

## Deviations from Plan

None - plan executed exactly as written.

## Authentication gates

None.

## Verification

- `make test` - PASS, 775 total tests (unchanged by this plan; matches 6 renames / 0 delta).
- `make lint` - clean, 0 issues.
- Sentinel (negative): `rg -E '^func (TestMocked_|TestCompute)' internal/` - 0 results.
- Sentinel (positive): `rg '^func TestUnit_(Transport_ComputeDelay|ObjectStoreUser(Resource|Policy|DataSource))' internal/` - 6 results.

## Success criteria

- [x] R-010 closed: no test in `internal/` uses `TestMocked_` or `TestCompute` prefixes.
- [x] 6 renames, 0 behavioral changes.
- [x] Baseline test count preserved.

## Commits

- `c2e2f3e` refactor(52-05): rename tests to TestUnit_ prefix

## Self-Check: PASSED

- FOUND: internal/client/transport_internal_test.go (TestUnit_Transport_ComputeDelayJitter, TestUnit_Transport_ComputeDelayCap)
- FOUND: internal/provider/object_store_user_resource_test.go (TestUnit_ObjectStoreUserResource_Lifecycle, TestUnit_ObjectStoreUserResource_FullAccess)
- FOUND: internal/provider/object_store_user_policy_resource_test.go (TestUnit_ObjectStoreUserPolicyResource_Lifecycle)
- FOUND: internal/provider/object_store_user_data_source_test.go (TestUnit_ObjectStoreUserDataSource_Basic)
- FOUND commit: c2e2f3e
