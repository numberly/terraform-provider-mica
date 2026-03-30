---
gsd_state_version: 1.0
milestone: v2.1
milestone_name: Bucket Advanced Features
status: defining_requirements
stopped_at: null
last_updated: "2026-03-30T08:00:00.000Z"
last_activity: 2026-03-30 — Milestone v2.1 started
progress:
  total_phases: 0
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-30)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Milestone v2.1 — Bucket Advanced Features

## Current Position

Phase: Not started (defining requirements)
Plan: —
Status: Defining requirements
Last activity: 2026-03-30 — Milestone v2.1 started

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

### Pending Todos

None yet.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-03-30
Stopped at: null
Resume file: None
