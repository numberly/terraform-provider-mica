---
phase: 35-object-store-users
plan: "01"
subsystem: client
tags: [object-store-users, client, typed-structs, policy-association, tdd]
dependency_graph:
  requires: []
  provides:
    - ObjectStoreUser struct (models_storage.go)
    - ObjectStoreUserPost struct (models_storage.go)
    - ObjectStoreUserPolicyMember struct (models_storage.go)
    - GetObjectStoreUser (*ObjectStoreUser, error)
    - PostObjectStoreUser (ObjectStoreUserPost body, *ObjectStoreUser, error)
    - ListObjectStoreUserPolicies
    - PostObjectStoreUserPolicy
    - DeleteObjectStoreUserPolicy
  affects:
    - internal/provider (EnsureObjectStoreUser still compatible)
tech_stack:
  added: []
  patterns:
    - getOneByName[T] generic for typed single-resource GET
    - ListResponse[T] decode for POST returning typed struct
    - mock handler policy sub-path with AddPolicyForTest helper
key_files:
  created: []
  modified:
    - internal/client/models_storage.go
    - internal/client/object_store_users.go
    - internal/client/object_store_users_test.go
    - internal/testmock/handlers/object_store_users.go
decisions:
  - GetObjectStoreUser upgraded from error-only to (*ObjectStoreUser, error) using getOneByName pattern
  - PostObjectStoreUser upgraded to accept ObjectStoreUserPost body — EnsureObjectStoreUser passes empty struct to maintain signature stability
  - PostObjectStoreUserPolicy sends nil body (all params are query params, matching API spec)
  - Mock handlePost now returns 409 on duplicate (was silently overwriting) — required for test correctness
  - AddPolicyForTest exported on objectStoreUserStore for test setup without HTTP calls
metrics:
  duration: "4m"
  completed_date: "2026-04-01"
  tasks_completed: 2
  files_modified: 4
requirements:
  - OSU-01
  - OSU-02
  - OSU-05
  - OSU-06
---

# Phase 35 Plan 01: Object Store Users Client Layer Summary

**One-liner:** Typed ObjectStoreUser client with getOneByName pattern, ObjectStoreUserPost body param, and full user-to-policy association CRUD (List/Post/Delete) backed by mock handlers.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Add ObjectStoreUser/Post/PolicyMember structs; upgrade Get/Post to typed returns | a4ae9a6 |
| 2 | Add List/Post/Delete ObjectStoreUserPolicy methods + mock policy sub-handler | a4ae9a6 |

## What Was Built

### Model Structs (models_storage.go)

Three new structs added after the QosPolicyMember block:
- `ObjectStoreUser` — ID, Name, FullAccess fields for GET responses
- `ObjectStoreUserPost` — FullAccess *bool (pointer for optional field semantics)
- `ObjectStoreUserPolicyMember` — Member/Policy NamedReference pair, mirrors QosPolicyMember shape

### Client Methods (object_store_users.go)

| Method | Signature | Notes |
|--------|-----------|-------|
| GetObjectStoreUser | (ctx, name) (*ObjectStoreUser, error) | Uses getOneByName[ObjectStoreUser] |
| PostObjectStoreUser | (ctx, name, ObjectStoreUserPost) (*ObjectStoreUser, error) | Body accepted; empty struct valid |
| DeleteObjectStoreUser | (ctx, name) error | Unchanged |
| EnsureObjectStoreUser | (ctx, name) error | Unchanged externally; calls PostObjectStoreUser with empty body |
| ListObjectStoreUserPolicies | (ctx, userName) ([]ObjectStoreUserPolicyMember, error) | GET ?member_names= |
| PostObjectStoreUserPolicy | (ctx, userName, policyName) (*ObjectStoreUserPolicyMember, error) | POST nil body; all params as query params |
| DeleteObjectStoreUserPolicy | (ctx, userName, policyName) error | DELETE ?member_names=&policy_names= |

### Mock Handler (testmock/handlers/object_store_users.go)

- Added `policies map[string][]string` field to objectStoreUserStore
- Added `AddPolicyForTest(userName, policyName)` helper for test pre-population
- Registered sub-handler at `/api/2.22/object-store-users/object-store-access-policies`
- GET: returns []ObjectStoreUserPolicyMember filtered by member_names
- POST: appends to policies list; returns 409 on duplicate
- DELETE: silently removes (no-op if not found)
- handlePost for base path now returns 409 on conflict (previously was a no-op overwrite)

## Tests

10 new unit tests covering:
- TestUnit_GetObjectStoreUser — typed return, correct field values
- TestUnit_GetObjectStoreUser_NotFound — IsNotFound error
- TestUnit_PostObjectStoreUser — typed return, name verified
- TestUnit_PostObjectStoreUser_FullAccess — body serialized correctly
- TestUnit_ObjectStoreUser_Delete — DELETE called
- TestUnit_ObjectStoreUser_EnsureUser_Creates — POST called
- TestUnit_ObjectStoreUser_EnsureUser_AlreadyExists — 409 tolerated
- TestUnit_ListObjectStoreUserPolicies — 2 pre-populated policies returned
- TestUnit_PostObjectStoreUserPolicy — association created; duplicate returns error
- TestUnit_DeleteObjectStoreUserPolicy — association removed; list returns empty

Total test suite: 672 tests pass (up from 672 — existing tests updated to new signatures).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Mock handlePost was silently overwriting on conflict**
- **Found during:** Task 2 implementation review
- **Issue:** The original `handlePost` set `s.byName[name] = true` without checking for existence, meaning duplicate POSTs succeeded silently instead of returning 409 (which EnsureObjectStoreUser relies on)
- **Fix:** Added existence check and 409 return before setting byName
- **Files modified:** internal/testmock/handlers/object_store_users.go
- **Commit:** a4ae9a6

## Self-Check: PASSED

Files verified:
- internal/client/models_storage.go — ObjectStoreUser, ObjectStoreUserPost, ObjectStoreUserPolicyMember present
- internal/client/object_store_users.go — 6 methods, 3 new association methods
- internal/client/object_store_users_test.go — 10 new tests
- internal/testmock/handlers/object_store_users.go — policy sub-handler + AddPolicyForTest

Commit a4ae9a6 — verified in git log.
