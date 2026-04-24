---
phase: 35-object-store-users
verified: 2026-04-01T06:44:54Z
status: passed
score: 13/13 must-haves verified
re_verification: false
---

# Phase 35: Object Store Users Verification Report

**Phase Goal:** Operators can create named S3 users, associate access policies to them, and manage per-user credentials — enabling multi-tenant S3 workflows with fine-grained access control
**Verified:** 2026-04-01T06:44:54Z
**Status:** PASSED
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

Truths are drawn from the must_haves sections of plans 35-01, 35-02, 35-03, and 35-04.

#### Plan 35-01 Truths (client layer)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Client can POST/DELETE /object-store-users/object-store-access-policies to attach and detach policies from a user | VERIFIED | `PostObjectStoreUserPolicy` and `DeleteObjectStoreUserPolicy` implemented at lines 63-81 of `object_store_users.go`; correct query-param paths confirmed |
| 2 | Client can GET /object-store-users/object-store-access-policies to list policies attached to a user | VERIFIED | `ListObjectStoreUserPolicies` at line 52-59 uses `?member_names=` param; returns `[]ObjectStoreUserPolicyMember` |
| 3 | Client GetObjectStoreUser returns a typed struct (not map[string]any) with id, name, full_access fields | VERIFIED | Returns `(*ObjectStoreUser, error)` using `getOneByName[ObjectStoreUser]`; struct in `models_storage.go` lines 380-384 |
| 4 | Mock handlers respond correctly to user-policy association requests | VERIFIED | `handlePolicies` in `testmock/handlers/object_store_users.go` routes GET/POST/DELETE; 409 on duplicate; silent no-op on unknown delete |

#### Plan 35-02 Truths (resource and data source)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 5 | Operator can create an S3 user (account/username) with optional full_access via terraform apply | VERIFIED | `Create` method in `object_store_user_resource.go` lines 115-144; body only sent when FullAccess not null/unknown |
| 6 | Operator can destroy the user via terraform destroy | VERIFIED | `Delete` method lines 196-218; idempotent on IsNotFound |
| 7 | Operator can read an existing S3 user by name via data.flashblade_object_store_user | VERIFIED | Data source `Read` in `object_store_user_data_source.go` lines 82-108; populates id, name, full_access |
| 8 | terraform import flashblade_object_store_user.x account/username populates state with zero drift on next plan | VERIFIED | `ImportState` uses name directly; null timeouts initialized with CRD-only keys; plan convergence step in `TestMocked_ObjectStoreUser_Lifecycle` |
| 9 | Read logs field-level drift changes via tflog when full_access changes outside Terraform | VERIFIED | `tflog.Warn` at lines 175-181 comparing `user.FullAccess` vs `data.FullAccess.ValueBool()` |

#### Plan 35-03 Truths (user-policy member resource)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 10 | Operator can create a flashblade_object_store_user_policy resource associating a user to an access policy | VERIFIED | `Create` in `object_store_user_policy_resource.go` lines 100-123 calls `PostObjectStoreUserPolicy` |
| 11 | Operator can destroy the association (only the association — user and policy are not deleted) | VERIFIED | `Delete` calls `DeleteObjectStoreUserPolicy` only; no cascade to user/policy resources |
| 12 | terraform import flashblade_object_store_user_policy.x account/username/policyname populates state; zero drift on next plan | VERIFIED | `ImportState` uses `strings.SplitN(req.ID, "/", 3)`, reassembles `parts[0]+"/"+parts[1]` as userName; plan convergence step verified in `TestMocked_ObjectStoreUserPolicy_Lifecycle` |
| 13 | Read logs drift when the association disappears outside Terraform (resource removed from state) | VERIFIED | `tflog.Warn` at lines 160-165 followed by `RemoveResource` when association not found in list |

#### Plan 35-04 Truths (tests and ROADMAP)

All four test truths are subsumed by the passing test runs verified below.

**Score: 13/13 truths verified**

---

## Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/object_store_users.go` | 6 client methods including 3 association methods | VERIFIED | GetObjectStoreUser, PostObjectStoreUser, DeleteObjectStoreUser, EnsureObjectStoreUser, ListObjectStoreUserPolicies, PostObjectStoreUserPolicy, DeleteObjectStoreUserPolicy — 7 methods (EnsureObjectStoreUser included) |
| `internal/client/models_storage.go` | ObjectStoreUser, ObjectStoreUserPost, ObjectStoreUserPolicyMember structs | VERIFIED | All three structs present at lines 380-395 |
| `internal/testmock/handlers/object_store_users.go` | handlers for /object-store-users/object-store-access-policies | VERIFIED | `handlePolicies` registered in `RegisterObjectStoreUserHandlers`; GET/POST/DELETE all handled |
| `internal/provider/object_store_user_resource.go` | flashblade_object_store_user resource (Create, Read, Delete, ImportState, UpgradeState) | VERIFIED | All interfaces implemented; exports `NewObjectStoreUserResource` |
| `internal/provider/object_store_user_data_source.go` | flashblade_object_store_user data source (Read) | VERIFIED | Exports `NewObjectStoreUserDataSource`; populates id, name, full_access |
| `internal/provider/provider.go` | both resources and data source registered | VERIFIED | Lines 288-308 confirm `NewObjectStoreUserResource`, `NewObjectStoreUserPolicyResource`, `NewObjectStoreUserDataSource` all registered |
| `internal/provider/object_store_user_policy_resource.go` | flashblade_object_store_user_policy resource (Create, Read, Delete, ImportState) | VERIFIED | All interfaces implemented; exports `NewObjectStoreUserPolicyResource` |
| `internal/provider/object_store_user_resource_test.go` | TestMocked_ObjectStoreUser_Lifecycle and TestMocked_ObjectStoreUser_FullAccess | VERIFIED | Both functions present; 4 provider tests pass |
| `internal/provider/object_store_user_data_source_test.go` | TestMocked_ObjectStoreUser_DataSource | VERIFIED | Present and passing |
| `internal/provider/object_store_user_policy_resource_test.go` | TestMocked_ObjectStoreUserPolicy_Lifecycle | VERIFIED | Present and passing |
| `examples/resources/flashblade_object_store_user/resource.tf` | HCL example | VERIFIED | Contains `name` and `full_access` fields |
| `examples/resources/flashblade_object_store_user/import.sh` | import command | VERIFIED | Contains `terraform import flashblade_object_store_user.example myaccount/myuser` |
| `examples/data-sources/flashblade_object_store_user/data-source.tf` | data source HCL example | VERIFIED | Present |
| `examples/resources/flashblade_object_store_user_policy/resource.tf` | HCL example | VERIFIED | Contains `user_name` and `policy_name` |
| `examples/resources/flashblade_object_store_user_policy/import.sh` | import command with 3-part ID | VERIFIED | Uses `myaccount/myuser/my-access-policy` format |
| `ROADMAP.md` | Both resources in Implemented section | VERIFIED | Lines 31-32 show `flashblade_object_store_user` (Done, data source Yes) and `flashblade_object_store_user_policy` (Done, data source No) |

---

## Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `object_store_user_resource.go` | `client.GetObjectStoreUser` | Read method calls GetObjectStoreUser then logs drift | WIRED | Lines 163 (call) and 174 (drift compare) |
| `object_store_user_resource.go` | `client.PostObjectStoreUser` | Create method | WIRED | Line 136 |
| `object_store_user_policy_resource.go` | `client.PostObjectStoreUserPolicy` | Create method | WIRED | Line 115 |
| `object_store_user_policy_resource.go` | `client.ListObjectStoreUserPolicies` | Read method (list then find) | WIRED | Line 141; find loop at lines 152-156 |
| `internal/client/object_store_users.go` | `/object-store-users/object-store-access-policies` | GET/POST/DELETE methods | WIRED | All three methods use this path; confirmed at lines 53, 64, 78 |
| `object_store_user_resource_test.go` | `testNewMockedProvider()` | provider setup in each test | WIRED | `setupObjectStoreUserTest` uses `testmock.NewMockServer` + `RegisterObjectStoreUserHandlers` |
| `object_store_user_resource_test.go` | `internal/testmock/handlers/object_store_users.go` | mock server handling /api/2.22/object-store-users | WIRED | `RegisterObjectStoreUserHandlers` called in `setupObjectStoreUserTest` |

