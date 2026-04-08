---
phase: 41-certificate-groups
plan: 02
subsystem: provider
tags: [go, terraform, flashblade, certificate-groups, rest-api, provider]

# Dependency graph
requires:
  - phase: 41-certificate-groups
    plan: 01
    provides: CertificateGroup/CertificateGroupMember client CRUD, mock handler with Seed/SeedMember
provides:
  - flashblade_certificate_group resource (CRD, no Update)
  - flashblade_certificate_group data source
  - flashblade_certificate_group_member resource (CRD, 4 interface assertions)
  - 6 provider unit tests (TestUnit_CertificateGroup*)
  - Provider registration in provider.go
  - HCL examples for all 3 resources
  - Regenerated docs
affects:
  - ROADMAP.md (Certificate Groups moved to Implemented)
  - CONVENTIONS.md (baseline updated to 735)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - CRD resource with no Update: returns AddError "cannot be updated — all attributes force a new resource"
    - CRD inline null timeouts (create/read/delete keys only) in ImportState — same as tls_policy_member
    - Realms volatile list field with NO UseStateForUnknown — empty-safe mapping via ListValueMust
    - Member resource composite ID (groupName/certificateName) via parseCompositeID

key-files:
  created:
    - internal/provider/certificate_group_resource.go
    - internal/provider/certificate_group_data_source.go
    - internal/provider/certificate_group_member_resource.go
    - internal/provider/certificate_group_resource_test.go
    - internal/provider/certificate_group_data_source_test.go
    - internal/provider/certificate_group_member_resource_test.go
    - examples/resources/flashblade_certificate_group/resource.tf
    - examples/resources/flashblade_certificate_group/import.sh
    - examples/data-sources/flashblade_certificate_group/data-source.tf
    - examples/resources/flashblade_certificate_group_member/resource.tf
    - examples/resources/flashblade_certificate_group_member/import.sh
    - docs/resources/certificate_group.md
    - docs/resources/certificate_group_member.md
    - docs/data-sources/certificate_group.md
  modified:
    - internal/provider/provider.go
    - ROADMAP.md
    - CONVENTIONS.md

key-decisions:
  - "CRD resource: Update method returns AddError (no PATCH in FlashBlade certificate-groups API)"
  - "Realms field is Computed with NO UseStateForUnknown — volatile (set by array), must detect drift"
  - "CRD inline null timeouts in ImportState: only create/read/delete keys (matches schema, not the 4-key nullTimeoutsValue helper)"
  - "listToStringSlice reused from tls_policy_resource.go for realms drift comparison"

requirements-completed: [CERTG-01, CERTG-02, CERTG-03, CERTG-04, CERTG-05]

# Metrics
duration: ~7min
completed: 2026-04-08
---

# Phase 41 Plan 02: Certificate Groups Provider Layer Summary

**flashblade_certificate_group resource + data source + flashblade_certificate_group_member resource: CRD-only pattern with 4-interface assertions, drift detection on realms, composite ID import, 735 total tests**

## Performance

- **Duration:** ~7 min
- **Started:** 2026-04-08T09:30:39Z
- **Completed:** 2026-04-08T09:37:22Z
- **Tasks:** 3
- **Files modified:** 17

## Accomplishments

- flashblade_certificate_group resource: 4 interface assertions, Version 0 schema, CRD timeouts, Update returns AddError, drift detection on name/realms via tflog.Debug, ImportState with CRD inline null timeouts
- flashblade_certificate_group data source: 2 interface assertions, inline realms mapping with empty-safe ListValueMust pattern
- flashblade_certificate_group_member resource: 4 interface assertions including ResourceWithUpgradeState, composite ID import (groupName/certificateName), CRD timeouts
- 6 TestUnit_CertificateGroup* tests: Lifecycle, Import, DriftDetection for resource; Basic for DS; Lifecycle, Import for member
- Provider registration: NewCertificateGroupResource, NewCertificateGroupMemberResource, NewCertificateGroupDataSource added to provider.go
- HCL examples: resource.tf, import.sh for both resources; data-source.tf for DS
- make docs: generated docs/resources/certificate_group.md, certificate_group_member.md, data-sources/certificate_group.md
- ROADMAP.md: Certificate Groups moved to Implemented (Done), header counters updated
- CONVENTIONS.md: test baseline updated from 716 to 735

## Task Commits

1. **Task 1: certificate_group resource + data source** - `6fa75cf` (feat)
2. **Task 2: certificate_group_member resource** - `f2a6ddc` (feat)
3. **Task 3: tests, registration, examples, docs** - `81d9c07` (feat)

## Files Created/Modified

### Created
- `internal/provider/certificate_group_resource.go` — CRD resource, 4 interface assertions, drift detection
- `internal/provider/certificate_group_data_source.go` — DS, 2 interface assertions, inline mapping
- `internal/provider/certificate_group_member_resource.go` — CRD member resource, 4 interface assertions
- `internal/provider/certificate_group_resource_test.go` — 3 tests
- `internal/provider/certificate_group_data_source_test.go` — 1 test
- `internal/provider/certificate_group_member_resource_test.go` — 2 tests
- HCL examples for all 3 resources (resource.tf, import.sh / data-source.tf)
- Generated docs for all 3 resources

### Modified
- `internal/provider/provider.go` — registered 3 new constructors
- `ROADMAP.md` — Certificate Groups: Candidate → Done, counters updated
- `CONVENTIONS.md` — baseline: 716 → 735

## Decisions Made

- CRD resource: Update method returns AddError (no PATCH in FlashBlade certificate-groups API)
- Realms field is Computed with NO UseStateForUnknown — volatile (set by array), must detect drift
- CRD inline null timeouts in ImportState: only create/read/delete keys (matches schema, not the 4-key nullTimeoutsValue helper)
- listToStringSlice reused from tls_policy_resource.go for realms drift comparison

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

---
*Phase: 41-certificate-groups*
*Completed: 2026-04-08*

## Self-Check: PASSED
