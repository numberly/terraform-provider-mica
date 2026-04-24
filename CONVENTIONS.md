# CONVENTIONS.md — Terraform Provider FlashBlade

Single source of truth for code conventions. Every resource/data source/modification MUST follow these rules. Code examples: see `flashblade-resource-builder` skill or referenced files.

## File Structure

| Component | Location | Naming |
|-----------|----------|--------|
| Model structs | `internal/client/models_<domain>.go` | Append to existing domain file |
| Client CRUD | `internal/client/<resource>.go` | One file per resource |
| Client tests | `internal/client/<resource>_test.go` | Same package (`client_test`) |
| Mock handler | `internal/testmock/handlers/<resource>.go` | One file per resource |
| Resource | `internal/provider/<resource>_resource.go` | One file per resource |
| Resource tests | `internal/provider/<resource>_resource_test.go` | Same package (`provider`) |
| Data source | `internal/provider/<resource>_data_source.go` | One file per resource |
| Data source tests | `internal/provider/<resource>_data_source_test.go` | Same package (`provider`) |
| HCL examples | `examples/resources/flashblade_<resource>/resource.tf` | + `import.sh` |
| HCL DS examples | `examples/data-sources/flashblade_<resource>/data-source.tf` | |
| Generated docs | `docs/resources/<resource>.md` | **Never edit manually** |

Reference: `internal/provider/provider.go` for registration pattern.

## Model Structs

Three structs per resource in `internal/client/models_<domain>.go`:

| Struct | Suffix | Purpose | Field types |
|--------|--------|---------|-------------|
| `Target` | (none) | GET response | Plain types, `NamedReference` for refs |
| `TargetPost` | `Post` | POST body | Plain types, `omitempty` on optional |
| `TargetPatch` | `Patch` | PATCH body | ALL fields are `*string` or `**NamedReference` |

### Pointer rules

- **GET**: No pointers on scalars. `NamedReference` for refs, `*NamedReference` for optional refs.
- **POST**: No pointers on scalars except: `*bool` when API default is `true`/non-zero (omitempty drops `false`); `*int64`/`*string` when zero is a valid user choice (e.g. `VLAN=0`). Pointers on nested structs (`*NFSConfig`) and optional refs (`*NamedReference`) are expected. Name field: `json:"-"` (goes via `?names=` query param).
- **PATCH**: Every field is a pointer. `nil` = omit, non-nil = send.
  - `*string` for scalars
  - `**NamedReference` for refs (outer nil = omit, outer non-nil + inner nil = clear to null, both non-nil = set value)
  - `*NestedConfig` for atomic config blocks
  - `*[]T` with `omitempty` for lists: `nil` = omit, `&[]T{}` = clear, `&[...]` = set. `Update()` must assign `&slice` so empty `ElementsAs` transmits `[]`.

**Exception — "always send" lists**: When API treats absent key ≠ empty list, use plain `[]T` without `omitempty`. Known: `NetworkInterfacePatch.Services`, `NetworkInterfacePatch.AttachedServers` (`models_network.go`). Default to `*[]T`+`omitempty` for new fields.

### JSON tags

`json:"field_name"` | `json:"field_name,omitempty"` | `json:"-"` (name excluded from body)

Reference: `internal/client/models_storage.go`.

## Client CRUD Methods

One file per resource. Signatures: `Get/Post/Patch/Delete` on `*FlashBladeClient`, always `(ctx, name, [body])` → `(*Xxx, error)` or `error`.

**Rules**: GET uses `getOneByName[T]` (never hand-roll). Name via `?names=` + `url.QueryEscape`. No API version prefix in paths (added by `client.do()`). Return `APIError` directly — no `fmt.Errorf` wrapping. Always propagate caller `ctx`.

**List shapes** (pick one, do not invent a fourth):

| Shape | When | Example |
|-------|------|---------|
| `ListXxxOpts` struct | API has filters/pagination | `ListBuckets(ctx, opts)` |
| Plain string parent | Sub-collection scoped to parent | `ListNfsExportPolicyRules(ctx, policyName)` |
| No args beyond `ctx` | Global flat set | `ListSubnets(ctx)` |

Reference: `internal/client/targets.go`.

## Mock Handlers

One file per resource. Store struct: `sync.Mutex` + `byName map[string]*client.Xxx` + `nextID int`. `RegisterXxxHandlers(mux)` returns `*xxxStore` (for `Seed()`). Path includes API version: `/api/2.22/...`.

**Critical GET rule**: `?names=` matches nothing → **return empty list with HTTP 200**, NOT 404. This matches real API and lets `getOneByName[T]` detect not-found.

| Method | Success | Name exists | Name missing | Body invalid |
|--------|---------|-------------|--------------|--------------|
| POST | 200 + item | 409 Conflict | — | 400 Bad Request |
| PATCH | 200 + item | Apply non-nil fields | 404 Not Found | 400 Bad Request |
| DELETE | 200 (empty) | Remove from store | 404 Not Found | — |

**Shared helpers** (`handlers/helpers.go`): `ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`, `WriteJSONError`.

Reference: `internal/testmock/handlers/targets.go`.

## Resource Implementation

One file per resource. **All 4 interfaces mandatory**: `Resource`, `ResourceWithConfigure`, `ResourceWithImportState`, `ResourceWithUpgradeState`. Schema `Version` starts at 0.

### Plan modifiers

