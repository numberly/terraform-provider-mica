---
phase: 09-bug-fixes
plan: 02
subsystem: api-client
tags: [error-handling, json-serialization, flashblade-api]

# Dependency graph
requires:
  - phase: none
    provides: n/a
provides:
  - Scoped IsNotFound with sub-error message suffix matching
  - Clean omitempty tags on GET-only model struct fields
affects: [all-resources, error-classification]

# Tech tracking
tech-stack:
  added: []
  patterns: [sub-error-suffix-matching, no-omitempty-on-struct-types]

key-files:
  created: []
  modified:
    - internal/client/errors.go
    - internal/client/errors_test.go
    - internal/client/models.go

key-decisions:
  - "Use HasSuffix on Errors[0].Message instead of Contains on Error() string for IsNotFound"
  - "Remove omitempty from non-pointer struct fields rather than converting to pointers (GET-only structs)"

patterns-established:
  - "IsNotFound checks sub-error message suffix, not formatted Error() string"
  - "Non-pointer struct fields in GET-only models must NOT have omitempty tags"

requirements-completed: [BUG-03, BUG-04]

# Metrics
duration: 3min
completed: 2026-03-28
---

# Phase 9 Plan 2: Client Error & Model Bug Fixes Summary

**Scoped IsNotFound to sub-error suffix matching and removed misleading omitempty on 14 non-pointer struct fields across 8 GET-only structs**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-28T19:07:36Z
- **Completed:** 2026-03-28T19:10:50Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- IsNotFound now checks `Errors[0].Message` with `HasSuffix` instead of `Error()` with `Contains`, preventing validation errors from being silently swallowed as "not found"
- Added 11 comprehensive unit tests covering all IsNotFound edge cases (nil, non-APIError, 404, legitimate 400 not-found, validation 400, no sub-errors)
- Removed misleading `omitempty` from 14 non-pointer struct fields across FileSystem, ObjectStoreAccount, Bucket, ObjectStoreAccessKey, and 4 policy rule structs

## Task Commits

Each task was committed atomically:

1. **Task 1: Scope IsNotFound (TDD RED)** - `8c89265` (test)
2. **Task 1: Scope IsNotFound (TDD GREEN)** - `7261ae5` (fix)
3. **Task 2: Audit and fix omitempty** - `0bded98` (fix)

_TDD task had separate RED and GREEN commits_

## Files Created/Modified
- `internal/client/errors.go` - Refined IsNotFound to use sub-error suffix matching
- `internal/client/errors_test.go` - Added 11 table-driven unit tests for IsNotFound
- `internal/client/models.go` - Removed omitempty from 14 non-pointer struct fields

## Decisions Made
- Used `HasSuffix` on raw sub-error message rather than regex: simpler, covers both "does not exist." and "does not exist" suffixes
- Removed omitempty tags rather than converting to pointers: GET-only structs never serialized, pointers would add unnecessary nil-check complexity

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Error classification is now robust against false-positive not-found matching
- Model struct tags are consistent and non-misleading
- Full test suite (280 tests) passes with zero regressions

---
*Phase: 09-bug-fixes*
*Completed: 2026-03-28*

## Self-Check: PASSED

All 3 source files exist. All 3 commits verified (8c89265, 7261ae5, 0bded98).
