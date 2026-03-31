# Phase 28: LAG Data Source & Subnet Resource - Research

**Researched:** 2026-03-30
**Domain:** Terraform provider (Go) — FlashBlade REST API v2.22 subnets and link-aggregation-groups
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| LAG-01 | Operator can read an existing LAG by name via data source (name, status, ports, port_speed, lag_speed, mac_address) | LinkAggregationGroup model confirmed in FLASHBLADE_API.md line 1075; GET-only endpoint; pure data source pattern |
| SUB-01 | Operator can create a subnet with name, prefix, gateway, mtu, vlan, and link_aggregation_group via Terraform | POST /api/2.22/subnets confirmed; name is `ro` → passed via ?names= query param (user-provided); all fields confirmed writable |
| SUB-02 | Operator can update subnet settings (gateway, prefix, mtu, vlan, link_aggregation_group) via Terraform apply | PATCH /api/2.22/subnets confirmed; same writable field set as POST |
| SUB-03 | Operator can delete a subnet via Terraform destroy | DELETE /api/2.22/subnets confirmed; standard ?names= pattern |
| SUB-04 | Operator can read any existing subnet by name via data source | GET /api/2.22/subnets confirmed; ?names= filter pattern |
| SUB-05 | Operator can import an existing subnet into Terraform state with no drift on subsequent plan | ImportState via name (user-provided name, unlike VIP auto-generated); name is the import ID |
| SUB-06 | Drift detection logs changes when subnet is modified outside Terraform | Standard Read→plan diff pattern; all writable fields in state |
</phase_requirements>

---

## Summary

Phase 28 adds two new Terraform types: `flashblade_link_aggregation_group` (data source only — read existing LAGs) and `flashblade_subnet` (full CRUD resource + data source — manage subnets referencing LAGs). Both are pure mechanical extensions of established provider patterns. No new dependencies are required.

The critical design difference from Phase 29 (VIP) is that **subnet name IS user-provided**, passed via the `?names=` query parameter at POST time — the same pattern as `alert-watchers` and `object-store-virtual-hosts`. This makes the schema simpler: `name` is `Required + RequiresReplace`, not `Computed`-only. The subnet resource has one nested reference field (`link_aggregation_group`) that bridges `types.String` (Terraform) to `*NamedReference` (API JSON), following the exact same pattern as `subnet` on the network interface.

The LAG data source is the simplest possible data source in the codebase: read-only, no POST/PATCH/DELETE, all fields computed. The reference pattern is `array_connection_data_source.go` (read-only object, flat attributes, single GET by name).

**Primary recommendation:** Build in dependency order — models → client methods → mock handlers → subnet resource → subnet data source → LAG data source → provider registration. Mock handlers for both endpoints are needed before provider-level integration tests. Subnet and LAG mock handlers are independent of each other and can be written in parallel.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `terraform-plugin-framework` | v1.19.0 | Resource/data source CRUD scaffold | Only supported SDK for new providers; SDKv2 maintenance-only |
| `stdlib net/http` | — | REST API calls via FlashBladeClient | Existing RoundTripper chain; no new deps needed |
| `net/http/httptest` | stdlib | Mock API server for integration tests | Zero external deps; established pattern in all 28 handlers |
| `terraform-plugin-testing` | v1.15.0 | Acceptance test runner | Existing test infrastructure |
| `github.com/google/uuid` | v1.x | UUID generation in mock handlers | Already imported in all existing handlers |

### Supporting

No new libraries required. All primitives (`NamedReference`, `ListResponse[T]`, `getOneByName`, `IsNotFound`) already exist in the client package.

### Alternatives Considered

None. The stack is established and not open for choice at this phase.

**Installation:** No new `go get` commands needed. All dependencies already in `go.mod`.

## Architecture Patterns

### Recommended Project Structure

