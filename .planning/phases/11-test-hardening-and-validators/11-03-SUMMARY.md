---
phase: 11-test-hardening-and-validators
plan: 03
subsystem: testing
tags: [idempotence, update, unit-test, terraform-provider, mock-server]

# Dependency graph
requires:
  - phase: 11-02
    provides: query param validation for mock handlers
provides:
  - 9 idempotence tests for v1.1 resources (Create->Read drift detection)
  - 3 standalone Update tests for SMB client policy, syslog server, S3 export policy
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: [TestUnit_Xxx_Idempotent pattern for all resources, TestUnit_Xxx_Update standalone pattern]

key-files:
  created: []
  modified:
    - internal/provider/smb_client_policy_resource_test.go
    - internal/provider/smb_client_policy_rule_resource_test.go
    - internal/provider/syslog_server_resource_test.go
    - internal/provider/s3_export_policy_resource_test.go
    - internal/provider/s3_export_policy_rule_resource_test.go
    - internal/provider/object_store_virtual_host_resource_test.go
    - internal/provider/server_resource_test.go
    - internal/provider/file_system_export_resource_test.go
    - internal/provider/object_store_account_export_resource_test.go

key-decisions:
  - "Idempotence tests compare scalar fields only (not list/object fields like DNS or attached_servers) to keep assertions clear and avoid false positives from list ordering"

patterns-established:
  - "TestUnit_Xxx_Idempotent: Create resource, Read back state, compare before/after models field-by-field"
  - "TestUnit_Xxx_Update: Create resource, build update plan with changed field, call Update, verify changed field and unchanged name"

requirements-completed: [TST-01, TST-03]

# Metrics
duration: 4min
completed: 2026-03-29
---

# Phase 11 Plan 03: Idempotence & Update Tests Summary

**9 idempotence tests + 3 standalone Update tests for all v1.1 resource families catching Create->Read drift and explicit field-change paths**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-29T06:46:31Z
- **Completed:** 2026-03-29T06:50:36Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- 9 idempotence tests verify Create followed by Read produces zero attribute drift for all v1.1 resources
- 3 standalone Update tests exercise specific mutable fields (enabled toggle, URI change) isolated from CRUD sequences
- Full test suite (329 tests) passes with no regressions

## Task Commits

Each task was committed atomically:

1. **Task 1: Add idempotence tests for all v1.1 resources** - `c99d936` (test)
2. **Task 2: Add standalone Update lifecycle tests** - `65fcac6` (test)

## Files Created/Modified
- `internal/provider/smb_client_policy_resource_test.go` - Added Idempotent + Update tests
- `internal/provider/smb_client_policy_rule_resource_test.go` - Added Idempotent test
- `internal/provider/syslog_server_resource_test.go` - Added Idempotent + Update tests
- `internal/provider/s3_export_policy_resource_test.go` - Added Idempotent + Update tests
- `internal/provider/s3_export_policy_rule_resource_test.go` - Added Idempotent test
- `internal/provider/object_store_virtual_host_resource_test.go` - Added Idempotent test
- `internal/provider/server_resource_test.go` - Added Idempotent test
- `internal/provider/file_system_export_resource_test.go` - Added Idempotent test
- `internal/provider/object_store_account_export_resource_test.go` - Added Idempotent test

## Decisions Made
- Idempotence tests compare scalar fields only (ID, Name, Enabled, etc.) rather than complex nested objects to keep assertions clear and avoid false positives from list ordering differences

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 11 (Test Hardening & Validators) is now complete (all 3 plans done)
- All v1.2 milestone requirements addressed
- Full test suite of 329 tests passes

---
*Phase: 11-test-hardening-and-validators*
*Completed: 2026-03-29*

## Self-Check: PASSED
- All 9 modified test files exist
- Both task commits verified (c99d936, 65fcac6)
- Idempotent test pattern found in all 9 files
- 329 tests pass with no regressions
