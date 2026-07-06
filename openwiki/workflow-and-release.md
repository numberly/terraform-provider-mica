# Workflow & release

This page covers how work is planned and shipped: the GSD structured-development
workflow (`.planning/`), the build/CI/release pipelines, and docs generation.

## GSD development workflow (`.planning/`)

Development is organized as **milestones → phases → plans**, state-tracked in
`.planning/`. This is the "harness → gsd → agents → code → review" loop:

```
harness (orchestrator)
  └─ GSD  (milestone → phase → numbered plans)          .planning/
       └─ specialized agents + skills                    flashblade-resource-builder,
            └─ code                                        api-upgrade, gsd-* …
                 └─ review / verification gate            *-VERIFICATION.md, UAT
```

| File | Role |
|------|------|
| `.planning/PROJECT.md` | Living charter: what/why, validated vs active vs out-of-scope requirements, key decisions |
| `.planning/STATE.md` | Current position: active milestone, phase/plan progress, decisions, todos, session log |
| `.planning/ROADMAP.md` | Milestone list (v1.0 → v2.23.0); archives in `.planning/milestones/` |
| `.planning/phases/<NN>-…/` | One dir per phase |

Each phase follows a **plan → execute → summarize → verify** cadence: numbered
plans have `<NN>-<MM>-PLAN.md` + `-SUMMARY.md`, and the phase has
`-VERIFICATION.md` / `-HUMAN-UAT.md` (plus specialized outputs like `-RETRO.md`,
`-PR.md`, `-MERGE.md`, `-RELEASE.md`). GSD skills (`gsd-new-milestone`,
`gsd-plan-phase`, `gsd-execute-phase`, …, exposed as `mcp__gsd-workflow__*`) are
the orchestration layer; execution delegates to code-writing agents; review is
captured in the verification/UAT files and code-review gates before merge.

Example: v2.23.0 shipped via **Phase 59** (API 2.23 consolidation & validation)
and **Phase 60** (release/merge).

Not part of the provider's runtime — this is process/state only.

## Build & release

Root build file: **`GNUmakefile`** (not `Makefile`). Key targets:

| Target | Action |
|--------|--------|
| `build` | `go build -trimpath -o terraform-provider-mica` |
| `test` | `go test ./internal/...` + `TEST_BASELINE` regression gate |
| `testacc` | `TF_ACC=1 go test ./... -timeout 120m` |
| `lint` | `golangci-lint run ./...` |
| `generate` / `docs` | `go generate ./...` (docs is an alias) |
| `install` | build + copy into `~/.terraform.d/plugins/…/numberly/mica/dev/` |
| `install-hooks` | `core.hooksPath = scripts/git-hooks` (rejects `Co-Authored-By`) |

**GoReleaser** cuts releases:

- `.goreleaser.yml` — TF provider: linux/darwin/windows × amd64/arm64,
  SHA256SUMS + **GPG detached signature** (Terraform Registry requirement),
  `terraform-registry-manifest.json`.
- `.goreleaser.pulumi.yml` — Pulumi bridge artifacts.

**CI (`.github/workflows/`)**:

| Workflow | Trigger | Does |
|----------|---------|------|
| `ci.yml` | push/PR to main | unit tests + coverage; `workflow_call`-able release gate |
| `release.yml` | `v*` tags (excl. `-pulumi*`) | CI gate → GoReleaser + Cosign keyless (Sigstore) signing |
| `pulumi-ci.yml` | path-filtered (`pulumi/**`, `internal/**`) | bridge lint/test |
| `pulumi-prerequisites.yml` | `workflow_call` | provider tests, schema-drift gate, `go mod tidy` gate, uploads schema artifacts |
| `pulumi-release.yml` | `v*-pulumi*` tags | prerequisites + CHANGELOG preflight → provider binary + Python wheel + independent Go SDK tag |