```
internal/
├── client/
│   ├── models_network.go        # NEW — Subnet, SubnetPost, SubnetPatch, LinkAggregationGroup structs
│   ├── subnets.go               # NEW — GetSubnet, ListSubnets, PostSubnet, PatchSubnet, DeleteSubnet
│   ├── link_aggregation_groups.go  # NEW — GetLinkAggregationGroup, ListLinkAggregationGroups (GET-only)
│   └── [existing files unchanged]
├── provider/
│   ├── subnet_resource.go              # NEW — flashblade_subnet resource (full CRUD + import)
│   ├── subnet_data_source.go           # NEW — flashblade_subnet data source
│   ├── link_aggregation_group_data_source.go  # NEW — flashblade_link_aggregation_group data source
│   └── provider.go                     # MODIFIED — register new resource + 2 data sources
└── testmock/
    └── handlers/
        ├── subnets.go                  # NEW — in-memory CRUD mock for /api/2.22/subnets
        └── link_aggregation_groups.go  # NEW — in-memory read-only mock for /api/2.22/link-aggregation-groups
```

### Pattern 1: User-Provided Name via ?names= Query Param at POST

**What:** Subnet `name` is marked `ro` in the API model but is passed as the `?names=` query parameter at POST, not in the request body. The name is determined by the operator, not auto-assigned by the array. This is identical to `alert-watchers` and (partially) `object-store-virtual-hosts`.

**When to use:** `PostSubnet` and `DeleteSubnet` and `PatchSubnet` all use `?names=<name>`.

**Key implication:** `name` in the Terraform schema is `Required: true` with `stringplanmodifier.RequiresReplace()`. No `UseStateForUnknown()` needed. Import ID is the user-chosen name.

**Example:**
```go
// internal/client/subnets.go
func (c *FlashBladeClient) PostSubnet(ctx context.Context, name string, body SubnetPost) (*Subnet, error) {
    path := "/subnets?names=" + url.QueryEscape(name)
    var resp ListResponse[Subnet]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostSubnet: empty response from server")
    }
    return &resp.Items[0], nil
}

func (c *FlashBladeClient) GetSubnet(ctx context.Context, name string) (*Subnet, error) {
    return getOneByName[Subnet](c, ctx, "/subnets?names="+url.QueryEscape(name), "subnet", name)
}

func (c *FlashBladeClient) PatchSubnet(ctx context.Context, name string, body SubnetPatch) (*Subnet, error) {
    path := "/subnets?names=" + url.QueryEscape(name)
    var resp ListResponse[Subnet]
    if err := c.patch(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PatchSubnet: empty response from server")
    }
    return &resp.Items[0], nil
}

func (c *FlashBladeClient) DeleteSubnet(ctx context.Context, name string) error {
    return c.delete(ctx, "/subnets?names="+url.QueryEscape(name))
}
```

### Pattern 2: NamedReference Bridge for link_aggregation_group

**What:** The `link_aggregation_group` field is `*NamedReference` in the API JSON (`{"link_aggregation_group": {"name": "..."}}`) but exposed as a flat `types.String` (`lag_name`) in Terraform state. The expand/flatten functions bridge between these representations.

**When to use:** In both `SubnetPost` and `SubnetPatch` bodies. In the Terraform schema, `lag_name` is `Optional + Computed`. On Read, populate it from the response's `LinkAggregationGroup.Name`.

**Why flat string:** Consistent with all other resource reference patterns in this provider (`subnet_name` on network interfaces, `server` on file system exports). The LAG is always referenced by name; users never need the internal LAG ID.

**Example:**
```go
// Expand: types.String → *NamedReference for POST/PATCH body
func lagNameToRef(lagName types.String) *client.NamedReference {
    if lagName.IsNull() || lagName.IsUnknown() || lagName.ValueString() == "" {
        return nil
    }
    return &client.NamedReference{Name: lagName.ValueString()}
}

// Flatten: *NamedReference from GET response → types.String for state
func refToLagName(ref *client.NamedReference) types.String {
    if ref == nil || ref.Name == "" {
        return types.StringNull()
    }
    return types.StringValue(ref.Name)
}
```

### Pattern 3: Read-Only Data Source (LAG)

**What:** The LAG data source is identical in structure to `array_connection_data_source.go`. It has one `Required` field (`name`) used as lookup key, and all remaining fields are `Computed: true`. No POST/PATCH/DELETE client methods are needed.

**When to use:** For infrastructure objects that are pre-existing physical/logical constructs not managed by Terraform. LAGs are configured at the hardware level; Terraform operators only need to reference them by name.

