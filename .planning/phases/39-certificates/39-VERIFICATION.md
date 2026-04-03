---
phase: 39-certificates
verified: 2026-04-03T00:00:00Z
status: passed
score: 10/10 must-haves verified
re_verification: null
gaps: []
human_verification:
  - test: "terraform import flashblade_certificate.example my-tls-cert"
    expected: "State populated with all non-sensitive attributes; subsequent plan shows 0 diff"
    why_human: "Requires a live FlashBlade array — cannot verify programmatically"
  - test: "terraform plan after out-of-band cert renewal"
    expected: "Plan shows updated issued_by, issued_to, valid_from, valid_to, key_algorithm, key_size, status"
    why_human: "Drift detection exercised in unit tests but live plan output needs human review"
---

# Phase 39: Certificates Verification Report

**Phase Goal:** Operators can import and manage TLS certificates on a FlashBlade through Terraform with full CRUD, import, and drift detection
**Verified:** 2026-04-03
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Client can GET a certificate by name | VERIFIED | `GetCertificate` in `internal/client/certificates.go` uses `getOneByName[Certificate]`; `TestUnit_Certificate_Get_Found` passes |
| 2 | Client can POST a new certificate with PEM body, private key, and optional fields | VERIFIED | `PostCertificate` in `internal/client/certificates.go`; `TestUnit_Certificate_Post` passes and verifies computed fields |
| 3 | Client can PATCH an existing certificate (cert renewal with new PEM + key) | VERIFIED | `PatchCertificate` uses pointer semantics; `TestUnit_Certificate_Patch` passes |
| 4 | Client can DELETE a certificate by name | VERIFIED | `DeleteCertificate` deletes then `TestUnit_Certificate_Delete` confirms subsequent GET returns IsNotFound |
| 5 | Mock handler GET returns HTTP 200 + empty list when name not found | VERIFIED | `handleGet` in `internal/testmock/handlers/certificates.go` lines 55-83: filter miss → `items = []client.Certificate{}` → `WriteJSONListResponse(w, http.StatusOK, items)` |
| 6 | Operator can import a certificate via terraform apply with name, PEM, private key | VERIFIED | `flashblade_certificate` resource Create wired to `PostCertificate`; `TestUnit_CertificateResource_Lifecycle` passes |
| 7 | Operator can update (cert renewal) and destroy via terraform apply/destroy | VERIFIED | Update builds `CertificatePatch` with changed fields; Delete calls `DeleteCertificate`; tested in Lifecycle test |
| 8 | terraform import populates all non-sensitive attributes; private_key and passphrase set to empty string | VERIFIED | `ImportState` sets `data.PrivateKey = types.StringValue("")` and `data.Passphrase = types.StringValue(""`); `TestUnit_CertificateResource_Import` asserts both are `""` and Timeouts.IsNull() |
| 9 | data.flashblade_certificate reads an existing certificate by name | VERIFIED | `certificateDataSource.Read` calls `d.client.GetCertificate`; `TestUnit_CertificateDataSource_Basic` passes |
| 10 | Drift detection logs field-level changes via tflog with resource/field/was/now keys | VERIFIED | 17 `tflog.Debug(ctx, "drift detected", ...)` blocks in Read, each with keys `resource`, `field`, `was`, `now`; `TestUnit_CertificateResource_DriftDetection` passes |

