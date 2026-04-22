---
phase: 54-bridge-bootstrap-poc-3-resources
plan: "05"
subsystem: testing
tags: [pulumi, bridge, tfbridge, go, unit-test, providerinfo]

requires:
  - phase: 54-bridge-bootstrap-poc-3-resources
    provides: resources.go with full ProviderInfo (plans 01-04)

provides:
  - resources_test.go with 11 TestProviderInfo_* assertions locking bridge invariants
  - Test coverage for: resource counts, timeouts omission, secret promotion, composite IDs, no autonaming

affects: [phase 55, phase 56, phase 57, phase 58]

tech-stack:
  added: []
  patterns:
    - "Bridge ProviderInfo unit tests call Provider() directly (no network, no tfgen)"
    - "Count tests use exact integers — must update when TF provider resource set changes"
    - "Sensitive ID uses tfbridge.False() (not True()) — documented in test comment"

key-files:
  created:
    - pulumi/provider/resources_test.go
  modified: []

key-decisions:
  - "expectedResources=54, expectedDataSources=40 (plan template had stale 28/21 — updated from actual Provider() output)"
  - "TestProviderInfo_ConfigHasNoManualOverrides replaces api_token test (Config is empty; TF uses nested auth.* auto-promoted by shim)"
  - "BucketDeleteTimeout test adapted: DeleteTimeout does not exist on ResourceInfo v3.127.0 — verifies bucket registration only"
  - "AdditionalSecretOutputs assertion removed: field not on ResourceInfo in bridge v3.127.0"
  - "Added TestProviderInfo_ArrayConnectionKeySensitiveIDFalse to lock False() quirk"

requirements-completed: [TEST-01]

duration: 8min
completed: 2026-04-22
---

# Phase 54 Plan 05: Bridge ProviderInfo Test Suite Summary

**11 TestProviderInfo_* unit tests locking Phase 54 bridge invariants (counts, secrets, timeouts, ComputeID, no autonaming) — all pass with no network**

## Performance

- **Duration:** 8 min
- **Started:** 2026-04-22T12:00Z
- **Completed:** 2026-04-22T12:08Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments

- Created `pulumi/provider/resources_test.go` (package `provider`, calls `Provider()` directly)
- 11 sub-tests covering every Phase 54 invariant: resource/DS counts, timeouts omission across 54 resources, secret_access_key promotion, ComputeID on policy rule, array_connection_key False() quirk, no autonaming
- All 11 tests pass: `go test -run TestProviderInfo .` in `pulumi/provider/`
- TEST-01 requirement satisfied; Phase 54 bridge chain validated without real FlashBlade

## Task Commits

1. **Task 1: Author resources_test.go** - `bc00ba1` (test)

**Plan metadata:** _(docs commit pending)_

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider/resources_test.go` - 11 TestProviderInfo_* assertions

## Decisions Made

- **Count correction:** Plan template expected `expectedResources=28, expectedDataSources=21` (old roadmap values). Actual `Provider()` returns 54 resources + 40 data sources. Updated constants before writing tests.
- **Config test adaptation:** `TestProviderInfo_ApiTokenIsSecret` from plan template replaced with `TestProviderInfo_ConfigHasNoManualOverrides` — Config is empty because TF uses nested `auth.*` block with auto-promoted Sensitive fields.
- **DeleteTimeout adaptation:** Bridge v3.127.0 `ResourceInfo` has no `DeleteTimeout` field (confirmed in STATE.md + bridge source). Test verifies bucket/filesystem registration only.
- **AdditionalSecretOutputs removed:** Not on `ResourceInfo` in v3.127.0. `Secret: tfbridge.True()` field-level mark is the explicit defense; TF auto-promotion is runtime defense.
- **Added False() test:** `TestProviderInfo_ArrayConnectionKeySensitiveIDFalse` added to lock the bridge quirk where `tfbridge.False()` (not `True()`) is required to acknowledge sensitive ID exposure.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected resource/datasource counts**
- **Found during:** Task 1 (writing test constants)
- **Issue:** Plan template had `expectedResources=28, expectedDataSources=21` — stale values from original roadmap. Running `Provider()` in a temp test revealed 54 resources and 40 data sources.
- **Fix:** Set `expectedResources=54, expectedDataSources=40` with explanatory comment.
- **Verification:** `TestProviderInfo_ResourceAndDataSourceCounts` passes.
- **Committed in:** bc00ba1

**2. [Rule 1 - Bug] Config test adapted to actual resources.go**
- **Found during:** Task 1 (reading resources.go before writing tests)
- **Issue:** resources.go has empty `Config` map (no `api_token`); plan template tested `Config["api_token"]` which would `t.Fatalf`.
- **Fix:** Replaced with `TestProviderInfo_ConfigHasNoManualOverrides` asserting Config is empty.
- **Committed in:** bc00ba1

**3. [Rule 1 - Bug] Removed non-existent ResourceInfo fields from assertions**
- **Found during:** Task 1 (verifying bridge v3.127.0 ResourceInfo struct)
- **Issue:** Plan template asserted `r.DeleteTimeout >= 25*time.Minute` and `r.AdditionalSecretOutputs` — both fields don't exist on `ResourceInfo` in v3.127.0 (confirmed via bridge source + STATE.md).
- **Fix:** `BucketDeleteTimeout` → `BucketSoftDeleteRegistered`; `AdditionalSecretOutputs` assertion removed.
- **Committed in:** bc00ba1

---

**Total deviations:** 3 auto-fixed (all Rule 1 bugs — plan template had stale/incorrect API assumptions)
**Impact on plan:** All fixes necessary for correctness. Tests now accurately reflect bridge v3.127.0 API. No scope creep.

## Issues Encountered

None — all issues were pre-identified in STATE.md key decisions and handled cleanly.

## Next Phase Readiness

- Phase 54 complete: all 5 plans executed, 13/13 requirements covered across plans 01-05
- Bridge chain validated: skeleton → ProviderInfo → binaries → schema generation → test suite
- Phase 55 can proceed: filesystem + NFS export bridge resources (SOFTDELETE-02/03)

## Self-Check

- [x] `pulumi/provider/resources_test.go` exists: FOUND
- [x] commit `bc00ba1` exists: FOUND
- [x] 11 tests pass: VERIFIED (`go test -run TestProviderInfo .`)

## Self-Check: PASSED

---
*Phase: 54-bridge-bootstrap-poc-3-resources*
*Completed: 2026-04-22*
