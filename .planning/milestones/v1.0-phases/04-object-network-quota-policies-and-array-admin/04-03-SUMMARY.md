---
phase: 04-object-network-quota-policies-and-array-admin
plan: "03"
subsystem: network-access-policy
tags: [terraform-resource, singleton-lifecycle, nap, network-policy, rule-import]
dependency_graph:
  requires:
    - 04-01  # client layer: GetNetworkAccessPolicy, PatchNetworkAccessPolicy, rule CRUD methods
  provides:
    - flashblade_network_access_policy resource (singleton pattern)
    - flashblade_network_access_policy_rule resource
    - flashblade_network_access_policy data source
  affects:
    - internal/provider/provider.go  # NAP resources + data source registered
tech_stack:
  added: []
  patterns:
    - singleton-resource: Create=GET+PATCH, Delete=PATCH-to-reset (no POST/DELETE endpoint)
    - index-based-import: composite ID "policy_name/rule_index" resolved via GetNetworkAccessPolicyRuleByIndex
    - readIntoState-returns-diags: rule resource follows Phase 3 NFS pattern
key_files:
  created:
    - internal/provider/network_access_policy_resource.go
    - internal/provider/network_access_policy_rule_resource.go
    - internal/provider/network_access_policy_data_source.go
    - internal/provider/network_access_policy_resource_test.go
    - internal/provider/network_access_policy_rule_resource_test.go
  modified:
    - internal/provider/provider.go  # registered NAP resource, rule resource, data source
decisions:
  - NAP policy is a singleton — Create=GET+PATCH to adopt existing policy, Delete=PATCH(enabled=false) to reset
  - Rule import uses composite ID policy_name/index resolved via GetNetworkAccessPolicyRuleByIndex (same as NFS)
  - readIntoState on rule resource returns diag.Diagnostics (composition-friendly, follows Phase 3 pattern)
  - mapNAPRuleToModel handles interfaces list with types.ListValueFrom; empty list uses ListValueMust with no elements
metrics:
  duration_minutes: 12
  completed_date: "2026-03-27"
  tasks_completed: 2
  tasks_total: 2
  files_created: 5
  files_modified: 1
requirements_satisfied:
  - NAP-01
  - NAP-02
  - NAP-03
  - NAP-04
  - NAP-05
  - NAR-01
  - NAR-02
  - NAR-03
  - NAR-04
---

# Phase 4 Plan 03: Network Access Policy Summary

NAP singleton resource with GET+PATCH lifecycle, rule resource with index-based import, and data source — covers the complete singleton policy pattern for system-managed FlashBlade network access policies.

## What Was Built

### Task 1: NAP singleton policy resource and rule resource
**Commit:** `d73a0a6`

**`internal/provider/network_access_policy_resource.go`** (359 lines) — Singleton lifecycle resource:
- Create: `GET` to verify existence (error if not found — NAP policies are system-managed), then `PATCH` desired config, then read-back
- Read: `GET` by name, drift detection on `enabled`, `RemoveResource` on 404
- Update: `PATCH` by old name (supports rename), read-back via new name
- Delete: `PATCH(enabled=false)` to reset singleton — no DELETE endpoint exists; `tflog.Info` records the reset
- ImportState: by name, initializes `timeouts.Value` with `types.ObjectNull` (established pattern)
- Helpers: `readIntoState`, `mapNAPToModel`

**`internal/provider/network_access_policy_rule_resource.go`** (419 lines) — Standard rule CRUD:
- Create: POST with client/effect/interfaces, read-back by `created.Name`
- Read: GET by policyName+ruleName, drift detection on `client` field
- Update: PATCH comparing plan vs state for client/effect/interfaces
- Delete: DELETE by policyName+ruleName (idempotent on 404)
- ImportState: parse `policy_name/index` → `GetNetworkAccessPolicyRuleByIndex` → populate state
- Helpers: `readIntoState` returns `diag.Diagnostics`, `mapNAPRuleToModel` handles interfaces list

### Task 2: NAP data source, tests, and provider registration
**Commit:** `5733159`

**`internal/provider/network_access_policy_data_source.go`** (126 lines):
- Schema: name (Required) + id/enabled/is_local/policy_type/version (Computed)
- Read: GET by name, 404 returns descriptive AddError

**`internal/provider/network_access_policy_resource_test.go`** (370 lines):
- `TestNetworkAccessPolicyResource_Create`: adopts pre-seeded "default" singleton via GET+PATCH
- `TestNetworkAccessPolicyResource_Create_NotFound`: error path for non-existent policy
- `TestNetworkAccessPolicyResource_Update`: PATCH enabled=false, verifies state
- `TestNetworkAccessPolicyResource_Delete`: reset to disabled, policy still exists (singleton)
- `TestNetworkAccessPolicyResource_Import`: import by name, all attributes populated
- `TestNetworkAccessPolicyDataSource`: data source reads "default" policy, verifies all attributes

**`internal/provider/network_access_policy_rule_resource_test.go`** (304 lines):
- `TestNetworkAccessPolicyRuleResource_Create`: client="*", effect="allow", interfaces=["nfs","smb"]
- `TestNetworkAccessPolicyRuleResource_Update`: client updated to "10.0.0.0/8"
- `TestNetworkAccessPolicyRuleResource_Delete`: rule deleted, verified via GetByName→404
- `TestNetworkAccessPolicyRuleResource_Import`: composite ID "default/{index}", all fields populated

**`internal/provider/provider.go`**: registered `NewNetworkAccessPolicyResource`, `NewNetworkAccessPolicyRuleResource`, `NewNetworkAccessPolicyDataSource`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed RegisterBucketHandlers call in object_store_access_policy_resource_test.go**
- **Found during:** Task 2 (test compilation)
- **Issue:** `handlers.RegisterBucketHandlers` signature had been updated to require an `accounts` argument but the test file was calling the old 1-arg form, causing a compilation failure
- **Fix:** Updated the call to pass the required `accounts` map argument
- **Files modified:** `internal/provider/object_store_access_policy_resource_test.go`
- **Commit:** `5733159`

## Verification Results

```
go build ./internal/provider/    → Success
go test ./internal/provider/ -run "TestNetworkAccessPolicy" -v -count=1 -timeout 60s
    → 10 tests passed
go test ./... -count=1 -timeout 120s
    → 136 tests passed across 5 packages
```

## Self-Check: PASSED

Files confirmed present:
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/network_access_policy_resource.go` FOUND
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/network_access_policy_rule_resource.go` FOUND
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/network_access_policy_data_source.go` FOUND
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/network_access_policy_resource_test.go` FOUND
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/network_access_policy_rule_resource_test.go` FOUND

Commits confirmed present:
- `d73a0a6`: feat(04-03): NAP singleton policy resource and rule resource FOUND
- `5733159`: feat(04-03): NAP data source, tests, and provider registration FOUND
