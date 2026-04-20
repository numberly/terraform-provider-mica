---
phase: 52-important-conformance
plan: 02
subsystem: provider
tags: [conformance, interfaces, upgradestate, terraform-plugin-framework]

requires:
  - phase: 51
    provides: subnet_resource ResourceWithUpgradeState assertion (already present)
  - phase: 52-01
    provides: qos_policy_resource ResourceWithUpgradeState assertion + real v0→v1 upgrader
provides:
  - 7 additional resources now declaring all 4 mandatory framework interface assertions
  - No-op UpgradeState methods returning empty upgrader maps on those 7 resources
affects: [52-03, 52-05, convention-audit]

tech-stack:
  added: []
  patterns:
    - "No-op UpgradeState pattern for resources at SchemaVersion 0 without pending migrations"

key-files:
  created: []
  modified:
    - internal/provider/bucket_access_policy_resource.go
    - internal/provider/bucket_access_policy_rule_resource.go
    - internal/provider/bucket_audit_filter_resource.go
    - internal/provider/network_interface_resource.go
    - internal/provider/object_store_user_policy_resource.go
    - internal/provider/qos_policy_member_resource.go
    - internal/provider/tls_policy_member_resource.go

key-decisions:
  - "Assertion-only compliance: no SchemaVersion bump on any of the 7 resources (D-52-02)"
  - "Empty map returned from UpgradeState matches canonical pattern in nfs_export_policy_rule_resource.go"

patterns-established:
  - "No-op UpgradeState: resources at v0 with no migration declare the 4th interface via empty-map method"

requirements-completed: [R-007]

duration: 5min
completed: 2026-04-20
---

# Phase 52 Plan 02: ResourceWithUpgradeState assertion on 7 resources Summary

**Seven regular resources now declare the `ResourceWithUpgradeState` interface plus a no-op `UpgradeState` method — closing the CONVENTIONS.md §Resource Implementation gap for R-007 without artificial schema bumps.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-04-20T14:44:00Z
- **Completed:** 2026-04-20T14:49:30Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- All 54 regular resource files in `internal/provider/` now declare the 4 mandatory framework interface assertions (Resource, ResourceWithConfigure, ResourceWithImportState, ResourceWithUpgradeState).
- Each of the 7 targeted resources now has a no-op `UpgradeState(_ context.Context) map[int64]resource.StateUpgrader` method.
- No SchemaVersion field was touched — assertion-only compliance as locked by D-52-02.
- `make test` still green: 775 tests (baseline 752).
- `make lint`: 0 issues.

## Task Commits

1. **Task 1: Pre-flight — verify subnet + qos_policy already have assertion** — no edits (verified via grep that both files carry the assertion from Phase 51 / plan 52-01).
2. **Task 2: Add ResourceWithUpgradeState assertion + no-op UpgradeState on 7 resources** — `5e1d9b1` (fix)

## Files Created/Modified
- `internal/provider/bucket_access_policy_resource.go` — +1 assertion, +1 UpgradeState method
- `internal/provider/bucket_access_policy_rule_resource.go` — +1 assertion, +1 UpgradeState method
- `internal/provider/bucket_audit_filter_resource.go` — +1 assertion, +1 UpgradeState method
- `internal/provider/network_interface_resource.go` — +1 assertion, +1 UpgradeState method
- `internal/provider/object_store_user_policy_resource.go` — +1 assertion, +1 UpgradeState method
- `internal/provider/qos_policy_member_resource.go` — +1 assertion, +1 UpgradeState method
- `internal/provider/tls_policy_member_resource.go` — +1 assertion, +1 UpgradeState method

## Decisions Made
- Followed the plan verbatim. `subnet_resource.go` was confirmed already compliant (Phase 51); `qos_policy_resource.go` is owned by plan 52-01 and therefore skipped here to avoid double-edit.

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
- Minor: `tls_policy_member_resource.go` has two identical terminating blocks (Read + ImportState). Disambiguated the Edit by anchoring on the preceding timeouts stanza that is unique to the ImportState path.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- R-007 is closed. Plans 52-03 (done), 52-04 (done), and 52-05 (remaining) remain on track.
- No blockers.

## Self-Check: PASSED

- FOUND: assertion on all 7 target files (verified via `rg -c 'ResourceWithUpgradeState = &'`).
- FOUND: commit `5e1d9b1` in `git log`.
- FOUND: `make test` → 775 tests, `make lint` → 0 issues.

---
*Phase: 52-important-conformance*
*Completed: 2026-04-20*
