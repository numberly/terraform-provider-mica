---
name: swagger-to-reference
description: "Convert a FlashBlade swagger.json (OpenAPI 3.0) into the AI-optimized markdown reference format matching FLASHBLADE_API.md. Handles allOf/$ref resolution, groups endpoints by tag, and emits a compact Data Models section. Use this skill when a new API version swagger file is available and needs to be converted into the project standard reference format."
---

# swagger-to-reference

## Purpose

Converts a FlashBlade OpenAPI 3.0 `swagger.json` file into the AI-optimized markdown reference
format used by this provider (matching `FLASHBLADE_API.md`). Resolves all `$ref`/`allOf`
references, groups endpoints by tag, deduplicates common query parameters into a shared table,
and emits a compact Data Models section. The output is consumed by Claude for API exploration
and serves as the base input for diff and upgrade skills.

## When to Use

- A new FlashBlade API version is available (e.g., `swagger-2.23.json`) and needs a reference file
- Regenerating the reference after swagger corrections or schema fixes
- Bootstrapping an `api_references/<X.XX>.md` file for a version not yet covered

## Prerequisites

- `swagger.json` file present in the project root or at a known path
- `.claude/skills/_shared/swagger_utils.py` present (shared library from Phase 43)
- Python 3.9+ (stdlib only — no pip installs required)

## Workflow

### Step 1 — Ask for the API version (MANDATORY before running the script)

Before running the script, ask the user:

> "What is the API version string for this swagger file? (e.g., 2.22, 2.23)"

Do NOT infer the version from `swagger["info"]["version"]` without confirming with the user.
The answer the user provides becomes the `--version` argument. This is required because
swagger `info.version` may not match the canonical API version used in the project.

### Step 2 — Run the converter

```bash
PYTHONPATH=.claude/skills python3 .claude/skills/swagger-to-reference/scripts/parse_swagger.py \
  <path/to/swagger.json> \
  --version <X.XX> \
  --output api_references/<X.XX>.md
```

Example for version 2.23:

```bash
PYTHONPATH=.claude/skills python3 .claude/skills/swagger-to-reference/scripts/parse_swagger.py \
  swagger-2.23.json \
  --version 2.23 \
  --output api_references/2.23.md
```

### Step 3 — Verify output

```bash
# Check header and summary line
head -3 api_references/<X.XX>.md

# Confirm no unresolved refs (should print 0 or "0 - PASS")
grep -c "\$ref\|allOf" api_references/<X.XX>.md || echo "0 - PASS"

# Check section presence
grep "^## " api_references/<X.XX>.md
```

Expected sections in output:

```
## Auth
## Common Parameters
## Endpoints
## Data Models (Key Resources)
```

### Step 4 — Spot-check path count against swagger

```bash
python3 -c "import json; d=json.load(open('<path/to/swagger.json>')); print(len(d['paths']), 'paths')"
```

The `N paths | M ops` summary line in the generated markdown must match this path count.
If counts differ, check whether the swagger file was truncated or if any paths were filtered.

## Output Format

The generated file has four sections:

| Section | Content |
|---------|---------|
| Header + summary line | `# FlashBlade REST API {version} — AI-Optimized Reference` + base URL, version, path/op counts, auth method |
| `## Auth` | OAuth2, session login/logout, `api_version` endpoint |
| `## Common Parameters` | All unique query params across all operations, deduplicated, sorted alphabetically |
| `## Endpoints` | Grouped by tag (H3 headings, kebab-case → Title Case), HTTP methods in canonical order (GET, POST, PUT, PATCH, DELETE) |
| `## Data Models (Key Resources)` | One entry per resolved schema, fields sorted alphabetically, descriptions truncated at 50 chars |

See `FLASHBLADE_API.md` at the project root as the ground-truth format example.

## Troubleshooting

| Error | Fix |
|-------|-----|
| `ModuleNotFoundError: No module named '_shared'` | Ensure `PYTHONPATH=.claude/skills` is set before the `python3` call |
| Path count mismatch in summary line | Run `python3 -c "import json; print(len(json.load(open('swagger.json'))['paths']))"` and compare |
| `allOf` still present in output | Verify `resolve_all_of` was called; check `self.resolved_schemas` is populated before `_build_data_models()` |
| `KeyError: 'paths'` | Swagger file may be malformed or wrong file provided; validate with `python3 -m json.tool swagger.json` |
| Empty Data Models section | Check that `swagger["components"]["schemas"]` is non-empty and schemas are referenced by at least one path |