**Score:** 10/10 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/client/models_network.go` | Certificate, CertificatePost, CertificatePatch structs | VERIFIED | All three structs present at lines 91, 117, 127 |
| `internal/client/certificates.go` | GetCertificate, PostCertificate, PatchCertificate, DeleteCertificate | VERIFIED | All four methods implemented, 1.7K file |
| `internal/client/certificates_test.go` | 5 unit tests with TestUnit_ prefix | VERIFIED | 5 tests: Get_Found, Get_NotFound, Post, Patch, Delete — all pass |
| `internal/testmock/handlers/certificates.go` | Mock handler with Seed, empty-list GET | VERIFIED | RegisterCertificateHandlers + Seed + handle dispatch + empty-list GET |
| `internal/provider/certificate_resource.go` | flashblade_certificate resource with CRUD, Import, drift detection | VERIFIED | 19.3K file, all 4 interface assertions, Sensitive on private_key/passphrase |
| `internal/provider/certificate_resource_test.go` | 3 resource tests | VERIFIED | Lifecycle, Import, DriftDetection — all pass |
| `internal/provider/certificate_data_source.go` | flashblade_certificate data source | VERIFIED | certificateDataSource struct, wired to GetCertificate |
| `internal/provider/certificate_data_source_test.go` | 1 data source test | VERIFIED | TestUnit_CertificateDataSource_Basic passes |
| `internal/provider/provider.go` | NewCertificateResource and NewCertificateDataSource registered | VERIFIED | Lines 293 and 333 |
| `examples/resources/flashblade_certificate/resource.tf` | HCL example | VERIFIED | 302B file exists |
| `examples/resources/flashblade_certificate/import.sh` | Import example | VERIFIED | 60B file exists |
| `examples/data-sources/flashblade_certificate/data-source.tf` | Data source example | VERIFIED | 243B file exists |
| `docs/resources/certificate.md` | Generated docs | VERIFIED | 4.3K file, generated by tfplugindocs |
| `docs/data-sources/certificate.md` | Generated docs | VERIFIED | 2.1K file, generated by tfplugindocs |
| `ROADMAP.md` | Certificates entry in Implemented section | VERIFIED | Line 87: `| Certificates | flashblade_certificate | Yes | Done | ...` |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/client/certificates.go` | `internal/client/models_network.go` | Certificate/CertificatePost/CertificatePatch types | WIRED | `getOneByName[Certificate]`, `ListResponse[Certificate]`, `CertificatePost`, `CertificatePatch` all used |
| `internal/client/certificates_test.go` | `internal/testmock/handlers/certificates.go` | mock server handlers | WIRED | `handlers.RegisterCertificateHandlers(mux)` called in `newCertificateServer` |
| `internal/provider/certificate_resource.go` | `internal/client/certificates.go` | r.client.GetCertificate / PostCertificate / PatchCertificate / DeleteCertificate | WIRED | All four client methods called in Create/Read/Update/Delete/ImportState |
| `internal/provider/certificate_data_source.go` | `internal/client/certificates.go` | d.client.GetCertificate | WIRED | Line 178: `cert, err := d.client.GetCertificate(ctx, name)` |
| `internal/provider/provider.go` | `internal/provider/certificate_resource.go` | NewCertificateResource factory | WIRED | Line 293: `NewCertificateResource,` in Resources() slice |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `certificate_resource.go` Create | `cert *client.Certificate` | `r.client.PostCertificate` → mock/real API → `ListResponse[Certificate].Items[0]` | Yes — populated from POST response | FLOWING |
| `certificate_resource.go` Read | `cert *client.Certificate` | `r.client.GetCertificate` → `getOneByName` → API GET → decoded JSON | Yes — populated from GET response | FLOWING |
| `certificate_resource.go` ImportState | `cert *client.Certificate` | `r.client.GetCertificate` → API GET | Yes — all computed fields mapped by `mapCertificateToModel` | FLOWING |
| `certificate_data_source.go` Read | `cert *client.Certificate` | `d.client.GetCertificate` → API GET | Yes — inline mapping to all config fields | FLOWING |

Note: private_key and passphrase are write-only — API never returns them. Resource intentionally preserves them from plan/state. This is documented behavior, not a hollow prop.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| 5 client certificate tests pass | `go test ./internal/client/... -run TestUnit_Certificate -count=1` | 5 passed | PASS |
| 4 provider certificate tests pass | `go test ./internal/provider/... -run TestUnit_Certificate -count=1` | 4 passed | PASS |
| Full test suite (701 tests) | `go test ./... -count=1` | 701 passed, 0 failed | PASS |
| Provider builds | `go build ./...` | Success | PASS |
| Vet clean | `go vet ./...` | 0 issues | PASS |

### Requirements Coverage

