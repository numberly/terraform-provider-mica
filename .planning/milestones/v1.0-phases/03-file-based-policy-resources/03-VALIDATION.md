---
phase: 3
slug: file-based-policy-resources
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard testing + terraform-plugin-framework test helpers |
| **Config file** | none (go test standard — framework from Phase 1) |
| **Quick run command** | `go test ./internal/... -run "TestNfs\|TestSmb\|TestSnapshot" -count=1 -timeout 60s` |
| **Full suite command** | `go test ./... -count=1 -timeout 120s` |
| **Estimated runtime** | ~25 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -run "TestNfs\|TestSmb\|TestSnapshot" -count=1 -timeout 60s`
- **After every plan wave:** Run `go test ./... -count=1 -timeout 120s`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 25 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | NFP-01 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Create -v` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | NFP-02 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Update -v` | ❌ W0 | ⬜ pending |
| 03-01-03 | 01 | 1 | NFP-03 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Delete -v` | ❌ W0 | ⬜ pending |
| 03-01-04 | 01 | 1 | NFP-04 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyResource_Import -v` | ❌ W0 | ⬜ pending |
| 03-01-05 | 01 | 1 | NFP-05 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyDataSource -v` | ❌ W0 | ⬜ pending |
| 03-01-06 | 01 | 1 | NFR-01 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Create -v` | ❌ W0 | ⬜ pending |
| 03-01-07 | 01 | 1 | NFR-02 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Update -v` | ❌ W0 | ⬜ pending |
| 03-01-08 | 01 | 1 | NFR-03 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Delete -v` | ❌ W0 | ⬜ pending |
| 03-01-09 | 01 | 1 | NFR-04 | unit (mock) | `go test ./internal/provider/ -run TestNfsExportPolicyRuleResource_Import -v` | ❌ W0 | ⬜ pending |
| 03-02-01 | 02 | 2 | SMP-01..05 | unit (mock) | `go test ./internal/provider/ -run TestSmbSharePolicy -v` | ❌ W0 | ⬜ pending |
| 03-02-02 | 02 | 2 | SMR-01..04 | unit (mock) | `go test ./internal/provider/ -run TestSmbSharePolicyRule -v` | ❌ W0 | ⬜ pending |
| 03-03-01 | 03 | 2 | SNP-01..05 | unit (mock) | `go test ./internal/provider/ -run TestSnapshotPolicy -v` | ❌ W0 | ⬜ pending |
| 03-03-02 | 03 | 2 | SNR-01..04 | unit (mock) | `go test ./internal/provider/ -run TestSnapshotPolicyRule -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/client/nfs_export_policies.go` — NFS policy + rule client methods
- [ ] `internal/client/smb_share_policies.go` — SMB policy + rule client methods
- [ ] `internal/client/snapshot_policies.go` — Snapshot policy + rule client methods
- [ ] `internal/testmock/handlers/nfs_export_policies.go` — NFS mock handler
- [ ] `internal/testmock/handlers/smb_share_policies.go` — SMB mock handler
- [ ] `internal/testmock/handlers/snapshot_policies.go` — Snapshot mock handler
- [ ] Models in `internal/client/models.go` — Phase 3 structs
- [ ] All 12 resource/data source files + test files

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| NFS rule ordering preserved after apply | NFR-01 | Index assigned server-side, need real array to verify | Create 3 NFS rules, verify indexes match insertion order |
| Snapshot rule add/remove via PATCH | SNR-01 | Opaque `add_rules` body needs real API validation | Add snapshot rule, verify schedule appears in policy |
| Policy attachment guard on real array | NFP-03, SMP-03, SNP-03 | Attachment lookup endpoint needs live validation | Attach policy to FS, try to delete policy |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 25s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
