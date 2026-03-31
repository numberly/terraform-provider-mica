---
phase: 33-client-hardening
plan: "01"
subsystem: client
tags: [oauth2, context, retry, refactor, hardening]
dependency_graph:
  requires: []
  provides: [context-aware-oauth2-token-source, explicit-retry-delay]
  affects: [internal/client, internal/provider]
tech_stack:
  added: []
  patterns: [context-propagation, explicit-duration-over-heuristic]
key_files:
  created: []
  modified:
    - internal/client/auth.go
    - internal/client/auth_test.go
    - internal/client/client.go
    - internal/client/transport_test.go
    - internal/client/filesystems_test.go
    - internal/provider/provider.go
    - internal/provider/helpers_test.go
decisions:
  - "FlashBladeTokenSource stores caller ctx at construction — oauth2.TokenSource.Token() has no context param so ctx is captured at NewFlashBladeTokenSource call site"
  - "RetryBaseDelay removed from Config entirely — callers cannot set it; 1000ms hardcoded as const in NewClient; buildTransport accepts explicit delay param"
  - "retry_base_delay removed from provider Terraform schema — breaking provider schema change accepted for v2.1.3"
metrics:
  duration: "~10m"
  tasks_completed: 2
  files_changed: 61
  completed_date: "2026-03-31"
---

# Phase 33 Plan 01: Client Hardening — OAuth2 Context + RetryBaseDelay Removal Summary

**One-liner:** Context propagation through OAuth2 token source via stored ctx field; RetryBaseDelay footgun eliminated with hardcoded 1000ms const replacing int-to-ms heuristic.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Fix OAuth2 context propagation in FlashBladeTokenSource | 26705a6 | auth.go, auth_test.go, client.go |
| 2 | Remove RetryBaseDelay heuristic from Config and update all callers | 4392587 | client.go, provider.go, 59 test files |

## What Was Built

### Task 1: OAuth2 Context Propagation (CH-01)

- Added `ctx context.Context` field to `FlashBladeTokenSource` struct
- Updated `NewFlashBladeTokenSource` signature: `func NewFlashBladeTokenSource(ctx context.Context, endpoint, apiToken string, httpClient *http.Client)`
- `Token()` now calls `ts.fetchToken(ts.ctx)` — when Terraform cancels via SIGINT, the in-flight OAuth2 token exchange is cancelled
- Removed the TODO comment about `context.Background()` workaround
- `NewClient` passes `ctx` to `NewFlashBladeTokenSource` — caller context flows end-to-end
- All 5 auth_test.go call sites updated to pass `context.Background()`

### Task 2: RetryBaseDelay Removal (CH-02)

- Removed `RetryBaseDelay time.Duration` from `Config` struct
- Deleted heuristic block: `if RetryBaseDelay < time.Millisecond { RetryBaseDelay * time.Millisecond }`
- Added `const defaultRetryDelay = 1000 * time.Millisecond` — used in `buildTransport` and OAuth2 transport wiring
- `buildTransport` now accepts `retryDelay time.Duration` as explicit parameter
- Removed `RetryBaseDelay` from `flashBladeProviderModel`, `Schema()`, and `Configure()` in provider.go
- Removed `retry_base_delay` Terraform schema attribute (breaking change, accepted for v2.1.3)
- Removed `time` import from provider.go (no longer needed)
- Batch-removed `RetryBaseDelay: 1 * time.Millisecond` lines from all 54 provider test files via sed
- Removed resulting unused `time` imports from 54 provider test files

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Dangling code in provider.go after sed removal**
- **Found during:** Task 2
- **Issue:** `sed -i '/RetryBaseDelay:/d'` removed the inner line `RetryBaseDelay: retryBaseDelay,` but left orphaned code: the `retryBaseDelay` variable assignment block, the `if delayStr := ...` block, and the closing brace remained as invalid Go
- **Fix:** Manually cleaned up the entire `retryBaseDelay` parsing block (lines 213-223) from Configure() — the variable is now unused since client.Config no longer accepts it
- **Files modified:** internal/provider/provider.go
- **Commit:** 4392587

**2. [Rule 2 - Missing cleanup] Unused `time` imports in 54 test files**
- **Found during:** Task 2 (`go vet` reported unused import)
- **Issue:** After removing `RetryBaseDelay: 1 * time.Millisecond` lines, many test files had `"time"` imported but no longer used
- **Fix:** Batch-removed unused `"time"` imports via shell loop (only files not actually using `time.X`)
- **Files modified:** 54 files in internal/provider/
- **Commit:** 4392587

## Self-Check: PASSED

- internal/client/auth.go: FOUND
- internal/client/client.go: FOUND
- internal/provider/provider.go: FOUND
- .planning/phases/33-client-hardening/33-01-SUMMARY.md: FOUND
- Commit 26705a6 (Task 1): FOUND
- Commit 4392587 (Task 2): FOUND
