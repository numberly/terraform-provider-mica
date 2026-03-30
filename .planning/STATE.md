---
gsd_state_version: 1.0
milestone: v2.1
milestone_name: Bucket Advanced Features
status: completed
stopped_at: Completed 26-03-PLAN.md
last_updated: "2026-03-30T11:24:16.160Z"
last_activity: 2026-03-30 — Completed 26-03 QoS policy resource, member resource, and data source
progress:
  total_phases: 27
  completed_phases: 25
  total_plans: 63
  completed_plans: 61
  percent: 97
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-30)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Phase 26 — Audit Filters & QoS Policies (v2.1)

## Current Position

Phase: 26 of 27 (Audit Filters & QoS Policies)
Plan: 3 of 3 in current phase
Status: Phase Complete
Last activity: 2026-03-30 — Completed 26-03 QoS policy resource, member resource, and data source

Progress: [██████████] 97% (61/63 plans complete)

## Performance Metrics

**Velocity (from v1.0 through v2.0.1):**
- Total plans completed: 54
- Phases completed: 22
- Total execution time: ~56 hours

*Updated after each plan completion*

## Accumulated Context

### Decisions

- [v2.0.1]: Access key name param required when providing secret_access_key (API constraint)
- [v2.0.1]: Remote credentials name must be <remote-name>/<credentials-name> format
- [v2.0.1]: Bucket replica link DELETE uses ?ids= not bucket names
- [v2.0.1]: Volatile attrs (lag, recovery_point, backlog) should NOT use UseStateForUnknown
- [v2.0.1]: Bucket Update must check IsUnknown() before including fields in PATCH
- [v2.1]: Bucket inline attrs (eradication, object lock, public access) as schema extensions, not separate resources
- [v2.1]: QoS + audit filters combined in Phase 26 (coarse granularity)
- [v2.1]: Testing + docs consolidated into Phase 27 (cross-cutting)
- [Phase 23]: public_access_config excluded from POST (API spec constraint, PATCH only)
- [Phase 23]: Mock handler defaults: 24h eradication delay, retention-based mode (matches real API)
- [Phase 24-lifecycle-rules]: GetLifecycleRule uses bucket_names filter + iteration (not getOneByName) — API uses bucket_names param
- [Phase 24-lifecycle-rules]: Composite name format bucketName/ruleID for PATCH and DELETE identification
- [Phase 24-lifecycle-rules]: Optional int64 fields mapped to null when API returns 0 (preserves Terraform null semantics)
- [Phase 24-lifecycle-rules]: PATCH sends only changed fields via pointer comparison
- [Phase 25-bucket-access-policies]: Policy store keyed by bucket name (one policy per bucket)
- [Phase 25-bucket-access-policies]: Rules stored inside policy object, separate endpoint handlers
- [Phase 25]: No Update method for bucket access policies (no PATCH endpoint, RequiresReplace)
- [Phase 25]: CRD-only resources use 3-key timeouts (no update key) to match schema
- [Phase 26]: QosPolicyMember/QosPolicyMemberPost types instead of PolicyMember/PolicyMemberPost to avoid collision with existing PolicyMember in models_common.go
- [Phase 26]: Bucket name as single-string import ID for audit filters (one per bucket)
- [Phase 26]: QoS policy name uses RequiresReplace (rename via PATCH not exposed to avoid drift)
- [Phase 26]: QoS member CRD-only with 3-key timeouts, composite "policyName/memberName" import

### Pending Todos

None yet.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-03-30T11:22:00Z
Stopped at: Completed 26-03-PLAN.md
Resume file: None
