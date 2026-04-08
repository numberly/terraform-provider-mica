# CONVENTIONS.md — Terraform Provider FlashBlade

AI-oriented coding conventions. This document is the single source of truth
for how code MUST be written in this provider. Every new resource, data source,
or modification MUST follow these rules. No exceptions.

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

- **GET struct**: No pointers on scalars. Use `NamedReference` for refs, `*NamedReference` for optional refs.
- **POST struct**: No pointers. Name goes via query param, not body.
- **PATCH struct**: Every field is a pointer. `nil` = omit from JSON. Non-nil = send.
  - `*string` for scalar fields
  - `**NamedReference` for reference fields (outer nil = omit, outer non-nil + inner nil = set to null, outer non-nil + inner non-nil = set value)

### JSON tags

- Present field: `json:"field_name"`
- Optional/omittable: `json:"field_name,omitempty"`
- Name field excluded from body: `json:"-"` (name goes via query param)

Reference: `internal/client/models_storage.go` — Target structs.

## Client CRUD Methods

One file per resource: `internal/client/<resource>.go`

### Function signatures

```go
func (c *FlashBladeClient) GetXxx(ctx context.Context, name string) (*Xxx, error)
func (c *FlashBladeClient) PostXxx(ctx context.Context, name string, body XxxPost) (*Xxx, error)
func (c *FlashBladeClient) PatchXxx(ctx context.Context, name string, body XxxPatch) (*Xxx, error)
func (c *FlashBladeClient) DeleteXxx(ctx context.Context, name string) error
```

### Rules

- **GET**: Use `getOneByName[T]` generic helper — never hand-roll list+filter logic.
- **POST/PATCH/DELETE**: Name goes via `?names=` query param with `url.QueryEscape(name)`.
- **Endpoint path**: No API version prefix — `"/targets?names=..."` not `"/api/2.22/targets?names=..."`. Version is added by `client.do()`.
- **Error wrapping**: Use `fmt.Errorf("FuncName: %w", err)`.
- **Context**: Always propagate caller `ctx` — never use `context.Background()`.

Reference: `internal/client/remote_credentials.go`, `internal/client/targets.go`.

## Mock Handlers

One file per resource: `internal/testmock/handlers/<resource>.go`

### Store struct pattern

```go
type xxxStore struct {
    mu     sync.Mutex
    byName map[string]*client.Xxx
    nextID int
}
```

- Always `sync.Mutex` — mock server is used concurrently.
- Always `byName` map keyed by resource name.
- `nextID int` for generating synthetic IDs (`fmt.Sprintf("xxx-%d", s.nextID)`).

### Registration

```go
func RegisterXxxHandlers(mux *http.ServeMux) *xxxStore {
    store := &xxxStore{byName: make(map[string]*client.Xxx), nextID: 1}
    mux.HandleFunc("/api/2.22/<endpoint>", store.handle)
    return store
}
```

- Returns `*xxxStore` so tests can call `Seed()`.
- Path includes API version: `/api/2.22/...` (mock is hit directly, not through client URL builder).

### Seed method

```go
func (s *xxxStore) Seed(item *client.Xxx) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.byName[item.Name] = item
}
```

### GET handler — critical rule

When `?names=` filter matches nothing: **return empty list with HTTP 200**, NOT 404.

```go
if namesFilter != "" {
    item, ok := s.byName[namesFilter]
    if ok {
        items = append(items, *item)
    }
}
// falls through to WriteJSONListResponse(w, http.StatusOK, items)
```

This matches real FlashBlade API behavior and lets `getOneByName[T]` handle not-found detection.

### Other handlers

| Method | Success | Name exists | Name missing | Body invalid |
|--------|---------|-------------|--------------|--------------|
| POST | 200 + item | 409 Conflict | — | 400 Bad Request |
| PATCH | 200 + item | Apply non-nil fields | 404 Not Found | 400 Bad Request |
| DELETE | 200 (empty) | Remove from store | 404 Not Found | — |

### Shared helpers (from `handlers/helpers.go`)

- `ValidateQueryParams(w, r, []string{"allowed", "params"})` — reject unknown params
- `RequireQueryParam(w, r, "names")` — 400 if missing
- `WriteJSONListResponse(w, statusCode, items)` — wrap in `{"items": [...]}`
- `WriteJSONError(w, statusCode, "message")` — standard error envelope

