---
phase: 39-certificates
plan: "01"
subsystem: client
tags: [certificates, client, mock, tests]
dependency_graph:
  requires: []
  provides: [Certificate-client-CRUD, certificate-mock-handler, certificate-unit-tests]
  affects: [39-02-certificate-resource]
tech_stack:
  added: []
  patterns: [getOneByName, certificateStore, WriteJSONListResponse]
key_files:
  created:
    - internal/client/certificates.go
    - internal/testmock/handlers/certificates.go
    - internal/client/certificates_test.go
  modified:
    - internal/client/models_network.go
decisions:
  - Certificate models appended to models_network.go (network/TLS domain)
  - POST struct excludes X.509 subject fields (extracted from PEM by API)
  - passphrase and private_key are write-only — never stored or returned by mock handler
  - GET handler returns HTTP 200 + empty list when name not found (matches real API)
  - CertificateType defaults to appliance in mock POST if not provided
metrics:
  duration: "~5 minutes"
  completed: "2026-04-03"
  tasks: 2
  files: 4
---

# Phase 39 Plan 01: Certificate Client CRUD Summary

Certificate client CRUD (GetCertificate, PostCertificate, PatchCertificate, DeleteCertificate) with mock handler and 5 unit tests following the targets.go pattern.

## What Was Built

### Task 1: Certificate model structs and client CRUD methods

Added `Certificate`, `CertificatePost`, and `CertificatePatch` structs to `internal/client/models_network.go` (certificates belong to the network/TLS domain alongside subnets and network interfaces).

Created `internal/client/certificates.go` with four CRUD methods:
- `GetCertificate` — uses `getOneByName[Certificate]` generic helper
- `PostCertificate` — POSTs to `/certificates?names=<name>`, decodes `ListResponse[Certificate]`
- `PatchCertificate` — PATCHes with pointer semantics (nil = omit field)
- `DeleteCertificate` — DELETEs by name via query param

Key model decisions:
- GET struct: no pointers on scalars, `[]string` for SANs
- POST struct: only import-relevant fields (certificate, certificate_type, intermediate_certificate, passphrase, private_key) — X.509 subject fields excluded because API extracts them from PEM
- PATCH struct: all pointer fields — only 4 mutable fields for cert renewal

### Task 2: Mock handler and client unit tests

Created `internal/testmock/handlers/certificates.go`:
- Thread-safe `certificateStore` with `Seed()` method
- GET returns HTTP 200 + empty list when name not found (critical — not 404)
- POST generates all computed fields (IssuedBy, IssuedTo, Status, ValidFrom, ValidTo, KeyAlgorithm, KeySize)
- POST never stores `private_key` or `passphrase` (write-only fields)
- PATCH applies non-nil pointer fields; passphrase/private_key accepted but not stored

Created `internal/client/certificates_test.go` with 5 tests:
- `TestUnit_Certificate_Get_Found` — seeds cert, verifies all computed fields returned
- `TestUnit_Certificate_Get_NotFound` — verifies `client.IsNotFound(err)` on missing name
- `TestUnit_Certificate_Post` — verifies computed fields populated after POST
- `TestUnit_Certificate_Patch` — seeds + patches certificate PEM, verifies update
- `TestUnit_Certificate_Delete` — deletes then verifies subsequent GET returns not-found

## Verification Results

- `go build ./internal/client/...` — PASSED
- `go test ./internal/client/... -run TestUnit_Certificate -count=1` — 5/5 PASSED
- `go vet ./internal/client/...` — PASSED (0 issues)

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED
