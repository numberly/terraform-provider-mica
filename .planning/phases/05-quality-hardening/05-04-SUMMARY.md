---
phase: 05-quality-hardening
plan: 04
subsystem: testing
tags: [terraform-plugin-framework, testmock, lifecycle-testing, import-idempotency, go-test]

# Dependency graph
requires:
  - phase: 05-01
    provides: validator and plan modifier tests establishing unit test infrastructure
  - phase: 05-02
    provides: pagination, error-path tests, and mock server CRUD handler coverage
provides:
  - Full Create->Read->Update->Read->Delete lifecycle tests for all 19 resources
  - ImportState->Read->0-diff idempotency tests for all 18 importable resources
  - 37 new tests bringing provider package total to 187 unit tests
affects: [05-05, acceptance-testing]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Lifecycle test pattern: single mock server instance, step-by-step CRUD with state assertions"
    - "Import idempotency pattern: Create->ImportState->Read->field comparison"
    - "Singleton lifecycle: Create=PATCH(configure), Delete=PATCH(reset), import ID='default'"
    - "Composite import ID format: policy_name/rule_index (NFS/NAP/snapshot) or policy_name/rule_name (SMB/OAP)"

key-files:
  created: []
  modified:
    - internal/provider/filesystem_resource_test.go
    - internal/provider/bucket_resource_test.go
    - internal/provider/object_store_account_resource_test.go
    - internal/provider/object_store_access_key_resource_test.go
    - internal/provider/nfs_export_policy_resource_test.go
    - internal/provider/nfs_export_policy_rule_resource_test.go
    - internal/provider/smb_share_policy_resource_test.go
    - internal/provider/smb_share_policy_rule_resource_test.go
    - internal/provider/snapshot_policy_resource_test.go
    - internal/provider/snapshot_policy_rule_resource_test.go
    - internal/provider/object_store_access_policy_resource_test.go
    - internal/provider/object_store_access_policy_rule_resource_test.go
    - internal/provider/network_access_policy_resource_test.go
    - internal/provider/network_access_policy_rule_resource_test.go
    - internal/provider/quota_group_resource_test.go
    - internal/provider/quota_user_resource_test.go
    - internal/provider/array_dns_resource_test.go
    - internal/provider/array_ntp_resource_test.go
    - internal/provider/array_smtp_resource_test.go

key-decisions:
  - "AccessKey lifecycle is Create->Read->Delete only (no Update — all fields RequiresReplace; no Import — secret unavailable after creation)"
  - "Singleton lifecycle (NAP/DNS/NTP/SMTP) uses import ID='default' and Delete=PATCH(reset) pattern"
  - "OAP lifecycle Update uses name mutation (not description — that is RequiresReplace)"
  - "OAP rule lifecycle Update modifies actions list (effect is RequiresReplace)"
  - "Composite import IDs: NFS/NAP/snapshot rules use policy_name/index; SMB/OAP rules use policy_name/rule_name; quota resources use fs_name/gid or fs_name/uid"

patterns-established:
  - "Lifecycle test: create plan -> Create -> assert state -> Read -> assert state -> update plan -> Update -> assert state -> Read -> assert state -> Delete"
  - "Idempotency test: Create -> ImportState(id) -> Read -> compare key fields between post-create and post-import model structs"
  - "Policy rule composite ID: always policy_name + separator + identifier (no standalone ID)"
  - "Special-case detection: if all fields RequiresReplace, skip Update step; if no ImportState, skip idempotency test"

requirements-completed:
  - QUA-04

# Metrics
duration: 90min
completed: 2026-03-28
---

# Phase 5 Plan 04: Lifecycle and Import Idempotency Tests Summary

**37 new mocked integration tests covering Create->Read->Update->Read->Delete lifecycle and ImportState->Read->0-diff idempotency for all 19 resources (187 total tests passing)**

## Performance

- **Duration:** ~90 min
- **Started:** 2026-03-28T07:41Z (continuation from 05-02 session)
- **Completed:** 2026-03-28
- **Tasks:** 2
- **Files modified:** 19

## Accomplishments

- Added `TestUnit_{Resource}_Lifecycle` for all 19 resources exercising full CRUD method chain on a single mock server instance with state assertions at each step
- Added `TestUnit_{Resource}_ImportIdempotency` for all 18 importable resources (AccessKey excluded — no ImportState); each test verifies that ImportState->Read produces field values matching the original Create state
- All 187 provider unit tests pass with zero regressions (up from 141 before phase 5 work began)

## Task Commits

1. **Task 1+2: Bucket, filesystem, account, access_key families** - `002ce3a` (feat(05-02)) — lifecycle + idempotency for 4 resources
2. **Task 1+2: NFS rule, SMB, SMB rule, snapshot, snapshot rule, OAP rule** - `1043844` (test(05-02)) — lifecycle + idempotency for 6 resources, also NFS and OAP policies
3. **Task 1+2: NAP policy and rule** - `19424be` (test(05-02)) — lifecycle + idempotency for 2 resources
4. **Task 1+2: Quota group, quota user** - `129a591` (test(05-02)) — lifecycle + idempotency for 2 resources
5. **Task 1+2: Array DNS, NTP, SMTP singletons** - `4bf414a` (feat(05-04)) — lifecycle + idempotency for 3 resources

