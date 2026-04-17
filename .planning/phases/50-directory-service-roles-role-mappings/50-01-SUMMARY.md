---
phase: 50-directory-service-roles-role-mappings
plan: "01"
subsystem: client
tags: [go, flashblade, directory-service, ldap, rbac, management-access-policy]

# Dependency graph
requires:
  - phase: 49-directory-service-management
    provides: DirectoryService* structs in models_admin.go, getOneByName/patchOne generic usage patterns
provides:
  - DirectoryServiceRole, DirectoryServiceRolePost, DirectoryServiceRolePatch model structs
  - ManagementAccessPolicyDirectoryServiceRoleMembership model struct
  - Get/Post/Patch/Delete DirectoryServiceRole client methods
  - Get/Post/Delete ManagementAccessPolicyDirectoryServiceRoleMembership client methods
  - 9 TestUnit_* client tests covering both resources
affects:
  - 50-02 (mock handlers)
  - 50-03 (resource implementation)
  - 50-04 (data source)
  - 50-05 (provider registration, docs, ROADMAP)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - getOneByName[T] for GET-by-name with empty-list not-found detection
    - postOne[TBody,TResp] for POST without names query param (server-generated name)
    - patchOne[TBody,TResp] for PATCH with names= query param
    - c.delete() for DELETE with names= query param
    - dsrmPath() helper for composite policy_names+role_names query params
    - DirectoryServiceRolePatch excludes management_access_policies (readonly on PATCH per swagger)
    - ManagementAccessPolicyDirectoryServiceRoleMembership uses struct{} as POST body

key-files:
  created:
    - internal/client/directory_service_roles.go
    - internal/client/directory_service_roles_test.go
    - internal/client/management_access_policy_directory_service_role_memberships.go
    - internal/client/management_access_policy_directory_service_role_memberships_test.go
  modified:
    - internal/client/models_admin.go

key-decisions:
  - "POST /directory-services/roles has no names query param — name is server-generated from management_access_policies (D-03)"
  - "DirectoryServiceRolePatch omits ManagementAccessPolicies — readonly on PATCH per swagger/api_references line 434"
  - "dsrmPath uses url.Values to correctly encode policy_names (contains : and /) and role_names"
  - "Composite key for getOneByName is role_name/policy_name (role first, D-05) so SplitN works with special chars in policy name"

patterns-established:
  - "dsrmPath: composite query param builder for management-access-policies/directory-services/roles"
  - "Empty struct{} body for POST membership endpoints with query-param-only identity"

requirements-completed: [DSR-01, DSR-02, DSR-03, DSRM-01, DSRM-02, DSRM-05, QA-01, QA-02]

# Metrics
duration: 3min
completed: "2026-04-17"
---

# Phase 50 Plan 01: Directory Service Roles and DSRM Client Layer Summary

**Go client layer for LDAP→RBAC wiring: Get/Post/Patch/Delete DirectoryServiceRole and Get/Post/Delete ManagementAccessPolicyDirectoryServiceRoleMembership with 9 TestUnit_* tests**

## Performance

- **Duration:** 3 min
- **Started:** 2026-04-17T09:23:43Z
- **Completed:** 2026-04-17T09:26:50Z
- **Tasks:** 5
- **Files modified:** 5

## Accomplishments

- Four model structs added to models_admin.go (DSR GET/POST/PATCH + DSRM membership)
- Four DSR client methods in directory_service_roles.go using getOneByName/postOne/patchOne/c.delete
- Three DSRM client methods in management_access_policy_directory_service_role_memberships.go
- 9 unit tests (5 DSR + 4 DSRM) exercising all operations including special-char policy name `pure:policy/array_admin`
- `go build ./internal/client/...` and `go vet` both clean

## Task Commits

1. **Task 1: Add DirectoryServiceRole model structs** - `b9c35a9` (feat)
2. **Task 2: Create directory_service_roles.go** - `1b2bff6` (feat)
3. **Task 3: Create directory_service_roles_test.go** - `2b000d8` (test)
4. **Task 4: Create management_access_policy_directory_service_role_memberships.go** - `cbdc2fc` (feat)
5. **Task 5: Create management_access_policy_directory_service_role_memberships_test.go** - `bff6391` (test)

## Files Created/Modified

- `internal/client/models_admin.go` — appended 4 struct types for DSR and DSRM
- `internal/client/directory_service_roles.go` — GetDirectoryServiceRole, PostDirectoryServiceRole, PatchDirectoryServiceRole, DeleteDirectoryServiceRole
- `internal/client/directory_service_roles_test.go` — 5 TestUnit_DirectoryServiceRole_* tests
- `internal/client/management_access_policy_directory_service_role_memberships.go` — Get/Post/Delete + dsrmPath helper
- `internal/client/management_access_policy_directory_service_role_memberships_test.go` — 4 TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_* tests

## Decisions Made

- POST uses no `names` query param per swagger (D-03: name is server-generated)
- DirectoryServiceRolePatch struct excludes ManagementAccessPolicies — swagger marks it readonly on PATCH
- dsrmPath uses `url.Values.Encode()` to safely encode policy names containing colons and slashes
- getOneByName composite key is `role_name/policy_name` (D-05) so `SplitN(..., "/", 2)` returns correct parts

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Client layer complete; 50-02 (mock handlers) can now implement RegisterDirectoryServiceRolesHandlers and RegisterManagementAccessPolicyDSRMHandlers using these types
- All 9 new tests pass; existing test suite unaffected

---
*Phase: 50-directory-service-roles-role-mappings*
*Completed: 2026-04-17*
