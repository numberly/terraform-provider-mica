---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: In Progress
last_updated: "2026-04-20T08:45:51.463Z"
last_activity: 2026-04-20 — Completed 52-01-PLAN.md (R-006, +2 tests, commit 4b83e27)
progress:
  total_phases: 3
  completed_phases: 1
  total_plans: 9
  completed_plans: 5
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-20)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Milestone v2.22.3 convention-compliance — defining requirements

## Current Position

Milestone: v2.22.3 convention-compliance
Phase: 52-important-conformance (1/5 plans complete)
Plan: 52-02 (next) — R-007 9-resource UpgradeState rollout
Status: In Progress
Last activity: 2026-04-20 — Completed 52-01-PLAN.md (R-006, +2 tests → 772, commit 4b83e27)

## Recent Milestones

- ✅ **v2.22.2** — Directory Service Roles & Role Mappings (shipped 2026-04-17, 818 tests, [archive](milestones/v2.22.2-ROADMAP.md))
- ✅ **v2.22.1** — Directory Service – Array Management (shipped 2026-04-17, 798 tests, [archive](milestones/v2.22.1-ROADMAP.md))

## Accumulated Context

### Key Decisions

Full project decision log in `.planning/PROJECT.md`. Highlights across recent milestones:

- v2.22.2: DSR name is user-supplied via `?names=` (D-03 superseded post-50.1); composite ID for DSRM membership uses `/` not `:` (D-05) because role names can contain `:` (e.g. `pure:policy/array_admin`).
- v2.22.1: Directory Service Management is a singleton resource; `bind_password` Sensitive write-only; Delete is full-reset PATCH (no DELETE endpoint).
- v2.22.3 (2026-04-20): CONVENTIONS.md §Pointer rules clarified — `*bool` allowed in POST when API default is non-zero (e.g., policy `Enabled` defaults to `true`); `*int64`/`*string` allowed when zero is a semantic user choice (e.g., `VLAN=0` means untagged); pointers on nested structs and optional refs are the correct Go idiom for `omitempty`.
- D-52-01 (2026-04-20): `QosPolicyPost.MaxTotal*` migrated to `*int64` + `Name` → `json:"-"`; QoS resource schema v0→v1 with identity upgrader (R-006 closed, commit 4b83e27).

### Open Blockers

_(none)_

## Next Steps

Draft `.planning/REQUIREMENTS.md` for v2.22.3 (27 violations grouped in Critical/Important/Cosmetic phases), then `/gsd:plan-phase 51` to start execution.
