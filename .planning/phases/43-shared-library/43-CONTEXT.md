# Phase 43: Shared Library - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning
**Mode:** Auto-generated (infrastructure phase — discuss skipped)

<domain>
## Phase Boundary

All Python tooling shares a single, well-tested utility library that resolves allOf schemas, normalizes API paths, and flattens nested schemas. Output: `.claude/skills/_shared/swagger_utils.py` with Python 3.10+ stdlib only.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — pure infrastructure phase. Use ROADMAP phase goal, success criteria, and codebase conventions to guide decisions.

Key technical constraints:
- Python 3.10+ stdlib only (json, re, pathlib, collections, typing)
- Must handle OpenAPI 3.0 format (components/schemas, $ref, allOf)
- 404/709 schemas in swagger-2.22.json use allOf — resolver must handle multi-level inheritance
- Private schemas (_-prefixed) should be resolved inline, not emitted standalone
- Wrapper schemas (*Response, *GetResponse) should be skipped
- Path normalization must strip `/api/<version>/` prefix deterministically

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `swagger-2.22.json` (1.5MB, 226 paths, OpenAPI 3.0.1) — primary test fixture
- `swagger-2.23.json` (1.6MB, 233 paths) — secondary test fixture
- `FLASHBLADE_API.md` (1361 lines) — target output format reference

### Established Patterns
- Skills live in `.claude/skills/<skill-name>/`
- Shared code goes in `.claude/skills/_shared/`

### Integration Points
- `parse_swagger.py` (Phase 44) will import from this library
- `diff_swagger.py` (Phase 46) will import from this library
- `browse_api.py` (Phase 45) may import for schema parsing

</code_context>

<specifics>
## Specific Ideas

No specific requirements — infrastructure phase. Refer to ROADMAP phase description and success criteria.

</specifics>

<deferred>
## Deferred Ideas

None — infrastructure phase.

</deferred>
