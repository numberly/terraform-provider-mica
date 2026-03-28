---
phase: 05-quality-hardening
plan: 01
subsystem: provider/client
tags: [validators, plan-modifiers, error-helpers, schema-testing]
dependency_graph:
  requires: []
  provides:
    - IsConflict and IsUnprocessable error helpers in client layer
    - stringvalidator.OneOf on bucket versioning field
    - int64validator.AtLeast(0) on quota_group and quota_user quota fields
    - PlanModifier assertion tests for all 19 resources
    - Validator rejection tests for 4 validated fields
  affects:
    - internal/client/errors.go
    - internal/provider/bucket_resource.go
    - internal/provider/quota_group_resource.go
    - internal/provider/quota_user_resource.go
    - all 19 internal/provider/*_resource_test.go files
tech_stack:
  added:
    - terraform-plugin-framework-validators/stringvalidator (already present, now used in bucket_resource.go)
    - terraform-plugin-framework-validators/int64validator (new import in quota resources)
  patterns:
    - validator.StringRequest / validator.StringResponse for inline validator testing
    - validator.Int64Request / validator.Int64Response for inline int64 validator testing
    - schema.StringAttribute / schema.BoolAttribute / schema.Int64Attribute type assertions for PlanModifier testing
key_files:
  created:
    - internal/client/errors_test.go
  modified:
    - internal/client/errors.go (added IsConflict, IsUnprocessable)
    - internal/provider/bucket_resource.go (added versioning validator)
    - internal/provider/quota_group_resource.go (added quota AtLeast validator)
    - internal/provider/quota_user_resource.go (added quota AtLeast validator)
    - internal/provider/bucket_resource_test.go (PlanModifiers + VersioningValidator tests)
    - internal/provider/object_store_access_policy_rule_resource_test.go (PlanModifiers + EffectValidator tests)
    - internal/provider/quota_group_resource_test.go (PlanModifiers + QuotaValidator tests)
    - internal/provider/quota_user_resource_test.go (PlanModifiers + QuotaValidator tests)
    - internal/provider/network_access_policy_rule_resource_test.go (PlanModifiers test)
    - internal/provider/nfs_export_policy_rule_resource_test.go (PlanModifiers test)
    - internal/provider/smb_share_policy_rule_resource_test.go (PlanModifiers test)
    - internal/provider/snapshot_policy_resource_test.go (PlanModifiers test)
    - internal/provider/snapshot_policy_rule_resource_test.go (PlanModifiers test)
    - internal/provider/filesystem_resource_test.go (PlanModifiers test)
    - internal/provider/nfs_export_policy_resource_test.go (PlanModifiers test)
    - internal/provider/smb_share_policy_resource_test.go (PlanModifiers test)
    - internal/provider/array_dns_resource_test.go (PlanModifiers test)
    - internal/provider/array_ntp_resource_test.go (PlanModifiers test)
    - internal/provider/array_smtp_resource_test.go (PlanModifiers test)
    - internal/provider/network_access_policy_resource_test.go (PlanModifiers test)
    - internal/provider/object_store_access_policy_resource_test.go (PlanModifiers test)
    - internal/provider/object_store_access_key_resource_test.go (PlanModifiers test)
    - internal/provider/object_store_account_resource_test.go (PlanModifiers test)
decisions:
  - IsConflict and IsUnprocessable follow exact pattern of IsNotFound — consistent error helper API
  - Validator tests call ValidateString/ValidateInt64 directly without spinning up a full provider — fast unit tests
  - Snapshot policy enabled field has no PlanModifier (uses booldefault.StaticBool) — test correctly asserts is_local and retention_lock instead
  - Only 3 resources received new validators (bucket, quota_group, quota_user) — others lack clear enum/range fields per RESEARCH guidance
metrics:
  duration: ~25 minutes
  completed: "2026-03-26"
  tasks: 2
  files: 24
---

# Phase 5 Plan 01: Validators, Plan Modifier Tests, and Error Helpers Summary

Added IsConflict/IsUnprocessable error helpers to client layer, added enum/range validators to bucket versioning and quota resources, and added PlanModifier assertion tests plus validator rejection tests across all 19 resources.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Add IsConflict/IsUnprocessable error helpers with TDD tests | d887c8c |
| 2 | Add validators to 3 resource schemas + PlanModifier tests for all 19 + validator rejection tests | 02dc586 |

## What Was Built

### Error Helpers (QUA-01 partial)

Two new functions in `internal/client/errors.go`:
- `IsConflict(err error) bool` — true when err is `*APIError` with HTTP 409
- `IsUnprocessable(err error) bool` — true when err is `*APIError` with HTTP 422

Both follow the exact pattern of `IsNotFound`. Both handle nil safely. Tested with 7 unit tests in `internal/client/errors_test.go`.

### Validators Added (QUA-02)

| Resource | Field | Validator |
|----------|-------|-----------|
| bucket | versioning | stringvalidator.OneOf("none", "enabled", "suspended") |
| quota_group | quota | int64validator.AtLeast(0) |
| quota_user | quota | int64validator.AtLeast(0) |

### Plan Modifier Tests (QUA-03)

All 19 resource test files now contain a `TestUnit_{Resource}_PlanModifiers` function asserting every `RequiresReplace` and `UseStateForUnknown` plan modifier present in the production schema.

### Validator Rejection Tests (QUA-02)

| Test | File | Rejects | Accepts |
|------|------|---------|---------|
| TestUnit_OAPRule_EffectValidator | object_store_access_policy_rule_resource_test.go | "invalid" | "allow", "deny" |
| TestUnit_Bucket_VersioningValidator | bucket_resource_test.go | "invalid" | "none", "enabled", "suspended" |
| TestUnit_QuotaGroup_QuotaValidator | quota_group_resource_test.go | -1 | 0, 1048576 |
| TestUnit_QuotaUser_QuotaValidator | quota_user_resource_test.go | -1 | 0, 1048576 |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Snapshot policy PlanModifiers test corrected**
- **Found during:** Task 2 — first test run
- **Issue:** Test asserted `enabled` field has PlanModifier, but `enabled` uses `booldefault.StaticBool(true)` with no PlanModifier
- **Fix:** Changed test to assert `is_local` and `retention_lock` fields which correctly have `UseStateForUnknown` modifiers
- **Files modified:** `internal/provider/snapshot_policy_resource_test.go`
- **Commit:** Included in 02dc586

## Test Results

- Pre-plan tests: ~147
- Post-plan tests: 177
- New tests: 30 (7 client error helpers + 19 PlanModifiers + 4 validator rejection)
- Regressions: 0

## Self-Check: PASSED

Verified:
- `internal/client/errors.go` — exists, contains `IsConflict`
- `internal/client/errors_test.go` — exists, contains `TestUnit_IsConflict`
- `internal/provider/bucket_resource.go` — contains `stringvalidator.OneOf`
- `internal/provider/quota_group_resource.go` — contains `int64validator.AtLeast`
- `internal/provider/quota_user_resource.go` — contains `int64validator.AtLeast`
- Git commits d887c8c and 02dc586 exist
- All 177 tests pass: `go test ./internal/... -count=1`
