---
phase: 56
plan: 01
name: Schema Artifacts and Makefile Targets
subsystem: pulumi-bridge
tags: [pulumi, makefile, schema, sdk]
requires: []
provides: [56-02, 56-03]
affects: [pulumi/Makefile, pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json]
tech-stack:
  added: [jq]
  patterns: []
key-files:
  created: [pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json]
  modified: [pulumi/Makefile]
key-decisions:
  - Use jq -c for schema-embed.json minification (simple, no extra deps)
  - Use tfgen binary language subcommands for SDK generation (standard bridge pattern)
  - Preserve go.mod during Go SDK generation via backup/restore
requirements-completed: [SDK-03, SDK-04]
duration: 12 min
completed: 2026-04-22
---

# Phase 56 Plan 01: Schema Artifacts and Makefile Targets Summary

**One-liner:** Added schema-embed.json generation and Python/Go SDK Makefile targets with scope boundary enforcement.

## What Was Built

1. **schema-embed.json generation** — `make tfgen` now minifies `schema.json` into a single-line `schema-embed.json` via `jq -c`. Required for SDK generation and embedding.
2. **generate_python target** — Runs `tfgen python --out sdk/python --skip-docs --skip-examples`. Produces a `pulumi_flashblade` package with `setup.py`.
3. **generate_go target** — Runs `tfgen go --out sdk/go --skip-docs --skip-examples`. Preserves existing `go.mod` via backup/restore.
4. **test_python_sdk / test_go_sdk targets** — Verification wrappers that build the SDKs and confirm artifacts (wheel compilation for Python, `go build ./...` for Go).
5. **SDK-04 scope boundary** — Makefile explicitly documents that NodeJS, .NET, and Java are out of scope. No targets for those languages exist.

## Tasks Executed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Add schema-embed.json generation to tfgen target | ac498cf |
| 2 | Add generate_python Makefile target | db89bcb |
| 3 | Add generate_go Makefile target | db89bcb |
| 4 | Ensure no NodeJS/C#/Java targets exist | db89bcb |
| 5 | Commit schema artifacts and Makefile changes | ac498cf, db89bcb |

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## Verification

- `make tfgen` produces non-empty `schema-embed.json` (1 line, 174.6K)
- `make generate_python` exits 0, creates `sdk/python/pulumi_flashblade/` with 54 resource files
- `make generate_go` exits 0, creates `sdk/go/flashblade/` with 54 resource files
- `grep -E 'generate_(nodejs|dotnet|java)' pulumi/Makefile` returns no matches

## Next

Ready for Wave 2: Python SDK wheel build (56-02) and Go SDK compilation (56-03).
