---
phase: 08-smb-client-policies-syslog-and-acceptance-tests
verified: 2026-03-28T18:00:00Z
status: passed
score: 12/12 must-haves verified
re_verification: false
human_verification:
  - test: "Acceptance test idempotency — tofu plan after apply shows 0 changes"
    expected: "No resource drift on SMB client policy, syslog server, S3 export policy, and virtual host after apply"
    why_human: "Requires live FlashBlade array; cannot verify programmatically"
  - test: "Acceptance test destroy — tofu destroy removes all resources cleanly"
    expected: "All test-gule-* resources deleted with no errors"
    why_human: "Requires live FlashBlade array; cannot verify programmatically"
---

# Phase 8: SMB Client Policies, Syslog, and Acceptance Tests — Verification Report

**Phase Goal:** Remaining resource types are implemented and all v1.1 resources pass acceptance tests against a live FlashBlade
**Verified:** 2026-03-28
**Status:** PASSED
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| #  | Truth                                                                              | Status     | Evidence                                                                 |
|----|------------------------------------------------------------------------------------|------------|--------------------------------------------------------------------------|
| 1  | Operator can create an SMB client policy with enabled toggle and read it back      | VERIFIED   | `TestUnit_SmbClientPolicy_CRUD` passes; resource uses `PostSmbClientPolicy` + `GetSmbClientPolicy` |
| 2  | Operator can create SMB client policy rules with client/encryption/permission      | VERIFIED   | `TestUnit_SmbClientPolicyRule_CRUD` passes; `SmbClientPolicyRulePost` has `Client`, `Encryption`, `Permission` fields |
| 3  | Operator can update and delete SMB client policy rules independently               | VERIFIED   | CRUD test exercises Update (client field change) and Delete; `PatchSmbClientPolicyRule` + `DeleteSmbClientPolicyRule` wired |
| 4  | Operator can import SMB client policies and rules into Terraform state             | VERIFIED   | `TestUnit_SmbClientPolicy_Import` and `TestUnit_SmbClientPolicyRule_Import` pass; composite ID `policy_name/rule_name` |
| 5  | SMB client policy data source reads policy by name                                 | VERIFIED   | `TestUnit_SmbClientPolicy_DataSource` passes; `GetSmbClientPolicy` called from data source |
| 6  | Operator can create a syslog server with URI, services, and sources                | VERIFIED   | `TestUnit_SyslogServer_CRUD` passes; `PostSyslogServer` wired in resource |
| 7  | Operator can update syslog server URI, services, and sources in-place              | VERIFIED   | CRUD test updates URI and services; `PatchSyslogServer` called with pointer fields |
| 8  | Operator can import an existing syslog server into Terraform state                 | VERIFIED   | `TestUnit_SyslogServer_Import` passes; import sets Name and calls `readIntoState` |
| 9  | Syslog server data source reads server by name                                     | VERIFIED   | `TestUnit_SyslogServer_DataSource` passes; `GetSyslogServer` wired in data source |
| 10 | Null-vs-empty drift does not occur on services and sources list attributes         | VERIFIED   | `listdefault.StaticValue(types.ListValueMust(...))` at lines 84 and 91 of `syslog_server_resource.go` |
| 11 | SMB client policy and rule can be created, read, and destroyed on live FlashBlade  | VERIFIED*  | Acceptance test HCL present; SUMMARY reports full apply+destroy cycle passed (commit `f3894ba` + `fa8cd8a`) |
| 12 | Syslog server can be created, read, and destroyed on live FlashBlade               | VERIFIED*  | `test_syslog_servers.tf` present; SUMMARY reports live validation passed |

*Human-executed acceptance test: automated checks confirm HCL files exist and are syntactically valid; live run confirmed by operator per SUMMARY.

**Score:** 12/12 truths verified

---

## Required Artifacts

### Plan 08-01 — SMB Client Policies

| Artifact                                                         | Provides                                      | Status     | Details                                              |
|------------------------------------------------------------------|-----------------------------------------------|------------|------------------------------------------------------|
| `internal/client/models.go`                                      | 7 SmbClientPolicy* model structs              | VERIFIED   | Lines 350–413: all 7 types confirmed                 |
| `internal/client/smb_client_policies.go`                        | 12 CRUD methods                               | VERIFIED   | All 12 methods present (Get, List, Post, Patch, Delete for policy + rules + Members) |
| `internal/testmock/handlers/smb_client_policies.go`             | Mock handler — `RegisterSmbClientPolicyHandlers` | VERIFIED | Function present at line 22                          |
| `internal/provider/smb_client_policy_resource.go`               | `flashblade_smb_client_policy` resource       | VERIFIED   | `NewSmbClientPolicyResource` at line 33; 12.6K substantive |
| `internal/provider/smb_client_policy_rule_resource.go`          | `flashblade_smb_client_policy_rule` resource  | VERIFIED   | `NewSmbClientPolicyRuleResource` present; 12.4K substantive |
| `internal/provider/smb_client_policy_data_source.go`            | `flashblade_smb_client_policy` data source    | VERIFIED   | `NewSmbClientPolicyDataSource` at line 24            |
| `internal/provider/smb_client_policy_resource_test.go`          | Policy unit tests                             | VERIFIED   | 3 test functions: CRUD, Import, DataSource; 15.3K    |
| `internal/provider/smb_client_policy_rule_resource_test.go`     | Rule unit tests                               | VERIFIED   | 2 test functions: CRUD, Import; 11.9K                |

