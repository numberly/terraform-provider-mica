---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: executing
last_updated: "2026-04-22T09:16:07.231Z"
last_activity: 2026-04-22
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 5
  completed_plans: 1
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-21)

**Core value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises.
**Current focus:** Phase 54 — bridge-bootstrap-poc-3-resources

## Current Position

Milestone: pulumi-2.22.3 (Pulumi Bridge Alpha)
Phase: 54 (bridge-bootstrap-poc-3-resources) — EXECUTING
Plan: 2 of 5
Status: Ready to execute
Last activity: 2026-04-22

Progress: [██░░░░░░░░] 20% (1/5 plans)

## Recent Milestones

- ✅ **v2.22.3** — Convention Compliance (shipped 2026-04-20, 779 tests, 12/12 requirements, [archive](milestones/v2.22.3-ROADMAP.md))
- ✅ **v2.22.2** — Directory Service Roles & Role Mappings (shipped 2026-04-17, 818 tests, [archive](milestones/v2.22.2-ROADMAP.md))
- ✅ **v2.22.1** — Directory Service – Array Management (shipped 2026-04-17, 798 tests, [archive](milestones/v2.22.1-ROADMAP.md))

## Accumulated Context

### Key Decisions (pulumi-2.22.3)

- **Module path:** `github.com/numberly/opentofu-provider-flashblade` (TF provider root). Bridge modules: `./pulumi/provider/` and `./pulumi/sdk/go/` each with own `go.mod`; TF provider wired via `replace ../../`.
- **Bridge versions:** `pulumi-terraform-bridge/v3 v3.127.0`, `pulumi/sdk/v3 v3.231.0`, `pulumi/pkg/v3 v3.231.0`. Replace SHA: `v2.0.0-20260318212141-5525259d096b`.
- **Schema commit policy:** `schema.json`, `schema-embed.json`, `bridge-metadata.json` committed to git. CI enforces `git diff --exit-code` after `make tfgen`.
- **Secrets pattern:** `Secret: tfbridge.True()` + `AdditionalSecretOutputs` belt-and-braces (Write-Only Fields pattern deferred).
- **Composite IDs:** All 4 composite resources use `/` separator with string IDs (NOT colon + integer). Verified against `readIntoState` in `internal/provider/`.
- **No SetAutonaming:** Storage names are operational identifiers — no random suffix.
- **Soft-delete defense:** `DeleteTimeout: 30*time.Minute` on bucket + filesystem (bridge default 5min kills `pollUntilGone`).
- **SDK scope:** Python + Go only. No TS, C#, Java. No PyPI, npm, NuGet, Pulumi Registry.
- **Distribution:** GitHub Releases private. Go SDK via git tag `sdk/go/vX.Y.Z` + `GOPRIVATE`. Python SDK via `.whl` attached to release.

### Critical Pitfalls (pre-mitigated by phase design)

- **PB1 (CRITICAL):** Default 5-min `DeleteTimeout` kills `pollUntilGone` → mitigated in Phase 54 (SOFTDELETE-01) and Phase 55 (SOFTDELETE-02/03).
- **PB2 (CRITICAL):** Wrong composite ID separator → all `ComputeID` implementations must read `readIntoState` first. `/` separator, string rule names.
- **PB3 (HIGH):** Secret-ness lost on state update → `Secret: tfbridge.True()` + `AdditionalSecretOutputs` on all 6 sensitive fields.
- **PB4 (MEDIUM):** Replace SHA coupled to bridge version → must re-verify SHA on every bridge bump.
- **PB5 (HIGH):** Go SDK `go get` requires `sdk/go/vX.Y.Z` tag in addition to release tag → post-goreleaser step in `pulumi-release.yml`.
- **PB7 (MEDIUM):** `timeouts {}` leaks into SDK → `omitTimeoutsOnAll` helper applied BEFORE first `make tfgen`.

### Open Blockers

_(none — 3 open questions from research resolved via REQUIREMENTS.md decisions: module path = numberly, schema committed, Write-Only Fields deferred)_

## Next Steps

Execute plan 02: `resources.go` ProviderInfo + pf.ShimProvider wiring, `go mod tidy`.

## Session Log

- 2026-04-22T09:15Z — Plan 54-01 completed: pulumi/ module skeleton (3 go.mod files, Makefile, .gitignore)
