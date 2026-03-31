# Architecture Research

**Domain:** Terraform provider — FlashBlade network interface (VIP) integration
**Researched:** 2026-03-30
**Confidence:** HIGH (all conclusions drawn directly from reading existing source files)

## Standard Architecture

### System Overview

```
┌───────────────────────────────────────────────────────────────┐
│                     Provider Layer                             │
│  (internal/provider/)                                          │
│  ┌──────────────┐  ┌────────────────┐  ┌──────────────────┐   │
│  │ network_     │  │ server_        │  │ server_          │   │
│  │ interface_   │  │ resource.go    │  │ data_source.go   │   │
│  │ resource.go  │  │ (MODIFIED)     │  │ (MODIFIED)       │   │
│  │ (NEW)        │  └────────────────┘  └──────────────────┘   │
│  └──────────────┘                                             │
│  ┌───────────────────────────────────────────────────────┐    │
│  │ network_interface_data_source.go (NEW)                 │    │
│  └───────────────────────────────────────────────────────┘    │
├───────────────────────────────────────────────────────────────┤
│                     Client Layer                               │
│  (internal/client/)                                            │
│  ┌─────────────────────┐  ┌──────────────────────────────┐    │
│  │ network_interfaces. │  │ models_network.go (NEW)       │    │
│  │ go (NEW)            │  │ NetworkInterface              │    │
│  │ GetNetworkInterface │  │ NetworkInterfacePost          │    │
│  │ ListNetworkInterfaces│  │ NetworkInterfacePatch         │    │
│  │ PostNetworkInterface│  └──────────────────────────────┘    │
│  │ PatchNetworkInterface│                                      │
│  │ DeleteNetworkInterface│                                     │
│  └─────────────────────┘                                      │
├───────────────────────────────────────────────────────────────┤
│                     FlashBlade REST API v2.22                  │
│  GET/POST/PATCH/DELETE /api/2.22/network-interfaces            │
└───────────────────────────────────────────────────────────────┘

Mock Layer (internal/testmock/handlers/)
  ┌──────────────────────────────────────┐
  │ network_interfaces.go (NEW)           │
  │ networkInterfaceStore (in-memory)     │
  └──────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Status |
|-----------|----------------|--------|
| `internal/client/models_network.go` | Go structs for NetworkInterface, NetworkInterfacePost, NetworkInterfacePatch | NEW |
| `internal/client/network_interfaces.go` | CRUD methods on FlashBladeClient for /network-interfaces | NEW |
| `internal/provider/network_interface_resource.go` | flashblade_network_interface resource (full CRUD + import) | NEW |
| `internal/provider/network_interface_data_source.go` | flashblade_network_interface data source (read by name) | NEW |
| `internal/provider/server_resource.go` | Add `network_interfaces` computed attribute showing VIPs attached to server | MODIFIED |
| `internal/provider/server_data_source.go` | Add `network_interfaces` computed attribute showing VIPs | MODIFIED |
| `internal/testmock/handlers/network_interfaces.go` | In-memory mock for /api/2.22/network-interfaces | NEW |
| `internal/provider/provider.go` | Register new resource + data source factory functions | MODIFIED |

## Recommended Project Structure

```
internal/
├── client/
│   ├── models_network.go        # NEW — NetworkInterface, Post, Patch structs
│   ├── network_interfaces.go    # NEW — CRUD client methods
│   └── [existing files unchanged]
├── provider/
│   ├── network_interface_resource.go      # NEW
│   ├── network_interface_data_source.go   # NEW
│   ├── server_resource.go                 # MODIFIED — add network_interfaces attribute
│   ├── server_data_source.go              # MODIFIED — add network_interfaces attribute
│   └── provider.go                        # MODIFIED — register new types
└── testmock/
    └── handlers/
        ├── network_interfaces.go          # NEW
        └── [existing files unchanged]
