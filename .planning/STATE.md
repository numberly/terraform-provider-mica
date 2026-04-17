---
gsd_state_version: 1.0
milestone: v2.22.1
milestone_name: Directory Service – Array Management
status: verifying
stopped_at: Completed 49-directory-service-management-49-05-PLAN.md
last_updated: "2026-04-17T08:17:35.714Z"
last_activity: 2026-04-17
progress:
  total_phases: 1
  completed_phases: 1
  total_plans: 5
  completed_plans: 5
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-17)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Phase 49 — directory-service-management

## Current Position

Phase: 49 (directory-service-management) — EXECUTING
Plan: 5 of 5
Status: Phase complete — ready for verification
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
- [Phase 49-directory-service-management]: DSM data source: namedRefAttrTypes/namedRefObjectValue helpers in data source file, 2 interface assertions, no name/bind_password per D-06
- [Phase 49-directory-service-management]: DSM resource: no name attribute in schema (D-01), hardcoded to 'management'
- [Phase 49-directory-service-management]: DSM resource: Delete sends full-reset PATCH (D-02), bind_password omitted
- [Phase 49-directory-service-management]: Import key is literal 'management' (singleton resource)
- [Phase 49-directory-service-management]: resource.tf uses ldaps:// per enterprise expectation, bind_password via var

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-04-17T08:17:35.708Z
Stopped at: Completed 49-directory-service-management-49-05-PLAN.md
Resume file: None
