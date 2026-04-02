---
phase: 36-target-resource
plan: 02
subsystem: provider
tags: [go, terraform-plugin-framework, flashblade, replication, targets, resource, datasource, mocked-tests]

requires:
  - 36-01

provides:
  - flashblade_target resource (CRUD, drift detection, import) in internal/provider/target_resource.go
  - data.flashblade_target data source in internal/provider/target_data_source.go
  - 4 mocked integration tests (3 resource + 1 data source)
  - NewTargetResource and NewTargetDataSource registered in provider.go
  - HCL examples in examples/resources/flashblade_target/ and examples/data-sources/flashblade_target/
  - Auto-generated docs in docs/resources/target.md and docs/data-sources/target.md

affects:
  - Phase 37 (RC + BRL extension may reference target name)

tech-stack:
  added: []
  patterns:
    - "Flat string attribute for ca_certificate_group (not nested object) — keeps schema simple"
    - "Drift detection via tflog.Debug with field/was/now keys on every Read"
    - "**NamedReference outer-nil=omit, inner-nil=clear semantics for PATCH ca_certificate_group"
    - "nullTimeoutsValue() on ImportState to initialize timeouts without plan values"

key-files:
  created:
    - internal/provider/target_resource.go
    - internal/provider/target_resource_test.go
    - internal/provider/target_data_source.go
    - internal/provider/target_data_source_test.go
    - examples/resources/flashblade_target/resource.tf
    - examples/data-sources/flashblade_target/data-source.tf
    - docs/resources/target.md
    - docs/data-sources/target.md
  modified:
    - internal/provider/provider.go

key-decisions:
  - "Flat ca_certificate_group string in schema (not nested object) — consistent with plan spec, simpler HCL"
  - "Drift detection on Read logs all four mutable/computed fields: address, ca_certificate_group, status, status_details"
  - "targetStoreFacade removed — handlers.RegisterTargetHandlers returns *targetStore directly, Seed is callable"

duration: ~6min
completed: 2026-04-02
---

# Phase 36 Plan 02: Target Provider Resource and Data Source Summary

**flashblade_target resource with full CRUD, drift detection, import, and data source; registered in provider; 686 tests passing**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-04-02T15:15:24Z
- **Completed:** 2026-04-02T15:21:03Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments

- `flashblade_target` resource: Create, Read (with tflog drift detection), Update (patch address and ca_certificate_group), Delete, ImportState all implemented following remoteCredentialsResource pattern
- Flat `ca_certificate_group` string attribute with Optional+Computed — null when unset, stores group name string
- Data source `data.flashblade_target` reads by name, exposes address, status, status_details, ca_certificate_group, id
- Provider registration: NewTargetResource appended to Resources(), NewTargetDataSource appended to DataSources()
- HCL examples created for both resource and data source
- `make docs` regenerated — docs/resources/target.md and docs/data-sources/target.md created automatically
- 686 tests pass (up from 668 baseline), lint clean

## Task Commits

1. **Task 1: flashblade_target resource with drift detection** - `c82360b` (feat)
2. **Task 2: data source, provider registration, HCL examples, docs** - `2c4719b` (feat)
3. **Lint fix: remove unused targetStoreFacade** - `983ca7b` (fix)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/target_resource.go` — CRUD resource, drift detection, import
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/target_resource_test.go` — lifecycle, import, drift detection tests
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/target_data_source.go` — read-only data source by name
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/target_data_source_test.go` — basic data source test
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider.go` — NewTargetResource + NewTargetDataSource registered
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/examples/resources/flashblade_target/resource.tf` — minimal + ca_certificate_group examples
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/examples/data-sources/flashblade_target/data-source.tf` — data source lookup example
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/docs/resources/target.md` — auto-generated
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/docs/data-sources/target.md` — auto-generated

## Decisions Made

- Used flat string `ca_certificate_group` attribute (not nested object) as specified in plan — simpler HCL, consistent with schema design intent
- Drift detection logs all relevant fields (address, ca_certificate_group, status, status_details) via `tflog.Debug` with `field/was/now` keys — follows project CLAUDE.md convention
- Removed `targetStoreFacade` wrapper planned in task spec — `handlers.RegisterTargetHandlers` already returns `*targetStore` with exported `Seed` method, making the facade redundant

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed unused targetStoreFacade type**
- **Found during:** make lint (after Task 2 verification)
- **Issue:** targetStoreFacade was defined in task spec as a pattern but proved unnecessary since *targetStore returned by RegisterTargetHandlers already exposes Seed() directly
- **Fix:** Removed the struct from target_resource_test.go
- **Files modified:** internal/provider/target_resource_test.go
- **Commit:** 983ca7b

## Self-Check: PASSED

- target_resource.go: FOUND
- target_resource_test.go: FOUND
- target_data_source.go: FOUND
- target_data_source_test.go: FOUND
- provider.go (NewTargetResource registered): FOUND
- provider.go (NewTargetDataSource registered): FOUND
- examples/resources/flashblade_target/resource.tf: FOUND
- examples/data-sources/flashblade_target/data-source.tf: FOUND
- docs/resources/target.md: FOUND
- docs/data-sources/target.md: FOUND
- Commit c82360b: FOUND
- Commit 2c4719b: FOUND
- Commit 983ca7b: FOUND
- 686 tests pass: CONFIRMED
- make lint clean: CONFIRMED
- make build clean: CONFIRMED
