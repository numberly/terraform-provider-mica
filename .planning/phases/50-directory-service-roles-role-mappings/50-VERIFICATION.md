---
phase: 50-directory-service-roles-role-mappings
verified: 2026-04-17T00:00:00Z
status: passed
score: 6/6 must-haves verified
re_verification: false
---

# Phase 50: Directory Service Roles & Role Mappings — Verification Report

**Phase Goal:** Users can map LDAP groups to FlashBlade roles through Terraform and attach those roles to management access policies — full CRUD on both resources, composite ID import for the membership, and drift detection on all mutable fields.
**Verified:** 2026-04-17
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `terraform apply` with role + memberships converges (HCL examples exist, resource + membership compile) | VERIFIED | `examples/resources/flashblade_directory_service_role/resource.tf` includes both `flashblade_directory_service_role` and membership resource; `go build ./...` exits 0 |
| 2 | `terraform import` works for role (by name) and membership (by composite `role_name/policy_name`) | VERIFIED | `ImportState` in both resources confirmed; `import.sh` files contain correct commands; `parseCompositeID(req.ID, 2)` with `parts[0]=roleName` (D-05) present |
| 3 | PATCH on group/group_base updates in-place; changing `management_access_policies` forces replace; drift on `role` logged | VERIFIED | `listplanmodifier.RequiresReplace()` on `management_access_policies`; no modifier on `group`/`group_base`/`role`; `tflog.Debug` drift detection on 4 fields (group, group_base, management_access_policies, role.name) — 8 matches |
| 4 | Destroying membership leaves policy + role intact (DELETE is association-only) | VERIFIED | `DeleteManagementAccessPolicyDirectoryServiceRoleMembership` calls `DELETE /management-access-policies/directory-services/roles?policy_names=&role_names=` only; mock is dissociation-only with no cross-store deletes |
| 5 | `make test` >= 812 | VERIFIED | 814 tests pass across 4 packages; `make test` exits 0 |
| 6 | `make lint` clean | VERIFIED | `golangci-lint run ./...` reports 0 issues |

**Score:** 6/6 truths verified

---

## Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/models_admin.go` | 4 new struct types (DirectoryServiceRole, Post, Patch, Membership) | VERIFIED | Lines 208, 220, 231, 240; Patch excludes ManagementAccessPolicies; Post excludes Name |
| `internal/client/directory_service_roles.go` | Get/Post/Patch/Delete client methods | VERIFIED | All 4 methods present; use `getOneByName`, `postOne`, `patchOne` generics |
| `internal/client/directory_service_roles_test.go` | >=4 TestUnit_DirectoryServiceRole_* tests | VERIFIED | 5 tests: Get_Found, Get_NotFound, Post, Patch_Group, Delete |
| `internal/client/management_access_policy_directory_service_role_memberships.go` | Get/Post/Delete membership client methods | VERIFIED | 3 methods; `policy_names`+`role_names` query params |
| `internal/client/management_access_policy_directory_service_role_memberships_test.go` | >=3 TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_* | VERIFIED | 4 tests: Get_Exists, Get_NotExists, Post, Delete |
| `internal/testmock/handlers/directory_service_roles.go` | Thread-safe mock with Register+Seed, name-derivation, readonly-PATCH enforcement | VERIFIED | `RegisterDirectoryServiceRolesHandlers`, `strings.TrimPrefix`, `management_access_policies is readonly on PATCH` error present |
| `internal/testmock/handlers/management_access_policy_directory_service_role_memberships.go` | Thread-safe mock with Register+Seed, GET/POST/DELETE | VERIFIED | `RegisterManagementAccessPolicyDirectoryServiceRoleMembershipsHandlers`; `ValidateQueryParams` for `policy_names`+`role_names`; idempotent POST |
| `internal/provider/directory_service_role_resource.go` | 4 interface assertions, RequiresReplace on management_access_policies, drift detection, ImportState | VERIFIED | 4 assertions confirmed; `listplanmodifier.RequiresReplace()`; 8 drift-detection log calls; `nullTimeoutsValue()`; `RemoveResource` |
| `internal/provider/directory_service_role_resource_test.go` | >=3 TestUnit_DirectoryServiceRoleResource_* | VERIFIED | 3 tests: Lifecycle, Import, DriftDetection |
| `internal/provider/directory_service_role_data_source.go` | 2 interface assertions, no timeouts, no plan modifiers | VERIFIED | Exactly 2 `var _ datasource.` assertions; 0 timeouts/planmodifier references; `AddError` on not-found; 0 `RemoveResource` |
| `internal/provider/directory_service_role_data_source_test.go` | >=1 TestUnit_DirectoryServiceRoleDataSource_Basic | VERIFIED | 1 test; exercises all computed fields |
| `internal/provider/management_access_policy_directory_service_role_membership_resource.go` | 4 interface assertions, composite ID `role/policy`, RequiresReplace on both, RemoveResource on miss | VERIFIED | 4 assertions; `compositeID(m.Role.Name, m.Policy.Name)`; `parseCompositeID(req.ID, 2)` with role-first; `RemoveResource`; `Update not supported` |
| `internal/provider/management_access_policy_directory_service_role_membership_resource_test.go` | >=3 tests incl MissingAssociation | VERIFIED | 3 tests: Lifecycle, Import, MissingAssociation; `readResp.State.Raw.IsNull()` assertion; `do NOT seed` comment; `priorState := tfsdk.State` |
| `internal/provider/provider.go` | NewDirectoryServiceRoleResource + NewManagementAccessPolicyDirectoryServiceRoleMembershipResource + NewDirectoryServiceRoleDataSource | VERIFIED | Lines 318, 319, 385 |
| `examples/resources/flashblade_directory_service_role/resource.tf` | HCL with role + membership chaining | VERIFIED | Contains `management_access_policies`, `flashblade_directory_service_role.admins.name` reference |
| `examples/resources/flashblade_directory_service_role/import.sh` | Import by name | VERIFIED | `terraform import flashblade_directory_service_role.admins array_admin` |
| `examples/data-sources/flashblade_directory_service_role/data-source.tf` | Data source lookup example | VERIFIED | `data "flashblade_directory_service_role"` block present |
| `examples/resources/flashblade_management_access_policy_directory_service_role_membership/resource.tf` | HCL membership example | VERIFIED | `pure:policy/readonly` policy, `role = "array_admin"` |
| `examples/resources/flashblade_management_access_policy_directory_service_role_membership/import.sh` | Composite ID import with D-05 comment | VERIFIED | `array_admin/pure:policy/array_admin`; D-05 supersession comment present |
| `docs/resources/directory_service_role.md` | Auto-generated via `make docs` | VERIFIED | File exists; contains `management_access_policies` attribute descriptions |
| `docs/data-sources/directory_service_role.md` | Auto-generated via `make docs` | VERIFIED | File exists |
| `docs/resources/management_access_policy_directory_service_role_membership.md` | Auto-generated via `make docs` | VERIFIED | File exists; contains `role_name/policy_name` composite ID description |
| `ROADMAP.md` | Both resources in Implemented → Array Administration; counters refreshed; Last updated 2026-04-17 | VERIFIED | Lines 100-101 show both resources as Done; Last updated 2026-04-17; Provider version v2.22.2 |
| `CONVENTIONS.md` | Baseline bumped from 798 to >=812 | VERIFIED | Line 451: `Current baseline: **814 tests**` |

---

## Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `directory_service_roles.go:GetDirectoryServiceRole` | `client.go:getOneByName[T]` | `getOneByName[DirectoryServiceRole]` | WIRED | Confirmed in source |
| `directory_service_roles.go:PostDirectoryServiceRole` | `client.go:postOne[TBody, TResp]` | `postOne[DirectoryServiceRolePost, DirectoryServiceRole]` — no names query param | WIRED | Confirmed; POST path is `/directory-services/roles` with no `?names=` |
| `directory_service_roles.go:PatchDirectoryServiceRole` | `client.go:patchOne[TBody, TResp]` | `patchOne[DirectoryServiceRolePatch, DirectoryServiceRole]` | WIRED | Confirmed |
| `directory_service_role_resource.go:Create` | `client.PostDirectoryServiceRole` | `r.client.PostDirectoryServiceRole(ctx, body)` | WIRED | Confirmed at line 118 |
| `directory_service_role_resource.go:Update` | `client.PatchDirectoryServiceRole` | `r.client.PatchDirectoryServiceRole(ctx, name, body)` — only *string ptrs, no management_access_policies | WIRED | Confirmed |
| `directory_service_role_resource.go:Read` | `tflog.Debug` drift detection | 4 separate drift checks on group, group_base, management_access_policies, role.name | WIRED | 8 total `tflog.Debug` calls (4 fields × 2 log invocations) |
| `management_access_policy_directory_service_role_membership_resource.go:Create` | `client.PostManagementAccessPolicyDirectoryServiceRoleMembership` | `r.client.PostManagementAccessPolicyDirectoryServiceRoleMembership(ctx, policy, role)` | WIRED | Confirmed |
| `management_access_policy_directory_service_role_membership_resource.go:Read` | `resp.State.RemoveResource(ctx)` on `client.IsNotFound` | Empty-list-on-miss from mock → `getOneByName` not-found → `RemoveResource` | WIRED | Confirmed; exercised by MissingAssociation test |
| `management_access_policy_directory_service_role_membership_resource.go:ImportState` | `helpers.go:parseCompositeID` | `parseCompositeID(req.ID, 2)` with `parts[0]=roleName` (D-05) | WIRED | Confirmed |
| `provider.go:Resources()` | `NewDirectoryServiceRoleResource` + `NewManagementAccessPolicyDirectoryServiceRoleMembershipResource` | slice entries at lines 318-319 | WIRED | Confirmed |
| `provider.go:DataSources()` | `NewDirectoryServiceRoleDataSource` | slice entry at line 385 | WIRED | Confirmed |

