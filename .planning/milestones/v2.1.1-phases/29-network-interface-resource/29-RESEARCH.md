# Phase 29: Network Interface Resource & Data Source — Research

**Researched:** 2026-03-30
**Domain:** Terraform provider (Go) — FlashBlade VIP (Virtual IP) resource, REST API v2.22
**Confidence:** HIGH

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| NI-01 | Operator can create a network interface with name, address, subnet, type, services, and attached_servers via Terraform | POST /api/2.22/network-interfaces?names=<name> with NetworkInterfacePost body; name is user-provided via ?names= param |
| NI-02 | Operator can update network interface settings (address, services, attached_servers) via Terraform apply | PATCH /api/2.22/network-interfaces?names=<name> with NetworkInterfacePatch body (3 fields only) |
| NI-03 | Operator can delete a network interface via Terraform destroy | DELETE /api/2.22/network-interfaces?names=<name> |
| NI-04 | subnet and type are immutable after creation (RequiresReplace) | stringplanmodifier.RequiresReplace() on subnet_name and type attributes |
| NI-05 | services accepts a single value from: data, sts, egress-only, replication | StringAttribute with enum validator; single value, not a list |
| NI-06 | attached_servers is required for data/sts services and forbidden for egress-only/replication services | ConfigValidator or ValidateConfig method checking services + attached_servers combination |
| NI-07 | Operator can read an existing network interface by name via data source | GET /api/2.22/network-interfaces?names=<name> returns all fields |
| NI-08 | Operator can import an existing network interface into Terraform state with no drift on subsequent plan | ImportState using name as import ID; identical to subnet pattern |
| NI-09 | Drift detection logs changes when network interface is modified outside Terraform | tflog.Info on field diff in Read(), same pattern as subnet_resource.go |
| NI-10 | All read-only fields exposed as computed (enabled, gateway, mtu, netmask, vlan, realms) | Computed: true, UseStateForUnknown() on stable fields; derived from subnet |
</phase_requirements>

## Summary

Phase 29 implements `flashblade_network_interface` resource and data source. The VIP (Virtual IP) is the primary networking primitive for FlashBlade data access — it binds an IP address to a subnet and to one or more servers (for data/sts services).

The critical design decision resolved from user discussion: **VIP `name` IS user-provided**, passed via the `?names=` query parameter on POST (same pattern as `flashblade_subnet`). The FLASHBLADE_API.md marks `name` as `(ro string)` in the POST body model, but this only means it cannot be set in the JSON body — it is provided as a query param. The UI showing a "Name" field with placeholder "Interface1" confirms this. This resolves the Phase 29 blocker documented in STATE.md.

The second critical design decision: `services` is a **single-value string** (`data`, `sts`, `egress-only`, `replication`), not a list. This drives a cross-field validation requirement: `data` and `sts` must have exactly one entry in `attached_servers`; `egress-only` and `replication` must have no `attached_servers`.

All other patterns are direct ports from Phase 28 (subnet) with the addition of the `attached_servers` handling from `object_store_virtual_host_resource.go`.

**Primary recommendation:** Model after `subnet_resource.go` for the base structure and `object_store_virtual_host_resource.go` for `attached_servers` handling. Use `?names=` on POST (same as subnet). Add a `ValidateConfig` method for the services/attached_servers cross-field constraint.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| terraform-plugin-framework | v1.19.0 | Provider CRUD scaffold | Only supported SDK for new providers; SDKv2 maintenance-only |
| terraform-plugin-framework-timeouts | current | Timeout attributes | Consistent with all existing resources |
| terraform-plugin-log/tflog | current | Drift detection logging | Established pattern in every resource's Read() |
| net/http/httptest | stdlib | Mock server in tests | Zero external deps; established testmock pattern |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/google/uuid | current | UUID generation in mock handler | Mock POST response auto-generates ID |
| github.com/hashicorp/terraform-plugin-go/tftypes | current | Unit test state construction | Used in all `_test.go` files for tfsdk.Plan/State setup |

**No new go.mod entries required.** All dependencies already present.

