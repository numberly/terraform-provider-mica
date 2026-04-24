---
phase: 10-architecture-cleanup
plan: 01
subsystem: api
tags: [go, refactoring, code-organization]

# Dependency graph
requires:
  - phase: 09-bug-fixes
    provides: stable models.go with all model structs
provides:
  - domain-specific model files (models_common, models_storage, models_policies, models_exports, models_admin)
affects: [all-phases]

# Tech tracking
tech-stack:
  added: []
  patterns: ["domain-scoped model files in client package"]

key-files:
  created:
    - internal/client/models_common.go
    - internal/client/models_storage.go
    - internal/client/models_policies.go
    - internal/client/models_exports.go
    - internal/client/models_admin.go
  modified: []

key-decisions:
  - "Split monolithic models.go into 5 domain files for navigability and reduced merge conflicts"

patterns-established:
  - "Domain-scoped model files: common, storage, policies, exports, admin"

requirements-completed: [ARC-01]

# Metrics
duration: 3min
completed: 2026-03-28
---

# Phase 10 Plan 01: Split Models Summary

**Split 858-line monolithic models.go into 5 domain-scoped files (common, storage, policies, exports, admin) with zero logic changes**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-28T21:05:36Z
- **Completed:** 2026-03-28T21:08:55Z
- **Tasks:** 2
- **Files modified:** 6 (1 deleted, 5 created)

## Accomplishments
- Split monolithic models.go (858 lines) into 5 domain-specific files
- Zero import changes needed for consumers (all types remain in client package)
- All 46 client package tests pass without modification

## Task Commits

Each task was committed atomically:

1. **Task 1: Create 5 domain files from models.go and delete the original** - `621a9d0` (refactor)
2. **Task 2: Run full test suite to confirm zero regressions** - no commit (verification-only task, no files changed)

**Plan metadata:** (pending)

## Files Created/Modified
- `internal/client/models_common.go` - Space, ListResponse, VersionResponse, NamedReference, NumericIDReference, PolicyMember
- `internal/client/models_storage.go` - FileSystem, Bucket, ObjectStoreAccount, ObjectStoreAccessKey and Post/Patch variants
- `internal/client/models_policies.go` - All NFS/SMB/Snapshot/OAP/NAP/Quota/S3 export policy structs
- `internal/client/models_exports.go` - Server, FileSystemExport, ObjectStoreAccountExport, VirtualHost and variants
- `internal/client/models_admin.go` - ArrayDns, ArrayInfo, SmtpServer, AlertWatcher, SyslogServer and variants
- `internal/client/models.go` - DELETED

## Decisions Made
- Split monolithic models.go into 5 domain files for navigability and reduced merge conflicts
- Domain grouping: common (shared types), storage (fs/bucket/account/keys), policies (all policy families), exports (servers/exports/virtual hosts), admin (array config)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Pre-existing `go vet` failure in `internal/provider/helpers_test.go` (undefined compositeID) and build failure in `internal/provider/helpers.go` (unused imports) -- both completely unrelated to the models split. Not addressed (out of scope).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Domain model files ready for future development
- Smaller files reduce merge conflicts when adding new model structs
- All existing tests pass, consumers unaffected

---
*Phase: 10-architecture-cleanup*
*Completed: 2026-03-28*
