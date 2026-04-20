---
phase: 52-important-conformance
plan: 01
subsystem: qos-policies
tags: [wire-format, pointer-semantics, state-upgrader, convention-compliance]
dependency-graph:
  requires: []
  provides:
    - "QosPolicyPost now POST-encodes 0-valued limits (semantic unlimited)"
    - "qosPolicyResource implements ResourceWithUpgradeState (4/4 interfaces)"
    - "schema v0 -> v1 identity upgrader for QoS policy"
  affects:
    - internal/client/models_storage.go
    - internal/provider/qos_policy_resource.go
    - internal/testmock/handlers/qos_policies.go
tech-stack:
  added: []
  patterns:
    - "scalar pointer migration (int64 -> *int64 where zero has meaning)"
    - "identity state upgrader via Go type conversion (newModel(oldModel))"
    - "name-via-query-param wire convention (json:\"-\")"
key-files:
  created:
    - .planning/phases/52-important-conformance/52-01-SUMMARY.md
  modified:
    - internal/client/models_storage.go
    - internal/client/qos_policies_test.go
    - internal/provider/qos_policy_resource.go
    - internal/provider/qos_policy_resource_test.go
    - internal/testmock/handlers/qos_policies.go
decisions:
  - id: D-52-01
    summary: "QosPolicyPost.MaxTotal* -> *int64, Name -> json:\"-\", schema v0->v1"
  - id: LOCAL-01
    summary: "PriorSchema copies full v0 attribute structure (PlanModifiers + Default) verbatim from current Schema(); identity upgrader uses Go type conversion since attribute shape is unchanged"
metrics:
  duration-min: 5
  completed: "2026-04-20T08:44:51Z"
  tasks: 2
  files-modified: 5
  tests-delta: +2
  tests-total-after: 772
requirements: [R-006]
---

# Phase 52 Plan 01: QosPolicyPost scalar pointers + schema v0->v1 Summary

Preserve 0 = unlimited semantics end-to-end through POST by migrating
`QosPolicyPost.MaxTotalBytesPerSec` / `MaxTotalOpsPerSec` to `*int64`,
drop the `name` JSON field (convention requires name via `?names=`), add
the missing `ResourceWithUpgradeState` interface, and bump the QoS policy
resource schema from v0 to v1 with a no-op identity upgrader.

## Scope

| # | Task | Commit |
|---|---|---|
| 1 | Migrate `QosPolicyPost` to `*int64` scalars + `json:"-"` on Name; update Create builder, mock handler decoder, and affected client test | 4b83e27 |
| 2 | Add `ResourceWithUpgradeState` assertion, bump schema to v1, define `qosPolicyV0Model` + identity `UpgradeState`, add `TestUnit_QosPolicyResource_StateUpgrade_V0toV1` | 4b83e27 |

Both tasks ship as one atomic commit per D-52-06 (one commit per
requirement; R-006 spans both wire-format fix and schema bump).

## Changes by File

- `internal/client/models_storage.go`: `QosPolicyPost.Name` -> `json:"-"`;
  `MaxTotalBytesPerSec` / `MaxTotalOpsPerSec` -> `*int64` with `omitempty`.
- `internal/provider/qos_policy_resource.go`: add
  `var _ resource.ResourceWithUpgradeState` assertion; `Version: 0` -> `1`;
  Create body builder uses `&v` pattern for `MaxTotal*`; new
  `qosPolicyV0Model` struct + `UpgradeState` identity method with full
  PriorSchema copy including `PlanModifiers` (`UseStateForUnknown`,
  `RequiresReplace`) and `booldefault.StaticBool(true)` default.
- `internal/testmock/handlers/qos_policies.go`: POST handler now reads
  name via `RequireQueryParam(... "names")` (not body) and dereferences
  `*int64` fields via new local `derefInt64` helper.
- `internal/client/qos_policies_test.go`: add
  `TestUnit_QosPolicyPost_JSONEncoding` (wire-format contract: no `name`
  key, zero pointer preserved, nil pointer omitted); adjust existing
  `TestUnit_QosPolicy_Post` to supply `*int64` pointers.
