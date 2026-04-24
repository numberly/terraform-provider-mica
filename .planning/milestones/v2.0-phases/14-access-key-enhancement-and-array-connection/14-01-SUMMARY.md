---
phase: 14-access-key-enhancement-and-array-connection
plan: 01
subsystem: api
tags: [access-key, s3, replication, bucket-versioning, terraform-framework]

# Dependency graph
requires:
  - phase: 13-object-store-access-key
    provides: base access key resource with Computed+Sensitive secret
provides:
  - Optional secret_access_key input for cross-array key sharing
  - Bucket ValidateConfig warning for replication readiness
  - RequiresReplace on secret_access_key (changing secret forces new key)
affects: [15-replication-resources, 16-docs-workflow]

# Tech tracking
tech-stack:
  added: []
  patterns: [ValidateConfig for plan-time warnings, Optional+Computed+Sensitive for user-providable secrets]

key-files:
  created: []
  modified:
    - internal/client/models_storage.go
    - internal/provider/object_store_access_key_resource.go
    - internal/provider/object_store_access_key_resource_test.go
    - internal/testmock/handlers/object_store_access_keys.go
    - internal/provider/bucket_resource.go
    - internal/provider/bucket_resource_test.go

key-decisions:
  - "secret_access_key uses Optional+Computed+Sensitive (not WriteOnly) to preserve state for cross-resource references"
  - "RequiresReplace on secret_access_key because API does not support PATCH on secret"
  - "Bucket versioning check is a warning not error -- bucket is valid without versioning, error comes from replica link"

patterns-established:
  - "ValidateConfig for plan-time replication readiness warnings"
  - "Optional+Computed+Sensitive+RequiresReplace for user-providable immutable secrets"

requirements-completed: [AKE-01, AKE-02, AKE-03]

# Metrics
duration: 7min
completed: 2026-03-29
---

# Phase 14 Plan 01: Access Key Enhancement Summary

**Optional secret_access_key input for cross-array replication with bucket versioning warning for replication readiness**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-29T09:00:17Z
- **Completed:** 2026-03-29T09:06:51Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Access key resource accepts optional secret_access_key for cross-array credential sharing
- Mock handler honors caller-provided secret instead of generating random one
- Bucket resource emits plan-time warning when versioning is not "enabled" (replication prerequisite)
- Full backward compatibility -- all existing tests pass unchanged

## Task Commits

Each task was committed atomically:

1. **Task 1: Add optional secret_access_key to model, client, resource schema, and mock handler** - `ef12bd9` (feat)
2. **Task 2: Tests for create-with-secret and bucket versioning warning** - `1ddc579` (test)

## Files Created/Modified
- `internal/client/models_storage.go` - Added SecretAccessKey field to ObjectStoreAccessKeyPost with omitempty
- `internal/provider/object_store_access_key_resource.go` - Schema: Optional+Computed+Sensitive+RequiresReplace; Create: conditional secret pass-through
- `internal/provider/object_store_access_key_resource_test.go` - CreateWithSecret, SecretOptionalComputed, SecretRequiresReplace tests
- `internal/testmock/handlers/object_store_access_keys.go` - Mock accepts and honors optional secret_access_key in POST
- `internal/provider/bucket_resource.go` - ValidateConfig with versioning warning for replication readiness
- `internal/provider/bucket_resource_test.go` - VersioningWarning test with subtests for none/suspended/enabled/null

## Decisions Made
- secret_access_key uses Optional+Computed+Sensitive (not WriteOnly) to preserve state for cross-resource references (e.g., writing to Vault)
- RequiresReplace on secret_access_key because API does not support PATCH on secret -- changing it forces resource recreation
- Bucket versioning check is a warning, not an error -- bucket is valid without versioning; the error will come from the replica link resource in Phase 15

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Access key enhancement complete -- cross-array credential sharing is ready for replication workflow
- Bucket versioning warning will give operators early feedback before replica link creation
- Ready for Phase 14 Plan 02 (array connection data source)

## Self-Check: PASSED

All 6 files verified present. Both task commits (ef12bd9, 1ddc579) verified in git log.

---
*Phase: 14-access-key-enhancement-and-array-connection*
*Completed: 2026-03-29*
