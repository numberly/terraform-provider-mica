---
phase: 01-foundation
plan: "04"
subsystem: provider-resources
tags: [resource, data-source, filesystem, crud, import, soft-delete, drift, timeouts, tdd]
dependency_graph:
  requires: [01-02, 01-03]
  provides: [flashblade_file_system resource, flashblade_file_system data source]
  affects: [provider registration, all future resource plans as template]
tech_stack:
  added:
    - github.com/hashicorp/terraform-plugin-framework-timeouts v0.7.0
  patterns:
    - TDD red-green for each task
    - Read-at-end-of-write pattern for Create and Update
    - Two-phase soft-delete + eradicate with polling
    - Drift detection via tflog.Info structured logging
    - Per-resource timeouts.Attributes() for create/read/update/delete
    - Import by name via GetFileSystem
key_files:
  created:
    - internal/provider/filesystem_resource.go
    - internal/provider/filesystem_resource_test.go
    - internal/provider/filesystem_data_source.go
    - internal/provider/filesystem_data_source_test.go
    - examples/provider/provider.tf
    - examples/resources/flashblade_file_system/resource.tf
    - examples/data-sources/flashblade_file_system/data-source.tf
  modified:
    - internal/provider/provider.go
    - internal/testmock/server.go
    - go.mod
    - go.sum
decisions:
  - "timeouts implemented as SingleNestedAttribute (Attributes() not Block()) — consistent with provider auth pattern"
  - "Optional/Computed blocks (nfs, smb, multi_protocol, default_quotas) use SingleNestedBlock — allows nil pointer in model when absent"
  - "destroy_eradicate_on_delete defaults true via booldefault.StaticBool — matches ops-team expectation of clean teardown"
  - "ImportState initializes timeouts.Value with types.ObjectNull to satisfy timeouts.Type custom serialization"
  - "Data source uses separate model structs (not shared with resource) — avoids timeouts field carrying over to data source schema"
metrics:
  duration_minutes: 158
  completed_date: "2026-03-27"
  tasks_completed: 2
  files_created: 7
  files_modified: 4
---

# Phase 01 Plan 04: flashblade_file_system Resource and Data Source Summary

**One-liner:** flashblade_file_system resource with full CRUD, two-phase soft-delete, per-resource timeouts, drift detection, and import by name, plus matching data source.

## Tasks Completed

| # | Name | Commit | Key Files |
|---|------|--------|-----------|
| 1 | filesystem_resource full CRUD | 40888f0 | filesystem_resource.go, filesystem_resource_test.go, provider.go |
| 2 | filesystem_data_source + examples | d7cac79 | filesystem_data_source.go, filesystem_data_source_test.go, examples/ |
| dep | Add timeouts dependency | f514e30 | go.mod, go.sum |

## What Was Built

### flashblade_file_system Resource (`internal/provider/filesystem_resource.go`)

- **Create:** Builds `FileSystemPost` from plan (name, provisioned, NFS, SMB), calls `PostFileSystem`, then Read-at-end-of-write via `readIntoState`.
- **Read:** Calls `GetFileSystem` by name. If 404 → `RemoveResource`. Compares `provisioned` against state and emits `tflog.Info` drift log with structured fields (resource, field, state_value, api_value). Maps all API fields to model.
- **Update:** Computes diff between plan and state, sends only changed fields via `PatchFileSystem`. Supports in-place rename (no RequiresReplace on `name`). Read-at-end-of-write.
- **Delete:** Phase 1 soft-delete (`PATCH destroyed=true`). If `destroy_eradicate_on_delete=true` (default): Phase 2 `DeleteFileSystem` + Phase 3 `PollUntilEradicated`. All phases use the delete timeout context.
- **ImportState:** Accepts name as import ID, calls `GetFileSystem`, populates full state including null-initialized `timeouts.Value`.
- **Timeouts:** `timeouts.Attributes(ctx, Opts{Create, Read, Update, Delete: true})` with defaults 20m/5m/20m/30m.
- **Schema:** `id`, `name`, `provisioned`, `destroyed`, `destroy_eradicate_on_delete`, `time_remaining`, `created`, `promotion_status`, `writable`, `nfs_export_policy`, `smb_share_policy` as flat attributes. `space`, `nfs`, `smb`, `http`, `multi_protocol`, `default_quotas`, `source` as SingleNestedBlock.

### flashblade_file_system Data Source (`internal/provider/filesystem_data_source.go`)

- **Read:** Gets `name` from config, calls `GetFileSystem`. If 404 → diagnostic "File system not found". Maps all API fields to model.
- **Schema:** All attributes Computed except `name` (Required). No timeouts block. Separate model structs to avoid coupling.

### Example HCL (`examples/`)

- `examples/provider/provider.tf` — provider block with endpoint and auth.api_token.
- `examples/resources/flashblade_file_system/resource.tf` — resource with NFS block and timeouts.
- `examples/data-sources/flashblade_file_system/data-source.tf` — data source read with output.

## Test Results

All 22 tests pass:
- 13 `TestUnit_FileSystem_*` tests covering Create, CreateWithNFS, CreateWithSMB, Read, ReadDestroyed, ReadNotFound, Update, UpdateRename, Destroy, DestroySoftOnly, Import, DriftLog, Idempotent.
- 2 `TestUnit_FileSystemDataSource_*` tests covering Read and NotFound.
- 7 existing `TestUnit_Provider*` tests unchanged.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed testmock server login path**
- **Found during:** Task 1 (all tests failing with "unexpected status 404")
- **Issue:** `MockServer` registered `/login` but `LoginWithAPIToken` calls `{endpoint}/api/login`
- **Fix:** Changed `mux.HandleFunc("/login", ...)` to `mux.HandleFunc("/api/login", ...)` in `internal/testmock/server.go`
- **Files modified:** `internal/testmock/server.go`
- **Commit:** 40888f0

**2. [Rule 1 - Bug] Fixed timeouts.Value initialization in ImportState**
- **Found during:** Task 1 (`TestUnit_FileSystem_Import` failing with type conversion error)
- **Issue:** Zero-value `timeouts.Value{}` with empty `types.Object{}` is not valid for `timeouts.Type` serialization
- **Fix:** Initialize `data.Timeouts` with `types.ObjectNull(attrTypes)` in `ImportState` before `resp.State.Set`
- **Files modified:** `internal/provider/filesystem_resource.go`
- **Commit:** 40888f0

## Self-Check: PASSED

All created files found on disk. All commits (f514e30, 40888f0, d7cac79) confirmed in git log.
