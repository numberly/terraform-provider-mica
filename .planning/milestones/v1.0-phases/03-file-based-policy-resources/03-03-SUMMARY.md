---
phase: 03-file-based-policy-resources
plan: "03"
subsystem: provider
tags: [smb, share-policy, resource, data-source, tdd]
dependency_graph:
  requires: [03-01, 03-02]
  provides: [flashblade_smb_share_policy, flashblade_smb_share_policy_rule, flashblade_smb_share_policy data source]
  affects: [provider.go]
tech_stack:
  added: []
  patterns: [resource-with-import, name-based-rule-import, delete-guard, tdd-unit-tests]
key_files:
  created:
    - internal/provider/smb_share_policy_resource.go
    - internal/provider/smb_share_policy_data_source.go
    - internal/provider/smb_share_policy_resource_test.go
    - internal/provider/smb_share_policy_rule_resource.go
    - internal/provider/smb_share_policy_rule_resource_test.go
  modified: []
decisions:
  - "SMB policy has no Version field — omitted from schema (unlike NFS export policy)"
  - "SMB rule import uses composite ID policy_name/rule_name (string) — not numeric index like NFS"
  - "SMB rule name is Computed+UseStateForUnknown (server-assigned stable identifier)"
  - "Delete guard for policy calls ListSmbSharePolicyMembers; requires file-systems handler in delete test"
  - "provider.go already registered all SMB and Snapshot types (pre-committed by a prior agent)"
metrics:
  duration_minutes: 30
  completed_date: "2026-03-26"
  tasks_completed: 2
  files_created: 5
  files_modified: 0
---

# Phase 3 Plan 03: SMB Share Policy Resources Summary

**One-liner:** SMB share policy resource (in-place rename, delete guard), rule resource (name-based import — no index), and read-only data source — all backed by mock-server unit tests.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | SMB share policy resource + data source | 802c719 | smb_share_policy_resource.go, smb_share_policy_data_source.go, smb_share_policy_resource_test.go |
| 2 | SMB share policy rule resource + provider registration | 6d1d994 | smb_share_policy_rule_resource.go, smb_share_policy_rule_resource_test.go |

## What Was Built

**flashblade_smb_share_policy resource** (`smb_share_policy_resource.go`):
- Full CRUD with timeout support
- `name` has no RequiresReplace — rename is applied via PATCH in a single operation
- `enabled` defaults to `true` via `booldefault.StaticBool(true)`
- No `version` field (unlike NFS export policy — SMB policy has no version concept)
- Delete guard: calls `ListSmbSharePolicyMembers` and blocks with a clear diagnostic if attached
- ImportState: reads by name, initializes null timeouts object

**flashblade_smb_share_policy data source** (`smb_share_policy_data_source.go`):
- Read-only by name; surfaces id, name, enabled, is_local, policy_type (no version field)

**flashblade_smb_share_policy_rule resource** (`smb_share_policy_rule_resource.go`):
- `policy_name` has RequiresReplace — rules belong to one policy and cannot be moved
- `name` is Computed+UseStateForUnknown (server-assigned string, stable identifier for PATCH/DELETE)
- NO `index` attribute — SMB rules have no ordering concept (key difference from NFS)
- Rule fields: principal (Required), change (Optional/Computed), full_control (Optional/Computed), read (Optional/Computed)
- Composite import ID: `policy_name/rule_name` where rule_name is the server-assigned string (not a numeric index)
- `readIntoState` returns `diag.Diagnostics` for clean caller composition

**Provider registration** (`provider.go`):
- All 3 SMB types were already registered from a prior agent execution (plans 03-04 pre-committed)
- No changes needed to provider.go in this plan

## Test Results

- 9 SMB-specific tests pass (Create, Update+rename, Delete, Import for resource + DataSource; Create, Update, Delete, Import for rule resource)
- 101 total tests pass — no regressions
- Delete test registers FileSystem handler to satisfy `ListSmbSharePolicyMembers` (GET /file-systems)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Pre-existing orphan test files blocked compilation**
- **Found during:** Task 1 RED phase (and Task 2 RED phase)
- **Issue:** `snapshot_policy_resource_test.go` and `snapshot_policy_rule_resource_test.go` were untracked files referencing `snapshotPolicyResource` and `snapshotPolicyRuleResource` types that appeared undefined. Build failed with multiple "undefined" errors.
- **Investigation:** The implementation files `snapshot_policy_resource.go` and `snapshot_policy_rule_resource.go` already existed on disk (14.5KB+ each) and `go build ./internal/provider/` succeeded — the "undefined" errors in earlier test runs were false alarms from the rtk tool output ordering. Tests could run after the SMB files were created.
- **Fix:** No action needed — the snapshot files and provider.go registration were already present from a prior agent run (commits 9b4eb3b and 6f5bbe7).
- **Impact:** None — compilation succeeded, all 101 tests pass.

## Self-Check: PASSED

- internal/provider/smb_share_policy_resource.go: FOUND
- internal/provider/smb_share_policy_data_source.go: FOUND
- internal/provider/smb_share_policy_rule_resource.go: FOUND
- commit 802c719 (Task 1): FOUND
- commit 6d1d994 (Task 2): FOUND
- go test ./... → 101 passed, 0 failed
