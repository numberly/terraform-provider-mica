---
phase: 03-file-based-policy-resources
plan: 02
subsystem: provider
tags: [nfs, export-policy, resource, data-source, tdd]
dependency_graph:
  requires: [03-01]
  provides: [flashblade_nfs_export_policy, flashblade_nfs_export_policy_rule, flashblade_nfs_export_policy data source]
  affects: [provider.go]
tech_stack:
  added: []
  patterns: [resource-with-import, composite-import-id, delete-guard, tdd-unit-tests]
key_files:
  created:
    - internal/provider/nfs_export_policy_resource.go
    - internal/provider/nfs_export_policy_data_source.go
    - internal/provider/nfs_export_policy_resource_test.go
    - internal/provider/nfs_export_policy_rule_resource.go
    - internal/provider/nfs_export_policy_rule_resource_test.go
  modified:
    - internal/provider/provider.go
decisions:
  - "NFS policy name has no RequiresReplace — rename is in-place via PATCH (existing project decision)"
  - "Delete guard calls ListNfsExportPolicyMembers; requires file-systems handler in delete test"
  - "Rule import uses composite ID policy_name/rule_index resolved via GetNfsExportPolicyRuleByIndex"
  - "readIntoState returns diag.Diagnostics (not interface{}) for clean caller composition"
  - "Security list mapped via types.ListValueFrom; empty list maps to types.ListValueMust with empty slice"
metrics:
  duration_minutes: 35
  completed_date: "2026-03-27"
  tasks_completed: 2
  files_created: 5
  files_modified: 1
---

# Phase 3 Plan 02: NFS Export Policy Resources Summary

**One-liner:** NFS export policy resource (in-place rename, delete guard), rule resource (composite import), and read-only data source — all backed by mock-server unit tests.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | NFS export policy resource + data source | 3ffa1d2 | nfs_export_policy_resource.go, nfs_export_policy_data_source.go, nfs_export_policy_resource_test.go |
| 2 | NFS export policy rule resource + provider registration | c23f1e2 | nfs_export_policy_rule_resource.go, nfs_export_policy_rule_resource_test.go, provider.go |

## What Was Built

**flashblade_nfs_export_policy resource** (`nfs_export_policy_resource.go`):
- Full CRUD with timeout support
- `name` has no RequiresReplace — rename is applied via PATCH in a single operation
- `enabled` defaults to `true` via `booldefault.StaticBool(true)`
- `version` is Computed but NOT UseStateForUnknown — it changes on every update
- Delete guard: calls `ListNfsExportPolicyMembers` and blocks with a clear diagnostic if attached
- ImportState: reads by name, initializes null timeouts object

**flashblade_nfs_export_policy data source** (`nfs_export_policy_data_source.go`):
- Read-only by name; surfaces id, name, enabled, is_local, policy_type, version

**flashblade_nfs_export_policy_rule resource** (`nfs_export_policy_rule_resource.go`):
- `policy_name` has RequiresReplace — rules belong to one policy and cannot be moved
- `name` and `index` are Computed (server-assigned); `name` is used for PATCH/DELETE targeting
- All optional/computed rule fields: access, client, permission, anonuid, anongid, atime, fileid_32bit, secure, security (list), required_transport_security
- PATCH uses `*string` for anonuid/anongid (API schema difference from GET integer)
- Composite import ID: `policy_name/rule_index` — resolved via `GetNfsExportPolicyRuleByIndex`

**Provider registration** (`provider.go`):
- Added `NewNfsExportPolicyResource`, `NewNfsExportPolicyRuleResource` to `Resources()`
- Added `NewNfsExportPolicyDataSource` to `DataSources()`

## Test Results

- 9 NFS-specific tests pass (Create, Update/rename, Delete, Import for both resource types + DataSource)
- 83 total tests pass — no regressions
- Delete test registers FileSystem handler to satisfy `ListNfsExportPolicyMembers` (GET /file-systems)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Delete test required file-systems mock handler**
- **Found during:** Task 1 GREEN phase
- **Issue:** `TestNfsExportPolicyResource_Delete` called `ListNfsExportPolicyMembers` which hits `GET /file-systems?filter=...`. No file-systems handler was registered, returning HTTP 404.
- **Fix:** Added `handlers.RegisterFileSystemHandlers(ms.Mux)` to the delete test setup.
- **Files modified:** `internal/provider/nfs_export_policy_resource_test.go`
- **Commit:** 3ffa1d2

**2. [Rule 1 - Bug] readIntoState interface mismatch**
- **Found during:** Task 2 compilation
- **Issue:** Initial `readIntoState` used a hand-crafted `interface{AddError; HasError; Append}` that was incompatible with `*diag.Diagnostics`. The `Append` method signature didn't match.
- **Fix:** Changed `readIntoState` to return `diag.Diagnostics` instead of taking a diags argument — cleaner and idiomatic for the framework.
- **Files modified:** `internal/provider/nfs_export_policy_rule_resource.go`
- **Commit:** c23f1e2

**3. [Rule 1 - Bug] Test used non-existent `ValueBigFloat()` on types.Int64**
- **Found during:** Task 2 test compilation
- **Issue:** Test called `createdModel.Index.ValueBigFloat().String()` but `types.Int64` has `ValueInt64()` not `ValueBigFloat()`.
- **Fix:** Changed to `strconv.FormatInt(createdModel.Index.ValueInt64(), 10)`.
- **Files modified:** `internal/provider/nfs_export_policy_rule_resource_test.go`
- **Commit:** c23f1e2

## Self-Check: PASSED

- nfs_export_policy_resource.go: FOUND
- nfs_export_policy_data_source.go: FOUND
- nfs_export_policy_rule_resource.go: FOUND
- commit 3ffa1d2 (Task 1): FOUND
- commit c23f1e2 (Task 2): FOUND
