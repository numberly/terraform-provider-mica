---
phase: 06-server-resource-and-export-consolidation
plan: 01
subsystem: api
tags: [terraform, flashblade, server, dns, crud, import]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: client layer, mock framework, provider registration pattern
provides:
  - flashblade_server resource with full CRUD lifecycle and import
  - Extended server data source with DNS and created attributes
  - ServerDNS, ServerPost, ServerPatch client types
  - PostServer, PatchServer, DeleteServer client methods
  - Full CRUD mock handler for /servers
affects: [06-02, 07, 08]

# Tech tracking
tech-stack:
  added: []
  patterns: [ListNestedAttribute for DNS nested objects, create_ds query param for server creation]

key-files:
  created:
    - internal/provider/server_resource.go
    - internal/provider/server_resource_test.go
  modified:
    - internal/client/models.go
    - internal/client/servers.go
    - internal/testmock/handlers/servers.go
    - internal/provider/server_data_source.go
    - internal/provider/server_data_source_test.go
    - internal/provider/provider.go

key-decisions:
  - "DNS modeled as ListNestedAttribute (ordered) since API preserves DNS entry order"
  - "cascade_delete is a write-only list attribute used only on Delete, not stored in API state"
  - "Server creation uses ?create_ds= query param (not ?names=) per FlashBlade API convention"

patterns-established:
  - "ListNestedAttribute with nested string lists: DNS entries with domain/nameservers/services"
  - "Write-only attributes: cascade_delete used only on destroy, not reconciled from API"

requirements-completed: [SRV-01, SRV-02, SRV-03, SRV-04, SRV-05]

# Metrics
duration: 5min
completed: 2026-03-28
---

# Phase 6 Plan 1: Server Resource Summary

**FlashBlade server resource with DNS CRUD lifecycle, import, cascade delete, and extended data source**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-28T14:21:04Z
- **Completed:** 2026-03-28T14:26:30Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Server resource with Create/Read/Update/Delete/Import and DNS configuration
- Extended server data source with dns and created attributes
- Full CRUD mock handler for /servers with create_ds param support
- 9 unit tests passing (Create, Read, Update, Delete, Import, NotFound, PlanModifiers, DataSource, DataSource_NotFound)

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend Server model, client CRUD, and mock handler** - `27271d3` (feat)
2. **Task 2: Server resource, data source update, provider registration, and tests** - `2580c34` (feat)

## Files Created/Modified
- `internal/client/models.go` - Added ServerDNS, ServerPost, ServerPatch types; extended Server with DNS/Created
- `internal/client/servers.go` - Added PostServer, PatchServer, DeleteServer client methods
- `internal/testmock/handlers/servers.go` - Full CRUD mock handler with POST/PATCH/DELETE and DNS seed data
- `internal/provider/server_resource.go` - flashblade_server resource with full CRUD/import lifecycle
- `internal/provider/server_resource_test.go` - 7 unit tests for resource operations
- `internal/provider/server_data_source.go` - Extended with dns and created attributes
- `internal/provider/server_data_source_test.go` - Updated tests with dns/created type support
- `internal/provider/provider.go` - Registered NewServerResource

## Decisions Made
- DNS modeled as ListNestedAttribute (ordered) since the FlashBlade API preserves DNS entry order
- cascade_delete is a write-only list attribute used only during Delete, not reconciled from API state
- Server creation uses ?create_ds= query param (not ?names=) per FlashBlade API convention for servers

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Pre-existing TestUnit_FileSystemExport_Update failure detected during regression -- unrelated to this plan, not addressed (out of scope)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Server resource complete, ready for export consolidation (plan 06-02)
- cascade_delete param ready for acceptance testing in Phase 8

## Self-Check: PASSED

All 8 files verified present. Both task commits (27271d3, 2580c34) verified in git log.

---
*Phase: 06-server-resource-and-export-consolidation*
*Completed: 2026-03-28*
