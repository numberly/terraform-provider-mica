---
phase: 07-s3-export-policies-and-virtual-hosts
verified: 2026-03-28T15:30:00Z
status: passed
score: 9/9 must-haves verified
re_verification: false
---

# Phase 7: S3 Export Policies and Virtual Hosts — Verification Report

**Phase Goal:** Operators can manage S3 export access policies and virtual-hosted-style S3 endpoints through Terraform
**Verified:** 2026-03-28
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

All truths are sourced from the three PLAN must_haves blocks (Plans 01, 02, 03).

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | S3 export policy client can POST/GET/PATCH/DELETE policies | VERIFIED | 11 methods in `internal/client/s3_export_policies.go` (lines 20–184) |
| 2 | S3 export policy rule client can POST/GET/PATCH/DELETE rules with actions/effect/resources | VERIFIED | `PostS3ExportPolicyRule`, `GetS3ExportPolicyRuleByIndex`, `GetS3ExportPolicyRuleByName`, `PatchS3ExportPolicyRule`, `DeleteS3ExportPolicyRule` all present |
| 3 | Virtual host client can POST/GET/PATCH/DELETE with attached_servers | VERIFIED | 5 methods in `internal/client/object_store_virtual_hosts.go` (lines 20–100) |
| 4 | Mock handlers respond correctly for all CRUD operations on both resource families | VERIFIED | `RegisterS3ExportPolicyHandlers` and `RegisterObjectStoreVirtualHostHandlers` confirmed in handlers package; 258 tests pass |
| 5 | Operator can create an S3 export policy with enabled toggle via Terraform | VERIFIED | `flashblade_s3_export_policy` resource exists with `Enabled` field; `TestUnit_S3ExportPolicy_CreateReadUpdateDelete` passes |
| 6 | Operator can create S3 export policy rules with actions/effect/resources | VERIFIED | `flashblade_s3_export_policy_rule` resource with effect/actions/resources fields; `TestUnit_S3ExportPolicyRule_CreateReadUpdateDelete` passes |
| 7 | Operator can update effect, actions, resources on an S3 rule without replacing it | VERIFIED | Effect is NOT `RequiresReplace`; `PatchS3ExportPolicyRule` called in Update path; in-place update tested |
| 8 | Operator can create a virtual host with hostname and attached servers via Terraform | VERIFIED | `flashblade_object_store_virtual_host` resource; `TestUnit_ObjectStoreVirtualHost_CreateReadUpdateDelete` passes |
| 9 | Operator can import an existing virtual host by name into Terraform state | VERIFIED | Import by server-assigned name confirmed; `TestUnit_ObjectStoreVirtualHost_Import` passes |

**Score:** 9/9 truths verified

---

### Required Artifacts

| Artifact | Provided by | Status | Line Count | Notes |
|----------|-------------|--------|------------|-------|
| `internal/client/models.go` | Phase 7 model structs | VERIFIED | Phase 7 section at lines 703–770 | S3ExportPolicy, S3ExportPolicyRule, ObjectStoreVirtualHost + Post/Patch variants all present |
| `internal/client/s3_export_policies.go` | S3 export policy/rule CRUD | VERIFIED | 6.7K, 11 functions | All 11 methods confirmed by `grep ^func` |
| `internal/client/object_store_virtual_hosts.go` | Virtual host CRUD | VERIFIED | 3.5K, 5 functions | All 5 methods confirmed |
| `internal/testmock/handlers/s3_export_policies.go` | Mock CRUD for S3 export policies | VERIFIED | 11.1K | `RegisterS3ExportPolicyHandlers` exported |
| `internal/testmock/handlers/object_store_virtual_hosts.go` | Mock CRUD for virtual hosts | VERIFIED | 5.2K | `RegisterObjectStoreVirtualHostHandlers` exported |
| `internal/provider/s3_export_policy_resource.go` | `flashblade_s3_export_policy` resource | VERIFIED | 10.4K | `NewS3ExportPolicyResource` exported, full CRUD |
| `internal/provider/s3_export_policy_rule_resource.go` | `flashblade_s3_export_policy_rule` resource | VERIFIED | 13.5K | `NewS3ExportPolicyRuleResource` exported, full CRUD |
| `internal/provider/s3_export_policy_data_source.go` | `flashblade_s3_export_policy` data source | VERIFIED | 4.3K | `NewS3ExportPolicyDataSource` exported |
| `internal/provider/object_store_virtual_host_resource.go` | `flashblade_object_store_virtual_host` resource | VERIFIED | 11.9K | `NewObjectStoreVirtualHostResource` exported, full CRUD |
| `internal/provider/object_store_virtual_host_data_source.go` | `flashblade_object_store_virtual_host` data source | VERIFIED | 5.3K | `NewObjectStoreVirtualHostDataSource` exported |
| `internal/provider/s3_export_policy_resource_test.go` | S3 export policy lifecycle tests | VERIFIED | 14.2K | 4 test functions including `TestUnit_S3ExportPolicy_CreateReadUpdateDelete` |
| `internal/provider/s3_export_policy_rule_resource_test.go` | S3 rule lifecycle tests | VERIFIED | 13.9K | 4 test functions including `TestUnit_S3ExportPolicyRule_IndependentDelete` |
| `internal/provider/object_store_virtual_host_resource_test.go` | Virtual host lifecycle tests | VERIFIED | 13.4K | 4 test functions including `TestUnit_ObjectStoreVirtualHost_EmptyServers` |

