---
phase: 01-foundation
verified: 2026-03-26T00:00:00Z
status: passed
score: 21/21 must-haves verified
re_verification: false
---

# Phase 1: Foundation Verification Report

**Phase Goal:** Operators can configure the provider and manage file systems via Terraform with full CRUD, import, and drift detection — all shared infrastructure patterns established for replication
**Verified:** 2026-03-26
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (from ROADMAP.md Success Criteria + PLAN must_haves)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Provider connects to FlashBlade via API token or OAuth2, respecting env var fallbacks and custom CA | VERIFIED | provider.go:141-173 reads FLASHBLADE_HOST/API_TOKEN/OAUTH2_* via os.Getenv; auth.go implements FlashBladeTokenSource + LoginWithAPIToken; transport.go builds TLS transport with custom CA pool |
| 2 | FlashBladeClient can authenticate via OAuth2 token exchange (non-standard grant type) | VERIFIED | auth.go:65-131 implements FlashBladeTokenSource with urn:ietf:params:oauth:grant-type:token-exchange grant; token caching with sync.Mutex present |
| 3 | FlashBladeClient negotiates API version v2.22 on startup | VERIFIED | client.go:20 `const APIVersion = "2.22"`; client.go:164 NegotiateVersion checks versions array; provider.go:248 calls NegotiateVersion during Configure |
| 4 | FlashBladeClient retries transient errors (429, 503, 5xx) with exponential backoff | VERIFIED | transport.go:34-91 RoundTrip checks IsRetryable; transport.go:92 computeDelay with exponential growth; errors.go:66 IsRetryable returns true for 429/503/5xx>=500; 26 client tests pass |
| 5 | Custom CA certificate is loaded into TLS transport | VERIFIED | client.go:122-163 buildTransport handles CACertFile (file read) and CACert (inline PEM) into x509.CertPool |
| 6 | Provider schema accepts endpoint, auth block (api_token or oauth2), TLS settings, and retry config | VERIFIED | provider.go:65-129 Schema defines all attributes; SingleNestedAttribute for auth block with oauth2 sub-block |
| 7 | Provider marks api_token and oauth2 fields as Sensitive | VERIFIED | provider.go:106-113 Sensitive: true on api_token, client_id, key_id |
| 8 | Provider injects FlashBladeClient into ResourceData and DataSourceData | VERIFIED | provider.go:260-263 sets resp.ResourceData and resp.DataSourceData to *FlashBladeClient |
| 9 | Provider Configure calls NegotiateVersion and returns diagnostic error on v2.22 mismatch | VERIFIED | provider.go:248 NegotiateVersion call with diagnostic on error |
| 10 | Client can create/read/update/soft-delete/eradicate file systems | VERIFIED | filesystems.go implements GetFileSystem, PostFileSystem, PatchFileSystem, DeleteFileSystem, PollUntilEradicated; all TestUnit_FileSystem_* pass |
| 11 | Mock server simulates full CRUD lifecycle including soft-delete | VERIFIED | testmock/handlers/filesystems.go RegisterFileSystemHandlers with in-memory store; TestUnit_MockServer_FullCRUDLifecycle passes |
| 12 | User can create a file system with name, provisioned size, NFS/SMB config | VERIFIED | filesystem_resource.go:355-395 Create calls PostFileSystem with FileSystemPost; schema includes nfs/smb SingleNestedBlock |
| 13 | User can update file system attributes in place including rename | VERIFIED | filesystem_resource.go:460-510 Update builds FileSystemPatch with only changed fields; no RequiresReplace on name attribute |
| 14 | User can destroy with two-phase soft-delete + eradicate | VERIFIED | filesystem_resource.go:520-558 Phase 1 PatchFileSystem(destroyed=true), Phase 2 DeleteFileSystem, Phase 3 PollUntilEradicated; destroy_eradicate_on_delete defaults true |
| 15 | User can read file system state including all computed attributes | VERIFIED | filesystem_resource.go:590-650 readIntoState maps space, created, nfs, smb, http, promotion_status, writable |
| 16 | User can import an existing file system by name | VERIFIED | filesystem_resource.go:560-588 ImportState accepts name, calls GetFileSystem, populates full state |
| 17 | Data source returns file system attributes by name | VERIFIED | filesystem_data_source.go:298 GetFileSystem by name; all attributes Computed except name (Required) |
| 18 | Drift detection logs field-level diffs via tflog | VERIFIED | filesystem_resource.go:430 tflog.Info("drift detected on file system") with structured fields; TestUnit_FileSystem_DriftLog passes |
| 19 | Per-resource timeouts configurable for create/read/update/delete | VERIFIED | filesystem_resource.go:8 imports terraform-plugin-framework-timeouts; timeouts.Attributes(ctx, Opts{Create,Read,Update,Delete:true}) with 20m/5m/20m/30m defaults |
| 20 | terraform plan after apply shows 0 changes (idempotency) | VERIFIED | TestUnit_FileSystem_Idempotent passes; Read-at-end-of-write pattern used in Create and Update |
| 21 | All tests pass; build compiles; go vet clean | VERIFIED | `go test ./internal/... -run TestUnit`: 51 tests pass (26 client, 3 testmock, 22 provider); `go build ./...`: success; `go vet ./...`: clean |

