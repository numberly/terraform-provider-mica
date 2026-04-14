---
name: api-diff
description: "Compare two FlashBlade swagger.json files to produce a structured diff of endpoint and schema changes, annotate items as real_change/swagger_artifact/needs_verification, and generate a migration plan cross-referenced with ROADMAP.md."
---

# api-diff

## Purpose

Detects breaking changes, new endpoints, and schema field additions between two FlashBlade API versions.
Produces a structured JSON diff consumed by the api-upgrade skill.
Annotates each diff item with confidence: real_change (confirmed), swagger_artifact (doc error), needs_verification (unknown).

## When to Use

- New swagger file available for a new FlashBlade API version
- Determining which provider changes are needed before an upgrade
- Annotating diff items with known discrepancy status
- Generating a migration plan to update provider models and add new resources

## Prerequisites

- swagger-OLD.json and swagger-NEW.json in project root (or at known paths)
- .claude/skills/_shared/swagger_utils.py present (from Phase 43 shared library)
- Python 3.10+ (stdlib only — no pip installs required)

## Workflow

### Step 1 — Run the diff

```bash
PYTHONPATH=.claude/skills python3 .claude/skills/api-diff/scripts/diff_swagger.py \
  swagger-2.22.json swagger-2.23.json \
  --format json \
  --output /tmp/diff-2.22-2.23.json
```

Note: Path prefix normalization is automatic — /api/2.22/buckets and /api/2.23/buckets are treated as the same endpoint.

### Step 2 — Review and annotate discrepancies (optional)

Edit `.claude/skills/api-diff/references/known_discrepancies.json` to add annotation overrides for known swagger inaccuracies. Then re-run with the `--discrepancies` flag:

```bash
PYTHONPATH=.claude/skills python3 .claude/skills/api-diff/scripts/diff_swagger.py \
  swagger-2.22.json swagger-2.23.json \
  --format json \
  --discrepancies .claude/skills/api-diff/references/known_discrepancies.json \
  --output /tmp/diff-annotated.json
```

### Step 3 — Generate migration plan

```bash
python3 .claude/skills/api-diff/scripts/generate_migration_plan.py \
  /tmp/diff-annotated.json \
  ROADMAP.md \
  --format markdown \
  --output /tmp/migration-plan.md
```

Note: roadmap_gaps entries are new API endpoints that match existing ROADMAP.md Candidate entries — these should be prioritized for scheduling.

## Output

- **diff.json**: structured diff with 6 categories (new/removed/modified endpoints and schemas). Each item has an `annotation` field set to `real_change`, `swagger_artifact`, or `needs_verification`.
- **migration-plan.json or migration-plan.md**: 4-category actionable plan — `update_models`, `new_resources`, `deprecated`, `roadmap_gaps`.

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| All paths appear as new+removed | Version prefix not stripped | Use current diff_swagger.py — normalize_path() is called automatically |
| ROADMAP.md parse error | File not found in working directory | Run from project root or pass absolute path |
| ImportError: No module named swagger_utils | Missing PYTHONPATH | Prefix command with PYTHONPATH=.claude/skills |