```

### Structure Rationale

- **models_network.go:** Follows established pattern — one models file per domain area (models_admin.go, models_exports.go, models_storage.go). Network interface structs do not belong in models_exports.go or models_storage.go because VIPs are a network primitive, not a storage or export concept.
- **network_interfaces.go:** Follows the one-client-file-per-resource-type pattern (servers.go, buckets.go, object_store_virtual_hosts.go). Keeps routing/transport logic separate from models.
- No new subdirectory needed — the existing flat structure scales to this addition.

## Architectural Patterns

### Pattern 1: Auto-Named POST (no ?names=, no ?create_ds=)

**What:** The FlashBlade API auto-generates the VIP name on POST. The POST body carries `address`, `attached_servers`, `services`, `subnet`, and `type`. No name parameter exists in the query string at creation time.

**When to use:** Used exclusively for `PostNetworkInterface`. This differs from server (uses `?create_ds=`) and object-store-virtual-host (uses `?names=`).

**Trade-offs:** The resource cannot know its Terraform ID before the API call completes. The `name` attribute must be `Computed: true` with `stringplanmodifier.UseStateForUnknown()`. Import works via the server-assigned name.

**Example:**
```go
// client/network_interfaces.go
func (c *FlashBladeClient) PostNetworkInterface(ctx context.Context, body NetworkInterfacePost) (*NetworkInterface, error) {
    // No ?names= or ?create_ds= — name is assigned by the array
    path := "/network-interfaces"
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

### Pattern 2: NamedReference List for attached_servers (API []NamedReference → Terraform []string)

**What:** The API returns `attached_servers` as an array of NamedReference objects (`{name, id}`). Both `objectStoreVirtualHostResource` (VH → server) and `NetworkInterface` (VIP → server) use this pattern. In Terraform state, expose only the name strings — consumers never need the internal ID.

**When to use:** Whenever an API field is `[]NamedReference` and the user manages the list by name. The `modelServersToNamedRefs` helper in `object_store_virtual_host_resource.go` is the reference implementation and can be duplicated locally in the new resource file.

**Trade-offs:** Full-replace semantics on PATCH — always send the complete desired list, not a delta. This matches the API behavior for `attached_servers` on network interfaces (same as virtual hosts). The mock handler must also apply full-replace on PATCH for `attached_servers`.

**Example:**
```go
// Convert []NamedReference from API response into types.List of string for state
serverNames := make([]string, len(ni.AttachedServers))
for i, s := range ni.AttachedServers {
    serverNames[i] = s.Name
}
serverList, d := types.ListValueFrom(ctx, types.StringType, serverNames)
```

### Pattern 3: Server enrichment — computed read-only list of VIP names

**What:** The server resource/data source gains a new `network_interfaces` computed attribute listing VIP names attached to the server. This data is NOT returned by `GET /servers` — it requires a second API call: `ListNetworkInterfaces` (no filter), then filter in Go by `attached_servers[].name == serverName`.

**When to use:** When Terraform consumers need to discover endpoints without managing VIPs as a separate resource. The operational workflow is: create server → create network interface pointing at server → read server data source to discover VIP address.

**Trade-offs:**
- Option A (recommended): `ListNetworkInterfaces` with no filter, filter client-side. No risk of undocumented query params.
- Option B: API filter expression `filter=attached_servers[name='<server>']` — not documented for this endpoint.

Use Option A. Arrays have tens of VIPs at most; the O(n*m) client-side join is negligible.

**Example:**
```go
// In serverResource.Read(), after GetServer():
vips, err := r.client.ListNetworkInterfaces(ctx, ListNetworkInterfacesOpts{})
if err != nil {
    // log warning but don't fail Read — enrichment is best-effort
}
var attachedVIPNames []string
for _, vip := range vips {
    for _, s := range vip.AttachedServers {
        if s.Name == serverName {
            attachedVIPNames = append(attachedVIPNames, vip.Name)
        }
    }
}
```

The `network_interfaces` attribute is `Computed: true` only — users cannot set it. No `PlanModifier` is needed since it is always fetched fresh from the API on Read.

### Pattern 4: Mock handler — auto-name generation on POST

**What:** The mock must replicate the real API behavior of auto-assigning names on POST. Use an incrementing counter protected by the store mutex: `name = fmt.Sprintf("vip%d", store.nextID)`.

**When to use:** In `RegisterNetworkInterfaceHandlers`, the POST handler increments a counter and assigns a name, then returns it. Tests that need to assert the name call `AddNetworkInterface` to seed the store with a known name, or inspect the POST response.

**Trade-offs:** The mock's ImportState test must use the name from the POST response as the import ID. This is the same pattern as `objectStoreVirtualHostResource`.

## Data Flow

### VIP Create Flow

```
terraform apply
    ↓
networkInterfaceResource.Create()
    ↓ reads plan (address, subnet.name, services, attached_servers, type)
PostNetworkInterface(ctx, NetworkInterfacePost{...})
    ↓ POST /api/2.22/network-interfaces (no name param)
FlashBlade assigns name (e.g. "vip0")
    ↓ returns NetworkInterface{Name:"vip0", Address:"10.0.0.1", ...}
mapNetworkInterfaceToModel() → state.Name = "vip0", state.ID = "<uuid>"
    ↓
resp.State.Set(ctx, &data)
```

### VIP PATCH / DELETE Flow

```
terraform apply (update or destroy)
    ↓
state.Name.ValueString() → "vip0"
PatchNetworkInterface(ctx, "vip0", NetworkInterfacePatch{...})
    ↓ PATCH /api/2.22/network-interfaces?names=vip0
or
DeleteNetworkInterface(ctx, "vip0")
    ↓ DELETE /api/2.22/network-interfaces?names=vip0
```

### Server Read with VIP Enrichment Flow

```
terraform plan/apply (server resource or data source Read)
    ↓
GetServer(ctx, serverName)           → Server{Name, ID, DNS, ...}
ListNetworkInterfaces(ctx, opts{})   → []NetworkInterface
    ↓ client-side filter: vip.AttachedServers contains serverName
attachedVIPNames = ["vip0", "vip1"]
    ↓
mapServerToModel() appends network_interfaces = types.List["vip0", "vip1"]
```

### Key Data Flows

1. **VIP-to-server association:** NetworkInterface.AttachedServers is the authoritative source of the relationship. The server object returned by `GET /servers` has no VIP field — enrichment is always a client-side join in the provider.
2. **Import:** `terraform import flashblade_network_interface.example vip0` sets `state.Name = "vip0"`, then Read fetches by `?names=vip0`. Name is the import ID, consistent with all other resources.

## Integration Points

### New Components

| Component | Integration | Notes |
|-----------|-------------|-------|
| `models_network.go` | Consumed by `network_interfaces.go` and provider files | Contains `NamedReference` for `subnet` field — already defined in `models_common.go`, reuse directly |
| `network_interfaces.go` | Methods added to `*FlashBladeClient` | Identical signature pattern as `object_store_virtual_hosts.go` |
| `network_interface_resource.go` | Registered in `provider.go` Resources() | `NewNetworkInterfaceResource` factory function |
| `network_interface_data_source.go` | Registered in `provider.go` DataSources() | `NewNetworkInterfaceDataSource` factory function |
| `handlers/network_interfaces.go` | Used in testmock server setup | Wire into existing mock server registration entrypoint alongside other `RegisterXHandlers` calls |

### Modified Components

| Component | What Changes | Risk |
|-----------|-------------|------|
| `server_resource.go` | Add `NetworkInterfaces types.List` to `serverResourceModel`; add schema attribute `network_interfaces`; add `ListNetworkInterfaces` call in Read + filter logic | LOW — strictly additive, no existing field touched |
| `server_data_source.go` | Same additions in `serverDataSourceModel` and Read | LOW — additive |
| `provider.go` | Append two entries to Resources() and one entry to DataSources() slices | TRIVIAL |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `server_resource.go` ↔ `network_interfaces.go` | Direct method call on `*FlashBladeClient` | No new interface needed — existing client injection pattern handles it |
| `network_interface_resource.go` ↔ `models_network.go` | Struct field access via `client` import | Same import path as all other resources |
| `handlers/network_interfaces.go` ↔ mock server | `RegisterNetworkInterfaceHandlers(mux)` returns `*networkInterfaceStore` | Consistent with `RegisterServerHandlers` return pattern (store used by tests to seed data) |

## API Constraint: POST name auto-generation

The POST body for `/network-interfaces` includes `name` as a field in the API schema, but it is marked read-only (`ro`). Consequences:

- `NetworkInterfacePost` must NOT include `Name` — it will be ignored or rejected.
- `NetworkInterfacePatch` must NOT include `Name` — VIP names are immutable after creation.
- The Terraform `name` attribute is `Computed: true` only, with `stringplanmodifier.UseStateForUnknown()`.
- Import uses the server-assigned name (e.g. `vip0`) as the Terraform import ID.

## Model Design

### NetworkInterface (GET response)

```go
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
```

### NetworkInterfacePost (writable fields on POST)

```go
type NetworkInterfacePost struct {
    Address         string           `json:"address,omitempty"`
    Services        []string         `json:"services,omitempty"`
    Subnet          *NamedReference  `json:"subnet,omitempty"`
    Type            string           `json:"type,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}
```

### NetworkInterfacePatch (writable fields on PATCH)

```go
type NetworkInterfacePatch struct {
    Address         *string          `json:"address,omitempty"`
    Services        []string         `json:"services,omitempty"`
    AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}
```

Notes:
- `subnet` and `type` are absent from Patch — per API docs, only `address`, `services`, `attached_servers` are patchable on PATCH.
- `Address` uses a pointer so an unset patch field is truly omitted from the JSON body.
- `Services` and `AttachedServers` use value types because an empty slice must serialize as `[]` (not omitted) to allow clearing the list.
- `NamedReference` is already defined in `models_common.go` — no new type needed.

## Recommended Build Order

Dependencies flow top-to-bottom. Each step is independently testable before the next begins.

| Step | Task | Depends On | Notes |
|------|------|-----------|-------|
| 1 | `models_network.go` | nothing | Foundation — structs used by all other components |
| 2 | `network_interfaces.go` | step 1 | Client CRUD; unit-testable with mock HTTP server |
| 3 | `handlers/network_interfaces.go` | step 1 | Mock handler; enables integration tests for steps 4-7 |
| 4 | `network_interface_resource.go` | steps 2, 3 | Full CRUD resource + integration tests |
| 5 | `network_interface_data_source.go` | steps 2, 3 | Data source + integration tests; parallel with step 4 |
| 6 | `server_resource.go` enrichment | steps 2, 3 | Modify server Read to add VIP list; mock needs cross-resource data setup |
| 7 | `server_data_source.go` enrichment | steps 2, 3 | Same enrichment for data source; parallel with step 6 |
| 8 | `provider.go` registration | steps 4, 5 | Two-line additions to Resources() and DataSources() slices |

Steps 4+5 can be done in parallel after step 3. Steps 6+7 can be done in parallel after step 5.

## Anti-Patterns

### Anti-Pattern 1: Including `name` in NetworkInterfacePost body

**What people do:** Mirror the GET model into the POST struct, including the `name` field.
**Why it's wrong:** The API marks `name` as `ro` in the POST body — it is auto-assigned by the array. Including it may cause an API error, or produce a plan that shows an impossible known name value before creation.
**Do this instead:** `NetworkInterfacePost` has no `Name` field. The resource's `name` attribute is `Computed: true` and populated only after the POST response is received.

### Anti-Pattern 2: Querying VIPs per-server via undocumented filter params

**What people do:** Try `GET /network-interfaces?server_names=<name>` or a filter expression without verifying API support.
**Why it's wrong:** The API docs do not document a `server_names` filter for this endpoint. Using undocumented params risks 400 errors or silently empty results.
**Do this instead:** Call `ListNetworkInterfaces` with no filter, then filter `AttachedServers` in Go. The array has at most a few dozen VIPs — the in-memory join is negligible.

### Anti-Pattern 3: Storing subnet derived fields as managed plan attributes

**What people do:** Model gateway, mtu, vlan, netmask as Optional attributes on the resource, expecting users to set them.
**Why it's wrong:** All these fields are derived (`ro` in the API) — they cannot be set on POST or PATCH and will cause perpetual plan drift if treated as user-specified.
**Do this instead:** Expose `subnet` as a single `types.String` (the subnet name) on the resource. Expose `gateway`, `mtu`, `netmask`, `vlan`, `enabled` as separate `Computed: true` attributes populated from the GET response. They inform users but are never part of the desired state.

### Anti-Pattern 4: Managing VIP attachment from the server resource

**What people do:** Add `network_interfaces` as a writable attribute on the server resource, attempting to attach VIPs by setting `server.network_interfaces = ["vip0"]`.
**Why it's wrong:** The association is owned by the network interface object (`NetworkInterface.AttachedServers`), not the server. Managing it from the server side creates conflicting state between the two resources and has no API backing.
**Do this instead:** `network_interfaces` on server resource/data source is strictly `Computed: true` — a read-only discovery mechanism. VIP-to-server attachment is managed exclusively through `flashblade_network_interface.attached_servers`.

## Sources

- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/FLASHBLADE_API.md` lines 370-383 (network-interfaces endpoints), line 1119 (NetworkInterface model), line 1123 (NetworkInterfacePatch model) — HIGH confidence
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/client/object_store_virtual_hosts.go` — reference for auto-named POST and NamedReference handling — HIGH confidence
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/object_store_virtual_host_resource.go` — reference for `attached_servers` as `[]NamedReference` to `types.List[string]` — HIGH confidence
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/server_resource.go` — existing server model; enrichment is additive — HIGH confidence
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/testmock/handlers/servers.go` — mock handler pattern (store struct, RegisterXHandlers, AddX seeder) — HIGH confidence
- `/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/internal/provider/provider.go` lines 270-343 — registration pattern for Resources() and DataSources() — HIGH confidence

---
*Architecture research for: FlashBlade network interface (VIP) integration into existing Terraform provider*
*Researched: 2026-03-30*
