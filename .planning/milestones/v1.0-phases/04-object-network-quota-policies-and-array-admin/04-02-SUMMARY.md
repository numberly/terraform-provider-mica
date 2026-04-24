---
phase: 04-object-network-quota-policies-and-array-admin
plan: "02"
subsystem: api
tags: [terraform, flashblade, object-store, iam, s3-policy, crud, import]

# Dependency graph
requires:
  - phase: 04-object-network-quota-policies-and-array-admin
    plan: "01"
    provides: OAP client layer (Post/Get/Patch/Delete ObjectStoreAccessPolicy + Rule methods) and mock handlers
provides:
  - flashblade_object_store_access_policy resource with CRUD, rename-in-place, delete guard, RequiresReplace on description
  - flashblade_object_store_access_policy_rule resource with CRUD, RequiresReplace on effect, conditions JSON round-trip
  - flashblade_object_store_access_policy data source reads OAP by name
  - 10 provider-layer tests (policy CRUD/import + data source + rule CRUD/import/conditions)
affects:
  - 04-03
  - 04-04
  - 04-05

# Tech tracking
tech-stack:
  added:
    - github.com/hashicorp/terraform-plugin-framework-validators (stringvalidator.OneOf for effect field)
  patterns:
    - OAP uses NFS export policy resource pattern (readIntoState helper, mapXxxToModel)
    - OAP rule uses SMB share policy rule pattern (composite import ID policy_name/rule_name)
    - conditions stored as types.String with jsonencode convention (json.RawMessage round-trip)
    - RequiresReplace on description (POST-only field) and effect (read-only after creation)
    - delete guard: ListObjectStoreAccessPolicyMembers before DELETE to prevent detachment errors

key-files:
  created:
    - internal/provider/object_store_access_policy_resource.go
    - internal/provider/object_store_access_policy_rule_resource.go
    - internal/provider/object_store_access_policy_data_source.go
    - internal/provider/object_store_access_policy_resource_test.go
    - internal/provider/object_store_access_policy_rule_resource_test.go
  modified:
    - internal/provider/provider.go (registration of 3 new OAP factory functions)
    - go.mod / go.sum (added terraform-plugin-framework-validators)

key-decisions:
  - "OAP description has RequiresReplace — POST-only field, API rejects PATCH with description"
  - "OAP rule effect has RequiresReplace — read-only after creation, PATCH only accepts actions/resources/conditions"
  - "OAP rule conditions stored as types.String (plain JSON string) — jsonencode() convention, marshal/unmarshal via json.RawMessage"
  - "OAP delete guard checks ListObjectStoreAccessPolicyMembers before DELETE — prevents detach errors on policies attached to buckets"
  - "OAP rule synthetic ID is policy_name/rule_name — no server-issued UUID for rules, composite key matches import format"

patterns-established:
  - "IAM-style conditions: store as types.String, convert to/from json.RawMessage at API boundary"
  - "RequiresReplace on POST-only fields: description (policy), effect (rule)"
  - "delete guard pattern: list members, error with count if non-zero, then DELETE"

requirements-completed: [OAP-01, OAP-02, OAP-03, OAP-04, OAP-05, OAR-01, OAR-02, OAR-03, OAR-04]

# Metrics
duration: 25min
completed: 2026-03-27
---

# Phase 4 Plan 02: Object Store Access Policy Summary

**IAM-style S3 access policy resource + rule resource with CRUD, RequiresReplace on description/effect, conditions JSON round-trip, delete guard, and import support**

## Performance

- **Duration:** ~25 min (interrupted, completed in follow-up)
- **Started:** 2026-03-27T16:44:36Z
- **Completed:** 2026-03-27T17:52:16Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- OAP policy resource with CRUD, rename-in-place (PATCH name), description RequiresReplace (POST-only), delete guard blocking deletion when policy is attached to buckets
- OAP rule resource with CRUD, composite import (policy_name/rule_name), effect RequiresReplace (read-only after creation), conditions JSON round-trip via types.String/json.RawMessage
- OAP data source reads policy by name returning all attributes (id, arn, enabled, is_local, policy_type, description)
- 10 tests covering policy CRUD + import + data source + rule CRUD + conditions round-trip + import

## Task Commits

1. **Task 1: OAP policy resource + rule resource with CRUD and import** - `d87e253` (feat)
2. **Task 2: OAP data source, tests, and provider registration** - `6df3b8b` (feat)

Note: `object_store_access_policy_resource_test.go` and provider.go registration were batched in adjacent plan commits (04-03 `5733159`, 04-04 `0c1cff5`).

## Files Created/Modified

- `internal/provider/object_store_access_policy_resource.go` - OAP resource: Create/Read/Update/Delete/ImportState + delete guard
- `internal/provider/object_store_access_policy_rule_resource.go` - OAP rule resource: Create/Read/Update/Delete/ImportState + conditions JSON
- `internal/provider/object_store_access_policy_data_source.go` - OAP data source: Read by name
- `internal/provider/object_store_access_policy_resource_test.go` - Policy CRUD + import + data source tests (5 test functions)
- `internal/provider/object_store_access_policy_rule_resource_test.go` - Rule CRUD + import + conditions round-trip (5 test functions)
- `internal/provider/provider.go` - Registered NewObjectStoreAccessPolicyResource, NewObjectStoreAccessPolicyRuleResource, NewObjectStoreAccessPolicyDataSource
- `go.mod` / `go.sum` - Added terraform-plugin-framework-validators dependency

## Decisions Made

- **OAP description RequiresReplace** — API is POST-only for description; PATCH does not accept this field. Changing description forces resource replacement.
- **OAP rule effect RequiresReplace** — Effect is read-only after creation per FlashBlade API (research pitfall). Only actions, resources, and conditions are patchable.
- **conditions as types.String** — JSON string via jsonencode() convention. Converted to/from json.RawMessage at API boundary. Empty `{}` or null API response becomes types.StringNull().
- **delete guard** — ListObjectStoreAccessPolicyMembers called before DELETE; returns diagnostic error if buckets still attached.
- **Synthetic rule ID** — "policy_name/rule_name" composite ID; no server-issued UUID for rules.

## Deviations from Plan

None - plan executed exactly as written. All 5 artifacts produced at or above min_lines requirements.

## Issues Encountered

Rate limit interruption during original execution. The plan was resumed and verified:
- All 5 implementation files confirmed present and correct
- 10 OAP-specific tests pass (`go test ./internal/provider/ -run "TestObjectStoreAccessPolicy"`)
- Full suite (136 tests across 5 packages) green (`go test ./... -count=1`)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- OAP resource layer complete; ready for Phase 4 Plan 03 (Network Access Policy) and Plan 04 (Quotas)
- Conditions JSON pattern established for future IAM-style resources
- delete guard pattern documented for policies with bucket attachments

---
*Phase: 04-object-network-quota-policies-and-array-admin*
*Completed: 2026-03-27*
