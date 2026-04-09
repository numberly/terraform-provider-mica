---
phase: quick
plan: 260409-b4l
subsystem: provider/array-dns
tags: [refactor, named-resource, state-upgrader]
dependency_graph:
  requires: []
  provides: [named-dns-crud, dns-state-v1]
  affects: [flashblade_array_dns, flashblade_array_dns-data-source]
tech_stack:
  added: []
  patterns: [getOneByName, postOne, patchOne, byName-mock-store, v0-v1-state-upgrader]
key_files:
  created: []
  modified:
    - internal/client/array_admin.go
    - internal/client/array_admin_test.go
    - internal/client/models_admin.go
    - internal/testmock/handlers/array_admin.go
    - internal/provider/array_dns_resource.go
    - internal/provider/array_dns_resource_test.go
    - internal/provider/array_dns_data_source.go
    - examples/resources/flashblade_array_dns/resource.tf
    - examples/resources/flashblade_array_dns/import.sh
    - examples/data-sources/flashblade_array_dns/data-source.tf
    - docs/resources/array_dns.md
    - docs/data-sources/array_dns.md
    - CONVENTIONS.md
decisions:
  - "State upgrader sets Name='default' for v0->v1 migration"
  - "DNS mock store separated as arrayDnsStore with SeedDns delegator on arrayAdminStore"
  - "Delete now sends real DELETE instead of PATCH-reset-to-defaults"
metrics:
  duration: 659s
  completed: "2026-04-09"
  tasks: 2
  files: 13
---

# Quick Task 260409-b4l: Transform flashblade_array_dns from Singleton to Named Resource

Named DNS CRUD with getOneByName/postOne/patchOne generics, byName mock store, schema v1 state upgrader, import by name.

## Task Summary

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Client + Models + Mock | ffb3e58 | array_admin.go, models_admin.go, handlers/array_admin.go, array_admin_test.go |
| 2 | Resource + Data Source + Tests + Docs | c05cf44 | array_dns_resource.go, array_dns_data_source.go, tests, examples, docs, CONVENTIONS.md |

## Changes Made

### Client Layer
- All DNS methods now accept `name string` parameter with `?names=` query param
- `GetArrayDns` uses `getOneByName[ArrayDns]` generic (was hand-rolled list+filter)
- `PostArrayDns` uses `postOne` generic with name query param
- `PatchArrayDns` uses `patchOne` generic with name query param
- Added `DeleteArrayDns` method (did not exist before)
- `ArrayDnsPost` model has `Name` field with `json:"-"` tag

### Mock Handler
- DNS store separated from `arrayAdminStore` into dedicated `arrayDnsStore` with `byName` map
- `arrayAdminStore` exposes `DnsStore` field and `SeedDns()` convenience method
- DNS handler supports GET (empty list on miss), POST (409 conflict), PATCH (404 on miss), DELETE
- Uses `ValidateQueryParams` and `RequireQueryParam` shared helpers

### Resource
- Added `name` attribute: Required, RequiresReplace
- Schema version bumped 0 -> 1
- v0->v1 state upgrader: sets `Name = "default"` for existing states
- Create: standard POST (removed singleton GET-first-then-POST-or-PATCH logic)
- Delete: real DELETE (removed PATCH-reset-to-defaults)
- Import: by DNS config name (was hardcoded "default")
- Drift detection added on `name` field

### Data Source
- Added `name` attribute: Required
- Read uses name for lookup via `GetArrayDns(ctx, name)`
- Not-found produces AddError

### Tests
- 5 client tests: Get_Found, Get_NotFound, Post, Patch, Delete
- 6 resource tests: Lifecycle, Import, DriftDetection, StateUpgrade_V0toV1, PlanModifiers, SchemaV1HasName
- 1 data source test: Basic
- Total: 753 tests (was 752, net +1)

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Decisions Made

1. **State upgrader Name default**: v0->v1 sets `Name = types.StringValue("default")` since name is Required+RequiresReplace. Next `terraform plan` will show the real name if it differs.
2. **DNS mock store separation**: Created `arrayDnsStore` struct with its own mutex and byName map, embedded in `arrayAdminStore.DnsStore`. NTP/SMTP/alert watchers remain on the parent store unchanged.
3. **Real DELETE**: Delete now sends HTTP DELETE instead of PATCH-reset-to-defaults, matching standard named resource pattern.

## Self-Check: PASSED

- All 13 modified files exist on disk
- Both commits verified: ffb3e58, c05cf44
- make test: 753 passed (>= 752 baseline)
- make lint: 0 issues
- make docs: regenerated
