---
gsd_state_version: 1.0
milestone: v2.23.1
milestone_name: flashblade_snmp_manager
status: planning
last_updated: "2026-05-20T10:00:00.000Z"
last_activity: 2026-05-20
progress:
  total_phases: 1
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-20)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises.
**Current focus:** v2.23.1 — `flashblade_snmp_manager` resource & data source (full CRUD on `/api/2.23/snmp-managers`)

## Current Position

Milestone: v2.23.1 (`flashblade_snmp_manager`) — **PLANNING**
Phase: 61 — `flashblade_snmp_manager` Resource & Data Source (context gathered)
Plan: pending (`/gsd:plan-phase 61`)
Status: Context gathered, ready to plan
Last activity: 2026-05-20 — Phase 61 context captured (20 decisions D-01..D-20)

Progress: [          ] 0% (0/1 phases, 0/0 plans)

## Recent Milestones

- 🚧 **v2.23.1** — `flashblade_snmp_manager` (in planning, started 2026-05-20)
- ✅ **v2.23.0** — FlashBlade API 2.23 Upgrade (shipped 2026-05-20, 807 tests, 33/33 requirements, [release](https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0), [archive](milestones/v2.23.0-ROADMAP.md))
- ✅ **pulumi-2.22.3** — Pulumi Bridge Alpha (shipped 2026-04-24, 836 TF tests + 23 bridge tests, [archive](milestones/pulumi-2.22.3-ROADMAP.md))
- ✅ **v2.22.3** — Convention Compliance (shipped 2026-04-20, 779 tests, 12/12 requirements, [archive](milestones/v2.22.3-ROADMAP.md))
- ✅ **v2.22.2** — Directory Service Roles & Role Mappings (shipped 2026-04-17, 818 tests, [archive](milestones/v2.22.2-ROADMAP.md))

## Performance Metrics

- **Provider tests:** 807 (TEST_BASELINE at last shipped milestone v2.23.0)
- **TEST_BASELINE (GNUmakefile):** 807 — must NOT be bumped in this milestone; expected ≥ 816 after Phase 61 lands.
- **Lint:** 0 issues at last release
- **Resources / Data sources:** 55 / 43 — expected delta on merge: **+1 resource**, **+1 data source** (`flashblade_snmp_manager`)

## Accumulated Context

### Key Decisions (v2.23.1)

- Resource scope = pure CRUD on `/snmp-managers`. The connectivity test endpoint `GET /snmp-managers/test` is OUT of scope (resource-action pattern, future milestone alongside `/dns/test`, `/smtp/test`, etc.).
- Branch from clean `main`: `implem-snmp-managers`.
- Domain placement: `internal/client/models_admin.go` (with `SmtpServer`, `SyslogServer`, `AlertWatcher`). Confirmed via `mcp__serena__get_symbols_overview`.
- Pre-check (Serena `find_symbol` on `SnmpManager` / `Snmp*` / `snmp_manager`): no existing code, greenfield implementation.
- Sensitive write-once fields: `v2c.community`, `v3.auth_passphrase`, `v3.privacy_passphrase` — never returned by API GET → keep state value, never overwrite in Read; null in ImportState.
- Validators choose the **stricter POST-time constraints** from `_snmp_v3_post` (privacy_passphrase 8-63, auth_passphrase ≤ 32) for safer UX.
- No cross-field validator on `version` vs. `v2c`/`v3` — let API validate (alignment with provider conventions).
- `TEST_BASELINE` (GNUmakefile) must NOT be bumped in v2.23.1 — reserved for release milestones.

### Key Decisions (carried from v2.23.0, for context)

- Pulumi SDK regen / publish is owned by a separate `pulumi-2.23.x` milestone (out of scope here too).

### Open Todos

- Plan Phase 61 via `/gsd:plan-phase 61`.

### Open Blockers

_(none)_

## Next Steps

Run `/gsd:plan-phase 61` (or `/gsd:discuss-phase 61` first) to decompose Phase 61 into executable plans following the *New Resource* 16-item checklist from `CONVENTIONS.md`. Implementation must be driven by the `flashblade-resource-builder` skill.

## Session Log

- 2026-05-20 — Milestone v2.23.1 created (`flashblade_snmp_manager` CRUD, branch `implem-snmp-managers`). Pre-check Serena: no collision. API schemas validated via `swagger-to-reference` + raw `swagger-2.23.json`.
- 2026-05-20 — Roadmap created: Phase 61 (Implement `flashblade_snmp_manager` Resource & Data Source). 13/13 requirements mapped.
- 2026-05-20 — Phase 61 context gathered. 20 decisions locked (D-01..D-20) in `phases/61-flashblade-snmp-manager/61-CONTEXT.md`. Next: `/gsd:plan-phase 61`.
