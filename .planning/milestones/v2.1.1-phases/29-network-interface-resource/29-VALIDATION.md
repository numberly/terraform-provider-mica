---
phase: 29
slug: network-interface-resource
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-31
---

# Phase 29 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing test infrastructure |
| **Quick run command** | `go test ./internal/provider/ -run TestNetworkInterface -count=1 -v` |
| **Full suite command** | `go test ./internal/... -count=1 -v` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/provider/ -run TestNetworkInterface -count=1 -v`
- **After every plan wave:** Run `go test ./internal/... -count=1 -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 29-01-01 | 01 | 1 | NI-01..06 | unit | `go build ./internal/client/` | ❌ W0 | ⬜ pending |
| 29-01-02 | 01 | 1 | NI-01..06 | unit | `go build ./internal/testmock/...` | ❌ W0 | ⬜ pending |
| 29-02-01 | 02 | 2 | NI-01..10 | unit | `go test ./internal/provider/ -run TestNetworkInterface -count=1 -v` | ❌ W0 | ⬜ pending |
| 29-02-02 | 02 | 2 | NI-07,NI-10 | unit | `go test ./internal/provider/ -run TestNetworkInterfaceDataSource -count=1 -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing test infrastructure covers all phase requirements (mock server, test helpers, provider test framework)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| VIP CRUD on live FlashBlade | NI-01..03 | Requires real array | `terraform apply` with VIP config |
| Service/server validation on real API | NI-05,NI-06 | API validation behavior | Create VIP with each service type |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
