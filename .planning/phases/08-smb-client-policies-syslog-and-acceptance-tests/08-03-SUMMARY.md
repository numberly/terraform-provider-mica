---
phase: 08-smb-client-policies-syslog-and-acceptance-tests
plan: "03"
subsystem: testing
tags: [terraform, acceptance-tests, smb-client-policy, syslog, s3-export-policy, virtual-host, flashblade]

# Dependency graph
requires:
  - phase: 08-01
    provides: SMB client policy and rule resources
  - phase: 08-02
    provides: Syslog server resource
  - phase: 07-02
    provides: S3 export policy and rule resources
  - phase: 07-03
    provides: Virtual host resource
provides:
  - Acceptance test HCL for SMB client policy + rule (live FlashBlade validated)
  - Acceptance test HCL for syslog server (live FlashBlade validated)
  - Acceptance test HCL for S3 export policy + rule + virtual host (live FlashBlade validated)
  - Confirmed API constraints not captured in research (S3 rule name, attached_servers, pure:S3Access action)
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Acceptance test HCL in tmp/test-purestorage/ with test-gule-* naming convention
    - destroy_eradicate_on_delete = true for file system resources
    - Acceptance cycle: tofu validate → plan → apply → plan (0 changes) → destroy

key-files:
  created:
    - /home/gule/Workspace/team-infrastructure/tmp/test-purestorage/test_smb_client_policies.tf
    - /home/gule/Workspace/team-infrastructure/tmp/test-purestorage/test_syslog_servers.tf
    - /home/gule/Workspace/team-infrastructure/tmp/test-purestorage/test_v11_server_exports.tf
  modified:
    - internal/client/s3_export_policies.go
    - internal/provider/object_store_virtual_host_resource.go
    - internal/provider/s3_export_policy_rule_resource.go

key-decisions:
  - "S3 export policy rule name must be alphanumeric only (no hyphens), Required field, ?names= query param on POST"
  - "S3 export policy rule only valid action is pure:S3Access (not s3:* wildcards)"
  - "Virtual host attached_servers listdefault removed — API auto-attaches default server on creation"

patterns-established:
  - "Acceptance tests live in tmp/test-purestorage/ with test-gule-* prefix for all resource names"
  - "Full cycle validation: apply + idempotency plan (0 changes) + destroy"
  - "API discovery: acceptance testing against live FlashBlade surfaces constraints not in API docs"

requirements-completed: [EXP-03]

# Metrics
duration: 30min
completed: 2026-03-28
---

# Phase 8 Plan 03: Acceptance Tests Summary

**Live FlashBlade acceptance test cycle validated all 26 v1.1 resources — uncovering 3 API constraints fixed inline (S3 rule name format, S3 action value, virtual host server auto-attachment)**

## Performance

- **Duration:** ~30 min
- **Started:** 2026-03-28T16:30:00Z
- **Completed:** 2026-03-28T16:58:53Z
- **Tasks:** 2 (Task 1 auto + Task 2 human-verify checkpoint)
- **Files modified:** 6 (3 created, 3 fixed)

## Accomplishments

- Wrote HCL acceptance test configurations for SMB client policy + rule, syslog server, and S3 export policy + rule + virtual host
- Ran full apply/plan/destroy cycle against live FlashBlade — all 26 resources created and destroyed without errors
- Discovered and fixed 3 API behavioral constraints not captured during research phase

## Task Commits

Each task was committed atomically:

1. **Task 1: Write acceptance test HCL configurations** - `f3894ba` (feat)
2. **Task 2: Acceptance test fixes (live FlashBlade)** - `fa8cd8a` (fix)

**Plan metadata:** (this commit)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage/test_smb_client_policies.tf` — Acceptance test: SMB client policy + rule resource
- `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage/test_syslog_servers.tf` — Acceptance test: syslog server resource + data source
- `/home/gule/Workspace/team-infrastructure/tmp/test-purestorage/test_v11_server_exports.tf` — Acceptance test: S3 export policy + rule + virtual host
- `internal/client/s3_export_policies.go` — Fixed S3 export policy rule POST to use ?names= query param
- `internal/provider/object_store_virtual_host_resource.go` — Removed listdefault for attached_servers
- `internal/provider/s3_export_policy_rule_resource.go` — Fixed actions validator: only pure:S3Access allowed

## Decisions Made

- S3 export policy rule name is a Required field (not Computed) and must be alphanumeric — the API uses ?names= on POST, not a body name field
- The only valid S3 action is `pure:S3Access` — s3:* wildcard actions are rejected by the FlashBlade API
- Virtual host `attached_servers` listdefault removed: the API automatically attaches the default server on creation; providing an explicit empty list conflicts with the API's auto-assignment behavior

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] S3 export policy rule name must be alphanumeric (no hyphens)**
- **Found during:** Task 2 (acceptance test against live FlashBlade)
- **Issue:** Rule name `allow-all` rejected by API — names must be alphanumeric only; POST uses ?names= query param
- **Fix:** Updated S3 export policy rule resource to mark name as Required, added ?names= to POST call; renamed test resource to `allowall`
- **Files modified:** internal/client/s3_export_policies.go, internal/provider/s3_export_policy_rule_resource.go
- **Verification:** tofu apply succeeded with alphanumeric rule name
- **Committed in:** fa8cd8a

**2. [Rule 1 - Bug] S3 export policy rule action value incorrect**
- **Found during:** Task 2 (acceptance test against live FlashBlade)
- **Issue:** Action `s3:*` rejected by FlashBlade API — only valid value is `pure:S3Access`
- **Fix:** Updated actions validator and test HCL to use `pure:S3Access`
- **Files modified:** internal/provider/s3_export_policy_rule_resource.go, test_v11_server_exports.tf
- **Verification:** tofu apply succeeded with correct action value
- **Committed in:** fa8cd8a

**3. [Rule 1 - Bug] Virtual host attached_servers listdefault conflicts with API auto-attachment**
- **Found during:** Task 2 (acceptance test against live FlashBlade)
- **Issue:** API auto-attaches default server on virtual host creation; listdefault static empty list caused drift on subsequent plan
- **Fix:** Removed `listdefault.StaticValue(empty list)` — the attribute now reflects API state directly
- **Files modified:** internal/provider/object_store_virtual_host_resource.go
- **Verification:** tofu plan after apply shows 0 changes (idempotency confirmed)
- **Committed in:** fa8cd8a

---

**Total deviations:** 3 auto-fixed (all Rule 1 - Bug)
**Impact on plan:** All auto-fixes surfaced by live API testing. Critical for provider correctness — these would have caused silent failures in production use. No scope creep.

## Issues Encountered

The live FlashBlade API revealed three behavioral constraints absent from API documentation:
1. S3 rule names are alphanumeric-only with ?names= POST semantics
2. S3 actions only accept `pure:S3Access` (no wildcard or AWS-style actions)
3. Virtual host server attachment is automatic — explicit empty list causes drift

All were resolved inline during the acceptance test cycle.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 8 complete: all v1.1 resources (SMB client policies, syslog server, S3 export policies, virtual host) are live-validated
- Milestone v1.1 (Servers & Exports) is fully implemented and acceptance-tested
- The provider is ready for release — 26 resources and data sources work against live FlashBlade with 0 drift after apply

---
*Phase: 08-smb-client-policies-syslog-and-acceptance-tests*
*Completed: 2026-03-28*