Requirements CERT-01 through CERT-05 are phase-local identifiers defined in the plan frontmatter. They are not registered in `.planning/REQUIREMENTS.md` (which covers v2.1.3 and v2.2 milestones). This is not a gap — the certificate feature is a new API coverage increment beyond v2.2 scope, and the plan serves as the authoritative requirement document.

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CERT-01 | 39-01, 39-02 | Certificate model structs (Certificate, CertificatePost, CertificatePatch) | SATISFIED | All three structs in `models_network.go` |
| CERT-02 | 39-01, 39-02 | Client CRUD methods (Get, Post, Patch, Delete) | SATISFIED | `internal/client/certificates.go` — 4 methods |
| CERT-03 | 39-02 | flashblade_certificate Terraform resource with full CRUD, import, drift detection | SATISFIED | `internal/provider/certificate_resource.go` — 575 lines |
| CERT-04 | 39-02 | flashblade_certificate data source reads by name | SATISFIED | `internal/provider/certificate_data_source.go` — 222 lines |
| CERT-05 | 39-02 | Provider registration, HCL examples, generated docs, ROADMAP updated | SATISFIED | provider.go lines 293/333, examples dir, docs dir, ROADMAP line 87 |

### CONVENTIONS.md Compliance

| Rule | Expected | Actual | Status |
|------|----------|--------|--------|
| Test naming uses `TestUnit_` prefix | All test functions start with `TestUnit_` | All 9 tests use `TestUnit_` prefix | PASS |
| `UseStateForUnknown` only on stable computed fields (`id`, `certificate_type`) | Only stable fields get `UseStateForUnknown` | Applied only at schema lines 82 and 101; no occurrence on volatile fields (status, valid_from, valid_to, issued_by, issued_to, key_algorithm, key_size) | PASS |
| Mock handler GET returns empty list (HTTP 200), not 404 | `WriteJSONListResponse(w, 200, [])` on miss | `handleGet` lines 55-83: empty slice + `WriteJSONListResponse(w, http.StatusOK, items)` | PASS |
| `private_key` and `passphrase` are `Sensitive: true` | Both fields have `Sensitive: true` in schema | Resource schema lines 149-156: `Sensitive: true` on both | PASS |
| Minimum tests: ≥4 client, ≥3 resource, ≥1 data source | 4 client + 3 resource + 1 DS = 8 minimum | 5 client + 3 resource + 1 DS = 9 tests | PASS |
| Drift detection uses `tflog.Debug` with `resource/field/was/now` keys | `map[string]any{"resource": ..., "field": ..., "was": ..., "now": ...}` | 17 drift blocks in Read, all with exact 4 keys | PASS |
| Total test count must not decrease from 692 baseline | ≥ 692 | 701 tests | PASS |

### Anti-Patterns Found

No blockers or warnings detected.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None | — | — | — | — |

No TODO/FIXME/placeholder comments found in any certificate files. No empty implementations or hardcoded stubs. No return null or return []/{} patterns in non-test code that flow to user-visible output.

### Human Verification Required

#### 1. Live Import Test

**Test:** On a real FlashBlade: `terraform import flashblade_certificate.example my-tls-cert`
**Expected:** State populated with all non-sensitive fields (id, name, certificate, certificate_type, issued_by, etc.); `private_key` and `passphrase` set to `""`; subsequent `terraform plan` shows 0 diff
**Why human:** Requires live FlashBlade array with an existing certificate

#### 2. Cert Renewal Drift Detection

**Test:** Import a cert, renew it out-of-band on the array, then run `terraform plan`
**Expected:** Plan shows changed fields (issued_by, issued_to, valid_from, valid_to, key_algorithm, key_size); `terraform apply` converges to 0 diff
**Why human:** Requires live array and actual cert renewal; mock unit test covers the logic but not the end-to-end plan output format

### Gaps Summary

No gaps found. All must-haves verified at all levels (exists, substantive, wired, data-flowing). The phase goal — "Operators can import and manage TLS certificates on a FlashBlade through Terraform with full CRUD, import, and drift detection" — is achieved by the implementation.

---

_Verified: 2026-04-03_
_Verifier: Claude (gsd-verifier)_