- `internal/provider/qos_policy_resource_test.go`: add
  `TestUnit_QosPolicyResource_StateUpgrade_V0toV1` — seeds v0 state with
  `max_total_bytes_per_sec=0` and asserts the unlimited value lands
  unchanged in v1 state.

## Verification

- `go test ./internal/client/... ./internal/provider/... -run 'QosPolicy' -count=1`: 23 pass.
- `make test`: all packages pass; count 772 (+2 vs prior 770). Makefile
  `TEST_BASELINE=752` still satisfied.
- `make lint`: 0 issues.
- `rg '"name"' internal/client/models_storage.go` -> no `json:"name"` tag
  on `QosPolicyPost` (only on the `QosPolicy` GET struct, as expected).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Update `TestUnit_QosPolicy_Post` to build `*int64` args**

- **Found during:** Task 1 (compile break)
- **Issue:** Existing test constructed `QosPolicyPost{MaxTotalBytesPerSec: 536870912, MaxTotalOpsPerSec: 5000}` with bare `int64` literals.
- **Fix:** Introduce local `maxBytes` / `maxOps` variables and pass `&maxBytes` / `&maxOps`. Mock handler server-side dereferences before building the `QosPolicy` response.
- **Files modified:** `internal/client/qos_policies_test.go`
- **Commit:** 4b83e27

**2. [Rule 2 - Correctness] Mock handler must not rely on `body.Name`**

- **Found during:** Task 1
- **Issue:** Plan §action assumed only dereferencing the `*int64` fields was needed, but `body.Name` is now `json:"-"` and always `""` after decode. Previous handler rejected with 400 `"name is required"` and used `body.Name` as map key, which would break every provider test.
- **Fix:** Mirror the real API — read name from `?names=` via `RequireQueryParam`. Preserves existing test surface with no provider changes needed.
- **Files modified:** `internal/testmock/handlers/qos_policies.go`
- **Commit:** 4b83e27

**3. [Rule 2 - Correctness] PriorSchema must include PlanModifiers + Default**

- **Found during:** Task 2
- **Issue:** Plan skeleton showed a minimal PriorSchema (e.g. `"id": schema.StringAttribute{Computed: true}`) without `UseStateForUnknown`, `RequiresReplace`, or the `enabled` boolean default. CONVENTIONS.md §State Upgraders requires PriorSchema to be an exact copy of the schema at that version.
- **Fix:** Copy `PlanModifiers` (on `id`, `name`, `policy_type`) and `booldefault.StaticBool(true)` (on `enabled`) verbatim from the current `Schema()` since these were already present at v0.
- **Files modified:** `internal/provider/qos_policy_resource.go`
- **Commit:** 4b83e27

### Auth Gates

None.

### Baseline discrepancy (informational)

The orchestrator prompt states "Baseline test count: 827 (post Phase 51)".
Actual pre-change count was 770 (Makefile `TEST_BASELINE=752`). No code
adjustment made — the "must not decrease" contract is satisfied
(770 -> 772). Likely a mis-reported number in the prompt; recorded here
for Phase 52 continuation plans.

## Known Stubs

None.

## Self-Check: PASSED

- Commit 4b83e27 present: `git log --oneline | grep 4b83e27` -> found.
- `internal/client/models_storage.go` — `QosPolicyPost` has `*int64` on both MaxTotal fields and `json:"-"` on Name: verified via Read.
- `internal/provider/qos_policy_resource.go` — `ResourceWithUpgradeState` assertion + `Schema.Version: 1` + `UpgradeState` method present: verified.
- `internal/provider/qos_policy_resource_test.go` — `TestUnit_QosPolicyResource_StateUpgrade_V0toV1` exists and passes: verified via `go test -run`.
- `internal/testmock/handlers/qos_policies.go` — POST uses `RequireQueryParam("names")` + `derefInt64`: verified.
- `make test`: 772 passing. `make lint`: 0 issues.
