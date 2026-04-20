---
phase: 52-important-conformance
plan: 04
subsystem: nfs-export-policy-rules
tags: [wire-format, pointer-semantics, slice-patch, state-upgrader, convention-compliance]
dependency-graph:
  requires: []
  provides:
    - "NfsExportPolicyRulePatch.Security is *[]string (distinguish omit vs clear)"
    - "schema v0 -> v1 identity upgrader for nfs_export_policy_rule"
    - "terraform apply can now clear the security flavors list on a rule"
  affects:
    - internal/client/models_policies.go
    - internal/provider/nfs_export_policy_rule_resource.go
tech-stack:
  added: []
  patterns:
    - "slice pointer migration ([]string -> *[]string where empty-slice has meaning)"
    - "identity state upgrader via Go type conversion (newModel(oldModel))"
    - "defensive nil-to-empty slice coercion before pointer assignment"
key-files:
  created:
    - .planning/phases/52-important-conformance/52-04-SUMMARY.md
  modified:
    - internal/client/models_policies.go
    - internal/provider/nfs_export_policy_rule_resource.go
    - internal/provider/nfs_export_policy_rule_resource_test.go
decisions:
  - id: D-52-04
    summary: "NfsExportPolicyRulePatch.Security -> *[]string, schema v0->v1 + clear test"
  - id: LOCAL-04
    summary: "Test uses body captor + client-driven seed (no mock SeedPolicy/SeedRule helpers needed)"
metrics:
  duration-min: 6
  completed: "2026-04-20T00:00:00Z"
  tasks: 2
  files-modified: 3
  tests-delta: +2
  tests-total-after: 832
requirements: [R-009]
---

# Phase 52 Plan 04: NfsExportPolicyRulePatch.Security *[]string + schema v0->v1 Summary

Restore clear-to-empty semantics for the NFS export policy rule security
flavors list by migrating `NfsExportPolicyRulePatch.Security` from
`[]string + omitempty` (which silently dropped empty slices) to `*[]string`
per CONVENTIONS.md §PATCH struct. Bump the resource schema v0 -> v1 with
an identity upgrader (mandated whenever a client model type changes, even
if the Terraform model shape is unchanged).

## Scope

| # | Task | Commit |
|---|---|---|
| 1 | Migrate `Security` to `*[]string`; update Update() builder with `&security` + nil-to-empty coercion; add `TestUnit_NfsExportPolicyRule_Patch_Security_Clear` with a body captor | 6804f59 |
| 2 | Bump `nfsExportPolicyRuleResource` schema to v1; add `nfsExportPolicyRuleV0Model` + identity `UpgradeState` with full PriorSchema copy; add `TestUnit_NfsExportPolicyRuleResource_StateUpgrade_V0toV1` | 6804f59 |

Single atomic commit per D-52-06 (one commit per requirement).

## Changes by File

- `internal/client/models_policies.go`: `NfsExportPolicyRulePatch.Security`
  type changed from `[]string` to `*[]string` (tag unchanged). The Get,
  Post, and RuleInPolicy structs keep plain `[]string` — only PATCH needs
  the pointer to distinguish omit vs clear.
- `internal/provider/nfs_export_policy_rule_resource.go`:
  - `Schema.Version: 0` -> `1`.
  - `Update()` builder now assigns `patch.Security = &security`, coercing
    a nil slice to `[]string{}` defensively so the body is always `[]`
    never `null` when the attribute is non-null in plan.
  - New `nfsExportPolicyRuleV0Model` struct (structurally identical to
    `nfsExportPolicyRuleModel`) + non-empty `UpgradeState` method with
    key `0` containing the full v0 PriorSchema attribute copy and an
    identity upgrader using Go type conversion
    `nfsExportPolicyRuleModel(old)`.
