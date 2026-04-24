---
phase: 02-object-store-resources
plan: "02"
subsystem: bucket
tags: [resource, data-source, crud, soft-delete, tdd, mock-handler]
dependency_graph:
  requires:
    - 02-01  # models.go (Bucket, BucketPost, BucketPatch), account client, mock helpers
  provides:
    - flashblade_bucket resource with full CRUD + Import + soft-delete
    - flashblade_bucket data source
    - internal/client/buckets.go CRUD methods
    - internal/testmock/handlers/buckets.go full mock handler
  affects:
    - internal/provider/provider.go (adds NewBucketResource + NewBucketDataSource)
    - internal/provider/object_store_account_resource.go (uses ListBuckets for DELETE guard)
tech_stack:
  added: []
  patterns:
    - Two-phase soft-delete (PatchBucket destroyed=true, then optional DeleteBucket+poll)
    - ForceNew on name + account via RequiresReplace
    - destroy_eradicate_on_delete defaults false (production data safety, inverse of filesystem)
    - PATCH uses raw map[string]json.RawMessage for true partial-update semantics
    - Cross-reference account store in mock handler POST validation
    - TDD RED/GREEN/REFACTOR via failing tests first
key_files:
  created:
    - internal/testmock/handlers/buckets.go
    - internal/provider/bucket_resource.go
    - internal/provider/bucket_resource_test.go
    - internal/provider/bucket_data_source.go
    - internal/provider/bucket_data_source_test.go
  modified:
    - internal/client/buckets.go
    - internal/provider/provider.go
    - internal/provider/object_store_account_resource_test.go
decisions:
  - "destroy_eradicate_on_delete defaults false for buckets (vs true for filesystems) — buckets hold production S3 data, eradication must be opt-in"
  - "Name and account are RequiresReplace (ForceNew) — S3 clients hardcode bucket names, account is immutable post-creation"
  - "Non-empty bucket guard checks ObjectCount from state before soft-delete — prevents accidental data loss"
  - "Mock handler cross-references account store in POST — validates account existence before creating bucket"
  - "ListBuckets Destroyed filter added to support PollBucketUntilEradicated query pattern"
metrics:
  duration_seconds: 488
  completed_date: "2026-03-26"
  tasks_completed: 2
  files_created: 5
  files_modified: 3
---

# Phase 02 Plan 02: Bucket Resource and Data Source Summary

**One-liner:** Bucket resource with two-phase soft-delete (eradicate=false default), account cross-reference mock, ForceNew on name+account, drift logging, and full TDD coverage.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Bucket client CRUD + mock handler | ebc4ea8 | buckets.go (client), buckets.go (handlers) |
| 2 | Bucket resource + data source + tests | 443505e | bucket_resource.go, bucket_data_source.go, tests, provider.go |

## Artifacts Produced

| Path | Exports | Notes |
|------|---------|-------|
| `internal/client/buckets.go` | `GetBucket, ListBuckets, PostBucket, PatchBucket, DeleteBucket, PollBucketUntilEradicated` | Full CRUD with Destroyed filter |
| `internal/testmock/handlers/buckets.go` | `RegisterBucketHandlers` | Account cross-reference on POST, raw PATCH, eradication guard on DELETE |
| `internal/provider/bucket_resource.go` | `NewBucketResource` | 310 lines, two-phase soft-delete, ForceNew on name+account, drift tflog |
| `internal/provider/bucket_data_source.go` | `NewBucketDataSource` | Reads by name, maps all fields including account ref |

## Decisions Made

- **destroy_eradicate_on_delete defaults false for buckets**: Inverse of filesystem resource default. S3 buckets hold production object data — eradication is irreversible and must be an explicit choice.
- **Name + account RequiresReplace (ForceNew)**: S3 clients embed bucket names in ARNs and policies; rename is not possible. Account is immutable post-creation per FlashBlade semantics.
- **Non-empty bucket guard on Delete**: Checks `ObjectCount > 0` from state before soft-delete. Provides a clear error "Bucket contains X objects. Empty the bucket before deleting it."
- **Mock RegisterBucketHandlers signature changed**: Now accepts `*objectStoreAccountStore` for cross-reference. Updated `object_store_account_resource_test.go` to pass the store pointer.

## Test Coverage

9 tests pass (TDD RED/GREEN):
- `TestUnit_Bucket_Create` — POST creates bucket with name + account, returns id + created
- `TestUnit_Bucket_Update` — PATCH updates quota_limit, versioning, hard_limit_enabled
- `TestUnit_Bucket_Destroy` — soft-delete only (destroy_eradicate_on_delete=false default)
- `TestUnit_Bucket_Destroy_WithEradicate` — full eradication when destroy_eradicate_on_delete=true
- `TestUnit_Bucket_Import` — import by name populates all attributes including account ref
- `TestUnit_Bucket_DriftLog` — Read logs field-level diffs via tflog, corrects state to API values
- `TestUnit_Bucket_NonEmptyDelete` — ObjectCount>0 returns diagnostic blocking deletion
- `TestUnit_BucketDataSource` — data source reads bucket by name with full attribute map
- `TestUnit_BucketDataSource_NotFound` — returns error diagnostic for missing bucket

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] RegisterBucketHandlers signature change broke existing test**
- **Found during:** Task 1 implementation — changed stub to full CRUD with account cross-reference
- **Issue:** `object_store_account_resource_test.go` called `handlers.RegisterBucketHandlers(ms.Mux)` (no accounts param)
- **Fix:** Updated call site to `accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)` + `handlers.RegisterBucketHandlers(ms.Mux, accountStore)`
- **Files modified:** `internal/provider/object_store_account_resource_test.go`
- **Commit:** ebc4ea8

**2. [Scope note] Access key provider tests pre-exist as build failures**
- Pre-existing build failures in `object_store_access_key_*_test.go` reference undefined types from plan 02-03 (not yet implemented). These are out of scope and unchanged by this plan.

## Self-Check: PASSED

Files created exist:
- FOUND: internal/client/buckets.go
- FOUND: internal/testmock/handlers/buckets.go
- FOUND: internal/provider/bucket_resource.go
- FOUND: internal/provider/bucket_resource_test.go
- FOUND: internal/provider/bucket_data_source.go
- FOUND: internal/provider/bucket_data_source_test.go

Commits exist:
- FOUND: ebc4ea8 (Task 1)
- FOUND: 443505e (Task 2)

Build: `go build ./internal/...` — SUCCESS
Tests: `go test ./internal/provider/ -run TestUnit_Bucket` — 9 PASSED
