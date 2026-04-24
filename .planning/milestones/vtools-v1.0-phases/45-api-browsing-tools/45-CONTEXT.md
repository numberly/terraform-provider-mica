# Phase 45: API Browsing Tools - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning
**Mode:** Auto-generated (infrastructure phase — discuss skipped)

<domain>
## Phase Boundary

Claude can search, inspect, and compare API endpoints and schemas interactively from a generated reference file. Creates `browse_api.py` CLI tool with subcommands: endpoints, schema, compare-schemas, stats, search, tags, fields.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — pure infrastructure phase.

Key constraints:
- Input: generated markdown reference from parse_swagger.py (api_references/<version>.md)
- CLI: `python3 browse_api.py <reference.md> <command> [args]`
- Must parse the markdown format back into structured data (endpoints list, schema dict)
- Subcommands: endpoints [--tag TAG] [--method METHOD] [--search PATTERN], schema <name>, fields <schema>, compare-schemas <s1> <s2>, tags, stats, search <pattern>
- Use _shared/swagger_utils.py if needed, but browse_api.py mainly parses markdown
- Output: filtered, human-readable text to stdout suitable for AI consumption
- Python 3.10+ stdlib only

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `api_references/2.22.md` (or generated via parse_swagger.py) — input format
- `.claude/skills/_shared/swagger_utils.py` — may reuse normalize_path
- FLASHBLADE_API.md format documentation

### Integration Points
- Lives at `.claude/skills/swagger-to-reference/scripts/browse_api.py` (same skill dir as parse_swagger.py)
- Phase 46 (api-diff) may reuse parsing logic

</code_context>

<specifics>
## Specific Ideas

No specific requirements — infrastructure phase.

</specifics>

<deferred>
## Deferred Ideas

None.

</deferred>