**Example:**
```go
// internal/client/link_aggregation_groups.go — GET-only client
func (c *FlashBladeClient) GetLinkAggregationGroup(ctx context.Context, name string) (*LinkAggregationGroup, error) {
    return getOneByName[LinkAggregationGroup](c, ctx, "/link-aggregation-groups?names="+url.QueryEscape(name), "link aggregation group", name)
}

func (c *FlashBladeClient) ListLinkAggregationGroups(ctx context.Context) ([]LinkAggregationGroup, error) {
    var resp ListResponse[LinkAggregationGroup]
    if err := c.get(ctx, "/link-aggregation-groups", &resp); err != nil {
        return nil, err
    }
    return resp.Items, nil
}
```

### Pattern 4: Mock Handler for Read-Only Endpoint (LAG)

**What:** The LAG mock handler registers only GET. It exposes a `Seed` method for tests to insert LAGs. No POST/PATCH/DELETE handlers are needed, so the mock is simpler than resource mocks.

**Example:**
```go
// internal/testmock/handlers/link_aggregation_groups.go
type lagStore struct {
    mu   sync.Mutex
    lags map[string]*client.LinkAggregationGroup
}

func RegisterLinkAggregationGroupHandlers(mux *http.ServeMux) *lagStore {
    store := &lagStore{lags: make(map[string]*client.LinkAggregationGroup)}
    mux.HandleFunc("/api/2.22/link-aggregation-groups", store.handleGet)
    return store
}

func (s *lagStore) Seed(lag *client.LinkAggregationGroup) {
    s.mu.Lock(); defer s.mu.Unlock()
    s.lags[lag.Name] = lag
}
```

### Pattern 5: Subnet Mock Handler (Full CRUD with ?names= at POST)

**What:** Unlike the VIP mock (which auto-generates names with a counter), the subnet mock reads the `?names=` query parameter at POST — identical to the server mock's `?create_ds=` pattern but using `?names=`.

**Example:**
```go
func (s *subnetStore) handlePost(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("names")
    if name == "" {
        WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
        return
    }
    // decode body, create subnet with given name, store and return
}
```

### Recommended Build Order

| Step | File | Depends On | Parallel With |
|------|------|-----------|---------------|
| 1 | `internal/client/models_network.go` | nothing | — |
| 2 | `internal/client/subnets.go` | step 1 | step 3 |
| 3 | `internal/client/link_aggregation_groups.go` | step 1 | step 2 |
| 4 | `internal/testmock/handlers/subnets.go` | step 1 | step 5 |
| 5 | `internal/testmock/handlers/link_aggregation_groups.go` | step 1 | step 4 |
| 6 | `internal/provider/subnet_resource.go` | steps 2, 4 | step 7, 8 |
| 7 | `internal/provider/subnet_data_source.go` | steps 2, 4 | step 6, 8 |
| 8 | `internal/provider/link_aggregation_group_data_source.go` | steps 3, 5 | step 6, 7 |
| 9 | `internal/provider/provider.go` | steps 6, 7, 8 | — |

Steps 2+3 are parallel (both depend only on step 1). Steps 4+5 are parallel. Steps 6+7+8 are parallel after their respective mock handlers are ready.

### Anti-Patterns to Avoid

- **Including `name` in SubnetPost/SubnetPatch body:** Name is `ro` in the API model — it is passed exclusively via `?names=` query param. Including `Name` in the struct body will cause it to be sent as JSON and may be silently ignored or cause API errors.
- **Exposing `enabled`, `interfaces`, `services` as Optional on the resource:** These are `ro` in the API spec. They must be `Computed: true` only in the subnet resource schema.
- **Adding POST/PATCH/DELETE to LAG client:** The LAG endpoint only supports GET in the context of this provider (POST/PATCH/DELETE on LAGs manage hardware ports — not a Terraform concern). Only `GetLinkAggregationGroup` and `ListLinkAggregationGroups` should be implemented.
- **Not calling Read at the end of Create/Update:** State diverges from API reality. Every `Create` and `Update` must terminate with a `Read` call to populate all computed fields (e.g., `enabled`, `interfaces`, `services`) from the actual API response.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Fetch-by-name with 404 handling | Custom GET with manual error parsing | `getOneByName[T]` in `internal/client` | Already handles empty list → 404, consistent error type |
| List pagination | Manual token loop | `ListResponse[T]` + `get()` helper | Pagination handled transparently; consistent across all endpoints |
| Thread-safe mock state | Custom mutex-per-struct | `sync.Mutex` + map pattern (see `serverStore`) | Established pattern across all 28 mock handlers |
| JSON PATCH semantics | Full struct replacement | `map[string]json.RawMessage` raw decode in PATCH handler | Allows true field-level updates without zero-value clobbering |

