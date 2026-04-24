---
phase: 21-dead-code-removal-modernization
plan: 01
subsystem: api
tags: [dead-code, math-rand-v2, terraform-framework, cleanup]

# Dependency graph
requires:
  - phase: 20-code-quality-validators-dedup
    provides: shared helpers and dedup patterns
provides:
  - "5 unused List* policy functions and Opts types removed"
  - "IsUnprocessable helper removed"
  - "SourceReference type consolidated into NamedReference"
  - "30 empty UpgradeState stubs removed from all resources"
  - "math/rand modernized to math/rand/v2"
affects: [22-acceptance-tests]

# Tech tracking
tech-stack:
  added: [math/rand/v2]
  patterns: [no-empty-upgrade-state, named-reference-only]

key-files:
  modified:
    - internal/client/nfs_export_policies.go
    - internal/client/smb_share_policies.go
    - internal/client/smb_client_policies.go
    - internal/client/snapshot_policies.go
    - internal/client/s3_export_policies.go
    - internal/client/errors.go
    - internal/client/errors_test.go
    - internal/client/models_storage.go
    - internal/client/transport.go
    - internal/testmock/handlers/object_store_access_keys.go
    - internal/provider/*_resource.go (30 files)

key-decisions:
  - "Kept encoding/json import in models_storage.go since json.RawMessage is used by ObjectStoreAccountPost"
  - "Updated rand.Intn to rand.IntN for math/rand/v2 API compatibility"

patterns-established:
  - "No empty UpgradeState methods: only implement ResourceWithUpgradeState when actual state migration is needed"
  - "Use NamedReference universally: no per-type reference structs with identical fields"

requirements-completed: [DCR-01, DCR-02, DCR-03, DCR-04, MOD-01]

# Metrics
duration: 8min
completed: 2026-03-29
---

# Phase 21 Plan 01: Dead Code Removal and Modernization Summary

**Removed 5 unused List* functions, IsUnprocessable helper, SourceReference type, 30 empty UpgradeState stubs, and modernized math/rand to v2**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-29T18:39:37Z
- **Completed:** 2026-03-29T18:47:40Z
- **Tasks:** 2
- **Files modified:** 40

## Accomplishments
- Removed 251 lines of dead code from client package (5 List* functions, Opts types, IsUnprocessable, SourceReference)
- Removed 154 lines of empty UpgradeState boilerplate from 30 resource files
- Modernized math/rand to math/rand/v2 in transport and test mock, eliminating deprecation warnings
- All 375 tests pass, go vet clean, zero compilation errors

## Task Commits

Each task was committed atomically:

1. **Task 1: Remove unused List* functions, IsUnprocessable, and SourceReference** - `d9da0cc` (refactor)
2. **Task 2: Remove empty UpgradeState implementations and update math/rand to v2** - `fd0e4ce` (refactor)

## Files Created/Modified
- `internal/client/nfs_export_policies.go` - Removed ListNfsExportPoliciesOpts and ListNfsExportPolicies
- `internal/client/smb_share_policies.go` - Removed ListSmbSharePoliciesOpts and ListSmbSharePolicies
- `internal/client/smb_client_policies.go` - Removed ListSmbClientPoliciesOpts and ListSmbClientPolicies
- `internal/client/snapshot_policies.go` - Removed ListSnapshotPoliciesOpts and ListSnapshotPolicies
- `internal/client/s3_export_policies.go` - Removed ListS3ExportPoliciesOpts and ListS3ExportPolicies
- `internal/client/errors.go` - Removed IsUnprocessable function
- `internal/client/errors_test.go` - Removed 4 IsUnprocessable test functions
- `internal/client/models_storage.go` - Removed SourceReference type, updated FileSystem.Source to *NamedReference
- `internal/client/transport.go` - Updated math/rand to math/rand/v2
- `internal/testmock/handlers/object_store_access_keys.go` - Updated math/rand to math/rand/v2, Intn to IntN
- `internal/provider/*_resource.go` (30 files) - Removed UpgradeState methods, interface assertions, and Version: 0

## Decisions Made
- Kept `encoding/json` import in models_storage.go because `json.RawMessage` is actively used by ObjectStoreAccountPost
- Updated `rand.Intn` to `rand.IntN` in test mock as required by math/rand/v2 API rename

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed math/rand/v2 API incompatibility (Intn -> IntN)**
- **Found during:** Task 2
- **Issue:** math/rand/v2 renamed `Intn` to `IntN`, causing compilation error
- **Fix:** Updated `rand.Intn` to `rand.IntN` in object_store_access_keys.go
- **Files modified:** internal/testmock/handlers/object_store_access_keys.go
- **Verification:** go build ./... passes
- **Committed in:** fd0e4ce (part of Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Necessary API adaptation for math/rand/v2. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All dead code removed, codebase is leaner
- math/rand/v2 in use, no deprecation warnings from go vet
- Ready for Phase 22 (acceptance tests) to validate all v2.0.1 changes

---
*Phase: 21-dead-code-removal-modernization*
*Completed: 2026-03-29*
