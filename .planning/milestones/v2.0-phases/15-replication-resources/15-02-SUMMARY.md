---
phase: 15-replication-resources
plan: 02
subsystem: api
tags: [terraform, remote-credentials, s3-replication, flashblade, mock-handler]

requires:
  - phase: 15-01
    provides: "Client models (ObjectStoreRemoteCredentials*) and CRUD methods (Get/Post/Patch/DeleteRemoteCredentials)"
provides:
  - "flashblade_object_store_remote_credentials resource with full CRUD lifecycle"
  - "flashblade_object_store_remote_credentials data source (read by name)"
  - "Mock CRUD handler for /api/2.22/object-store-remote-credentials"
affects: [15-03, docs, testing]

tech-stack:
  added: []
  patterns: [secret-preservation-in-state, query-param-based-create]

key-files:
  created:
    - internal/testmock/handlers/remote_credentials.go
    - internal/provider/remote_credentials_resource.go
    - internal/provider/remote_credentials_data_source.go
  modified:
    - internal/provider/provider.go

key-decisions:
  - "Secret preservation: secret_access_key kept from plan values in state (API strips it on GET)"
  - "Import sets secret_access_key to empty string; user must provide in config or use ignore_changes"

patterns-established:
  - "Secret-preserving resource: user-provided secrets kept from plan, not overwritten by API GET response"

requirements-completed: [RCR-01, RCR-02, RCR-03]

duration: 3min
completed: 2026-03-29
---

# Phase 15 Plan 02: Remote Credentials Resource and Data Source Summary

**S3 remote credentials resource with secret-preserving CRUD, data source, and mock handler for cross-array replication**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-29T11:29:33Z
- **Completed:** 2026-03-29T11:32:18Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Mock handler implementing GET/POST/PATCH/DELETE with proper secret stripping on GET and PATCH responses
- Resource supporting full CRUD lifecycle: Create (POST with name+remote_name query params), Read, Update (key rotation via PATCH), Delete, Import by name
- Data source exposing id, name, access_key_id, remote_name (no sensitive secret_access_key)
- Provider registration for both resource and data source

## Task Commits

Each task was committed atomically:

1. **Task 1: Mock handler for remote credentials** - `40fd94b` (feat)
2. **Task 2: Remote credentials resource, data source, and provider registration** - `461c97b` (feat)

## Files Created/Modified
- `internal/testmock/handlers/remote_credentials.go` - Mock CRUD handler with secret stripping
- `internal/provider/remote_credentials_resource.go` - Resource with Create/Read/Update/Delete/Import
- `internal/provider/remote_credentials_data_source.go` - Read-only data source by name
- `internal/provider/provider.go` - Registered resource and data source

## Decisions Made
- Secret preservation: secret_access_key kept from plan values in state since API GET does not return it
- Import sets secret_access_key to empty string; user must provide in config or use ignore_changes

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Remote credentials resource ready for cross-array bucket replication workflows
- Plan 15-03 (bucket replica links) can proceed; all prerequisite resources are in place

---
*Phase: 15-replication-resources*
*Completed: 2026-03-29*
