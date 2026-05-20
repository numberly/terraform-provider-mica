---
gsd_state_version: 1.0
milestone: v2.23.0
milestone_name: FlashBlade API 2.23 Upgrade
status: shipped
last_updated: "2026-05-20T09:30:00.000Z"
last_activity: 2026-05-20
progress:
  total_phases: 2
  completed_phases: 2
  total_plans: 10
  completed_plans: 10
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-20)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises.
**Current focus:** No active milestone — `/gsd:new-milestone` for next cycle

## Current Position

Milestone: v2.23.0 (FlashBlade API 2.23 Upgrade) — **SHIPPED 2026-05-20**
Status: Archived — no active milestone
Last activity: 2026-05-20 — milestone archived

Progress: [██████████] 100% (2/2 phases, 10/10 plans)

## Recent Milestones

- ✅ **v2.23.0** — FlashBlade API 2.23 Upgrade (shipped 2026-05-20, 807 tests, 33/33 requirements, [release](https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0), [archive](milestones/v2.23.0-ROADMAP.md))
- ✅ **pulumi-2.22.3** — Pulumi Bridge Alpha (shipped 2026-04-24, 836 TF tests + 23 bridge tests, [archive](milestones/pulumi-2.22.3-ROADMAP.md))
- ✅ **v2.22.3** — Convention Compliance (shipped 2026-04-20, 779 tests, 12/12 requirements, [archive](milestones/v2.22.3-ROADMAP.md))
- ✅ **v2.22.2** — Directory Service Roles & Role Mappings (shipped 2026-04-17, 818 tests, [archive](milestones/v2.22.2-ROADMAP.md))
- ✅ **v2.22.1** — Directory Service – Array Management (shipped 2026-04-17, 798 tests, [archive](milestones/v2.22.1-ROADMAP.md))

## Performance Metrics

- **Provider tests:** 836 (baseline at last shipped milestone pulumi-2.22.3)
- **TEST_BASELINE (GNUmakefile):** 807 — to refresh once API 2.23 work lands on main (RELEASE-06)
- **Lint:** 0 issues at last release
- **Resources / Data sources:** 54 / 40 pre-API-2.23. Expected delta on merge: +1 resource (workload), +3 data sources (workload, resiliency_group, resiliency_group_member)

## Accumulated Context

### Key Decisions (v2.23.0)

- Retro milestone: 19/33 requirements already implemented on branch `test/api-upgrade-2.23`. They are mapped to Phase 59 for traceability only, not re-execution.
- 14 requirements are active work: VALID-01..06 (Phase 59), RELEASE-01..07 (Phase 60).
- Tight 2-phase split (coarse granularity): consolidation+validation, then release.
- Acceptance validation on par5 + pa7 is mandatory before merge (VALID-04).
- Pulumi SDK regen / publish (`pulumi-2.23.0`) is OUT of scope — separate milestone.

### Key Decisions (pulumi-2.22.3, kept for context)

- Module path: `github.com/numberly/opentofu-provider-flashblade`. Bridge modules under `./pulumi/provider/` and `./pulumi/sdk/go/` with `replace ../../`.
- Bridge: `pulumi-terraform-bridge/v3 v3.127.0`, `pulumi/sdk/v3 v3.231.0`, `pulumi/pkg/v3 v3.231.0`.
- Schema commit policy: `schema.json` + `bridge-metadata.json` committed; CI gate via `git diff --exit-code` after `make tfgen` — directly relevant to VALID-05.
- Composite IDs use `/` separator with string keys.
- Tokens via SingleModule (`flashblade:index/*`).

### Open Todos

- Plan Phase 59 via `/gsd:plan-phase 59`.
- At Phase 60 release time: bump `TEST_BASELINE` in `GNUmakefile` (RELEASE-06).

### Open Blockers

_(none)_

## Next Steps

Run `/gsd:plan-phase 59` to decompose Phase 59 into executable plans (consolidation + validation work covering VALID-01..06, plus retro traceability for the 19 already-shipped API/WORKLOAD/RESILIENCY/SCHEMA/BRIDGE requirements).

## Session Log

- 2026-05-20 — Milestone v2.23.0 created (retro + finalisation for API 2.23 upgrade on branch `test/api-upgrade-2.23`).
- 2026-05-20 — Roadmap created: Phase 59 (API 2.23 Consolidation & Validation), Phase 60 (v2.23.0 Release). 33/33 requirements mapped.
