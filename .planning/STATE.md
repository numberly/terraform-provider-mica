---
gsd_state_version: 1.0
milestone: v2.1
milestone_name: Bucket Advanced Features
status: Roadmap ready, awaiting plan-phase 32
stopped_at: Completed 32-01-PLAN.md
last_updated: "2026-03-31T16:19:25.579Z"
last_activity: 2026-03-31 — Roadmap created for v2.1.3
progress:
  total_phases: 34
  completed_phases: 31
  total_plans: 71
  completed_plans: 69
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
Last activity: 2026-03-31 — Phase 32-01 complete (5 code correctness fixes)

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

## Accumulated Context

### Decisions

- [quick-7]: ServerDNS struct deleted — DNS field is []NamedReference matching real API response format
- [quick-7]: directory_services added as Computed-only []NamedReference; schema v1->v2 with state upgrader chain
- [quick-7]: v1 nested DNS objects lack name field so v1->v2 upgrader resets DNS to null (refreshed on next Read)
- [Phase 32-01]: JSON tag freeze_locked_objects unchanged — only Go field name renamed to FreezeLockedObjects
- [Phase 32-01]: DiagnosticReporter.AddWarning added — backward compatible since *diag.Diagnostics already satisfies the extended interface
- [Phase 32-01]: nfs_export_policy and smb_share_policy removed from filesystem schema — had no API backing in filesystem CRUD

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

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-03-31T16:19:25.569Z
Stopped at: Completed 32-01-PLAN.md
Resume file: None
