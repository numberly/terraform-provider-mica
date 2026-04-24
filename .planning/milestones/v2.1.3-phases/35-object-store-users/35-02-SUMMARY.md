---
phase: 35-object-store-users
plan: "02"
subsystem: provider
tags: [object-store-users, terraform-resource, terraform-datasource, crd, drift-detection, import]

dependency_graph:
  requires:
    - phase: 35-01
      provides: GetObjectStoreUser, PostObjectStoreUser, DeleteObjectStoreUser, ObjectStoreUser, ObjectStoreUserPost structs
  provides:
    - flashblade_object_store_user resource (Create, Read, Delete, ImportState, UpgradeState)
    - flashblade_object_store_user data source (Read)
    - both registered in provider.go
    - example HCL files and docs generated
  affects:
    - 35-03 (ObjectStoreUserPolicy resource — same account/username pattern)

tech-stack:
  added: []
  patterns:
    - CRD resource with no Update stub returning AddError
    - FullAccess optional pointer in Create: only send body field when explicitly set (not null/unknown)
    - Drift detection via tflog.Warn comparing state vs API value in Read
    - ImportState with inline nullTimeoutsValue-equivalent (CRD keys only: create/read/delete)
    - mapObjectStoreUserToModel helper: pure value mapper, no diagnostics needed (no nested objects)

key-files:
  created:
    - internal/provider/object_store_user_resource.go
    - internal/provider/object_store_user_data_source.go
    - examples/resources/flashblade_object_store_user/resource.tf
    - examples/resources/flashblade_object_store_user/import.sh
    - examples/data-sources/flashblade_object_store_user/data-source.tf
  modified:
    - internal/provider/provider.go
    - ROADMAP.md
    - docs/ (auto-generated via make docs)

key-decisions:
  - "Update stub returns AddError — all attributes are RequiresReplace so no in-place update is ever needed; stub prevents silent no-op"
  - "FullAccess uses boolplanmodifier.UseStateForUnknown — API always returns a value, no drift on computed default"
  - "ImportState uses inline CRD-only timeout null object (no update key) — avoids reusing the 4-key nullTimeoutsValue helper"
  - "ROADMAP.md updated synchronously with resource implementation per CLAUDE.md mandatory convention"

patterns-established:
  - "CRD resource null timeouts object uses only the keys matching the schema (create/read/delete — no update key)"

requirements-completed: [OSU-01, OSU-02, OSU-03, OSU-04, OSU-07]

duration: 8min
completed: 2026-03-31
---

# Phase 35 Plan 02: Object Store User Resource and Data Source Summary

**flashblade_object_store_user CRD resource with full_access drift detection, ImportState, and data source — registered in provider and docs generated.**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-03-31T06:25:00Z
- **Completed:** 2026-03-31T06:33:00Z
- **Tasks:** 2
- **Files modified:** 9 (2 new Go files, 3 new HCL examples, provider.go, ROADMAP.md, docs regenerated)

## Accomplishments

- CRD resource `flashblade_object_store_user` with Create (optional full_access body param), Read (tflog.Warn drift), Delete (idempotent), ImportState (account/username), UpgradeState (v0 empty map)
- Data source `flashblade_object_store_user` reading id, name, full_access from API
- Both registered in provider.go alongside existing object_store_* entries
- ROADMAP.md updated: Object Store Users moved from Not Implemented to Storage/Implemented (Done)
- Provider docs regenerated via make docs

## Task Commits

1. **Task 1: Implement flashblade_object_store_user resource** - `5dad19a` (feat)
2. **Task 2: Register data source, provider, examples, docs** - `0e98c29` (feat)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_user_resource.go` — CRD resource with full CRUD+Import+UpgradeState
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_user_data_source.go` — read-only data source
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider.go` — NewObjectStoreUserResource and NewObjectStoreUserDataSource registered
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/examples/resources/flashblade_object_store_user/resource.tf` — HCL example
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/examples/resources/flashblade_object_store_user/import.sh` — import command
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/examples/data-sources/flashblade_object_store_user/data-source.tf` — data source example
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/ROADMAP.md` — Object Store Users marked Done
- `docs/` — regenerated via make docs

## Decisions Made

- Update stub returns AddError: all mutable attributes are RequiresReplace, so Update is never called in practice. The stub prevents silent no-ops if that contract were violated.
- FullAccess uses `boolplanmodifier.UseStateForUnknown` as it is Optional+Computed and the API always returns the current value — avoids spurious plan diffs.
- ImportState uses inline CRD-only null timeouts object (`create/read/delete` keys only) instead of the shared `nullTimeoutsValue()` helper which includes `update` key not present in this schema.
- ROADMAP.md updated synchronously as mandated by CLAUDE.md; make docs run to keep docs/ consistent.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `flashblade_object_store_user` resource ready for use in acceptance tests and production workflows
- Phase 35-03 (ObjectStoreUserPolicy resource) can now use the same account/username pattern
- No blockers

---
*Phase: 35-object-store-users*
*Completed: 2026-03-31*
