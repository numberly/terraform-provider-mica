---
phase: 33-client-hardening
plan: "02"
subsystem: linting
tags: [golangci-lint, gosec, bodyclose, noctx, exhaustive, security, testing]
dependency_graph:
  requires:
    - phase: 33-01
      provides: context-aware-oauth2-token-source, explicit-retry-delay
  provides:
    - golangci-lint config with gosec/bodyclose/noctx/exhaustive linters enabled
    - zero linter violations in codebase
  affects: [ci, code-review, internal/client, internal/testmock]
tech_stack:
  added: []
  patterns: [nolint-with-rationale, context-propagation-in-tests, explicit-error-discarding]
key_files:
  created: []
  modified:
    - .golangci.yml
    - internal/client/transport.go
    - internal/client/auth_test.go
    - internal/client/transport_test.go
    - internal/testmock/handlers/object_store_access_keys.go
    - internal/testmock/handlers/query_params_test.go
    - internal/testmock/server_test.go
key_decisions:
  - "G104 body close errors suppressed with _ assignment (not nolint) — errors are structural cleanup, not actionable"
  - "G706 and G404 in transport.go suppressed via nolint with rationale — format string is hardcoded, jitter randomness needs no crypto strength"
  - "G404 in testmock handlers suppressed via nolint — test mock key generation is not a security boundary"
  - "noctx violations in test files fixed by replacing .Get() and httptest.NewRequest() with context-aware variants using context.Background()"
  - "test files excluded from gosec linter (path: _test.go exclusion rule) but NOT from noctx — noctx violations in tests fixed with real code changes"
requirements-completed:
  - CL-02
duration: ~10m
completed: "2026-03-31"
---

# Phase 33 Plan 02: Golangci-lint Expansion — gosec/bodyclose/noctx/exhaustive Summary

**Four security and correctness linters added to .golangci.yml v2 config; all 13 violations resolved through targeted code fixes and documented nolint suppressions.**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-03-31T16:35:00Z
- **Completed:** 2026-03-31T16:45:00Z
- **Tasks:** 1
- **Files modified:** 7

## Accomplishments

- Added gosec, bodyclose, noctx, and exhaustive to `.golangci.yml` (golangci-lint v2 format, alongside existing errcheck/govet/staticcheck/unused/ineffassign)
- Fixed 7 noctx violations by replacing `.Get()` and `httptest.NewRequest()` with context-aware variants across 4 test files
- Fixed 2 gosec G104 violations by using `_ = body.Close()` explicit discard pattern in transport.go
- Suppressed 4 documented false positives (G706 log injection, G404 jitter rand, G404 testmock rand) with inline nolint rationale
- `golangci-lint run ./...` exits 0, `go build ./...` exits 0, 250 unit tests pass

## Task Commits

1. **Task 1: Create .golangci.yml and fix all linter violations** - `b6f462a` (feat)

**Plan metadata:** _(docs commit follows)_

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.golangci.yml` - Added gosec, bodyclose, noctx, exhaustive with linters-settings; test file exclusion for gosec
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/transport.go` - `_ = req.Body.Close()`, `_ = resp.Body.Close()`; nolint for G706 and G404 jitter
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/auth_test.go` - `.Get()` replaced with `NewRequestWithContext + .Do()`
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/transport_test.go` - Three `.Get()` calls replaced with `NewRequestWithContext + .Do()`
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/object_store_access_keys.go` - nolint:gosec on two G404 rand calls in generateAccessKeyID/generateSecretAccessKey
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/query_params_test.go` - All `httptest.NewRequest()` calls updated to `httptest.NewRequestWithContext(context.Background(), ...)`
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/server_test.go` - `http.NewRequest` replaced with `http.NewRequestWithContext(context.Background(), ...)`

## Decisions Made

- G104 (`req.Body.Close()` and `resp.Body.Close()`) fixed with `_ =` discard — these are cleanup closes after body already consumed; error not actionable. Cleaner than nolint.
- G706 log injection suppressed with nolint — the Printf format string is a string literal; all args are int/duration/string values from internal state, not user-controlled input.
- G404 jitter rand suppressed with nolint — exponential backoff jitter has no security requirement; crypto/rand would add unnecessary overhead.
- G404 testmock key generators suppressed with nolint — these generate placeholder fake keys for unit test mocks, not real credentials.
- noctx test violations fixed with real code (not nolint) — tests should demonstrate correct patterns for the codebase they test.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Coverage] Fixed all noctx violations in test files beyond the flagged subset**
- **Found during:** Task 1 (iterating through linter output)
- **Issue:** The plan mentioned lines 32 and 65 in transport_test.go, but line 101 also contained `.Get()`. Additionally, query_params_test.go had 5 NewRequest calls (not 3 as originally flagged since linter may report first N).
- **Fix:** Fixed all occurrences in each file systematically rather than only the flagged lines.
- **Files modified:** transport_test.go, query_params_test.go
- **Committed in:** b6f462a

---

**Total deviations:** 1 auto-fixed (Rule 2 - missing coverage)
**Impact on plan:** Minimal — caught all occurrences instead of just flagged subset. No scope creep.

## Issues Encountered

- golangci-lint wrapper script on PATH (`golangci-lint`) had broken argument parsing (interpreted `run ./...` as file paths). Used full binary path `/home/gule/go/bin/golangci-lint` for all invocations.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 33 (client-hardening) complete: CH-01, CH-02, CL-02 all done.
- Phase 34 (test quality) is next: TQ-01 (ExpectNonEmptyPlan removal), TQ-02 (acceptance test expansion).
- Linter configuration is now a safety net for future PRs.

## Self-Check: PASSED

- .golangci.yml: FOUND
- internal/client/transport.go: FOUND
- .planning/phases/33-client-hardening/33-02-SUMMARY.md: FOUND
- Commit b6f462a (Task 1): FOUND

---
*Phase: 33-client-hardening*
*Completed: 2026-03-31*
