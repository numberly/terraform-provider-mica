# Phase 6: Server Resource & Export Consolidation - Research

**Researched:** 2026-03-28
**Domain:** Terraform Provider Go — FlashBlade Server CRUD + Mock Handler TDD Backfill
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SRV-01 | Operator can create a FlashBlade server with DNS configuration via Terraform | `POST /servers?create_ds=<name>` with `dns` array body; new client methods + resource CRUD |
| SRV-02 | Operator can update server DNS configuration via Terraform apply | `PATCH /servers?names=<name>` with `dns` array body; resource Update method |
| SRV-03 | Operator can destroy a server with cascade delete option for dependent exports | `DELETE /servers?names=<name>&cascade_delete=<names>` — optional param; resource Delete |
| SRV-04 | Operator can import an existing server into Terraform state | `ImportState` method using `GetServer`; must expose all attributes incl. `dns` |
| SRV-05 | Server data source reads existing server by name (consolidate existing) | Existing `serverDataSource` + `GetServer` already work; needs `dns` attribute added |
| EXP-01 | File system export resource has proper TDD unit tests and mock handlers | Mock handler exists (GET only); needs POST/PATCH/DELETE + full test suite |
| EXP-02 | Account export resource has proper TDD unit tests and mock handlers | No mock handler exists at all; needs full handler + test suite |
</phase_requirements>

---

## Summary

Phase 6 covers two distinct work areas. The first is building a full `flashblade_server` Terraform resource (CRUD + import) on top of the existing GET-only client and data source. The second is backfilling TDD coverage for two resources (`flashblade_file_system_export` and `flashblade_object_store_account_export`) that were written during v1.0 without mock handlers or unit tests.

The `Server` model in `models.go` is currently minimal — it holds only `name` and `id`. The API spec confirms servers have a `dns` array field (the only writable runtime attribute) and a `created` timestamp (read-only). The `Server` model must be extended before either the resource or the data source can expose DNS configuration.

The POST endpoint for servers is unconventional: the server name comes from the `?create_ds=<name>` query parameter (not `?names=`), and the body contains only `dns`. The DELETE endpoint accepts `?cascade_delete=<array>` — this is an array of export names to delete alongside the server, not a simple boolean flag. Both quirks must be reflected accurately in the mock handler and the client wrapper.

The export mock handlers are nearly complete for file system exports (GET exists) but entirely absent for account exports. The TDD backfill follows the identical pattern used for all 19 v1.0 resources: `handlers/<resource>.go` with full CRUD + `provider/<resource>_resource_test.go` covering Create/Read/Update/Delete/Import/NotFound.

**Primary recommendation:** Execute in two sequential plans — (1) extend `Server` model + full server client/handler/resource/data-source, then (2) add mock handlers and unit tests for both export resources.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `terraform-plugin-framework` | v1.19.0 | Resource/DataSource interfaces | Already in use across all 22 resources |
| `terraform-plugin-framework-timeouts` | v0.7.0 | Timeout attributes on resources | All mutable resources use it |
| `net/http` (stdlib) | — | Mock handler implementation | Used by all `handlers/*.go` |
| `encoding/json` (stdlib) | — | JSON decode in mock handlers | Same pattern in every handler |
| `github.com/google/uuid` | v1.6.0 | ID generation in mock handlers | Used by all existing handlers |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `sync.Mutex` | stdlib | Thread-safe store in mock handlers | All handlers wrap their map in a mutex |
| `tfsdk.State` / `tfsdk.Plan` | part of plugin-framework | Unit test scaffolding | Used by all existing `_test.go` files |
| `tftypes.Object` | plugin-go v0.31.0 | Type descriptor for test helpers | Required by `tfsdk.State.Raw` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `net/http` mock handlers | `httptest.NewServer` fixture | Project uses `testmock.MockServer` wrapping `httptest` — do not bypass |
| Inline test assertions | `testify/assert` | Project uses stdlib `testing.T` — do not add new dependencies |

**Installation:** No new dependencies required.

---

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── client/
│   ├── models.go                  # extend Server struct (add DNS, Created)
│   └── servers.go                 # add PostServer, PatchServer, DeleteServer
├── testmock/handlers/
│   ├── servers.go                 # extend: add POST, PATCH, DELETE handlers
│   ├── file_system_exports.go     # new: POST, PATCH, DELETE handlers (GET exists)
│   └── object_store_account_exports.go  # new: full CRUD handlers
└── provider/
    ├── server_data_source.go      # update: expose dns, created attributes
    ├── server_data_source_test.go # update: test dns attribute populated
    ├── server_resource.go         # new: full CRUD + import resource
    ├── server_resource_test.go    # new: Create/Read/Update/Delete/Import tests
    ├── file_system_export_resource_test.go  # new: full TDD test suite
    └── object_store_account_export_resource_test.go  # new: full TDD test suite