Reference: `internal/testmock/handlers/remote_credentials.go`, `internal/testmock/handlers/targets.go`.

## Resource Implementation

One file per resource: `internal/provider/<resource>_resource.go`

### Interface assertions (top of file)

```go
var _ resource.Resource = &xxxResource{}
var _ resource.ResourceWithConfigure = &xxxResource{}
var _ resource.ResourceWithImportState = &xxxResource{}
var _ resource.ResourceWithUpgradeState = &xxxResource{}
```

All four interfaces are mandatory for every resource.

### Schema

- `Version`: Start at `0`. Increment when adding/changing/renaming attributes.
- `Description`: One sentence describing the resource.

### Plan modifiers — critical rules

| Field type | Modifier | Example |
|------------|----------|---------|
| `id` (computed, stable) | `UseStateForUnknown()` | Never changes after creation |
| `created` (computed, stable) | `UseStateForUnknown()` | Never changes after creation |
| `name` (required, immutable) | `RequiresReplace()` | Forces new resource on change |
| `status`, `status_details` (volatile) | **NONE** | Changes outside Terraform |
| `lag`, `recovery_point`, `backlog` (volatile) | **NONE** | Changes outside Terraform |

**DO**: `UseStateForUnknown()` on computed fields that never change after creation.
**DO NOT**: `UseStateForUnknown()` on any field that can change outside Terraform. This masks drift.

### Timeouts (standard values)

```go
"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
    Create: true, Read: true, Update: true, Delete: true,
})
```

| Operation | Default |
|-----------|---------|
| Create | `20 * time.Minute` |
| Read | `5 * time.Minute` |
| Update | `20 * time.Minute` |
| Delete | `30 * time.Minute` |

### Configure

```go
func (r *xxxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil { return }
    c, ok := req.ProviderData.(*client.FlashBladeClient)
    if !ok {
        resp.Diagnostics.AddError("Unexpected Provider Data Type", "...")
        return
    }
    r.client = c
}
```

### Drift detection in Read — mandatory

Every mutable or computed field MUST be checked and logged:

```go
if data.Address.ValueString() != apiObj.Address {
    tflog.Debug(ctx, "drift detected", map[string]any{
        "resource": name,
        "field":    "address",
        "was":      data.Address.ValueString(),
        "now":      apiObj.Address,
    })
}
```

- Use `tflog.Debug` (not Info, not Warn).
- Keys: `resource`, `field`, `was`, `now`.
- Log only — never error on drift.

### ImportState

