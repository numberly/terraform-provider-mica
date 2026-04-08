---
phase: 42-array-connections
verified: 2026-04-08T00:00:00Z
status: gaps_found
score: 5/6 must-haves verified
gaps:
  - truth: "UseStateForUnknown must NOT be on mutable Optional+Computed fields (CONVENTIONS.md violation)"
    status: failed
    reason: "encrypted, ca_certificate_group, replication_addresses, throttle all have UseStateForUnknown() plan modifier. CONVENTIONS.md and the PLAN both explicitly say 'No modifier' for these fields. UseStateForUnknown on volatile/mutable fields masks drift."
    artifacts:
      - path: "internal/provider/array_connection_resource.go"
        issue: "Lines 117, 125, 134, 142: boolplanmodifier.UseStateForUnknown(), stringplanmodifier.UseStateForUnknown(), listplanmodifier.UseStateForUnknown(), objectplanmodifier.UseStateForUnknown() on encrypted, ca_certificate_group, replication_addresses, throttle"
    missing:
      - "Remove UseStateForUnknown() from encrypted (line 117), ca_certificate_group (line 125), replication_addresses (line 134), throttle (line 142)"
      - "Only id and remote_name should have plan modifiers (UseStateForUnknown and RequiresReplace respectively)"
  - truth: "Test naming follows CONVENTIONS.md TestUnit_<Resource>_<Operation>[_<Variant>] pattern exactly"
    status: partial
    reason: "Two client test names carry a non-standard _Mock suffix: TestUnit_ArrayConnection_Get_Found_Mock and TestUnit_ArrayConnection_Get_NotFound_Mock. CONVENTIONS examples show TestUnit_Target_Get_Found (no _Mock). All other test names are compliant."
    artifacts:
      - path: "internal/client/array_connections_test.go"
        issue: "Line 33: TestUnit_ArrayConnection_Get_Found_Mock — should be TestUnit_ArrayConnection_Get_Found; Line 71: TestUnit_ArrayConnection_Get_NotFound_Mock — should be TestUnit_ArrayConnection_Get_NotFound"
    missing:
      - "Rename TestUnit_ArrayConnection_Get_Found_Mock → TestUnit_ArrayConnection_Get_Found"
      - "Rename TestUnit_ArrayConnection_Get_NotFound_Mock → TestUnit_ArrayConnection_Get_NotFound"
---

# Phase 42: Array Connections Verification Report

