# Phase 52 Context — Important Conformance

**Milestone:** v2.22.3 convention-compliance
**Goal:** All resources declare the 4 mandatory interface assertions, QoS POST encodes unlimited quotas correctly, NFS policy rule PATCH can clear security, and test naming is uniform.
**Requirements:** R-006, R-007, R-008, R-009, R-010

---

## Canonical Refs

- `CONVENTIONS.md` §Resource Implementation (interface assertions)
- `CONVENTIONS.md` §Pointer rules (post-R-005 clarifications)
- `CONVENTIONS.md` §Test Conventions (mandatory `TestUnit_` prefix)
- `internal/client/models_storage.go` — `QosPolicyPost` (lines ~356–365)
- `internal/client/models_policies.go` — `NfsExportPolicyRulePatch.Security`
- `internal/provider/object_store_access_key_resource.go` — documented exception
- Resources missing `ResourceWithUpgradeState`:
  - `bucket_access_policy_resource.go`
  - `bucket_access_policy_rule_resource.go`
  - `bucket_audit_filter_resource.go`
  - `network_interface_resource.go`
  - `object_store_user_policy_resource.go`
  - `qos_policy_member_resource.go`
  - `qos_policy_resource.go`
  - `subnet_resource.go` (may already be covered by Phase 51 — verify)
  - `tls_policy_member_resource.go`

---

## Prior Decisions (inherited)

- **Phase 51** added `doublePointerRefForPatch` helper in `internal/provider/helpers.go` — available for any future PATCH ref migrations (not needed here — Phase 52 focuses on scalars and slices).
- **Phase 51** established null-aware PATCH mock handler decoder pattern — NOT needed here since no `**NamedReference` migration in Phase 52.
- **Test baseline** after Phase 51: 827.

---

## Gray Areas — Decisions

### GA-1 — QosPolicyPost scalar pointer migration

**Decision:** Migrate `MaxTotalBytesPerSec` and `MaxTotalOpsPerSec` from `int64` to `*int64`. `Name` changes from `json:"name"` to `json:"-"` (convention: name goes only via `?names=` query param). Resource `Create()` builds the POST body with `ptr(plan.MaxTotalBytesPerSec.ValueInt64())` (and similar for ops) when the attribute is non-null; nil when null.

**Schema bump:** Increment QoS policy `Version` to 1, add no-op v0→v1 identity upgrader (consistent with Phase 51 D-51-04 rationale — convention-driven).

**Rationale:** `0=unlimited` semantics require the same fix pattern as VLAN=0.

### GA-2 — `ResourceWithUpgradeState` assertion on 9 resources

**Decision:** For each of the 9 resources, add:
1. `var _ resource.ResourceWithUpgradeState = &xxxResource{}` assertion at top of file.
2. A no-op `UpgradeState` method returning `map[int64]resource.StateUpgrader{}`.

Do **not** bump `SchemaVersion` — these resources are at v0 and have no pending migration; the assertion alone satisfies convention.

**Verification:** `subnet_resource.go` will likely already have this after Phase 51 — verify and skip if present.

### GA-3 — `object_store_access_key_resource.go` ImportState exception

**Decision:** Document the exception in CONVENTIONS.md §Resource Implementation AND add a shim `ImportState` that calls `resp.Diagnostics.AddError("Import not supported", "object store access keys cannot be imported because secret_access_key is only available at creation time. Recreate the resource via terraform apply.")`. Also add the 4th interface assertion.

**Rationale:** Resource becomes convention-compliant at the assertion level while the runtime behavior preserves the security semantics (user can't import a key without its secret).

### GA-4 — NfsExportPolicyRulePatch.Security slice pointer

**Decision:** Change `Security []string \`json:"security,omitempty"\`` → `Security *[]string \`json:"security,omitempty"\``. Resource `Update()` builds `&values` when the attribute is in plan, `nil` otherwise. Mock handler: adapt decode to accept pointer slice. Add `TestUnit_NfsExportPolicyRule_Patch_Security_Clear` that sends `[]string{}` and asserts cleared state.

**Schema bump:** NfsExportPolicyRule resource gets `Version: 1` + no-op upgrader.

### GA-5 — Test renames (6 tests)

**Decision:** Pure rename — no behavior change, no file relocation:
- `internal/client/transport_internal_test.go`: `TestComputeDelayJitter` → `TestUnit_Transport_ComputeDelayJitter`; `TestComputeDelayCap` → `TestUnit_Transport_ComputeDelayCap`
- `internal/provider/object_store_user_resource_test.go`: `TestMocked_ObjectStoreUser_Lifecycle` → `TestUnit_ObjectStoreUserResource_Lifecycle`; `TestMocked_ObjectStoreUser_FullAccess` → `TestUnit_ObjectStoreUserResource_FullAccess`
- `internal/provider/object_store_user_policy_resource_test.go`: `TestMocked_ObjectStoreUserPolicy_Lifecycle` → `TestUnit_ObjectStoreUserPolicyResource_Lifecycle`
- `internal/provider/object_store_user_data_source_test.go`: `TestMocked_ObjectStoreUser_DataSource` → `TestUnit_ObjectStoreUserDataSource_Basic`

**Verification:** `grep -rE "^func (TestMocked_|TestCompute)" internal/` returns 0 results.

### GA-6 — Commit strategy

**Decision:** 5 atomic commits (one per requirement):
- `fix(52-01): QosPolicyPost scalar pointers + schema v0→v1` (R-006)
- `fix(52-02): add ResourceWithUpgradeState assertion on 9 resources` (R-007)
- `fix(52-03): implement ImportState shim on object_store_access_key and document exception` (R-008)
- `fix(52-04): NfsExportPolicyRulePatch.Security *[]string + schema v0→v1` (R-009)
- `refactor(52-05): rename tests to TestUnit_ prefix` (R-010)

---

## Deferred Ideas

- Generic `doublePointerSliceForPatch` helper — wait until Phase 53 establishes more slice-PATCH call sites before extracting.

---

## Decisions Locked

| ID | Decision | Applies to |
|---|---|---|
| D-52-01 | `QosPolicyPost.MaxTotal*` → `*int64`, `Name` → `json:"-"`, schema v0→v1 | R-006 |
| D-52-02 | 9 resources get `ResourceWithUpgradeState` assertion + no-op method; no SchemaVersion bump | R-007 |
| D-52-03 | `object_store_access_key` gets 4th interface + shim ImportState rejecting with clear error; CONVENTIONS.md documents exception | R-008 |
| D-52-04 | `NfsExportPolicyRulePatch.Security` → `*[]string`, schema v0→v1 + clear test | R-009 |
| D-52-05 | Pure test renames, 6 functions across 4 files | R-010 |
| D-52-06 | 5 atomic commits, one per requirement | phase-wide |

---

## Scope for planner

- 5 plans (one per requirement)
- Net test delta: +3 (QoS state upgrader + NfsExportPolicyRule state upgrader + Security clear test). Target: 827 → ≥ 830.
- Lint must stay clean.
