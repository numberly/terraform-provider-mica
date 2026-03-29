---
phase: 14-access-key-enhancement-and-array-connection
plan: 02
subsystem: api
tags: [flashblade, array-connection, data-source, replication, terraform]

# Dependency graph
requires:
  - phase: 13-object-store-access-key
    provides: "Provider data source patterns and client method conventions"
provides:
  - "ArrayConnection model struct"
  - "GetArrayConnection and ListArrayConnections client methods"
  - "flashblade_array_connection data source (read-only)"
  - "Mock handler for /api/2.22/array-connections with Seed method"
affects: [15-replication-credentials-and-replica-links]

# Tech tracking
tech-stack:
  added: []
  patterns: [data-source-only resource pattern, Seed-based mock test setup]

key-files:
  created:
    - internal/client/array_connections.go
    - internal/testmock/handlers/array_connections.go
    - internal/provider/array_connection_data_source.go
    - internal/provider/array_connection_data_source_test.go
  modified:
    - internal/client/models_admin.go
    - internal/provider/provider.go

key-decisions:
  - "Array connection is data-source-only (no POST/PATCH/DELETE) -- resource deferred to v2.1"
  - "Mock handler uses Seed method instead of AddXxx for pre-populating test data"

patterns-established:
  - "Data-source-only pattern: GET client methods + mock GET handler + Seed for tests"

requirements-completed: [ACN-01, ACN-02]

# Metrics
duration: 4min
completed: 2026-03-29
---

# Phase 14 Plan 02: Array Connection Data Source Summary

**Read-only data source for FlashBlade array connections by remote name, exposing connection ID, status, management/replication addresses for replication configuration**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-29T09:00:08Z
- **Completed:** 2026-03-29T09:03:52Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- ArrayConnection model with all API fields (id, status, remote, management_address, replication_addresses, encrypted, type, version)
- Client methods GetArrayConnection (by remote_names filter) and ListArrayConnections (with pagination)
- flashblade_array_connection data source with remote_name Required lookup and all other fields Computed
- 3 unit tests: successful read with all attributes, not-found error, schema validation

## Task Commits

Each task was committed atomically:

1. **Task 1: ArrayConnection model, client methods, and mock handler** - `48a54b2` (feat)
2. **Task 2 RED: Failing tests for array connection data source** - `80eb29c` (test)
3. **Task 2 GREEN: Array connection data source implementation** - `7035dd7` (feat)

## Files Created/Modified
- `internal/client/models_admin.go` - Added ArrayConnection struct
- `internal/client/array_connections.go` - GetArrayConnection and ListArrayConnections client methods
- `internal/testmock/handlers/array_connections.go` - Mock handler with GET + remote_names filter and Seed method
- `internal/provider/array_connection_data_source.go` - flashblade_array_connection data source implementation
- `internal/provider/array_connection_data_source_test.go` - 3 unit tests (Read, NotFound, Schema)
- `internal/provider/provider.go` - Registered NewArrayConnectionDataSource in DataSources()

## Decisions Made
- Array connection is data-source-only (no POST/PATCH/DELETE) -- resource deferred to v2.1
- Mock handler uses Seed method for pre-populating test data (simpler than AddXxx pattern for read-only resources)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Array connection data source provides connection IDs needed for Phase 15 (replication credentials and replica links)
- All patterns established for data-source-only resources

---
*Phase: 14-access-key-enhancement-and-array-connection*
*Completed: 2026-03-29*
