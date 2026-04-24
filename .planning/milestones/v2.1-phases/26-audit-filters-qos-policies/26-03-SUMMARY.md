---
phase: 26-audit-filters-qos-policies
plan: 03
subsystem: api
tags: [terraform, qos, bandwidth, iops, flashblade, provider]

requires:
  - phase: 26-01
    provides: QoS policy client CRUD functions and models

provides:
  - flashblade_qos_policy resource with full CRUD and import
  - flashblade_qos_policy_member CRD-only resource with composite import
  - flashblade_qos_policy data source
  - Provider registration for all three

affects: [27-testing-docs]

tech-stack:
  added: []
  patterns: [CRD-only member resource with 3-key timeouts, optional int64 null mapping for QoS limits]

key-files:
  created:
    - internal/provider/qos_policy_resource.go
    - internal/provider/qos_policy_resource_test.go
    - internal/provider/qos_policy_member_resource.go
    - internal/provider/qos_policy_member_resource_test.go
    - internal/provider/qos_policy_data_source.go
    - internal/provider/qos_policy_data_source_test.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "QoS policy name uses RequiresReplace (API rename via PATCH exists but not exposed to avoid drift)"
  - "Member resource is CRD-only with 3-key timeouts matching Phase 25 convention"
  - "Optional int64 fields (bandwidth/IOPS) mapped to null when API returns 0"

patterns-established:
  - "QoS resource follows lifecycle_rule_resource pattern for optional int64 null mapping"

requirements-completed: [QOS-01, QOS-02, QOS-03, QOS-04, QOS-05, QOS-06]

duration: 7min
completed: 2026-03-30
---

# Phase 26 Plan 03: QoS Policies Summary

**QoS policy resource with bandwidth/IOPS CRUD, CRD-only member assignment, data source, and 14 unit tests**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-30T11:15:02Z
- **Completed:** 2026-03-30T11:22:00Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- QoS policy resource with Create, Read, Update (PATCH pointer semantics), Delete, Import by name
- QoS policy member resource (CRD-only) with composite "policyName/memberName" import
- QoS policy data source reading all attributes by name
- All 14 QoS unit tests passing, 444 total tests passing across project

## Task Commits

Each task was committed atomically:

1. **Task 1: QoS policy resource with full CRUD and import** - `a534f65` (feat)
2. **Task 2: QoS policy member resource, data source, and provider registration** - `8819859` (feat)

## Files Created/Modified
- `internal/provider/qos_policy_resource.go` - QoS policy resource CRUD + import + pointer-based PATCH
- `internal/provider/qos_policy_resource_test.go` - 7 unit tests: Create, Read, Read_NotFound, Update, Delete, Import, Schema
- `internal/provider/qos_policy_member_resource.go` - CRD-only member resource with composite import
- `internal/provider/qos_policy_member_resource_test.go` - 5 unit tests: Create, Read, Read_NotFound, Delete, Import
- `internal/provider/qos_policy_data_source.go` - Data source reading QoS policy by name
- `internal/provider/qos_policy_data_source_test.go` - 2 unit tests: Read, Read_NotFound
- `internal/provider/provider.go` - Registered 2 resources + 1 data source

## Decisions Made
- QoS policy name uses RequiresReplace to avoid drift from rename via PATCH
- Member resource is CRD-only with 3-key timeouts (Create, Read, Delete) matching Phase 25 convention
- Optional int64 fields (max_total_bytes_per_sec, max_total_ops_per_sec) mapped to null when API returns 0

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 26 complete (all 3 plans: client layer, audit filters, QoS policies)
- Ready for Phase 27 testing and documentation consolidation

---
*Phase: 26-audit-filters-qos-policies*
*Completed: 2026-03-30*
