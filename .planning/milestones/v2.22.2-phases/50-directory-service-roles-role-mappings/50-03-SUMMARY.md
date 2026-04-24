---
phase: 50-directory-service-roles-role-mappings
plan: "03"
subsystem: provider
tags: [go, terraform, flashblade, directory-service, ldap, rbac, management-access-policy]

# Dependency graph
requires:
  - phase: 50-directory-service-roles-role-mappings
    plan: "01"
    provides: "DirectoryServiceRole*/client methods"
  - phase: 50-directory-service-roles-role-mappings
    plan: "02"
    provides: "RegisterDirectoryServiceRolesHandlers mock with Seed API"
provides:
  - "flashblade_directory_service_role Terraform resource (full CRUD, 4 interface assertions, SchemaVersion 0)"
  - "flashblade_directory_service_role data source (2 interface assertions, no timeouts)"
  - "3 TestUnit_DirectoryServiceRoleResource_* tests"
  - "1 TestUnit_DirectoryServiceRoleDataSource_* test"
affects:
  - 50-04 (DSRM resource tests may reuse testmock patterns)
  - 50-05 (provider registration: NewDirectoryServiceRoleResource + NewDirectoryServiceRoleDataSource)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "listplanmodifier.RequiresReplace() on management_access_policies (readonly on PATCH — forces re-creation)"
    - "D-02 supersession: role attribute Computed-only (deprecated per swagger); SC-3 replacement trigger on management_access_policies not role"
    - "listvalidator.SizeAtLeast(1) enforces non-empty management_access_policies list on POST"
    - "mapDirectoryServiceRoleToModel helper: shared by Create/Read/Update/ImportState"
    - "Drift detection on 4 fields: group, group_base, management_access_policies, role.name"
    - "roleAttrTypes() map for nested Computed-only deprecated role sub-object"

key-files:
  created:
    - internal/provider/directory_service_role_resource.go
    - internal/provider/directory_service_role_resource_test.go
    - internal/provider/directory_service_role_data_source.go
    - internal/provider/directory_service_role_data_source_test.go
  modified: []

key-decisions:
  - "D-02 confirmed: role attribute is Computed-only nested object, no plan modifier, no RequiresReplace — SC-3 replacement trigger lives exclusively on management_access_policies via listplanmodifier.RequiresReplace()"
  - "stringSlicesEqual reused from helpers.go — no new definition in resource file"
  - "listvalidator.SizeAtLeast(1) added per Q2 resolution: empty management_access_policies list must be rejected at schema validation time"

patterns-established:
  - "mapDirectoryServiceRoleToModel: shared helper returned as diag.Diagnostics (not pointer-receiver) for reuse by multiple CRUD methods"
  - "Data source test pattern: client.NewClient directly (no shared newTestClientForMock helper — none exists in this package)"

requirements-completed: [DSR-01, DSR-02, DSR-03, DSR-04, DSR-05, DSR-06, QA-03, QA-05, QA-07]

# Metrics
duration: 5min
completed: "2026-04-17"
---

# Phase 50 Plan 03: Directory Service Role Resource and Data Source Summary

**flashblade_directory_service_role resource with LDAP-group-to-policy mapping (RequiresReplace on management_access_policies, Computed-only deprecated role sub-object) plus matching data source and 4 TestUnit tests**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-17T09:28:59Z
- **Completed:** 2026-04-17T09:34:00Z
- **Tasks:** 3
- **Files modified:** 4 (created)

## Accomplishments

- Resource with all 4 interface assertions, SchemaVersion 0, correct plan modifiers per D-01/D-02
- management_access_policies: RequiresReplace (readonly on PATCH) + SizeAtLeast(1) validator
- role attribute: Computed-only nested object (D-02 supersession of SC-3 documented in code comment)
- Drift detection on all 4 mutable/computed fields via tflog.Debug
- Data source with exactly 2 interfaces, no timeouts, AddError on not-found
- 4 TestUnit tests all passing (Lifecycle, Import, DriftDetection, DataSource_Basic)

## Task Commits

1. **Task 1: directory_service_role_resource.go** - `9041712` (feat)
2. **Task 2: directory_service_role_resource_test.go** - `434a01e` (test)
3. **Task 3: directory_service_role_data_source.go + test** - `54674a7` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/directory_service_role_resource.go` — Full CRUD resource: 4 interface assertions, SchemaVersion 0, drift detection, ImportState
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/directory_service_role_resource_test.go` — Lifecycle, Import, DriftDetection tests
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/directory_service_role_data_source.go` — Data source: 2 interfaces, no timeouts, AddError on not-found
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/directory_service_role_data_source_test.go` — TestUnit_DirectoryServiceRoleDataSource_Basic

## Decisions Made

- D-02 supersession confirmed and documented in code: `role` attribute has NO plan modifier — SC-3 replacement trigger lives on `management_access_policies` via `listplanmodifier.RequiresReplace()`
- `stringSlicesEqual` reused from `helpers.go` — not redefined in resource file
- `listvalidator.SizeAtLeast(1)` added (Q2 resolution: empty list must be rejected)
- Data source test wires client via `client.NewClient` directly (no shared helper in this package)

## Deviations from Plan

None — plan executed exactly as written.

The plan noted `stringSlicesEqual` should be defined in the resource file, but it already exists in `helpers.go`. Reusing it is correct per CONVENTIONS.md (no duplicate definitions). This is not a deviation but a correct application of existing code.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Resource and data source are implemented but not yet registered in `provider.go` — Plan 50-05 handles registration
- All 4 new tests pass; go build ./... exits 0
- Plan 50-04 (DSRM membership resource) can proceed in parallel

---
*Phase: 50-directory-service-roles-role-mappings*
*Completed: 2026-04-17*
