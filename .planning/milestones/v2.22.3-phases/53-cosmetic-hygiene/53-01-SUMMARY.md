---
phase: 53-cosmetic-hygiene
plan: 01
subsystem: client-models, provider-resources, docs
tags: [conventions, patch-pointer-discipline, policy-rules, R-011]
requires:
  - CONVENTIONS.md §Pointer rules
  - internal/client/models_network.go (NetworkInterfacePatch — kept as-is)
provides:
  - uniform *[]string PATCH slice discipline across 3 policy rule models
  - documented "always send" list-field exception in CONVENTIONS.md
affects:
  - internal/client/models_policies.go
  - internal/provider/object_store_access_policy_rule_resource.go
  - internal/provider/s3_export_policy_rule_resource.go
  - internal/provider/network_access_policy_rule_resource.go
tech-stack:
  added: []
  patterns:
    - "`*[]string` with `omitempty` for PATCH list fields: nil = omit, &[]string{} = clear, &[...] = set"
    - "Resource Update() takes `&slice` at assignment line; no body restructuring"
key-files:
  created:
    - .planning/phases/53-cosmetic-hygiene/53-01-SUMMARY.md
  modified:
    - internal/client/models_policies.go
    - internal/client/object_store_access_policies_test.go
    - internal/client/s3_export_policies_test.go
    - internal/client/network_access_policies_test.go
    - internal/provider/object_store_access_policy_rule_resource.go
    - internal/provider/s3_export_policy_rule_resource.go
    - internal/provider/network_access_policy_rule_resource.go
    - CONVENTIONS.md
decisions:
  - "Consolidated into a single `fix(53-01):` commit per plan success criteria (initial split into refactor+test commits was soft-reset)."
  - "Mock PATCH handlers already decode into `map[string]json.RawMessage`, not the Patch struct directly, so no handler changes were required."
metrics:
  duration_seconds: ~600
  completed_date: 2026-04-20
  test_count_before: 832
  test_count_after: 835
  tests_added: 3
  commits: 1
---

# Phase 53 Plan 01: PATCH slice discipline on policy rules (R-011)

## One-liner

Migrate `ObjectStoreAccessPolicyRulePatch`, `S3ExportPolicyRulePatch`, and `NetworkAccessPolicyRulePatch` list fields to `*[]string` with `omitempty`; formalize the `NetworkInterfacePatch` "always send" carve-out in `CONVENTIONS.md`.

## Files touched

| File | Change |
|------|--------|
| `internal/client/models_policies.go` | `Actions`, `Resources`, `Interfaces` → `*[]string` on 3 PATCH structs |
| `internal/provider/object_store_access_policy_rule_resource.go` | `patch.Actions = &actions`; `patch.Resources = &resources` |
| `internal/provider/s3_export_policy_rule_resource.go` | `patch.Actions = &actions`; `patch.Resources = &resources` |
| `internal/provider/network_access_policy_rule_resource.go` | `patch.Interfaces = &interfaces` |
| `internal/client/object_store_access_policies_test.go` | Updated existing Patch test; added `TestUnit_ObjectStoreAccessPolicyRule_Patch_ClearList` |
| `internal/client/s3_export_policies_test.go` | Added `TestUnit_S3ExportPolicyRule_Patch_ClearList` (incl. nil-omit sub-assertion) |
| `internal/client/network_access_policies_test.go` | Added `TestUnit_NetworkAccessPolicyRule_Patch_ClearList` |
| `CONVENTIONS.md` | Added `*[]T` bullet under §Pointer rules and new "Exception — 'always send' list fields" subsection |

## Test deltas

| Metric | Value |
|--------|-------|
| Before | 832 |
| After | 835 |
| Added | +3 (ObjectStore clear, S3 clear, NetworkAccess clear) |
| Removed | 0 |

## Verification

- `make test` → 835 tests, all packages `ok`.
- `make lint` → `0 issues.`
- `rg -n 'Actions\s+\[\]string' internal/client/models_policies.go` → only matches on POST structs (correct).
- `rg -n 'patch\.(Actions|Resources|Interfaces)\s*=\s*[a-z]'` → 0 matches in the 3 patched resources.
- `rg -n 'always send' CONVENTIONS.md` → 1 match under §Pointer rules.

## Commit

- `97a603e fix(53-01): PATCH slice fields use *[]string in 3 policy rule models`

No `Co-Authored-By` trailer. `--no-verify` used per project policy.

## Mock handler fixes

None required. All three PATCH mock handlers (`object_store_access_policies.go`, `s3_export_policies.go`, `network_access_policies.go`) already decode the body into `map[string]json.RawMessage` and apply per-key unmarshals, so the Patch struct type change had zero impact on them.

## Deviations from Plan

### Process deviations

**1. [Process] Two intermediate commits were consolidated into one.**
- **Found during:** Task 3 final commit step.
- **Issue:** I had separately committed Task 1 (`refactor(53-01): ...`) and Task 2 (`test(53-01): ...`) while executing atomically. The plan's success criteria explicitly require "Single commit on current branch" with `CONVENTIONS.md` in the same commit as code.
- **Fix:** `git reset --soft HEAD~2`, then single consolidated commit `fix(53-01): ...` covering all 8 files.
- **Files modified:** None additional — just repackaged the same diff.
- **Commit:** `97a603e` (final single commit).

### Code deviations

None — the migration followed the plan exactly. No handler changes, no additional fields touched, no structural modifications.

## Known Stubs

None.

## Self-Check: PASSED

- models_policies.go: `*[]string` confirmed on lines 316, 361, 447 via Read.
- CONVENTIONS.md: "always send" paragraph present.
- Commit `97a603e` exists on current branch.
- `make test` 835 passed. `make lint` clean.
