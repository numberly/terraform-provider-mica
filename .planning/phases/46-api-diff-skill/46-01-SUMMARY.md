---
phase: 46-api-diff-skill
plan: "01"
subsystem: skills/api-diff
tags: [python, swagger, diff, cli, tooling]
dependency_graph:
  requires: [swagger_utils.py (Phase 43)]
  provides: [diff_swagger.py CLI, structured diff JSON]
  affects: [plan 46-02 (generate_migration_plan.py consumes diff output)]
tech_stack:
  added: []
  patterns: [normalize_path for version-agnostic keying, resolve_all_of + flatten_schema for schema diff]
key_files:
  created:
    - .claude/skills/api-diff/scripts/diff_swagger.py
  modified: []
decisions:
  - HTTP_METHODS whitelist filters non-operation path item fields (parameters, summary, servers)
  - Schema comparison uses flatten_schema from shared lib to resolve $ref/allOf before field diffing
  - Endpoint annotation defaults to needs_verification; overrideable via --discrepancies JSON
metrics:
  duration: "~5 minutes"
  completed: "2026-04-14T08:33:26Z"
  tasks_completed: 1
  tasks_total: 1
  files_created: 1
  files_modified: 0
---

# Phase 46 Plan 01: diff_swagger.py CLI Tool Summary

**One-liner:** Swagger diff CLI using normalize_path + resolve_all_of to produce zero-false-positive JSON/markdown diffs across API versions.

## Tasks Completed

| Task | Name | Commit | Files |
| ---- | ---- | ------ | ----- |
| 1 | Create diff_swagger.py | a17ba13 | .claude/skills/api-diff/scripts/diff_swagger.py |

## What Was Built

`diff_swagger.py` is a CLI script that:
- Loads two swagger JSON files and builds endpoint maps keyed by `(normalize_path, method)` — stripping version prefix to avoid false positives
- Computes six diff categories: new/removed/modified endpoints, new/removed/modified schemas
- For modified endpoints: compares summary, parameter names, requestBody presence, response codes
- For modified schemas: uses `flatten_schema` to compare property keys and types
- Supports `--format json` (default, structured) and `--format markdown` (human-readable summary table + section tables)
- Supports `--discrepancies` JSON file to override `annotation` fields on matched `(normalized_path, method)` pairs
- Writes to `--output file` or stdout

Verified:
- Self-diff on swagger-2.22.json → all-zero summary (no false positives from version prefix)
- Real 2.22→2.23 diff: 16 new endpoints, 19 modified, 14 new schemas, 10 modified schemas
- All 6 category keys present; all items have required fields
- Markdown output renders valid tables with section headers

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

- `.claude/skills/api-diff/scripts/diff_swagger.py` exists: FOUND
- Commit a17ba13: FOUND (git log confirms)
- Self-diff all-zero: VERIFIED
- Real diff structure valid: VERIFIED