## Files Created/Modified

- `internal/provider/filesystem_resource_test.go` - TestUnit_FileSystem_Lifecycle, TestUnit_FileSystem_ImportIdempotency
- `internal/provider/bucket_resource_test.go` - TestUnit_Bucket_Lifecycle (update versioning), TestUnit_Bucket_ImportIdempotency
- `internal/provider/object_store_account_resource_test.go` - TestUnit_ObjectStoreAccount_Lifecycle, TestUnit_ObjectStoreAccount_ImportIdempotency
- `internal/provider/object_store_access_key_resource_test.go` - TestUnit_AccessKey_Lifecycle (Create->Read->Delete, no Update/Import)
- `internal/provider/nfs_export_policy_resource_test.go` - TestUnit_NfsExportPolicy_Lifecycle, TestUnit_NfsExportPolicy_ImportIdempotency
- `internal/provider/nfs_export_policy_rule_resource_test.go` - TestUnit_NfsExportPolicyRule_Lifecycle (update client CIDR), TestUnit_NfsExportPolicyRule_ImportIdempotency
- `internal/provider/smb_share_policy_resource_test.go` - TestUnit_SmbSharePolicy_Lifecycle, TestUnit_SmbSharePolicy_ImportIdempotency
- `internal/provider/smb_share_policy_rule_resource_test.go` - TestUnit_SmbSharePolicyRule_Lifecycle (update change field), TestUnit_SmbSharePolicyRule_ImportIdempotency
- `internal/provider/snapshot_policy_resource_test.go` - TestUnit_SnapshotPolicy_Lifecycle, TestUnit_SnapshotPolicy_ImportIdempotency
- `internal/provider/snapshot_policy_rule_resource_test.go` - TestUnit_SnapshotPolicyRule_Lifecycle (update keep_for 7d->14d), TestUnit_SnapshotPolicyRule_ImportIdempotency
- `internal/provider/object_store_access_policy_resource_test.go` - TestUnit_ObjectStoreAccessPolicy_Lifecycle (update name), TestUnit_ObjectStoreAccessPolicy_ImportIdempotency
- `internal/provider/object_store_access_policy_rule_resource_test.go` - TestUnit_OAPRule_Lifecycle (update actions list), TestUnit_OAPRule_ImportIdempotency
- `internal/provider/network_access_policy_resource_test.go` - TestUnit_NetworkAccessPolicy_Lifecycle (singleton), TestUnit_NetworkAccessPolicy_ImportIdempotency
- `internal/provider/network_access_policy_rule_resource_test.go` - TestUnit_NetworkAccessPolicyRule_Lifecycle, TestUnit_NetworkAccessPolicyRule_ImportIdempotency
- `internal/provider/quota_group_resource_test.go` - TestUnit_QuotaGroup_Lifecycle (update quota 1GiB->2GiB), TestUnit_QuotaGroup_ImportIdempotency
- `internal/provider/quota_user_resource_test.go` - TestUnit_QuotaUser_Lifecycle (update quota 1GiB->2GiB), TestUnit_QuotaUser_ImportIdempotency
- `internal/provider/array_dns_resource_test.go` - TestUnit_ArrayDns_Lifecycle (singleton, update nameservers), TestUnit_ArrayDns_ImportIdempotency
- `internal/provider/array_ntp_resource_test.go` - TestUnit_ArrayNtp_Lifecycle (singleton, update servers), TestUnit_ArrayNtp_ImportIdempotency
- `internal/provider/array_smtp_resource_test.go` - TestUnit_ArraySmtp_Lifecycle (singleton with alert_watchers), TestUnit_ArraySmtp_ImportIdempotency

## Decisions Made

- AccessKey gets only a Create->Read->Delete lifecycle test — no Update because all fields are RequiresReplace, no Import because the secret_access_key is unavailable after creation. This accurately reflects the resource's operational semantics.
- OAP lifecycle Update step mutates the policy name (not description which is RequiresReplace), and OAP rule Update step adds s3:PutObject to the actions list (effect is RequiresReplace).
- Singleton resources (NAP, DNS, NTP, SMTP) use import ID `"default"` to match their ImportState implementation. Delete step verifies the reset-to-defaults behavior (PATCH with empty/false values).
- Composite import IDs follow the existing convention: numeric index for NFS/NAP/snapshot rules, string rule_name for SMB/OAP rules, `fs_name/gid` or `fs_name/uid` for quota resources.

## Deviations from Plan

None - plan executed exactly as written. All 19 lifecycle tests and 18 import idempotency tests implemented using existing test helpers and mock handlers without any production code changes.

## Issues Encountered

None — all tests passed on first run. The existing mock server handlers already support full CRUD lifecycles, and the test helper functions (plan builders, schema helpers) were already available in each test file.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Full mocked integration coverage achieved: 187 unit tests across all 19 resources
- All lifecycle paths verified: Create->Read->Update->Read->Delete + ImportState->Read->0-diff
- Ready for phase 05-05: acceptance test scaffolding and documentation generation
- No blockers identified

---
*Phase: 05-quality-hardening*
*Completed: 2026-03-28*
