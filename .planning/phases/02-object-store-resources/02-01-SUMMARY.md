---
phase: 02-object-store-resources
plan: "01"
subsystem: object-store-accounts
tags: [terraform, flashblade, object-store, resource, data-source, client, mock]
dependency_graph:
  requires: []
  provides:
    - flashblade_object_store_account resource (Create/Read/Update/Delete/Import)
    - flashblade_object_store_account data source (Read by name)
    - client.ObjectStoreAccount, ObjectStoreAccountPost, ObjectStoreAccountPatch model structs
    - client.Bucket, BucketPost, BucketPatch model structs (for Plans 02-02/02-03)
    - client.ObjectStoreAccessKey, ObjectStoreAccessKeyPost model structs (for Plan 02-03)
    - client.NamedReference struct
    - client.ListBuckets, GetBucket methods (prerequisite for account bucket guard)
    - handlers.WriteJSONListResponse, WriteJSONError generic helpers
    - handlers.RegisterObjectStoreAccountHandlers mock handler
    - handlers.RegisterBucketHandlers stub handler
  affects:
    - internal/provider/provider.go (Resources + DataSources registration)
    - internal/testmock/handlers/filesystems.go (migrated to generic helpers)
tech_stack:
  added:
    - objectStoreAccountResource (terraform-plugin-framework resource)
    - objectStoreAccountDataSource (terraform-plugin-framework data source)
    - objectStoreAccountModel (tfsdk-tagged struct)
    - objectStoreAccountDataSourceModel (tfsdk-tagged struct)
  patterns:
    - Same client CRUD pattern as filesystems.go (GET by ?names= query param)
    - Same mock handler pattern as handlers/filesystems.go
    - Same TDD pattern as Phase 1 (tfsdk.Plan/State struct-based unit tests)
    - Name is query param on POST (not in body) — matches FlashBlade object store API
    - Single-phase DELETE (no soft-delete) with bucket-existence guard
    - RequiresReplace on name attribute (no rename support for accounts)
key_files:
  created:
    - internal/client/object_store_accounts.go
    - internal/client/buckets.go
    - internal/testmock/handlers/helpers.go
    - internal/testmock/handlers/object_store_accounts.go
    - internal/testmock/handlers/buckets.go
    - internal/provider/object_store_account_resource.go
    - internal/provider/object_store_account_resource_test.go
    - internal/provider/object_store_account_data_source.go
    - internal/provider/object_store_account_data_source_test.go
  modified:
    - internal/client/models.go (added all Phase 2 model structs)
    - internal/testmock/handlers/filesystems.go (use generic helpers)
    - internal/provider/provider.go (added account resource + data source registration)
decisions:
  - "Object store account name passed as ?names= query param on POST (not in body) — matches FlashBlade API design"
  - "Single-phase DELETE with no soft-delete for accounts — per plan spec and FlashBlade behavior"
  - "Bucket guard in Delete calls ListBuckets with ?account_names= filter — prevents orphaned buckets"
  - "All Phase 2 model structs added in this plan — Bucket, BucketPost/Patch, ObjectStoreAccessKey, ObjectStoreAccessKeyPost — Plans 02-02/03 won't need to touch models.go"
  - "client/buckets.go added as prerequisite for the bucket guard; Plan 02-02 will expand it with full CRUD"
  - "handlers.WriteJSONListResponse/WriteJSONError extracted as generic package-level helpers — filesystems.go migrated"
  - "handlers.RegisterBucketHandlers stub returns empty list — prevents 404 in Delete test; Plan 02-02 replaces it with full handler"
metrics:
  duration_seconds: 576
  completed_date: "2026-03-27"
  tasks_completed: 2
  files_created: 9
  files_modified: 3
---

# Phase 2 Plan 01: Object Store Account Resource and Data Source Summary

