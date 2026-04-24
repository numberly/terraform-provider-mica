---
phase: 55-full-mapping-28-resources-21-data-sources
plan: 01
subsystem: infra
tags: [pulumi, terraform-bridge, tfbridge, go, resources, composite-id, secrets]

# Dependency graph
requires:
  - phase: 54-bridge-bootstrap-poc-3-resources
    provides: resources.go skeleton with POC overrides (1 ComputeID, 1 Secret, omitTimeoutsOnAll)
provides:
  - 3 additional ComputeID closures for all composite-ID resources
  - 6 additional Secret:tfbridge.True() marks for all remaining sensitive fields
  - SOFTDELETE-02 registration confirmed for flashblade_file_system
affects:
  - 55-02 (plan 02 depends on correct resources.go for make tfgen)
  - pulumi schema generation (schema.json will reflect Secret marks)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "ComputeID closure pattern: read Pulumi camelCase state keys, validate, return bucketName/ruleName ID"
    - "Secret:tfbridge.True() belt-and-braces for all Sensitive:true TF fields"

key-files:
  created: []
  modified:
    - pulumi/provider/resources.go

key-decisions:
  - "COMPOSITE-03 uses policyName+name (not rule_index) — network_access_policy_rule CRUD uses string rule name for GetByName"
  - "COMPOSITE-04 role FIRST in composite ID — built-in policy names (pure:policy/array_admin) contain slashes"
  - "array_connection_key.connection_key gets Secret:True() in addition to the existing id:False() override"
  - "omitTimeoutsOnAll runs before per-resource overrides, so r.Fields is already initialized — Secret assignments append safely"

patterns-established:
  - "ComputeID pattern: state[camelCaseKey].V.(string) with mapKeys() in error message"
  - "Secret marks appended after omitTimeoutsOnAll — r.Fields already non-nil"

requirements-completed:
  - MAPPING-01
  - MAPPING-04
  - COMPOSITE-02
  - COMPOSITE-03
  - COMPOSITE-04
  - SECRETS-02
  - SOFTDELETE-02

# Metrics
duration: 8min
completed: 2026-04-22
---

# Phase 55 Plan 01: ComputeID + Secret Marks Summary

**4 ComputeID closures and 7 Secret:tfbridge.True() marks covering all composite-ID and sensitive-field resources in resources.go**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-04-22T11:10Z
- **Completed:** 2026-04-22T11:18Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Added ComputeID closures for 3 remaining composite-ID resources (bucket_access_policy_rule, network_access_policy_rule, management_access_policy_directory_service_role_membership)
- Added Secret:tfbridge.True() for 6 remaining sensitive fields across 5 resources
- Confirmed SOFTDELETE-02 registration for flashblade_file_system already present
- go build + go vet clean in pulumi/provider/

## Task Commits

1. **Task 1: Add ComputeID closures + Secret marks + SOFTDELETE-02** - `cc07ddc` (feat)

**Plan metadata:** (docs commit below)

## Files Created/Modified
- `pulumi/provider/resources.go` — 130 lines added: 3 ComputeID closures + 6 Secret marks

## Decisions Made
- COMPOSITE-03 (network_access_policy_rule) uses `policyName` + `name` state keys — the CRUD path uses string rule names via `GetNetworkAccessPolicyRuleByName`, even though ImportState parses an integer index
- COMPOSITE-04 (management_access_policy_dsr_membership) uses `role`/`policy` (single-word, no camelCase conversion needed)
- `omitTimeoutsOnAll` pre-initializes `r.Fields` for every resource, so Secret assignments for array_connection_key append safely without nil-check on Fields

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- resources.go now has all 4 ComputeID closures and 7 Secret:True() marks
- Plan 55-02 can proceed to `make tfgen` + schema.json regeneration to capture the new Secret marks

---
*Phase: 55-full-mapping-28-resources-21-data-sources*
*Completed: 2026-04-22*
