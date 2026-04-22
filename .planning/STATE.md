---
gsd_state_version: 1.0
milestone: pulumi-2.22.3
milestone_name: "Pulumi Bridge Alpha"
status: "defining-requirements"
stopped_at: "Milestone pulumi-2.22.3 started"
last_updated: "2026-04-21T00:00:00.000Z"
last_activity: 2026-04-21 — Milestone pulumi-2.22.3 started (defining requirements)
progress:
  total_phases: 0
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-21)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises.
**Current focus:** pulumi-2.22.3 — Bridge the FlashBlade TF provider to Pulumi (Python + Go, private distribution).

## Current Position

Milestone: pulumi-2.22.3 (Pulumi Bridge Alpha)
Phase: Not started (defining requirements)
Plan: —
Status: Defining requirements
Last activity: 2026-04-21 — Milestone pulumi-2.22.3 started

## Recent Milestones

- ✅ **v2.22.3** — Convention Compliance (shipped 2026-04-20, 779 tests, 12/12 requirements, [archive](milestones/v2.22.3-ROADMAP.md))
- ✅ **v2.22.2** — Directory Service Roles & Role Mappings (shipped 2026-04-17, 818 tests, [archive](milestones/v2.22.2-ROADMAP.md))
- ✅ **v2.22.1** — Directory Service – Array Management (shipped 2026-04-17, 798 tests, [archive](milestones/v2.22.1-ROADMAP.md))

## Accumulated Context

### Key Decisions

Full project decision log in `.planning/PROJECT.md`. Highlights across recent milestones:

- v2.22.3: CONVENTIONS.md §Pointer rules formalized with three exception classes (POST `*bool`/`*int64` for non-zero API defaults, `*[]T` for PATCH slices, "always send" carve-out for `NetworkInterfacePatch`). `doublePointerRefForPatch` helper becomes the canonical pattern for all `**NamedReference` PATCH call sites. Identity state upgraders (`type(oldState)` conversion) are the standard when only wire format changes.
- v2.22.2: DSR name user-supplied via `?names=` (D-03 superseded post-50.1); composite ID for DSRM membership uses `/` not `:` because role names can contain `:` (e.g. `pure:policy/array_admin`).
- v2.22.1: Directory Service Management is a singleton resource; `bind_password` Sensitive write-only; Delete is full-reset PATCH (no DELETE endpoint).

### Open Blockers

_(none)_

## Next Steps

Research (4 parallel agents) → REQUIREMENTS.md → ROADMAP.md → `/gsd:discuss-phase [N]` on first phase.
