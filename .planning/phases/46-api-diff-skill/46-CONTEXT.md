# Phase 46: api-diff Skill - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning
**Mode:** Auto-generated (infrastructure phase — discuss skipped)

<domain>
## Phase Boundary

Claude can produce a structured diff between two swagger versions, annotate discrepancies, and generate a migration plan cross-referenced with ROADMAP.md. Creates diff_swagger.py, generate_migration_plan.py, known_discrepancies.md, and api-diff SKILL.md.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — pure infrastructure phase.

Key constraints:
- diff_swagger.py CLI: `python3 diff_swagger.py <old.json> <new.json> [--output path] [--format json|markdown] [--discrepancies path]`
- CRITICAL: Normalize paths by stripping `/api/<version>/` prefix BEFORE comparing — otherwise every path appears as both removed and added
- Use _shared/swagger_utils.py for allOf resolution and path normalization
- Diff output: new_endpoints, removed_endpoints, modified_endpoints, new_schemas, removed_schemas, modified_schemas, field_changes
- Each diff item annotatable: real_change / swagger_artifact / needs_verification
- generate_migration_plan.py: cross-ref diff JSON + ROADMAP.md → actionable plan (update_models, new_resources, deprecated, roadmap_gaps)
- known_discrepancies.md: living document tracking confirmed swagger vs API divergences
- Output JSON + markdown formats
- Python 3.10+ stdlib only
- Skill files go in `.claude/skills/api-diff/`

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `.claude/skills/_shared/swagger_utils.py` — resolve_all_of, normalize_path, flatten_schema
- `swagger-2.22.json` + `swagger-2.23.json` — test inputs for diff
- `ROADMAP.md` — cross-reference target for migration plan

### Integration Points
- `.claude/skills/api-diff/scripts/diff_swagger.py`
- `.claude/skills/api-diff/scripts/generate_migration_plan.py`
- `.claude/skills/api-diff/references/known_discrepancies.md`
- `.claude/skills/api-diff/SKILL.md`
- Phase 47 (api-upgrade) consumes the diff + migration plan output

</code_context>

<specifics>
## Specific Ideas

The swagger is not always accurate vs real FlashBlade API. The skill must:
- Flag endpoints where swagger behavior may differ from known API behavior
- Include a known_discrepancies.md file for manual annotation
- Allow marking diff items as swagger_artifact vs real_change

</specifics>

<deferred>
## Deferred Ideas

None.

</deferred>
