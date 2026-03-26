---
phase: 1
slug: foundation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-26
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | terraform-plugin-testing v1.15.0 (acceptance) + `testing` stdlib (unit) |
| **Config file** | none — Wave 0 installs |
| **Quick run command** | `go test ./internal/... -run TestUnit -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~15 seconds (unit only, no acceptance) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -run TestUnit -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | PROV-01 | unit | `go test ./internal/provider/... -run TestUnit_ProviderSchema` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 1 | PROV-02 | unit | `go test ./internal/provider/... -run TestUnit_ProviderSchema` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | PROV-03 | unit | `go test ./internal/client/... -run TestUnit_EnvVarFallback` | ❌ W0 | ⬜ pending |
| 01-01-04 | 01 | 1 | PROV-04 | unit (mock) | `go test ./internal/client/... -run TestUnit_NegotiateVersion` | ❌ W0 | ⬜ pending |
| 01-01-05 | 01 | 1 | PROV-05 | unit | `go test ./internal/provider/... -run TestUnit_SensitiveAttributes` | ❌ W0 | ⬜ pending |
| 01-01-06 | 01 | 1 | PROV-06 | unit (mock) | `go test ./internal/provider/... -run TestUnit_TflogOutput` | ❌ W0 | ⬜ pending |
| 01-01-07 | 01 | 1 | PROV-07 | unit | `go test ./internal/client/... -run TestUnit_CustomCATLS` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | FS-01 | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_Create` | ❌ W0 | ⬜ pending |
| 01-02-02 | 02 | 1 | FS-02 | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_Update` | ❌ W0 | ⬜ pending |
| 01-02-03 | 02 | 1 | FS-03 | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_Destroy` | ❌ W0 | ⬜ pending |
| 01-02-04 | 02 | 1 | FS-04 | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_Read` | ❌ W0 | ⬜ pending |
| 01-02-05 | 02 | 1 | FS-05 | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_Import` | ❌ W0 | ⬜ pending |
| 01-02-06 | 02 | 1 | FS-06 | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystemDataSource` | ❌ W0 | ⬜ pending |
| 01-02-07 | 02 | 1 | FS-07 | unit (mock) | `go test ./internal/provider/... -run TestUnit_FileSystem_DriftLog` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` / `go.sum` — module initialization: `go mod init github.com/soulkyu/terraform-provider-flashblade`
- [ ] `main.go` — provider binary entry point
- [ ] `internal/provider/provider.go` — FlashBladeProvider skeleton
- [ ] `internal/client/client.go` + `auth.go` + `transport.go` + `errors.go` + `models.go` — client layer
- [ ] `internal/client/filesystems.go` — file system CRUD methods
- [ ] `internal/provider/filesystem_resource.go` — flashblade_file_system resource
- [ ] `internal/provider/filesystem_data_source.go` — flashblade_file_system data source
- [ ] `internal/testmock/server.go` + `handlers/filesystems.go` — mock HTTP server
- [ ] `GNUmakefile` — `make build`, `make test`, `make testacc`, `make lint`
- [ ] `.golangci.yml` — linter config (golangci-lint v2 format)
- [ ] `examples/provider/provider.tf` — required for tfplugindocs
- [ ] `examples/resources/flashblade_file_system/resource.tf` — required for tfplugindocs
- [ ] All `*_test.go` files listed in the test map above

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| OAuth2 token exchange against live array | PROV-02 | Non-standard grant type requires real FlashBlade | Configure OAuth2 creds, run `terraform plan`, verify auth succeeds |
| Soft-delete eradication timing | FS-03 | Eradication polling needs real array to confirm timing | Destroy a file system, verify name is reusable immediately |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
