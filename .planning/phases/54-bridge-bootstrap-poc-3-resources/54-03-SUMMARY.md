---
phase: 54-bridge-bootstrap-poc-3-resources
plan: "03"
subsystem: pulumi-bridge
tags: [pulumi, bridge, entry-points, tfgen, embed]
dependency_graph:
  requires: [54-02]
  provides: [54-04]
  affects: [pulumi/provider/cmd]
tech_stack:
  added: []
  patterns:
    - "//go:embed for schema-embed.json and bridge-metadata.json"
    - "pftfbridge.Main signature: (ctx, pkg, ProviderInfo, ProviderMetadata) — no error return"
    - "tfgen.Main signature: (provider, ProviderInfo) — no error return"
key_files:
  created:
    - pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go
    - pulumi/provider/cmd/pulumi-resource-flashblade/main.go
    - pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json
    - pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json
  modified: []
decisions:
  - "pftfbridge.Main does not return error in v3.127.0 — removed if-err pattern from plan template"
  - "BridgeMetadata field deprecated but still accepted — kept for forward compat"
  - "Placeholder {} JSON files for embed targets — overwritten by plan 04 make tfgen"
metrics:
  duration: "~2 minutes"
  completed: "2026-04-22"
  tasks_completed: 2
  tasks_total: 2
  files_created: 4
  files_modified: 0
---

# Phase 54 Plan 03: Binary Entry Points Summary

Two Pulumi bridge binary entry points created and verified to compile against v3.127.0.

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | pulumi-tfgen-flashblade main.go | c264cec | cmd/pulumi-tfgen-flashblade/main.go |
| 2 | pulumi-resource-flashblade main.go + embed placeholders | 6783399 | cmd/pulumi-resource-flashblade/main.go, schema-embed.json, bridge-metadata.json |

## What Was Built

**pulumi-tfgen-flashblade** (build-time schema generator):
- Calls `tfgen.Main("flashblade", flashblade.Provider())`
- No version parameter (PF tfgen differs from SDK v2)

**pulumi-resource-flashblade** (runtime gRPC provider):
- Calls `pftfbridge.Main(ctx, "flashblade", flashblade.Provider(), meta)`
- Embeds `schema-embed.json` and `bridge-metadata.json` via `//go:embed`
- Placeholder `{}` files allow compilation before first `make tfgen` run

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] pftfbridge.Main has no error return in v3.127.0**
- **Found during:** Task 2
- **Issue:** Plan template showed `if err := pftfbridge.Main(...); err != nil { panic(err) }` but the actual signature is `func Main(...) {}` (void, calls os.Exit internally)
- **Fix:** Removed the error-check wrapper — called `pftfbridge.Main(...)` directly
- **Files modified:** cmd/pulumi-resource-flashblade/main.go
- **Commit:** 6783399

## Success Criteria

- [x] BRIDGE-04 satisfied: two entry points exist and compile
- [x] `go build ./cmd/pulumi-tfgen-flashblade` succeeds (binary: 101.5MB)
- [x] `go build ./cmd/pulumi-resource-flashblade` succeeds (binary: 95.7MB)
- [x] Runtime binary embeds placeholder schema (ready to be replaced by plan 04's `make tfgen`)

## Self-Check: PASSED

- [x] pulumi/provider/cmd/pulumi-tfgen-flashblade/main.go — FOUND
- [x] pulumi/provider/cmd/pulumi-resource-flashblade/main.go — FOUND
- [x] pulumi/provider/cmd/pulumi-resource-flashblade/schema-embed.json — FOUND
- [x] pulumi/provider/cmd/pulumi-resource-flashblade/bridge-metadata.json — FOUND
- [x] Commit c264cec — FOUND
- [x] Commit 6783399 — FOUND
