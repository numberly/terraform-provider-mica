# Requirements — Milestone v2.22.3 convention-compliance

**Source:** Convention audit report (2026-04-20) crossing CONVENTIONS.md rules with `internal/client`, `internal/provider`, `internal/testmock/handlers`, and test files. Findings cross-verified against swagger-2.22.json and git history.

**Scope:** Technical debt — no new user-facing features. Fix real violations that mask bugs, break PATCH semantics, or diverge from canonical patterns. Raise the test baseline above 818.

**Out of scope:**
- New resources or data sources (tracked in ROADMAP.md backlog)
- Refactoring legitimate exceptions (`*bool` for `Enabled` policies, nested struct pointers) — already documented in CONVENTIONS.md §Pointer rules.

---

## Active Requirements

### R-001 — SubnetPost.VLAN must use `*int64` (Critical)

**Why:** Convention §POST struct: `*int64` required when zero value is a semantic user choice. Swagger + resource schema document `VLAN=0` as "untagged". Plain `int64 + omitempty` silently drops `vlan=0` from POST body, letting the API apply its own default.

**Validation:** Integration test creating a subnet with `vlan=0` reads back `vlan=0` (not the API default).

**Primary owner:** Phase 51-critical-pointer-fixes

---

### R-002 — SubnetPatch.LinkAggregationGroup must use `**NamedReference` (Critical)

**Why:** Convention §PATCH struct: reference fields MUST be `**NamedReference` to distinguish omit/clear/set. Current `*NamedReference` makes it impossible to detach a subnet from a LAG via `terraform apply`.

**Validation:** Integration test: create subnet with LAG, then `terraform apply` with LAG unset → subnet LAG cleared on array.

**Primary owner:** Phase 51-critical-pointer-fixes

---

### R-003 — FileSystemExportPatch.Server and .SharePolicy must use `**NamedReference` (Critical)

**Why:** Same PATCH semantics violation as R-002. Users cannot clear the export policy or change the server binding without the double pointer.

**Validation:** Integration test: patch an export with `share_policy` set to null clears the policy on the array.

**Primary owner:** Phase 51-critical-pointer-fixes

---

### R-004 — ObjectStoreAccountExportPatch.Policy must use `**NamedReference` (Critical)

**Why:** Same PATCH semantics violation as R-002.

**Validation:** Integration test for patch with null policy reference.

**Primary owner:** Phase 51-critical-pointer-fixes

---

### R-005 — Schema version bumps + state upgraders for R-001..R-004 impacted resources (Critical)

**Why:** Convention §State Upgraders requires `Version` increment + upgrader chain whenever a model attribute changes type. R-001..R-004 affect `subnet_resource`, `file_system_export_resource`, `object_store_account_export_resource`. Each resource that changes must add a Vn→Vn+1 upgrader test.

**Validation:** `TestUnit_<Resource>_StateUpgrade_V*toV*` tests pass for each bumped resource. `make test` count ≥ 818 + 1 per upgrader.

**Primary owner:** Phase 51-critical-pointer-fixes

---

### R-006 — QosPolicyPost scalar pointer fixes (Important) ✅ DONE

**Why:** Convention §POST struct: `*int64` for semantic zero. `MaxTotalBytesPerSec=0` and `MaxTotalOpsPerSec=0` mean "unlimited" on FlashBlade. Current `int64 + omitempty` drops them, letting API apply default quota instead. Also: `Name` field has `json:"name"` (must be `json:"-"` per convention — name goes via `?names=` query).

**Validation:** Integration test creating a QoS policy with `max_total_bytes_per_sec=0` reads back `0`.

**Primary owner:** Phase 52-important-conformance

**Status:** Closed 2026-04-20 by 52-01 (commit 4b83e27). `TestUnit_QosPolicyPost_JSONEncoding` enforces wire-format contract; `TestUnit_QosPolicyResource_StateUpgrade_V0toV1` enforces identity upgrader.

---

### R-007 — Add `ResourceWithUpgradeState` assertion to 9 resources (Important) ✅ DONE

**Why:** Convention §Resource Implementation: all 4 interface assertions are mandatory. The following resources have only 3:
- `bucket_access_policy_resource.go`
- `bucket_access_policy_rule_resource.go`
- `bucket_audit_filter_resource.go`
- `network_interface_resource.go`
- `object_store_user_policy_resource.go`
- `qos_policy_member_resource.go`
- `qos_policy_resource.go`
- `subnet_resource.go`
- `tls_policy_member_resource.go`

**Validation:** `grep -c "ResourceWithUpgradeState" internal/provider/*_resource.go` shows assertion on every resource file. Each file has a no-op `UpgradeState` method returning `map[int64]resource.StateUpgrader{}` (when version is 0).

