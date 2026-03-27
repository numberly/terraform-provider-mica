---
phase: 03-file-based-policy-resources
plan: "01"
subsystem: client-layer
tags: [client, mock, nfs, smb, snapshot, policy, models]
dependency_graph:
  requires: []
  provides:
    - NfsExportPolicy client CRUD + mock handler
    - SmbSharePolicy client CRUD + mock handler
    - SnapshotPolicy client CRUD + mock handler
    - All Phase 3 model structs in models.go
  affects:
    - internal/client/models.go
    - All Phase 3 resource and data source plans (depend on these client methods)
tech_stack:
  added: []
  patterns:
    - FlashBlade ListResponse[T] generic wrapper for all client GET responses
    - PATCH raw map decode for true partial-update semantics in mock handlers
    - Sequential index auto-assignment for NFS rules
    - Name-based rule identity for SMB rules (no index)
    - add_rules/remove_rules PATCH body pattern for snapshot rules
key_files:
  created:
    - internal/client/nfs_export_policies.go
    - internal/client/smb_share_policies.go
    - internal/client/snapshot_policies.go
    - internal/testmock/handlers/nfs_export_policies.go
    - internal/testmock/handlers/smb_share_policies.go
    - internal/testmock/handlers/snapshot_policies.go
  modified:
    - internal/client/models.go
decisions:
  - "NFS rule GET model uses int for anonuid/anongid (API integer type), PATCH model uses *string (API schema difference — confirmed in FLASHBLADE_API.md)"
  - "Snapshot policy name is read-only on PATCH — client enforces this by omitting name from SnapshotPolicyPatch struct"
  - "Snapshot mock PATCH handler processes remove_rules before add_rules to allow atomic replace semantics"
  - "ListNfsExportPolicyMembers and ListSmbSharePolicyMembers use file-systems?filter= endpoint for delete-guard; ListSnapshotPolicyMembers uses /policies/file-systems?policy_names="
metrics:
  duration_minutes: 27
  completed_date: "2026-03-27"
  tasks_completed: 2
  files_created: 6
  files_modified: 1
---

# Phase 3 Plan 01: File-Based Policy Client Layer Summary

**One-liner:** NFS/SMB/Snapshot policy CRUD client methods and mock HTTP handlers establishing three distinct rule management patterns (NFS: sequential index, SMB: name-based, Snapshot: parent PATCH add_rules/remove_rules).

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Add all Phase 3 model structs and NFS client + mock | 8a1fa53 | models.go (+160 lines), nfs_export_policies.go (client), handlers/nfs_export_policies.go (mock) |
| 2 | Add SMB and Snapshot client methods + mock handlers | de040dd | smb_share_policies.go, snapshot_policies.go, handlers/smb_share_policies.go, handlers/snapshot_policies.go |

## What Was Built

### Model Structs (models.go)

9 new struct families appended after the Phase 2 section:

- `PolicyMember` — shared delete-guard member reference
- NFS: `NfsExportPolicy`, `NfsExportPolicyPost`, `NfsExportPolicyPatch`, `NfsExportPolicyRule`, `NfsExportPolicyRuleInPolicy`, `NfsExportPolicyRulePost`, `NfsExportPolicyRulePatch`
- SMB: `SmbSharePolicy`, `SmbSharePolicyPost`, `SmbSharePolicyPatch`, `SmbSharePolicyRule`, `SmbSharePolicyRuleInPolicy`, `SmbSharePolicyRulePost`, `SmbSharePolicyRulePatch`
- Snapshot: `SnapshotPolicy`, `SnapshotPolicyPost`, `SnapshotPolicyPatch`, `SnapshotPolicyRuleInPolicy`, `SnapshotPolicyRulePost`, `SnapshotPolicyRuleRemove`

### Client Methods