| Field type | Modifier |
|------------|----------|
| `id`, `created` (computed, stable) | `UseStateForUnknown()` |
| `name` (required, immutable) | `RequiresReplace()` |
| volatile (`status`, `lag`, `recovery_point`, etc.) | **NONE** — masks drift |

### Timeouts

All 4 operations enabled. Defaults: Create 20m, Read 5m, Update 20m, Delete 30m.

### Drift detection — mandatory

Every mutable/computed field: `tflog.Debug(ctx, "drift detected", map[string]any{"resource": name, "field": "xxx", "was": old, "now": new})`. Log only, never error.

### ImportState

Import by **name** (not UUID). Always call `nullTimeoutsValue()`. Set write-once/sensitive fields to null.

### Soft-delete (buckets, filesystems only)

Two-phase: PATCH `destroyed=true` → `pollUntilGone[T]`. Only for resources with `destroy_eradicate_on_delete`.

Reference: `internal/provider/target_resource.go`, `internal/provider/bucket_resource.go` (soft-delete).

## Data Source Implementation

One file per data source. **2 interfaces only**: `DataSource`, `DataSourceWithConfigure`. No timeouts, no plan modifiers. `name` = Required, all others = Computed. Not-found → `AddError` (not `RemoveResource`). Inline field mapping is fine for simple schemas.

Reference: `internal/provider/target_data_source.go`.

## State Upgraders

Bump `SchemaVersion` when adding/changing/renaming attributes. Naming: `xxxV0Model`, `xxxV1Model`; current = `xxxModel` (no suffix). `PriorSchema` must be exact copy of that version's schema. New fields → `types.StringNull()` / `types.ListNull(...)`. Chain runs sequentially (0→1→2). Entry key = prior version number. Empty map when version is 0.

Reference: `internal/provider/server_resource.go` (v0→v1→v2).

## Test Conventions

### Naming — mandatory

`TestUnit_<Resource>_<Operation>[_<Variant>]` — e.g. `TestUnit_Target_Get_Found`, `TestUnit_TargetResource_Lifecycle`, `TestUnit_TargetDataSource_Basic`.

**Exception — bridge tests:** tests under `pulumi/provider/` use the `TestProviderInfo_*` pattern because they test the Pulumi bridge ProviderInfo configuration, not individual TF resource/data source logic. This naming is intentional and approved.

### Client tests

`httptest.NewServer()`, mock `/api/login` for `x-auth-token`, use `newTestClient(t, srv)`. One test per CRUD + edge cases.

### Provider tests

5 required helpers per resource: `newTestXxxResource`, `xxxResourceSchema`, `buildXxxType`, `nullXxxConfig`, `xxxPlanWith`. DS tests: same pattern with `datasource.*` types + 4 helpers (`newTestXxxDataSource`, `xxxDSSchema`, `buildXxxDSType`, `nullXxxDSConfig`).

### Assertions

`t.Fatalf` for setup failures, `t.Errorf` for assertion failures.

## Test Coverage

| Component | Minimum tests |
|-----------|---------------|
| Client CRUD | 4 (Get_Found, Get_NotFound, Post, Patch, Delete) |
| Resource | 3 (Lifecycle, Import, DriftDetection) |
| Data source | 1 (Basic) |
| State upgrader | 1 per version bump |

**Baseline: 818 tests.** Must not decrease. New resource adds ≥8 tests. Run `make test` + `make lint` before every commit.

## Provider Registration

Register in `internal/provider/provider.go`: append `NewXxxResource` to `Resources()`, `NewXxxDataSource` to `DataSources()`.

## Documentation

`make docs` after any schema change (never edit `docs/` manually). Every resource needs `resource.tf` + `import.sh` examples. Every data source needs `data-source.tf`. Import by **name** (not UUID). Update `ROADMAP.md` in same commit as new resource.

## Checklist — New Resource

1. [ ] Model structs (Get/Post/Patch) in `models_<domain>.go`
2. [ ] Client CRUD using `getOneByName[T]`
3. [ ] Mock handler with Seed, empty-list GET, shared helpers
4. [ ] Client tests (≥4) with `TestUnit_` prefix
5. [ ] Resource with all 4 interfaces, schema version 0, correct plan modifiers
6. [ ] Drift detection on all mutable/computed fields
7. [ ] ImportState with `nullTimeoutsValue()`
8. [ ] Resource tests (≥3): Lifecycle, Import, DriftDetection
9. [ ] Data source with Configure + Read
10. [ ] Data source test (≥1): Basic
11. [ ] Registration in `provider.go`
12. [ ] HCL examples (`resource.tf`, `import.sh`, `data-source.tf`)
13. [ ] `make docs` regenerated
14. [ ] `make test` passes, total count ≥ 818 baseline
15. [ ] `make lint` clean
16. [ ] ROADMAP.md updated

## Checklist — Modify Existing Resource

1. [ ] Schema version incremented
2. [ ] State upgrader with PriorSchema + intermediate model
3. [ ] State upgrader test (≥1) with `TestUnit_` prefix
4. [ ] New fields → null in upgrader
5. [ ] ImportState updated for new fields
6. [ ] Drift detection for new mutable/computed fields
7. [ ] Plan modifiers correct (UseStateForUnknown only on stable computed)
8. [ ] `make test` passes, count ≥ previous baseline
9. [ ] `make lint` clean
10. [ ] `make docs` regenerated if schema changed