**Key insight:** Every pattern needed for this phase already exists in the codebase. The risk of hand-rolling is introducing inconsistency with established error handling, pagination, and thread-safety patterns.

## Common Pitfalls

### Pitfall 1: Treating subnet name as Computed (auto-generated) like VIP name

**What goes wrong:** If `name` is `Computed: true` with `UseStateForUnknown()`, the plan will show `(known after apply)` and the mock POST handler will have no name to use — breaking the ID lifecycle.

**Why it happens:** Confusion with Phase 29 (VIP), where name IS auto-generated. For subnets, the API model marks `name` as `ro` in the body schema, but this means "do not send in body" — not "we generate it." The name is always user-chosen, passed as `?names=`.

**How to avoid:** Set `name` as `Required: true` with `stringplanmodifier.RequiresReplace()` in the subnet resource schema. Import ID is the subnet name.

**Warning signs:** Mock POST handler receiving empty `?names=` parameter.

### Pitfall 2: Sending read-only fields in POST/PATCH body

**What goes wrong:** Fields like `enabled`, `id`, `interfaces`, `services` appear in the FLASHBLADE_API.md POST/PATCH body documentation but are marked `ro`. Sending them causes either a 422 error or silent rejection.

**Why it happens:** FLASHBLADE_API.md line 797-798 lists the full object schema in the body column, not just writable fields. Fields must be individually checked for the `(ro)` annotation.

**How to avoid:** `SubnetPost` contains only: `gateway`, `link_aggregation_group`, `mtu`, `prefix`, `vlan`. `SubnetPatch` contains the same writable set (all five can be updated). Use pointer types in `SubnetPatch` for fields that should be omittable when unchanged.

**Writable fields confirmed from API spec:**
- POST writable: `gateway`(string), `link_aggregation_group`(object), `mtu`(integer), `prefix`(string), `vlan`(integer)
- POST read-only (do NOT include in SubnetPost): `enabled`, `id`, `interfaces`, `name`, `services`
- PATCH writable: same five fields as POST
- PATCH read-only: same as POST read-only

### Pitfall 3: Missing Read at end of Create/Update

**What goes wrong:** After `PostSubnet` returns the created object, computed fields (`enabled`, `interfaces`, `services`) may differ from zero values. If state is set from the POST body rather than the POST response, drift is detected on the first plan.

**Why it happens:** POST returns the full created object — using it to populate state is correct. But any subsequent `Update` that only patches and sets state from the plan (not the PATCH response) will miss server-side field changes.

**How to avoid:** Always call `mapSubnetToModel` using the API response from both POST and PATCH, not the plan data.

### Pitfall 4: LAG mock not seeded before data source test

**What goes wrong:** The LAG data source test fetches a LAG by name. If the mock handler has no LAGs seeded, it returns an empty list and the data source returns a not-found error, causing a false test failure.

**Why it happens:** Unlike resource mocks (which seed via POST), a read-only mock has no POST handler — LAGs must be pre-seeded via `store.Seed()` in the test setup.

**How to avoid:** Every LAG data source integration test must call `lagStore.Seed(...)` in `TestMain` or in the individual test's setup function. Document the seeding requirement in the mock file's package comment.

### Pitfall 5: vlan field zero-value omission in JSON

**What goes wrong:** `vlan` is an integer. If the subnet has VLAN 0 (untagged), the `omitempty` JSON tag will omit it from the PATCH body, preventing the operator from setting VLAN back to 0 after setting it to a non-zero value.

**Why it happens:** `omitempty` treats 0 as the zero value for integers and omits it.

**How to avoid:** In `SubnetPatch`, use `*int64` for `Vlan` (and `Mtu` for the same reason) so nil means "not in PATCH body" and 0 is serializable. In `SubnetPost`, use `int64` directly since POST always includes the user's intent.

## Code Examples

Verified patterns from existing source files:

### Model Struct Design (Subnet)