```

### Pattern 1: Server POST with `create_ds` query parameter
**What:** The FlashBlade API uses `?create_ds=<name>` (not `?names=`) to specify the server name on POST. This is documented in `FLASHBLADE_API.md` line 781.
**When to use:** `PostServer` client method only.
**Example:**
```go
// internal/client/servers.go
func (c *FlashBladeClient) PostServer(ctx context.Context, name string, body ServerPost) (*Server, error) {
    path := "/servers?create_ds=" + url.QueryEscape(name)
    var resp ListResponse[Server]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostServer: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### Pattern 2: Server DELETE with optional `cascade_delete`
**What:** `?cascade_delete=<array>` is an array of export names. In Terraform, this is modeled as an optional `List(String)` attribute. When populated, the names are joined as comma-separated values in the query param.
**When to use:** Resource Delete method when `cascade_delete` is non-empty.
**Example:**
```go
// internal/client/servers.go
func (c *FlashBladeClient) DeleteServer(ctx context.Context, name string, cascadeDelete []string) error {
    path := "/servers?names=" + url.QueryEscape(name)
    if len(cascadeDelete) > 0 {
        path += "&cascade_delete=" + url.QueryEscape(strings.Join(cascadeDelete, ","))
    }
    return c.delete(ctx, path)
}
```

### Pattern 3: Server model extension — `dns` is a list of objects
**What:** The API `Server` model has `dns`(array) — a list of DNS configuration objects. The current `Server` struct in `models.go` has only `Name` and `ID`. This must be extended with a `DNS` field (`[]ServerDNS`) and a `Created` field.
**When to use:** When defining `ServerPost` and `ServerPatch` models.

Key fields from API spec (line 1261):
- `created` (ro integer) — creation timestamp
- `dns` (array) — writable DNS config
- `directory_services` (ro array) — read-only, do not expose in resource
- `realms` (ro array) — read-only, do not expose in resource

**Example:**
```go
// internal/client/models.go — replace minimal Server struct
type ServerDNS struct {
    Domain      string   `json:"domain,omitempty"`
    Nameservers []string `json:"nameservers,omitempty"`
    Services    []string `json:"services,omitempty"`
}

type Server struct {
    Name    string      `json:"name"`
    ID      string      `json:"id"`
    Created int64       `json:"created,omitempty"`
    DNS     []ServerDNS `json:"dns,omitempty"`
}

type ServerPost struct {
    DNS []ServerDNS `json:"dns,omitempty"`
}

type ServerPatch struct {
    DNS []ServerDNS `json:"dns,omitempty"`
}
```

### Pattern 4: Mock handler for POST with `create_ds`
**What:** The mock handler for servers must dispatch on `create_ds` param (not `names`) for POST.
**When to use:** Extending `internal/testmock/handlers/servers.go`.
**Example:**
```go
// handlers/servers.go — handlePost
func (s *serverStore) handlePost(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("create_ds")
    if name == "" {
        WriteJSONError(w, http.StatusBadRequest, "create_ds query parameter is required for POST")
        return
    }
    // ... decode body, create server, store, respond
}
```

### Pattern 5: Mock handler for DELETE with `cascade_delete`
**What:** The `cascade_delete` param lists exports to delete together with the server. In the mock, simply delete those exports from the export store if provided.
**When to use:** `serverStore.handleDelete` — requires a cross-reference to the file system export store and/or account export store if cascade is tested.
**Example:**
```go
// handlers/servers.go — handleDelete
func (s *serverStore) handleDelete(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("names")
    cascadeRaw := r.URL.Query().Get("cascade_delete")
    // cascade_delete handling: if tests don't need to verify cascade, simply ignore it
    // in the mock and verify via real API in acceptance tests (Phase 8)
    _ = cascadeRaw
    // ... delete server from store
}
```

### Pattern 6: Export mock handler — file system exports
**What:** `internal/testmock/handlers/file_system_exports.go` needs to be created (or the existing GET-only file extended) with POST/PATCH/DELETE. The POST uses `?member_names=<fs_name>&policy_names=<policy_name>`. PATCH uses `?ids=<id>`. DELETE uses `?member_names=<fs_name>&names=<export_name>`.

### Pattern 7: Resource schema for `dns` as a ListNestedAttribute
**What:** The `dns` field is a list of objects. Use `schema.ListNestedAttribute` with nested attributes for `domain`, `nameservers` (List of String), `services` (List of String).
**When to use:** `server_resource.go` and updated `server_data_source.go`.

### Pattern 8: Server data source update
**What:** The existing `serverDataSource` exposes only `id` and `name`. It must be updated to also expose `dns` (computed) and `created` (computed). The `serverDataSourceModel` must gain corresponding fields.

### Anti-Patterns to Avoid
- **Using `?names=` for server POST:** The API uses `?create_ds=`, not `?names=`. Using the wrong param causes a 400.
- **Treating `cascade_delete` as a boolean:** It is an array of names, not a toggle flag. Model it as `types.List` in the resource schema.
- **Ignoring `dns` in import:** ImportState must call GetServer and map the full `dns` array, otherwise subsequent `plan` will show drift.
- **Adding `directory_services` or `realms` as writable:** These are read-only in the API. Mark as Computed only, or omit.
- **Skipping `UseStateForUnknown` on computed ID fields:** All computed ID-like fields need `stringplanmodifier.UseStateForUnknown()` to avoid plan churn.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON list response wrapper | Custom deserialization | `ListResponse[T]` generic already in `models.go` | Reused by all 22 existing client methods |
| Thread-safe in-memory store | Custom sync logic | `sync.Mutex` + `byName`/`byID` maps (existing pattern) | Identical in all 13 existing handlers |
| Mock server HTTP setup | Custom `httptest` scaffolding | `testmock.NewMockServer()` + `RegisterXxxHandlers(ms.Mux)` | Established in Phase 1, used by all tests |
| Timeout handling | Manual context deadline | `timeouts.Value` + `data.Timeouts.Create(ctx, 20*time.Minute)` | Used by all mutable resources |
| Not-found detection | Inline status code check | `client.IsNotFound(err)` | Standard pattern, prevents double-error |

---

## Common Pitfalls

### Pitfall 1: Server POST param is `create_ds`, not `names`
**What goes wrong:** Using `?names=<name>` on POST gives HTTP 400 or creates a nameless server.
**Why it happens:** FlashBlade uses different param names for POST vs GET/PATCH/DELETE on `/servers`.
**How to avoid:** Use `?create_ds=` in `PostServer` and the mock `handlePost`. Verified from `FLASHBLADE_API.md` line 781.
**Warning signs:** Mock handler returns 400 with "names query parameter is required".

### Pitfall 2: `cascade_delete` is an array, not a boolean
**What goes wrong:** Sending `?cascade_delete=true` is silently ignored or returns 400.
**Why it happens:** FlashBlade uses the param to list the specific export names to cascade-delete.
**How to avoid:** Model as `types.List` in the Terraform schema; join with comma in the client call.
**Warning signs:** Server deletes but dependent exports remain, causing orphaned state.

### Pitfall 3: Import drift on `dns` field
**What goes wrong:** Import succeeds but `terraform plan` shows a diff on `dns` because the list was not mapped from the API response.
**Why it happens:** `ImportState` calls `GetServer` but if `DNS` is not in the model, it stays null.
**How to avoid:** Ensure `mapServerToModel` maps the full `dns` list. Test with an `ImportState` unit test that verifies dns is populated.
**Warning signs:** "Plan shows diff after import" acceptance criterion fails.

### Pitfall 4: Missing `UseStateForUnknown` on computed fields causes plan churn
**What goes wrong:** `id` and `name` show `(known after apply)` on every plan after creation.
**Why it happens:** Missing `stringplanmodifier.UseStateForUnknown()` on Computed string attributes.
**How to avoid:** Apply to `id`, `name`, and any other Computed-only string in the resource schema.
**Warning signs:** `terraform plan` after `apply` shows 0 changes logically but still marks fields as unknown.

### Pitfall 5: `objectStoreAccountExport` Delete uses `data.Name` not `data.AccountName` as member
**What goes wrong:** The existing `Delete` implementation on line 263 of `object_store_account_export_resource.go` passes `data.Name` (the combined `account/export_name`) as `exportName`, not the short export name.
**Why it happens:** The combined name is used where only the short part was needed.
**How to avoid:** When writing tests, verify DELETE sends correct `member_names` and `names` query params. This may surface an existing bug.
**Warning signs:** DELETE returns 404 from mock handler because the param mismatch.

### Pitfall 6: File system export mock handler needs cross-reference to policy store
**What goes wrong:** A real `PostFileSystemExport` call passes `?policy_names=<policy>`. If the mock validates policy existence, tests fail unless the policy store is also seeded.
**Why it happens:** Over-strict mock validation.
**How to avoid:** In the mock, do NOT validate that the policy exists — just record `policy_names` as a string. Keep the mock simple (same as all other handlers).

---

## Code Examples

Verified patterns from existing codebase:

### Registering a handler with cross-reference (bucket pattern)
```go
// Source: internal/testmock/handlers/buckets.go (RegisterBucketHandlers signature)
func RegisterBucketHandlers(mux *http.ServeMux, accountStore *objectStoreAccountStore) *bucketStore {
    // cross-reference used for account validation on bucket creation
}
// Used in tests as:
// accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
// handlers.RegisterBucketHandlers(ms.Mux, accountStore)
```

### Resource unit test scaffold (generic pattern)
```go
// Source: internal/provider/object_store_account_resource_test.go
ms := testmock.NewMockServer()
defer ms.Close()
handlers.RegisterXxxHandlers(ms.Mux)
r := newTestXxxResource(t, ms)
s := xxxResourceSchema(t).Schema
plan := xxxPlanWith(t, ...)
resp := &resource.CreateResponse{
    State: tfsdk.State{Raw: tftypes.NewValue(buildXxxType(), nil), Schema: s},
}
r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)
```

### ImportState pattern
```go
// Source: internal/provider/file_system_export_resource.go ImportState
func (r *fileSystemExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    name := req.ID
    export, err := r.client.GetFileSystemExport(ctx, name)
    // ...
    var data fileSystemExportModel
    data.Timeouts = timeouts.Value{
        Object: types.ObjectNull(map[string]attr.Type{
            "create": types.StringType,
            "read":   types.StringType,
            "update": types.StringType,
            "delete": types.StringType,
        }),
    }
    mapFileSystemExportToModel(export, &data)
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### List attribute type for `dns` in unit tests
```go
// Pattern: tftypes.List{ElementType: <nested object type>}
dnsElemType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
    "domain":      tftypes.String,
    "nameservers": tftypes.List{ElementType: tftypes.String},
    "services":    tftypes.List{ElementType: tftypes.String},
}}
dnsType := tftypes.List{ElementType: dnsElemType}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| GET-only server handler + data source | Full CRUD handler + resource + enriched data source | Phase 6 | Operators can manage server lifecycle |
| Export resources without tests | Export resources with mock handlers + unit test suite | Phase 6 | TDD compliance, CI-safe |
| `Server{Name, ID}` only | `Server{Name, ID, Created, DNS}` | Phase 6 | Exposes DNS config in data source and resource |

