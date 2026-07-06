# Conventions

This page is a navigable summary. The **authoritative** rules live in
[`CONVENTIONS.md`](../CONVENTIONS.md) (code layout, structs, tests) and
[`CLAUDE.md`](../CLAUDE.md) (project instructions). When they conflict with this
page, they win. The conventions are strict on purpose: they make every resource
identical in shape, so adding/upgrading one is mechanical and reviewable. The
`flashblade-resource-builder` skill encodes the full lifecycle.

## File layout (one file per resource, per layer)

| Component | Location |
|-----------|----------|
| Model structs | `internal/client/models_<domain>.go` (append to domain file) |
| Client CRUD | `internal/client/<resource>.go` |
| Client tests | `internal/client/<resource>_test.go` |
| Mock handler | `internal/testmock/handlers/<resource>.go` |
| Resource / DS | `internal/provider/<resource>_resource.go` / `_data_source.go` |
| Tests | `internal/provider/<resource>_{resource,data_source}_test.go` |
| Examples | `examples/resources/flashblade_<resource>/{resource.tf,import.sh}` |
| Generated docs | `docs/…` — **never hand-edit** (`make docs`) |

## Model structs

Three per resource — `Target` (GET), `TargetPost` (POST), `TargetPatch` (PATCH).
The pointer rules are the sharpest edge in the codebase:

- **GET**: no pointers on scalars; `NamedReference` / `*NamedReference` for refs.
- **POST**: plain + `omitempty`; use `*bool` only when the API default is `true`
  (so `omitempty` doesn't drop a deliberate `false`); `*int64`/`*string` when
  zero is a valid choice (e.g. `VLAN=0`). Name field is `json:"-"` (sent via
  `?names=`).
- **PATCH**: **every field a pointer**. `nil` = omit, non-nil = send.
  `*string` scalars; `**NamedReference` for refs (outer nil = omit, outer set +
  inner nil = clear to null); `*[]T`+`omitempty` for lists (`nil` omit, `&[]T{}`
  clear, populated = set). Exception — "always send" lists (`NetworkInterfacePatch.Services`)
  use plain `[]T`.

## Client CRUD

Signatures are `Get/Post/Patch/Delete(ctx, name, [body])`. GET always uses
`getOneByName[T]`. Names go via `?names=` + `url.QueryEscape`. **No API-version
prefix in paths** (the client adds it). Return `APIError` directly — no
`fmt.Errorf` wrapping. Always propagate the caller `ctx` (never
`context.Background()` in auth paths). Pick one of three list shapes
(`ListXxxOpts` struct / plain parent string / `ctx`-only) — don't invent a fourth.

## Mock handlers

One store per resource: `sync.Mutex` + `byName` map + `nextID`.
`RegisterXxxHandlers(mux)` returns the store (for `Seed()`). Paths **include**
the API version (`/api/2.23/...`). **Critical**: a `?names=` miss returns an
empty list with HTTP 200 (not 404) — this matches the real API and lets
`getOneByName[T]` detect not-found. Use the shared helpers in
`handlers/helpers.go`.

## Resource implementation

- All four interfaces mandatory (`Resource`, `…WithConfigure`,
  `…WithImportState`, `…WithUpgradeState`); schema `Version` starts at 0.
- **Plan modifiers**: `UseStateForUnknown()` on stable computed (`id`, `created`);
  `RequiresReplace()` on immutable `name`; **none** on volatile fields
  (`status`, `lag`, `recovery_point`, `backlog`) — a modifier there masks drift.
- **Timeouts**: all four ops; defaults Create 20m / Read 5m / Update 20m /
  Delete 30m.
- **Drift detection is mandatory**: every mutable/computed field logs
  `tflog.Debug(ctx, "drift detected", …)` on `Read`. Log only, never error.
- **ImportState**: by name; always call `nullTimeoutsValue()`; null write-once
  and sensitive fields.
- **Soft-delete** only for buckets & file systems.

## State upgraders

Bump `SchemaVersion` on any attribute add/change/rename. Naming: `xxxV0Model`,
`xxxV1Model`, current = `xxxModel`. `PriorSchema` must be an exact copy of that
version's schema. New fields default to `types.StringNull()` /
`types.ListNull(...)`. The chain runs sequentially (0→1→2); entry key = prior
version number. Reference: `server_resource.go` (v0→v1→v2).

> **Repo idiom** (from project memory): write the upgrader as
> `newState := currentModel(oldState)` — a defensive cast that fails *loud* at
> the next schema change. Do **not** regress to an explicit struct literal
> (silent zero-init bugs; staticcheck S1016).

## Data sources

Two interfaces only (`DataSource`, `…WithConfigure`). No timeouts, no plan
modifiers. `name` Required, rest Computed. Not-found → `AddError`.

## Git & tooling

- **Conventional Commits** (`feat:`/`fix:`/`chore:`/`docs:`/`perf:`).
  **Never** add `Co-Authored-By` trailers — a `commit-msg` hook enforces this.
- Subagent commits use `--no-verify`.
- Rebase on the target branch before pushing / opening a PR.
- `gofmt` is **not** enforced repo-wide and the source isn't gofmt-clean; don't
  reformat whole files — trust `make lint` (`.golangci.yml`).
- **Code navigation uses Serena MCP** (project rule) — `find_symbol` /
  `get_references` over Grep/Glob for exploration.

## New-resource checklist (abridged)

Models → client CRUD (`getOneByName`) → mock handler (empty-list GET, `Seed`) →
client tests ≥4 → resource (4 interfaces, drift, ImportState) → resource tests ≥3
→ data source + test → register in `provider.go` → HCL examples → `make docs` →
`make test` (≥ baseline) → `make lint` → update `ROADMAP.md` **in the same
commit**. Full checklist: [`CONVENTIONS.md`](../CONVENTIONS.md).
