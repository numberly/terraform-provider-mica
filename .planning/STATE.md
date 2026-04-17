---
gsd_state_version: 1.0
milestone: v2.22.2
milestone_name: Directory Service Roles & Role Mappings
status: active
stopped_at: v2.22.1 archived; v2.22.2 (Phase 50) complete and ready for audit
last_updated: "2026-04-17T12:15:00Z"
last_activity: 2026-04-17
last_shipped_milestone: v2.22.1
progress:
  total_phases: 1
  completed_phases: 1
  total_plans: 5
  completed_plans: 5
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-17)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Phase 50 — directory-service-roles-role-mappings

## Current Position

Phase: 50
Plan: Not started
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
- [Phase 50-roadmap]: DSR resource: full CRUD (POST/GET/PATCH/DELETE) — NOT singleton like DSM
- [Phase 50-roadmap]: DSR resource: role attribute is NamedReference, triggers RequiresReplace on change
- [Phase 50-roadmap]: DSR resource: management_access_policies is computed-only list (populated by API via membership associations)
- [Phase 50-roadmap]: DSRM resource: composite ID format policy_name:role_name — follows qos_policy_member/tls_policy_member/certificate_group_member pattern
- [Phase 50-roadmap]: DSRM resource: GET/POST/DELETE only — no PATCH, both fields trigger RequiresReplace
- [Phase 50-roadmap]: DSRM resource: Read calls GET with policy_names= and role_names= query params; empty list → RemoveResource
- [Phase 50-roadmap]: Test baseline: 798 (v2.22.1) → target ≥ 812 (14 new tests: 5 client + 6 resource + 1 data source + 2 margin)
- [Phase 50-02]: DSRM POST idempotent (Q3 resolved): 200-always, create-or-return, Terraform replays never 409
- [Phase 50-02]: DSR PATCH readonly guard: raw JSON decode first to detect management_access_policies before typed decode
- [Phase 50]: POST /directory-services/roles has no names query param — name is server-generated from management_access_policies
- [Phase 50]: DirectoryServiceRolePatch omits ManagementAccessPolicies — readonly on PATCH per swagger
- [Phase 50]: DSRM composite key: role_name/policy_name (role first, D-05) so SplitN works with colons/slashes in policy name
- [Phase 50-directory-service-roles-role-mappings]: Composite ID puts role_name FIRST (role/policy) so SplitN correctly handles built-in policy names containing : and / like pure:policy/array_admin
- [Phase 50-directory-service-roles-role-mappings]: Used nullTimeoutsValueCRD() in ImportState for CRD-only DSRM resource (no Update timeout)
- [Phase 50-directory-service-roles-role-mappings]: D-02 confirmed: role attribute Computed-only, SC-3 replacement trigger on management_access_policies via listplanmodifier.RequiresReplace()
- [Phase 50-directory-service-roles-role-mappings]: DSRM resource was already registered in provider.go by Plan 50-04 executor; only DSR resource + data source needed adding in Plan 50-05
- [Phase 50-directory-service-roles-role-mappings]: errcheck lint violations fixed in client test files (w.Write -> _, _ = w.Write) — 6 instances across DSR + DSRM test files

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-04-17T09:44:30.396Z
Stopped at: Completed 50-directory-service-roles-role-mappings-50-05-PLAN.md
Resume file: None
