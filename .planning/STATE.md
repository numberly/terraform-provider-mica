---
gsd_state_version: 1.0
milestone: v2.1
milestone_name: Bucket Advanced Features
status: completed
stopped_at: Completed 35-04-PLAN.md
last_updated: "2026-04-01T06:46:25.451Z"
last_activity: 2026-03-31 — Phase 32-01 complete (5 code correctness fixes)
progress:
  total_phases: 35
  completed_phases: 33
  total_plans: 77
  completed_plans: 75
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-31)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** v2.1.3 Code Review Fixes — Phase 32 next

## Current Position

Phase: 32 (Code Correctness Fixes) — complete
Plan: 32-01 complete
Status: Phase 32 done — Phase 33 next
Last activity: 2026-04-01 — Quick task 8: fix quota_limit guard in object store account PATCH

```
v2.1.3 Progress: [███                 ] 1/3 phases
```

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases defined | 3 |
| Phases complete | 0 |
| Plans defined | 3 (1 per phase, TBD detail) |
| Plans complete | 0 |
| Requirements mapped | 10/10 |
| Phase 32-code-correctness-fixes P01 | 15 | 3 tasks | 6 files |
| Phase 33-client-hardening P01 | 10 | 2 tasks | 61 files |
| Phase 33-client-hardening P02 | 10 | 1 tasks | 7 files |
| Phase 35-object-store-users P01 | 4 | 2 tasks | 4 files |
| Phase 35-object-store-users P03 | 5 | 1 tasks | 4 files |
| Phase 35 P02 | 8 | 2 tasks | 9 files |
| Phase 35-object-store-users P04 | 15 | 2 tasks | 5 files |

## Accumulated Context

### Decisions

- [quick-7]: ServerDNS struct deleted — DNS field is []NamedReference matching real API response format
- [quick-7]: directory_services added as Computed-only []NamedReference; schema v1->v2 with state upgrader chain
- [quick-7]: v1 nested DNS objects lack name field so v1->v2 upgrader resets DNS to null (refreshed on next Read)
- [Phase 32-01]: JSON tag freeze_locked_objects unchanged — only Go field name renamed to FreezeLockedObjects
- [Phase 32-01]: DiagnosticReporter.AddWarning added — backward compatible since *diag.Diagnostics already satisfies the extended interface
- [Phase 32-01]: nfs_export_policy and smb_share_policy removed from filesystem schema — had no API backing in filesystem CRUD
- [Phase 33-01]: FlashBladeTokenSource stores caller ctx at construction — oauth2.TokenSource.Token() has no context param so ctx captured at NewFlashBladeTokenSource call site
- [Phase 33-01]: RetryBaseDelay removed from Config entirely — 1000ms hardcoded as const; retry_base_delay removed from provider Terraform schema (breaking change for v2.1.3)
- [Phase 33-02]: G104 body close errors suppressed with _ assignment — errors from cleanup Body.Close() are not actionable
- [Phase 33-02]: G706/G404 in transport.go suppressed via nolint with rationale — format string is hardcoded, jitter rand needs no crypto strength
- [Phase 33-02]: noctx violations in test files fixed with real code (not nolint) — NewRequestWithContext + context.Background()
- [Phase 35-01]: GetObjectStoreUser upgraded from error-only to (*ObjectStoreUser, error) using getOneByName pattern
- [Phase 35-01]: PostObjectStoreUser upgraded to accept ObjectStoreUserPost body; EnsureObjectStoreUser passes empty struct
- [Phase 35-01]: PostObjectStoreUserPolicy sends nil body — all params are query params per API spec
- [Phase 35-03]: ImportState uses strings.SplitN(id, '/', 3) not parseCompositeID — 3-part ID where first two parts form the qualified user name (account/username)
- [Phase 35-03]: Read method checks both member.Name == userName AND policy.Name == policyName to avoid false matches when user has multiple policies
- [Phase 35]: Update stub returns AddError — all attributes are RequiresReplace so Update is never called in practice
- [Phase 35]: ImportState uses inline CRD-only null timeouts (create/read/delete) instead of shared nullTimeoutsValue which includes update key
- [Phase 35-04]: Mock handler fixed: objectStoreUserStore stores ObjectStoreUser with UUID id (was bool + empty string)
- [Phase 35-04]: ImportStateId must be explicit for name-based import when id attribute holds UUID
- [Phase 35-04]: ImportStateVerifyIdentifierAttribute=user_name for policy resource (no id field in schema)

### v2.1.3 Phase Groupings

- Phase 32: CC-01, CC-02, CC-03, CH-03, CL-01 — code correctness (typo, dead schema, diagnostic severity, unused ctx, dead helper)
- Phase 33: CH-01, CH-02, CL-02 — client hardening (OAuth2 context, RetryBaseDelay removal, linter expansion)
- Phase 34: TQ-01, TQ-02 — test quality (ExpectNonEmptyPlan removal, acceptance test expansion)

### Pending Todos

None yet.

### Quick Tasks Completed

| # | Description | Date | Commit | Directory |
|---|-------------|------|--------|-----------|
| 7 | Refactor server DNS to NamedReference, add directory_services, schema v2 | 2026-03-31 | c1df886 | [7-refactor-server-dns](./quick/7-refactor-server-dns-to-namedreference-ad/) |
| 8 | fix: skip quota_limit in object store account PATCH when unchanged | 2026-04-01 | 8612b5b | [8-fix-skip-quota-limit](./quick/8-fix-skip-quota-limit-in-object-store-acc/) |

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-04-01T06:40:56.246Z
Stopped at: Completed 35-04-PLAN.md
Resume file: None
