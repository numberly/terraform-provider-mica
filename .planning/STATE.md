---
gsd_state_version: 1.0
milestone: v2.1
milestone_name: Bucket Advanced Features
status: planning
stopped_at: Completed 38-documentation-workflow/38-01-PLAN.md
last_updated: "2026-04-02T16:24:49.061Z"
last_activity: 2026-04-02 — v2.2 roadmap created, Phases 36-38 defined
progress:
  total_phases: 38
  completed_phases: 36
  total_plans: 81
  completed_plans: 79
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-02)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** v2.2 S3 Target Replication — Phase 36 ready to plan

## Current Position

Phase: 36 of 38 (Target Resource)
Plan: —
Status: Ready to plan
Last activity: 2026-04-02 — v2.2 roadmap created, Phases 36-38 defined

Progress: [░░░░░░░░░░] 0% (v2.2)

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases defined | 3 |
| Phases complete | 0 |
| Plans defined | TBD |
| Plans complete | 0 |
| Requirements mapped | 11/11 |
| Phase 36-target-resource P01 | 12 | 3 tasks | 4 files |
| Phase 36-target-resource P02 | 348s | 2 tasks | 9 files |
| Phase 37-remote-credentials-replica-link-enhancement P01 | 388s | 2 tasks | 6 files |
| Phase 38-documentation-workflow P01 | 135 | 3 tasks | 4 files |

## Accumulated Context

### Decisions

- [Phase 35-04]: Mock handler fixed: objectStoreUserStore stores ObjectStoreUser with UUID id (was bool + empty string)
- [Phase 35-04]: ImportStateId must be explicit for name-based import when id attribute holds UUID
- [Phase 35-04]: ImportStateVerifyIdentifierAttribute=user_name for policy resource (no id field in schema)
- [Phase 35]: Update stub returns AddError — all attributes are RequiresReplace so Update is never called in practice
- [Phase 35]: ImportState uses inline CRD-only null timeouts (create/read/delete) instead of shared nullTimeoutsValue which includes update key
- [v2.2 roadmap]: 3 phases at coarse granularity — Phase 36 (target CRUD), Phase 37 (RC + BRL extension), Phase 38 (docs)
- [Phase 36-01]: Use **NamedReference for TargetPatch.CACertificateGroup to support nil=omit vs inner-nil=set-null PATCH semantics
- [Phase 36-01]: Mock GET handler returns HTTP 404 (not empty list) when ?names= filter finds no match so getOneByName detects not-found via HTTP status
- [Phase 36-01]: targetStoreFacade wrapper in test file exposes Seed without making internal targetStore type public
- [Phase 36-02]: Flat ca_certificate_group string in resource schema (not nested object) — keeps HCL simple and consistent with plan spec
- [Phase 36-02]: Drift detection on Read logs all four mutable/computed fields via tflog.Debug with field/was/now keys
- [Phase 37-01]: remote_name changed to Optional+Computed: API always populates Remote.Name field
- [Phase 37-01]: target_name preserved from plan/state like SecretAccessKey (not returned by GET)
- [Phase 37-01]: v0->v1 upgrader uses remoteCredentialsV0Model intermediate struct; sets target_name=null
- [Phase 38-01]: DOC-01: import.sh uses the target name (not UUID) as the import identifier, matching the ImportState implementation
- [Phase 38-01]: DOC-02: s3-target-replication workflow uses single-provider pattern (one FlashBlade, one external S3) — no provider aliases
- [Phase 38-01]: DOC-03: make docs regenerates target.md with Import section; object_store_remote_credentials.md updated to reflect target_name attribute from Phase 37

### v2.2 Phase Groupings

- Phase 36: TGT-01, TGT-02, TGT-03, TGT-04, TGT-05 — new flashblade_target resource + data source
- Phase 37: RC-01, RC-02, BRL-01 — extend existing remote credentials + validate replica link with target
- Phase 38: DOC-01, DOC-02, DOC-03 — import docs, workflow example, tfplugindocs

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-04-02T16:22:16.127Z
Stopped at: Completed 38-documentation-workflow/38-01-PLAN.md
Resume file: None