All 13 artifacts: VERIFIED (exists, substantive, wired).

---

### Key Link Verification

| From | To | Via | Status | Evidence |
|------|----|-----|--------|----------|
| `internal/client/s3_export_policies.go` | `internal/client/models.go` | `S3ExportPolicy`/`S3ExportPolicyRule` structs | WIRED | Types used throughout all client methods |
| `internal/testmock/handlers/s3_export_policies.go` | `internal/client/models.go` | `client.S3ExportPolicy` in store | WIRED | 18 references to `client.S3ExportPolicy`/`client.S3ExportPolicyRule` confirmed |
| `internal/testmock/handlers/object_store_virtual_hosts.go` | `internal/client/models.go` | `client.ObjectStoreVirtualHost` in store | WIRED | 8 references to `client.ObjectStoreVirtualHost` confirmed |
| `internal/provider/s3_export_policy_resource.go` | `internal/client/s3_export_policies.go` | `client.Post/Get/Patch/DeleteS3ExportPolicy` | WIRED | All 4 CRUD methods confirmed in resource (Create/Read/Update/Delete paths) |
| `internal/provider/s3_export_policy_rule_resource.go` | `internal/client/s3_export_policies.go` | `client.Post/Get/Patch/DeleteS3ExportPolicyRule` | WIRED | 6 distinct client calls confirmed |
| `internal/provider/object_store_virtual_host_resource.go` | `internal/client/object_store_virtual_hosts.go` | `client.Post/Get/Patch/DeleteObjectStoreVirtualHost` | WIRED | All 4 CRUD methods + extra Get in import path confirmed |
| `internal/provider/provider.go` | `internal/provider/s3_export_policy_resource.go` | `NewS3ExportPolicyResource` registration | WIRED | Confirmed at provider.go lines 292–293 |
| `internal/provider/provider.go` | `internal/provider/s3_export_policy_rule_resource.go` | `NewS3ExportPolicyRuleResource` registration | WIRED | Confirmed at provider.go line 293 |
| `internal/provider/provider.go` | `internal/provider/object_store_virtual_host_resource.go` | `NewObjectStoreVirtualHostResource` registration | WIRED | Confirmed at provider.go line 291 |
| `internal/provider/provider.go` | `internal/provider/s3_export_policy_data_source.go` | `NewS3ExportPolicyDataSource` registration | WIRED | Confirmed at provider.go line 318 |
| `internal/provider/provider.go` | `internal/provider/object_store_virtual_host_data_source.go` | `NewObjectStoreVirtualHostDataSource` registration | WIRED | Confirmed at provider.go line 317 |

