---
gsd_state_version: 1.0
milestone: v2.22.1
milestone_name: Directory Service – Array Management
status: Defining plan
stopped_at: Phase 49 context gathered
last_updated: "2026-04-17T07:31:12.137Z"
last_activity: 2026-04-17 — Roadmap created for milestone v2.22.1
progress:
  total_phases: 1
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-17)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Directory Service – Array Management (v2.22.1)

## Current Position

Phase: 49 — Directory Service Management
Plan: — (defining plan)
Status: Defining plan
Last activity: 2026-04-17 — Roadmap created for milestone v2.22.1

```
Progress: [                    ] 0% (0/1 phases)
```

## Accumulated Context

### Decisions

- Shared Python lib in .claude/skills/_shared/ (cross-skill)
- PYTHONPATH=.claude/skills required for _shared imports
- Python 3.10+ stdlib only, no external deps
- Swagger inaccuracies tracked in known_discrepancies.md (living doc)
- DSM resource: singleton PATCH-only (no POST, no DELETE), backed by /directory-services?names=management
- DSM resource: bind_password is sensitive, write-only, never returned by API
- DSM resource: ca_certificate and ca_certificate_group use NamedReference pattern
- DSM resource: management sub-object holds user_login_attribute, user_object_class, ssh_public_key_attribute
- DSM resource: Delete = PATCH reset (enabled=false, empty uris, nil references)
- DSM resource: Import by name "management", nullTimeoutsValue(), bind_password left empty

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-04-17T07:31:12.131Z
Stopped at: Phase 49 context gathered
Resume file: .planning/phases/49-directory-service-management/49-CONTEXT.md