```go
func (r *xxxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    name := req.ID
    obj, err := r.client.GetXxx(ctx, name)
    // handle error...
    var data xxxModel
    data.Timeouts = nullTimeoutsValue()   // always initialize timeouts
    mapXxxToModel(obj, &data)
    // set write-once fields to null/empty (SecretAccessKey, TargetName, etc.)
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

- Import by **name**, not UUID.
- Always call `nullTimeoutsValue()` — import has no plan to read timeouts from.
- Set write-once/sensitive fields to null or empty string.

### Soft-delete (buckets, filesystems only)

Two-phase destroy: PATCH `destroyed=true`, then optional eradication with `pollUntilGone[T]`.
Only applies to resources with the `destroy_eradicate_on_delete` attribute.

Reference: `internal/provider/remote_credentials_resource.go`, `internal/provider/target_resource.go`, `internal/provider/bucket_resource.go` (soft-delete).

## Data Source Implementation

One file per data source: `internal/provider/<resource>_data_source.go`

### Interface assertions

```go
var _ datasource.DataSource = &xxxDataSource{}
var _ datasource.DataSourceWithConfigure = &xxxDataSource{}
```

Only two interfaces — no ImportState, no UpgradeState.

### Schema differences from resource

- **No timeouts block.**
- **No plan modifiers** (no UseStateForUnknown, no RequiresReplace).
- `name`: Required (lookup key).
- All other fields: Computed.

### Read pattern

```go
func (d *xxxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var config xxxDataSourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

    name := config.Name.ValueString()
    obj, err := d.client.GetXxx(ctx, name)
    if err != nil {
        if client.IsNotFound(err) {
            resp.Diagnostics.AddError("Xxx not found", fmt.Sprintf("No xxx named %q", name))
        } else {
            resp.Diagnostics.AddError("Error reading xxx", err.Error())
        }
        return
    }
    // map fields inline — no helper needed for simple data sources
    config.ID = types.StringValue(obj.ID)
    // ...
    resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
```

- No timeout wrapping (uses context directly).
- Not-found produces `AddError`, not `RemoveResource` (data sources don't exist in state).
- Field mapping can be inline for simple schemas — no `mapXxxToModel` helper required.

Reference: `internal/provider/target_data_source.go`, `internal/provider/remote_credentials_data_source.go`.

## State Upgraders

### When to bump SchemaVersion

Increment `Version` in `Schema()` when:
- Adding a new attribute
- Changing an attribute type
- Renaming an attribute

### Pattern

1. **Define intermediate model struct** for the prior version:

```go
type xxxV0Model struct {
    ID       types.String   `tfsdk:"id"`
    Name     types.String   `tfsdk:"name"`
    // ... v0 fields only
    Timeouts timeouts.Value `tfsdk:"timeouts"`
}
```

- Naming: `xxxV0Model`, `xxxV1Model`, etc.
- Current version model has NO version suffix (just `xxxModel`).

2. **Add upgrader to UpgradeState chain:**

```go
func (r *xxxResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
    return map[int64]resource.StateUpgrader{
        0: {
            PriorSchema: &schema.Schema{
                // EXACT copy of v0 schema as it existed
                Attributes: map[string]schema.Attribute{...},
            },
            StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
                var old xxxV0Model
                resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
                if resp.Diagnostics.HasError() { return }

                newState := xxxModel{
                    // copy existing fields from old
                    ID:       old.ID,
                    Name:     old.Name,
                    // initialize new fields
                    NewField: types.StringNull(),
                    Timeouts: old.Timeouts,
                }
                resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
            },
        },
    }
}
```

### Rules

- **PriorSchema**: Must be an exact copy of the schema at that version. Include timeouts with `timeouts.Attributes(ctx, ...)`.
- **New fields**: Initialize to `types.StringNull()`, `types.ListNull(...)`, or appropriate zero value.
- **Chain**: Upgraders run sequentially (0→1, 1→2). Each entry key is the **prior** version number.
- **SchemaVersion** in `Schema()` must equal `len(UpgradeState) + initial version`.
- **Empty map** when version is 0: `return map[int64]resource.StateUpgrader{}`

Reference: `internal/provider/server_resource.go` (v0→v1→v2), `internal/provider/remote_credentials_resource.go` (v0→v1).

## Test Conventions

### Naming — mandatory prefix

```
TestUnit_<Resource>_<Operation>[_<Variant>]
```

Examples:
- `TestUnit_Target_Get_Found`
- `TestUnit_Target_Post`
- `TestUnit_Target_Patch_Address`
- `TestUnit_Target_Patch_CACertGroup`
- `TestUnit_Target_Delete`
- `TestUnit_TargetResource_Lifecycle`
- `TestUnit_TargetResource_Import`
- `TestUnit_TargetResource_DriftDetection`
- `TestUnit_TargetDataSource_Basic`
- `TestUnit_RemoteCredentials_StateUpgrade_V0toV1`

**DO NOT**: Use `TestGetTarget_found`, `TestTargetResource_lifecycle`, or any other convention.

### Client tests (`internal/client/<resource>_test.go`)

- Use `httptest.NewServer()` with inline handler OR mock handler package.
- Mock `/api/login` to return `x-auth-token` header.
- Use shared `newTestClient(t, srv)` helper.
- One test per CRUD operation + edge cases (not-found, conflict).

### Provider resource tests (`internal/provider/<resource>_resource_test.go`)

Required helpers per resource:

```go
func newTestXxxResource(t *testing.T, ms *testmock.MockServer) *xxxResource     // wired client
func xxxResourceSchema(t *testing.T) resource.SchemaResponse                     // parsed schema
func buildXxxType() tftypes.Object                                               // tftypes for raw values
func nullXxxConfig() map[string]tftypes.Value                                    // all-null baseline
func xxxPlanWith(t *testing.T, ...) tfsdk.Plan                                  // plan builder
```

Test setup:
```go
ms := testmock.NewMockServer()
defer ms.Close()
store := handlers.RegisterXxxHandlers(ms.Mux)
r := newTestXxxResource(t, ms)
s := xxxResourceSchema(t).Schema
```

### Provider data source tests (`internal/provider/<resource>_data_source_test.go`)

Same pattern with `datasource.*` types:

```go
func newTestXxxDataSource(t *testing.T, ms *testmock.MockServer) *xxxDataSource
func xxxDSSchema(t *testing.T) datasource.SchemaResponse
func buildXxxDSType() tftypes.Object
func nullXxxDSConfig() map[string]tftypes.Value
```

### Assertions

- Use `t.Fatalf` for setup failures that block the rest of the test.
- Use `t.Errorf` for assertion failures that should continue checking other fields.
- String comparison: `if got.Name != "expected" { t.Errorf("expected %q, got %q", "expected", got.Name) }`
- Not-found check: `if !client.IsNotFound(err) { t.Errorf(...) }`

## Test Coverage

### Minimum tests per component

| Component | Minimum | Tests |
|-----------|---------|-------|
| Client CRUD | 4 | Get_Found, Get_NotFound, Post, Patch (one variant), Delete |
| Resource | 3 | Lifecycle (create→read→update→delete), Import, DriftDetection |
| Data source | 1 | Basic (seed + read + verify all fields) |
| State upgrader | 1 per version bump | V0toV1, V1toV2, etc. |

### Coverage rules

- **Total test count MUST NOT decrease.** Current baseline: **735 tests**.
- Every new resource adds at minimum **8 tests** (4 client + 3 resource + 1 data source).
- Every state upgrader adds at minimum **1 test**.
- Run `make test` and `make lint` before every commit. Both must be clean.
- Update the baseline count in this document when adding tests.

## Provider Registration

When adding a new resource or data source, register it in `internal/provider/provider.go`:

- Append `NewXxxResource` to the `Resources()` method return slice.
- Append `NewXxxDataSource` to the `DataSources()` method return slice.

Factory functions are defined in the resource/data source files:
```go
func NewXxxResource() resource.Resource { return &xxxResource{} }
func NewXxxDataSource() datasource.DataSource { return &xxxDataSource{} }
```

## Documentation

### Generated docs

- Run `make docs` after any schema change. This invokes `tfplugindocs`.
- Output goes to `docs/resources/<resource>.md` and `docs/data-sources/<resource>.md`.
- **Never edit files in `docs/` manually.** They are overwritten by `make docs`.

### HCL examples

Every resource needs:
- `examples/resources/flashblade_<resource>/resource.tf` — minimal working config
- `examples/resources/flashblade_<resource>/import.sh` — import command

Every data source needs:
- `examples/data-sources/flashblade_<resource>/data-source.tf` — lookup example

Import uses the resource **name** (not UUID):
```bash
terraform import flashblade_target.example my-target
```

### ROADMAP.md

When implementing a new resource, update `ROADMAP.md` in the same commit:
1. Move entry from "Not Implemented" to "Implemented"
2. Update header counters
3. Update `Last updated` date

## Checklist — New Resource

Before considering a new resource complete, every item must be done:

1. [ ] Model structs (Get/Post/Patch) in `models_<domain>.go`
2. [ ] Client CRUD in `<resource>.go` using `getOneByName[T]`
3. [ ] Mock handler in `handlers/<resource>.go` with Seed, empty-list GET, shared helpers
4. [ ] Client tests (≥4) with `TestUnit_` prefix
5. [ ] Resource with all 4 interfaces, schema version 0, correct plan modifiers
6. [ ] Drift detection on all mutable/computed fields in Read
7. [ ] ImportState with `nullTimeoutsValue()`
8. [ ] Resource tests (≥3): Lifecycle, Import, DriftDetection
9. [ ] Data source with Configure + Read
10. [ ] Data source test (≥1): Basic
11. [ ] Registration in `provider.go` (Resources + DataSources)
12. [ ] HCL examples (`resource.tf`, `import.sh`, `data-source.tf`)
13. [ ] `make docs` regenerated
14. [ ] `make test` passes, total count ≥ 716 baseline
15. [ ] `make lint` clean (0 issues)
16. [ ] ROADMAP.md updated

## Checklist — Modify Existing Resource

1. [ ] Schema version incremented
2. [ ] State upgrader added with PriorSchema + intermediate model
3. [ ] State upgrader test (≥1) with `TestUnit_` prefix
4. [ ] New fields initialized to null in upgrader
5. [ ] ImportState updated for new fields
6. [ ] Drift detection added for new mutable/computed fields
7. [ ] Plan modifiers correct (UseStateForUnknown only on stable computed fields)
8. [ ] `make test` passes, total count ≥ previous baseline
9. [ ] `make lint` clean
10. [ ] `make docs` regenerated if schema changed
