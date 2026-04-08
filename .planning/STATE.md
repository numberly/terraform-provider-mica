---
gsd_state_version: 1.0
milestone: v2.1
milestone_name: Bucket Advanced Features
status: executing
stopped_at: Completed 40-tls-policies/40-02-PLAN.md
last_updated: "2026-04-03T07:42:49.559Z"
last_activity: 2026-04-03
progress:
  total_phases: 39
  completed_phases: 38
  total_plans: 85
  completed_plans: 83
  percent: 60
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-02)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Phase 40 — tls-policies

## Current Position

Phase: 40
Plan: Not started
Status: Ready to execute
Last activity: 2026-04-08 - Completed quick task 260408-bif: Fix lifecycle rule int64 zero-value handling

Progress: [██████░░░░] 60% (v2.2 — 3/5 phases complete)

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
| Phase 39-certificates P01 | 300 | 2 tasks | 4 files |
| Phase 39-certificates P02 | 513 | 2 tasks | 11 files |
| Phase 40-tls-policies P02 | 480 | 2 tasks | 17 files |

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
- [Phase 39-01]: Certificate models appended to models_network.go (network/TLS domain)
- [Phase 39-01]: POST struct excludes X.509 subject fields (extracted from PEM by API) — import-only mode
- [Phase 39-01]: passphrase and private_key are write-only — never stored or returned by mock handler
- [Phase 39-certificates]: UseStateForUnknown only on id and certificate_type; all renewal-volatile fields (issued_by, issued_to, valid_from, valid_to, key_algorithm, key_size, status) have no plan modifier to avoid masking drift
- [Phase 39-certificates]: private_key and passphrase Sensitive, preserved from plan/state; set to empty string on ImportState — API never returns write-only fields
- [Phase 40-tls-policies]: is_local gets UseStateForUnknown: computed stable field set by API at creation
- [Phase 40-tls-policies]: policy_type is Computed-only without UseStateForUnknown: volatile, drift detection still logs
- [Phase 40-tls-policies]: listToStringSlice defined locally in tls_policy_resource.go (not helpers.go) - single consumer

### v2.2 Phase Groupings

- Phase 36: TGT-01, TGT-02, TGT-03, TGT-04, TGT-05 — new flashblade_target resource + data source ✓
- Phase 37: RC-01, RC-02, BRL-01 — extend existing remote credentials + validate replica link with target ✓
- Phase 38: DOC-01, DOC-02, DOC-03 — import docs, workflow example, tfplugindocs ✓
- Phase 39: CERT-01 to CERT-05 — flashblade_certificate resource + data source (import PEM only)
- Phase 40: TLSP-01 to TLSP-06 — flashblade_tls_policy resource + DS + flashblade_tls_policy_member

### Pending Todos

None.

### Blockers/Concerns

None.

### Quick Tasks Completed

| # | Description | Date | Commit | Directory |
|---|-------------|------|--------|-----------|
| 260408-bif | Fix lifecycle rule int64 zero-value handling: use *int64 for duration fields | 2026-04-08 | bbd00d3 | [260408-bif-fix-lifecycle-rule-int64-zero-value-hand](./quick/260408-bif-fix-lifecycle-rule-int64-zero-value-hand/) |

## Session Continuity

Last session: 2026-04-08
Stopped at: Completed 40-tls-policies/40-02-PLAN.md
Resume file: None
