---
phase: 27-testing-documentation
plan: 01
subsystem: testing
tags: [terraform, tfplugindocs, unit-tests, hcl-examples, registry-docs]

# Dependency graph
requires:
  - phase: 24-lifecycle-rules
    provides: lifecycle rule resource, data source, and mock handlers
  - phase: 25-bucket-access-policies
    provides: bucket access policy resource, rules, data source, and mock handlers
  - phase: 26-audit-filters-qos-policies
    provides: audit filter and QoS policy resources, data sources, and mock handlers
provides:
  - bucket access policy data source unit tests (Read + NotFound)
  - example HCL for all v2.1 resources (resource.tf) and data sources (data-source.tf)
  - import.sh with correct composite ID formats for all importable v2.1 resources
  - bucket-advanced-features workflow demonstrating full v2.1 stack
  - regenerated registry docs for all v2.1 resources and data sources
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Example HCL per resource/data source for tfplugindocs"
    - "import.sh with composite ID format per importable resource"
    - "Workflow examples in examples/workflows/ for multi-resource stacks"

key-files:
  created:
    - internal/provider/bucket_access_policy_data_source_test.go
    - examples/workflows/bucket-advanced-features/main.tf
    - examples/resources/flashblade_lifecycle_rule/resource.tf
    - examples/resources/flashblade_lifecycle_rule/import.sh
    - examples/resources/flashblade_bucket_access_policy/resource.tf
    - examples/resources/flashblade_bucket_access_policy/import.sh
    - examples/resources/flashblade_bucket_access_policy_rule/resource.tf
    - examples/resources/flashblade_bucket_access_policy_rule/import.sh
    - examples/resources/flashblade_bucket_audit_filter/resource.tf
    - examples/resources/flashblade_bucket_audit_filter/import.sh
    - examples/resources/flashblade_qos_policy/resource.tf
    - examples/resources/flashblade_qos_policy/import.sh
    - examples/resources/flashblade_qos_policy_member/resource.tf
    - examples/resources/flashblade_qos_policy_member/import.sh
    - examples/data-sources/flashblade_lifecycle_rule/data-source.tf
    - examples/data-sources/flashblade_bucket_access_policy/data-source.tf
    - examples/data-sources/flashblade_bucket_audit_filter/data-source.tf
    - examples/data-sources/flashblade_qos_policy/data-source.tf
    - docs/resources/lifecycle_rule.md
    - docs/resources/bucket_access_policy.md
    - docs/resources/bucket_access_policy_rule.md
    - docs/resources/bucket_audit_filter.md
    - docs/resources/qos_policy.md
    - docs/resources/qos_policy_member.md
    - docs/data-sources/lifecycle_rule.md
    - docs/data-sources/bucket_access_policy.md
    - docs/data-sources/bucket_audit_filter.md
    - docs/data-sources/qos_policy.md
  modified: []

key-decisions:
  - "tftypes.Number used for Int64 schema attributes in tftypes object (tftypes.Int64 does not exist)"

patterns-established:
  - "Data source unit test pattern: newTest*DataSource + schema helper + buildType + nullConfig + Read/NotFound tests"
  - "Workflow examples demonstrate full resource stacks with operational comments"

requirements-completed: [TST-01, TST-02, DOC-01, DOC-02]

# Metrics
duration: 3min
completed: 2026-03-30
---

# Phase 27 Plan 01: Testing & Documentation Summary

**Bucket access policy data source unit tests, example HCL + import.sh for all 6 v2.1 resources, workflow example, and regenerated registry docs via tfplugindocs**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-30T11:30:50Z
- **Completed:** 2026-03-30T11:33:49Z
- **Tasks:** 2
- **Files modified:** 30

## Accomplishments
- Added the one missing unit test (bucket access policy data source) bringing v2.1 test coverage to 100%
- Created 6 resource.tf, 6 import.sh, and 4 data-source.tf example files for all v2.1 resources
- Built bucket-advanced-features workflow demonstrating all 8 v2.1 resource types working together
- Regenerated registry docs with tfplugindocs -- 10 new doc pages (6 resources + 4 data sources)

## Task Commits

Each task was committed atomically:

1. **Task 1: Missing data source test + all example HCL and import.sh files** - `49fa998` (test)
2. **Task 2: Workflow example + tfplugindocs regeneration** - `7f5943c` (docs)

## Files Created/Modified
- `internal/provider/bucket_access_policy_data_source_test.go` - Unit tests for Read + NotFound
- `examples/resources/flashblade_lifecycle_rule/resource.tf` - Lifecycle rule example
- `examples/resources/flashblade_lifecycle_rule/import.sh` - Import with bucketName/ruleID format
- `examples/resources/flashblade_bucket_access_policy/resource.tf` - Access policy example
- `examples/resources/flashblade_bucket_access_policy/import.sh` - Import with bucket name
- `examples/resources/flashblade_bucket_access_policy_rule/resource.tf` - Access policy rule example
- `examples/resources/flashblade_bucket_access_policy_rule/import.sh` - Import with bucketName/ruleName format
- `examples/resources/flashblade_bucket_audit_filter/resource.tf` - Audit filter example
- `examples/resources/flashblade_bucket_audit_filter/import.sh` - Import with bucket name
- `examples/resources/flashblade_qos_policy/resource.tf` - QoS policy example
- `examples/resources/flashblade_qos_policy/import.sh` - Import with policy name
- `examples/resources/flashblade_qos_policy_member/resource.tf` - QoS member example
- `examples/resources/flashblade_qos_policy_member/import.sh` - Import with policyName/memberName format
- `examples/data-sources/flashblade_lifecycle_rule/data-source.tf` - Lifecycle rule lookup
- `examples/data-sources/flashblade_bucket_access_policy/data-source.tf` - Access policy lookup
- `examples/data-sources/flashblade_bucket_audit_filter/data-source.tf` - Audit filter lookup
- `examples/data-sources/flashblade_qos_policy/data-source.tf` - QoS policy lookup
- `examples/workflows/bucket-advanced-features/main.tf` - Complete v2.1 workflow
- `docs/resources/lifecycle_rule.md` - Registry docs for lifecycle rule
- `docs/resources/bucket_access_policy.md` - Registry docs for bucket access policy
- `docs/resources/bucket_access_policy_rule.md` - Registry docs for access policy rule
- `docs/resources/bucket_audit_filter.md` - Registry docs for audit filter
- `docs/resources/qos_policy.md` - Registry docs for QoS policy
- `docs/resources/qos_policy_member.md` - Registry docs for QoS member
- `docs/data-sources/lifecycle_rule.md` - Registry docs for lifecycle rule DS
- `docs/data-sources/bucket_access_policy.md` - Registry docs for access policy DS
- `docs/data-sources/bucket_audit_filter.md` - Registry docs for audit filter DS
- `docs/data-sources/qos_policy.md` - Registry docs for QoS policy DS

## Decisions Made
- Used tftypes.Number for Int64 schema attributes in tftypes object definitions (tftypes.Int64 does not exist in the Go SDK)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- v2.1 milestone is fully complete: all resources, data sources, tests, and documentation are in place
- Registry docs are regenerated and ready for publishing
- All 446 unit tests pass with no regressions

## Self-Check: PASSED

All key files verified present. Both task commits (49fa998, 7f5943c) confirmed in git log.

---
*Phase: 27-testing-documentation*
*Completed: 2026-03-30*