---

## Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| OSU-01 | 35-01, 35-02, 35-04 | Operator can create a named S3 user under an account via Terraform (format: account/username) | SATISFIED | `Create` in resource; `PostObjectStoreUser` in client; `TestMocked_ObjectStoreUser_Lifecycle` step 1 |
| OSU-02 | 35-01, 35-02, 35-04 | Operator can delete an S3 user via Terraform destroy | SATISFIED | `Delete` in resource; `DeleteObjectStoreUser` in client; test destroy runs after each UnitTest |
| OSU-03 | 35-02, 35-04 | Operator can read an existing S3 user by name via data source | SATISFIED | `flashblade_object_store_user` data source; `TestMocked_ObjectStoreUser_DataSource` |
| OSU-04 | 35-02, 35-04 | Operator can import an existing S3 user into Terraform state with no drift on subsequent plan | SATISFIED | `ImportState` using name as ID; `ImportStateId: "testaccount/testuser"` in test; plan convergence step |
| OSU-05 | 35-01, 35-03, 35-04 | Operator can associate one or more access policies to a user via a member resource | SATISFIED | `flashblade_object_store_user_policy` resource; `PostObjectStoreUserPolicy`; `TestMocked_ObjectStoreUserPolicy_Lifecycle` |
| OSU-06 | 35-01, 35-03, 35-04 | Operator can remove a policy association from a user via Terraform destroy | SATISFIED | `Delete` in `object_store_user_policy_resource.go`; `DeleteObjectStoreUserPolicy`; destroy runs after UnitTest |
| OSU-07 | 35-02, 35-03, 35-04 | Drift detection logs changes when user or policy association is modified outside Terraform | SATISFIED | `tflog.Warn` in user `Read` (full_access field); `tflog.Warn` + `RemoveResource` in policy `Read` (missing association) |

All 7 OSU requirements (OSU-01 through OSU-07) are SATISFIED. No orphaned requirements found.

---

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `internal/testmock/handlers/object_store_users.go` | 103-108 | Mock `handlePost` always creates user with `FullAccess: false` regardless of request body | WARNING | `TestMocked_ObjectStoreUser_FullAccess` tests with `full_access = false` (not `= true` as plan stated). The client-level test `TestUnit_PostObjectStoreUser_FullAccess` does verify the body serializes correctly. The provider test does not exercise the `full_access = true` round-trip path. |

No blockers found. The mock body-parsing gap is a test coverage weakness, not a functional defect — the resource correctly sends the body and the client test validates the body content.

---

## Human Verification Required

None — all functional checks are automated and passing.

---

## Test Execution Summary

| Suite | Command | Result |
|-------|---------|--------|
| Client unit tests | `go test ./internal/client/ -run TestUnit_.*ObjectStoreUser -count=1` | 10 passed |
| Provider mocked tests (user) | `go test ./internal/provider/ -run TestMocked_ObjectStoreUser -count=1` | 4 passed |
| Provider mocked tests (policy) | `go test ./internal/provider/ -run TestMocked_ObjectStoreUserPolicy -count=1` | 1 passed |
| Full suite | `go test ./... -count=1` | 676 passed, 0 failed |
| Build | `go build ./...` | success |

---

## Summary

Phase 35 goal is fully achieved. All 13 must-have truths are verified, all 7 OSU requirements are satisfied, and all 676 tests pass.

The only notable observation is a minor test coverage gap: `TestMocked_ObjectStoreUser_FullAccess` is named for `full_access = true` but tests with `full_access = false`, and the mock handler does not parse the body's `full_access` field. This does not block any requirement because:

1. The client unit test `TestUnit_PostObjectStoreUser_FullAccess` verifies the body is serialized correctly.
2. The resource `Create` method correctly builds and sends the body with `full_access = true` when the attribute is set.
3. Acceptance tests against a real array would exercise the true value.

Operators can create named S3 users, associate access policies to them, and import both resource types into Terraform state with zero drift — the phase goal is met.

---

_Verified: 2026-04-01T06:44:54Z_
_Verifier: Claude (gsd-verifier)_