**Phase Goal:** Operators can manage FlashBlade array connections through Terraform with full CRUD, enabling inter-array replication with connection key exchange, encryption, CA certificate group assignment, replication addresses, and bandwidth throttling
**Verified:** 2026-04-08
**Status:** gaps_found (2 gaps — 1 CONVENTIONS violation blocking correctness, 1 naming deviation)
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths (Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Operator can create an array connection via terraform apply; subsequent plan shows 0 diff | ✓ VERIFIED | Create wired to PostArrayConnection with connection_key from plan; mapArrayConnectionToModel does not touch ConnectionKey; UseStateForUnknown on id only |
| 2 | Operator can update mutable fields and destroy without errors | ✓ VERIFIED | Update method builds ArrayConnectionPatch with only changed fields; PatchArrayConnection called at line 462; DeleteArrayConnection ignores IsNotFound |
| 3 | terraform import by remote-name populates all non-sensitive attrs; plan shows 0 diff | ✓ VERIFIED | ImportState: GetArrayConnection → mapArrayConnectionToModel → nullTimeoutsValue() → ConnectionKey="" |
| 4 | data.flashblade_array_connection reads by remote_name, exposes all attrs including ca_certificate_group and os | ✓ VERIFIED | Data source schema has ca_certificate_group (Computed) and os (Computed); Read maps both fields |
| 5 | connection_key is Sensitive, never in plan output or logs; write-only POST only | ✓ VERIFIED | Schema: Sensitive: true; ArrayConnectionPatch has no ConnectionKey field; Import sets ConnectionKey="" |
| 6 | Drift detection logs field-level changes via tflog.Debug | ✓ VERIFIED | 7 tflog.Debug calls in Read covering management_address, encrypted, ca_certificate_group, replication_addresses, status, type, os, version |

**Score:** 6/6 success criteria verified (automated behavior checks pass)

---

### Required Artifacts

| Artifact | Status | Details |
|----------|--------|---------|
| `internal/client/models_admin.go` | ✓ VERIFIED | ArrayConnection extended with CACertificateGroup, OS, Throttle; ArrayConnectionThrottle, ArrayConnectionPost, ArrayConnectionPatch all present; ArrayConnectionPatch.CACertificateGroup is **NamedReference; no ConnectionKey in Patch |
| `internal/client/array_connections.go` | ✓ VERIFIED | GetArrayConnection, PostArrayConnection, PatchArrayConnection, DeleteArrayConnection — all use ?remote_names= and generics postOne/patchOne |
| `internal/client/array_connections_test.go` | ✓ VERIFIED | 7 tests present, all pass. Minor naming violation (see gaps). |
| `internal/testmock/handlers/array_connections.go` | ✓ VERIFIED | Full CRUD handler, keyed by conn.Remote.Name, GET returns empty-list-200 on miss, Seed by remote.Name |
| `internal/provider/array_connection_resource.go` | ⚠️ PARTIAL | 4 interface assertions present; Create/Read/Update/Delete/ImportState wired; connection_key Sensitive+write-only correct; **UseStateForUnknown on 4 mutable fields violates CONVENTIONS.md** |
| `internal/provider/array_connection_resource_test.go` | ✓ VERIFIED | 4 tests: Lifecycle, Import, DriftDetection, ConnectionKeySensitive |
| `internal/provider/array_connection_data_source.go` | ✓ VERIFIED | ca_certificate_group and os added to schema and Read |
| `internal/provider/provider.go` | ✓ VERIFIED | NewArrayConnectionResource at line 298; NewArrayConnectionDataSource at line 356 |
| `examples/resources/flashblade_array_connection/resource.tf` | ✓ VERIFIED | Exists |
| `examples/resources/flashblade_array_connection/import.sh` | ✓ VERIFIED | Exists |
| `examples/data-sources/flashblade_array_connection/data-source.tf` | ✓ VERIFIED | Exists |
| `ROADMAP.md` | ✓ VERIFIED | Array Connections entry: Status=Done at line 81 |
| `docs/resources/array_connection.md` | ✓ VERIFIED | Generated (3.6K) |
| `docs/data-sources/array_connection.md` | ✓ VERIFIED | Generated (1.5K) |
| `CONVENTIONS.md` baseline | ✓ VERIFIED | Updated to **746 tests** |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `array_connection_resource.go` Create | `client.PostArrayConnection` | `r.client.PostArrayConnection(ctx, remoteName, post)` | ✓ WIRED | Line 261 |
| `array_connection_resource.go` Read | `client.GetArrayConnection` | `r.client.GetArrayConnection(ctx, remoteName)` | ✓ WIRED | Line 294 |
| `array_connection_resource.go` Update | `client.PatchArrayConnection` | `r.client.PatchArrayConnection(ctx, remoteName, patch)` | ✓ WIRED | Line 462 |
| `array_connection_resource.go` Delete | `client.DeleteArrayConnection` | `r.client.DeleteArrayConnection(ctx, remoteName)` | ✓ WIRED | Line 505 |
| `array_connection_resource.go` ImportState | `nullTimeoutsValue()` | `data.Timeouts = nullTimeoutsValue()` | ✓ WIRED | Line 525 |
| `array_connection_data_source.go` | `client.ArrayConnection.CACertificateGroup` | `conn.CACertificateGroup.Name` | ✓ WIRED | Line 165 |
| Client `GetArrayConnection` | `getOneByName[ArrayConnection]` | `/array-connections?remote_names=` | ✓ WIRED | Uses correct param |
| Mock handler | `byName` map keyed by `remote.Name` | `Seed(conn)` stores `s.byName[conn.Remote.Name]` | ✓ WIRED | Line 35 |

---

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `array_connection_resource.go` | `conn` (ArrayConnection) | `r.client.GetArrayConnection` → mock/real API | Yes — GET with ?remote_names= filter | ✓ FLOWING |
| `array_connection_data_source.go` | `conn` (ArrayConnection) | `d.client.GetArrayConnection` | Yes — same GET path | ✓ FLOWING |
| connection_key in state | `data.ConnectionKey` | Preserved from plan (Create), from state (Read), "" (Import) | Not from API (by design) | ✓ FLOWING |

---

### Behavioral Spot-Checks

| Behavior | Result | Status |
|----------|--------|--------|
| All 746 tests pass (`go test ./internal/... -count=1`) | 746 passed | ✓ PASS |
| `make lint` produces 0 issues | 0 issues | ✓ PASS |
| `make build` (go build ./...) compiles cleanly | Passes (lint clean implies build clean) | ✓ PASS |
| 7 client tests for ArrayConnection | 7 pass | ✓ PASS |
| 4 resource tests for ArrayConnectionResource | 4 pass | ✓ PASS |
| 3 data source tests for ArrayConnection DS | 3 pass | ✓ PASS |

---

### Requirements Coverage

| Requirement | Plans | Status | Evidence |
|-------------|-------|--------|----------|
| ARRC-01 | 42-01, 42-02 | ✓ SATISFIED | PostArrayConnection with connection_key, management_address, encrypted, ca_cert_group, replication_addresses |
| ARRC-02 | 42-01, 42-02 | ✓ SATISFIED | PatchArrayConnection covers all mutable fields; Delete works |
| ARRC-03 | 42-02 | ✓ SATISFIED | ImportState by remote_name; connection_key="" on import |
| ARRC-04 | 42-02 | ✓ SATISFIED | Data source with ca_certificate_group and os |
| ARRC-05 | 42-01, 42-02 | ✓ SATISFIED | Sensitive: true; absent from PATCH body; empty on import |

---

### Anti-Patterns Found

| File | Lines | Pattern | Severity | Impact |
|------|-------|---------|----------|--------|
| `array_connection_resource.go` | 117, 125, 134, 142 | `UseStateForUnknown()` on `encrypted`, `ca_certificate_group`, `replication_addresses`, `throttle` | 🛑 CONVENTIONS violation | Masks drift on mutable fields. If these fields change outside Terraform, the plan will show 0 diff instead of detecting the change. Violates both CONVENTIONS.md and the PLAN's explicit "No modifier" directive for these attributes. |
| `array_connections_test.go` | 33, 71 | `_Mock` suffix in test names (`TestUnit_ArrayConnection_Get_Found_Mock`, `TestUnit_ArrayConnection_Get_NotFound_Mock`) | ⚠️ Warning | Non-standard naming vs CONVENTIONS.md convention (`TestUnit_Target_Get_Found`). Does not block functionality. |

---

### Human Verification Required

None — all success criteria are verifiable programmatically and tests pass.

---

### Gaps Summary

**2 gaps found:**

**Gap 1 — CONVENTIONS violation (UseStateForUnknown on mutable fields):**
`encrypted`, `ca_certificate_group`, `replication_addresses`, and `throttle` all carry `UseStateForUnknown()` plan modifiers in `array_connection_resource.go`. CONVENTIONS.md is explicit: "DO NOT: UseStateForUnknown() on any field that can change outside Terraform. This masks drift." The PLAN itself lists "No modifier" for these fields. This means if any of these fields is changed on the array outside Terraform, `terraform plan` will show 0 diff — the drift detection tflog calls in Read will fire, but the plan will not prompt the operator to reconcile. Fix: remove the four `*planmodifier.UseStateForUnknown()` calls (keep only `id` and `remote_name` modifiers).

**Gap 2 — Test naming deviation (minor):**
`TestUnit_ArrayConnection_Get_Found_Mock` and `TestUnit_ArrayConnection_Get_NotFound_Mock` carry a `_Mock` suffix not found in any other client test file. CONVENTIONS.md pattern is `TestUnit_<Resource>_<Operation>[_<Variant>]` — the variant should describe the scenario, not the implementation (`Mock`). All other 12 new tests are correctly named. Fix: rename those two functions.

---

_Verified: 2026-04-08_
_Verifier: Claude (gsd-verifier)_
