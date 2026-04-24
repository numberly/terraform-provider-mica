---
phase: 32-code-correctness-fixes
plan: "01"
subsystem: provider
tags: [terraform, go, flashblade, bucket, filesystem, diagnostics]

# Dependency graph
requires: []
provides:
  - ObjectLockConfig.FreezeLockedObjects field with correct spelling (JSON tag unchanged)
  - Bucket resource extract functions without unused ctx parameter
  - Filesystem schema without dead nfs_export_policy/smb_share_policy attributes
  - DiagnosticReporter interface with AddWarning method
  - readIntoState severity-preserving diagnostic dispatch
  - mapFSToModel using types.ObjectValue directly (no mustObjectValue wrapper)
affects: [33-client-hardening, 34-test-quality]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Extract functions take only the data they need — no ctx if unused"
    - "Severity-preserving diagnostic dispatch: check d.Severity() == diag.SeverityWarning"
    - "Inline types.ObjectValue directly — no passthrough wrappers"

key-files:
  created: []
  modified:
    - internal/client/models_storage.go
    - internal/provider/bucket_resource.go
    - internal/provider/bucket_resource_test.go
    - internal/provider/filesystem_resource.go
    - internal/provider/filesystem_resource_test.go
    - internal/provider/helpers.go

key-decisions:
  - "JSON tag freeze_locked_objects unchanged — only Go field name renamed to FreezeLockedObjects"
  - "DiagnosticReporter.AddWarning added — all existing callers use *diag.Diagnostics which already has AddWarning"
  - "nfs_export_policy and smb_share_policy removed — they were schema-only fields with no API backing in filesystem CRUD"

patterns-established:
  - "DiagnosticReporter interface: AddError, AddWarning, HasError — any new readIntoState must dispatch by severity"

requirements-completed: [CC-01, CC-02, CC-03, CH-03, CL-01]

# Metrics
duration: 15min
completed: 2026-03-31
---

# Phase 32 Plan 01: Code Correctness Fixes Summary

**Five mechanical correctness fixes: FreezeLockgedObjects typo, dead filesystem schema attributes, diagnostic severity promotion bug, unused ctx in extract functions, and mustObjectValue passthrough helper**

## Performance

- **Duration:** 15 min
- **Started:** 2026-03-31T00:00:00Z
- **Completed:** 2026-03-31T00:15:00Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Renamed `FreezeLockgedObjects` to `FreezeLockedObjects` across all Go files (struct, resource, test) — JSON wire format unchanged
- Removed dead `nfs_export_policy` and `smb_share_policy` fields from filesystemModel struct, Schema(), and test helpers
- Fixed `readIntoState` severity loss: warnings from `mapFSToModel` are now forwarded as warnings, not promoted to errors
- Removed unused `ctx context.Context` from `extractEradicationConfig`, `extractObjectLockConfig`, `extractPublicAccessConfig` and their 5 call sites
- Removed `mustObjectValue` passthrough helper; inlined 6 `types.ObjectValue` calls directly in `mapFSToModel`
- Extended `DiagnosticReporter` interface with `AddWarning` — backward compatible since `*diag.Diagnostics` already satisfies it

## Task Commits

1. **Task 1: Rename FreezeLockgedObjects typo, remove unused ctx from extract functions** - `9ac9a74` (fix)
2. **Task 2: Remove dead filesystem schema attrs, inline mustObjectValue helper** - `4df996d` (fix)
3. **Task 3: Preserve diagnostic severity in readIntoState** - `b7f17d2` (fix)

## Files Created/Modified
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/models_storage.go` - FreezeLockedObjects field renamed
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/bucket_resource.go` - FreezeLockedObjects references updated, ctx removed from 3 extract functions and 5 call sites
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/bucket_resource_test.go` - FreezeLockedObjects field updated
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/filesystem_resource.go` - Removed NFSExportPolicy/SMBSharePolicy from struct+schema, inlined 6 types.ObjectValue calls, fixed readIntoState severity dispatch
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/filesystem_resource_test.go` - Removed nfs_export_policy/smb_share_policy from type and value maps
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/helpers.go` - Added AddWarning to DiagnosticReporter interface, removed mustObjectValue function

## Decisions Made
- JSON tag `freeze_locked_objects` was already correct — only the Go struct field name was the typo. Tag not touched.
- `DiagnosticReporter.AddWarning` addition is backward compatible: all callers pass `&resp.Diagnostics` which is `*diag.Diagnostics` already implementing both `AddError` and `AddWarning`.
- `nfs_export_policy` and `smb_share_policy` had no mapping in `mapFSToModel` and no API field — confirmed they were placeholder attributes with no behavior.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

The final verification grep `grep -r "nfs_export_policy\|smb_share_policy" --include="*.go" .` matched other resource files (nfs_export_policy_resource.go, smb_share_policy_resource.go) that are entirely separate resources. These are not dead code — they are legitimate resources. The verification was refined to check only the target files (filesystem_resource.go, helpers.go, filesystem_resource_test.go) where all matches confirm clean removal.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Phase 32 complete. All five code correctness issues resolved.
- Phase 33 (client hardening: OAuth2 context, RetryBaseDelay removal, linter expansion) is unblocked.
- Phase 34 (test quality: ExpectNonEmptyPlan removal, acceptance test expansion) is unblocked.

---
*Phase: 32-code-correctness-fixes*
*Completed: 2026-03-31*
