---
phase: 08-smb-client-policies-syslog-and-acceptance-tests
plan: 01
subsystem: api
tags: [terraform, smb, client-policy, flashblade, crud]

# Dependency graph
requires:
  - phase: 03-file-policies
    provides: SMB share policy pattern (resource, rule, data source, mock handler)
provides:
  - flashblade_smb_client_policy resource (CRUD + import + rename)
  - flashblade_smb_client_policy_rule resource (CRUD + import)
  - flashblade_smb_client_policy data source
  - Client CRUD methods for SMB client policies and rules (12 methods)
  - Mock handler for SMB client policies and rules
affects: [acceptance-tests, smb-client-policy-exports]

# Tech tracking
tech-stack:
  added: []
  patterns: [smb-client-policy-resource-pattern, version-read-only-field, access-based-enumeration-bool]

key-files:
  created:
    - internal/client/smb_client_policies.go
    - internal/testmock/handlers/smb_client_policies.go
    - internal/provider/smb_client_policy_resource.go
    - internal/provider/smb_client_policy_rule_resource.go
    - internal/provider/smb_client_policy_data_source.go
    - internal/provider/smb_client_policy_resource_test.go
    - internal/provider/smb_client_policy_rule_resource_test.go
  modified:
    - internal/client/models.go
    - internal/provider/provider.go

key-decisions:
  - "Version field is Computed+UseStateForUnknown (read-only, never sent in POST/PATCH)"
  - "AccessBasedEnumerationEnabled defaults to false via booldefault.StaticBool(false)"
  - "Rule Index field is Computed+UseStateForUnknown (server-assigned)"
  - "SMB client policy delete uses member guard via ListSmbClientPolicyMembers (file-systems filter)"

patterns-established:
  - "SMB client policy mirrors SMB share policy pattern with additional version and access_based_enumeration_enabled fields"
  - "Rule fields are client/encryption/permission instead of principal/change/full_control/read"

requirements-completed: [SMC-01, SMC-02, SMC-03, SMC-04]

# Metrics
duration: 7min
completed: 2026-03-28
---

# Phase 08 Plan 01: SMB Client Policies Summary

**SMB client policy resource family with CRUD+import for policies and rules, version read-only field, and access-based enumeration toggle**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-28T15:33:54Z
- **Completed:** 2026-03-28T15:40:34Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- Full SMB client policy resource with Create/Read/Update/Delete/Import including rename support
- SMB client policy rule resource with client/encryption/permission fields and composite ID import
- Data source for reading SMB client policies by name
- 7 unit tests passing (CRUD, import, data source, plan modifiers for both policy and rule)
- 265 total tests passing with zero regressions

## Task Commits

Each task was committed atomically:

1. **Task 1: Models, client CRUD, and mock handler** - `6260601` (feat)
2. **Task 2: Resource, rule resource, data source, provider registration, and tests** - `7711c8a` (feat)

## Files Created/Modified
- `internal/client/models.go` - Added SmbClientPolicy* structs (7 new types)
- `internal/client/smb_client_policies.go` - 12 client CRUD methods for policies and rules
- `internal/testmock/handlers/smb_client_policies.go` - Mock handler with full CRUD for policies and rules
- `internal/provider/smb_client_policy_resource.go` - flashblade_smb_client_policy resource
- `internal/provider/smb_client_policy_rule_resource.go` - flashblade_smb_client_policy_rule resource
- `internal/provider/smb_client_policy_data_source.go` - flashblade_smb_client_policy data source
- `internal/provider/smb_client_policy_resource_test.go` - Policy unit tests (CRUD, import, data source, plan modifiers)
- `internal/provider/smb_client_policy_rule_resource_test.go` - Rule unit tests (CRUD, import, plan modifiers)
- `internal/provider/provider.go` - Registered 3 new resources/data sources

## Decisions Made
- Version field is Computed+UseStateForUnknown (read-only, never sent in POST/PATCH) per RESEARCH.md Pitfall 1
- AccessBasedEnumerationEnabled defaults to false via booldefault.StaticBool(false)
- Rule Index field is Computed+UseStateForUnknown (server-assigned)
- SMB client policy delete uses member guard via ListSmbClientPolicyMembers (file-systems filter on smb.client_policy.name)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- SMB client policy resource family complete and ready for acceptance tests
- All unit tests pass, provider registration verified

---
*Phase: 08-smb-client-policies-syslog-and-acceptance-tests*
*Completed: 2026-03-28*

## Self-Check: PASSED

All 7 created files verified. Both task commits (6260601, 7711c8a) verified.