### Plan 08-02 — Syslog Server

| Artifact                                                    | Provides                              | Status   | Details                                              |
|-------------------------------------------------------------|---------------------------------------|----------|------------------------------------------------------|
| `internal/client/models.go`                                 | SyslogServer, SyslogServerPost, SyslogServerPatch | VERIFIED | Lines 833–856: all 3 types confirmed      |
| `internal/client/syslog_servers.go`                        | 5 CRUD methods                        | VERIFIED | Get, List, Post, Patch, Delete all present           |
| `internal/testmock/handlers/syslog_servers.go`             | Mock handler — `RegisterSyslogServerHandlers` | VERIFIED | Function present at line 20                    |
| `internal/provider/syslog_server_resource.go`              | `flashblade_syslog_server` resource   | VERIFIED | `NewSyslogServerResource` at line 33; 11.1K          |
| `internal/provider/syslog_server_data_source.go`           | `flashblade_syslog_server` data source | VERIFIED | `NewSyslogServerDataSource` at line 25               |
| `internal/provider/syslog_server_resource_test.go`         | Unit tests                            | VERIFIED | 3 test functions: CRUD, Import, DataSource; 14.3K    |

### Plan 08-03 — Acceptance Tests

| Artifact                                                                    | Provides                              | Status   | Details                                  |
|-----------------------------------------------------------------------------|---------------------------------------|----------|------------------------------------------|
| `tmp/test-purestorage/test_smb_client_policies.tf`                         | HCL for SMB client policy + rule      | VERIFIED | 3 resources/data sources, valid HCL      |
| `tmp/test-purestorage/test_syslog_servers.tf`                              | HCL for syslog server                 | VERIFIED | 1 resource + data source, valid HCL      |
| `tmp/test-purestorage/test_v11_server_exports.tf`                          | HCL for S3 export policy + virtual host | VERIFIED | 3 resources with API-correct values (alphanumeric name, `pure:S3Access` action) |

---

## Key Link Verification

| From                                        | To                                    | Via                                   | Status    | Details                                                   |
|---------------------------------------------|---------------------------------------|---------------------------------------|-----------|-----------------------------------------------------------|
| `smb_client_policy_resource.go`             | `client/smb_client_policies.go`       | Post/Get/Patch/Delete/ListMembers     | WIRED     | All 5 call sites confirmed (lines 161, 192, 264, 300, 313) |
| `smb_client_policy_rule_resource.go`        | `client/smb_client_policies.go`       | Post/GetByName/Patch/Delete           | WIRED     | All 4 call sites confirmed (lines 163, 198, 260, 294)     |
| `provider.go`                               | `smb_client_policy_resource.go`       | `NewSmbClientPolicyResource`          | WIRED     | Registered at line 277                                    |
| `provider.go`                               | `smb_client_policy_rule_resource.go`  | `NewSmbClientPolicyRuleResource`      | WIRED     | Registered at line 278                                    |
| `provider.go`                               | `smb_client_policy_data_source.go`    | `NewSmbClientPolicyDataSource`        | WIRED     | Registered at line 309                                    |
| `syslog_server_resource.go`                 | `client/syslog_servers.go`            | Post/Get/Patch/Delete                 | WIRED     | All 4 call sites confirmed (lines 144, 171, 237, 269)     |
| `syslog_server_resource.go`                 | `listdefault.StaticValue`             | Null-vs-empty on services/sources     | WIRED     | Lines 84 and 91 of `syslog_server_resource.go`            |
| `provider.go`                               | `syslog_server_resource.go`           | `NewSyslogServerResource`             | WIRED     | Registered at line 296                                    |
| `provider.go`                               | `syslog_server_data_source.go`        | `NewSyslogServerDataSource`           | WIRED     | Registered at line 323                                    |
| `test_smb_client_policies.tf`               | `flashblade_smb_client_policy`        | terraform apply vs live FlashBlade    | VERIFIED* | SUMMARY: passed apply+plan(0)+destroy cycle               |
| `test_syslog_servers.tf`                    | `flashblade_syslog_server`            | terraform apply vs live FlashBlade    | VERIFIED* | SUMMARY: passed apply+plan(0)+destroy cycle               |
| `internal/client/s3_export_policies.go`     | S3 rule POST fix                      | `?policy_names=...&names=...`         | WIRED     | `PostS3ExportPolicyRule` at line 151 uses correct params  |
| `object_store_virtual_host_resource.go`     | No `listdefault` on `attached_servers` | API auto-attaches default server     | WIRED     | `listdefault` import and usage absent from file           |

