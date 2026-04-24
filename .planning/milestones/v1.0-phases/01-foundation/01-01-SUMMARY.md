---
phase: 01-foundation
plan: 01
subsystem: infra
tags: [go, terraform-provider, terraform-plugin-framework, oauth2, tls, retry, http-client]

# Dependency graph
requires: []
provides:
  - Go module github.com/soulkyu/terraform-provider-flashblade initialized with terraform-plugin-framework v1.19.0
  - FlashBladeClient with session-token and OAuth2 token-exchange auth
  - retryTransport with exponential backoff (429/5xx), X-Request-ID injection
  - TLS transport with custom CA cert (file or inline PEM) and InsecureSkipVerify
  - NegotiateVersion validates API v2.22 support on startup
  - Pure Go models: FileSystem, FileSystemPost, FileSystemPatch, Space, NFSConfig, SMBConfig, ListResponse[T], VersionResponse
  - APIError type with IsNotFound, IsRetryable, ParseAPIError helpers
  - Placeholder provider satisfying provider.Provider interface (stubs for Phase 02)
  - Build tooling: GNUmakefile, .golangci.yml, .goreleaser.yml
affects: [02-provider-schema, 03-file-system-resource, all subsequent phases]

# Tech tracking
tech-stack:
  added:
    - terraform-plugin-framework v1.19.0
    - terraform-plugin-log v0.10.0
    - terraform-plugin-go v0.31.0
    - terraform-plugin-testing v1.15.0
    - terraform-plugin-docs v0.24.0
    - golang.org/x/oauth2 v0.36.0
    - Go 1.25.0 (upgraded from 1.22 by framework requirement)
  patterns:
    - Client layer has zero terraform-plugin-framework imports — pure Go, fully testable with httptest.NewServer
    - retryTransport wraps base http.RoundTripper; retry logic is independent of business logic
    - FlashBladeTokenSource implements oauth2.TokenSource with mutex-protected caching
    - NewClient accepts Config struct — all options centralized, easy to extend
    - buildTransport returns http.RoundTripper to enable composability

key-files:
  created:
    - go.mod
    - go.sum
    - main.go
    - GNUmakefile
    - .golangci.yml
    - .goreleaser.yml
    - terraform-registry-manifest.json
    - tools/tools.go
    - internal/provider/provider.go
    - internal/client/client.go
    - internal/client/auth.go
    - internal/client/transport.go
    - internal/client/errors.go
    - internal/client/models.go
    - internal/client/client_test.go
    - internal/client/auth_test.go
    - internal/client/transport_test.go
  modified: []

key-decisions:
  - "Client layer is pure Go with zero terraform-plugin-framework imports — testable with httptest.NewServer"
  - "OAuth2 uses custom FlashBladeTokenSource (token-exchange grant) not standard clientcredentials.Config"
  - "RetryBaseDelay in Config.RetryBaseDelay accepts time.Duration; <1ms treated as raw milliseconds for test ergonomics"
  - "HTTPClient() exported method on FlashBladeClient allows transport-layer testing without mocking internals"
  - "Test CA helper generates separate CA cert + server cert (IP SAN 127.0.0.1) — CA cert alone as server cert fails ExtKeyUsageServerAuth validation"

patterns-established:
  - "Pattern: retryTransport.RoundTrip snapshots body bytes for replay on retry"
  - "Pattern: ParseAPIError reads response body and returns *APIError for HTTP >= 400"
  - "Pattern: FlashBladeTokenSource.Token() is goroutine-safe via sync.Mutex"
  - "Pattern: buildTransport isolates TLS+retry plumbing from auth plumbing"

requirements-completed: [PROV-01, PROV-02, PROV-04, PROV-05, PROV-07]

# Metrics
duration: 35min
completed: 2026-03-26
---

# Phase 1, Plan 01: Project Scaffold and HTTP Client Summary

**Go module + pure-Go FlashBlade HTTP client with session-token auth, OAuth2 token-exchange, TLS custom CA, exponential-backoff retry transport, and API v2.22 version negotiation — 15 unit tests, zero framework dependencies in client layer**

## Performance

- **Duration:** ~35 min
- **Started:** 2026-03-26T16:30:00Z
- **Completed:** 2026-03-26T17:05:00Z
- **Tasks:** 2
- **Files modified:** 17

## Accomplishments

