---
name: flashblade-resource-builder
description: "Guide for implementing new Terraform resources and data sources for the Pure Storage FlashBlade provider. Covers the full lifecycle from API discovery to tested, documented resources. Enforces strict code quality conventions including generic helper usage, test naming, plan modifier rules, and drift detection patterns. This skill should be used when adding a new FlashBlade API resource, modifying an existing resource, or planning provider work."
---

# FlashBlade Resource Builder

## Purpose

Implement Terraform resources against the FlashBlade REST API v2.22 with the
provider's strict quality standards: full CRUD, import, drift detection,
comprehensive tests, and zero lint issues.

## When to Use

- Adding a new Terraform resource or data source
- Modifying an existing resource (new attributes, schema upgrades)
- Planning a new milestone or phase for the provider
- Reviewing code for convention compliance

## Mandatory First Steps

### 1. Read the Conventions

Before writing any code, read the full conventions document. This is the
single source of truth — every rule is non-negotiable:

```
Read: references/CONVENTIONS.md
```

Also read `CONVENTIONS.md` at the repo root (same content, kept in sync).
Key sections: File Structure, Model Structs, Plan Modifiers, Test Naming,
Coverage Baseline, Checklists.

### 2. Read the API Reference

Search `FLASHBLADE_API.md` at the repo root for the relevant endpoint:

```bash
grep -n "section_name\|endpoint_path" FLASHBLADE_API.md
```

Understand: GET/POST/PATCH/DELETE paths, query parameters, request body
fields (required vs optional vs read-only), response model fields and types.

### 3. Plan with GSD

All new work goes through the GSD planning workflow. Never implement
resources without a plan.

**New milestone (multiple related resources):**
```
/gsd:new-milestone
```

**New phase within current milestone:**
```
/gsd:add-phase
/gsd:plan-phase <N>
```

**Execute planned work:**
```
/gsd:execute-phase <N>
```

The plan ensures requirement coverage, test minimums, and dependency ordering.

## Architecture

```
internal/
  client/           # HTTP client, models, CRUD methods
    client.go       # Base client, auth, generics (getOneByName, postOne, patchOne)
    models_*.go     # Domain-grouped model structs
    <resource>.go   # One file per resource CRUD
  provider/         # Terraform resources, data sources, validators
    <resource>_resource.go
    <resource>_data_source.go
    provider.go     # Registration organized by domain
  testmock/
    handlers/       # One mock handler file per API resource
```

## Generic Helpers — Mandatory

The client layer provides three generics. **Always use them.** Hand-rolling
the GET/POST/PATCH + ListResponse unwrap pattern is a convention violation.

### getOneByName[T]

For any GET returning a single item:

```go
func (c *FlashBladeClient) GetXxx(ctx context.Context, name string) (*Xxx, error) {
    return getOneByName[Xxx](c, ctx, "/endpoint?names="+url.QueryEscape(name), "label", name)
}
```

Works with any query params — path is built before the call.

### postOne[TBody, TResp]

For any POST returning a ListResponse:

```go
func (c *FlashBladeClient) PostXxx(ctx context.Context, name string, body XxxPost) (*Xxx, error) {
    return postOne[XxxPost, Xxx](c, ctx, "/endpoint?names="+url.QueryEscape(name), body, "PostXxx")
}
```

Custom path logic stays in the wrapper — only the HTTP call + unwrap uses the generic.

### patchOne[TBody, TResp]

Identical pattern for PATCH:

```go
func (c *FlashBladeClient) PatchXxx(ctx context.Context, name string, body XxxPatch) (*Xxx, error) {
    return patchOne[XxxPatch, Xxx](c, ctx, "/endpoint?names="+url.QueryEscape(name), body, "PatchXxx")
}
```

### Delete — No Generic

Delete is already 1-2 lines (`c.delete(ctx, path)`). No generic needed.

## Implementation Steps

### Step 1: Model Structs

Append to `internal/client/models_<domain>.go`. Three structs per resource:

| Struct | Purpose | Pointer Rules |
|--------|---------|---------------|
| `Xxx` | GET response | No pointers. `NamedReference` for refs. |
| `XxxPost` | POST body | No pointers. Name via query param. |
| `XxxPatch` | PATCH body | ALL pointers (`*string`, `**NamedReference`). |

### Step 2: Client CRUD

Create `internal/client/<resource>.go` using the generic helpers. No API
version prefix in paths — `client.do()` adds it. Always propagate caller `ctx`.

### Step 3: Mock Handler

Create `internal/testmock/handlers/<resource>.go`:

- Thread-safe store: `sync.Mutex` + `byName` map + `nextID`
- `RegisterXxxHandlers(mux)` returns `*xxxStore` for test Seed
- **CRITICAL**: GET returns **empty list HTTP 200** when not found — NOT 404
- Use shared helpers: `ValidateQueryParams`, `RequireQueryParam`, `WriteJSONListResponse`, `WriteJSONError`

### Step 4: Client Tests

Create `internal/client/<resource>_test.go`. Minimum 4 tests:

```
TestUnit_Xxx_Get_Found
TestUnit_Xxx_Get_NotFound
TestUnit_Xxx_Post
TestUnit_Xxx_Patch_<Field>
TestUnit_Xxx_Delete
```

### Step 5: Terraform Resource

Create `internal/provider/<resource>_resource.go`.

**4 interface assertions mandatory.** Schema version starts at 0.

**Plan modifier rules (critical — violations mask drift):**

| Field type | Modifier |
|------------|----------|
| `id`, `created` (stable computed) | `UseStateForUnknown()` |
| `name` (immutable) | `RequiresReplace()` |
| `status`, `lag`, volatile fields | **NONE** |

**Drift detection in Read — mandatory for every mutable/computed field:**
```go
tflog.Debug(ctx, "drift detected", map[string]any{
    "resource": name, "field": "address",
    "was": data.Address.ValueString(), "now": apiObj.Address,
})
```

**ImportState:** Import by name, call `nullTimeoutsValue()`, set sensitive
fields to null or empty string.

**Sensitive fields (private_key, passphrase, etc.):**
- `Sensitive: true` in schema
- Preserve from plan in Create, from state in Read
- In Update: only send in PATCH if content changed (`.Equal()` comparison)
- If only sensitive fields changed: skip PATCH, copy computed fields from state

**Optional+Computed fields with API defaults:**
- Add `Computed: true` + custom plan modifier (unknown on Create, state on Update)
- Map empty API values to actual empty values (empty list, empty string) — NOT null
- Null in state + plan modifier = perpetual drift

### Step 6: Data Source

Create `internal/provider/<resource>_data_source.go`:

- 2 interface assertions only (no Import, no UpgradeState)
- No timeouts, no plan modifiers
- `name`: Required. Others: Computed.
- Not-found: `AddError`, NOT `RemoveResource`

### Step 7: Tests

**Resource (minimum 3):** Lifecycle, Import, DriftDetection

**Data source (minimum 1):** Basic

**State upgrader (if version bumped, minimum 1):** V0toV1

### Step 8: Registration & Documentation

1. Register in `provider.go` — correct domain group in `Resources()` and `DataSources()`
2. HCL examples: `resource.tf`, `import.sh`, `data-source.tf`
3. `make docs` — never edit `docs/` manually
4. Update `ROADMAP.md` in the same commit

## Validation

Before considering ANY resource complete:

```bash
make test    # All tests pass, count >= baseline (722+)
make lint    # 0 issues
make docs    # Regenerate
```

Run the full checklist from `references/CONVENTIONS.md`.

## Common Pitfalls

| Pitfall | Fix |
|---------|-----|
| `UseStateForUnknown` on volatile field | Remove the modifier |
| Mock GET returns 404 for not-found | Return empty list with HTTP 200 |
| Tests without `TestUnit_` prefix | Always prefix |
| Sending write-only fields on every PATCH | Compare plan vs state with `.Equal()` |
| Mapping empty API list to `types.ListNull()` | Map to `types.ListValueMust(...)` |
| Import by composite key (ambiguous) | Use UUID or add disambiguating filter |
| Hand-rolling POST/PATCH response unwrap | Use `postOne[T,R]` / `patchOne[T,R]` generics |

## Quality Gates

Every change must pass all four:

1. **CONVENTIONS.md** — coding standards
2. **Desloppify** — automated code quality (strict target: 85+)
3. **Test baseline** — count must never decrease
4. **Lint** — zero tolerance
