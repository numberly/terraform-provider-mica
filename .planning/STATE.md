---
gsd_state_version: 1.0
milestone: v2.23.1
milestone_name: "**Goal:** Ship `flashblade_snmp_manager` resource + data source"
status: verifying
last_updated: "2026-05-20T12:39:03.209Z"
last_activity: 2026-05-20
progress:
  total_phases: 1
  completed_phases: 1
  total_plans: 1
  completed_plans: 1
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-20)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises.
**Current focus:** Phase 61 — flashblade-snmp-manager

## Current Position

Milestone: v2.23.1 (`flashblade_snmp_manager`) — **EXECUTION COMPLETE, AWAITING VERIFICATION**
Phase: 61 (flashblade-snmp-manager) — VERIFYING
Plan: 1 of 1 (complete)
Status: Phase complete — ready for verification
Last activity: 2026-05-20 -- Plan 61-01 executed, 816 tests, branch implem-snmp-managers (11 commits)

Progress: [██████████] 100% (1/1 phases, 1/1 plans)

## Recent Milestones

- 🚧 **v2.23.1** — `flashblade_snmp_manager` (in planning, started 2026-05-20)
- ✅ **v2.23.0** — FlashBlade API 2.23 Upgrade (shipped 2026-05-20, 807 tests, 33/33 requirements, [release](https://github.com/numberly/terraform-provider-mica/releases/tag/v2.23.0), [archive](milestones/v2.23.0-ROADMAP.md))
- ✅ **pulumi-2.22.3** — Pulumi Bridge Alpha (shipped 2026-04-24, 836 TF tests + 23 bridge tests, [archive](milestones/pulumi-2.22.3-ROADMAP.md))
- ✅ **v2.22.3** — Convention Compliance (shipped 2026-04-20, 779 tests, 12/12 requirements, [archive](milestones/v2.22.3-ROADMAP.md))
- ✅ **v2.22.2** — Directory Service Roles & Role Mappings (shipped 2026-04-17, 818 tests, [archive](milestones/v2.22.2-ROADMAP.md))

## Performance Metrics

- **Provider tests:** 816 (post-Phase-61, baseline 807 + 9 new for `flashblade_snmp_manager`)
- **TEST_BASELINE (GNUmakefile):** 807 — NOT bumped (reserved for release milestones, will move to 816 at v2.23.1 ship)
- **Lint:** 0 issues
- **Resources / Data sources:** 56 / 44 (post-Phase-61, +1 resource +1 data source `flashblade_snmp_manager`)
- **Phase 61 plan 01 execution:** 13 tasks, 11 atomic commits on `implem-snmp-managers`, ~15 min

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

### Key Decisions (Phase 61 execution)

- **Mock handler wiring follows codebase pattern** (per-test registration via `ms.Mux`), not plan literal text (`server.go`). The existing codebase never wires resource handlers in `NewMockServer`. Deviation documented in SUMMARY.md.
- **Drift logs inlined to satisfy "exactly 6" contract** (not routed through a helper) for grep-ability and explicit per-leaf branching.
- **Strict POST-time validators applied at provider schema level for both Create AND Update** (auth_passphrase ≤ 32, privacy_passphrase 8..63) — predictable validation before PATCH.

### Open Todos

- Run `/gsd:verify-phase 61` to validate the execution.
- After verification: tag `v2.23.1`, push branch `implem-snmp-managers`, open PR to `main`.
- At v2.23.1 release: bump `TEST_BASELINE` in `GNUmakefile` from 807 to 816.

### Open Blockers

_(none)_

## Next Steps

Run `/gsd:verify-phase 61` to validate the execution. On pass: tag v2.23.1, open PR `implem-snmp-managers` → `main`.

## Session Log

- 2026-05-20 — Milestone v2.23.1 created (`flashblade_snmp_manager` CRUD, branch `implem-snmp-managers`). Pre-check Serena: no collision. API schemas validated via `swagger-to-reference` + raw `swagger-2.23.json`.
- 2026-05-20 — Roadmap created: Phase 61 (Implement `flashblade_snmp_manager` Resource & Data Source). 13/13 requirements mapped.
- 2026-05-20 — Phase 61 context gathered. 20 decisions locked (D-01..D-20) in `phases/61-flashblade-snmp-manager/61-CONTEXT.md`. Next: `/gsd:plan-phase 61`.
- 2026-05-20 — **Phase 61 plan 01 executed.** 13 tasks, 11 atomic commits on `implem-snmp-managers` (a241ec1 → 24098d1). 816 tests (807 + 9), lint clean, docs idempotent, ROADMAP row moved to Implemented. SUMMARY at `.planning/phases/61-flashblade-snmp-manager/61-01-implement-snmp-manager-SUMMARY.md`.
