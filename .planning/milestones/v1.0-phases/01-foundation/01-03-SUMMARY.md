---
phase: 01-foundation
plan: 03
subsystem: testing
tags: [go, terraform-provider, httptest, mock-server, file-system, crud, tdd]

# Dependency graph
requires:
  - phase: 01-foundation
    plan: 01
    provides: "FlashBladeClient with get/post/patch/delete helpers, FileSystem/FileSystemPost/FileSystemPatch/ListResponse[T] models, IsNotFound, APIError"
provides:
  - File system CRUD client methods: GetFileSystem, ListFileSystems, PostFileSystem, PatchFileSystem, DeleteFileSystem, PollUntilEradicated
  - ListFileSystemsOpts struct for optional filter parameters
  - testmock.NewMockServer: reusable httptest.Server factory with built-in /login and /api/api_version handlers
  - testmock/handlers.RegisterFileSystemHandlers: in-memory thread-safe CRUD handler for provider-level tests
affects: [03-file-system-resource, all subsequent phases needing mock server]

# Tech tracking
tech-stack:
  added:
    - github.com/google/uuid v1.6.0 (promoted from indirect for handler use)
  patterns:
    - Client file system methods use get/post/patch/delete helpers from client.go — no raw HTTP in resource files
    - GetFileSystem returns IsNotFound when items is empty (FlashBlade API behavior: 200 with empty items, not 404)
    - PatchFileSystem uses IDs (not names) for PATCH/DELETE — stable across renames
    - PollUntilEradicated uses context.Done() for timeout, polls with 2s sleep
    - testmock handlers use raw map[string]json.RawMessage for PATCH to achieve true PATCH semantics
    - Mock server registers resource handlers via RegisterFileSystemHandlers(mux) pattern for composability

key-files:
  created:
    - internal/client/filesystems.go
    - internal/client/filesystems_test.go
    - internal/testmock/server.go
    - internal/testmock/server_test.go
    - internal/testmock/handlers/filesystems.go
  modified: []

key-decisions:
  - "GetFileSystem treats empty items list as not-found and returns APIError{StatusCode:404} — FlashBlade returns 200 with empty items, not HTTP 404"
  - "PatchFileSystem uses IDs as the selector (not names) for rename stability — confirmed per CONTEXT.md locked decision"
  - "testmock handlers use raw map[string]json.RawMessage for PATCH body parsing to faithfully implement PATCH semantics without overwriting absent fields"
  - "PollUntilEradicated polls ?destroyed=true endpoint (not the default endpoint) to avoid false positives from active file systems with same name"

patterns-established:
  - "Pattern: Client methods return *T from ListResponse[T].Items[0] — callers get a pointer, never deal with slices"
  - "Pattern: RegisterXxxHandlers(mux) — resource-specific handler registration keeps mock server composable"
  - "Pattern: In-memory store uses byName + byID dual-index with sync.Mutex for thread safety"
  - "Pattern: writeListResponse / writeError helpers in handlers for consistent JSON envelope format"

requirements-completed: [FS-01, FS-02, FS-03, FS-04]

# Metrics
duration: 25min
completed: 2026-03-27
---

# Phase 1, Plan 03: File System Client Methods and Mock Server Summary

**Pure-Go file system CRUD client (GetFileSystem, ListFileSystems, PostFileSystem, PatchFileSystem, DeleteFileSystem, PollUntilEradicated) with a reusable httptest mock server and thread-safe in-memory handlers — 14 unit tests, full CRUD lifecycle including soft-delete and eradication polling**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-03-27T06:28:46Z
- **Completed:** 2026-03-27T06:53:00Z
- **Tasks:** 2 (both TDD)
- **Files modified:** 5 created

## Accomplishments

- 6 file system client methods on `FlashBladeClient`: create, read, list, update (incl. rename), soft-delete, eradicate, poll-until-eradicated
- Reusable `testmock.NewMockServer` factory with built-in `/login` + `/api/api_version` — zero setup for test files
- `handlers.RegisterFileSystemHandlers` with thread-safe in-memory state: UUID generation, dual name/ID index, true PATCH semantics, delete guard requiring soft-delete first
- 14 passing unit tests covering all behaviors including timeout, PATCH omitempty semantics, and full CRUD lifecycle

## Task Commits

Each task was committed atomically:

1. **Task 1 TDD RED — failing file system client tests** - `fcf51ec` (test)
2. **Task 1 TDD GREEN — file system CRUD implementation** - `2ae46ab` (feat)
3. **Task 2 TDD RED — failing mock server tests** - `521efef` (test)
4. **Task 2 TDD GREEN — mock server and handlers implementation** - `9657812` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/filesystems.go` - File system CRUD methods: GetFileSystem, ListFileSystems, PostFileSystem, PatchFileSystem, DeleteFileSystem, PollUntilEradicated with ListFileSystemsOpts
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/filesystems_test.go` - 11 unit tests with per-test httptest.NewServer stubs
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/server.go` - MockServer factory with /login and /api/api_version built-in handlers
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/server_test.go` - 3 behavioral tests for full CRUD lifecycle via raw HTTP
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/filesystems.go` - RegisterFileSystemHandlers with in-memory CRUD, rename support, DELETE guard

## Decisions Made

- `GetFileSystem` synthesizes a 404 APIError when the items list is empty — the FlashBlade API returns HTTP 200 with empty items rather than HTTP 404 for non-existent resources. This keeps `IsNotFound()` usable throughout the provider layer.
- PATCH handler uses `map[string]json.RawMessage` rather than `FileSystemPatch` struct to achieve true PATCH semantics — the struct approach with `omitempty` would silently drop `false` booleans in some edge cases.
- `PollUntilEradicated` queries the `?destroyed=true` variant to ensure the poll only returns file systems that are in the soft-deleted state — avoids a race where a new file system with the same name is created before eradication completes.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

- All file system client methods functional; `go vet ./internal/...` clean; 14 tests green
- `testmock.MockServer` is ready for reuse in Plan 04 provider-level acceptance tests
- `handlers.RegisterFileSystemHandlers` establishes the composable pattern for future resource families (NFS exports, snapshots, etc.)
- No blockers for Plan 04

---
*Phase: 01-foundation*
*Completed: 2026-03-27*

## Self-Check: PASSED

- FOUND: `internal/client/filesystems.go`
- FOUND: `internal/client/filesystems_test.go`
- FOUND: `internal/testmock/server.go`
- FOUND: `internal/testmock/server_test.go`
- FOUND: `internal/testmock/handlers/filesystems.go`
- FOUND: `.planning/phases/01-foundation/01-03-SUMMARY.md`
- FOUND commit: `fcf51ec` (test: failing file system client tests)
- FOUND commit: `2ae46ab` (feat: file system CRUD implementation)
- FOUND commit: `521efef` (test: failing mock server tests)
- FOUND commit: `9657812` (feat: mock server and handlers implementation)
