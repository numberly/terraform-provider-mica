# Milestone tools-v1.0: API Tooling Pipeline

**Status:** SHIPPED 2026-04-14
**Phases:** 43-48
**Total Plans:** 9

## Overview

Automated swagger-to-reference conversion, API version diffing, and provider upgrade orchestration through Claude Code skills with Python tooling. Full pipeline: swagger.json → AI-optimized reference → browse → diff → migration plan → dry-run upgrade.

## Phases

### Phase 43: Shared Library
**Goal**: Shared utility library (allOf resolver, path normalizer, schema flattener)
**Plans**: 1/1 complete
- [x] 43-01: swagger_utils.py + 15 pytest tests

### Phase 44: swagger-to-reference Skill
**Goal**: Convert swagger.json to AI-optimized markdown matching FLASHBLADE_API.md
**Plans**: 2/2 complete
- [x] 44-01: parse_swagger.py converter (226 paths, 538 ops)
- [x] 44-02: swagger-to-reference SKILL.md

### Phase 45: API Browsing Tools
**Goal**: Search, inspect, compare endpoints and schemas from generated reference
**Plans**: 1/1 complete
- [x] 45-01: browse_api.py (6 subcommands: endpoints, schema, compare, stats, tags, search)

### Phase 46: api-diff Skill
**Goal**: Structured diff between swagger versions + migration plan
**Plans**: 3/3 complete
- [x] 46-01: diff_swagger.py (16 new endpoints 2.22→2.23)
- [x] 46-02: generate_migration_plan.py + known_discrepancies.md
- [x] 46-03: api-diff SKILL.md

### Phase 47: api-upgrade Skill
**Goal**: Mechanical version updater + 5-phase orchestration
**Plans**: 2/2 complete
- [x] 47-01: upgrade_version.py (38 files in dry-run)
- [x] 47-02: api-upgrade SKILL.md + upgrade_checklist.md

### Phase 48: Integration & Validation
**Goal**: CLAUDE.md update, end-to-end pipeline validation
**Plans**: inline execution
- [x] CLAUDE.md updated with 3 skills + api_references/ + API tools section
- [x] E2E pipeline validated on swagger-2.22.json and swagger-2.23.json

## Milestone Summary

**Key Decisions:**
- Shared Python lib in .claude/skills/_shared/ (not per-skill)
- api-upgrade includes mechanical scripts (not pure orchestration)
- PYTHONPATH=.claude/skills required for _shared imports (standard Python practice)
- Python 3.10+ stdlib only — no external dependencies

**Issues Resolved:**
- allOf resolution for 404/709 schemas (recursive $ref + allOf merging)
- Version prefix normalization prevents false diff positives

**Issues Deferred:**
- None

**Technical Debt:**
- PYTHONPATH requirement for bare python3 invocations (scripts handle internally)

---
*Archived: 2026-04-14*
