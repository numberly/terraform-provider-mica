---
phase: 11-test-hardening-and-validators
plan: 01
subsystem: testing
tags: [terraform-validators, stringvalidator, input-validation, schema]

# Dependency graph
requires:
  - phase: 08-remaining-resources
    provides: Resource schemas for S3 export policy rule, NFS, SMB, SMTP, virtual host
provides:
  - Custom AlphanumericValidator and HostnameNoDotValidator
  - Plan-time enum validation on 6 resource schemas
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: [custom-validator-interface, table-driven-validator-tests]

key-files:
  created:
    - internal/provider/validators.go
    - internal/provider/validators_test.go
  modified:
    - internal/provider/s3_export_policy_rule_resource.go
    - internal/provider/network_access_policy_rule_resource.go
    - internal/provider/smb_client_policy_rule_resource.go
    - internal/provider/nfs_export_policy_rule_resource.go
    - internal/provider/array_smtp_resource.go
    - internal/provider/object_store_virtual_host_resource.go

key-decisions:
  - "Custom validators implement validator.String interface directly (no external library needed)"

patterns-established:
  - "Custom validators: implement Description/MarkdownDescription/ValidateString, export factory function"
  - "Enum validators: use stringvalidator.OneOf from terraform-plugin-framework-validators"

requirements-completed: [VAL-01, VAL-02]

# Metrics
duration: 3min
completed: 2026-03-29
---

# Phase 11 Plan 01: Input Validators Summary

**Custom alphanumeric/hostname validators and OneOf enum validators wired into 6 resource schemas for plan-time input rejection**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-29T06:40:11Z
- **Completed:** 2026-03-29T06:43:14Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Created AlphanumericValidator (rejects non-alphanumeric S3 export policy rule names) and HostnameNoDotValidator (rejects dots in virtual host hostnames)
- 19 table-driven unit tests covering accept/reject cases for both custom validators
- Wired stringvalidator.OneOf validators for effect, encryption, permission, access, and encryption_mode fields across 6 resources
- All 262 existing tests continue to pass

## Task Commits

Each task was committed atomically:

1. **Task 1: Create custom name format validators and unit tests** - `b40a2da` (feat)
2. **Task 2: Wire validators into resource schemas** - `ebea207` (feat)

## Files Created/Modified
- `internal/provider/validators.go` - AlphanumericValidator and HostnameNoDotValidator implementations
- `internal/provider/validators_test.go` - Table-driven tests for both validators (19 cases)
- `internal/provider/s3_export_policy_rule_resource.go` - AlphanumericValidator on name, OneOf on effect
- `internal/provider/network_access_policy_rule_resource.go` - OneOf on effect
- `internal/provider/smb_client_policy_rule_resource.go` - OneOf on encryption, permission
- `internal/provider/nfs_export_policy_rule_resource.go` - OneOf on permission, access
- `internal/provider/array_smtp_resource.go` - OneOf on encryption_mode
- `internal/provider/object_store_virtual_host_resource.go` - HostnameNoDotValidator on hostname

## Decisions Made
- Custom validators implement the validator.String interface directly rather than wrapping external library types, keeping the dependency minimal

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All plan-time validators are in place
- Ready for additional test hardening plans in this phase

---
*Phase: 11-test-hardening-and-validators*
*Completed: 2026-03-29*
