# Requirements: API Tooling Pipeline

**Defined:** 2026-04-14
**Core Value:** Automate API reference generation, version comparison, and provider upgrade orchestration

## tools-v1.0 Requirements

### Swagger Conversion

- [x] **CONV-01**: Convert an OpenAPI 3.0 swagger.json into AI-optimized markdown matching existing FLASHBLADE_API.md format
- [x] **CONV-02**: Recursively resolve allOf/$ref schemas (404/709 schemas use allOf in 2.22)
- [x] **CONV-03**: Ask user for API version before processing
- [x] **CONV-04**: Generate output to `api_references/<version>.md`

### API Browsing

- [x] **BRWS-01**: Search endpoints by tag, HTTP method, or text pattern
- [x] **BRWS-02**: Display schema details (fields, types, readOnly annotations)
- [x] **BRWS-03**: Compare two schemas side-by-side (e.g., Post vs Patch)
- [x] **BRWS-04**: Display reference statistics (path count, schema count, method distribution)

### API Diffing

- [ ] **DIFF-01**: Compare two swagger files and produce structured diff (new/removed/modified endpoints + schemas)
- [ ] **DIFF-02**: Normalize paths (strip `/api/<version>/` prefix) before comparison
- [ ] **DIFF-03**: Annotate diff items as `real_change` / `swagger_artifact` / `needs_verification` via known_discrepancies.md
- [ ] **DIFF-04**: Generate migration plan cross-referenced with ROADMAP.md

### API Upgrade

- [ ] **UPGR-01**: Update `const APIVersion` in client.go, mock server versions, and mock handler paths automatically
- [ ] **UPGR-02**: Dry-run by default, --apply to execute
- [ ] **UPGR-03**: Orchestrate upgrade in 5 phases with review gates (infra → schemas → new resources → deprecations → docs)

### Shared Library

- [x] **SLIB-01**: Provide shared utilities (allOf resolver, path normalizer, schema flattener) in `.claude/skills/_shared/swagger_utils.py`
- [x] **SLIB-02**: Python 3.10+ stdlib only, no external dependencies

### Integration

- [ ] **INTG-01**: Update CLAUDE.md with API tools and `api_references/` convention
- [x] **INTG-02**: Create 3 SKILL.md files following skill-creator format (YAML frontmatter, structured sections)

## Future Requirements

### Enhanced Tooling

- **ETOOL-01**: Auto-detect swagger format (OpenAPI 3.0 vs Swagger 2.0) and handle both
- **ETOOL-02**: Validate generated reference against original swagger (round-trip check)
- **ETOOL-03**: Interactive discrepancy annotation workflow

## Out of Scope

| Feature | Reason |
|---------|--------|
| Automated code generation from swagger | Too error-prone, provider patterns require human judgment |
| Acceptance test generation | Requires real FlashBlade array access |
| CI/CD integration | Skills are dev-time tools, not pipeline components |
| Swagger correction/patching | Not our swagger to fix — track discrepancies instead |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SLIB-01 | Phase 43 | Complete |
| SLIB-02 | Phase 43 | Complete |
| CONV-01 | Phase 44 | Complete |
| CONV-02 | Phase 44 | Complete |
| CONV-03 | Phase 44 | Complete |
| CONV-04 | Phase 44 | Complete |
| BRWS-01 | Phase 45 | Complete |
| BRWS-02 | Phase 45 | Complete |
| BRWS-03 | Phase 45 | Complete |
| BRWS-04 | Phase 45 | Complete |
| DIFF-01 | Phase 46 | Pending |
| DIFF-02 | Phase 46 | Pending |
| DIFF-03 | Phase 46 | Pending |
| DIFF-04 | Phase 46 | Pending |
| UPGR-01 | Phase 47 | Pending |
| UPGR-02 | Phase 47 | Pending |
| UPGR-03 | Phase 47 | Pending |
| INTG-01 | Phase 48 | Pending |
| INTG-02 | Phase 44, 46, 47, 48 | Complete |

**Coverage:**
- tools-v1.0 requirements: 19 total
- Mapped to phases: 19
- Unmapped: 0

---
*Requirements defined: 2026-04-14*
*Last updated: 2026-04-14 after initial definition*
