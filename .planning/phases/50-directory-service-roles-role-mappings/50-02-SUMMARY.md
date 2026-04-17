---
phase: 50-directory-service-roles-role-mappings
plan: "02"
subsystem: testing
tags: [go, mock, testmock, directory-service-roles, management-access-policy, memberships]

# Dependency graph
requires:
  - phase: 50-directory-service-roles-role-mappings
    plan: "01"
    provides: "client.DirectoryServiceRole, client.DirectoryServiceRolePost, client.DirectoryServiceRolePatch, client.ManagementAccessPolicyDirectoryServiceRoleMembership types"
provides:
  - "Thread-safe mock handler for /api/2.22/directory-services/roles (GET/POST/PATCH/DELETE + Seed)"
  - "Thread-safe mock handler for /api/2.22/management-access-policies/directory-services/roles (GET/POST/DELETE + Seed)"
affects:
  - 50-directory-service-roles-role-mappings/50-03 (resource + data source tests depend on these handlers)
  - 50-directory-service-roles-role-mappings/50-04 (DSRM resource tests depend on membership handler)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "DSR mock: POST derives server-generated name by stripping pure:policy/ prefix from first ManagementAccessPolicies[0].Name"
    - "DSR mock: PATCH rejects management_access_policies with 400 (readonly-on-PATCH enforcement)"
    - "DSRM mock: idempotent POST (create-or-return, 200 always — resolves Q3 from CONTEXT.md)"
    - "DSRM mock: idempotent DELETE (missing pair silently ignored)"

key-files:
  created:
    - internal/testmock/handlers/directory_service_roles.go
    - internal/testmock/handlers/management_access_policy_directory_service_role_memberships.go
  modified: []

key-decisions:
  - "POST name derivation: strip pure:policy/ from first policy name, then role.Name fallback, then role-N sequential — mirrors D-10 from CONTEXT.md"
  - "PATCH readonly guard for management_access_policies: raw JSON decode first, reject if key present (D-01 swagger contract)"
  - "DSRM POST idempotent (Q3 resolved): no 409, create-or-return lets Terraform replay safely"
  - "DSRM DELETE idempotent: missing pair silently ignored (no 404) — matches D-07 + D-08 dissociate semantics"
  - "dsrMemberKey() named helper (not reusing key() from other handlers) to avoid future name collision risk"

patterns-established:
  - "Raw-decode-then-typed-decode pattern for PATCH readonly field enforcement"
  - "Set-based membership store with pipe-delimited composite key for policy|role pairs"

requirements-completed: [DSR-01, DSR-02, DSR-03, DSRM-01, DSRM-02, DSRM-05]

# Metrics
duration: 2min
completed: 2026-04-17
---

# Phase 50 Plan 02: Directory Service Roles Mock Handlers Summary

**Two thread-safe mock HTTP handlers for /directory-services/roles (full CRUD) and /management-access-policies/directory-services/roles (GET/POST/DELETE idempotent) backing Wave 2 provider tests**

## Performance

- **Duration:** 2 min
- **Started:** 2026-04-17T09:25:35Z
- **Completed:** 2026-04-17T09:27:00Z
- **Tasks:** 2
- **Files modified:** 2 (created)

## Accomplishments

- `directoryServiceRolesStore` with server-generated name derivation on POST (strips `pure:policy/` prefix) and readonly-on-PATCH guard for `management_access_policies`
- `mapDsrMembershipsStore` with idempotent POST + DELETE; resolves Q3 from CONTEXT.md (Terraform replays never 409)
- Both handlers: `go build ./internal/testmock/...` and `go vet ./internal/testmock/...` exit 0

## Task Commits

Each task was committed atomically:

1. **Task 1: Create handlers/directory_service_roles.go mock** - `5d0ded4` (feat)
2. **Task 2: Create handlers/management_access_policy_directory_service_role_memberships.go mock** - `bf63f31` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/directory_service_roles.go` - Thread-safe mock for /api/2.22/directory-services/roles with GET/POST/PATCH/DELETE + Seed
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/management_access_policy_directory_service_role_memberships.go` - Thread-safe mock for /api/2.22/management-access-policies/directory-services/roles with GET/POST/DELETE + Seed

## Decisions Made

- **PATCH readonly enforcement via raw decode**: decoded body as `map[string]json.RawMessage` first to detect `management_access_policies` presence before typed decode — cleanest approach without needing a separate sentinel struct
- **dsrMemberKey() vs key()**: introduced a distinct helper name rather than reusing any existing `key()` function to prevent future symbol collision risk as more handlers are added
- **DSRM idempotent POST (Q3)**: chose 200-always over 409-then-retry since the resource's Create doesn't do a Read-before-Create probe; idempotency is simpler and matches Terraform's replay expectations

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None. `go build ./internal/testmock/...` passed on first attempt, confirming 50-01 had already committed the client type definitions.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Both mock handlers are registered-on-demand via `Register*Handlers(ms.Mux)` — Wave 2 resource/data source tests can call them directly
- Wave 3 (plans 50-03, 50-04) can proceed to implement and test `flashblade_directory_service_role` and `flashblade_management_access_policy_directory_service_role_membership` resources

---
*Phase: 50-directory-service-roles-role-mappings*
*Completed: 2026-04-17*
