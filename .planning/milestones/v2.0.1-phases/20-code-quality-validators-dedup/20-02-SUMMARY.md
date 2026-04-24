---
phase: 20-code-quality-validators-dedup
plan: 02
subsystem: api
tags: [go-generics, client-layer, deduplication, code-quality]

# Dependency graph
requires:
  - phase: 20-code-quality-validators-dedup
    provides: "mustObjectValue returning diags (plan 01), DiagnosticReporter interface"
provides:
  - "getOneByName[T] generic helper for single-entity lookups"
  - "pollUntilGone[T] generic helper for eradication polling"
  - "Shared mapFSToModel between filesystem resource and data source"
affects: [21-dead-code-cleanup, 22-test-coverage]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Generic helpers for FlashBlade API list-then-pick-first pattern", "Shared model mapping via temporary struct copy"]

key-files:
  created: []
  modified:
    - "internal/client/client.go"
    - "internal/client/filesystems.go"
    - "internal/client/buckets.go"
    - "internal/provider/filesystem_data_source.go"

key-decisions:
  - "Used package-level generic functions (not methods) since Go does not support generic methods on structs"
  - "Shared mapFSToModel via temporary filesystemModel copy to data source model -- avoids interface overhead while eliminating 60+ lines of duplication"

patterns-established:
  - "getOneByName[T]: standard pattern for all single-entity lookups across client package"
  - "pollUntilGone[T]: standard pattern for polling destroyed resources until eradication"

requirements-completed: [DUP-06, DUP-07, DUP-08, MOD-02]

# Metrics
duration: 20min
completed: 2026-03-29
---

# Phase 20 Plan 02: Generic Client Helpers & Shared FS Mapping Summary

**Go generics eliminate 17 identical Get* bodies via getOneByName[T] and 2 poll loops via pollUntilGone[T]; shared mapFSToModel removes 60+ duplicated lines from data source**

## Performance

- **Duration:** 20 min
- **Started:** 2026-03-29T18:12:33Z
- **Completed:** 2026-03-29T18:33:15Z
- **Tasks:** 2
- **Files modified:** 19

## Accomplishments
- Extracted getOneByName[T] generic function called by 17 Get* methods across 17 client files
- Extracted pollUntilGone[T] generic function replacing 2 identical eradication polling loops
- Eliminated 60+ lines of duplicated mustObjectValue mapping from filesystem_data_source.go
- Verified zero panic() calls remain in production provider code

## Task Commits

Each task was committed atomically:

1. **Task 1: Add generic getOneByName and pollUntilGone to client.go, refactor callers** - `8e29567` (refactor)
2. **Task 2: Share mapFSToModel and verify mustObjectValue panic removal** - `4613d9c` (refactor)

## Files Created/Modified
- `internal/client/client.go` - Added getOneByName[T] and pollUntilGone[T] generic helpers
- `internal/client/filesystems.go` - Refactored GetFileSystem and PollUntilEradicated to use generics
- `internal/client/buckets.go` - Refactored GetBucket and PollBucketUntilEradicated to use generics
- `internal/client/servers.go` - Refactored GetServer to use getOneByName
- `internal/client/object_store_accounts.go` - Refactored GetObjectStoreAccount
- `internal/client/object_store_access_keys.go` - Refactored GetObjectStoreAccessKey
- `internal/client/object_store_access_policies.go` - Refactored GetObjectStoreAccessPolicy
- `internal/client/object_store_account_exports.go` - Refactored GetObjectStoreAccountExport
- `internal/client/object_store_virtual_hosts.go` - Refactored GetObjectStoreVirtualHost
- `internal/client/remote_credentials.go` - Refactored GetRemoteCredentials
- `internal/client/file_system_exports.go` - Refactored GetFileSystemExport
- `internal/client/nfs_export_policies.go` - Refactored GetNfsExportPolicy
- `internal/client/smb_share_policies.go` - Refactored GetSmbSharePolicy
- `internal/client/smb_client_policies.go` - Refactored GetSmbClientPolicy
- `internal/client/snapshot_policies.go` - Refactored GetSnapshotPolicy
- `internal/client/s3_export_policies.go` - Refactored GetS3ExportPolicy
- `internal/client/network_access_policies.go` - Refactored GetNetworkAccessPolicy
- `internal/client/syslog_servers.go` - Refactored GetSyslogServer
- `internal/provider/filesystem_data_source.go` - Replaced inline mapping with shared mapFSToModel

## Decisions Made
- Used package-level generic functions (not methods) since Go does not support generic methods on structs
- Shared mapFSToModel via temporary filesystemModel copy -- pragmatic approach avoiding interface overhead while eliminating all duplicated mapping logic

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All client Get* methods now use getOneByName[T] -- consistent pattern for future resources
- pollUntilGone[T] available for any future resource needing eradication polling
- Ready for Phase 21 (dead code cleanup) and Phase 22 (test coverage)

---
*Phase: 20-code-quality-validators-dedup*
*Completed: 2026-03-29*
