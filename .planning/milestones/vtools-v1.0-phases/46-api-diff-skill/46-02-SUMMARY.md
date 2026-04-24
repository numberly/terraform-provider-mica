---
phase: 46-api-diff-skill
plan: "02"
subsystem: tooling
tags: [python, api-diff, migration-plan, roadmap, swagger]

requires:
  - phase: 46-01
    provides: diff_swagger.py producing diff.json with new/removed/modified endpoints and schemas

provides:
  - generate_migration_plan.py: 4-category migration plan (update_models, new_resources, deprecated, roadmap_gaps) from diff.json
  - known_discrepancies.md: living doc for swagger vs real API divergences
  - known_discrepancies.json: machine-readable overrides for --discrepancies flag

affects:
  - 46-03 (api-upgrade skill will consume migration plans)
  - future ROADMAP.md updates triggered by new FlashBlade API versions

tech-stack:
  added: []
  patterns:
    - "Fuzzy ROADMAP.md cross-reference: >=2 words from api_section must appear in normalized_path"
    - "Migration plan deduplication: GET method as anchor per normalized_path"
    - "Action strings are template-generated (no LLM required)"

key-files:
  created:
    - .claude/skills/api-diff/scripts/generate_migration_plan.py
    - .claude/skills/api-diff/references/known_discrepancies.md
    - .claude/skills/api-diff/references/known_discrepancies.json
  modified: []

key-decisions:
  - "Fuzzy matching uses word count (>=2 words from api_section in path) — avoids false positives for short names"
  - "known_discrepancies.json starts empty (no overrides) — populated by investigation sessions"
  - "roadmap_gaps is a subset of new_resources, not a separate list — same endpoint appears in both"

patterns-established:
  - "Migration plan script: stdlib only, reads diff.json + ROADMAP.md, outputs JSON or markdown"
  - "ROADMAP.md parsing: find ## Not Implemented header, parse table rows, filter Candidate/Deferred status"

requirements-completed: [DIFF-03, DIFF-04]

duration: 12min
completed: 2026-04-14
---

# Phase 46 Plan 02: API Diff Skill — Migration Plan Generator Summary

**Migration plan generator (generate_migration_plan.py) that cross-references diff.json with ROADMAP.md Candidate/Deferred entries to surface roadmap_gaps, plus known_discrepancies.md living doc template**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-14T08:35:00Z
- **Completed:** 2026-04-14T08:47:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Created `generate_migration_plan.py` producing 4-category JSON/markdown migration plans from diff.json
- Implemented fuzzy ROADMAP.md cross-reference: 19 Candidate + 15 Deferred entries parsed; new endpoint "file-systems/replica-links" correctly matched to "File System Replica Links" Candidate
- Created `known_discrepancies.md` + `known_discrepancies.json` as living annotation layer

## Task Commits

1. **Task 1: Create known_discrepancies.md** - `d26da7e` (feat)
2. **Task 2: Create generate_migration_plan.py** - `d303cea` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `.claude/skills/api-diff/scripts/generate_migration_plan.py` - CLI migration plan generator (stdlib only)
- `.claude/skills/api-diff/references/known_discrepancies.md` - Living doc for swagger vs real API divergences
- `.claude/skills/api-diff/references/known_discrepancies.json` - Machine-readable overrides for --discrepancies flag

## Decisions Made
- Fuzzy matching threshold set at >=2 words from api_section appearing in normalized_path — single-word entries fall back to substring match
- `roadmap_gaps` is a subset of `new_resources` (same item appears in both lists) — consumers can use either for different views
- Action strings are purely template-driven — no LLM inference required for migration plan generation

## Deviations from Plan
None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Known Stubs
None - all data flows are wired (reads actual diff.json and ROADMAP.md).

## Next Phase Readiness
- Full api-diff skill pipeline ready: diff_swagger.py → diff.json → generate_migration_plan.py → migration plan
- known_discrepancies.json empty — populate after first real diff run against swagger-2.23.json
- Phase 47 (api-upgrade skill) can consume migration plan JSON for mechanical upgrade scripts

---
*Phase: 46-api-diff-skill*
*Completed: 2026-04-14*
