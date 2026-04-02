---
phase: 36-target-resource
plan: 01
subsystem: client
tags: [go, terraform-plugin-framework, flashblade, replication, targets, mock, unit-tests]

requires: []

provides:
  - Target, TargetPost, TargetPatch model structs in internal/client/models_storage.go
  - GetTarget, PostTarget, PatchTarget, DeleteTarget CRUD methods on FlashBladeClient
  - Thread-safe mock handler for /api/2.22/targets in internal/testmock/handlers/targets.go
  - 6 unit tests covering all CRUD operations in internal/client/targets_test.go

affects:
  - 36-02 (provider resource and data source consume these client methods)

tech-stack:
  added: []
  patterns:
    - "getOneByName[Target] for GET by name with automatic IsNotFound on empty response"
    - "**NamedReference in TargetPatch for nullable object PATCH semantics (nil outer = omit, nil inner = set null)"
    - "Mock handler Seed() method for pre-populating test state"
    - "targetStoreFacade wrapper in tests to expose opaque store's Seed method"

key-files:
  created:
    - internal/client/targets.go
    - internal/testmock/handlers/targets.go
    - internal/client/targets_test.go
  modified:
    - internal/client/models_storage.go

key-decisions:
  - "Use **NamedReference for TargetPatch.CACertificateGroup to distinguish omit vs set-null PATCH semantics"
  - "Mock GET handler returns 404 JSON error (not empty list) when ?names= filter has no match — consistent with how getOneByName detects not-found"
  - "targetStoreFacade wrapper in test file exposes typed Seed method without making the store type public"

patterns-established:
  - "GET mock with ?names= filter returns 404 JSON error (not empty list) so getOneByName can detect not-found via HTTP status"

requirements-completed:
  - TGT-01
  - TGT-02
  - TGT-03
  - TGT-04
  - TGT-05

duration: 12min
completed: 2026-04-01
---

# Phase 36 Plan 01: Target Client Layer Summary

**Target CRUD client with GetTarget/PostTarget/PatchTarget/DeleteTarget, model structs, mock handler, and 6 passing unit tests**

## Performance

- **Duration:** ~12 min
- **Started:** 2026-04-01T00:00:00Z
- **Completed:** 2026-04-01T00:12:00Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Target, TargetPost, TargetPatch model structs appended to models_storage.go with correct JSON tags and PATCH semantics
- Four client methods (GetTarget, PostTarget, PatchTarget, DeleteTarget) following established remote_credentials.go pattern
- Thread-safe in-memory mock handler with GET/POST/PATCH/DELETE, 404 on unknown name, 409 on duplicate POST
- 6 unit tests all passing: found, notFound, post, patchAddress, patchCACertGroup, delete

## Task Commits

1. **Task 1: Define Target model structs** - `42e4d7f` (feat)
2. **Task 2: Implement client CRUD methods** - `e6efec9` (feat)
3. **Task 3: Build mock handler and write unit tests** - `9e9781a` (test)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/models_storage.go` - Added Target, TargetPost, TargetPatch structs
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/targets.go` - GetTarget, PostTarget, PatchTarget, DeleteTarget methods
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/targets.go` - Thread-safe mock for /api/2.22/targets
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/targets_test.go` - 6 unit tests using RegisterTargetHandlers

## Decisions Made

- Used `**NamedReference` for `TargetPatch.CACertificateGroup` to support two PATCH modes: nil outer pointer = omit field entirely, non-nil outer with nil inner = set field to JSON null (clear cert group)
- Mock GET handler returns HTTP 404 JSON error when `?names=` filter finds no match — this is required for `getOneByName` to detect not-found via HTTP status, rather than returning an empty list (which would produce the same "not found" error but via a different path)
- Used a `targetStoreFacade` struct in the test file to expose the `Seed` method without making the internal `targetStore` type public from the handlers package

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Client layer complete; ready for plan 36-02 (Terraform provider resource + data source)
- GetTarget, PostTarget, PatchTarget, DeleteTarget all exported and tested
- Mock handler registered under /api/2.22/targets, ready for provider-level mocked integration tests

---
*Phase: 36-target-resource*
*Completed: 2026-04-01*

## Self-Check: PASSED

- targets.go: FOUND
- models_storage.go: FOUND
- handlers/targets.go: FOUND
- targets_test.go: FOUND
- Commit 42e4d7f: FOUND
- Commit e6efec9: FOUND
- Commit 9e9781a: FOUND
