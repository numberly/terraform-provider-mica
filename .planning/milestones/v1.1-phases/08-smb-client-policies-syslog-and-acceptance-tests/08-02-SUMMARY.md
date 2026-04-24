---
phase: 08-smb-client-policies-syslog-and-acceptance-tests
plan: 02
subsystem: infra
tags: [syslog, terraform-provider, flashblade, crud, listdefault]

# Dependency graph
requires:
  - phase: 08-01
    provides: SMB client policy resource family pattern and models.go structure
provides:
  - SyslogServer, SyslogServerPost, SyslogServerPatch model structs
  - Client CRUD methods for /syslog-servers endpoint
  - flashblade_syslog_server resource with CRUD + import
  - flashblade_syslog_server data source
  - Mock handler for syslog server test infrastructure
affects: [08-03-acceptance-tests]

# Tech tracking
tech-stack:
  added: []
  patterns: [listdefault.StaticValue for null-vs-empty on string list attributes]

key-files:
  created:
    - internal/client/syslog_servers.go
    - internal/testmock/handlers/syslog_servers.go
    - internal/provider/syslog_server_resource.go
    - internal/provider/syslog_server_data_source.go
    - internal/provider/syslog_server_resource_test.go
  modified:
    - internal/client/models.go
    - internal/provider/provider.go

key-decisions:
  - "Syslog server name uses RequiresReplace (not renameable per API)"
  - "Services and sources use listdefault.StaticValue with empty list to prevent null-vs-empty drift"

patterns-established:
  - "stringSliceFromList helper for extracting Go []string from types.List"

requirements-completed: [SYS-01, SYS-02, SYS-03]

# Metrics
duration: 4min
completed: 2026-03-28
---

# Phase 8 Plan 2: Syslog Server Resource Summary

**Syslog server resource with CRUD lifecycle, import, data source, and null-vs-empty drift prevention via listdefault.StaticValue**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-28T15:42:57Z
- **Completed:** 2026-03-28T15:47:00Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Syslog server models (SyslogServer, SyslogServerPost, SyslogServerPatch) added to models.go
- Client CRUD with 5 methods (Get, List, Post, Patch, Delete) following established pattern
- Resource with RequiresReplace on name, listdefault.StaticValue on services/sources
- Data source reads syslog server by name
- 3 unit tests passing: CRUD lifecycle, import 0-diff, data source read
- Full test suite (268 tests) passes with no regressions

## Task Commits

Each task was committed atomically:

1. **Task 1: Syslog server models, client CRUD, and mock handler** - `bf32258` (feat)
2. **Task 2: Syslog server resource, data source, provider registration, and unit tests** - `e4c3ad0` (feat)

## Files Created/Modified
- `internal/client/models.go` - Added SyslogServer, SyslogServerPost, SyslogServerPatch structs
- `internal/client/syslog_servers.go` - Client CRUD methods for /syslog-servers
- `internal/testmock/handlers/syslog_servers.go` - Mock handler with full CRUD and raw PATCH semantics
- `internal/provider/syslog_server_resource.go` - Resource with CRUD, import, listdefault.StaticValue
- `internal/provider/syslog_server_data_source.go` - Data source reads by name
- `internal/provider/syslog_server_resource_test.go` - 3 unit tests
- `internal/provider/provider.go` - Registered NewSyslogServerResource and NewSyslogServerDataSource

## Decisions Made
- Syslog server name uses RequiresReplace (not renameable per API research)
- Services and sources use listdefault.StaticValue with empty list default to prevent null-vs-empty drift (Pitfall 5)
- stringSliceFromList helper extracts Go []string from types.List for clean CRUD code

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Syslog server resource complete and tested
- Ready for acceptance tests in plan 08-03

---
*Phase: 08-smb-client-policies-syslog-and-acceptance-tests*
*Completed: 2026-03-28*
