---
phase: 8
slug: smb-client-policies-syslog-and-acceptance-tests
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-28
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test infrastructure |
| **Quick run command** | `go test ./internal/... -count=1 -timeout 5m` |
| **Full suite command** | `go test ./internal/... -count=1 -timeout 5m -v` |
| **Estimated runtime** | ~15 seconds |

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
| 08-01-01 | 01 | 1 | SMC-01..04 | unit | `go test ./internal/provider/ -run "TestUnit_SmbClientPolic" -count=1` | ❌ W0 | ⬜ pending |
| 08-02-01 | 02 | 1 | SYS-01..03 | unit | `go test ./internal/provider/ -run "TestUnit_SyslogServer" -count=1` | ❌ W0 | ⬜ pending |
| 08-03-01 | 03 | 2 | EXP-03 | acceptance | `terraform apply/destroy` in tmp/test-purestorage/ | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] SMB client policy mock handler in `internal/testmock/handlers/smb_client_policies.go`
- [ ] Syslog server mock handler in `internal/testmock/handlers/syslog_servers.go`

*Existing test infrastructure covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Live FlashBlade acceptance | EXP-03 | Requires real array | terraform apply/destroy cycle in tmp/test-purestorage/ |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
