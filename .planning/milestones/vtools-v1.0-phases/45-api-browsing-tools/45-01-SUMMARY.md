---
phase: 45-api-browsing-tools
plan: "01"
subsystem: tooling
tags: [python, cli, markdown-parser, api-reference, swagger]

requires:
  - phase: 44-swagger-to-reference-skill
    provides: parse_swagger.py that generates api_references/2.22.md

provides:
  - browse_api.py CLI tool for querying generated markdown API reference
  - api_references/2.22.md — generated reference file checked in

affects:
  - future plan sessions using browse_api.py for API discovery
  - 46+ phases requiring targeted endpoint or schema lookup

tech-stack:
  added: []
  patterns:
    - Single-pass state machine parser for structured markdown
    - Multi-line continuation via ' | ' prefix detection
    - Stdlib-only Python CLI (argparse, re, collections)

key-files:
  created:
    - .claude/skills/swagger-to-reference/scripts/browse_api.py
    - api_references/2.22.md

key-decisions:
  - "Continuation lines (starting ' | ') appended with explicit ' | ' separator, not raw strip() — avoids pipe-prefix collision in _parse_schema_line split"
  - "stats --paths shows unique paths (de-duplicated by URL), not total endpoint count"
  - "Tag match is case-insensitive substring — 'buckets' matches 'Buckets' tag"

patterns-established:
  - "ReferenceParser: _parse() state machine with section/current_tag variables, pending_schema_name/text for multi-line accumulation"
  - "cmd_* functions receive ReferenceParser, print to stdout, errors to stderr + exit(1)"

requirements-completed:
  - BRWS-01
  - BRWS-02
  - BRWS-03
  - BRWS-04

duration: 20min
completed: 2026-04-14
---

# Phase 45 Plan 01: API Browsing Tools Summary

**browse_api.py CLI that parses generated markdown reference to list endpoints by tag/method, inspect schema fields, diff two schemas, and show statistics — stdlib-only, single-pass state machine parser**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-04-14T00:00:00Z
- **Completed:** 2026-04-14
- **Tasks:** 2 (implemented together in single file)
- **Files modified:** 2

## Accomplishments

- ReferenceParser class with single-pass state machine handling endpoints section (tag headers + endpoint lines) and data models section (multi-line schema entries)
- Multi-line schema fix: continuation lines starting with ` | ` are stripped of `| ` prefix before appending with explicit ` | ` separator to avoid parse collision
- Six CLI subcommands: `--tag`, `--method`, `--search`, `--schema`, `--compare`, `--stats` all verified against api_references/2.22.md (226 paths, 312 schemas, 538 ops)

## Task Commits

1. **Tasks 1+2: Implement browse_api.py (parser + CLI subcommands)** - `34ebe58` (feat)

**Plan metadata:** (docs commit below)

## Files Created/Modified

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/.claude/skills/swagger-to-reference/scripts/browse_api.py` — CLI tool with ReferenceParser + 6 subcommands
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/api_references/2.22.md` — generated reference file (2098 lines, 226 paths, 538 ops, 312 schemas)

## Decisions Made

- Multi-line schema entries: continuation lines (` | field...`) must strip leading `| ` before appending with separator — avoids pipe prefix in field splitting
- Tag display name taken from first matched endpoint's tag field (preserves original casing)
- `--stats` paths count = unique URLs (de-duplicated), not total operation count

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed multi-line schema continuation parsing**
- **Found during:** Task 2 verification (`--compare BucketPost BucketPatch`)
- **Issue:** BucketPatch spans 2 lines in markdown; `eradication_config` was missing from parsed fields because ` | ` continuation line was stripped to `| field...` and the leading `|` prevented field regex match after split
- **Fix:** Strip `| ` prefix from continuation lines before appending with explicit ` | ` separator
- **Files modified:** `.claude/skills/swagger-to-reference/scripts/browse_api.py`
- **Verification:** `--compare BucketPost BucketPatch` shows `eradication_config: object | object`; `--schema BucketPatch` shows 10 fields
- **Committed in:** `34ebe58`

---

**Total deviations:** 1 auto-fixed (Rule 1 - bug)
**Impact on plan:** Essential for correctness. No scope creep.

## Issues Encountered

None beyond the multi-line continuation bug (auto-fixed).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- browse_api.py ready for use: `python3 .claude/skills/swagger-to-reference/scripts/browse_api.py api_references/2.22.md --tag <tag>`
- api_references/2.22.md checked in — available for all future plan sessions without regenerating

---
*Phase: 45-api-browsing-tools*
*Completed: 2026-04-14*