See [api-versioning.md](api-versioning.md#dual-release-lines-222--223) for the
dual 2.22/2.23 branch story.

### Cutting a release (2.22.X, 2.23.X, Pulumi)

There is **no `VERSION` file** — the version is the **git tag**, injected into the
binary by GoReleaser (`-X main.version={{.Version}}`). A release is therefore
just: a `release/vX.Y.Z` branch, one `chore(release): vX.Y.Z changelog` commit,
and an **annotated tag** whose push triggers the workflow.

Three lines ship in parallel, each with its own tag namespace:

| Line | Base branch | `APIVersion` | Tag | Workflow |
|------|-------------|--------------|-----|----------|
| **2.23.X** | `main` | `2.23` | `vX.Y.Z` | `release.yml` → Terraform Registry |
| **2.22.X** | `release/v2.22.<prev>` | `2.22` | `vX.Y.Z` | `release.yml` → Terraform Registry |
| **Pulumi** | `release/v2.22.<this>` | `2.22` | `vX.Y.Z-pulumi.beta` | `pulumi-release.yml` → GitHub release + PyPI wheel + Go SDK tag |

**⚠️ The 2.22 line is a frozen, reduced feature set — not "the same code with
`APIVersion=2.22`".** It lacks resources that landed after the 2.22→2.23 API bump
(e.g. `workload`, `resiliency_group`, and any resource added on the 2.23 line).
A new resource is **not** automatically present on 2.22; it must be **ported**.
`git diff release/v2.22.<n> main` is large by design — do not "reconcile" it.

**Before cutting from `main`, check for divergence.** `main` is the integration
line; the `release/v2.2x.Y` branches can carry hotfixes (and their CHANGELOG
history) that were never forward-ported to `main`, and `main`'s `CHANGELOG.md`
often lags the release lines. Verify `main` is a functional superset of the
latest release on that line (`git log --oneline release/vX.Y.Z ^main --no-merges`)
and forward-port any missing fixes first, else the new tag **regresses** shipped
fixes. Sync the release branch's `CHANGELOG.md` from the previous release
(`git checkout <prev-release> -- CHANGELOG.md`) then prepend the new entry.

**Per-line procedure:**

1. **2.23.X** (clean, from `main`): branch `release/vX.Y.Z` off `main`; sync +
   prepend `CHANGELOG.md`; `make build && make test && make lint`; annotated tag.

2. **2.22.X** (backport): branch off the previous `release/v2.22.*`; **port** the
   new resource file-by-file from `main` (new files copy cleanly; add struct
   blocks to `models_*.go` and registrations to `provider.go` surgically). **The
   2.22 line has no `handlers.APIPrefix` const** (that refactor lives only on
   `main`) — mock handlers there hardcode `/api/2.22/…`, so change any
   `APIPrefix+"/…"` to the literal. Then `make -C pulumi tfgen` (regenerate
   schema) + bump `expectedResources`/`expectedDataSources` in
   `pulumi/provider/resources_test.go` (the 2.22 line uses magic numbers, not the
   derived counts main has). `make build/test/lint`; changelog; annotated tag.

3. **Pulumi** (`vX.Y.Z-pulumi.beta`): branch off the `release/v2.22.<this>` you
   just cut; add an entry to **`pulumi/CHANGELOG.md`** (`## vX.Y.Z-pulumi.beta`,
   describing `mica:index:*` tokens — sync its history from the previous
   `*-pulumi.beta` first); commit; annotated tag. `pulumi-release.yml` rebuilds
   the schema/SDKs from source, so the regenerated schema on the 2.22 branch is
   what surfaces.

**Always validate against a real array (both API 2.22 and 2.23) before tagging** —
the mock and swagger have both misrepresented the real API before, shipping broken
releases. See [testing.md](testing.md#three-tiers).

## Docs generation

`docs/` is generated by HashiCorp **`tfplugindocs`** (pinned via `tools/tools.go`).
The `//go:generate` directive is in `main.go`; invoked through
`make docs` → `go generate ./...`. Generated files carry a
`# generated by … terraform-plugin-docs` header — **never hand-edit them**.
Inputs are the schema field descriptions and `examples/`:

- `examples/provider/provider.tf` → rendered into `docs/index.md`
- `examples/resources/flashblade_<name>/{resource.tf,import.sh}`
- `examples/data-sources/flashblade_<name>/data-source.tf`
- `examples/workflows/` — multi-resource end-to-end examples (also linked from
  the README)

To change docs, edit the schema descriptions or the `examples/` inputs and
regenerate. Import is always documented by **name**, not UUID.
