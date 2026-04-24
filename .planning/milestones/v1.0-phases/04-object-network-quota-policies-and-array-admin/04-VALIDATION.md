---
phase: 4
slug: object-network-quota-policies-and-array-admin
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + httptest.NewServer mock |
| **Config file** | none (go test standard) |
| **Quick run command** | `go test ./internal/... -run "TestObjectStoreAccess\|TestNetworkAccess\|TestQuota\|TestArray" -count=1 -timeout 60s` |
| **Full suite command** | `go test ./... -count=1 -timeout 120s` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run quick command
- **After every plan wave:** Run full suite
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-xx-01 | 01 | 1 | OAP-01..05 | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccessPolicy -v` | ❌ W0 | ⬜ pending |
| 04-xx-02 | 01 | 1 | OAR-01..04 | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccessPolicyRule -v` | ❌ W0 | ⬜ pending |
| 04-xx-03 | 02 | 2 | NAP-01..05 | unit (mock) | `go test ./internal/provider/ -run TestNetworkAccessPolicy -v` | ❌ W0 | ⬜ pending |
| 04-xx-04 | 02 | 2 | NAR-01..04 | unit (mock) | `go test ./internal/provider/ -run TestNetworkAccessPolicyRule -v` | ❌ W0 | ⬜ pending |
| 04-xx-05 | 03 | 2 | QTP-01..05 | unit (mock) | `go test ./internal/provider/ -run TestQuota -v` | ❌ W0 | ⬜ pending |
| 04-xx-06 | 03 | 2 | QTR-01..04 | unit (mock) | `go test ./internal/provider/ -run TestQuota -v` | ❌ W0 | ⬜ pending |
| 04-xx-07 | 04 | 3 | ADM-01..05 | unit (mock) | `go test ./internal/provider/ -run "TestArrayDns\|TestArrayNtp\|TestArraySmtp" -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] All client method files for OAP, NAP, quota, DNS, NTP, SMTP, alert-watchers
- [ ] All mock handler files
- [ ] All resource, data source, and test files
- [ ] Models in `internal/client/models.go`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| OAP IAM conditions round-trip | OAR-01 | JSON condition parsing needs real API validation | Create OAP rule with conditions, verify `plan` shows 0 diff |
| NAP singleton behavior on real array | NAP-01 | Singleton GET+PATCH needs real array confirmation | Configure NAP, verify `apply → plan → 0 changes` |
| DNS/NTP/SMTP reset on destroy | ADM-01..03 | Default state varies per array | Destroy DNS resource, verify nameservers cleared |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
