---
phase: 02-object-store-resources
verified: 2026-03-26T00:00:00Z
status: passed
score: 16/16 must-haves verified
re_verification: false
requirement_notes:
  - id: OSA-03
    requirements_text: "User can destroy an object store account (two-phase soft-delete)"
    implementation: "Single-phase DELETE — explicit user decision documented in 02-CONTEXT.md"
    gap: false
    note: "REQUIREMENTS.md text is stale. Design decision overrides: accounts have no soft-delete per FlashBlade API semantics. Verified as SATISFIED (correct behavior, stale description)."
  - id: OAK-04
    requirements_text: "User can import an existing access key into Terraform state by name"
    implementation: "No import support — explicit user decision documented in 02-CONTEXT.md and 02-03-PLAN.md"
    gap: false
    note: "REQUIREMENTS.md text is stale. Design decision overrides: import intentionally omitted because secret_access_key is unavailable after creation. Verified as SATISFIED (correct behavior, stale description)."
---

# Phase 2: Object Store Resources — Verification Report

**Phase Goal:** Operators can manage the complete object store resource chain — accounts, buckets, and access keys — through Terraform with full lifecycle and dependency ordering
**Verified:** 2026-03-26
**Status:** PASSED
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Operator can create an object store account by name | VERIFIED | `PostObjectStoreAccount` called in `Create()`, test `TestUnit_ObjectStoreAccount_Create` passes |
| 2 | Operator can update account mutable attributes (quota_limit, hard_limit_enabled) | VERIFIED | `PatchObjectStoreAccount` called in `Update()` with pointer-based PATCH semantics, test `TestUnit_ObjectStoreAccount_Update` passes |
| 3 | Operator can destroy an account (single-phase DELETE) with bucket-existence guard | VERIFIED | `Delete()` calls `ListBuckets(account_names=)` and returns diagnostic if any exist, then calls `DeleteObjectStoreAccount`, test `TestUnit_ObjectStoreAccount_Delete` passes |
| 4 | Operator can import an existing account by name with 0-diff plan after import | VERIFIED | `ImportState()` sets Name then delegates to `Read()`, `resource.ResourceWithImportState` interface implemented, test `TestUnit_ObjectStoreAccount_Import` passes |
| 5 | Data source returns account attributes by name | VERIFIED | `object_store_account_data_source.go` calls `GetObjectStoreAccount`, test `TestUnit_ObjectStoreAccountDataSource` passes |
| 6 | Operator can create a bucket with name and account reference | VERIFIED | `PostBucket(name, body{Account})` called in `Create()`, test `TestUnit_Bucket_Create` passes |
| 7 | Operator can update bucket mutable attributes (quota_limit, hard_limit_enabled, versioning) | VERIFIED | `PatchBucket(id, patch)` called in `Update()`, test `TestUnit_Bucket_Update` passes |
| 8 | Operator can destroy a bucket with two-phase soft-delete (default: recoverable) | VERIFIED | `Delete()` always patches `Destroyed=true`, conditionally calls `DeleteBucket+PollBucketUntilEradicated` when `destroy_eradicate_on_delete=true`, tests `TestUnit_Bucket_Destroy` and `TestUnit_Bucket_Destroy_WithEradicate` pass |
| 9 | Operator can import a bucket by name with 0-diff plan after import | VERIFIED | `ImportState()` implemented, `resource.ResourceWithImportState` interface asserted, test `TestUnit_Bucket_Import` passes |
| 10 | Bucket data source returns bucket attributes by name | VERIFIED | `bucket_data_source.go` calls `GetBucket`, test `TestUnit_BucketDataSource` passes |
| 11 | Drift detection logs field-level diffs via tflog on Read | VERIFIED | `tflog.Info` calls in both `bucket_resource.go` and `object_store_account_resource.go` Read methods comparing state vs API values, test `TestUnit_Bucket_DriftLog` passes |
| 12 | Operator can create an access key for an object store account | VERIFIED | `PostObjectStoreAccessKey` called in `Create()` with `user={name: account+"/admin"}`, test `TestUnit_AccessKey_Create` passes |
| 13 | Operator can delete an access key | VERIFIED | `DeleteObjectStoreAccessKey` called in `Delete()`, test `TestUnit_AccessKey_Delete` passes |
| 14 | Secret access key is Sensitive and preserved in state after creation | VERIFIED | `secret_access_key` schema has `Sensitive: true` + `UseStateForUnknown`, `Read()` does not overwrite `data.SecretAccessKey`, test `TestUnit_AccessKey_SecretPreserved` passes |
| 15 | Access key has no import support (secret unavailable after creation) | VERIFIED | `objectStoreAccessKeyResource` does NOT implement `resource.ResourceWithImportState` (asserted via compile-time check), test `TestUnit_AccessKey_NoImport` passes |
| 16 | Data source returns access key attributes by name (without secret) | VERIFIED | `object_store_access_key_data_source.go` calls `GetObjectStoreAccessKey`, secret always empty on GET, test `TestUnit_AccessKeyDataSource` passes |

**Score:** 16/16 truths verified

