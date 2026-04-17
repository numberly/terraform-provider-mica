---
gsd_state_version: 1.0
milestone: v2.22.1
milestone_name: Directory Service – Array Management
status: executing
stopped_at: Completed 49-directory-service-management-49-01-PLAN.md
last_updated: "2026-04-17T07:59:39.581Z"
last_activity: 2026-04-17
progress:
  total_phases: 1
  completed_phases: 0
  total_plans: 5
  completed_plans: 2
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-17)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Phase 49 — directory-service-management

## Current Position

Phase: 49 (directory-service-management) — EXECUTING
Plan: 3 of 5
Status: Ready to execute
Last activity: 2026-04-17

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
- [Phase 49-directory-service-management]: GET returns empty list HTTP 200 on filter miss — matches getOneByName[T] not-found detection contract
- [Phase 49-directory-service-management]: PATCH **NamedReference: outer non-nil + inner nil = set to null (clear), both non-nil = set value
- [Phase 49-directory-service-management]: No DirectoryServicePost struct: /directory-services endpoint supports only GET + PATCH
- [Phase 49-directory-service-management]: DirectoryServicePatch uses **NamedReference for ca_certificate and ca_certificate_group (outer nil=omit, outer+nil inner=null, outer+non-nil inner=set)

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-04-17T07:59:39.576Z
Stopped at: Completed 49-directory-service-management-49-01-PLAN.md
Resume file: None
