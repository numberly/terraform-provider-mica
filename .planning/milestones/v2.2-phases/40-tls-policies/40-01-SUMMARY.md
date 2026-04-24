---
phase: 40-tls-policies
plan: 01
subsystem: client
tags: [tls, go, client, mock, unit-tests]

requires:
  - phase: 37-remote-credentials-replica-link-enhancement
    provides: patterns for member CRUD (ListMembers/PostMember/DeleteMember) and **NamedReference patch semantics

provides:
  - TlsPolicy, TlsPolicyPost, TlsPolicyPatch, TlsPolicyMember model structs in models_network.go
  - 7 client methods in tls_policies.go (Get/Post/Patch/Delete TlsPolicy + List/Post/Delete TlsPolicyMember)
  - Mock handler in handlers/tls_policies.go with TlsPolicyStoreFacade for cross-package test access
  - 8 client unit tests in tls_policies_test.go

affects: [40-tls-policies-02]

tech-stack:
  added: []
  patterns:
    - TlsPolicyStoreFacade wraps unexported store to enable cross-package Seed/SeedMember calls (matches targetStoreFacade pattern)
    - Member POST/DELETE use /network-interfaces/tls-policies; member GET listing uses /tls-policies/members
    - TlsPolicyPatch uses **NamedReference for nullable reference fields (outer nil=omit, outer non-nil+inner nil=set null)

key-files:
  created:
    - internal/client/tls_policies.go
    - internal/testmock/handlers/tls_policies.go
    - internal/client/tls_policies_test.go
  modified:
    - internal/client/models_network.go

key-decisions:
  - "TlsPolicyStoreFacade added to handlers package (exported) to expose Seed/SeedMember from unexported tlsPolicyStore — matches targetStoreFacade pattern from Phase 36"
  - "Member POST/DELETE route to /network-interfaces/tls-policies per API spec; member GET listing uses /tls-policies/members"
  - "Mock GET for /tls-policies returns empty list HTTP 200 when name not found (not 404) — consistent with FlashBlade API and getOneByName behavior"

patterns-established:
  - "TlsPolicyStoreFacade: exported wrapper over unexported store for cross-package test access without making store public"

requirements-completed: [TLSP-01, TLSP-02, TLSP-05]

duration: <1min
completed: 2026-04-03
---

# Phase 40 Plan 01: TLS Policies — Client Layer Summary

**TLS policy client layer: 4 model structs, 7 CRUD methods (policy + member), mock handler with 3 endpoints, and 8 passing unit tests**

## Performance

- **Duration:** ~21 seconds
- **Started:** 2026-04-03T00:00:00Z
- **Completed:** 2026-04-03T00:00:21Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added TlsPolicy, TlsPolicyPost, TlsPolicyPatch, TlsPolicyMember structs to models_network.go with correct pointer rules
- Implemented 7 client methods following getOneByName generic and qos_policies.go ListQosPolicyMembers pagination pattern
- Mock handler registers 3 endpoints (policy CRUD, member GET, member POST/DELETE) with empty-list GET on not-found
- 8 unit tests (TestUnit_TlsPolicy_* prefix) pass; total suite grows from 692 to 700

## Task Commits

1. **Task 1: Model structs and client CRUD** - `da7b69c` (feat)
2. **Task 2: Mock handler and client unit tests** - `9d62726` (test)

## Files Created/Modified

- `internal/client/models_network.go` - Added TlsPolicy/TlsPolicyPost/TlsPolicyPatch/TlsPolicyMember structs
- `internal/client/tls_policies.go` - 7 client methods (GetTlsPolicy, PostTlsPolicy, PatchTlsPolicy, DeleteTlsPolicy, ListTlsPolicyMembers, PostTlsPolicyMember, DeleteTlsPolicyMember)
- `internal/testmock/handlers/tls_policies.go` - Mock handler + TlsPolicyStoreFacade for cross-package test access
- `internal/client/tls_policies_test.go` - 8 unit tests with TestUnit_ prefix

## Decisions Made

- TlsPolicyStoreFacade exported wrapper added to handlers package to expose Seed/SeedMember from unexported tlsPolicyStore, matching the targetStoreFacade pattern from Phase 36.
- Member POST/DELETE route to `/network-interfaces/tls-policies` per FlashBlade API spec; member GET listing uses `/tls-policies/members`.
- Mock GET for `/tls-policies` returns empty list with HTTP 200 when name not found — consistent with FlashBlade API behavior and getOneByName generic.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

- Client layer complete — Plan 40-02 can consume GetTlsPolicy, PostTlsPolicy, PatchTlsPolicy, DeleteTlsPolicy, ListTlsPolicyMembers, PostTlsPolicyMember, DeleteTlsPolicyMember
- Mock handler ready for provider resource tests in Plan 40-02 via RegisterTlsPolicyHandlers + TlsPolicyStoreFacade

---
*Phase: 40-tls-policies*
*Completed: 2026-04-03*
