---
phase: 6
slug: server-resource-and-export-consolidation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-28
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test infrastructure |
| **Quick run command** | `go test ./internal/... -count=1 -timeout 5m` |
| **Full suite command** | `go test ./internal/... -count=1 -timeout 5m -v` |
| **Estimated runtime** | ~10 seconds |

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
| 06-01-01 | 01 | 1 | SRV-01..05 | unit | `go test ./internal/provider/ -run TestUnit_Server -count=1` | ❌ W0 | ⬜ pending |
| 06-02-01 | 02 | 1 | EXP-01 | unit | `go test ./internal/provider/ -run TestUnit_FileSystemExport -count=1` | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 1 | EXP-02 | unit | `go test ./internal/provider/ -run TestUnit_AccountExport -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Server mock handler (full CRUD) in `internal/testmock/handlers/servers.go`
- [ ] File system export mock handler (full CRUD) in `internal/testmock/handlers/file_system_exports.go`
- [ ] Account export mock handler (full CRUD) in `internal/testmock/handlers/object_store_account_exports.go`

*Existing test infrastructure (go test, httptest) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Cascade delete removes exports | SRV-03 | Requires live FlashBlade with real exports | Create server + export, destroy with cascade, verify exports gone |

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
