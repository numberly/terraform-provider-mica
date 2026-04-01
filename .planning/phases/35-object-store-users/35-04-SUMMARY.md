---
phase: 35-object-store-users
plan: "04"
subsystem: provider
tags: [object-store-users, mocked-tests, resource-test, data-source-test, policy-test, roadmap]

dependency_graph:
  requires:
    - phase: 35-02
      provides: flashblade_object_store_user resource and data source
    - phase: 35-03
      provides: flashblade_object_store_user_policy member resource
  provides:
    - TestMocked_ObjectStoreUser_Lifecycle (Create/PlanConvergence/Import)
    - TestMocked_ObjectStoreUser_FullAccess (Create with full_access/PlanConvergence)
    - TestMocked_ObjectStoreUser_DataSource (resource + data source read)
    - TestMocked_ObjectStoreUserPolicy_Lifecycle (Create/PlanConvergence/Import)
    - ROADMAP.md updated with flashblade_object_store_user_policy entry
  affects:
    - test coverage totals (676 tests across 5 packages)
    - ROADMAP.md coverage counter

tech-stack:
  added: []
  patterns:
    - setupObjectStoreUserTest helper registers account + user mock handlers, sets env vars
    - setupObjectStoreUserPolicyTest helper same pattern for policy tests
    - ImportStateId used to pass name (not UUID) for import verification
    - ImportStateVerifyIdentifierAttribute: "user_name" for policy resource (no 'id' field)
    - depends_on in data source config to sequence resource creation before read

key-files:
  created:
    - internal/provider/object_store_user_resource_test.go
    - internal/provider/object_store_user_data_source_test.go
    - internal/provider/object_store_user_policy_resource_test.go
  modified:
    - internal/testmock/handlers/object_store_users.go
    - ROADMAP.md

key-decisions:
  - "Mock handler fixed: objectStoreUserStore now stores full ObjectStoreUser with UUID id (was bool + empty string) — Rule 1 auto-fix"
  - "ImportStateId='testaccount/testuser' required explicitly — TestCheckResourceAttrSet passes but terraform testing framework uses 'id' attr by default for import ID which would be the UUID"
  - "ImportStateVerifyIdentifierAttribute='user_name' for policy resource because it has no 'id' attribute in its schema"
  - "setupObjectStoreUser(Policy)Test helpers created per-file following acceptance_test.go setupAcceptanceTest pattern"

requirements-completed:
  - OSU-01
  - OSU-02
  - OSU-03
  - OSU-04
  - OSU-05
  - OSU-06
  - OSU-07

duration: 15min
completed: 2026-03-31
---

# Phase 35 Plan 04: Object Store User Mocked Tests and ROADMAP Summary

**Mocked provider tests for flashblade_object_store_user, flashblade_object_store_user data source, and flashblade_object_store_user_policy with full lifecycle coverage and ROADMAP.md updated.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-03-31T08:30:00Z
- **Completed:** 2026-03-31T08:45:00Z
- **Tasks:** 2 (+ 1 auto-approved checkpoint)
- **Files modified:** 5

## Accomplishments

- `TestMocked_ObjectStoreUser_Lifecycle`: Create → plan-convergence → ImportState (3 steps)
- `TestMocked_ObjectStoreUser_FullAccess`: Create with full_access=false → plan-convergence
- `TestMocked_ObjectStoreUser_DataSource`: resource creation + data source read verifying name/id/full_access
- `TestMocked_ObjectStoreUserPolicy_Lifecycle`: user + policy association Create → plan-convergence → ImportState
- Fixed mock handler to return proper UUID-based IDs (was empty string)
- ROADMAP.md: `flashblade_object_store_user_policy` added to Storage Implemented section
- All 676 tests pass; lint clean; docs regenerated

## Task Commits