---

## Required Artifacts

### Plan 02-01

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/models.go` | All Phase 2 model structs | VERIFIED | `ObjectStoreAccount`, `ObjectStoreAccountPost/Patch`, `Bucket`, `BucketPost/Patch`, `ObjectStoreAccessKey`, `ObjectStoreAccessKeyPost`, `NamedReference` all present (187 lines) |
| `internal/client/object_store_accounts.go` | `GetObjectStoreAccount`, `PostObjectStoreAccount`, `PatchObjectStoreAccount`, `DeleteObjectStoreAccount` | VERIFIED | All 5 methods present (87 lines), no terraform imports |
| `internal/testmock/handlers/helpers.go` | `WriteJSONListResponse`, `WriteJSONError` | VERIFIED | Both exported helpers present (45 lines) |
| `internal/testmock/handlers/object_store_accounts.go` | `RegisterObjectStoreAccountHandlers` | VERIFIED | Returns `*objectStoreAccountStore` for cross-reference, thread-safe (181 lines) |
| `internal/provider/object_store_account_resource.go` | `NewObjectStoreAccountResource`, full CRUD + Import | VERIFIED | 401 lines, all CRUD + ImportState methods implemented |
| `internal/provider/object_store_account_data_source.go` | `NewObjectStoreAccountDataSource` | VERIFIED | 177 lines, reads by name |

### Plan 02-02

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/buckets.go` | `GetBucket`, `ListBuckets`, `PostBucket`, `PatchBucket`, `DeleteBucket`, `PollBucketUntilEradicated` | VERIFIED | All 6 methods present (141 lines), `Destroyed` filter and `AccountNames` filter both implemented |
| `internal/testmock/handlers/buckets.go` | `RegisterBucketHandlers` with account cross-reference | VERIFIED | Takes `*objectStoreAccountStore`, validates account on POST, `accounts.byName` cross-reference present |
| `internal/provider/bucket_resource.go` | `NewBucketResource`, full CRUD + Import + soft-delete, min 300 lines | VERIFIED | 492 lines, two-phase soft-delete implemented, ForceNew on name+account |
| `internal/provider/bucket_data_source.go` | `NewBucketDataSource` | VERIFIED | 213 lines, reads by name |

