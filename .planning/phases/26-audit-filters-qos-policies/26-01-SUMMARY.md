---
phase: 26-audit-filters-qos-policies
plan: 01
subsystem: api
tags: [flashblade, qos, audit-filter, client, mock]

requires:
  - phase: 25-bucket-access-policies
    provides: "Client CRUD patterns and mock handler patterns"
provides:
  - "BucketAuditFilter, QosPolicy, QosPolicyMember model structs"
  - "Client CRUD for bucket audit filters (4 methods)"
  - "Client CRUD for QoS policies + member management (7 methods)"
  - "Mock handlers for bucket audit filters and QoS policies"
affects: [26-02, 26-03, 27]

tech-stack:
  added: []
  patterns: [QosPolicyMember renamed to avoid collision with existing PolicyMember in models_common.go]

key-files:
  created:
    - internal/client/bucket_audit_filters.go
    - internal/client/qos_policies.go
    - internal/testmock/handlers/bucket_audit_filters.go
    - internal/testmock/handlers/qos_policies.go
  modified:
    - internal/client/models_storage.go

key-decisions:
  - "QosPolicyMember/QosPolicyMemberPost types instead of PolicyMember/PolicyMemberPost to avoid name collision with existing PolicyMember in models_common.go"

patterns-established:
  - "QoS policy member management: separate /members sub-endpoint pattern with policy_names + member_names params"
  - "Mock handler rename support: policy rename moves members to new key"

requirements-completed: [BAF-01, BAF-02, BAF-03, QOS-01, QOS-02, QOS-03, QOS-04]

duration: 4min
completed: 2026-03-30
---

# Phase 26 Plan 01: Client Models and CRUD for Audit Filters and QoS Policies Summary

**Client CRUD methods and mock handlers for bucket audit filters (4 endpoints) and QoS policies with member management (7 endpoints)**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-30T11:01:22Z
- **Completed:** 2026-03-30T11:05:19Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- BucketAuditFilter, QosPolicy, QosPolicyMember model structs added to models_storage.go
- 4 client CRUD methods for bucket audit filters following bucket_access_policies.go pattern
- 7 client methods for QoS policies: Get/Post/Patch/Delete + ListMembers/PostMember/DeleteMember
- Mock handlers with thread-safe stores, Seed methods, and full CRUD dispatch

## Task Commits

Each task was committed atomically:

1. **Task 1: Client models and CRUD methods** - `4ae42ae` (feat)
2. **Task 2: Mock handlers** - `37f25ff` (feat)

## Files Created/Modified
- `internal/client/models_storage.go` - Added BucketAuditFilter, QosPolicy, QosPolicyMember model structs
- `internal/client/bucket_audit_filters.go` - Get/Post/Patch/Delete for bucket audit filters
- `internal/client/qos_policies.go` - Full CRUD + member management for QoS policies
- `internal/testmock/handlers/bucket_audit_filters.go` - Mock CRUD on /buckets/audit-filters
- `internal/testmock/handlers/qos_policies.go` - Mock CRUD on /qos-policies and /qos-policies/members

## Decisions Made
- Renamed PolicyMember/PolicyMemberPost to QosPolicyMember/QosPolicyMemberPost to avoid collision with existing PolicyMember type in models_common.go (used for delete-guard checks)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Renamed PolicyMember types to avoid redeclaration**
- **Found during:** Task 1 (Client models)
- **Issue:** PolicyMember already declared in models_common.go with different fields (name/id only)
- **Fix:** Renamed to QosPolicyMember and QosPolicyMemberPost, updated all references in qos_policies.go
- **Files modified:** internal/client/models_storage.go, internal/client/qos_policies.go
- **Verification:** go build ./... passes
- **Committed in:** 4ae42ae (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Type rename necessary to avoid compilation error. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Client CRUD and mock handlers ready for resource/data source implementation in plans 26-02 and 26-03
- All model types match API spec field names and types

---
*Phase: 26-audit-filters-qos-policies*
*Completed: 2026-03-30*