**Score:** 21/21 truths verified

### Required Artifacts

| Artifact | Min Lines | Actual Lines | Status | Key Exports |
|----------|-----------|--------------|--------|-------------|
| `internal/client/client.go` | — | 282 | VERIFIED | FlashBladeClient, NewClient, Config, APIVersion, NegotiateVersion |
| `internal/client/auth.go` | — | 144 | VERIFIED | FlashBladeTokenSource, LoginWithAPIToken, NewFlashBladeTokenSource |
| `internal/client/transport.go` | — | 99 | VERIFIED | retryTransport (unexported, used via buildTransport) |
| `internal/client/errors.go` | — | 68 | VERIFIED | APIError, IsNotFound, IsRetryable, ParseAPIError |
| `internal/client/models.go` | — | 103 | VERIFIED | FileSystem, FileSystemPost, FileSystemPatch, Space, ListResponse, VersionResponse |
| `internal/client/filesystems.go` | — | 4.4K | VERIFIED | GetFileSystem, ListFileSystems, PostFileSystem, PatchFileSystem, DeleteFileSystem, PollUntilEradicated |
| `internal/provider/provider.go` | 150 | 278 | VERIFIED | New, FlashBladeProvider, env var fallbacks, sensitive fields |
| `internal/provider/filesystem_resource.go` | 300 | 712 | VERIFIED | NewFilesystemResource, full CRUD, import, soft-delete, drift, timeouts |
| `internal/provider/filesystem_data_source.go` | — | 13.4K (file) | VERIFIED | NewFilesystemDataSource |
| `internal/testmock/server.go` | — | 1.9K | VERIFIED | NewMockServer, RegisterHandler |
| `internal/testmock/handlers/filesystems.go` | — | 7.3K | VERIFIED | RegisterFileSystemHandlers |
| `internal/provider/filesystem_resource_test.go` | — | 27.4K | VERIFIED | 13 TestUnit_FileSystem_* tests |
| `internal/testmock/server_test.go` | — | 8.0K | VERIFIED | TestUnit_MockServer_FullCRUDLifecycle |
| `examples/resources/flashblade_file_system/resource.tf` | — | present | VERIFIED | HCL example with nfs block and timeouts |
| `examples/data-sources/flashblade_file_system/data-source.tf` | — | present | VERIFIED | HCL data source example |
| `examples/provider/provider.tf` | — | present | VERIFIED | Provider block example |

### Key Link Verification