---

## Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `directory_service_role_resource.go` | `role` (API response) | `r.client.GetDirectoryServiceRole` → `getOneByName[DirectoryServiceRole]` → HTTP GET → mock store | Yes — store keyed by name, seeded in tests | FLOWING |
| `directory_service_role_data_source.go` | `config` (all computed fields) | `d.client.GetDirectoryServiceRole` → populated from `DirectoryServiceRole` struct fields | Yes — `ID`, `Name`, `Group`, `GroupBase`, `ManagementAccessPolicies`, `Role` all mapped | FLOWING |
| `management_access_policy_directory_service_role_membership_resource.go` | `m` (association) | `r.client.GetManagementAccessPolicyDirectoryServiceRoleMembership` → mock set lookup | Yes — set keyed by `policy|role` pair | FLOWING |

---

## Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| DSR client tests pass (5 tests) | `go test ./internal/client/... -run TestUnit_DirectoryServiceRole_` | 5 passed | PASS |
| Membership client tests pass (4 tests) | `go test ./internal/client/... -run TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_` | 4 passed | PASS |
| DSR provider tests pass (3 tests) | `go test ./internal/provider/... -run TestUnit_DirectoryServiceRoleResource_` | 3 passed | PASS |
| DSR data source test passes (1 test) | `go test ./internal/provider/... -run TestUnit_DirectoryServiceRoleDataSource_` | 1 passed | PASS |
| DSRM resource tests pass (3 tests incl MissingAssociation) | `go test ./internal/provider/... -run TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembershipResource_` | 3 passed | PASS |
| Full test suite >= 812 | `make test` | 814 passed, 4 packages | PASS |
| Lint clean | `make lint` | 0 issues | PASS |
| Docs idempotent | `make docs && git diff --name-only docs/` | No diff after second run | PASS |

---

## Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| DSR-01 | 50-01, 50-03 | Create directory service role via `flashblade_directory_service_role` | SATISFIED | `PostDirectoryServiceRole` wired; HCL example; server-generated name via `UseStateForUnknown` |
| DSR-02 | 50-01, 50-03 | Update group/group_base via PATCH without recreation | SATISFIED | `PatchDirectoryServiceRole` with `*string` pointers only; no `management_access_policies` in Patch struct |
| DSR-03 | 50-01, 50-03 | Destroy role cleanly | SATISFIED | `DeleteDirectoryServiceRole`; IsNotFound-tolerant delete |
| DSR-04 | 50-03, 50-05 | Import by name | SATISFIED | `ImportState` with `nullTimeoutsValue()`; import.sh |
| DSR-05 | 50-03 | Drift detection on all mutable/computed fields | SATISFIED | 4 fields detected: group, group_base, management_access_policies, role.name |
| DSR-06 | 50-03 | Read-only data source | SATISFIED | `flashblade_directory_service_role` data source; 2 interface assertions; AddError on not-found |
| DSRM-01 | 50-01, 50-04 | Associate management access policy with directory service role | SATISFIED | `PostManagementAccessPolicyDirectoryServiceRoleMembership`; Lifecycle test |
| DSRM-02 | 50-04 | Destroying membership does not affect policy or role | SATISFIED | DELETE is association-only; mock confirms dissociation without cross-store deletion |
| DSRM-03 | 50-04, 50-05 | Import by composite ID (D-05 corrected: `role_name/policy_name`) | SATISFIED | `parseCompositeID(req.ID, 2)` with role-first; import.sh with D-05 comment; Import test exercises `admin-role/pure:policy/array_admin` |
| DSRM-04 | 50-04 | Both `policy` and `role` trigger RequiresReplace | SATISFIED | `stringplanmodifier.RequiresReplace()` on both; `Update not supported` error |
| DSRM-05 | 50-04 | Read verifies association exists; RemoveResource if gone | SATISFIED | `IsNotFound` → `RemoveResource`; MissingAssociation test verifies `state.Raw.IsNull()` |
| DOC-01 | 50-05 | HCL examples for `flashblade_directory_service_role` | SATISFIED | `examples/resources/flashblade_directory_service_role/resource.tf` and `import.sh` |
| DOC-02 | 50-05 | HCL example for membership (D-05 supersedes original format) | SATISFIED | `examples/resources/.../import.sh` uses `role_name/policy_name`; D-05 supersession comment present |
| DOC-03 | 50-05 | `make docs` generates three new docs files | SATISFIED | `docs/resources/directory_service_role.md`, `docs/data-sources/directory_service_role.md`, `docs/resources/management_access_policy_directory_service_role_membership.md` all present and idempotent |
| QA-01 | 50-01 | >=4 TestUnit_DirectoryServiceRole_* client tests | SATISFIED | 5 tests: Get_Found, Get_NotFound, Post, Patch_Group, Delete |
| QA-02 | 50-01 | >=3 TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_* client tests | SATISFIED | 4 tests: Get_Exists, Get_NotExists, Post, Delete |
| QA-03 | 50-03 | TestUnit_DirectoryServiceRoleResource_Lifecycle | SATISFIED | Present and passing |
| QA-04 | 50-04 | TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembershipResource_MissingAssociation | SATISFIED | Present; `priorState := tfsdk.State`; `readResp.State.Raw.IsNull()` assertion; `do NOT seed` comment |
| QA-05 | 50-03 | TestUnit_DirectoryServiceRoleResource_DriftDetection | SATISFIED | Present and passing |
| QA-06 | 50-05 | `make test` >= 812 | SATISFIED | 814 tests pass |
| QA-07 | 50-03, 50-04, 50-05 | `make lint` clean; both resources declare 4 interface assertions | SATISFIED | 0 lint issues; 4 assertions confirmed on both resources |
| QA-08 | 50-05 | ROADMAP.md updated; CONVENTIONS.md baseline bumped | SATISFIED | ROADMAP lines 100-101 show Done; CONVENTIONS baseline = 814 |