**Deprecated/outdated:**
- `internal/testmock/handlers/servers.go` GET-only handler: will be replaced by full CRUD handler in this phase.
- `serverDataSourceModel{ID, Name}`: will gain `DNS` and `Created` fields.

---

## Open Questions

1. **DNS struct shape**
   - What we know: API has `dns`(array) on Server; `FLASHBLADE_API.md` does not detail the inner DNS object fields beyond the summary on line 1261
   - What's unclear: Exact subfields of each DNS entry (domain, nameservers, services — or more?)
   - Recommendation: Mirror the `ArrayDns` model already in `models.go` (it has `domain`, `nameservers`, `services`, `sources`). The server DNS config likely has the same shape. If the real API differs, the mock will still work and acceptance tests in Phase 8 will surface gaps.

2. **`cascade_delete` semantics — verified vs assumed**
   - What we know: API spec says `cascade_delete`(array*) on DELETE `/servers`
   - What's unclear: Whether the array contains export IDs or export names; whether it's comma-separated or multi-value
   - Recommendation: Model as `types.List(types.String)` with optional usage. Validate exact semantics in Phase 8 acceptance tests. For Phase 6 mock, accept and ignore the param.

3. **Account export Delete bug risk**
   - What we know: `object_store_account_export_resource.go` line 263 passes `data.Name` (combined name like `account/export`) as `exportName` arg, but `DeleteObjectStoreAccountExport` expects short export name
   - What's unclear: Whether this is intentional (the combined name works as the export name) or a latent bug
   - Recommendation: Write a unit test for Delete that verifies the mock receives correct query params. The test will confirm or deny the bug.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) — no external framework |
