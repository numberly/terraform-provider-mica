---
phase: 40-tls-policies
plan: 02
subsystem: provider
tags: [tls, terraform-resource, terraform-datasource, go, provider, tests, docs]

requires:
  - phase: 40-tls-policies-01
    provides: TlsPolicy client layer (GetTlsPolicy, PostTlsPolicy, PatchTlsPolicy, DeleteTlsPolicy, ListTlsPolicyMembers, PostTlsPolicyMember, DeleteTlsPolicyMember), mock handler with TlsPolicyStoreFacade

provides:
  - flashblade_tls_policy resource with full CRUD, drift detection on 10 fields, import by name
  - flashblade_tls_policy data source reading all config attributes by name
  - flashblade_tls_policy_member CRD resource for NI assignment with composite import
  - 7 provider/data-source tests (Lifecycle, Import, DriftDetection, Basic, Read_NotFound)
  - HCL examples for all 3 components
  - Generated docs in docs/resources/ and docs/data-sources/
  - ROADMAP.md updated: TLS Policies marked Done

affects: [phase-40-verification]

tech-stack:
  added: []
  patterns:
    - "TLS policy resource: flat string for NamedReference fields (appliance_certificate, trusted_client_certificate_authority) — consistent with ca_certificate_group in target_resource.go"
    - "**NamedReference PATCH semantics: outer nil=omit, outer non-nil+inner nil=set to null, outer non-nil+inner non-nil=set value — used for appliance_certificate and trusted_client_certificate_authority in Update"
    - "listToStringSlice helper: shared within tls_policy_resource.go to convert types.List to []string for drift comparison and PATCH body"
    - "tlsPolicyMemberResource: CRD-only following qosPolicyMemberResource pattern; inline null timeouts in ImportState (create/read/delete keys)"
    - "is_local has UseStateForUnknown — computed, stable after creation (unlike policy_type which is Computed-only without USFUK)"

key-files:
  created:
    - internal/provider/tls_policy_resource.go
    - internal/provider/tls_policy_data_source.go
    - internal/provider/tls_policy_member_resource.go
    - internal/provider/tls_policy_resource_test.go
    - internal/provider/tls_policy_data_source_test.go
    - internal/provider/tls_policy_member_resource_test.go
    - examples/resources/flashblade_tls_policy/resource.tf
    - examples/resources/flashblade_tls_policy/import.sh
    - examples/data-sources/flashblade_tls_policy/data-source.tf
    - examples/resources/flashblade_tls_policy_member/resource.tf
    - examples/resources/flashblade_tls_policy_member/import.sh
    - docs/resources/tls_policy.md
    - docs/resources/tls_policy_member.md
    - docs/data-sources/tls_policy.md
  modified:
    - internal/provider/provider.go
    - ROADMAP.md

decisions:
  - "is_local gets UseStateForUnknown: it is a computed field set by the API at creation and does not change — same treatment as id in other resources"
  - "policy_type is Computed-only with no UseStateForUnknown: it can in theory change (volatile API-side field) — drift detection still logs changes"
  - "listToStringSlice defined in tls_policy_resource.go (not helpers.go) since it is the only consumer — no premature extraction"
  - "tlsPolicyMemberResource has no UpgradeState (CRD-only, version 0, no state migration needed)"

metrics:
  duration: ~8 minutes
  completed: 2026-04-03T07:34:51Z
  tasks: 2
  files: 15 created + 2 modified
  tests_added: 7
  tests_total: 716
---

# Phase 40 Plan 02: TLS Policies — Provider Layer Summary

**TLS policy provider layer: resource + data source + member resource with full CRUD, drift detection, composite import, 7 tests, HCL examples, and generated docs**

## Performance

- **Duration:** ~8 minutes
- **Started:** 2026-04-03T07:26:00Z
- **Completed:** 2026-04-03T07:34:51Z
- **Tasks:** 2
- **Files:** 15 created, 2 modified

## Accomplishments

- Implemented `flashblade_tls_policy` resource with all 4 interfaces, schema v0, UseStateForUnknown only on `id` and `is_local`, drift detection on all 10 mutable/computed fields via tflog.Debug
- Implemented `flashblade_tls_policy` data source with 2 interfaces, no timeouts, all fields Computed except `name`
- Implemented `flashblade_tls_policy_member` CRD resource (3 interfaces — no UpgradeState needed), composite import `policyName/memberName`, list+find Read, inline null timeouts on import
- 7 tests with TestUnit_ prefix: 3 resource (Lifecycle, Import, DriftDetection), 1 data source (Basic), 3 member (Lifecycle, Read_NotFound, Import)
- HCL examples for all 3 components including import.sh with correct identifiers
- `make docs` regenerated docs/resources/tls_policy.md, tls_policy_member.md, docs/data-sources/tls_policy.md
- ROADMAP.md updated: TLS Policies Planned → Done, covered count 36→37, coverage 66%→67%
- Registered NewTlsPolicyResource, NewTlsPolicyMemberResource, NewTlsPolicyDataSource in provider.go

## Task Commits

1. **Task 1: Resource, data source, member resource, and provider registration** - `e80e53f` (feat)
2. **Task 2: Tests, HCL examples, docs, ROADMAP** - `e53552b` (feat)

## Files Created/Modified

- `internal/provider/tls_policy_resource.go` — full CRUD resource with drift detection and import
- `internal/provider/tls_policy_data_source.go` — read-only data source
- `internal/provider/tls_policy_member_resource.go` — CRD member resource with composite import
- `internal/provider/tls_policy_resource_test.go` — 3 tests: Lifecycle, Import, DriftDetection
- `internal/provider/tls_policy_data_source_test.go` — 1 test: Basic
- `internal/provider/tls_policy_member_resource_test.go` — 3 tests: Lifecycle, Read_NotFound, Import
- `internal/provider/provider.go` — registered 3 new components
- `examples/resources/flashblade_tls_policy/{resource.tf,import.sh}` — resource HCL examples
- `examples/data-sources/flashblade_tls_policy/data-source.tf` — data source HCL example
- `examples/resources/flashblade_tls_policy_member/{resource.tf,import.sh}` — member HCL examples
- `docs/resources/tls_policy.md` — generated
- `docs/resources/tls_policy_member.md` — generated
- `docs/data-sources/tls_policy.md` — generated
- `ROADMAP.md` — TLS Policies marked Done

## Decisions Made

- `is_local` gets `UseStateForUnknown` because it is set by the API at creation and does not change — same pattern as `id`
- `policy_type` is Computed-only without `UseStateForUnknown` — it can change on the API side, drift detection still logs it
- `listToStringSlice` defined locally in `tls_policy_resource.go` (not `helpers.go`) — single consumer, no premature abstraction
- `tlsPolicyMemberResource` has no `UpgradeState` — CRD-only resource at version 0, no state migration ever needed

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

Files created:
- internal/provider/tls_policy_resource.go: FOUND
- internal/provider/tls_policy_data_source.go: FOUND
- internal/provider/tls_policy_member_resource.go: FOUND
- internal/provider/tls_policy_resource_test.go: FOUND
- internal/provider/tls_policy_data_source_test.go: FOUND
- internal/provider/tls_policy_member_resource_test.go: FOUND
- docs/resources/tls_policy.md: FOUND
- docs/data-sources/tls_policy.md: FOUND

Commits: e80e53f (Task 1), e53552b (Task 2) — both exist.

Test count: 716 (baseline was 709, added 7 new tests — meets >= 706 requirement).
Lint: 0 issues.

---
*Phase: 40-tls-policies*
*Completed: 2026-04-03*