- `internal/provider/nfs_export_policy_rule_resource_test.go`:
  - Imports bumped for `bytes`, `encoding/json`, `io`, `net/http`,
    `net/http/httptest`.
  - New `nfsRuleBodyCaptor` + `newNFSRuleCaptorClient` helpers (mirrors
    `subnetBodyCaptor`).
  - `TestUnit_NfsExportPolicyRule_Patch_Security_Clear`: seeds policy +
    rule via the client (`Security=["sys"]`), drives Update with a plan
    containing `security=[]` (non-null empty list) and state carrying
    `security=["sys"]`. Three assertions:
    1. captured PATCH body contains `"security":[]` (not omitted, not
       null);
    2. mock-stored rule has empty `Security` after PATCH;
    3. returned state has non-null empty Security list.
  - `TestUnit_NfsExportPolicyRuleResource_StateUpgrade_V0toV1`: seeds a
    fully-populated v0 state object via `tftypes`, invokes the v0
    upgrader, and asserts every field (id, policy_name, name, index,
    access, client, permission, anonuid, secure, security length,
    required_transport_security) is preserved verbatim.

## Verification

- `go test ./internal/provider/... -run
  'TestUnit_NfsExportPolicyRule_Patch_Security_Clear|TestUnit_NfsExportPolicyRuleResource_StateUpgrade_V0toV1'
  -count=1 -v`: 2 pass.
- `make test`: all packages pass; count 775 internal (+2 vs prior 773).
  Full `go test ./...` reports 832 tests (830 baseline + 2 new).
- `make lint`: 0 issues.
- `rtk proxy rg 'Security\s+\*\[\]string' internal/client/models_policies.go`:
  matches the updated line (L91).
- Mock handler `internal/testmock/handlers/nfs_export_policies.go`
  (~L473-478) already uses raw-map JSON decode and correctly unmarshals
  `"security":[]` into an empty slice — no handler change required
  (confirmed by the Clear test passing without handler edits).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Test cannot call non-existent store Seed helpers**

- **Found during:** Task 1 (test scaffolding)
- **Issue:** The plan skeleton called `store.SeedPolicy(...)` and
  `store.SeedRule(...)`, but `nfsExportPolicyStore` exposes no public
  seed methods (only `RegisterNfsExportPolicyHandlers`).
- **Fix:** Drive seeding through the public client path
  (`createTestPolicy` + `c.PostNfsExportPolicyRule`). Equivalent
  coverage, no changes to mock handler surface, avoids a public helper
  this plan doesn't otherwise need.
- **Files modified:** `internal/provider/nfs_export_policy_rule_resource_test.go`
- **Commit:** 6804f59

**2. [Rule 1 - Lint] Staticcheck S1016 on struct literal**

- **Found during:** `make lint` after Task 2
- **Issue:** golangci-lint (staticcheck) flagged the identity upgrader
  as `S1016: should convert old (type nfsExportPolicyRuleV0Model) to
  nfsExportPolicyRuleModel instead of using struct literal` because the
  two structs are byte-identical.
- **Fix:** Replace the 16-line field-by-field copy with a Go type
  conversion `nfsExportPolicyRuleModel(old)`. Consistent with Phase 52
  Plan 01 `LOCAL-01` decision.
- **Files modified:** `internal/provider/nfs_export_policy_rule_resource.go`
- **Commit:** 6804f59

### Auth Gates

None.

## Known Stubs

None. The change is wire-format + schema-bump only; all affected paths
are fully wired and covered by unit tests.

## Self-Check: PASSED

- Commit 6804f59 present: `git log --oneline | grep 6804f59` -> found.
- `internal/client/models_policies.go` — `Security *[]string` on the
  Patch struct (L91): verified via Read.
- `internal/provider/nfs_export_policy_rule_resource.go` — `Version: 1`
  (L70), non-empty `UpgradeState` with v0 PriorSchema, Update() uses
  `patch.Security = &security`: verified via Read.
- `internal/provider/nfs_export_policy_rule_resource_test.go` —
  `TestUnit_NfsExportPolicyRule_Patch_Security_Clear` and
  `TestUnit_NfsExportPolicyRuleResource_StateUpgrade_V0toV1` both exist
  and both pass: verified via `go test -run`.
- `make test`: 775 internal tests passing (full `./...` = 832).
  `make lint`: 0 issues.