| Config file | None — `go test ./...` |
| Quick run command | `go test ./internal/provider/ -run TestUnit_Server -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SRV-01 | Create server with DNS config | unit | `go test ./internal/provider/ -run TestUnit_Server_Create -v` | ❌ Wave 0 |
| SRV-02 | Update server DNS via PATCH | unit | `go test ./internal/provider/ -run TestUnit_Server_Update -v` | ❌ Wave 0 |
| SRV-03 | Delete server (with cascade_delete) | unit | `go test ./internal/provider/ -run TestUnit_Server_Delete -v` | ❌ Wave 0 |
| SRV-04 | Import server, 0 drift on subsequent plan | unit | `go test ./internal/provider/ -run TestUnit_Server_Import -v` | ❌ Wave 0 |
| SRV-05 | Data source reads server + dns attributes | unit | `go test ./internal/provider/ -run TestUnit_ServerDataSource -v` | ✅ (partial — needs dns) |
| EXP-01 | FS export full CRUD + import tests | unit | `go test ./internal/provider/ -run TestUnit_FileSystemExport -v` | ❌ Wave 0 |
| EXP-02 | Account export full CRUD + import tests | unit | `go test ./internal/provider/ -run TestUnit_ObjectStoreAccountExport -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -run TestUnit_ -count=1`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/provider/server_resource_test.go` — covers SRV-01, SRV-02, SRV-03, SRV-04
- [ ] `internal/provider/file_system_export_resource_test.go` — covers EXP-01
- [ ] `internal/provider/object_store_account_export_resource_test.go` — covers EXP-02
- [ ] `internal/testmock/handlers/servers.go` — extend with POST/PATCH/DELETE (GET exists)
- [ ] `internal/testmock/handlers/file_system_exports.go` — new file with full CRUD mock handler
- [ ] `internal/testmock/handlers/object_store_account_exports.go` — new file with full CRUD mock handler
- [ ] `internal/client/models.go` — extend `Server` struct with `Created`, `DNS`; add `ServerPost`, `ServerPatch`, `ServerDNS`
- [ ] `internal/client/servers.go` — add `PostServer`, `PatchServer`, `DeleteServer`

