---
phase: 41-certificate-groups
plan: 01
subsystem: client
tags: [go, terraform, flashblade, certificate-groups, rest-api, mock-handler]

# Dependency graph
requires:
  - phase: 40-tls-policies
    provides: TlsPolicyMember store+facade pattern, generic helper usage (getOneByName, postOne, ListResponse pagination)
provides:
  - CertificateGroup, CertificateGroupPost, CertificateGroupMember model structs in models_network.go
  - GetCertificateGroup, PostCertificateGroup, DeleteCertificateGroup client CRUD methods
  - ListCertificateGroupMembers with pagination loop
  - PostCertificateGroupMember, DeleteCertificateGroupMember member management methods
  - Thread-safe mock handler with Seed/SeedMember and CertificateGroupStoreFacade
  - 7 client unit tests (TestUnit_CertificateGroup_*)
affects:
  - 41-certificate-groups/41-02 (provider layer depends on this client layer)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - CertificateGroupPost as empty struct — API creates group from name alone, no body fields
    - /certificate-groups/certificates registered before /certificate-groups in ServeMux to avoid prefix collision
    - Facade pattern for cross-package test access to unexported store type

key-files:
  created:
    - internal/client/certificate_groups.go
    - internal/testmock/handlers/certificate_groups.go
    - internal/client/certificate_groups_test.go
  modified:
    - internal/client/models_network.go

key-decisions:
  - "CertificateGroupPost is empty struct — FlashBlade API creates certificate groups from ?names= query param alone, no writable body fields"
  - "Register /certificate-groups/certificates before /certificate-groups in ServeMux to avoid prefix collision (longer path wins)"
  - "Realms field initialized to []string{} (empty slice, not nil) in mock POST to match API behavior"
  - "Member store keyed by group name matches TLS policy member pattern for consistency"

patterns-established:
  - "Empty POST body struct: CertificateGroupPost{} with postOne[CertificateGroupPost, CertificateGroup] for name-only resource creation"

requirements-completed: [CERTG-01, CERTG-02, CERTG-03, CERTG-05]

# Metrics
duration: ~5min
completed: 2026-04-08
---

# Phase 41 Plan 01: Certificate Groups Client Layer Summary

**FlashBlade certificate group client layer: 6 CRUD+member methods via getOneByName/postOne generics, thread-safe mock handler with facade, 7 passing unit tests**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-04-08T09:25:00Z
- **Completed:** 2026-04-08T09:28:21Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Model structs appended to models_network.go: CertificateGroup (GET), CertificateGroupPost (empty POST body), CertificateGroupMember (association)
- 6 client methods in certificate_groups.go using getOneByName/postOne generics with pagination loop for ListCertificateGroupMembers
- Mock handler with thread-safe store, correct empty-list HTTP 200 behavior, Seed/SeedMember, and cross-package CertificateGroupStoreFacade
- 7 TestUnit_CertificateGroup* tests all pass; total client package count 279 (up from 272)

## Task Commits

1. **Task 1: Model structs + client CRUD methods** - `2bfcb17` (feat)
2. **Task 2: Mock handler** - `e6521c6` (feat)
3. **Task 3: Client unit tests** - `1b455d3` (test)

## Files Created/Modified

- `internal/client/models_network.go` - Appended CertificateGroup, CertificateGroupPost, CertificateGroupMember structs
- `internal/client/certificate_groups.go` - 6 CRUD+member methods using getOneByName/postOne generics
- `internal/testmock/handlers/certificate_groups.go` - Thread-safe mock with Seed/SeedMember/CertificateGroupStoreFacade
- `internal/client/certificate_groups_test.go` - 7 unit tests with TestUnit_CertificateGroup* naming

## Decisions Made

- CertificateGroupPost is an empty struct — FlashBlade API creates certificate groups from the ?names= query param alone with no writable body fields
- /certificate-groups/certificates registered before /certificate-groups in ServeMux to prevent prefix collision (Go ServeMux longer-path-wins rule)
- Realms initialized to []string{} (not nil) in mock POST to match API behavior and avoid nil pointer in tests
- Member store keyed by group name, consistent with TLS policy member pattern established in phase 40

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

- Client layer complete; plan 41-02 (provider layer) can implement flashblade_certificate_group resource + data source + flashblade_certificate_group_member resource
- All 6 client methods and mock handlers are ready for provider-level mocked integration tests

---
*Phase: 41-certificate-groups*
*Completed: 2026-04-08*

## Self-Check: PASSED

All files verified on disk, all task commits confirmed in git history.
