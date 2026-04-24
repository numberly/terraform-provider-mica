# Phase 44: swagger-to-reference Skill - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning
**Mode:** Auto-generated (infrastructure phase — discuss skipped)

<domain>
## Phase Boundary

Claude can convert any FlashBlade swagger.json into the AI-optimized markdown format matching FLASHBLADE_API.md, with correct allOf resolution and versioned output. Creates `parse_swagger.py` script and `swagger-to-reference` SKILL.md.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — pure infrastructure phase.

Key constraints:
- Output must match FLASHBLADE_API.md format (1361 lines, same structure)
- Use _shared/swagger_utils.py for allOf resolution and path normalization
- parse_swagger.py CLI: `python3 parse_swagger.py <swagger.json> [--output path] [--version X.XX]`
- Output goes to `api_references/<version>.md`
- OpenAPI 3.0 format (components/schemas, requestBody, $ref)
- Skip wrapper schemas (*Response, *GetResponse), resolve private (_-prefixed) inline
- SKILL.md must ask user for API version before processing

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `.claude/skills/_shared/swagger_utils.py` — resolve_all_of, normalize_path, flatten_schema
- `FLASHBLADE_API.md` — target output format reference (MUST match structure)
- `swagger-2.22.json` — primary test input (226 paths, 538 ops)

### Integration Points
- Phase 45 (browse_api.py) parses the generated markdown reference
- Phase 46 (diff_swagger.py) may use parse_swagger.py internals
- SKILL.md at `.claude/skills/swagger-to-reference/SKILL.md`
- Script at `.claude/skills/swagger-to-reference/scripts/parse_swagger.py`

</code_context>

<specifics>
## Specific Ideas

No specific requirements — infrastructure phase.

</specifics>

<deferred>
## Deferred Ideas

None.

</deferred>
