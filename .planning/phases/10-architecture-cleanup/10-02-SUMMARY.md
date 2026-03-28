---
phase: 10-architecture-cleanup
plan: 02
subsystem: infra
tags: [terraform, go, refactoring, helpers, composite-id]

requires:
  - phase: 10-architecture-cleanup
    provides: "split model files (10-01 prerequisite for clean builds)"
provides:
  - "compositeID, parseCompositeID, stringOrNull shared helpers in helpers.go"
  - "Consistent import ID parsing across all 9 rule/quota resources"
affects: [future-rule-resources, provider-maintenance]

tech-stack:
  added: []
  patterns: [shared-helpers-for-composite-ids, stringOrNull-for-nullable-api-fields]

key-files:
  created:
    - internal/provider/helpers_test.go
  modified:
    - internal/provider/helpers.go
    - internal/provider/nfs_export_policy_rule_resource.go
    - internal/provider/smb_share_policy_rule_resource.go
    - internal/provider/smb_client_policy_rule_resource.go
    - internal/provider/snapshot_policy_rule_resource.go
    - internal/provider/s3_export_policy_rule_resource.go
    - internal/provider/network_access_policy_rule_resource.go
    - internal/provider/object_store_access_policy_rule_resource.go
    - internal/provider/quota_user_resource.go
    - internal/provider/quota_group_resource.go

key-decisions:
  - "parseCompositeID returns error instead of diagnostics for reusability outside ImportState"
  - "Quota resources retain extra empty-part validation on top of parseCompositeID"
  - "compositeID also used for ID construction (not just parsing) for consistency"

patterns-established:
  - "parseCompositeID pattern: all new resources with composite import IDs must use parseCompositeID from helpers.go"
  - "stringOrNull pattern: API fields that return empty string for null must use stringOrNull from helpers.go"
  - "compositeID pattern: use compositeID() instead of manual string concatenation for Terraform IDs"

requirements-completed: [ARC-02, ARC-03]

duration: 8min
completed: 2026-03-28
---

# Phase 10 Plan 02: Shared Helpers Summary

**compositeID/parseCompositeID/stringOrNull helpers replacing duplicated inline logic across 9 resource files with TDD-tested shared functions**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-28T21:05:37Z
- **Completed:** 2026-03-28T21:13:53Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments
- Created 3 shared helper functions (compositeID, parseCompositeID, stringOrNull) in helpers.go with full unit tests
- Replaced all 9 inline SplitN patterns in ImportState methods with parseCompositeID
- Consolidated stringOrNull from smb_share_policy_rule_resource.go to helpers.go (single definition)
- Replaced manual "/" concatenation with compositeID() in 4 model mapper functions
- Removed unused "strings" imports from all 9 resource files

## Task Commits

Each task was committed atomically:

1. **Task 1: Create helpers.go with compositeID, parseCompositeID, and stringOrNull** - `991c074` (feat, TDD)
2. **Task 2: Replace inline composite ID parsing and stringOrNull in all resource files** - `0ec22a7` (refactor)

## Files Created/Modified
- `internal/provider/helpers.go` - Added compositeID, parseCompositeID, stringOrNull functions
- `internal/provider/helpers_test.go` - Table-driven unit tests for all 3 helpers (12 test cases)
- `internal/provider/nfs_export_policy_rule_resource.go` - parseCompositeID in ImportState
- `internal/provider/smb_share_policy_rule_resource.go` - parseCompositeID in ImportState, removed stringOrNull definition
- `internal/provider/smb_client_policy_rule_resource.go` - parseCompositeID in ImportState
- `internal/provider/snapshot_policy_rule_resource.go` - parseCompositeID in ImportState, compositeID in model mapper
- `internal/provider/s3_export_policy_rule_resource.go` - parseCompositeID in ImportState
- `internal/provider/network_access_policy_rule_resource.go` - parseCompositeID in ImportState
- `internal/provider/object_store_access_policy_rule_resource.go` - parseCompositeID in ImportState, compositeID in mapper
- `internal/provider/quota_user_resource.go` - parseCompositeID in ImportState, compositeID in mapper
- `internal/provider/quota_group_resource.go` - parseCompositeID in ImportState, compositeID in mapper

## Decisions Made
- parseCompositeID returns ([]string, error) instead of adding diagnostics directly, keeping the helper reusable outside ImportState contexts
- Quota resources (user/group) retain the extra `parts[0] == "" || parts[1] == ""` validation because parseCompositeID succeeds on "foo/" (returns ["foo", ""])
- Extended compositeID usage to model mapper functions (ID construction) for consistency, not just ImportState parsing

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All shared helpers in place for future rule resources
- Pattern established: new resources with composite IDs use parseCompositeID/compositeID

## Self-Check: PASSED

All files exist, all commits verified.

---
*Phase: 10-architecture-cleanup*
*Completed: 2026-03-28*