---

## Sources

### Primary (HIGH confidence)
- `FLASHBLADE_API.md` lines 778-783 — `/servers` endpoint: GET/POST(`create_ds` param)/PATCH/DELETE(`cascade_delete` param)
- `FLASHBLADE_API.md` line 1261 — `Server` data model fields
- `internal/client/models.go` — current `Server`, `FileSystemExport`, `ObjectStoreAccountExport` structs
- `internal/client/servers.go` — existing GET-only client
- `internal/client/file_system_exports.go` — full CRUD client already exists
- `internal/client/object_store_account_exports.go` — full CRUD client already exists
- `internal/provider/server_data_source.go` — existing data source (GET-only, no dns)
- `internal/provider/file_system_export_resource.go` — full resource, no tests
- `internal/provider/object_store_account_export_resource.go` — full resource, no tests
- `internal/provider/server_data_source_test.go` — existing test (id + name only)
- `internal/testmock/handlers/servers.go` — GET-only mock handler
- `internal/testmock/handlers/object_store_accounts.go` — full CRUD handler reference pattern
- `internal/provider/object_store_account_resource_test.go` — full test pattern reference

### Secondary (MEDIUM confidence)
- `go.mod` — confirmed no new dependencies needed (all required libs at correct versions)
- `.planning/REQUIREMENTS.md` — requirement IDs and descriptions
- `.planning/STATE.md` — note about export resources needing TDD consolidation

### Tertiary (LOW confidence)
- DNS subfield shape: assumed to mirror `ArrayDns` struct — needs validation against live API in Phase 8

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries confirmed in `go.mod`, patterns confirmed in 22 existing resources
- Architecture: HIGH — patterns directly cloned from existing handlers and tests
- API quirks (create_ds, cascade_delete): HIGH — confirmed in `FLASHBLADE_API.md`
- DNS subfield shape: LOW — inferred from `ArrayDns` analogy, not directly documented
- Pitfalls: HIGH — identified from direct code inspection of existing implementations

**Research date:** 2026-03-28
**Valid until:** 2026-04-28 (stable Go provider codebase, no fast-moving dependencies)