---

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── client/
│   └── models_network.go      — ADD NetworkInterface, NetworkInterfacePost, NetworkInterfacePatch structs
│   └── network_interfaces.go  — ADD GetNetworkInterface, ListNetworkInterfaces, PostNetworkInterface,
│                                    PatchNetworkInterface, DeleteNetworkInterface
├── provider/
│   ├── network_interface_resource.go     — NEW (this phase)
│   └── network_interface_data_source.go  — NEW (this phase)
│   └── provider.go                       — MODIFY (register new types)
└── testmock/
    └── handlers/
        └── network_interfaces.go          — NEW (this phase)
```

Note: `models_network.go` and `network_interfaces.go` were added in Phase 28 (they live in the same file as Subnet/LAG models). If they don't exist yet as separate NI entries, they must be added as part of Wave 1 of this phase.

### Pattern 1: User-Provided Name via ?names= (same as Subnet)

**What:** POST body carries writable fields (`address`, `subnet`, `type`, `services`, `attached_servers`). Name is passed via `?names=<name>` query param. Name attribute is `Required: true` with `stringplanmodifier.RequiresReplace()`.

**When to use:** All POST/PATCH/DELETE operations on network interfaces.

**Example:**
```go
// internal/client/network_interfaces.go
func (c *FlashBladeClient) PostNetworkInterface(ctx context.Context, name string, body NetworkInterfacePost) (*NetworkInterface, error) {
    path := "/network-interfaces?names=" + url.QueryEscape(name)
    var resp ListResponse[NetworkInterface]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostNetworkInterface: empty response from server")
    }
    return &resp.Items[0], nil
}
```

### Pattern 2: Separate POST/PATCH/GET Structs (critical — mandatory)

**What:** Three distinct Go structs prevent accidental inclusion of read-only fields in PATCH bodies.

**When to use:** Always. PATCH body is exactly `address`, `services`, `attached_servers`. POST body adds `subnet` and `type`. GET response includes all fields.

**Example:**
```go
// internal/client/models_network.go (additions to Phase 28 file)