```go
// internal/client/models_network.go — add alongside existing NetworkInterface structs

// Subnet represents a FlashBlade subnet from GET /api/2.22/subnets.
type Subnet struct {
    ID                  string          `json:"id,omitempty"`
    Name                string          `json:"name"`
    Enabled             bool            `json:"enabled,omitempty"`
    Gateway             string          `json:"gateway,omitempty"`
    Interfaces          []NamedReference `json:"interfaces,omitempty"`
    LinkAggregationGroup *NamedReference `json:"link_aggregation_group,omitempty"`
    MTU                 int64           `json:"mtu,omitempty"`
    Prefix              string          `json:"prefix,omitempty"`
    Services            []string        `json:"services,omitempty"`
    VLAN                int64           `json:"vlan,omitempty"`
}

// SubnetPost contains writable fields for POST /api/2.22/subnets?names=<name>.
// Name is NOT included — it is passed as the ?names= query parameter.
type SubnetPost struct {
    Gateway             string          `json:"gateway,omitempty"`
    LinkAggregationGroup *NamedReference `json:"link_aggregation_group,omitempty"`
    MTU                 int64           `json:"mtu,omitempty"`
    Prefix              string          `json:"prefix,omitempty"`
    VLAN                int64           `json:"vlan,omitempty"`
}

// SubnetPatch contains writable fields for PATCH /api/2.22/subnets?names=<name>.
// Pointer types allow true omission of unchanged fields; *int64 handles vlan/mtu=0 correctly.
type SubnetPatch struct {
    Gateway             *string         `json:"gateway,omitempty"`
    LinkAggregationGroup *NamedReference `json:"link_aggregation_group,omitempty"`
    MTU                 *int64          `json:"mtu,omitempty"`
    Prefix              *string         `json:"prefix,omitempty"`
    VLAN                *int64          `json:"vlan,omitempty"`
}

// LinkAggregationGroup represents a FlashBlade LAG from GET /api/2.22/link-aggregation-groups.
// All fields are read-only — the struct is used only in GET responses.
type LinkAggregationGroup struct {
    ID         string   `json:"id,omitempty"`
    Name       string   `json:"name"`
    LagSpeed   int64    `json:"lag_speed,omitempty"`
    MacAddress string   `json:"mac_address,omitempty"`
    PortSpeed  int64    `json:"port_speed,omitempty"`
    Ports      []string `json:"ports,omitempty"`
    Status     string   `json:"status,omitempty"`
}
```

### Terraform Schema (Subnet Resource)

```go
// Required + user-controlled fields
"name": schema.StringAttribute{
    Required:    true,
    Description: "Name of the subnet. Changing this forces a new resource.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},
"prefix": schema.StringAttribute{
    Required:    true,
    Description: "IPv4 or IPv6 address prefix in CIDR notation (e.g. 10.0.0.0/24).",
},
"gateway": schema.StringAttribute{
    Optional:    true,
    Computed:    true,
    Description: "IPv4 or IPv6 gateway address for this subnet.",
},
"mtu": schema.Int64Attribute{
    Optional:    true,
    Computed:    true,
    Description: "Maximum transmission unit (MTU) in bytes.",
},
"vlan": schema.Int64Attribute{
    Optional:    true,
    Computed:    true,
    Description: "VLAN ID (0 = untagged).",
},
"lag_name": schema.StringAttribute{
    Optional:    true,
    Computed:    true,
    Description: "Name of the link aggregation group (LAG) associated with this subnet.",
},

// Computed read-only fields
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
},
"enabled": schema.BoolAttribute{
    Computed:    true,
    Description: "Whether the subnet is enabled.",
},
"services": schema.ListAttribute{
    Computed:    true,
    ElementType: types.StringType,
    Description: "Services provided by this subnet (inherited from LAG).",
},
"interfaces": schema.ListAttribute{
    Computed:    true,
    ElementType: types.StringType,
    Description: "Names of network interfaces associated with this subnet.",
},
```

### Terraform Schema (LAG Data Source)

