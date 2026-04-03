---
phase: 39-certificates
plan: "02"
subsystem: provider
tags: [certificates, resource, data-source, tests, docs, roadmap]
dependency_graph:
  requires: [39-01-Certificate-client-CRUD]
  provides: [flashblade_certificate-resource, flashblade_certificate-data-source, certificate-provider-tests]
  affects: [ROADMAP.md, provider.go]
tech_stack:
  added: []
  patterns: [mapCertificateToModel, nullTimeoutsValue, write-only-sensitive-fields, drift-detection]
key_files:
  created:
    - internal/provider/certificate_resource.go
    - internal/provider/certificate_resource_test.go
    - internal/provider/certificate_data_source.go
    - internal/provider/certificate_data_source_test.go
    - examples/resources/flashblade_certificate/resource.tf
    - examples/resources/flashblade_certificate/import.sh
    - examples/data-sources/flashblade_certificate/data-source.tf
    - docs/resources/certificate.md
    - docs/data-sources/certificate.md
  modified:
    - internal/provider/provider.go
    - ROADMAP.md
decisions:
  - UseStateForUnknown only on id and certificate_type (stable after creation); all renewal-volatile fields (issued_by, issued_to, valid_from, valid_to, key_algorithm, key_size, status) have no plan modifier
  - private_key and passphrase are Sensitive and preserved from plan/state on Create/Read/Update; set to empty string on ImportState
  - nullTimeoutsValue() used in ImportState — returns null Object as expected (no plan available during import)
  - mapCertificateToModel helper handles SubjectAlternativeNames as types.ListValueMust with empty-list fallback for nil/empty slices
  - Data source model excludes private_key, passphrase, and timeouts (write-only and timeout fields not relevant for DS)
metrics:
  duration: "~513 seconds"
  completed: "2026-04-03"
  tasks: 2
  files: 11
---

# Phase 39 Plan 02: Certificate Provider Resource Summary

flashblade_certificate resource with full CRUD, import, drift detection on all computed fields, and flashblade_certificate data source, backed by 4 tests (3 resource + 1 data source).

## What Was Built

### Task 1: flashblade_certificate resource

Created `internal/provider/certificate_resource.go` following the target_resource.go pattern:

- All 4 interface assertions: `resource.Resource`, `ResourceWithConfigure`, `ResourceWithImportState`, `ResourceWithUpgradeState`
- Schema version 0 with `RequiresReplace()` on `name`, `UseStateForUnknown()` only on `id` and `certificate_type` (stable after creation)
- `private_key` and `passphrase`: `Sensitive: true`, not returned by API — preserved from plan/state on Create/Read/Update; set to empty string on ImportState
- Volatile computed fields with NO `UseStateForUnknown`: `issued_by`, `issued_to`, `valid_from`, `valid_to`, `key_algorithm`, `key_size`, `status`, `common_name`, `country`, `email`, `locality`, `organization`, `organizational_unit`, `state`
- `subject_alternative_names`: `types.ListAttribute` with `types.StringType` elements
- Drift detection on all 17 computed/mutable fields via `tflog.Debug` with `resource/field/was/now` keys
- `mapCertificateToModel` helper: maps all GET fields, builds SANs list, does NOT set PrivateKey/Passphrase
- `ImportState` by name: calls `GetCertificate`, sets `nullTimeoutsValue()`, sets PrivateKey=`""` and Passphrase=`""`
- `UpgradeState`: empty map (version 0)
- Timeout defaults: Create 20m, Read 5m, Update 20m, Delete 30m

### Task 2: Tests, data source, registration, examples, docs, ROADMAP

**Resource tests** (`internal/provider/certificate_resource_test.go`):
- `TestUnit_CertificateResource_Lifecycle`: Create → Read → Update (cert renewal PEM) → Delete; verifies private_key preserved through Create and Read
- `TestUnit_CertificateResource_Import`: Seeds cert in mock, imports by name, verifies all computed fields, confirms PrivateKey=`""` and Passphrase=`""` and Timeouts are null
- `TestUnit_CertificateResource_DriftDetection`: Seeds cert, overwrites mock store with different certificate_type/key_algorithm/issued_by, calls Read, verifies state reflects new API values

**Data source** (`internal/provider/certificate_data_source.go`):
- Model excludes private_key, passphrase, timeouts
- Reads certificate by name, maps all fields inline, handles SANs list, not-found → AddError

**Data source test** (`internal/provider/certificate_data_source_test.go`):
- `TestUnit_CertificateDataSource_Basic`: Seeds cert with full fields including 2 SANs, reads via data source, verifies id/name/certificate_type/common_name/issued_by/key_algorithm/key_size/status/valid_from/valid_to/SANs

**Provider registration** (`internal/provider/provider.go`):
- `NewCertificateResource` appended to Resources()
- `NewCertificateDataSource` appended to DataSources()

**HCL examples**:
- `examples/resources/flashblade_certificate/resource.tf`: minimal config with certificate + private_key, optional intermediate_certificate and passphrase commented out
- `examples/resources/flashblade_certificate/import.sh`: `terraform import flashblade_certificate.example my-tls-cert`
- `examples/data-sources/flashblade_certificate/data-source.tf`: lookup + status and valid_to outputs

**Docs**: `make docs` generated `docs/resources/certificate.md` and `docs/data-sources/certificate.md`

**ROADMAP.md**: Certificates entry moved from Planned to Done; covered counter updated ~35→~36, coverage ~65%→~66%

## Verification Results

- `make build` — PASSED
- `make test` — PASSED (all packages pass, 4 new certificate tests)
- `make lint` — PASSED (0 issues; fixed one QF1008 staticcheck finding on Timeouts.Object accessor)
- `docs/resources/certificate.md` exists — YES
- `docs/data-sources/certificate.md` exists — YES
- ROADMAP.md updated — YES

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed incorrect Timeouts null assertion in Import test**
- **Found during:** Task 2 — TestUnit_CertificateResource_Import failed
- **Issue:** Test asserted `model.Timeouts.Object.IsNull()` must be false (expected initialized), but `nullTimeoutsValue()` intentionally returns a null timeouts Object (no plan available during import). This is the correct behavior.
- **Fix:** Inverted assertion to `model.Timeouts.IsNull()` must be true; also fixed redundant `.Object` accessor per staticcheck QF1008
- **Files modified:** internal/provider/certificate_resource_test.go
- **Commit:** 871c0c9 (included in Task 2 commit)

## Known Stubs

None — all fields are wired to the API response. private_key and passphrase are intentionally empty after import (write-only fields, API never returns them), which is documented behavior.

## Self-Check: PASSED
