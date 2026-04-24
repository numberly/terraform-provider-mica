---
phase: 50-directory-service-roles-role-mappings
plan: "04"
subsystem: provider
tags: [terraform, flashblade, directory-service, management-access-policy, composite-id, member-resource]

# Dependency graph
requires:
  - phase: 50-directory-service-roles-role-mappings
    plan: "01"
    provides: "ManagementAccessPolicyDirectoryServiceRoleMembership client struct + CRUD methods"
  - phase: 50-directory-service-roles-role-mappings
    plan: "02"
    provides: "RegisterManagementAccessPolicyDirectoryServiceRoleMembershipsHandlers mock handler"
provides:
  - flashblade_management_access_policy_directory_service_role_membership resource (Create/Read/Delete + ImportState)
  - ≥3 TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembershipResource_* tests
  - Resource registered in provider.go
affects:
  - phase: 50-directory-service-roles-role-mappings (plan 50-05 for DSR data source + examples)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Composite ID member resource with 4 interface assertions including ResourceWithUpgradeState (D-00-b)"
    - "role_name/policy_name composite ID (role FIRST) via compositeID() to handle policy names containing : and /"
    - "CRD-only resource: Update returns error, both fields RequiresReplace"
    - "Read calls resp.State.RemoveResource(ctx) on client.IsNotFound (D-08)"
    - "ImportState via parseCompositeID(req.ID, 2) — role at parts[0], policy at parts[1]"
    - "nullTimeoutsValueCRD() used in ImportState for CRD-only resources"

key-files:
  created:
    - internal/provider/management_access_policy_directory_service_role_membership_resource.go
    - internal/provider/management_access_policy_directory_service_role_membership_resource_test.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "Composite ID puts role_name FIRST (role/policy) so SplitN splits built-in policy names like pure:policy/array_admin correctly"
  - "Used nullTimeoutsValueCRD() in ImportState (not nullTimeoutsValue) since this resource has no Update timeout"
  - "Registered only DSRM resource in provider.go; DSR resource registration deferred to plan 50-03 (parallel wave)"

patterns-established:
  - "MissingAssociation test pattern: build priorState via tftypes.NewValue + verify State.Raw.IsNull() after Read"
  - "Parallel wave coordination: each plan registers only its own factory functions in provider.go"

requirements-completed: [DSRM-01, DSRM-02, DSRM-03, DSRM-04, DSRM-05, QA-04, QA-07]

# Metrics
duration: 15min
completed: 2026-04-17
---

# Phase 50 Plan 04: Directory Service Role Membership Resource Summary

**DSRM Terraform resource with composite ID `role_name/policy_name` (role first), CRD lifecycle, and RemoveResource drift detection for the management-access-policies/directory-services/roles API**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-04-17T09:17:00Z
- **Completed:** 2026-04-17T09:32:12Z
- **Tasks:** 2
- **Files modified:** 3 (2 created, 1 modified)

## Accomplishments

- `flashblade_management_access_policy_directory_service_role_membership` resource with all 4 interface assertions including `ResourceWithUpgradeState` with empty map (D-00-b)
- Composite ID `role_name/policy_name` (role FIRST per D-05) handles built-in policy names containing `:` and `/` (e.g. `pure:policy/array_admin`)
- Read uses `resp.State.RemoveResource(ctx)` on `client.IsNotFound` to handle associations removed outside Terraform (D-08)
- 3 tests: Lifecycle, Import (special-char policy name), MissingAssociation (DSRM-05 path)

## Task Commits

1. **Task 1: Create DSRM resource file** - `92431df` (feat)
2. **Task 2: Create DSRM resource tests** - `5f229e6` (test)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/management_access_policy_directory_service_role_membership_resource.go` - DSRM resource: 4 interface assertions, CRD lifecycle, composite ID, ImportState
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/management_access_policy_directory_service_role_membership_resource_test.go` - 3 unit tests: Lifecycle, Import, MissingAssociation
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider.go` - Registered NewManagementAccessPolicyDirectoryServiceRoleMembershipResource

## Decisions Made

- **role FIRST in composite ID**: `compositeID(role, policy)` → `role_name/policy_name` so `strings.SplitN(id, "/", 2)` correctly handles built-in policy names that contain both `:` and `/`
- **nullTimeoutsValueCRD()**: Used CRD-variant (create/read/delete only) since resource has no Update operation
- **Parallel wave**: Only registered DSRM resource in provider.go; DSR resource registration belongs to plan 50-03 running in the same wave

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- DSRM resource complete; usable immediately after plan 50-03 completes (DSR resource)
- Plans 50-05/50-06 (data source, examples, docs) can proceed
- Test baseline: +3 tests from this plan

---
*Phase: 50-directory-service-roles-role-mappings*
*Completed: 2026-04-17*