```go
// Mirror of array_connection_data_source.go pattern — all Computed except name
"name": schema.StringAttribute{
    Required:    true,
    Description: "Name of the link aggregation group to look up.",
},
"id": schema.StringAttribute{Computed: true, Description: "Unique identifier."},
"status": schema.StringAttribute{Computed: true, Description: "Health status of the LAG (critical, healthy, identifying, unhealthy, unrecognized, unknown)."},
"mac_address": schema.StringAttribute{Computed: true, Description: "Unique MAC address assigned to the LAG."},
"port_speed": schema.Int64Attribute{Computed: true, Description: "Configured speed of each port in bits-per-second."},
"lag_speed": schema.Int64Attribute{Computed: true, Description: "Combined speed of all ports in bits-per-second."},
"ports": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Ports associated with the LAG."},
```

### Map Function Pattern (Subnet → Model)

```go
// Source: mirrors mapServerToModel in internal/provider/server_resource.go
func mapSubnetToModel(ctx context.Context, s *client.Subnet, data *subnetResourceModel, diags *diag.Diagnostics) {
    data.ID = types.StringValue(s.ID)
    data.Name = types.StringValue(s.Name)
    data.Enabled = types.BoolValue(s.Enabled)
    data.Gateway = types.StringValue(s.Gateway)
    data.MTU = types.Int64Value(s.MTU)
    data.Prefix = types.StringValue(s.Prefix)
    data.VLAN = types.Int64Value(s.VLAN)
    data.LagName = refToLagName(s.LinkAggregationGroup)

    // services: []string → types.List
    if len(s.Services) > 0 {
        svcList, d := types.ListValueFrom(ctx, types.StringType, s.Services)
        diags.Append(d...)
        data.Services = svcList
    } else {
        data.Services = types.ListNull(types.StringType)
    }

    // interfaces: []NamedReference → types.List[string]
    ifaceNames := make([]string, len(s.Interfaces))
    for i, iface := range s.Interfaces {
        ifaceNames[i] = iface.Name
    }
    if len(ifaceNames) > 0 {
        ifaceList, d := types.ListValueFrom(ctx, types.StringType, ifaceNames)
        diags.Append(d...)
        data.Interfaces = ifaceList
    } else {
        data.Interfaces = types.ListNull(types.StringType)
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| SDKv2 resources | terraform-plugin-framework v1.x | ~2022 | SDKv2 is maintenance-only; all new resources use framework |
| Inline GET struct reuse for POST/PATCH | Separate Post/Patch/Get structs | v2.0.1 decision | Prevents sending `ro` fields; required for correctness |

**No deprecated patterns apply to this phase.**

## Open Questions

1. **Does POST /api/2.22/subnets actually accept ?names= or another query param?**
   - What we know: The API spec marks `name` as `ro` in the POST body. User input confirms "Subnet name IS user-provided." The established FlashBlade pattern for user-provided names is `?names=` (alert-watchers, object-store-virtual-hosts).
   - What's unclear: Whether another param like `?create_ds=` (servers pattern) or a completely different mechanism is used.
   - Recommendation: Treat as `?names=` (HIGH probability correct). The first integration test will confirm. If the API returns 400, switch to `?create_ds=` or check whether name is a writable body field. The mock should use `?names=` to match the most common FlashBlade pattern.
   - **Confidence: MEDIUM** — inferred from context, not directly verified against a live array.

2. **Does PATCH /api/2.22/subnets support partial updates or require all writable fields?**
   - What we know: The PATCH body in FLASHBLADE_API.md lists the same five writable fields as POST. All existing PATCH implementations in this provider use partial updates (pointer fields, raw map in mock).
   - What's unclear: Whether sending only a subset of writable fields is safe (partial PATCH) or whether omitting a field resets it to default.
   - Recommendation: Use pointer types in `SubnetPatch` and send only changed fields. If the API behaves as full-replace, all fields can be sent safely (they are all writable). The safer assumption is partial PATCH semantics, consistent with all other endpoints.
   - **Confidence: HIGH** — consistent with all existing provider PATCH implementations.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing stdlib + terraform-plugin-testing v1.15.0 |
| Config file | none (go test ./...) |
| Quick run command | `go test ./internal/client/... ./internal/testmock/... -count=1 -timeout 30s` |
| Full suite command | `go test ./... -count=1 -timeout 120s` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LAG-01 | Data source reads LAG by name, returns all fields | integration | `go test ./internal/provider/ -run TestLAGDataSource -count=1` | ❌ Wave 0 |
| SUB-01 | Create subnet with all writable fields | integration | `go test ./internal/provider/ -run TestSubnetResource_Create -count=1` | ❌ Wave 0 |
| SUB-02 | Update subnet fields via PATCH | integration | `go test ./internal/provider/ -run TestSubnetResource_Update -count=1` | ❌ Wave 0 |
| SUB-03 | Delete subnet via destroy | integration | `go test ./internal/provider/ -run TestSubnetResource_Delete -count=1` | ❌ Wave 0 |
| SUB-04 | Data source reads subnet by name | integration | `go test ./internal/provider/ -run TestSubnetDataSource -count=1` | ❌ Wave 0 |
| SUB-05 | Import existing subnet, no drift on plan | integration | `go test ./internal/provider/ -run TestSubnetResource_Import -count=1` | ❌ Wave 0 |
| SUB-06 | Drift detected when subnet modified externally | integration | `go test ./internal/provider/ -run TestSubnetResource_Drift -count=1` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/client/... ./internal/testmock/... -count=1 -timeout 30s`
- **Per wave merge:** `go test ./... -count=1 -timeout 120s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/provider/subnet_resource_test.go` — covers SUB-01 through SUB-06
- [ ] `internal/provider/subnet_data_source_test.go` — covers SUB-04
- [ ] `internal/provider/link_aggregation_group_data_source_test.go` — covers LAG-01
- [ ] `internal/testmock/handlers/subnets.go` — required before provider tests
- [ ] `internal/testmock/handlers/link_aggregation_groups.go` — required before LAG data source test
- [ ] `internal/client/subnets.go` — client methods under test
- [ ] `internal/client/link_aggregation_groups.go` — client methods under test
- [ ] `internal/client/models_network.go` additions — Subnet, SubnetPost, SubnetPatch, LinkAggregationGroup structs

