# API versioning

The provider tracks the FlashBlade REST API version explicitly. The current
target is **v2.23** (Purity//FB 4.6.7+); **v2.22** is maintained in parallel.
Because the version prefix is injected centrally in the client (see
[architecture.md](architecture.md#client-http-layer-internalclient)), a version
bump is a small code change plus per-resource schema work — driven by tooling
and skills so it stays consistent.

## Reference artifacts

| File | What |
|------|------|
| `swagger-2.22.json`, `swagger-2.23.json` | Raw OpenAPI 3.0 inputs (repo root) |
| `api_references/2.22.md`, `2.23.md` | Per-version AI-optimized references (~220–233 KB) generated from swagger |
| `FLASHBLADE_API.md` | Standing AI-optimized API reference (226 paths, 538 ops) |

## Tooling (skills in `.claude/skills/`)

The version-upgrade pipeline is three skills (invoked via the Skill tool /
`/gsd-*` flow):

1. **`swagger-to-reference`** — converts a `swagger.json` into the markdown
   reference format (resolves `allOf`/`$ref`, groups by tag, compact Data Models).
   Run when a new API version's swagger lands →
   produces `api_references/<ver>.md`.
2. **`api-diff`** — diffs two swagger files, annotates each change as
   `real_change` / `swagger_artifact` / `needs_verification` (swagger has lied
   before — see the memory note below), and emits a migration plan
   cross-referenced with `ROADMAP.md`.
3. **`api-upgrade`** — orchestrates the whole provider bump through **6
   review-gated phases**: (1) infra version bump, (2) schema updates, (3) new
   resources, (4) deprecations, (5) documentation, (6) Pulumi bridge alignment.
   It consumes the `api-diff` plan and delegates new-resource implementation to
   `flashblade-resource-builder` (and the `flashblade-resource-implementor` /
   `flashblade-resource-modifier` agents).

Helper scripts (see [`CLAUDE.md`](../CLAUDE.md) "API Reference Tools") wrap
`parse_swagger.py`, `browse_api.py`, `diff_swagger.py`,
`generate_migration_plan.py`, and `upgrade_version.py`.

> **Memory / hard-won lesson**: the mock server *and* swagger have both
> misrepresented the real API (e.g. the object-lock `enabled` field; CORS
> empty-`id` responses). Always confirm against a real array before releasing —
> see [testing.md](testing.md#three-tiers).

## Dual release lines (2.22 + 2.23)

Two API versions are maintained simultaneously, each with its own release
branches:

- `release/v2.22.x` (+ paired `-pulumi.beta` tags) and `release/v2.23.x`.
- Features generally ship to both: implement on 2.23, then cherry-pick to 2.22
  and `sed` the mock-handler paths `/api/2.23/ → /api/2.22/` (project memory).
- `ROADMAP.md` tracks coverage for the current line (~78 % of IaC-relevant CRUD
  at v2.23.0).

The TF-provider and Pulumi-bridge release lines are **separate** tag namespaces
(`v*` vs `v*-pulumi*`) with separate workflows — see
[workflow-and-release.md](workflow-and-release.md#build--release).

## `ROADMAP.md` is mandatory to update

When adding a resource/data source, in the **same commit**: move the entry to
"Implemented", fill Status/Notes, bump the header counters + `Last updated`, and
run `make docs`. This keeps the roadmap accurate across contributors and
sessions.
