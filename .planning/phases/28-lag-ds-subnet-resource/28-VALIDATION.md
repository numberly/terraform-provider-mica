---
phase: 28
slug: lag-ds-subnet-resource
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-31
---

# Phase 28 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing test infrastructure |
| **Quick run command** | `go test ./internal/provider/ -run TestSubnet -count=1 -v` |
| **Full suite command** | `go test ./internal/... -count=1 -v` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/provider/ -run TestSubnet -count=1 -v`
- **After every plan wave:** Run `go test ./internal/... -count=1 -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 28-01-01 | 01 | 1 | LAG-01, SUB-01 | unit | `go build ./internal/client/` | ❌ W0 | ⬜ pending |
| 28-01-02 | 01 | 1 | LAG-01, SUB-01..06 | unit | `go test ./internal/testmock/... -count=1` | ❌ W0 | ⬜ pending |
| 28-02-01 | 02 | 2 | LAG-01 | unit | `go test ./internal/provider/ -run TestLag -count=1 -v` | ❌ W0 | ⬜ pending |
| 28-02-02 | 02 | 2 | SUB-01..06 | unit | `go test ./internal/provider/ -run TestSubnet -count=1 -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing test infrastructure covers all phase requirements (mock server, test helpers, provider test framework)

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Subnet CRUD on live FlashBlade | SUB-01..06 | Requires real array | `terraform apply` with subnet config against test FlashBlade |
| LAG read on live FlashBlade | LAG-01 | Requires real array | `terraform plan` with LAG data source against test FlashBlade |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
