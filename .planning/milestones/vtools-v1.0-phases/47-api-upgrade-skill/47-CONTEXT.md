# Phase 47: api-upgrade Skill - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning
**Mode:** Auto-generated (infrastructure phase — discuss skipped)

<domain>
## Phase Boundary

Claude can mechanically update API version references across the codebase and orchestrate the provider upgrade sequence with explicit review gates. Creates upgrade_version.py script and api-upgrade SKILL.md with 5-phase workflow.

</domain>

<decisions>
## Implementation Decisions

### Claude's Discretion
All implementation choices are at Claude's discretion — pure infrastructure phase.

Key constraints:
- upgrade_version.py CLI: `python3 upgrade_version.py --from <old> --to <new> --project-root <path> [--dry-run|--apply]`
- Must update: `const APIVersion` in internal/client/client.go, mock server version list in internal/testmock/server.go, all mock handler paths `/api/<old>/` → `/api/<new>/` in internal/testmock/handlers/*.go
- Dry-run by default (--apply to execute)
- SKILL.md: 5 named phases (Infrastructure, Schema Updates, New Resources, Deprecations, Documentation) with explicit review gates
- SKILL.md references flashblade-resource-builder for new resources, swagger-to-reference for doc regen
- Python 3.10+ stdlib only
- Script at `.claude/skills/api-upgrade/scripts/upgrade_version.py`
- SKILL.md at `.claude/skills/api-upgrade/SKILL.md`
- Checklist at `.claude/skills/api-upgrade/references/upgrade_checklist.md`

</decisions>

<code_context>
## Existing Code Insights

### Files to update (version references)
- `internal/client/client.go` — `const APIVersion = "2.22"` (line ~21)
- `internal/testmock/server.go` — versions slice in handleAPIVersion
- `internal/testmock/handlers/*.go` — `/api/2.22/` in HandleFunc registrations (20+ files)

### Integration Points
- Consumes diff + migration plan from api-diff skill (Phase 46)
- Invokes flashblade-resource-builder for new resources
- Invokes swagger-to-reference for documentation regeneration

</code_context>

<specifics>
## Specific Ideas

No specific requirements — infrastructure phase.

</specifics>

<deferred>
## Deferred Ideas

None.

</deferred>
