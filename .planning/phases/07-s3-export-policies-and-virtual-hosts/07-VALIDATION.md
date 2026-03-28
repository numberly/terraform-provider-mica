---
phase: 7
slug: s3-export-policies-and-virtual-hosts
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-28
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test infrastructure |
| **Quick run command** | `go test ./internal/... -count=1 -timeout 5m` |
| **Full suite command** | `go test ./internal/... -count=1 -timeout 5m -v` |
| **Estimated runtime** | ~12 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -count=1 -timeout 5m`
- **After every plan wave:** Run `go test ./internal/... -count=1 -timeout 5m -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | S3P-01..04 | unit | `go test ./internal/provider/ -run "TestUnit_S3ExportPolic" -count=1` | ❌ W0 | ⬜ pending |
| 07-02-01 | 02 | 1 | VH-01..03 | unit | `go test ./internal/provider/ -run "TestUnit_VirtualHost" -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] S3 export policy mock handler in `internal/testmock/handlers/s3_export_policies.go`
- [ ] Virtual host mock handler in `internal/testmock/handlers/object_store_virtual_hosts.go`

*Existing test infrastructure (go test, httptest) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| S3 rule IAM-style enforcement | S3P-02 | Requires live S3 traffic | Create policy with deny rule, verify S3 access blocked |

*Acceptance tests deferred to Phase 8 (EXP-03).*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
