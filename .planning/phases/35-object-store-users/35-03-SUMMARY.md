---
phase: 35-object-store-users
plan: "03"
subsystem: provider
tags: [object-store-users, policy-association, terraform-resource, CRD, import]

dependency_graph:
  requires:
    - phase: 35-01
      provides: ListObjectStoreUserPolicies, PostObjectStoreUserPolicy, DeleteObjectStoreUserPolicy, ObjectStoreUserPolicyMember struct
  provides:
    - flashblade_object_store_user_policy resource (Create, Read with drift detection, Delete, ImportState)
    - NewObjectStoreUserPolicyResource factory registered in provider.go
  affects:
    - acceptance tests (35-04 or later)
    - docs generation (make docs)

tech-stack:
  added: []
  patterns:
    - CRD-only member resource following qos_policy_member_resource pattern
    - 3-part import ID (account/username/policyname) via strings.SplitN(id, "/", 3)
    - drift detection via tflog.Warn + RemoveResource when association missing on Read

key-files:
  created:
    - internal/provider/object_store_user_policy_resource.go
    - examples/resources/flashblade_object_store_user_policy/resource.tf
    - examples/resources/flashblade_object_store_user_policy/import.sh
  modified:
    - internal/provider/provider.go

key-decisions:
  - "ImportState uses strings.SplitN(id, '//', 3) not parseCompositeID — 3-part ID where first two parts form the user name (account/username)"
  - "Read uses both member.Name and policy.Name in search predicate — avoids false match when user has multiple policies"

patterns-established:
  - "3-part composite import ID: SplitN into 3, reassemble first two for qualified user name"

requirements-completed:
  - OSU-05
  - OSU-06
  - OSU-07

duration: 5min
completed: 2026-03-31
---

# Phase 35 Plan 03: Object Store User Policy Resource Summary

**flashblade_object_store_user_policy CRD resource associating S3 users to access policies with drift detection and 3-part import ID (account/username/policyname)**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-31T07:00:00Z
- **Completed:** 2026-03-31T07:05:00Z
- **Tasks:** 1
- **Files modified:** 4

## Accomplishments

- Implemented flashblade_object_store_user_policy resource following qos_policy_member_resource pattern
- Read method detects drift (association removed outside Terraform) with tflog.Warn + RemoveResource
- ImportState handles 3-part ID (account/username/policyname) via strings.SplitN
- Registered NewObjectStoreUserPolicyResource in provider.go Resources list

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement flashblade_object_store_user_policy resource and examples** - `5dad19a` (feat)

## Files Created/Modified

- `internal/provider/object_store_user_policy_resource.go` — Full CRD resource with Create, Read (drift detection), Delete, ImportState, Update stub
- `internal/provider/provider.go` — NewObjectStoreUserPolicyResource() registered in Resources list
- `examples/resources/flashblade_object_store_user_policy/resource.tf` — HCL example showing user_name + policy_name
- `examples/resources/flashblade_object_store_user_policy/import.sh` — Import example with 3-part ID format

## Decisions Made

- ImportState uses `strings.SplitN(req.ID, "/", 3)` directly (not `parseCompositeID`) because the 3-part split needs re-joining parts 0+1 to form `account/username`, which parseCompositeID does not support
- Read method checks both `member.Name == userName` and `policy.Name == policyName` to avoid false matches when a user has multiple policies attached

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- flashblade_object_store_user_policy resource fully implemented and registered
- Ready for acceptance tests (real array integration) in a subsequent plan
- make docs can now regenerate documentation for this resource

---
*Phase: 35-object-store-users*
*Completed: 2026-03-31*
