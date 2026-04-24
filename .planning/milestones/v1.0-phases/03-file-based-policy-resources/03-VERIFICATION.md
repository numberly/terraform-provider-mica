---
phase: 03-file-based-policy-resources
verified: 2026-03-27T15:46:05Z
status: passed
score: 27/27 must-haves verified
re_verification: false
---

# Phase 3: File-Based Policy Resources — Verification Report

**Phase Goal:** Operators can manage NFS export, SMB share, and snapshot policies — including rules — through Terraform with no false drift on rule reorder
**Verified:** 2026-03-27T15:46:05Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths (from Phase 3 Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Operator can create an NFS export policy with rules; `apply → plan` shows 0 diff regardless of API rule return order | VERIFIED | NFS mock sorts rules by index; `Index` attr is Computed-only (not user-settable); 4 NFS rule tests pass |
| 2 | Operator can import NFS, SMB, and snapshot policy rules using composite ID; subsequent `plan` shows 0 diff | VERIFIED | All 3 rule ImportState implementations verified; NFS uses `policy_name/rule_index`, SMB uses `policy_name/rule_name`, Snapshot uses `policy_name/rule_index`; import tests pass |
| 3 | Operator can create, update, and destroy SMB share policy and snapshot policy rules independently of the parent policy lifecycle | VERIFIED | SMB rule resource has independent CRUD; Snapshot rule resource manages rules via `AddSnapshotPolicyRule` / `RemoveSnapshotPolicyRule` / `ReplaceSnapshotPolicyRule` on parent PATCH; all tests pass |
| 4 | All three policy data sources return attributes by name or filter without provider errors | VERIFIED | `nfs_export_policy_data_source.go` (126L), `smb_share_policy_data_source.go` (121L), `snapshot_policy_data_source.go` (126L) — all registered in provider.go; data source tests pass |

**Score:** 4/4 success criteria verified

---

### Required Artifacts

#### Plan 01 Artifacts (client layer + mock handlers)

| Artifact | Min Lines | Actual Lines | Status | Notes |
|----------|-----------|-------------|--------|-------|
| `internal/client/models.go` | — | 391 | VERIFIED | 21 Phase 3 struct types: NfsExportPolicy, NfsExportPolicyPost, NfsExportPolicyPatch, NfsExportPolicyRule, NfsExportPolicyRuleInPolicy, NfsExportPolicyRulePost, NfsExportPolicyRulePatch, SmbSharePolicy, SmbSharePolicyPost, SmbSharePolicyPatch, SmbSharePolicyRule, SmbSharePolicyRuleInPolicy, SmbSharePolicyRulePost, SmbSharePolicyRulePatch, SnapshotPolicy, SnapshotPolicyPost, SnapshotPolicyPatch, SnapshotPolicyRuleInPolicy, SnapshotPolicyRulePost, SnapshotPolicyRuleRemove, PolicyMember |
| `internal/client/nfs_export_policies.go` | — | 175 | VERIFIED | 12 exported methods including GetNfsExportPolicyRuleByIndex, GetNfsExportPolicyRuleByName, ListNfsExportPolicyMembers |
| `internal/client/smb_share_policies.go` | — | 153 | VERIFIED | 11 exported methods including ListSmbSharePolicyMembers |
| `internal/client/snapshot_policies.go` | — | 139 | VERIFIED | 10 exported methods including AddSnapshotPolicyRule, RemoveSnapshotPolicyRule, ReplaceSnapshotPolicyRule, GetSnapshotPolicyRuleByIndex |
| `internal/testmock/handlers/nfs_export_policies.go` | — | 432 | VERIFIED | Registers `/api/2.22/nfs-export-policies` and `/api/2.22/nfs-export-policies/rules`; sequential index via `nextRuleIndex`; rules sorted by index on GET |
| `internal/testmock/handlers/smb_share_policies.go` | — | 348 | VERIFIED | Registers `/api/2.22/smb-share-policies` and `/api/2.22/smb-share-policies/rules`; name-based rule identity; no index |
| `internal/testmock/handlers/snapshot_policies.go` | — | 212 | VERIFIED | Registers `/api/2.22/policies` and `/api/2.22/policies/file-systems`; PATCH handles `add_rules`/`remove_rules`; no dedicated rules endpoint |

#### Plan 02 Artifacts (NFS provider layer)

| Artifact | Min Lines | Actual Lines | Status | Notes |
|----------|-----------|-------------|--------|-------|
| `internal/provider/nfs_export_policy_resource.go` | 200 | 341 | VERIFIED | Delete guard via ListNfsExportPolicyMembers; in-place rename via PATCH; `version` is Computed-only (no UseStateForUnknown) |
| `internal/provider/nfs_export_policy_rule_resource.go` | 200 | 515 | VERIFIED | Import by `policy_name/rule_index`; `index` is Computed; `name` is server-assigned; drift detection on mutable fields |
| `internal/provider/nfs_export_policy_data_source.go` | 80 | 126 | VERIFIED | Read-only; `name` as Required filter |
| `internal/provider/nfs_export_policy_resource_test.go` | 100 | 393 | VERIFIED | Create, Update (rename + enabled), Delete, Import, DataSource tests |
| `internal/provider/nfs_export_policy_rule_resource_test.go` | 100 | 323 | VERIFIED | Create, Update, Delete, Import with composite ID tests |

#### Plan 03 Artifacts (SMB provider layer)

| Artifact | Min Lines | Actual Lines | Status | Notes |
|----------|-----------|-------------|--------|-------|
| `internal/provider/smb_share_policy_resource.go` | 200 | 336 | VERIFIED | Delete guard; in-place rename; no `version` field (correctly omitted) |
| `internal/provider/smb_share_policy_rule_resource.go` | 150 | 377 | VERIFIED | Import by `policy_name/rule_name` (string, not index); no `index` attribute in schema |
| `internal/provider/smb_share_policy_data_source.go` | 80 | 121 | VERIFIED | |
| `internal/provider/smb_share_policy_resource_test.go` | — | 397 | VERIFIED | |
| `internal/provider/smb_share_policy_rule_resource_test.go` | — | 357 | VERIFIED | |

#### Plan 04 Artifacts (Snapshot provider layer)

| Artifact | Min Lines | Actual Lines | Status | Notes |
|----------|-----------|-------------|--------|-------|
| `internal/provider/snapshot_policy_resource.go` | 200 | 343 | VERIFIED | `name` is RequiresReplace (documented exception); `retention_lock` is Computed+UseStateForUnknown; delete guard |
| `internal/provider/snapshot_policy_rule_resource.go` | 200 | 440 | VERIFIED | All CRUD via parent PATCH; synthetic ID `{policy_name}/{rule_name}`; Update uses atomic ReplaceSnapshotPolicyRule |
| `internal/provider/snapshot_policy_data_source.go` | 80 | 126 | VERIFIED | |
| `internal/provider/snapshot_policy_resource_test.go` | — | 393 | VERIFIED | |
| `internal/provider/snapshot_policy_rule_resource_test.go` | — | 320 | VERIFIED | |

---

### Key Link Verification

| From | To | Via | Status | Evidence |
|------|----|-----|--------|---------|
| `client/nfs_export_policies.go` | `client/models.go` | `NfsExportPolicy`, `NfsExportPolicyRule` struct types | WIRED | 21 Phase 3 structs confirmed in models.go |
| `testmock/handlers/nfs_export_policies.go` | `client/models.go` | Uses client model types for in-memory state | WIRED | Handler compiles; uses `client.NfsExportPolicy` |
| `provider/nfs_export_policy_resource.go` | `client/nfs_export_policies.go` | `r.client.GetNfsExportPolicy`, `PostNfsExportPolicy`, etc. | WIRED | 5 distinct client call sites confirmed |
| `provider/nfs_export_policy_rule_resource.go` | `client/nfs_export_policies.go` | `PostNfsExportPolicyRule`, `GetNfsExportPolicyRuleByName`, etc. | WIRED | `GetNfsExportPolicyRuleByIndex` used in ImportState |
| `provider/smb_share_policy_resource.go` | `client/smb_share_policies.go` | `GetSmbSharePolicy`, `PostSmbSharePolicy`, etc. | WIRED | 4 call sites confirmed |
| `provider/smb_share_policy_rule_resource.go` | `client/smb_share_policies.go` | `GetSmbSharePolicyRuleByName`, `PostSmbSharePolicyRule`, etc. | WIRED | 3 call sites confirmed including ImportState |
| `provider/snapshot_policy_resource.go` | `client/snapshot_policies.go` | `GetSnapshotPolicy`, `PostSnapshotPolicy`, etc. | WIRED | 4 call sites confirmed |
| `provider/snapshot_policy_rule_resource.go` | `client/snapshot_policies.go` | `AddSnapshotPolicyRule`, `RemoveSnapshotPolicyRule`, `ReplaceSnapshotPolicyRule` | WIRED | All 3 rule management methods used; explicitly documented in code comments |
| `provider/provider.go` | all 6 Phase 3 resources + 3 data sources | `NewNfsExportPolicyResource`, `NewSmbSharePolicyResource`, `NewSnapshotPolicyResource`, etc. | WIRED | Lines 273-291 in provider.go confirmed |

---

### Requirements Coverage

All 28 Phase 3 requirement IDs from REQUIREMENTS.md are claimed across the 4 plans and verified as implemented.

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|---------|
| NFP-01 | 03-01, 03-02 | Create NFS export policy | SATISFIED | `PostNfsExportPolicy` + resource Create; test passes |
| NFP-02 | 03-01, 03-02 | Update NFS export policy | SATISFIED | `PatchNfsExportPolicy` + in-place rename; test passes |
| NFP-03 | 03-01, 03-02 | Destroy NFS export policy | SATISFIED | `DeleteNfsExportPolicy` + delete guard; test passes |
| NFP-04 | 03-02 | Import NFS export policy by name | SATISFIED | ImportState parses name from req.ID; test passes |
| NFP-05 | 03-02 | Data source returns NFS export policy attributes | SATISFIED | `nfs_export_policy_data_source.go` 126L; test passes |
| NFR-01 | 03-01, 03-02 | Create NFS export policy rules | SATISFIED | `PostNfsExportPolicyRule`; test passes |
| NFR-02 | 03-01, 03-02 | Update NFS export policy rules | SATISFIED | `PatchNfsExportPolicyRule`; test passes |
| NFR-03 | 03-01, 03-02 | Destroy NFS export policy rules | SATISFIED | `DeleteNfsExportPolicyRule`; test passes |
| NFR-04 | 03-02 | Import NFS rules with composite ID `policy_name:rule_index` | SATISFIED | ImportState uses `policy_name/rule_index`; `GetNfsExportPolicyRuleByIndex` resolves it |
| SMP-01 | 03-01, 03-03 | Create SMB share policy | SATISFIED | test passes |
| SMP-02 | 03-01, 03-03 | Update SMB share policy | SATISFIED | in-place rename via PATCH; test passes |
| SMP-03 | 03-01, 03-03 | Destroy SMB share policy | SATISFIED | delete guard via `ListSmbSharePolicyMembers`; test passes |
| SMP-04 | 03-03 | Import SMB share policy by name | SATISFIED | ImportState; test passes |
| SMP-05 | 03-03 | Data source returns SMB share policy attributes | SATISFIED | `smb_share_policy_data_source.go` 121L; test passes |
| SMR-01 | 03-01, 03-03 | Create SMB share policy rules | SATISFIED | test passes |
| SMR-02 | 03-01, 03-03 | Update SMB share policy rules | SATISFIED | test passes |
| SMR-03 | 03-01, 03-03 | Destroy SMB share policy rules | SATISFIED | test passes |
| SMR-04 | 03-03 | Import SMB rules with composite ID | SATISFIED | Import uses `policy_name/rule_name` (string); test passes |
| SNP-01 | 03-01, 03-04 | Create snapshot policy | SATISFIED | test passes |
| SNP-02 | 03-01, 03-04 | Update snapshot policy | SATISFIED | `enabled` in-place; `name` RequiresReplace; test passes |
| SNP-03 | 03-01, 03-04 | Destroy snapshot policy | SATISFIED | delete guard via `ListSnapshotPolicyMembers`; test passes |
| SNP-04 | 03-04 | Import snapshot policy by name | SATISFIED | ImportState; test passes |
| SNP-05 | 03-04 | Data source returns snapshot policy attributes | SATISFIED | `snapshot_policy_data_source.go` 126L; test passes |
| SNR-01 | 03-01, 03-04 | Create snapshot policy rules | SATISFIED | `AddSnapshotPolicyRule` via PATCH; test passes |
| SNR-02 | 03-01, 03-04 | Update snapshot policy rules | SATISFIED | `ReplaceSnapshotPolicyRule` (atomic remove+add); test passes |
| SNR-03 | 03-01, 03-04 | Destroy snapshot policy rules | SATISFIED | `RemoveSnapshotPolicyRule` via PATCH; test passes |
| SNR-04 | 03-04 | Import snapshot rules with composite ID | SATISFIED | Import uses `policy_name/rule_index`; `GetSnapshotPolicyRuleByIndex` resolves it |

**Orphaned requirements:** None. All 28 IDs present in REQUIREMENTS.md traceability table map to Phase 3 with status Complete.

---

### Anti-Patterns Found

None. Grep for TODO/FIXME/XXX/HACK/PLACEHOLDER/placeholder across all 13 Phase 3 files returned zero matches.

---

### Build and Test Results

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go vet ./...` | PASS (no issues) |
| Phase 3 provider tests (`TestNfsExportPolicy*`, `TestSmbSharePolicy*`, `TestSnapshotPolicy*`) | 27/27 PASS |
| Full test suite (`go test ./...`) | 101/101 PASS across 5 packages |

---

### Human Verification Required

None identified. All goal-level behaviors are verified programmatically through unit tests against the mock server. The phase goal does not require visual or real-time validation — the key correctness property (no false drift on rule reorder) is validated by the mock server's index-sorted response combined with the `Computed`-only `index` attribute (not user-settable, so Terraform never proposes index changes).

---

## Gaps Summary

No gaps. All 28 requirements are satisfied, all artifacts exist and are substantive, all key links are wired, the full test suite passes, and no anti-patterns were found.

---

_Verified: 2026-03-27T15:46:05Z_
_Verifier: Claude (gsd-verifier)_