1. **Task 1: Resource and data source tests + mock fix** - `6d4cad5` (test)
2. **Task 2: Policy resource test + ROADMAP.md** - `f520ff2` (test)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_user_resource_test.go` — 2 lifecycle test functions
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_user_data_source_test.go` — data source test function
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_user_policy_resource_test.go` — policy lifecycle test function
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/object_store_users.go` — mock now stores ObjectStoreUser with UUID id; AddUserForTest helper added; handleDelete also cleans policies
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/ROADMAP.md` — flashblade_object_store_user_policy added; covered count updated to ~32

## Decisions Made

- **Mock handler upgraded**: `byName map[string]bool` → `map[string]*client.ObjectStoreUser` so each created user gets a UUID. Previous empty string caused `TestCheckResourceAttrSet("id")` to fail. (Rule 1 auto-fix)
- **`ImportStateId` explicit**: The terraform testing framework defaults to using the `id` attribute as the import ID. Since our `id` is a UUID but import uses the user name, we supply `ImportStateId: "testaccount/testuser"` explicitly.
- **`ImportStateVerifyIdentifierAttribute: "user_name"`**: The policy resource has no `id` field. The testing framework would fail with "missing identifier attribute 'id'" without this override.
- **`setupObjectStoreUser(Policy)Test` per-file helpers**: Follow the `setupAcceptanceTest` pattern from `acceptance_test.go` — each registers account + user handlers and sets env vars.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Mock handler returned empty string for object store user ID**

- **Found during:** Task 1 execution — `TestCheckResourceAttrSet("id")` failed
- **Issue:** `objectStoreUserStore.byName` was `map[string]bool`, so GET/POST responses returned `"id": ""`
- **Fix:** Changed to `map[string]*client.ObjectStoreUser` with UUID generated at POST time; GET returns stored struct
- **Files modified:** `internal/testmock/handlers/object_store_users.go`
- **Commit:** `6d4cad5`

**2. [Rule 1 - Bug] ImportState used UUID as import ID instead of user name**

- **Found during:** Task 1 — Step 3 failed with "Object store user not found (UUID)"
- **Issue:** Default import uses `id` attribute value (UUID) but the resource ImportState expects the name
- **Fix:** Added `ImportStateId: "testaccount/testuser"` to the import step
- **Files modified:** `internal/provider/object_store_user_resource_test.go`
- **Commit:** `6d4cad5`

**3. [Rule 1 - Bug] Policy resource import failed with "missing identifier attribute 'id'"**

- **Found during:** Task 2 — policy resource has no `id` attribute
- **Fix:** Added `ImportStateVerifyIdentifierAttribute: "user_name"` to the import step
- **Files modified:** `internal/provider/object_store_user_policy_resource_test.go`
- **Commit:** `f520ff2`

## Issues Encountered

None remaining.

## User Setup Required

None — all tests use the mock server.

## OSU Requirements Coverage

| Requirement | Description | Test |
|-------------|-------------|------|
| OSU-01 | Create object store user | TestMocked_ObjectStoreUser_Lifecycle Step 1 |
| OSU-02 | Read/refresh user state | TestMocked_ObjectStoreUser_Lifecycle Step 2 (plan convergence) |
| OSU-03 | Delete user | Terraform destroy after each test run |
| OSU-04 | Import user by name | TestMocked_ObjectStoreUser_Lifecycle Step 3 |
| OSU-05 | Create policy association | TestMocked_ObjectStoreUserPolicy_Lifecycle Step 1 |
| OSU-06 | Import policy association | TestMocked_ObjectStoreUserPolicy_Lifecycle Step 3 |
| OSU-07 | Data source read | TestMocked_ObjectStoreUser_DataSource |

---
*Phase: 35-object-store-users*
*Completed: 2026-03-31*

## Self-Check: PASSED

- `internal/provider/object_store_user_resource_test.go` — FOUND
- `internal/provider/object_store_user_data_source_test.go` — FOUND
- `internal/provider/object_store_user_policy_resource_test.go` — FOUND
- `internal/testmock/handlers/object_store_users.go` — FOUND (modified)
- `ROADMAP.md` — flashblade_object_store_user_policy entry FOUND
- Commit `6d4cad5` — FOUND in git log
- Commit `f520ff2` — FOUND in git log
- `go test ./... -count=1` — 676 passed
- `make lint` — 0 issues
- `make docs` — generated successfully