- Full project scaffold: Go module, main.go, provider placeholder, build tooling (GNUmakefile, golangci-lint v2, goreleaser, tfplugindocs)
- HTTP client layer (`internal/client/`) with all auth, TLS, retry, error, and model primitives — zero terraform-plugin-framework imports
- 15 unit tests covering all behaviors: NewClient, custom CA (file + inline PEM), InsecureSkipVerify, version negotiation, retry transport (429/503/max), OAuth2 token source with caching, LoginWithAPIToken, and APIError classification

## Task Commits

Each task was committed atomically:

1. **Task 1: Project scaffold and build tooling** - `a34ea37` (feat)
2. **Task 2: TDD RED — failing client tests** - `3e56b21` (test)
3. **Task 2: TDD GREEN — HTTP client implementation** - `4b11307` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/main.go` - Provider binary entry point using providerserver.Serve
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider.go` - Placeholder provider implementing provider.Provider interface stubs
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/client.go` - FlashBladeClient struct, NewClient, buildTransport, NegotiateVersion, CRUD helpers
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/auth.go` - LoginWithAPIToken (session), FlashBladeTokenSource (OAuth2 token-exchange with caching)
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/transport.go` - retryTransport: exponential backoff, X-Request-ID injection, body replay
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/errors.go` - APIError, IsNotFound, IsRetryable, ParseAPIError
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/models.go` - FileSystem, FileSystemPost, FileSystemPatch, Space, NFS/SMBConfig, ListResponse[T], VersionResponse
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/GNUmakefile` - build/test/testacc/lint/generate/install targets
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.golangci.yml` - golangci-lint v2 with errcheck/govet/staticcheck/unused
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.goreleaser.yml` - Multi-platform release config (linux/darwin/windows, amd64/arm64)
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/terraform-registry-manifest.json` - Protocol v6.0 manifest
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/tools/tools.go` - tfplugindocs blank import

## Decisions Made

- Client layer has zero terraform-plugin-framework imports — this is intentional. Resources and the provider layer import the framework; the client is a plain Go HTTP library.
- Used custom `FlashBladeTokenSource` instead of `clientcredentials.Config` because FlashBlade uses non-standard `urn:ietf:params:oauth:grant-type:token-exchange` grant type.
- `HTTPClient()` exported on `FlashBladeClient` to allow transport-layer testing (RetryTransport tests) without needing to mock internal methods.
- `RetryBaseDelay` in Config accepts `time.Duration` but values `< 1ms` are treated as raw milliseconds — this allows tests to pass `RetryBaseDelay: 1` (1ms) without verbose `time.Millisecond` syntax.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] TLS test helper generated invalid server cert**
- **Found during:** Task 2 (TestUnit_CustomCATLS, TestUnit_CustomCATLS_InlinePEM)
- **Issue:** Original test helper used the CA certificate itself as the server TLS certificate. The CA cert lacks `ExtKeyUsageServerAuth`, causing TLS handshake to fail with "bad certificate". The `x509.Certificate` for a server cert must have `ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}` and `IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}` as SAN.
- **Fix:** Rewrote `generateTestCerts()` to generate a proper CA cert + a separate server cert signed by the CA, with correct ExtKeyUsage and IP SAN.
- **Files modified:** `internal/client/client_test.go`
- **Verification:** Both TLS tests pass (`TestUnit_CustomCATLS`, `TestUnit_CustomCATLS_InlinePEM`)
- **Committed in:** `4b11307` (Task 2 feat commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 — bug in test helper)
**Impact on plan:** Auto-fix required for TLS tests to be valid. No scope creep.

## Issues Encountered

- `terraform-plugin-framework@v1.19.0` requires Go 1.25.0 — `go get` automatically upgraded go.mod from `go 1.22.2` to `go 1.25.0`. This is expected per RESEARCH.md.
- `x509.SystemCertPool()` succeeds on this Linux host but test helper used CA cert directly as server cert — fails `ExtKeyUsageServerAuth` check. Fixed per Rule 1.

## Next Phase Readiness

- Go module and all dependencies resolved; `go build ./...` clean
- Client layer fully functional; ready for provider schema implementation (Plan 02)
- Provider placeholder compiles and satisfies `provider.Provider` interface
- 15 unit tests green, `go vet` clean, zero framework imports in client layer
- No blockers for Phase 1 Plan 02

---
*Phase: 01-foundation*
*Completed: 2026-03-26*