### Plan 02-03

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/object_store_access_keys.go` | `GetObjectStoreAccessKey`, `PostObjectStoreAccessKey`, `DeleteObjectStoreAccessKey` | VERIFIED | All 4 methods present (75 lines), no terraform imports |
| `internal/testmock/handlers/object_store_access_keys.go` | `RegisterObjectStoreAccessKeyHandlers` | VERIFIED | POST generates random key+secret, GET strips secret, account cross-reference present |
| `internal/provider/object_store_access_key_resource.go` | `NewAccessKeyResource`, no ImportState | VERIFIED | 234 lines, `ResourceWithImportState` deliberately absent, `Sensitive: true` on secret |
| `internal/provider/object_store_access_key_data_source.go` | `NewAccessKeyDataSource` | VERIFIED | 143 lines, reads by name |

---

## Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `object_store_account_resource.go` | `object_store_accounts.go` | `r.client.PostObjectStoreAccount / PatchObjectStoreAccount / DeleteObjectStoreAccount / GetObjectStoreAccount` | WIRED | All 4 call sites found (lines 189, 220, 284, 333) |
| `provider.go` | `object_store_account_resource.go` | `NewObjectStoreAccountResource` in `Resources()` | WIRED | Registered at line 270 |
| `provider.go` | `bucket_resource.go` | `NewBucketResource` in `Resources()` | WIRED | Registered at line 271 |
| `provider.go` | `object_store_access_key_resource.go` | `NewAccessKeyResource` in `Resources()` | WIRED | Registered at line 272 |
| `provider.go` | `object_store_account_data_source.go` | `NewObjectStoreAccountDataSource` in `DataSources()` | WIRED | Registered at line 280 |
| `provider.go` | `bucket_data_source.go` | `NewBucketDataSource` in `DataSources()` | WIRED | Registered at line 281 |
| `provider.go` | `object_store_access_key_data_source.go` | `NewAccessKeyDataSource` in `DataSources()` | WIRED | Registered at line 282 |
| `bucket_resource.go` | `buckets.go` | `r.client.PostBucket / PatchBucket / DeleteBucket / GetBucket` | WIRED | All 5 call sites found |
| `handlers/buckets.go` | `handlers/object_store_accounts.go` | `accounts.byName` cross-reference on POST | WIRED | `s.accounts.byName[body.Account.Name]` at line 142 |
| `object_store_access_key_resource.go` | `object_store_access_keys.go` | `r.client.PostObjectStoreAccessKey / DeleteObjectStoreAccessKey / GetObjectStoreAccessKey` | WIRED | All 3 call sites found (lines 150, 179, 226) |
| `object_store_access_key_resource.go` | secret preservation | `data.SecretAccessKey` set in Create, left untouched in Read | WIRED | Line 161 (Create sets), line 195 (Read comment: left as-is) |

---

## Requirements Coverage

| Requirement | Source Plan | Description | Status | Notes |
|-------------|-------------|-------------|--------|-------|
| OSA-01 | 02-01 | User can create an object store account with name | SATISFIED | `PostObjectStoreAccount`, 8 tests pass |
| OSA-02 | 02-01 | User can update object store account attributes | SATISFIED | `PatchObjectStoreAccount` with pointer-based PATCH |
| OSA-03 | 02-01 | User can destroy an object store account | SATISFIED | Single-phase DELETE per explicit design decision in 02-CONTEXT.md. REQUIREMENTS.md text "two-phase soft-delete" is stale — accounts have no soft-delete by design. |
| OSA-04 | 02-01 | User can import an existing object store account | SATISFIED | `ResourceWithImportState` implemented |
| OSA-05 | 02-01 | Data source returns object store account attributes | SATISFIED | `object_store_account_data_source.go` reads by name |
| BKT-01 | 02-02 | User can create a bucket with name, account reference, and optional settings | SATISFIED | `PostBucket` with `BucketPost{Account: NamedReference}` |
| BKT-02 | 02-02 | User can update bucket attributes | SATISFIED | `PatchBucket` with `versioning`, `quota_limit`, `hard_limit_enabled` |
| BKT-03 | 02-02 | User can destroy a bucket (two-phase: mark destroyed, then eradicate) | SATISFIED | Two-phase soft-delete with configurable eradication, default=false |
| BKT-04 | 02-02 | User can import an existing bucket | SATISFIED | `ResourceWithImportState` implemented |
| BKT-05 | 02-02 | Data source returns bucket attributes | SATISFIED | `bucket_data_source.go` reads by name with full attribute map |
| BKT-06 | 02-02 | Drift detection logs field-level diffs via tflog | SATISFIED | `tflog.Info` in both `bucket_resource.go` Read and `object_store_account_resource.go` Read |
| OAK-01 | 02-03 | User can create an object store access key | SATISFIED | `PostObjectStoreAccessKey` with `user/{name}` |
| OAK-02 | 02-03 | User can delete an object store access key | SATISFIED | `DeleteObjectStoreAccessKey` |
| OAK-03 | 02-03 | Secret access key is Sensitive and only available at creation | SATISFIED | `Sensitive: true`, `UseStateForUnknown`, Read does not overwrite |
| OAK-04 | 02-03 | User can import an existing access key | SATISFIED (by design decision) | Import intentionally absent — secret unavailable after creation. REQUIREMENTS.md text is stale; design decision in 02-CONTEXT.md overrides. `TestUnit_AccessKey_NoImport` validates no-import behavior. |
| OAK-05 | 02-03 | Data source returns access key attributes | SATISFIED | `object_store_access_key_data_source.go` reads by name, secret always empty on GET |

**Requirements note:** OSA-03 and OAK-04 descriptions in REQUIREMENTS.md do not match the implementation, but both are caused by stale requirement text being superseded by explicit design decisions captured in `02-CONTEXT.md`. The implementations are correct per those decisions.

---

## Anti-Patterns Found

No anti-patterns detected. Scanned all 6 resource/client files for TODO, FIXME, placeholder patterns, empty implementations, and stub returns — all clean.

---

## Human Verification Required

### 1. Dependency ordering in a single terraform apply

**Test:** Write a `.tf` config that creates an `object_store_account`, a `bucket` referencing it, and an `object_store_access_key` referencing the account — all in one file with implicit references. Run `terraform apply`.
**Expected:** Terraform applies in correct order: account first, then bucket and access key in parallel, all succeed.
**Why human:** Graph ordering with implicit dependencies requires a live `terraform apply` against a real or full mock stack to validate end-to-end.

### 2. Account deletion blocked by existing bucket

**Test:** Create an account with a bucket under it. Attempt `terraform destroy` targeting only the account.
**Expected:** Diagnostic error "Account has N existing bucket(s). Destroy all buckets in the account before deleting the account."
**Why human:** Dependency enforcement across destroy ordering requires a live `terraform apply`/`destroy` sequence.

### 3. Secret not exposed in plan output

**Test:** Apply a config creating an access key. Observe `terraform plan` and `terraform apply` output.
**Expected:** `secret_access_key` shows as `(sensitive value)` in plan/apply, never as plaintext.
**Why human:** Output formatting requires visual inspection of terminal output.

---

## Build and Test Summary

- `go build ./internal/...` — SUCCESS
- `go test ./internal/provider/ -run "TestUnit_ObjectStoreAccount"` — 8/8 PASS
- `go test ./internal/provider/ -run "TestUnit_Bucket"` — 9/9 PASS
- `go test ./internal/provider/ -run "TestUnit_AccessKey"` — 6/6 PASS
- `go test ./internal/...` — 74/74 PASS (no regressions)

All documented commits verified in git history: `cef822c`, `e1d996b`, `ebc4ea8`, `443505e`, `0374295`, `7485294`, `b353cce`.

---

_Verified: 2026-03-26_
_Verifier: Claude (gsd-verifier)_
