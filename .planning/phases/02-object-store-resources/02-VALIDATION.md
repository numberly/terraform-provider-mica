---
phase: 2
slug: object-store-resources
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go test (stdlib) + terraform-plugin-testing v1.15.0 |
| **Config file** | none (go test native — framework installed in Phase 1) |
| **Quick run command** | `go test ./internal/... -run TestUnit -count=1 -timeout 60s` |
| **Full suite command** | `go test ./internal/... -count=1 -timeout 300s` |
| **Estimated runtime** | ~20 seconds (unit only, no acceptance) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/... -run TestUnit -count=1 -timeout 60s`
- **After every plan wave:** Run `go test ./internal/... -count=1 -timeout 300s`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 20 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | OSA-01 | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Create -count=1` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 1 | OSA-02 | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Update -count=1` | ❌ W0 | ⬜ pending |
| 02-01-03 | 01 | 1 | OSA-03 | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Delete -count=1` | ❌ W0 | ⬜ pending |
| 02-01-04 | 01 | 1 | OSA-04 | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccount_Import -count=1` | ❌ W0 | ⬜ pending |
| 02-01-05 | 01 | 1 | OSA-05 | unit (mock) | `go test ./internal/provider/ -run TestObjectStoreAccountDataSource -count=1` | ❌ W0 | ⬜ pending |
| 02-02-01 | 02 | 1 | BKT-01 | unit (mock) | `go test ./internal/provider/ -run TestBucket_Create -count=1` | ❌ W0 | ⬜ pending |
| 02-02-02 | 02 | 1 | BKT-02 | unit (mock) | `go test ./internal/provider/ -run TestBucket_Update -count=1` | ❌ W0 | ⬜ pending |
| 02-02-03 | 02 | 1 | BKT-03 | unit (mock) | `go test ./internal/provider/ -run TestBucket_Delete -count=1` | ❌ W0 | ⬜ pending |
| 02-02-04 | 02 | 1 | BKT-04 | unit (mock) | `go test ./internal/provider/ -run TestBucket_Import -count=1` | ❌ W0 | ⬜ pending |
| 02-02-05 | 02 | 1 | BKT-05 | unit (mock) | `go test ./internal/provider/ -run TestBucketDataSource -count=1` | ❌ W0 | ⬜ pending |
| 02-02-06 | 02 | 1 | BKT-06 | unit (mock) | `go test ./internal/provider/ -run TestBucket_Drift -count=1` | ❌ W0 | ⬜ pending |
| 02-03-01 | 03 | 2 | OAK-01 | unit (mock) | `go test ./internal/provider/ -run TestAccessKey_Create -count=1` | ❌ W0 | ⬜ pending |
| 02-03-02 | 03 | 2 | OAK-02 | unit (mock) | `go test ./internal/provider/ -run TestAccessKey_Delete -count=1` | ❌ W0 | ⬜ pending |
| 02-03-03 | 03 | 2 | OAK-03 | unit (mock) | `go test ./internal/provider/ -run TestAccessKey_SecretPreserved -count=1` | ❌ W0 | ⬜ pending |
| 02-03-04 | 03 | 2 | OAK-04 | n/a | — (no import by design) | n/a | ⬜ pending |
| 02-03-05 | 03 | 2 | OAK-05 | unit (mock) | `go test ./internal/provider/ -run TestAccessKeyDataSource -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/client/object_store_accounts.go` — account CRUD client methods
- [ ] `internal/client/buckets.go` — bucket CRUD client methods
- [ ] `internal/client/object_store_access_keys.go` — access key client methods
- [ ] `internal/testmock/handlers/object_store_accounts.go` — mock account handlers
- [ ] `internal/testmock/handlers/buckets.go` — mock bucket handlers with account cross-ref
- [ ] `internal/testmock/handlers/object_store_access_keys.go` — mock access key handlers
- [ ] `internal/provider/object_store_account_resource.go` — account resource
- [ ] `internal/provider/object_store_account_data_source.go` — account data source
- [ ] `internal/provider/bucket_resource.go` — bucket resource
- [ ] `internal/provider/bucket_data_source.go` — bucket data source
- [ ] `internal/provider/object_store_access_key_resource.go` — access key resource
- [ ] `internal/provider/object_store_access_key_data_source.go` — access key data source
- [ ] All corresponding `*_test.go` files

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Account → bucket → key single apply | BKT-01, OAK-01 | Dependency ordering needs real Terraform graph execution | Write full HCL with all 3 resources, run `terraform apply`, verify order |
| Access key secret availability | OAK-03 | Secret only at creation, need real apply to verify plan output redaction | Create access key, check `terraform show` output for redacted secret |
| Bucket soft-delete on real array | BKT-03 | Eradication timing needs real FlashBlade | Destroy bucket, verify name reusable |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 20s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
