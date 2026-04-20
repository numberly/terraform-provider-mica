---
phase: 51-critical-pointer-schema-fixes
plan: 04
subsystem: object_store_account_export
tags: [patch-semantics, state-upgrader, pointer-rules, R-004, R-005]
requires:
  - doublePointerRefForPatch helper (shipped in 51-01)
provides:
  - ObjectStoreAccountExportPatch with **NamedReference Policy field
  - object_store_account_export resource schema Version 1 with no-op v0→v1 upgrader
affects:
  - internal/client/models_exports.go
  - internal/provider/object_store_account_export_resource.go
  - internal/testmock/handlers/object_store_account_exports.go
  - internal/provider/object_store_account_export_resource_test.go
tech-stack:
  added: []
  patterns: [double-pointer-patch, no-op-identity-upgrader, null-aware-mock-decoder]
key-files:
  modified:
    - internal/client/models_exports.go
    - internal/provider/object_store_account_export_resource.go
    - internal/provider/object_store_account_export_resource_test.go
    - internal/testmock/handlers/object_store_account_exports.go
decisions:
  - Convert Policy to **NamedReference for struct-level consistency even though policy_name is RequiresReplace today
  - Apply the 51-03 null-aware mock decoder pattern to the policy PATCH branch, treating JSON `null` as CLEAR
  - Use type conversion `objectStoreAccountExportModel(oldState)` in the v0→v1 upgrader (lint S1016)
metrics:
  duration: "~6 min"
  completed: "2026-04-20"
---

# Phase 51 Plan 04: object_store_account_export PATCH pointer fix + schema v0→v1 Summary

One-liner: ObjectStoreAccountExportPatch.Policy becomes `**NamedReference` so future relaxations of `policy_name` RequiresReplace can expose CLEAR semantics; schema bumped 0→1 with a no-op identity upgrader.

## Model diffs

`internal/client/models_exports.go`:

```diff
 type ObjectStoreAccountExportPatch struct {
-    ExportEnabled *bool           `json:"export_enabled,omitempty"`
-    Policy        *NamedReference `json:"policy,omitempty"`
+    ExportEnabled *bool            `json:"export_enabled,omitempty"`
+    Policy        **NamedReference `json:"policy,omitempty"`
 }
```

GET struct and POST struct untouched (R-004 scope is PATCH only).

## Resource changes

- `Version: 0 → 1` in Schema (R-005, D-51-04).
- `objectStoreAccountExportV0Model` added (identical shape to current model).
- `UpgradeState` now returns a real v0→v1 upgrader with full `PriorSchema` (including timeouts) and an identity transform via type conversion `objectStoreAccountExportModel(oldState)`. Defensive bump per D-51-04: attribute shape unchanged but PATCH wire semantics changed.
- Create()'s second-stage PATCH (policy bootstrap after POST) now builds a double-pointer SET: `ref := &client.NamedReference{Name: ...}; patch.Policy = &ref`.
- Update() now calls `patch.Policy = doublePointerRefForPatch(state.PolicyName, plan.PolicyName)` — today OMIT is the only branch exercised (RequiresReplace short-circuits SET/CLEAR), but the wire format is future-proof.

## Mock handler — Case A with 51-03 null-aware fix

`internal/testmock/handlers/object_store_account_exports.go` uses raw-map decoding (Case A), so the struct shape change is transparent. Applied the 51-03 pattern to the `policy` PATCH branch: `string(v) == "null"` now sets the stored ref to `nil` (CLEAR), and other payloads decode into a fresh `NamedReference` (SET). This keeps the mock round-trip consistent with the wire-format semantics now exposed by the client struct, even though the current Terraform surface does not exercise CLEAR due to `RequiresReplace` on `policy_name`.

## New tests

- `TestUnit_ObjectStoreAccountExport_StateUpgrade_V0toV1` — builds a v0 raw state via PriorSchema, runs the upgrader, asserts all attributes preserved (name, policy_name, account_name).
- `TestUnit_ObjectStoreAccountExport_Patch_ClearPolicy` — constructs an `ObjectStoreAccountExportPatch` via `doublePointerRefForPatch(state="policy1", plan=null)` and asserts the marshaled JSON contains `"policy":null`.

Existing helpers `accountExportResourceSchema` and `buildAccountExportType` were reused (already defined in the test file) — no redeclaration.

## Test count delta

- Before (full `go test ./...`): 825 tests
- After (full `go test ./...`): 827 tests (+2)
- Internal baseline (`make test`): 768 → 770 (+2)

## Deviations from plan

- **[Lint fix] Identity upgrader rewritten as type conversion** — matches the 51-03 finding. staticcheck (S1016) flags the literal struct copy in the upgrader. Used `newState := objectStoreAccountExportModel(oldState)` instead of the explicit field-by-field copy shown in the plan. Semantically identical; lint stays green.
- **[Rule 1 / 51-03 pattern] Null-aware mock decoder applied to `policy` branch** — plan Task 3 offered "no change required" for raw-map handlers (Case A), but the 51-03 SUMMARY established that the `string(v) == "null"` branch is required for CLEAR semantics to round-trip through the mock. Applied the same fix here for consistency and to support the (currently theoretical) CLEAR path end-to-end.

## Verification

- `make lint` → 0 issues.
- `make test` → all packages pass, 770 internal tests (+2 from baseline).
- `go test ./...` → 827 total tests.
- `make docs` → no diff (attribute schema unchanged).

## Self-Check: PASSED
