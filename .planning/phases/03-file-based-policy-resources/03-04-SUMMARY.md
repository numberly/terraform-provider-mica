---
phase: 03-file-based-policy-resources
plan: "04"
subsystem: provider
tags: [snapshot-policy, snapshot-policy-rule, data-source, terraform-plugin-framework, patch-based-rules]
dependency_graph:
  requires: ["03-01", "03-02"]
  provides:
    - flashblade_snapshot_policy resource
    - flashblade_snapshot_policy_rule resource
    - flashblade_snapshot_policy data source
  affects:
    - internal/provider/provider.go (10 resources, 7 data sources)
tech_stack:
  added: []
  patterns:
    - RequiresReplace on name (snapshot policy name is API read-only)
    - PATCH add_rules/remove_rules for rule lifecycle (no dedicated endpoint)
    - Synthetic composite ID: {policy_name}/{rule_name} for rules
    - ReplaceSnapshotPolicyRule: atomic remove+add in single PATCH for updates
    - findRuleByName helper for O(n) scan through embedded rules array
key_files:
  created:
    - internal/provider/snapshot_policy_resource.go
    - internal/provider/snapshot_policy_resource_test.go
    - internal/provider/snapshot_policy_data_source.go
    - internal/provider/snapshot_policy_rule_resource.go
    - internal/provider/snapshot_policy_rule_resource_test.go
  modified:
    - internal/provider/provider.go
decisions:
  - "snapshot policy name has RequiresReplace — API does not support rename via PATCH (SnapshotPolicyPatch has no Name field)"
  - "Rule ID is synthetic: {policy_name}/{rule_name} — snapshot rules have no server-issued UUID"
  - "Update uses ReplaceSnapshotPolicyRule (atomic PATCH with remove_rules + add_rules) — no dedicated rule PATCH endpoint"
  - "New rule after Create/Replace identified as last element in policy.Rules array"
  - "findRuleByName: O(n) scan is acceptable for typical rule counts (<100)"
metrics:
  duration_seconds: 474
  completed_date: "2026-03-26"
  tasks_completed: 2
  files_changed: 6
---

# Phase 3 Plan 4: Snapshot Policy Resources Summary

Snapshot policy resource, rule resource, and data source — CRUD via parent PATCH add_rules/remove_rules with RequiresReplace name constraint and synthetic composite IDs for rules.

## What Was Built

### flashblade_snapshot_policy resource
- `name`: RequiresReplace — snapshot policy names are read-only after creation (API constraint). This is the documented exception to the uniform rename pattern used by NFS/SMB policies.
- `enabled`: only in-place mutable field via PATCH (no `name` in SnapshotPolicyPatch body)
- Delete guard: `ListSnapshotPolicyMembers` blocks delete if file systems are attached
- Import by name: restores all computed fields including `retention_lock`

### flashblade_snapshot_policy_rule resource
- No dedicated API endpoint — all CRUD goes through PATCH on the parent policy:
  - **Create**: `AddSnapshotPolicyRule` (PATCH add_rules), new rule identified as last in `policy.Rules`
  - **Read**: `GetSnapshotPolicy` + `findRuleByName` scan
  - **Update**: `ReplaceSnapshotPolicyRule` (atomic PATCH remove_rules + add_rules)
  - **Delete**: `RemoveSnapshotPolicyRule` (PATCH remove_rules)
- Synthetic ID: `{policy_name}/{rule_name}` — rules have no server-issued UUID
- Import: composite ID `policy_name/rule_index` resolved via `GetSnapshotPolicyRuleByIndex`
- Schedule fields: `at`, `every`, `keep_for`, `suffix`, `client_name` — all Optional+Computed

### flashblade_snapshot_policy data source
- Reads policy by name: returns `id`, `enabled`, `is_local`, `policy_type`, `retention_lock`

### provider.go
Now registers **10 resources** and **7 data sources**:
- Resources: Filesystem, ObjectStoreAccount, Bucket, AccessKey, NfsExportPolicy, NfsExportPolicyRule, SmbSharePolicy, SmbSharePolicyRule, SnapshotPolicy, SnapshotPolicyRule
- DataSources: Filesystem, ObjectStoreAccount, Bucket, AccessKey, NfsExportPolicy, SmbSharePolicy, SnapshotPolicy

## Test Results

- `TestSnapshotPolicyResource_Create` — PASS
- `TestSnapshotPolicyResource_Update` — PASS (enabled toggle, name unchanged)
- `TestSnapshotPolicyResource_Delete` — PASS
- `TestSnapshotPolicyResource_Import` — PASS
- `TestSnapshotPolicyDataSource` — PASS
- `TestSnapshotPolicyRuleResource_Create` — PASS (every=86400000, keep_for=604800000)
- `TestSnapshotPolicyRuleResource_Update` — PASS (keep_for updated to 1209600000 via replace)
- `TestSnapshotPolicyRuleResource_Delete` — PASS (rules array empty after delete)
- `TestSnapshotPolicyRuleResource_Import` — PASS (policy_name/0 composite ID)

Full test suite: **101 tests pass, 0 failures** across 5 packages.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] SMB rule resource was missing (plan 03-03 incomplete)**

- **Found during**: Task 1 RED phase — package failed to compile due to `smb_share_policy_rule_resource_test.go` referencing `smbSharePolicyRuleResource` which did not exist
- **Issue**: Plan 03-03 (SMB) created test files but the SMB rule resource implementation (`smb_share_policy_rule_resource.go`) was never created. This prevented the test package from building at all.
- **Fix**: Confirmed `smb_share_policy_rule_resource.go` actually exists at 12.4KB — it was already there. The package compiled successfully after writing the new snapshot files. The test errors were caused by the snapshot types not existing yet (which was expected — RED phase).
- **Outcome**: No action needed — SMB rule resource was already implemented as part of plan 03-03.

None — plan executed exactly as written, plus corrected initial misdiagnosis of SMB missing file.

## Self-Check: PASSED

- snapshot_policy_resource.go: FOUND (343 lines, exceeds 200 minimum)
- snapshot_policy_resource_test.go: FOUND
- snapshot_policy_data_source.go: FOUND (126 lines, exceeds 80 minimum)
- snapshot_policy_rule_resource.go: FOUND (440 lines, exceeds 200 minimum)
- snapshot_policy_rule_resource_test.go: FOUND
- Task 1 commit 9b4eb3b: FOUND
- Task 2 commit 6f5bbe7: FOUND
- Full test suite: 101 passed, 0 failures
