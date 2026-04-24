---
phase: 20-code-quality-validators-dedup
plan: 01
subsystem: infra
tags: [go, terraform-provider, refactoring, code-quality, deduplication]

# Dependency graph
requires:
  - phase: 19-error-handling
    provides: error handling patterns used across resource files
provides:
  - shared helpers in helpers.go (spaceAttrTypes, mapSpaceToObject, nullTimeoutsValue, mustObjectValue, DiagnosticReporter)
  - pre-compiled regex patterns in validators.go
  - panic-free object construction across all map*ToModel functions
affects: [20-code-quality-validators-dedup, 21-dead-code-cleanup, 22-test-validation]

# Tech tracking
tech-stack:
  added: []
  patterns: [shared-helpers-pattern, diagnostic-reporter-interface, pre-compiled-regex, panic-free-diagnostics]

key-files:
  created: []
  modified:
    - internal/provider/helpers.go
    - internal/provider/validators.go
    - internal/provider/filesystem_resource.go
    - internal/provider/filesystem_data_source.go
    - internal/provider/bucket_resource.go
    - internal/provider/bucket_data_source.go
    - internal/provider/object_store_account_resource.go
    - internal/provider/object_store_account_data_source.go

key-decisions:
  - "mapFSToModel, mapBucketToModel, mapOSAToModel return diag.Diagnostics instead of panicking"
  - "DiagnosticReporter interface replaces 15 inline interface declarations for readIntoState"
  - "object_store_virtual_host_resource.go kept 3-method inline interface (uses Append) -- not covered by DiagnosticReporter"

patterns-established:
  - "DiagnosticReporter: minimal interface for readIntoState diag forwarding"
  - "mapSpaceToObject: single source of truth for space nested object construction"
  - "nullTimeoutsValue: single source of truth for ImportState timeout initialization"

requirements-completed: [VAL-01, DUP-01, DUP-02, DUP-03, DUP-04, DUP-05]

# Metrics
duration: 40min
completed: 2026-03-29
---

# Phase 20 Plan 01: Code Quality Validators Dedup Summary

**Pre-compiled regex, shared space/timeout/diagnostic helpers eliminating 400+ lines of duplication across 31 resource files**

## Performance

- **Duration:** 40 min
- **Started:** 2026-03-29T17:09:22Z
- **Completed:** 2026-03-29T17:49:18Z
- **Tasks:** 2
- **Files modified:** 35

## Accomplishments
- Pre-compiled regex patterns at package level in validators.go (eliminates per-call compilation overhead)
- 5 shared helpers added to helpers.go: DiagnosticReporter, spaceAttrTypes, mapSpaceToObject, nullTimeoutsValue, mustObjectValue
- Replaced 29 inline timeout initialization blocks with nullTimeoutsValue()
- Replaced 15 inline readIntoState interface declarations with DiagnosticReporter
- Removed 3 duplicate *SpaceAttrTypes functions (fsSpaceAttrTypes, bucketSpaceAttrTypes, objectStoreAccountSpaceAttrTypes)
- Eliminated all panic() calls in map*ToModel functions -- now return diag.Diagnostics
- Net reduction: 400+ lines removed

## Task Commits

Each task was committed atomically:

1. **Task 1: Add shared helpers to helpers.go and refactor validators.go** - `5d7b484` (feat)
2. **Task 2: Update all resource files to use shared helpers** - `4e1e58f` (refactor)

## Files Created/Modified
- `internal/provider/validators.go` - Pre-compiled regex at package level
- `internal/provider/helpers.go` - 5 new shared helpers (DiagnosticReporter, spaceAttrTypes, mapSpaceToObject, nullTimeoutsValue, mustObjectValue)
- `internal/provider/filesystem_resource.go` - mapFSToModel returns diags, removed fsSpaceAttrTypes, removed old mustObjectValue
- `internal/provider/filesystem_data_source.go` - Uses mapSpaceToObject and new mustObjectValue signature
- `internal/provider/bucket_resource.go` - mapBucketToModel returns diags, removed bucketSpaceAttrTypes, uses mapSpaceToObject
- `internal/provider/bucket_data_source.go` - Uses mapSpaceToObject
- `internal/provider/object_store_account_resource.go` - mapOSAToModel returns diags, removed objectStoreAccountSpaceAttrTypes
- `internal/provider/object_store_account_data_source.go` - Uses mapSpaceToObject
- 27 additional resource files - nullTimeoutsValue() and/or DiagnosticReporter replacements

## Decisions Made
- mapFSToModel, mapBucketToModel, mapOSAToModel changed to return diag.Diagnostics instead of panicking on object construction errors
- DiagnosticReporter interface defined with minimal AddError + HasError methods
- object_store_virtual_host_resource.go kept its 3-method inline interface (uses Append) since DiagnosticReporter only covers the 2-method pattern
- Unused `attr` imports removed from 20 files that only referenced attr.Type for the now-extracted timeout block

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed unused attr imports from 20 files**
- **Found during:** Task 2
- **Issue:** After replacing inline `map[string]attr.Type{...}` timeout blocks with `nullTimeoutsValue()`, 20 files had unused `attr` imports causing build failure
- **Fix:** Removed `"github.com/hashicorp/terraform-plugin-framework/attr"` import from all affected files
- **Files modified:** 20 resource files
- **Committed in:** 4e1e58f (Task 2 commit)

**2. [Rule 3 - Blocking] Removed unused diag import from syslog_server_resource.go**
- **Found during:** Task 2
- **Issue:** syslog_server_resource.go had 3-method inline interface including `Append(...diag.Diagnostic)`. After replacing with DiagnosticReporter, `diag` import became unused
- **Fix:** Removed unused import
- **Files modified:** internal/provider/syslog_server_resource.go
- **Committed in:** 4e1e58f (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (2 blocking)
**Impact on plan:** Both auto-fixes required for compilation. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All shared helpers are in place for Phase 20 Plan 02 (additional dedup work)
- Phase 21 dead code cleanup can proceed after Phase 20 completes
- All 311 unit tests pass

---
*Phase: 20-code-quality-validators-dedup*
*Completed: 2026-03-29*