*Live FlashBlade run: cannot verify programmatically.

---

## Requirements Coverage

| Requirement | Source Plan | Description                                                          | Status    | Evidence                                                            |
|-------------|-------------|----------------------------------------------------------------------|-----------|---------------------------------------------------------------------|
| SMC-01      | 08-01       | Operator can create an SMB client policy with enable toggle          | SATISFIED | `smb_client_policy_resource.go` with `Enabled` field; CRUD test passes |
| SMC-02      | 08-01       | Operator can create SMB client policy rules with client/encryption/permission | SATISFIED | `SmbClientPolicyRulePost` struct + rule resource; CRUD test passes  |
| SMC-03      | 08-01       | Operator can update and delete SMB client policy rules independently | SATISFIED | `PatchSmbClientPolicyRule` + `DeleteSmbClientPolicyRule` wired; CRUD test passes |
| SMC-04      | 08-01       | Operator can import SMB client policies and rules into Terraform state | SATISFIED | Import tests pass for both policy (by name) and rule (composite ID `policy_name/rule_name`) |
| SYS-01      | 08-02       | Operator can create a syslog server with URI, services, and sources  | SATISFIED | `syslog_server_resource.go` with `URI`, `Services`, `Sources` fields; CRUD test passes |
| SYS-02      | 08-02       | Operator can update syslog server configuration                      | SATISFIED | `PatchSyslogServer` wired; CRUD test updates URI and services in-place |
| SYS-03      | 08-02       | Operator can import an existing syslog server into Terraform state   | SATISFIED | `TestUnit_SyslogServer_Import` passes; 0-diff plan after import     |
| EXP-03      | 08-03       | All export resources pass acceptance tests against live FlashBlade   | SATISFIED | SUMMARY confirms all 26 resources passed full apply+plan(0)+destroy cycle; 3 API bugs found and fixed inline |

No orphaned requirements — all 8 requirement IDs declared in plan frontmatter are accounted for and satisfied.

---

## Anti-Patterns Found

No TODOs, FIXMEs, placeholder implementations, or empty handler bodies found across the 14 new files.

| File | Pattern | Severity | Impact |
|------|---------|----------|--------|
| `s3_export_policy_rule_resource.go` | Description says "e.g. 's3:GetObject'" for `actions` attribute — stale since only `pure:S3Access` is valid | INFO | Documentation inaccuracy only; no runtime impact |

---

## Human Verification Required

### 1. Acceptance test idempotency

**Test:** From `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage`, run `tofu plan` after a fresh `tofu apply`.
**Expected:** Output shows `No changes. Your infrastructure matches the configuration.`
**Why human:** Requires live FlashBlade array and valid provider credentials.

### 2. Acceptance test cleanup

**Test:** Run `tofu destroy -auto-approve` and confirm all `test-gule-*` resources are removed.
**Expected:** All 26 resources (SMB client policy + rule, syslog server, S3 export policy + rule, virtual host) destroyed without errors.
**Why human:** Requires live FlashBlade array.

Note: The operator confirmed both tests passed during the 08-03 execution cycle (per SUMMARY). The human verification items above document what would be needed to re-confirm from scratch.

---

## Build and Test Results

| Check                               | Result          |
|-------------------------------------|-----------------|
| `go build ./...`                    | SUCCESS         |
| `TestUnit_SmbClientPolic*` (7 tests)| PASS            |
| `TestUnit_SyslogServer*` (3 tests)  | PASS            |
| Full suite `go test ./internal/...` | PASS — 268 tests |

---

## Commits Verified

| Commit    | Plan  | Description                                              |
|-----------|-------|----------------------------------------------------------|
| `6260601` | 08-01 | SMB client policy models, client CRUD, mock handler      |
| `7711c8a` | 08-01 | SMB client policy resource, rule, data source, tests     |
| `bf32258` | 08-02 | Syslog server models, client CRUD, mock handler          |
| `e4c3ad0` | 08-02 | Syslog server resource, data source, tests               |
| `f3894ba` | 08-03 | Acceptance test HCL (in test-purestorage repo)           |
| `fa8cd8a` | 08-03 | S3 rule name fix, pure:S3Access fix, virtual host drift fix (in provider repo) |

---

_Verified: 2026-03-28_
_Verifier: Claude (gsd-verifier)_
