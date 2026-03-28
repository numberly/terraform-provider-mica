---
gsd_state_version: 1.0
milestone: v1.1
milestone_name: Servers & Exports
status: executing
stopped_at: Completed 08-02-PLAN.md
last_updated: "2026-03-28T15:48:49.524Z"
last_activity: 2026-03-28 — Completed plan 08-02 (Syslog server resource)
progress:
  total_phases: 8
  completed_phases: 7
  total_plans: 28
  completed_plans: 27
  percent: 96
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-28)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises
**Current focus:** Milestone v1.1 — Servers & Exports, Phase 8 in progress

## Current Position

Phase: 8 of 8 (SMB Client Policies, Syslog & Acceptance Tests)
Plan: 2 of 3
Status: In progress
Last activity: 2026-03-28 — Completed plan 08-02 (Syslog server resource)

Progress: [██████████] 96%

## Performance Metrics

**Velocity (from v1.0):**
- Total plans completed: 20
- Average duration: ~100 min
- Total execution time: ~33 hours

**By Phase (v1.0):**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-foundation | 4 | 270 min | 68 min |
| 02-object-store | 3 | 1329 min | 443 min |
| 03-file-policies | 4 | 566 min | 142 min |
| 04-obj-net-quota-admin | 5 | 423 min | 85 min |
| 05-quality-hardening | 4 | 143 min | 36 min |
| 06-server-resource | 1 | 5 min | 5 min |
| 06-export-tests | 1 | 8 min | 8 min |

*Updated after each plan completion*
| Phase 07 P01 | 3 | 3 tasks | 5 files |
| Phase 07 P02 | 5 min | 2 tasks | 6 files |
| Phase 07 P03 | 4 min | 2 tasks | 4 files |
| Phase 08 P01 | 7 min | 2 tasks | 9 files |
| Phase 08 P02 | 4 | 2 tasks | 7 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [v1.0]: Client layer is pure Go with zero terraform-plugin-framework imports — testable with httptest.NewServer
- [v1.0]: FlashBlade returns HTTP 200 with empty items for non-existent resources — synthesize 404
- [v1.0]: PATCH handler uses raw map[string]json.RawMessage for true PATCH semantics
- [v1.0]: Rule import uses composite ID policy_name/rule_index or policy_name/rule_name
- [v1.0]: Singleton admin resources use GET-first then PATCH-or-POST pattern
- [v1.1]: DNS modeled as ListNestedAttribute; cascade_delete is write-only list used only on Delete
- [v1.1]: Server creation uses ?create_ds= query param (not ?names=) per FlashBlade API convention
- [v1.1]: Mock DELETE handler for account exports uses lenient lookup (Pitfall 5: data.Name passed as combined name)
- [Phase 07]: S3 export policy GET does not embed rules (unlike NFS) - rules fetched from separate /rules endpoint
- [Phase 07]: S3 export policy rule effect is patchable in-place (unlike OAP rules which require replace)
- [Phase 07]: S3 export policy delete has no member guard (unlike NFS which checks file system attachments)
- [Phase 07]: Virtual host attached_servers uses listdefault.StaticValue with empty list to prevent null-vs-empty drift
- [Phase 07]: Virtual host name is Computed (server-assigned); hostname is Required (user-supplied); import uses name
- [Phase 08]: SMB client policy version is read-only Computed+UseStateForUnknown (never in POST/PATCH)
- [Phase 08]: SMB client policy rule fields are client/encryption/permission (not principal/change/full_control/read)
- [Phase 08]: Syslog server name uses RequiresReplace (not renameable per API)

### Pending Todos

None yet.

### Blockers/Concerns

- Server cascade delete semantics need API validation — confirm DELETE with ?cascade=true behavior
- Existing export resources (file_system_export, account_export, server data source) were created quickly during v1.0 — need TDD consolidation

## Session Continuity

Last session: 2026-03-28T15:48:49.519Z
Stopped at: Completed 08-02-PLAN.md
Resume file: None