**One-liner:** JWT-free object store account CRUD with single-phase delete, bucket guard, and RequiresReplace on name using pure Go client + mock handler pattern from Phase 1.

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Model structs + account client CRUD + mock helpers + mock handler | cef822c | models.go, object_store_accounts.go, helpers.go, object_store_accounts.go (handler), filesystems.go |
| 2 | Resource + data source + provider registration + tests (TDD) | e1d996b | object_store_account_resource.go, _test.go, object_store_account_data_source.go, _test.go, provider.go, buckets.go (client), buckets.go (handler) |

## Verification Results

- `go build ./internal/...` — PASS
- `go test ./internal/provider/ -run "TestUnit_ObjectStoreAccount|TestUnit_ObjectStoreAccountDataSource" -count=1` — 8/8 PASS
- `go test ./internal/... -count=1 -timeout 60s` — 59/59 PASS (no Phase 1 regressions)

## Decisions Made

1. **Name as query param on POST**: FlashBlade API passes account name as `?names=` query param on POST, not in the body. Client method `PostObjectStoreAccount(ctx, name string, body ObjectStoreAccountPost)` reflects this correctly.

2. **Single-phase DELETE**: Object store accounts have no soft-delete cycle (unlike file systems). `DeleteObjectStoreAccount` calls `DELETE /object-store-accounts?names=` directly.

3. **Bucket guard on Delete**: Before calling DELETE, `Delete()` calls `ListBuckets` with `account_names=` filter. If any buckets exist, returns a diagnostic error rather than letting the API fail with an obscure error.

4. **All Phase 2 models in models.go**: Added Bucket, BucketPost, BucketPatch, ObjectStoreAccessKey, ObjectStoreAccessKeyPost, NamedReference structs in this plan so Plans 02-02 and 02-03 do not need to touch models.go.

5. **buckets.go client stub**: Added `ListBuckets` and `GetBucket` to unblock the account bucket guard. Plan 02-02 will add PostBucket, PatchBucket, DeleteBucket.

6. **Generic mock helpers**: `WriteJSONListResponse` and `WriteJSONError` extracted from filesystems.go as exported package-level helpers. All handlers now use these instead of local unexported functions.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added client/buckets.go for ListBuckets prerequisite**
- **Found during:** Task 2 implementation of the account Delete bucket guard
- **Issue:** `r.client.ListBuckets` and `client.ListBucketsOpts` undefined — Plan 02-02 would normally create these, but the account Delete guard needed them immediately
- **Fix:** Created `internal/client/buckets.go` with `ListBuckets` + `GetBucket` methods. Plan 02-02 will add `PostBucket`, `PatchBucket`, `DeleteBucket` to this file
- **Files modified:** `internal/client/buckets.go` (new)
- **Commit:** e1d996b

**2. [Rule 3 - Blocking] Added handlers/buckets.go stub for Delete test**
- **Found during:** Task 2 test execution — `TestUnit_ObjectStoreAccount_Delete` got HTTP 404 on `/api/2.22/buckets`
- **Issue:** No buckets endpoint registered in mock server → ListBuckets call returned 404 error
- **Fix:** Created `internal/testmock/handlers/buckets.go` stub that returns empty `[]Bucket{}` on GET. Plan 02-02 will replace with full handler
- **Files modified:** `internal/testmock/handlers/buckets.go` (new)
- **Commit:** e1d996b

**3. [Rule 1 - Bug] Fixed ForceNew test type assertion approach**
- **Found during:** Task 2 test RED phase
- **Issue:** `schema.Attribute` interface doesn't expose `GetPlanModifiers()` — had to use concrete `resschema.StringAttribute` type assertion to inspect `PlanModifiers` field
- **Fix:** Import `resschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"` and cast to `resschema.StringAttribute`
- **Files modified:** `internal/provider/object_store_account_resource_test.go`
- **Commit:** e1d996b

## Self-Check: PASSED

All key artifacts verified:
- FOUND: internal/client/object_store_accounts.go
- FOUND: internal/testmock/handlers/object_store_accounts.go
- FOUND: internal/provider/object_store_account_resource.go
- FOUND: internal/provider/object_store_account_data_source.go
- FOUND: commits cef822c and e1d996b in git history
- VERIFIED: 59 tests pass, 0 failures
