---
phase: 49-directory-service-management
plan: "03"
subsystem: provider-resource
tags: [go, terraform-provider, flashblade, directory-service, ldap, resource, validator]

# Dependency graph
requires:
  - phase: 49-directory-service-management-plan-01
    provides: DirectoryService + DirectoryServicePatch + DirectoryServiceManagement structs
  - phase: 49-directory-service-management-plan-02
    provides: Mock handler RegisterDirectoryServicesHandlers + Seed for test fixtures

provides:
  - LDAPURIValidator() list validator in validators.go
  - directoryServiceManagementResource with all 4 interface assertions in directory_service_management_resource.go
  - NewDirectoryServiceManagementResource() factory
  - 3 TestUnit_DirectoryServiceManagementResource_* tests in directory_service_management_resource_test.go

affects:
  - 49-04 (data source plan — same mock handler, same model struct)
  - 49-05 (provider registration — registers NewDirectoryServiceManagementResource)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Singleton PATCH-only resource: Create/Update/Delete all call PatchDirectoryServiceManagement with hardcoded 'management' name"
    - "Delete full-reset PATCH: enabled=false, uris=[], base_dn='', bind_user='', nil CACert refs via **NamedReference, management sub-object cleared"
    - "LDAPURIValidator: list validator checking each element starts with ldap:// or ldaps://"
    - "Write-only sensitive field: bind_password UseStateForUnknown, never synced from API, left empty on import"
    - "Drift detection tflog.Debug on 10 fields: {resource, field, was, now} shape"

key-files:
  created:
    - internal/provider/directory_service_management_resource.go
    - internal/provider/directory_service_management_resource_test.go
  modified:
    - internal/provider/validators.go
    - internal/provider/validators_test.go

key-decisions:
  - "No name attribute in schema (D-01): resource hardcoded to managementDirectoryServiceName constant"
  - "Delete sends full-reset PATCH (D-02): all fields set to empty/false/nil — clean slate for next apply"
  - "UseStateForUnknown on id and bind_password only (D-05): services has no modifier so drift is visible"
  - "mapDirectoryServiceToModel helper never touches BindPassword: caller explicitly preserves write-only value"
  - "buildDSMPatchFromPlan skips null/unknown plan fields: Optional+Computed semantics respected on Create"

patterns-established:
  - "directoryServiceManagementModel: no Name field, all other fields Optional+Computed except services (Computed only)"
  - "buildDSMPatchFromPlan helper for Create patch construction"
  - "**NamedReference clear pattern: var nilRef *client.NamedReference; patch.CACertificate = &nilRef"

requirements-completed: [DSM-01, DSM-02, DSM-03, DSM-04, DSM-05, DSM-06, DSM-07, QA-02, QA-05]

# Metrics
duration: ~9min
completed: 2026-04-17
---

# Phase 49 Plan 03: Directory Service Management Resource Summary

**LDAPURIValidator list validator + directoryServiceManagementResource singleton PATCH-only resource with 10-field drift detection, full-reset Delete, write-only bind_password, and 3 TestUnit_DirectoryServiceManagementResource_* passing tests**

## Performance

- **Duration:** ~9 min
- **Started:** 2026-04-17T08:01:57Z
- **Completed:** 2026-04-17T08:10:22Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Extended `internal/provider/validators.go` with `LDAPURIValidator()` list validator; rejects any URI not starting with `ldap://` or `ldaps://`; error message: `uris[N] must start with ldap:// or ldaps://` (D-04)
- Created `internal/provider/directory_service_management_resource.go` (~658 lines):
  - All 4 interface assertions: Resource, WithConfigure, WithImportState, WithUpgradeState
  - SchemaVersion 0, empty UpgradeState map
  - No `name` attribute (D-01): hardcoded `managementDirectoryServiceName = "management"`
  - `bind_password`: Sensitive + UseStateForUnknown (D-05)
  - `uris`: LDAPURIValidator() wired
  - Delete: full-reset PATCH (D-02)
  - Read: 10 `tflog.Debug` drift detection calls ({resource, field, was, now})
  - ImportState: nullTimeoutsValue() + empty bind_password (D-00-e)
- Created `internal/provider/directory_service_management_resource_test.go` (~336 lines):
  - `TestUnit_DirectoryServiceManagementResource_Lifecycle`: create → update base_dn → delete; post-delete GET verifies reset
  - `TestUnit_DirectoryServiceManagementResource_Import`: import by "management"; verifies bind_password empty, fields populated
  - `TestUnit_DirectoryServiceManagementResource_DriftDetection`: out-of-band Seed mutation; verifies Read syncs new base_dn

## Drift Detection Fields

10 fields checked in Read with `tflog.Debug` and `{resource, field, was, now}` shape:
1. `enabled` (bool)
2. `uris` (list comparison via `types.List.Equal`)
3. `base_dn`
4. `bind_user`
5. `ca_certificate`
6. `ca_certificate_group`
7. `user_login_attribute`
8. `user_object_class`
9. `ssh_public_key_attribute`
10. `services` (list comparison)

## Task Commits

1. **Task 1: LDAPURIValidator** - `b14b637` (feat)
2. **Task 2: directoryServiceManagementResource** - `30aa64c` (feat)
3. **Task 3: resource tests** - `16e5682` (test)
4. **Lint fix: staticcheck QF1008** - `a87559f` (fix)

## Files Created/Modified

- `internal/provider/validators.go` — Added ldapURIListValidator struct + LDAPURIValidator() + imports (fmt, strings, types)
- `internal/provider/validators_test.go` — Added TestLDAPURIValidator table-driven test (7 cases, attr import)
- `internal/provider/directory_service_management_resource.go` — Full resource implementation (658 lines)
- `internal/provider/directory_service_management_resource_test.go` — 3 TestUnit tests + helpers (336 lines)

## Decisions Made

- No `name` attribute in schema (D-01): singleton contract enforced at resource level
- `buildDSMPatchFromPlan` helper skips null/unknown plan fields: respects Optional+Computed defaults for Create
- `mapDirectoryServiceToModel` never touches BindPassword: all callers explicitly preserve/restore it
- `emptyStringList()` used for nil/empty URI/Services slices: avoids perpetual plan diff

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed invalid no-op type assertion in Create**
- **Found during:** Task 2 compilation
- **Issue:** Plan had `data.BindPassword = req.Plan.Raw.Type().Is(req.Plan.Raw.Type())` as a placeholder trick that doesn't compile
- **Fix:** Captured `savedPassword := data.BindPassword` before `mapDirectoryServiceToModel`, restored after
- **Commit:** 30aa64c

**2. [Rule 1 - Bug] Fixed staticcheck QF1008 lint violation in test file**
- **Found during:** `make lint` after Task 3
- **Issue:** `data.Timeouts.Object.IsNull()` — redundant `.Object` embedded field in selector
- **Fix:** Changed to `data.Timeouts.IsNull()`
- **Commit:** a87559f

---

**Total deviations:** 2 auto-fixed (Rule 1 — compile error + lint)
**Impact on plan:** No scope change, no architectural deviation.

## Known Stubs

None — resource is fully wired. All data sources from API are mapped. bind_password empty on import is intentional (D-00-e, D-00-b).

## Self-Check: PASSED