All 11 key links: WIRED.

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| S3P-01 | Plans 01, 02 | Operator can create an S3 export policy with enable/disable toggle | SATISFIED | `flashblade_s3_export_policy` resource with `enabled` field; `TestUnit_S3ExportPolicy_CreateReadUpdateDelete` exercises toggle from true to false |
| S3P-02 | Plans 01, 02 | Operator can create S3 export policy rules with actions/effect/resources (IAM-style) | SATISFIED | `flashblade_s3_export_policy_rule` resource with effect/actions/resources; test creates rule with effect="allow", actions=["s3:GetObject"], resources=["*"] |
| S3P-03 | Plans 01, 02 | Operator can update and delete S3 export policy rules independently | SATISFIED | Effect is patchable in-place (no RequiresReplace); `TestUnit_S3ExportPolicyRule_IndependentDelete` verifies one rule can be deleted without affecting others |
| S3P-04 | Plan 02 | Operator can import S3 export policies and rules into Terraform state | SATISFIED | `TestUnit_S3ExportPolicy_Import` (by name) and `TestUnit_S3ExportPolicyRule_Import` (by `policy_name/rule_index` composite ID) both pass |
| VH-01 | Plans 01, 03 | Operator can create a virtual host with hostname and attached servers | SATISFIED | `flashblade_object_store_virtual_host` resource; `TestUnit_ObjectStoreVirtualHost_CreateReadUpdateDelete` creates VH with hostname and attached_servers |
| VH-02 | Plans 01, 03 | Operator can update attached servers list on a virtual host | SATISFIED | Full-replace PATCH semantics on `attached_servers`; update step in CRUD test adds a second server |
| VH-03 | Plan 03 | Operator can import an existing virtual host into Terraform state | SATISFIED | `TestUnit_ObjectStoreVirtualHost_Import` imports by server-assigned name with no drift on subsequent plan |

All 7 requirements: SATISFIED. No orphaned requirements.

**Requirement-to-plan traceability:**
- S3P-01 claimed by both Plan 01 (client layer) and Plan 02 (Terraform resource) — correct, spans foundation + resource layers
- S3P-02 through S3P-04 all confirmed claimed and implemented
- VH-01 and VH-02 claimed by Plans 01 and 03 — correct, spans client + resource layers
- VH-03 claimed by Plan 03 only — correct, import is a Terraform-layer concern

---

### Anti-Patterns Found

No anti-patterns detected:

- No TODO/FIXME/HACK/PLACEHOLDER comments in any Phase 7 file
- No stub return patterns (`return {}`, `Not implemented`) in resource or client files
- Two `return nil` instances in `object_store_virtual_host_resource.go` (lines 352, 359) are legitimate early returns in `modelServersToNamedRefs` helper when the list is empty or an error occurred — not stubs
- No console.log-equivalent patterns

---

### Human Verification Required

| Test | What to do | Expected | Why human |
|------|-----------|----------|-----------|
| S3 rule IAM enforcement | Create S3 export policy with effect="deny", actions=["s3:GetObject"], resources=["*"]. Attempt S3 GetObject against the FlashBlade endpoint | S3 request is denied with 403 | Requires live FlashBlade + real S3 traffic; unit tests only verify API call shape, not enforcement semantics |

This is the only item requiring human/acceptance testing. It is deferred to Phase 8 (EXP-03) per the project plan.

---

### Build and Test Results

```
go build ./internal/client/... ./internal/testmock/... ./internal/provider/...
=> Success (no compilation errors)

go test ./internal/provider/ -run "TestUnit_S3ExportPolic|TestUnit_ObjectStoreVirtualHost" -count=1 -timeout 5m
=> 12 passed

go test ./internal/... -count=1 -timeout 5m
=> 258 passed in 4 packages (no regressions)
```

---

### Summary

Phase 7 goal is fully achieved. All three plans (client layer, S3 export policy resources, virtual host resource) delivered working, tested, and wired code. The 12 new unit tests cover:

- Full CRUD lifecycle for both S3 export policies and virtual hosts
- In-place update semantics (effect on rules, hostname and servers on virtual hosts)
- Import by name (policy, virtual host) and by composite ID (policy_name/rule_index for rules)
- Independent rule deletion
- Empty attached_servers list handling without drift
- Plan modifiers (UseStateForUnknown, RequiresReplace)

All 7 requirements (S3P-01 through S3P-04, VH-01 through VH-03) are satisfied. The full test suite (258 tests) shows zero regressions against prior phases.

---

_Verified: 2026-03-28_
_Verifier: Claude (gsd-verifier)_
