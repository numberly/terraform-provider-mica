---
phase: 51-critical-pointer-schema-fixes
plan: 03
subsystem: file_system_export
tags: [patch-semantics, state-upgrader, pointer-rules, R-003, R-005]
requires:
  - doublePointerRefForPatch helper (shipped in 51-01)
provides:
  - FileSystemExportPatch with **NamedReference fields for Server and SharePolicy
  - file_system_export resource schema Version 1 with no-op v0→v1 upgrader
affects:
  - internal/client/models_exports.go
  - internal/provider/file_system_export_resource.go
  - internal/testmock/handlers/file_system_exports.go
  - internal/provider/file_system_export_resource_test.go
tech-stack:
  added: []
  patterns: [double-pointer-patch, no-op-identity-upgrader]
key-files:
  modified:
    - internal/client/models_exports.go
    - internal/provider/file_system_export_resource.go
    - internal/provider/file_system_export_resource_test.go
    - internal/testmock/handlers/file_system_exports.go
decisions:
  - Convert both Server and SharePolicy to **NamedReference for struct-level consistency even though server_name is RequiresReplace
  - Adapt mock handler raw-map decoder to treat JSON `null` as CLEAR (zero-value struct was a latent bug)
metrics:
  duration: "~8 min"
  completed: "2026-04-20"
---

# Phase 51 Plan 03: file_system_export PATCH pointer fix + schema v0→v1 Summary

One-liner: FileSystemExportPatch.Server and .SharePolicy become `**NamedReference` so users can clear share_policy_name via `terraform apply`; schema bumped 0→1 with a no-op identity upgrader.

## Model diffs

`internal/client/models_exports.go`:

```diff
 type FileSystemExportPatch struct {
     ExportName  *string          `json:"export_name,omitempty"`
-    Server      *NamedReference  `json:"server,omitempty"`
-    SharePolicy *NamedReference  `json:"share_policy,omitempty"`
+    Server      **NamedReference `json:"server,omitempty"`
+    SharePolicy **NamedReference `json:"share_policy,omitempty"`
 }
```

GET struct and POST struct untouched (R-003 scope is PATCH only).

## Resource changes

- `Version: 0 → 1` in Schema (R-005, D-51-03).
- `fileSystemExportV0Model` added (identical shape to current model).
- `UpgradeState` now returns a real v0→v1 upgrader with `PriorSchema` and an identity transform (`fileSystemExportModel(oldState)`). D-51-04 defensive bump: attribute shape unchanged but PATCH wire semantics changed.
- `Update()` now calls `patch.SharePolicy = doublePointerRefForPatch(state.SharePolicyName, plan.SharePolicyName)` — CLEAR / SET / OMIT three-state semantics exposed to Terraform users.
- `patch.Server` is never set in Update() (server_name is RequiresReplace — the field exists on the struct for convention compliance only).

## Mock handler — Case A with fix

`internal/testmock/handlers/file_system_exports.go` uses raw-map decoding, so the struct shape change is transparent. However, the existing code had a latent bug: `json.Unmarshal([]byte("null"), &ref)` leaves `ref` as a zero-value struct, which would not correctly represent a CLEAR. Adapted both `server` and `share_policy` PATCH branches to check `string(v) == "null"` and set the stored ref to `nil` in that case. This is Rule 1 (auto-fix bug) — necessary for the R-003 clear semantics to round-trip through the mock.

## New tests

- `TestUnit_FileSystemExport_StateUpgrade_V0toV1` — builds a v0 raw state via PriorSchema, runs the upgrader, asserts all attributes preserved.
- `TestUnit_FileSystemExport_Patch_ClearSharePolicy` — constructs a FileSystemExportPatch via the helper and asserts the marshaled JSON contains `"share_policy":null` when transitioning set→null.

## Test count delta

- Before: 823 total tests (full `go test ./...`)
- After: 825 total tests (+2)
- Internal baseline (`make test`): 766 → 768 (+2)

## Deviations from plan

- **[Rule 1 — Bug] Fixed mock handler JSON null decoding** — The plan Task 3 Case A said to leave the handler untouched if raw-map decoding was used. On inspection the existing code would decode `"share_policy": null` into a zero-value `NamedReference` struct rather than clearing the stored ref. This prevents a regression test of CLEAR semantics from succeeding end-to-end through the mock. Fixed both `server` and `share_policy` PATCH branches to treat JSON `null` as a clear.
- **[Lint fix] Identity upgrader rewritten as type conversion** — staticcheck (S1016) flagged the literal struct copy in the upgrader. Rewrote `newState := fileSystemExportModel{ID: oldState.ID, ...}` as `newState := fileSystemExportModel(oldState)`. Semantically identical; makes lint green.

## Verification

- `make lint` → 0 issues.
- `make test` → all packages pass, 768 tests (+2 from baseline).
- `make docs` → no diff (attribute schema unchanged).

## Self-Check: PASSED
