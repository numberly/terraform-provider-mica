---
phase: 51-critical-pointer-schema-fixes
plan: 02
subsystem: subnet
tags: [pointer-semantics, state-upgrader, subnet, R-001, R-002, R-005]
requires:
  - doublePointerRefForPatch (shipped in 51-01)
provides:
  - SubnetPost.VLAN *int64 (explicit vlan=0 in POST body)
  - SubnetPatch.LinkAggregationGroup **NamedReference (clearable LAG)
  - Subnet schema v0->v1 upgrader (no-op identity)
affects:
  - internal/client/models_network.go
  - internal/provider/subnet_resource.go
  - internal/testmock/handlers/subnets.go
  - internal/client/subnets_test.go
  - internal/provider/subnet_resource_test.go
tech_stack:
  added: []
  patterns:
    - "Subnet migrated to the double-pointer PATCH idiom (first concrete consumer of doublePointerRefForPatch)"
    - "Defensive schema bump with no-op identity upgrader using Go struct conversion (subnetResourceModel(oldState))"
key_files:
  created: []
  modified:
    - internal/client/models_network.go
    - internal/provider/subnet_resource.go
    - internal/testmock/handlers/subnets.go
    - internal/client/subnets_test.go
    - internal/provider/subnet_resource_test.go
decisions:
  - "Used Go struct conversion (newState := subnetResourceModel(oldState)) rather than field-by-field copy in the upgrader -- staticcheck S1016 enforces this when fields are identical"
metrics:
  duration_minutes: 8
  tasks_completed: 4
  files_modified: 5
  files_created: 0
  completed: 2026-04-20
---

# Phase 51 Plan 02: Subnet pointer semantics + schema v0->v1 Summary

Subnet is now convention-compliant: vlan=0 is sent explicitly on POST, LAG can be cleared via PATCH, schema bumped to v1 with a no-op upgrader.

## Model diffs applied

- `SubnetPost.VLAN int64 -> *int64` (R-001): allows explicit `{"vlan": 0}` in POST body.
- `SubnetPatch.LinkAggregationGroup *NamedReference -> **NamedReference` (R-002): three-state semantics (omit / clear / set) for clearable LAG ref.
- `Subnet` (GET) struct left unchanged per the plan (R-002 is PATCH-only).

## Schema version bump + upgrader

- `subnetResource.Schema.Version: 0 -> 1`.
- Added `resource.ResourceWithUpgradeState` interface assertion.
- Added `subnetV0Model` struct (identical attribute shape to current model; defensive bump per D-51-04/D-51-06).
- `UpgradeState` returns a single v0->v1 identity upgrader using `subnetResourceModel(oldState)` Go struct conversion.

## Resource code changes

- `Create()`: when `data.VLAN` is not null/unknown, sets `body.VLAN = &v` so vlan=0 is preserved.
- `Update()`: dropped the `if !plan.LagName.Equal(state.LagName)` guard and delegated to `doublePointerRefForPatch(state.LagName, plan.LagName)` (helper already handles the equality case by returning nil).
- `lagNameToRef` / `refToLagName` retained: still used by Create() and mapSubnetToModel respectively.

## Mock handler changes

- `handlers/subnets.go` handlePost: dereference `body.VLAN` only when non-nil before assigning to `subnet.VLAN`. PATCH handler unchanged (uses raw-map decoding which is agnostic to the client-side pointer shape).

## New tests (+3)

| Name                                  | Purpose                                                                                |
| ------------------------------------- | -------------------------------------------------------------------------------------- |
| TestUnit_Subnet_StateUpgrade_V0toV1   | Exercises the v0->v1 identity upgrader, verifies all attributes round-trip unchanged.  |
| TestUnit_Subnet_Create_VLANZero       | Captures POST body via httptest wrapper, asserts `"vlan":0` present (R-001 regression). |
| TestUnit_Subnet_Patch_ClearLag        | Marshals `SubnetPatch` with double-pointer, asserts `"link_aggregation_group":null` (R-002 regression). |

Helpers added alongside tests: `subnetBodyCaptor` (io-based POST/PATCH body interceptor) and `newCaptorClient` (builds a FlashBladeClient pointed at a captor-wrapped mock mux).

## Verification

- `make lint`: 0 issues.
- `make test`: all 4 packages OK, 766 tests (baseline 752, +3 this plan, previously 763).
- `make docs`: regenerated, no diff.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated existing client subnet test for new `*int64` VLAN type**
- **Found during:** Task 3 verification (`make test` after mock handler update).
- **Issue:** `internal/client/subnets_test.go` had `body.VLAN` assigned to `sub.VLAN` (int64) and `VLAN: 999` in a struct literal -- both broke after R-001 model change.
- **Fix:** Dereference `body.VLAN` when non-nil before assigning; use `vlan999 := int64(999); VLAN: &vlan999` in the call site.
- **Files modified:** internal/client/subnets_test.go
- **Commit:** d2dd828 (same atomic commit as the plan work)

**2. [Rule 1 - Lint] Replaced field-by-field struct copy with Go struct conversion in the upgrader**
- **Found during:** Task 2 post-implementation lint check.
- **Issue:** staticcheck S1016 flagged the identity copy `subnetResourceModel{ID: oldState.ID, ...}` as unnecessary when both structs have identical field sets.
- **Fix:** Collapsed to `newState := subnetResourceModel(oldState)` -- a direct Go struct conversion.
- **Files modified:** internal/provider/subnet_resource.go
- **Commit:** d2dd828

### Notes / Surprises

- The prompt mentioned "baseline test count 820" but the authoritative Makefile TEST_BASELINE is 752 and the real top-level test count before this plan was 763. After this plan: 766. Within tolerance and above the Makefile floor; reporting the discrepancy for traceability.

## Self-Check: PASSED

- internal/client/models_network.go: FOUND
- internal/provider/subnet_resource.go: FOUND (Version: 1, UpgradeState, doublePointerRefForPatch call, body.VLAN = &v)
- internal/testmock/handlers/subnets.go: FOUND (body.VLAN nil-check)
- internal/client/subnets_test.go: FOUND (updated)
- internal/provider/subnet_resource_test.go: FOUND (+3 tests)
- Commit d2dd828: FOUND
