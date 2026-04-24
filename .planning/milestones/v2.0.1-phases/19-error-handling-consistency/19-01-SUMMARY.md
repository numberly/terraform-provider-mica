---
phase: 19-error-handling-consistency
plan: 01
subsystem: error-handling
tags: [errors.As, wrapped-errors, reflect, api-error, go-idioms]

requires:
  - phase: 18-security-auth-hardening
    provides: hardened auth with context-aware token fetch and HTTP safety-net timeout
provides:
  - errors.As-based IsNotFound, IsConflict, IsUnprocessable handling wrapped errors
  - resilient ParseAPIError with io.ReadAll failure path
  - fresh-GET guard on bucket delete for accurate object count
  - reflect-based countItems test helper
affects: [20-helper-consolidation, 22-test-coverage]

tech-stack:
  added: []
  patterns: [errors.As for error classification, fresh-GET before destructive operations]

key-files:
  created: []
  modified:
    - internal/client/errors.go
    - internal/client/errors_test.go
    - internal/provider/bucket_resource.go
    - internal/provider/bucket_resource_test.go
    - internal/provider/quota_user_resource.go
    - internal/provider/quota_group_resource.go
    - internal/provider/object_store_account_resource.go
    - internal/testmock/handlers/helpers.go
    - internal/testmock/handlers/buckets.go

key-decisions:
  - "Exported BucketStore type to enable test helpers for mock object count manipulation"
  - "Used errors.As universally -- no direct *APIError type assertions remain in codebase"

patterns-established:
  - "errors.As pattern: always use errors.As for APIError classification to support wrapped errors"
  - "Fresh-GET before delete: read current state from API before destructive operations instead of relying on Terraform state"

requirements-completed: [ERR-01, ERR-02, ERR-03, CON-01, CON-02]

duration: 22min
completed: 2026-03-29
---

# Phase 19 Plan 01: Error Handling Consistency Summary

**errors.As-based error classification for wrapped errors, ParseAPIError resilience, fresh-GET bucket delete guard, and reflect-based test helper**

## Performance

- **Duration:** 22 min
- **Started:** 2026-03-29T16:21:17Z
- **Completed:** 2026-03-29T16:43:00Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments

- IsNotFound, IsConflict, IsUnprocessable correctly classify errors wrapped with fmt.Errorf %w at any depth
- ParseAPIError returns a descriptive error when io.ReadAll fails instead of silently swallowing the failure
- All provider resource files use errors.As instead of direct type assertion for *client.APIError
- Bucket delete performs a fresh GET to check real object count instead of relying on potentially stale Terraform state
- countItems test helper uses reflect.ValueOf().Len() instead of JSON marshal/unmarshal round-trip

## Task Commits

Each task was committed atomically:

1. **Task 1: Migrate error helpers to errors.As and harden ParseAPIError** - `3bc7f55` (fix)
2. **Task 2: Bucket delete fresh-GET guard and countItems reflect fix** - `d1edd7b` (fix)

## Files Created/Modified

- `internal/client/errors.go` - errors.As-based IsNotFound/IsConflict/IsUnprocessable + resilient ParseAPIError
- `internal/client/errors_test.go` - Tests for wrapped error classification and ParseAPIError failure path
- `internal/provider/bucket_resource.go` - Fresh GET guard before bucket delete object count check
- `internal/provider/bucket_resource_test.go` - Updated test to use BucketStore.SetObjectCount for non-empty delete
- `internal/provider/quota_user_resource.go` - errors.As pattern for APIError check
- `internal/provider/quota_group_resource.go` - errors.As pattern for APIError check
- `internal/provider/object_store_account_resource.go` - errors.As pattern for APIError check
- `internal/testmock/handlers/helpers.go` - reflect-based countItems replacing JSON round-trip
- `internal/testmock/handlers/buckets.go` - Exported BucketStore type + SetObjectCount test helper

## Decisions Made

- Exported BucketStore type to enable tests to manipulate mock state (ObjectCount) for the fresh-GET guard
- Used errors.As universally -- zero direct *APIError type assertions remain in the codebase

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed compilation error from shadowed err variable**
- **Found during:** Task 2 (bucket delete fresh-GET guard)
- **Issue:** Adding fresh GET introduced `err :=` which shadowed the existing `err` in PatchBucket call, causing "no new variables on left side of :="
- **Fix:** Changed PatchBucket call from `:=` to `=` assignment
- **Files modified:** internal/provider/bucket_resource.go
- **Verification:** go build ./... succeeds
- **Committed in:** d1edd7b (Task 2 commit)

**2. [Rule 3 - Blocking] Exported BucketStore and added SetObjectCount helper**
- **Found during:** Task 2 (bucket delete fresh-GET guard)
- **Issue:** TestUnit_Bucket_NonEmptyDelete relied on stale state ObjectCount but fresh GET returns mock's actual ObjectCount (0). Mock bucket store was unexported.
- **Fix:** Exported BucketStore type, added SetObjectCount method, updated setupBucketMockServer to return store, updated test to set mock ObjectCount
- **Files modified:** internal/testmock/handlers/buckets.go, internal/provider/bucket_resource_test.go
- **Verification:** TestUnit_Bucket_NonEmptyDelete passes with fresh-GET behavior
- **Committed in:** d1edd7b (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** Both fixes necessary for correctness. No scope creep.

## Issues Encountered

None beyond the auto-fixed deviations.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Error handling is now consistent across all error classification paths
- Ready for Phase 20 (helper consolidation) which may reference these patterns
- All 379 tests pass, go vet clean

---
*Phase: 19-error-handling-consistency*
*Completed: 2026-03-29*
