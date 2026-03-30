---
phase: 23-bucket-inline-attributes
plan: 01
subsystem: api
tags: [terraform, flashblade, bucket, eradication, object-lock, public-access]

requires:
  - phase: none
    provides: existing bucket resource CRUD and client models
provides:
  - EradicationConfig, ObjectLockConfig, PublicAccessConfig client structs
  - Bucket resource schema blocks for all 3 configs + public_status
  - Create/Update/Read/Import wiring for inline config attributes
affects: [23-bucket-inline-attributes]

tech-stack:
  added: []
  patterns: [SingleNestedAttribute with Optional+Computed for API config objects, extract helper pattern for nested object to client struct conversion]

key-files:
  created: []
  modified:
    - internal/client/models_storage.go
    - internal/provider/bucket_resource.go

key-decisions:
  - "public_access_config excluded from POST (API spec constraint, PATCH only)"
  - "Bool fields in nested configs use omitempty since API treats entire config object as unit on PATCH"
  - "No enum validators added yet -- kept simple for this plan"

patterns-established:
  - "extract*Config pattern: check IsNull/IsUnknown, return nil or built client struct"
  - "*ConfigAttrTypes + map*ConfigToObject pattern for nested config mapping"

requirements-completed: [BKT-01, BKT-02, BKT-03, BKT-04]

duration: 3min
completed: 2026-03-30
---

# Phase 23 Plan 01: Bucket Inline Config Attributes Summary

**Eradication, object lock, and public access config blocks added to bucket resource with full Create/Read/Update/Import wiring**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-30T09:58:23Z
- **Completed:** 2026-03-30T10:00:59Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Added 3 new config structs (EradicationConfig, ObjectLockConfig, PublicAccessConfig) to client models
- Extended Bucket/BucketPost/BucketPatch with all new fields following existing pointer-for-PATCH pattern
- Added 3 SingleNestedAttribute schema blocks + 1 computed StringAttribute (public_status) to bucket resource
- Wired eradication_config and object_lock_config into Create, all 3 configs into Update with IsUnknown() guards
- Added helper functions for attr types, object mapping, and config extraction

## Task Commits

Each task was committed atomically:

1. **Task 1: Add client model structs** - `5e2fa92` (feat)
2. **Task 2: Extend bucket resource schema, model, mapping, and CRUD** - `7deb5f6` (feat)

## Files Created/Modified
- `internal/client/models_storage.go` - Added EradicationConfig, ObjectLockConfig, PublicAccessConfig structs; extended Bucket, BucketPost, BucketPatch
- `internal/provider/bucket_resource.go` - Added schema blocks, model fields, mapping helpers, extract functions, Create/Update wiring

## Decisions Made
- public_access_config excluded from POST body (API spec constraint -- PATCH only)
- Bool fields in nested configs use omitempty since API treats entire config as a unit on PATCH
- No enum validators added yet -- simplicity first, validators can come later

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Client models and bucket resource fully extended
- Unit tests will need mock updates to include new config fields (handled by plan 02)
- Acceptance tests can verify against real FlashBlade API

---
*Phase: 23-bucket-inline-attributes*
*Completed: 2026-03-30*

## Self-Check: PASSED