**NFS Export Policies** (`/nfs-export-policies`, `/nfs-export-policies/rules`):
- `GetNfsExportPolicy`, `ListNfsExportPolicies`, `PostNfsExportPolicy`, `PatchNfsExportPolicy`, `DeleteNfsExportPolicy`
- `ListNfsExportPolicyRules`, `GetNfsExportPolicyRuleByIndex`, `GetNfsExportPolicyRuleByName`, `PostNfsExportPolicyRule`, `PatchNfsExportPolicyRule`, `DeleteNfsExportPolicyRule`
- `ListNfsExportPolicyMembers` (delete guard via filter query on /file-systems)

**SMB Share Policies** (`/smb-share-policies`, `/smb-share-policies/rules`):
- `GetSmbSharePolicy`, `ListSmbSharePolicies`, `PostSmbSharePolicy`, `PatchSmbSharePolicy`, `DeleteSmbSharePolicy`
- `ListSmbSharePolicyRules`, `GetSmbSharePolicyRuleByName`, `PostSmbSharePolicyRule`, `PatchSmbSharePolicyRule`, `DeleteSmbSharePolicyRule`
- `ListSmbSharePolicyMembers` (delete guard via filter query on /file-systems)

**Snapshot Policies** (`/policies`, `/policies/file-systems`):
- `GetSnapshotPolicy`, `ListSnapshotPolicies`, `PostSnapshotPolicy`, `PatchSnapshotPolicy`, `DeleteSnapshotPolicy`
- `AddSnapshotPolicyRule`, `RemoveSnapshotPolicyRule`, `ReplaceSnapshotPolicyRule` (all via PATCH parent)
- `GetSnapshotPolicyRuleByIndex` (for import support)
- `ListSnapshotPolicyMembers` (via /policies/file-systems?policy_names=)

### Mock Handlers

- `RegisterNfsExportPolicyHandlers` — registers `/api/2.22/nfs-export-policies` and `/api/2.22/nfs-export-policies/rules`; auto-assigns sequential index per policy
- `RegisterSmbSharePolicyHandlers` — registers `/api/2.22/smb-share-policies` and `/api/2.22/smb-share-policies/rules`; no index, name-based rule identity
- `RegisterSnapshotPolicyHandlers` — registers `/api/2.22/policies` and `/api/2.22/policies/file-systems`; PATCH handler processes remove_rules then add_rules

## Decisions Made

1. **NFS anonuid/anongid type divergence** — GET model uses `int` (API returns integer), PATCH model uses `*string` (API PATCH body accepts string). This matches the FLASHBLADE_API.md schema which lists `anongid/anonuid: integer` for POST and `string` for PATCH.

2. **Snapshot name read-only** — `SnapshotPolicyPatch` struct omits the `Name` field entirely, making it structurally impossible to send a name change. This matches API documentation where `name` is ro on snapshot policy PATCH.

3. **Snapshot mock rule ordering** — `remove_rules` is processed before `add_rules` in the PATCH handler, enabling atomic replace-rule semantics (used by `ReplaceSnapshotPolicyRule`).

4. **Delete-guard endpoint split** — NFS and SMB use `GET /file-systems?filter=<protocol>.policy.name='X'` while Snapshot uses `GET /policies/file-systems?policy_names=X`, matching the actual API surface for each family.

## Deviations from Plan

None — plan executed exactly as written.

## Verification

```
go build ./...  → SUCCESS
go vet ./...    → SUCCESS (no issues)
```

## Self-Check: PASSED

- `internal/client/models.go` — exists, contains NfsExportPolicy, SmbSharePolicy, SnapshotPolicy
- `internal/client/nfs_export_policies.go` — exists, 8a1fa53 confirmed
- `internal/client/smb_share_policies.go` — exists, de040dd confirmed
- `internal/client/snapshot_policies.go` — exists, de040dd confirmed
- `internal/testmock/handlers/nfs_export_policies.go` — exists, 8a1fa53 confirmed
- `internal/testmock/handlers/smb_share_policies.go` — exists, de040dd confirmed
- `internal/testmock/handlers/snapshot_policies.go` — exists, de040dd confirmed
