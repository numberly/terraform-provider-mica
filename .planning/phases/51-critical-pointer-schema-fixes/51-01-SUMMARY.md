---
phase: 51-critical-pointer-schema-fixes
plan: 01
subsystem: provider-helpers
tags: [helpers, patch-semantics, convention]
requires: []
provides:
  - doublePointerRefForPatch helper
affects:
  - internal/provider/helpers.go
  - internal/provider/helpers_test.go
tech-stack:
  added: []
  patterns: ["three-state PATCH encoding (OMIT/CLEAR/SET) for **NamedReference"]
key-files:
  created:
    - none (helpers_test.go already existed; tests appended)
  modified:
    - internal/provider/helpers.go
    - internal/provider/helpers_test.go
decisions:
  - D-51-01 helper in internal/provider/helpers.go
metrics:
  duration: ~3min
  completed: 2026-04-20
  tests_delta: +5 (815 -> 820)
commit: 8773eb0
---

# Phase 51 Plan 01: doublePointerRefForPatch Helper Summary

One-liner: Added `doublePointerRefForPatch(state, plan types.String) **client.NamedReference` helper in `internal/provider/helpers.go` encoding the three-state PATCH semantics (OMIT / CLEAR / SET) required by downstream plans 51-02/03/04.

## Helper

- Signature: `func doublePointerRefForPatch(state, plan types.String) **client.NamedReference`
- Location: `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/helpers.go` (inserted before the `// ---------- shared helpers` block)
- Behavior:
  - `state.Equal(plan)` -> returns nil (OMIT)
  - `plan.IsNull()` (and state != plan) -> `&(*NamedReference)(nil)` (CLEAR)
  - otherwise -> `&(&NamedReference{Name: plan.ValueString()})` (SET)

## Tests Added (5)

All at `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/helpers_test.go`:

1. `TestUnit_Helpers_DoublePointerRefForPatch_Omit_SameValue`
2. `TestUnit_Helpers_DoublePointerRefForPatch_Omit_BothNull`
3. `TestUnit_Helpers_DoublePointerRefForPatch_Clear`
4. `TestUnit_Helpers_DoublePointerRefForPatch_Set_FromNull`
5. `TestUnit_Helpers_DoublePointerRefForPatch_Set_Changed`

All pass. `client` import added to test file to reference `client.NamedReference`.

## Verification

- `make lint` -> 0 issues
- `make test` -> all 5 packages pass. Total count: 820 (was 815 at plan start; +5 delta as expected)
- Plan baseline note: Plan body stated baseline 818 and target 823. Actual on-disk baseline was 815 -> 820 (same +5 delta, different absolute). No scope change; documented here for traceability.

## Deviations from Plan

None — plan executed exactly as written. Only observation: plan claimed baseline 818; actual baseline at execution time was 815. +5 delta and all success criteria met.

## Commit

- SHA: `8773eb0`
- Message: `feat(51-01): add doublePointerRefForPatch helper`
- Staged files: `internal/provider/helpers.go`, `internal/provider/helpers_test.go`

## Self-Check: PASSED

- helpers.go contains `doublePointerRefForPatch` — FOUND
- helpers_test.go contains 5 `TestUnit_Helpers_DoublePointerRefForPatch_*` tests — FOUND
- Commit `8773eb0` exists — FOUND
