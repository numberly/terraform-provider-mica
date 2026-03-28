---
phase: 09-bug-fixes
plan: 01
subsystem: api
tags: [terraform, flashblade, delete, plan-modifier, bug-fix]

# Dependency graph
requires: []
provides:
  - "Account export Delete sends correct short name to FlashBlade API"
  - "Filesystem writable/destroyed attributes no longer cause plan drift"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "strings.LastIndex for extracting short name from combined API name format"
    - "boolplanmodifier.UseStateForUnknown for Computed-only bool attributes"

key-files:
  created: []
  modified:
    - "internal/provider/object_store_account_export_resource.go"
    - "internal/provider/object_store_account_export_resource_test.go"
    - "internal/provider/filesystem_resource.go"
    - "internal/testmock/handlers/object_store_account_exports.go"

key-decisions:
  - "Made mock DELETE handler strict instead of lenient -- ensures tests catch name format bugs"
  - "Used strings.LastIndex instead of strings.Split for robustness with edge cases"
  - "Fixed destroyed attribute alongside writable -- same bug class, same fix"

patterns-established:
  - "Strict mock handlers: mock should match real API behavior, not paper over resource bugs"

requirements-completed: [BUG-01, BUG-02]

# Metrics
duration: 3min
completed: 2026-03-28
---

# Phase 9 Plan 1: Resource Bug Fixes Summary

**Fixed account export Delete sending wrong name format and filesystem writable/destroyed causing permanent plan drift**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-28T19:07:39Z
- **Completed:** 2026-03-28T19:10:37Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Account export Delete now extracts short export name from combined "account/export" format before calling the API
- Filesystem `writable` and `destroyed` Computed-only attributes now have `UseStateForUnknown` plan modifier, eliminating false plan diffs
- Mock DELETE handler made strict to prevent future regressions of the name format bug
- Added `AddObjectStoreAccountExportWithName` helper for flexible test seeding

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: Account export Delete test** - `0632f30` (test)
2. **Task 1 GREEN: Account export Delete fix** - `4264607` (fix)
3. **Task 2: Filesystem writable drift fix** - `7a0ca80` (fix)

**Plan metadata:** (pending)

_Note: Task 1 followed TDD with RED/GREEN commits_

## Files Created/Modified
- `internal/provider/object_store_account_export_resource.go` - Fixed Delete to extract short name via strings.LastIndex
- `internal/provider/object_store_account_export_resource_test.go` - Updated Delete test for strict mock, added no-slash edge case test
- `internal/testmock/handlers/object_store_account_exports.go` - Made DELETE strict, added AddObjectStoreAccountExportWithName helper
- `internal/provider/filesystem_resource.go` - Added boolplanmodifier.UseStateForUnknown to writable and destroyed

## Decisions Made
- Made mock DELETE handler strict (no lenient fallback) to catch real API interaction bugs in tests
- Used `strings.LastIndex` instead of `strings.Split` for short name extraction -- handles names with multiple slashes
- Fixed `destroyed` alongside `writable` -- same Computed-only bool missing UseStateForUnknown

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed destroyed attribute missing UseStateForUnknown**
- **Found during:** Task 2 (filesystem writable drift)
- **Issue:** `destroyed` attribute was also Computed-only without UseStateForUnknown, same drift bug as `writable`
- **Fix:** Added `boolplanmodifier.UseStateForUnknown()` to `destroyed`
- **Files modified:** internal/provider/filesystem_resource.go
- **Verification:** All 25 filesystem tests pass
- **Committed in:** 7a0ca80 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug - same class as planned fix)
**Impact on plan:** Plan already suggested checking `destroyed`. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Both bugs fixed and tested
- All 32 combined AccountExport + FileSystem tests pass
- Ready for remaining bug fix plans in phase 09

---
*Phase: 09-bug-fixes*
*Completed: 2026-03-28*

## Self-Check: PASSED