## Sources

### Primary (HIGH confidence)

- `FLASHBLADE_API.md` line 350-355 — Link Aggregation Groups endpoints (GET/POST/PATCH/DELETE); confirms GET-only is appropriate for data source
- `FLASHBLADE_API.md` line 794-799 — Subnets endpoints (GET/POST/PATCH/DELETE); writable vs read-only fields per method
- `FLASHBLADE_API.md` line 1075 — LinkAggregationGroup data model (all fields confirmed `ro` except `ports`)
- `FLASHBLADE_API.md` line 1321 — Subnet data model (field types and `ro` annotations confirmed)
- `internal/client/servers.go` — `?create_ds=` pattern for user-provided names at POST; reference for `PostSubnet`
- `internal/client/models_exports.go` — `ObjectStoreVirtualHostPost` with `AttachedServers []NamedReference`; reference for NamedReference list handling
- `internal/provider/server_resource.go` — `mapServerToModel`, `serverResourceModel`, `serverDNSAttrTypes`; reference for model struct and map function pattern
- `internal/provider/array_connection_data_source.go` — direct reference implementation for the LAG data source (read-only object, flat attributes)
- `internal/testmock/handlers/servers.go` — mock handler pattern (`?create_ds=` POST, byName/byID maps, `AddServer` seeder)
- `internal/testmock/handlers/object_store_virtual_hosts.go` — mock with `?names=` POST; direct reference for subnet mock POST handler
- `internal/provider/provider.go` lines 270-343 — `Resources()` and `DataSources()` registration lists; shows where to add new entries

### Secondary (MEDIUM confidence)

- `internal/client/models_admin.go` — `AlertWatcherPost` pattern (body has no Name; name passed via query param); supporting evidence for subnet `?names=` approach
- `internal/provider/server_data_source.go` — data source model and `mapServerDNSToDataSourceModel`; reference for subnet data source structure

### Tertiary (LOW confidence — needs validation)

- Subnet POST uses `?names=` (inferred from patterns, not verified against live array) — validate on first integration test run

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies; all primitives verified in existing files
- Architecture: HIGH — all patterns derived from reading existing source files directly
- API field classification (ro vs writable): HIGH — read directly from FLASHBLADE_API.md spec
- POST query param mechanism: MEDIUM — inferred from pattern consistency; needs first-test validation
- Pitfalls: HIGH — derived from API spec + codebase analysis of established patterns

**Research date:** 2026-03-30
**Valid until:** 2026-05-15 (stable API, established patterns)