All 22 requirements: SATISFIED.

---

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None detected | — | — | — | — |

No stubs, placeholders, hardcoded empty data flows, or TODO comments found in the phase deliverables. `DirectoryServiceRolePatch` correctly excludes `ManagementAccessPolicies` (readonly on PATCH); `DirectoryServiceRolePost` correctly excludes `Name` (server-generated).

---

## Context Decisions Verification (D-01, D-02, D-05, D-00-b, D-12)

| Decision | Criterion | Status |
|----------|-----------|--------|
| D-01 | `listplanmodifier.RequiresReplace()` on `management_access_policies` | VERIFIED — line 84 of DSR resource |
| D-02 | `role` attr is Computed-only, no plan modifier; code comment `"D-02: role is deprecated per swagger"` present | VERIFIED — `role` schema block has no `PlanModifiers`; comment at line 87 |
| D-05 | Composite ID uses `compositeID(role, policy)` — role FIRST; `roleName := parts[0]` in ImportState | VERIFIED — lines 128, 161, 221 use `compositeID(m.Role.Name, m.Policy.Name)`; `parts[0]` = roleName |
| D-00-b | All 4 interface assertions on both new resources | VERIFIED — 4 assertions on `directoryServiceRoleResource` and `mapDsrMembershipResource` |
| D-12 | Data source has exactly 2 interface assertions | VERIFIED — exactly 2 `var _ datasource.` entries |

---

## Human Verification Required

None. All success criteria are verifiable programmatically and have been confirmed.

### Optional Acceptance Testing (non-blocking)

The following items can only be validated against a real FlashBlade array:

1. **Real API POST name derivation** — Verify that `POST /directory-services/roles` with `management_access_policies=[{"name":"pure:policy/array_admin"}]` actually generates the role name `array_admin` on a live array (Q1 from CONTEXT.md was resolved by the mock using `strings.TrimPrefix("pure:policy/", ...)` heuristic).

2. **Real array DSRM idempotency** — Confirm that `POST /management-access-policies/directory-services/roles` returns 200 (not 409) when the association already exists (Q3 from CONTEXT.md resolved in mock as idempotent-200).

3. **Real import workflow** — `terraform import flashblade_management_access_policy_directory_service_role_membership.example array_admin/pure:policy/array_admin` against a live array.

---

## Gaps Summary

No gaps. All 6 success criteria verified, all 22 requirements satisfied, all 25 artifacts present and substantive, all key links wired, 814 tests passing, 0 lint issues, docs idempotent.

---

_Verified: 2026-04-17_
_Verifier: Claude (gsd-verifier)_