// NetworkInterface represents GET /api/2.22/network-interfaces response.
type NetworkInterface struct {
    ID              string           `json:"id,omitempty"`
    Name            string           `json:"name"`
    Address         string           `json:"address,omitempty"`
    Enabled         bool             `json:"enabled,omitempty"`
    Gateway         string           `json:"gateway,omitempty"`
    MTU             int64            `json:"mtu,omitempty"`
    Netmask         string           `json:"netmask,omitempty"`
    Services        []string         `json:"services,omitempty"`
    Subnet          *NamedReference  `json:"subnet,omitempty"`
    Type            string           `json:"type,omitempty"`
    VLAN            int64            `json:"vlan,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}

// NetworkInterfacePost — writable fields on POST. Name passed via ?names=.
type NetworkInterfacePost struct {
    Address         string           `json:"address,omitempty"`
    Services        []string         `json:"services,omitempty"`
    Subnet          *NamedReference  `json:"subnet,omitempty"`
    Type            string           `json:"type,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}

// NetworkInterfacePatch — writable fields on PATCH (address, services, attached_servers only).
// Services uses value type (not pointer) so empty slice sends as [] to allow clearing.
// AttachedServers same: full-replace semantics, empty list must serialize as [].
type NetworkInterfacePatch struct {
    Address         *string          `json:"address,omitempty"`
    Services        []string         `json:"services,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}
```

**Source:** FLASHBLADE_API.md line 373 (POST fields), line 374 (PATCH fields: address, attached_servers, services only), lines 1119/1123 (model definitions).

### Pattern 3: services as Single StringAttribute with Enum Validator

**What:** The API field is `[]string` but the valid values and UI confirm single-select semantics. Model as `types.String` (not `types.List`) in Terraform to enforce single value.

**When to use:** Schema definition and all mappers.

**Example:**
```go
// Schema attribute
"services": schema.StringAttribute{
    Required:    true,
    Description: "Service type for this VIP. One of: data, sts, egress-only, replication.",
    Validators: []validator.String{
        serviceTypeValidator(),
    },
},

// POST: expand single string to []string for API
body := NetworkInterfacePost{
    Services: []string{data.Services.ValueString()},
    ...
}

// Read: collapse []string (first element) to string for state
if len(ni.Services) > 0 {
    data.Services = types.StringValue(ni.Services[0])
} else {
    data.Services = types.StringNull()
}
```

### Pattern 4: Cross-Field Validation (services x attached_servers)

**What:** NI-06 requires that `data`/`sts` services mandate `attached_servers` (exactly 1) and `egress-only`/`replication` services forbid `attached_servers`.

**When to use:** Implement as a `resource.ResourceWithConfigValidators` method on the resource struct.

**Example:**
```go
var _ resource.ResourceWithConfigValidators = &networkInterfaceResource{}

func (r *networkInterfaceResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
    return []resource.ConfigValidator{
        networkInterfaceServicesValidator{},
    }
}

// networkInterfaceServicesValidator checks that attached_servers is set iff services in {data, sts}.
type networkInterfaceServicesValidator struct{}

func (v networkInterfaceServicesValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    var config networkInterfaceResourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }
    if config.Services.IsUnknown() || config.AttachedServers.IsUnknown() {
        return // defer to apply
    }
    svc := config.Services.ValueString()
    hasServers := !config.AttachedServers.IsNull() && len(config.AttachedServers.Elements()) > 0
    switch svc {
    case "data", "sts":
        if !hasServers {
            resp.Diagnostics.AddError("attached_servers required",
                fmt.Sprintf("services=%q requires exactly one attached_server", svc))
        }
    case "egress-only", "replication":
        if hasServers {
            resp.Diagnostics.AddError("attached_servers forbidden",
                fmt.Sprintf("services=%q must not have attached_servers", svc))
        }
    }
}
```

### Pattern 5: attached_servers — NamedReference List (from object_store_virtual_host)

**What:** API field is `[]NamedReference`. Expose as `types.List` of `types.StringType` in Terraform state. Use `modelServersToNamedRefs` helper pattern (already exists in `object_store_virtual_host_resource.go`). Full-replace semantics on PATCH.

**When to use:** All Create, Update, Read mappers.

**Example:**
```go
// Expand: Terraform []string → API []NamedReference
func niServersToNamedRefs(ctx context.Context, servers types.List, diags *diag.Diagnostics) []NamedReference {
    if servers.IsNull() || servers.IsUnknown() || len(servers.Elements()) == 0 {
        return nil
    }
    var names []string
    diags.Append(servers.ElementsAs(ctx, &names, false)...)
    if diags.HasError() {
        return nil
    }
    refs := make([]NamedReference, len(names))
    for i, n := range names {
        refs[i] = NamedReference{Name: n}
    }
    return refs
}

// Flatten: API []NamedReference → Terraform types.List
serverNames := make([]string, len(ni.AttachedServers))
for i, s := range ni.AttachedServers {
    serverNames[i] = s.Name
}
serverList, d := types.ListValueFrom(ctx, types.StringType, serverNames)
diags.Append(d...)
```

### Pattern 6: subnet_name as flat types.String (RequiresReplace)

**What:** API `subnet` field is a `*NamedReference` object. Expose as flat `types.String` (`subnet_name`) in Terraform state. Subnet is immutable after creation — add `RequiresReplace()`.

**When to use:** Schema definition, POST expansion, Read flattening.

**Example:**
```go
// Schema
"subnet_name": schema.StringAttribute{
    Required:    true,
    Description: "Name of the subnet this VIP is attached to. Immutable — changing forces recreation.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},

// Expand for POST
body := NetworkInterfacePost{
    Subnet: &client.NamedReference{Name: data.SubnetName.ValueString()},
    ...
}

// Flatten from GET response
data.SubnetName = refToSubnetName(ni.Subnet) // analogous to refToLagName in subnet_resource.go
```

### Pattern 7: Mock Handler — User-Provided Name via ?names= (same as subnet mock)

**What:** POST handler reads name from `?names=` query param, not auto-generates. `AddNetworkInterface(name, address, subnetName, svc)` seeder for tests. PATCH applies full-replace for `attached_servers` and `services` (value semantics).

**When to use:** `RegisterNetworkInterfaceHandlers(mux)` in test setup, mirroring `RegisterSubnetHandlers`.

**Example:**
```go
func (s *networkInterfaceStore) handlePost(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("names")
    if name == "" {
        WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
        return
    }
    // ...decode body, create NetworkInterface with name from query param...
}
```

### Anti-Patterns to Avoid

- **Sending subnet/type in PATCH body:** PATCH accepts only `address`, `services`, `attached_servers`. Including immutable fields causes 422 errors.
- **Modelling services as types.List:** Services is single-select. Using List causes plan drift and complicates the cross-field validator.
- **Omitting cross-field validator:** Without NI-06 validation, invalid configs (e.g., `services=egress-only` with `attached_servers=["server1"]`) reach the API and fail at apply time with an opaque error.
- **Using attached_servers=nil vs attached_servers=[] inconsistently:** When the API returns an empty `attached_servers`, always store as `types.ListValueFrom(ctx, types.StringType, []string{})` (empty list, not null) to avoid spurious drift.
- **Not calling Read at end of Create/Update:** Causes state divergence. Every Create and Update must call the shared `mapNetworkInterfaceToModel` after the API call.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ?names= query param encoding | Custom URL builder | `url.QueryEscape(name)` + string concat (established pattern in subnets.go) | Already the project standard |
| NamedReference expand/flatten | Custom JSON marshal | `refToSubnetName()` / `NamedReference{Name: ...}` (pattern from subnet_resource.go) | One-liner, consistent with all other resources |
| Cross-field validation | Manual in Create/Update | `resource.ResourceWithConfigValidators` interface | Runs at plan time, not just apply time |
| Enum validation | Manual string comparison in Create | Custom `validator.String` (pattern from validators.go) | Validated at plan time, surfaced early |
| Empty list vs null | Custom nil checks | `types.ListValueFrom(ctx, types.StringType, []string{})` | Eliminates null/empty drift |

---

## Common Pitfalls

### Pitfall 1: Services PATCH — empty slice must serialize as `[]`, not omitted

**What goes wrong:** If `Services []string` uses `omitempty`, clearing services (setting to empty) is silently dropped from the PATCH body, leaving old value in place.

**Why it happens:** `omitempty` on a slice omits it when nil/empty. But clearing a service list requires sending `"services": []`.

**How to avoid:** In `NetworkInterfacePatch`, do NOT use `omitempty` on `Services` and `AttachedServers`. Use value types (`[]string`, `[]NamedReference`). Only include these fields in the patch struct when they have actually changed.

**Warning signs:** `terraform apply` to clear `attached_servers` shows no error but old server stays attached. Second plan still shows diff.

### Pitfall 2: attached_servers full-replace semantics

**What goes wrong:** User has `attached_servers = ["server1", "server2"]`, removes `server2`, applies. If PATCH only sends `["server1"]` but mock/API does partial update, `server2` stays attached.

**Why it happens:** Mock PATCH handler must implement full-replace semantics for `attached_servers`, not merge. API replaces the entire list.

**How to avoid:** Mock `handlePatch` for `attached_servers` must overwrite the entire slice, not append. Resource Update must always send the full desired list.

**Warning signs:** Integration test shows `len(attached_servers) == 2` after update that should have reduced to 1.

### Pitfall 3: subnet_name roundtrip — nil NamedReference

**What goes wrong:** After POST, the GET response has `subnet: {"name": "data-subnet", "id": "uuid"}`. If `refToSubnetName` doesn't handle nil correctly, state gets `subnet_name = null` instead of `"data-subnet"`, triggering perpetual RequiresReplace.

**How to avoid:** `refToSubnetName(ref *NamedReference) types.String` must check `ref == nil || ref.Name == ""` → return `types.StringNull()`. This mirrors `refToLagName` in `subnet_resource.go` exactly.

### Pitfall 4: Cross-field validator running on unknown values during plan

**What goes wrong:** During `terraform plan`, `services` or `attached_servers` may be unknown (computed from another resource). Validator panics or produces false error on valid config.

**How to avoid:** Always check `IsUnknown()` on both `services` and `attached_servers` before evaluating the constraint. If either is unknown, return without error — defer validation to apply.

### Pitfall 5: Import ID confusion (name, not IP address)

**What goes wrong:** Operator tries `terraform import flashblade_network_interface.vip 10.21.200.10` (using IP address as import ID). Import fails or produces wrong state.

**How to avoid:** ImportState uses the VIP name (e.g., `vip0`, `Interface1`) as the import ID — the same identifier used by PATCH/DELETE. Document this prominently in the schema description and examples.

---

## Code Examples

### Client: Full CRUD pattern (source: subnets.go)

```go
// Source: internal/client/subnets.go (adapted)
func (c *FlashBladeClient) GetNetworkInterface(ctx context.Context, name string) (*NetworkInterface, error) {
    return getOneByName[NetworkInterface](c, ctx, "/network-interfaces?names="+url.QueryEscape(name), "network interface", name)
}

func (c *FlashBladeClient) ListNetworkInterfaces(ctx context.Context) ([]NetworkInterface, error) {
    var resp ListResponse[NetworkInterface]
    if err := c.get(ctx, "/network-interfaces", &resp); err != nil {
        return nil, err
    }
    return resp.Items, nil
}

func (c *FlashBladeClient) PostNetworkInterface(ctx context.Context, name string, body NetworkInterfacePost) (*NetworkInterface, error) {
    path := "/network-interfaces?names=" + url.QueryEscape(name)
    var resp ListResponse[NetworkInterface]
    if err := c.post(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PostNetworkInterface: empty response from server")
    }
    return &resp.Items[0], nil
}

func (c *FlashBladeClient) PatchNetworkInterface(ctx context.Context, name string, body NetworkInterfacePatch) (*NetworkInterface, error) {
    path := "/network-interfaces?names=" + url.QueryEscape(name)
    var resp ListResponse[NetworkInterface]
    if err := c.patch(ctx, path, body, &resp); err != nil {
        return nil, err
    }
    if len(resp.Items) == 0 {
        return nil, fmt.Errorf("PatchNetworkInterface: empty response from server")
    }
    return &resp.Items[0], nil
}

func (c *FlashBladeClient) DeleteNetworkInterface(ctx context.Context, name string) error {
    return c.delete(ctx, "/network-interfaces?names="+url.QueryEscape(name))
}
```

### Resource: Schema attributes for NI-specific fields

```go
// Source: pattern from internal/provider/subnet_resource.go + object_store_virtual_host_resource.go
"name": schema.StringAttribute{
    Required:    true,
    Description: "The name of the VIP. Immutable — changing forces recreation.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},
"subnet_name": schema.StringAttribute{
    Required:    true,
    Description: "Name of the subnet this VIP is attached to. Immutable.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},
"type": schema.StringAttribute{
    Required:    true,
    Description: "The type of network interface. The only valid value is 'vip'. Immutable.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},
"services": schema.StringAttribute{
    Required:    true,
    Description: "Service type. One of: data, sts, egress-only, replication.",
    Validators: []validator.String{serviceTypeValidator()},
},
"address": schema.StringAttribute{
    Required:    true,
    Description: "IPv4 or IPv6 address for this VIP.",
},
"attached_servers": schema.ListAttribute{
    Optional:    true,
    Computed:    true,
    ElementType: types.StringType,
    Description: "Server names attached to this VIP. Required for data/sts; forbidden for egress-only/replication.",
},
// Computed-only read-only fields derived from subnet:
"enabled":  schema.BoolAttribute{Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
"gateway":  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
"mtu":      schema.Int64Attribute{Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
"netmask":  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
"vlan":     schema.Int64Attribute{Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
"realms":   schema.ListAttribute{Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
```

### Resource Model

```go
type networkInterfaceResourceModel struct {
    ID              types.String   `tfsdk:"id"`
    Name            types.String   `tfsdk:"name"`
    Address         types.String   `tfsdk:"address"`
    SubnetName      types.String   `tfsdk:"subnet_name"`
    Type            types.String   `tfsdk:"type"`
    Services        types.String   `tfsdk:"services"`      // single value, not list
    AttachedServers types.List     `tfsdk:"attached_servers"`
    Enabled         types.Bool     `tfsdk:"enabled"`
    Gateway         types.String   `tfsdk:"gateway"`
    MTU             types.Int64    `tfsdk:"mtu"`
    Netmask         types.String   `tfsdk:"netmask"`
    VLAN            types.Int64    `tfsdk:"vlan"`
    Realms          types.List     `tfsdk:"realms"`
    Timeouts        timeouts.Value `tfsdk:"timeouts"`
}
```

### Drift Detection in Read

```go
// Source: internal/provider/subnet_resource.go Read() pattern
if data.Address.ValueString() != ni.Address {
    tflog.Info(ctx, "network interface drift detected: address changed",
        map[string]any{"name": name, "state": data.Address.ValueString(), "api": ni.Address})
}
if data.Services.ValueString() != firstService(ni.Services) {
    tflog.Info(ctx, "network interface drift detected: services changed",
        map[string]any{"name": name, "state": data.Services.ValueString(), "api": ni.Services})
}
```

### Test Helper Pattern (source: subnet_resource_test.go)

```go
func newTestNetworkInterfaceResource(t *testing.T, ms *testmock.MockServer) *networkInterfaceResource {
    t.Helper()
    c, err := client.NewClient(context.Background(), client.Config{
        Endpoint:           ms.URL(),
        APIToken:           "test-token",
        InsecureSkipVerify: true,
        MaxRetries:         1,
        RetryBaseDelay:     1 * time.Millisecond,
    })
    if err != nil {
        t.Fatalf("NewClient: %v", err)
    }
    return &networkInterfaceResource{client: c}
}
```

---

## State of the Art

| Old Assumption | Corrected Understanding | Impact |
|----------------|------------------------|--------|
| VIP name is auto-generated (ro) | VIP name IS user-provided via ?names= query param | name attribute is Required + RequiresReplace, NOT Computed-only |
| services is a list | services is single-select (single string) | StringAttribute, not ListAttribute; simpler mapper |
| No cross-field validation in this codebase | ConfigValidators interface exists and is used | Need to implement resource.ResourceWithConfigValidators |

**Deprecated assumptions (from prior milestone research summary):**
- SUMMARY.md stated "VIP `name` is auto-generated" — this is INCORRECT per user discussion and UI evidence. Use `Required` + `RequiresReplace` like `subnet` name.
- ARCHITECTURE.md Pattern 1 (Auto-Named POST, no ?names=) is WRONG for this resource. The correct pattern is identical to subnets (user-provided via ?names=).

---

## Open Questions

1. **realms field handling**
   - What we know: `realms` is `(ro array)` in the NetworkInterface model. Derived from subnet.
   - What's unclear: Whether the API returns it as `[]string` (realm names) or `[]NamedReference` objects.
   - Recommendation: Start with `types.ListNull(types.StringType)` until confirmed. If it's an array of objects, use name extraction like `interfaces` in `subnet_resource.go`. Mark as LOW priority — realms are out of scope per REQUIREMENTS.md.

2. **services valid values confirmation**
   - What we know: User discussion confirmed: `data`, `sts`, `egress-only`, `replication`.
   - What's unclear: Whether `data` and `sts` are separate values or if there's a combined `data-s3`/`data-nfs` split.
   - Recommendation: Use exactly these four values in the validator: `data`, `sts`, `egress-only`, `replication`. The prior SUMMARY.md listed different values (`data-s3`, `data-nfs`) — the user's context explicitly overrides this.

3. **attached_servers: Optional vs Required conditionality**
   - What we know: Required for `data`/`sts`, forbidden for `egress-only`/`replication`.
   - What's unclear: What count is expected — exactly 1, or 1+?
   - Recommendation: Schema as `Optional: true, Computed: true`. Enforce count in `ConfigValidators`. User context says "exactly 1 server" for data/sts — validate `len == 1` in the cross-field validator.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing stdlib + terraform-plugin-testing |
| Config file | none (standard `go test`) |
| Quick run command | `go test ./internal/... -run TestUnit_ -count=1` |
| Full suite command | `go test ./internal/... -count=1` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NI-01 | Create with name, address, subnet, type, services, attached_servers | unit | `go test ./internal/provider/ -run TestUnit_NetworkInterface_Create -v` | Wave 0 |
| NI-02 | Update address, services, attached_servers via PATCH | unit | `go test ./internal/provider/ -run TestUnit_NetworkInterface_Update -v` | Wave 0 |
| NI-03 | Delete removes resource | unit | `go test ./internal/provider/ -run TestUnit_NetworkInterface_Delete -v` | Wave 0 |
| NI-04 | subnet_name and type are RequiresReplace | unit (schema) | `go test ./internal/provider/ -run TestUnit_NetworkInterface_Schema -v` | Wave 0 |
| NI-05 | services enum validator rejects invalid values | unit (validator) | `go test ./internal/provider/ -run TestUnit_NetworkInterface_ServicesValidator -v` | Wave 0 |
| NI-06 | Cross-field validation: services x attached_servers | unit (ConfigValidator) | `go test ./internal/provider/ -run TestUnit_NetworkInterface_ConfigValidator -v` | Wave 0 |
| NI-07 | Data source reads by name | unit | `go test ./internal/provider/ -run TestUnit_NetworkInterfaceDataSource_Read -v` | Wave 0 |
| NI-08 | Import populates all attributes, no drift | unit | `go test ./internal/provider/ -run TestUnit_NetworkInterface_Import -v` | Wave 0 |
| NI-09 | Drift detection logs changes | unit | `go test ./internal/provider/ -run TestUnit_NetworkInterface_Drift -v` | Wave 0 |
| NI-10 | Computed fields (enabled, gateway, mtu, netmask, vlan, realms) populated | unit (schema) | `go test ./internal/provider/ -run TestUnit_NetworkInterface_Schema -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -run TestUnit_ -count=1`
- **Per wave merge:** `go test ./internal/... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/testmock/handlers/network_interfaces.go` — mock handler (needed by all unit tests)
- [ ] `internal/client/models_network.go` additions — NetworkInterface, NetworkInterfacePost, NetworkInterfacePatch structs
- [ ] `internal/client/network_interfaces.go` — CRUD client methods
- [ ] `internal/provider/network_interface_resource.go` — resource implementation
- [ ] `internal/provider/network_interface_data_source.go` — data source implementation
- [ ] `internal/provider/network_interface_resource_test.go` — unit tests
- [ ] `internal/provider/network_interface_data_source_test.go` — data source unit tests

---

## Sources

### Primary (HIGH confidence)
- `FLASHBLADE_API.md` lines 372-375 — POST/PATCH/DELETE/GET endpoints, body field list with ro markers
- `FLASHBLADE_API.md` line 1119 — NetworkInterface model (all fields with ro/mutable markers)
- `FLASHBLADE_API.md` line 1123 — NetworkInterfacePatch model (address, attached_servers, services)
- `internal/client/subnets.go` — ?names= pattern for PostSubnet, GetSubnet, PatchSubnet, DeleteSubnet
- `internal/client/models_network.go` — SubnetPost/SubnetPatch struct patterns (pointer types for zero-value safety)
- `internal/provider/subnet_resource.go` — Base CRUD resource pattern, drift detection, RequiresReplace usage
- `internal/provider/subnet_data_source.go` — Data source pattern
- `internal/provider/object_store_virtual_host_resource.go` — attached_servers NamedReference expand/flatten, ConfigValidators interface usage point of reference
- `internal/provider/validators.go` — Enum validator pattern (alphanumericStringValidator)
- `internal/testmock/handlers/subnets.go` — Full mock handler CRUD with ?names= pattern
- `internal/provider/subnet_resource_test.go` — Unit test structure (buildXType, nullXConfig, planWith helpers)
- `internal/provider/subnet_data_source_test.go` — Data source test structure
- Phase prompt `additional_context` — User-confirmed: name is user-provided, services is single-select, cross-field rules for attached_servers

### Secondary (MEDIUM confidence)
- `.planning/research/ARCHITECTURE.md` — Model designs, component boundaries, data flow diagrams
- `.planning/research/SUMMARY.md` — Milestone context (NOTE: name auto-generation assumption is now CORRECTED)
- `.planning/STATE.md` — Accumulated decisions; Phase 29 blocker about POST name behavior is now RESOLVED

### Tertiary (LOW confidence, flagged)
- `services` enum values (`data`, `sts`, `egress-only`, `replication`) — from user context; confirmed as correct set but not verified against official FlashBlade API documentation
- `realms` field structure — assumed `[]string` based on how `services` is modelled; should be verified on first Read against real or mock API

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies, all patterns present in codebase
- Architecture: HIGH — POST ?names= pattern confirmed, struct separation confirmed from API spec
- Critical design decision (name): HIGH — user context explicitly states user-provided name with ?names= query param
- services single-select: HIGH — user context explicit; UI evidence provided
- Cross-field validator: HIGH — interface pattern exists in framework, need to implement
- Pitfalls: HIGH — derived from existing codebase analysis and API spec
- realms field structure: LOW — assumed, not verified

**Research date:** 2026-03-30
**Valid until:** 2026-04-30 (stable domain — FlashBlade API v2.22 is the pinned version)

**Key correction from prior research:** SUMMARY.md and ARCHITECTURE.md assumed VIP name is auto-generated (no ?names=). This is WRONG. The additional_context in the task prompt (from user discussion) confirms name IS user-provided. This changes the schema `name` attribute from `Computed: true, UseStateForUnknown()` to `Required: true, RequiresReplace()`, and the POST client method from `PostNetworkInterface(ctx, body)` to `PostNetworkInterface(ctx, name, body)`.
