# Phase 59 Plan 01: Retro Traceability — Evidence Report

**Generated:** 2026-05-20
**Branch:** test/api-upgrade-2.23
**Milestone:** v2.23.0 — FlashBlade API 2.23 Upgrade

## Summary

Total REQ-IDs verified: 20/20

All 20 retro REQ-IDs (API-01..03, WORKLOAD-01..03, RESILIENCY-01..03, SCHEMA-01..08, BRIDGE-01..03) have been traced to concrete artifacts on the branch. No gaps detected.

---

## API — Version Bump (3/3)

| REQ-ID | Evidence command | Result | Status |
|--------|------------------|--------|--------|
| API-01 | `git log test/api-upgrade-2.23 --oneline \| grep -E '9a9715b\|API 2\.23'` + `rg '2\.23' internal/client/ -c` | 9a9715b (version bump); 5 files with 2.23 refs in client/; stale 2.22 removed in later commits | OK |
| API-02 | `git log test/api-upgrade-2.23 --oneline \| grep bfe0ad0` + `rg '2\.22' internal/testmock/ internal/provider/*_test.go` | bfe0ad0 (chore: update stale 2.22 references); 0 remaining 2.22 refs in mock handlers or tests | OK |
| API-03 | `git log test/api-upgrade-2.23 --oneline \| grep 19ea8dc` + `grep 'internal/\*\*/\*\.go' .claude/skills/api-upgrade/scripts/upgrade_version.py` | 19ea8dc (fix: broaden upgrade_version.py to all internal/**/*.go); script scans internal/**/*.go for API version refs | OK |

---

## WORKLOAD — New Resource + Data Source (3/3)

| REQ-ID | Evidence command | Result | Status |
|--------|------------------|--------|--------|
| WORKLOAD-01 | `rg 'func NewWorkloadResource' internal/provider/ -l` + `ls internal/provider/workload_resource.go internal/provider/workload_resource_test.go` + `rg 'TestUnit_WorkloadResource_(Lifecycle\|Import\|DriftDetection)' internal/provider/ -l` | internal/provider/workload_resource.go exists; resource tests in workload_resource_test.go (Lifecycle, Import, DriftDetection present) | OK |
| WORKLOAD-02 | `rg 'func NewWorkloadDataSource' internal/provider/ -l` + `ls internal/provider/workload_data_source.go internal/provider/workload_data_source_test.go` | internal/provider/workload_data_source.go exists; DS tests in workload_data_source_test.go | OK |
| WORKLOAD-03 | `ls internal/client/workloads.go internal/client/workloads_test.go internal/testmock/handlers/workloads.go` + `rg 'TestUnit_Workload_(Get\|Post\|Patch\|Delete)' internal/client/ -c` | All client CRUD files exist; 5 unit tests in workloads_test.go (Get_Found, Get_NotFound, Post, Patch, Delete) | OK |

---

## RESILIENCY — Read-Only Data Sources (3/3)

| REQ-ID | Evidence command | Result | Status |
|--------|------------------|--------|--------|
| RESILIENCY-01 | `rg 'func NewResiliencyGroupDataSource' internal/provider/ -l` | internal/provider/resiliency_group_data_source.go | OK |
| RESILIENCY-02 | `rg 'func NewResiliencyGroupMemberDataSource' internal/provider/ -l` | internal/provider/resiliency_group_member_data_source.go | OK |
| RESILIENCY-03 | `ls internal/testmock/handlers/resiliency_group*.go` + `rg 'TestUnit_ResiliencyGroup' internal/ -l` | resiliency_groups.go, resiliency_group_members.go exist; tests in resiliency_group_data_source_test.go, resiliency_group_member_data_source_test.go, resiliency_groups_test.go, resiliency_group_members_test.go | OK |

---

## SCHEMA — Workload Field + Context (8/8)

| REQ-ID | Evidence command | Result | Status |
|--------|------------------|--------|--------|
| SCHEMA-01 | `rg 'SchemaVersion' internal/provider/file_system_resource.go` + `rg 'TestUnit_FileSystemResource_StateUpgrade_V0toV1' internal/provider/ -l` | file_system_resource_test.go has StateUpgrade test | OK |
| SCHEMA-02 | `rg 'TestUnit_FileSystemExportResource_StateUpgrade' internal/provider/ -l` | file_system_export_resource_test.go has StateUpgrade test | OK |
| SCHEMA-03 | `rg 'TestUnit_NfsExportPolicyResource_StateUpgrade' internal/provider/ -l` | nfs_export_policy_resource_test.go has StateUpgrade test | OK |
| SCHEMA-04 | `rg 'TestUnit_SmbClientPolicyResource_StateUpgrade' internal/provider/ -l` | smb_client_policy_resource_test.go has StateUpgrade test | OK |
| SCHEMA-05 | `rg 'TestUnit_SmbSharePolicyResource_StateUpgrade' internal/provider/ -l` | smb_share_policy_resource_test.go has StateUpgrade test | OK |
| SCHEMA-06 | `rg 'TestUnit_QosPolicyResource_StateUpgrade_V0toV1' internal/provider/ -A 5 \| head -1` | qos_policy_resource_test.go has V0→V1 upgrader test (workload field) | OK |
| SCHEMA-07 | `rg 'TestUnit_QosPolicyResource_StateUpgrade_V1toV2' internal/provider/ -A 5 \| head -1` | qos_policy_resource_test.go has V1→V2 upgrader test (context field for API 2.23) | OK |
| SCHEMA-08 | `git log test/api-upgrade-2.23 --oneline \| grep 5ab258a` + `rg 'type qosPolicyV[0-9]Model' internal/provider/qos_policy_resource.go` | 5ab258a (refactor: align v0→v1 upgrader on canonical chain pattern); naming uses Pattern B (qosPolicyV0Model, qosPolicyV1Model) per CONVENTIONS.md | OK |

---

## BRIDGE — Pulumi Bridge Alignment (3/3)

| REQ-ID | Evidence command | Result | Status |
|--------|------------------|--------|--------|
| BRIDGE-01 | `find pulumi/provider -name 'schema*.json' -o -name '*-metadata.json'` + `git log test/api-upgrade-2.23 --oneline \| grep c65d063` | c65d063 (chore: regenerate schema artefacts for API 2.23 additions); schema.json, schema-embed.json, bridge-metadata.json all exist in pulumi/provider/cmd/pulumi-resource-mica/ | OK |
| BRIDGE-02 | `rg 'TestProviderInfo_' pulumi/provider/ -l` + `git log test/api-upgrade-2.23 --oneline \| grep d216d24` | d216d24 (chore: bump test count + Pulumi bridge expectations for API 2.23); TestProviderInfo tests in pulumi/provider/resources_test.go | OK |
| BRIDGE-03 | `git log test/api-upgrade-2.23 --oneline \| grep -E '907623d\|de2697a'` + `grep -n 'make tfgen' .claude/skills/api-upgrade/SKILL.md` | de2697a (docs: document make tfgen step in Phase 6); SKILL.md lines 384, 388, 398 document tfgen workflow | OK |

---

## Gaps

None. All 19 retro REQ-IDs are present and traceable.

---

**Report generated:** 2026-05-20
**Executor:** Plan 59-01 (retro-traceability)
