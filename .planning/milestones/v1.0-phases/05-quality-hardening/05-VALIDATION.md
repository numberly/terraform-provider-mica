---
phase: 5
slug: quality-hardening
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-28
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing stdlib |
| **Config file** | none |
| **Quick run command** | `go test ./internal/... -count=1 -timeout 5m` |
| **Full suite command** | `go test ./internal/... -count=1 -race -timeout 10m` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -count=1 -timeout 5m`
- **After every plan wave:** Run `go test ./internal/... -count=1 -race -timeout 10m`
- **Before `/gsd:verify-work`:** Full suite green + `make generate` produces docs/
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-xx-01 | 01 | 1 | QUA-01 | unit | `go test ./internal/provider/... -run TestUnit.*PlanModifier -count=1` | ❌ W0 | ⬜ pending |
| 05-xx-02 | 01 | 1 | QUA-02 | unit | `go test ./internal/provider/... -run TestUnit.*Validator -count=1` | ❌ W0 | ⬜ pending |
| 05-xx-03 | 01 | 1 | QUA-03 | unit | `go test ./internal/provider/... -count=1` | Partial | ⬜ pending |
| 05-xx-04 | 02 | 1 | QUA-04 | unit | `go test ./internal/... -count=1` | Partial | ⬜ pending |
| 05-xx-05 | 02 | 1 | QUA-05 | unit | `go test ./internal/client/... -run TestUnit_RetryTransport -count=1` | ✅ | ✅ green |
| 05-xx-06 | 03 | 2 | QUA-06 | generate | `make generate && ls docs/` | ❌ W0 | ⬜ pending |

---

## Wave 0 Requirements

- [ ] Plan modifier test functions in existing `*_resource_test.go` files
- [ ] Validator test functions (direct validator calls)
- [ ] `IsConflict` and `IsUnprocessable` helpers in `errors.go`
- [ ] Auto-pagination in client list methods
- [ ] `examples/resources/*/resource.tf` for all resource types
- [ ] `examples/data-sources/*/data-source.tf` for all data source types
- [ ] `.github/workflows/ci.yml`
- [ ] `README.md`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| terraform-plugin-docs output quality | QUA-06 | Generated docs need human review for readability | Run `make generate`, review docs/ output |
| GitHub Actions workflow validity | QUA-06 | Needs actual GitHub push to validate | Push branch, check Actions tab |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