**Primary owner:** Phase 52-important-conformance

**Status:** Closed 2026-04-20 by 52-02 (commit 5e1d9b1) and prior phases. `subnet_resource.go` was already compliant after Phase 51. `qos_policy_resource.go` was closed by 52-01 (real v0→v1 upgrader). The remaining 7 resources got the assertion + a no-op `UpgradeState` in this plan. All 54 `*_resource.go` files now carry the 4th interface assertion; `make test`/`make lint` clean.

---

### R-008 — Document `object_store_access_key_resource.go` ImportState exception (Important)

**Why:** Convention §Resource Implementation says all 4 interfaces are mandatory, but access keys are genuinely non-importable (secret_access_key is unavailable post-creation). Either implement `ImportState` that returns an informative error, OR document the exception explicitly in CONVENTIONS.md §Resource Implementation.

**Decision:** Prefer documenting the exception (clearer user-facing contract) + implement `ImportState` that rejects with a clear error message pointing to the secret-unavailability reason.

**Validation:** CONVENTIONS.md documents the exception with rationale. Resource file declares the 4th interface and implements `ImportState` that returns `AddError("Import not supported", "...")`.

**Primary owner:** Phase 52-important-conformance

---

### R-009 — NfsExportPolicyRulePatch.Security must use `*[]string` (Important) ✅ DONE

**Why:** Convention §PATCH struct: every field is a pointer. Plain `[]string + omitempty` cannot distinguish "don't touch" from "clear to empty list".

**Validation:** `TestUnit_NfsExportPolicyRule_Patch_Security_Clear` sends empty list and asserts array state is cleared.

**Primary owner:** Phase 52-important-conformance

---

### R-010 — Test naming normalization (Important)

**Why:** Convention §Test Conventions: `TestUnit_<Resource>_<Operation>[_<Variant>]` prefix is mandatory. 6 tests deviate:
- `transport_internal_test.go`: `TestComputeDelayJitter`, `TestComputeDelayCap`
- `object_store_user_resource_test.go`: `TestMocked_ObjectStoreUser_Lifecycle`, `TestMocked_ObjectStoreUser_FullAccess`
- `object_store_user_policy_resource_test.go`: `TestMocked_ObjectStoreUserPolicy_Lifecycle`
- `object_store_user_data_source_test.go`: `TestMocked_ObjectStoreUser_DataSource`

**Validation:** `grep -rE "^func (TestMocked_|TestCompute)" internal/` returns no results.

**Primary owner:** Phase 52-important-conformance

---

### R-011 — PATCH slice fields use `*[]string` in policy rules (Cosmetic)

**Why:** Convention §PATCH struct: every field is a pointer. Applies to `ObjectStoreAccessPolicyRulePatch.Actions`/`.Resources`, `S3ExportPolicyRulePatch.Actions`/`.Resources`, `NetworkAccessPolicyRulePatch.Interfaces`, `NetworkInterfacePatch.Services`/`.AttachedServers`.

**Decision:** For `NetworkInterfacePatch`, document the "always send" pattern in CONVENTIONS.md §Pointer rules if intentional, OR migrate to `*[]T` with `omitempty`. Default: migrate, since clearing is API-supported.

**Validation:** PATCH tests verify that sending `*[]string{}` clears the list, and `nil` omits the field.

**Primary owner:** Phase 53-cosmetic-hygiene

---

### R-012 — Handler store convention hygiene (Cosmetic)

**Why:** Convention §Mock Handlers requires `byName` map naming and `nextID int` with `fmt.Sprintf("xxx-%d", s.nextID)` for synthetic IDs. Deviations:
- `qos_policies.go`: uses `policies` instead of `byName`
- `subnets.go`, `network_interfaces.go`: use `uuid.New().String()` instead of `nextID`

**Validation:** `grep -n "byName" internal/testmock/handlers/*.go` shows consistent naming. No `uuid.New` calls in handler files.

**Primary owner:** Phase 53-cosmetic-hygiene

---

## Validated Requirements

_(none yet — milestone started 2026-04-20)_

## Invalidated Requirements

_(none)_

## Cross-References

- **Source audit report:** conversation on 2026-04-20 (see git log after milestone commit)
- **Updated convention:** CONVENTIONS.md §Model Structs / Pointer rules (committed before R-001..R-012 fixes)
- **Baseline:** 818 tests, 0 lint issues (from v2.22.2)
- **Post-milestone target:** ≥ 825 tests (818 + upgrader tests + new PATCH-clear tests), 0 lint issues