| From | To | Via | Status | Evidence |
|------|----|-----|--------|----------|
| `internal/client/client.go` | `internal/client/transport.go` | buildTransport returns RoundTripper | WIRED | client.go:79 `buildTransport(cfg)` call present |
| `internal/client/client.go` | `internal/client/auth.go` | LoginWithAPIToken / NewFlashBladeTokenSource | WIRED | client.go:96 LoginWithAPIToken, client.go:106 NewFlashBladeTokenSource |
| `internal/client/transport.go` | `internal/client/errors.go` | retryTransport uses IsRetryable | WIRED | transport.go:67 `IsRetryable(resp.StatusCode)` |
| `internal/provider/provider.go` | `internal/client/client.go` | Configure creates FlashBladeClient via client.NewClient | WIRED | provider.go:225 client.Config{}, provider.go:238 client.NewClient(cfg) |
| `main.go` | `internal/provider/provider.go` | providerserver.Serve with provider.New() | WIRED | main.go:21-26 providerserver.ServeOpts + provider.New(version) |
| `internal/client/filesystems.go` | `internal/client/client.go` | CRUD methods use get/post/patch/delete helpers | WIRED | filesystems.go uses c.get, c.post, c.patch, c.delete |
| `internal/client/filesystems.go` | `internal/client/models.go` | Uses FileSystem, FileSystemPost, FileSystemPatch, ListResponse | WIRED | All types referenced throughout filesystems.go |
| `internal/provider/filesystem_resource.go` | `internal/client/filesystems.go` | CRUD calls PostFileSystem, GetFileSystem, PatchFileSystem, DeleteFileSystem, PollUntilEradicated | WIRED | filesystem_resource.go:386/417/497/534/547/553 confirmed |
| `internal/provider/filesystem_resource.go` | `internal/provider/provider.go` | Registered in Resources() | WIRED | provider.go:269 NewFilesystemResource |
| `internal/provider/filesystem_data_source.go` | `internal/client/filesystems.go` | Read calls GetFileSystem | WIRED | filesystem_data_source.go:298 d.client.GetFileSystem |
| `internal/provider/provider.go` | `internal/provider/filesystem_data_source.go` | Registered in DataSources() | WIRED | provider.go:276 NewFilesystemDataSource |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| PROV-01 | 01-01, 01-02 | Provider accepts endpoint URL, API token, TLS CA cert | SATISFIED | provider.go Schema includes endpoint, ca_cert_file, ca_cert, auth.api_token |
| PROV-02 | 01-01, 01-02 | Provider accepts OAuth2 client_id, key_id, issuer | SATISFIED | provider.go oauth2Model; auth.go FlashBladeTokenSource |
| PROV-03 | 01-02 | Env var fallbacks for FLASHBLADE_* | SATISFIED | provider.go:141-173 os.Getenv for all 5 FLASHBLADE_* vars |
| PROV-04 | 01-01, 01-02 | API version negotiation v2.22 on startup | SATISFIED | client.go NegotiateVersion; provider.go:248 calls on Configure |
| PROV-05 | 01-01, 01-02 | api_token and oauth2 secrets marked Sensitive | SATISFIED | provider.go:106-113 Sensitive: true on api_token, client_id, key_id |
| PROV-06 | 01-02, 01-04 | Structured tflog output for all operations | SATISFIED | provider.go Configure emits tflog.Info; filesystem_resource.go emits drift tflog |
| PROV-07 | 01-01, 01-02 | Custom CA cert (file or inline) | SATISFIED | client.go:122-163 buildTransport handles both CACertFile and CACert |
| FS-01 | 01-03, 01-04 | Create file system with name, provisioned, optional policies | SATISFIED | filesystem_resource.go Create + PostFileSystem; nfs_export_policy, smb_share_policy attributes |
| FS-02 | 01-03, 01-04 | Update file system attributes (size, policies, NFS, SMB, rename) | SATISFIED | filesystem_resource.go Update + PatchFileSystem with diff-only patch |
| FS-03 | 01-03, 01-04 | Destroy with two-phase soft-delete + eradicate | SATISFIED | filesystem_resource.go Delete: PATCH destroyed=true, DeleteFileSystem, PollUntilEradicated |
| FS-04 | 01-03, 01-04 | Read all computed attributes (space, created timestamp) | SATISFIED | readIntoState maps all FileSystem fields including space, created |
| FS-05 | 01-04 | Import existing file system by name | SATISFIED | ImportState calls GetFileSystem, populates full state |
| FS-06 | 01-04 | Data source returns attributes by name | SATISFIED | filesystem_data_source.go Read + GetFileSystem |
| FS-07 | 01-04 | Drift detection via tflog | SATISFIED | filesystem_resource.go:430 tflog.Info with field/state_value/api_value; TestUnit_FileSystem_DriftLog |

No orphaned requirements found — all 14 requirement IDs from PLAN frontmatter are accounted for.

### Anti-Patterns Found

None. Scan of all `internal/` files (excluding test files) found zero TODO/FIXME/HACK/PLACEHOLDER comments, no stub return values (return nil, return {}, return []) in non-test code paths, no empty handler implementations.

Notable architectural decision: `internal/client/` package has zero `terraform-plugin-framework` imports (confirmed by grep returning 0 matches) — client layer is pure Go, testable with httptest.NewServer without Terraform testing framework overhead.

### Human Verification Required

The following items cannot be verified programmatically and require a live FlashBlade endpoint or manual inspection:

1. **Provider connects to real FlashBlade endpoint**
   - Test: Configure the provider with a valid endpoint and API token, run `terraform plan`
   - Expected: Provider connects, negotiates v2.22, outputs no errors
   - Why human: Requires live FlashBlade hardware/simulator

2. **Two-phase delete prevents name collision on re-creation**
   - Test: Apply a file system, destroy it, immediately re-apply the same name
   - Expected: No conflict error — PollUntilEradicated ensures full removal before returning
   - Why human: Timing-dependent behavior only observable against a real API

3. **tflog drift output visible in terraform refresh output**
   - Test: Manually change a file system's provisioned size outside Terraform, run `terraform refresh`
   - Expected: Structured log line with field, state_value, api_value visible in TF_LOG=INFO output
   - Why human: Log capture in unit tests uses a mock; real Terraform log routing needs live validation

4. **Import produces 0-diff plan on real resource**
   - Test: Import an existing file system, run `terraform plan`
   - Expected: No planned changes
   - Why human: Real computed fields (space.*, promotion_status, time_remaining) may have values not covered by unit test mock

---

## Summary

Phase 1 goal is fully achieved. All 21 observable truths are verified in the codebase. Every required artifact exists, is substantive (no stubs), and is wired into the appropriate call chain. All 14 requirement IDs from the four PLAN frontmatter sections are satisfied with concrete implementation evidence.

Key implementation quality signals:
- 51 unit tests pass across 3 packages (26 client, 3 testmock, 22 provider)
- Full binary compiles (`go build ./...`)
- `go vet ./...` clean
- client layer has zero terraform-plugin-framework imports (clean separation)
- No anti-patterns (TODO/FIXME/stubs) found in production code paths
- All shared patterns (Read-at-end-of-write, soft-delete, drift detection, per-resource timeouts, import by name) established in `filesystem_resource.go` as the template for Phase 2+ resources

---

_Verified: 2026-03-26_
_Verifier: Claude (gsd-verifier)_
