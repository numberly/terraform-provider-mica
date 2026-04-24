---
phase: 22
name: test-coverage
status: passed
verified: 2026-03-29
---

# Phase 22 Verification: Test Coverage

## Must-Haves

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| 1 | 5 data sources have Read + NotFound tests | ✓ | 10 new test functions, all passing |
| 2 | OAuth2 provider config test | ✓ | TestUnit_Configure_OAuth2 in provider_test.go |
| 3 | HCL acceptance tests with resource.UnitTest | ⚠ | Deferred — requires terraform/tofu binary on PATH |
| 4 | Pagination tests for buckets + one policy type | ✓ | 5 new tests (3 bucket, 2 network access policy) |

## Requirement Coverage

| Requirement | Status |
|-------------|--------|
| TST-01 | ✓ Complete — virtual host DS tests |
| TST-02 | ✓ Complete — remote credentials DS tests |
| TST-03 | ✓ Complete — bucket replica link DS tests |
| TST-04 | ✓ Complete — file system export DS tests |
| TST-05 | ✓ Complete — account export DS tests |
| TST-06 | ✓ Complete — OAuth2 provider config test |
| TST-07 | ⚠ Deferred — needs terraform-plugin-testing + binary |
| TST-08 | ✓ Complete — bucket + network access policy pagination |

## Build & Test

- `go build ./...`: ✓ Clean
- `go test ./internal/...`: ✓ 391 tests pass (16 new)

## Score: 7/8 requirements verified (1 deferred)

## Note on TST-07

HCL-based acceptance tests using `resource.UnitTest` from `terraform-plugin-testing` require a
terraform or OpenTofu binary on PATH. This is an environment dependency that should be set up
in CI. The requirement is valid but implementation is gated on environment setup, not code.
