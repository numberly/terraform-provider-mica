---
phase: 12-infrastructure-hardening
plan: 02
subsystem: infra
tags: [terraform, plan-modifier, retry, jitter, backoff]

requires:
  - phase: none
    provides: n/a
provides:
  - "Shared int64UseStateForUnknown and float64UseStateForUnknown plan modifiers in helpers.go"
  - "Jittered exponential backoff in retryTransport to prevent thundering herds"
affects: [provider-resources, client-transport]

tech-stack:
  added: [math/rand]
  patterns: [plan-modifier-helpers-in-helpers.go, jittered-backoff]

key-files:
  created:
    - internal/client/transport_internal_test.go
  modified:
    - internal/provider/helpers.go
    - internal/provider/helpers_test.go
    - internal/provider/filesystem_resource.go
    - internal/client/transport.go

key-decisions:
  - "Refactored computeDelay from retryTransport method to package-level function for testability"
  - "Used Go 1.20+ global math/rand (auto-seeded, concurrency-safe) instead of explicit rand.NewSource"

patterns-established:
  - "Plan modifier helpers: all custom plan modifiers live in helpers.go, not in individual resource files"
  - "Internal test files: use package-internal _test.go files for testing unexported functions"

requirements-completed: [HLP-01, HLP-02, TRN-01]

duration: 3min
completed: 2026-03-29
---

# Phase 12 Plan 02: Helpers Consolidation and Retry Jitter Summary

**Consolidated plan modifier helpers into helpers.go and added +/-20% random jitter to retry backoff**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-29T07:41:16Z
- **Completed:** 2026-03-29T07:44:25Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Moved int64UseStateForUnknown from filesystem_resource.go to helpers.go (shared location)
- Added float64UseStateForUnknown for consistency with identical pattern
- Added +/-20% random jitter to computeDelay preventing thundering herd retries
- Full TDD coverage: 8 new unit tests for plan modifiers, 2 for jitter

## Task Commits

Each task was committed atomically:

1. **Task 1: Move int64UseStateForUnknown and add float64UseStateForUnknown**
   - `b2a2dfb` (test: failing tests for both modifiers)
   - `77c8f34` (feat: move + add implementation)
2. **Task 2: Add jitter to exponential backoff**
   - `008490f` (test: failing tests for jitter)
   - `0cd69fd` (feat: implement jitter in computeDelay)

_TDD tasks have RED/GREEN commits._

## Files Created/Modified
- `internal/provider/helpers.go` - Added int64UseStateForUnknown + float64UseStateForUnknown plan modifiers
- `internal/provider/helpers_test.go` - 8 unit tests for both plan modifiers
- `internal/provider/filesystem_resource.go` - Removed inlined int64UseStateForUnknown definition
- `internal/client/transport.go` - Refactored computeDelay to package-level with +/-20% jitter
- `internal/client/transport_internal_test.go` - 2 jitter unit tests (variance + cap)

## Decisions Made
- Refactored computeDelay from retryTransport method to package-level function for direct testability via internal test file
- Used Go 1.20+ global math/rand (auto-seeded, concurrency-safe) instead of creating explicit rand.NewSource

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- helpers.go now serves as the canonical location for all shared plan modifiers
- float64UseStateForUnknown available for future Float64 computed attributes
- Retry jitter active on all API calls, reducing thundering herd risk

---
*Phase: 12-infrastructure-hardening*
*Completed: 2026-03-29*
