---
phase: quick-7
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/client/models_exports.go
  - internal/client/servers_test.go
  - internal/testmock/handlers/servers.go
  - internal/provider/server_resource.go
  - internal/provider/server_resource_test.go
  - internal/provider/server_data_source.go
  - internal/provider/server_data_source_test.go
autonomous: true
requirements: [DNS-REFACTOR]

must_haves:
  truths:
    - "Server.DNS is a list of NamedReference (name/id), not inline objects"
    - "Server.DirectoryServices is a list of NamedReference, read-only computed"
    - "dns schema attribute is a flat list of strings (DNS config names)"
    - "directory_services schema attribute is a computed list of strings"
    - "Schema version is 2 with v1->v2 StateUpgrader"
    - "All existing tests pass with updated DNS format"
  artifacts:
    - path: "internal/client/models_exports.go"
      provides: "Server with []NamedReference DNS and DirectoryServices"
      contains: "DirectoryServices"
    - path: "internal/provider/server_resource.go"
      provides: "Resource schema v2 with string-list dns and directory_services"
      contains: "Version:     2"
    - path: "internal/provider/server_data_source.go"
      provides: "Data source schema with string-list dns and directory_services"
      contains: "directory_services"
  key_links:
    - from: "internal/provider/server_resource.go"
      to: "internal/client/models_exports.go"
      via: "mapServerToModel reads []NamedReference not []ServerDNS"
      pattern: "srv\\.DNS.*\\.Name"
    - from: "internal/provider/server_resource.go"
      to: "internal/client/models_exports.go"
      via: "Create/Update builds []NamedReference from string list"
      pattern: "NamedReference\\{Name:"
---

<objective>
Refactor flashblade_server DNS field from incorrect inline objects (domain/nameservers/services) to a list of DNS configuration names (NamedReference), matching the real FlashBlade API. Add directory_services as computed read-only. Bump schema v1 to v2 with StateUpgrader.

Purpose: The current DNS schema is wrong -- the API returns `[{name, id, resource_type}]` not `[{domain, nameservers, services}]`. This fixes the data model to match reality.
Output: All server files updated, tests green.
</objective>

<execution_context>
@/home/gule/.claude/get-shit-done/workflows/execute-plan.md
@/home/gule/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/client/models_exports.go
@internal/client/models_common.go
@internal/provider/server_resource.go
@internal/provider/server_data_source.go
@internal/provider/server_resource_test.go
@internal/provider/server_data_source_test.go
@internal/testmock/handlers/servers.go
@internal/client/servers_test.go

<interfaces>
<!-- Key types the executor needs -->

From internal/client/models_common.go:
```go
type NamedReference struct {
    Name string `json:"name,omitempty"`
    ID   string `json:"id,omitempty"`
}
```

Current (WRONG) ServerDNS from models_exports.go:
```go
type ServerDNS struct {
    Domain      string   `json:"domain,omitempty"`
    Nameservers []string `json:"nameservers,omitempty"`
    Services    []string `json:"services,omitempty"`
}
```

Real API response format:
```json
{
  "dns": [{"name": "management", "id": "343c...", "resource_type": "dns"}],
  "directory_services": [{"name": "srv-backup_nfs", "id": "c903...", "resource_type": "directory-services"}]
}
```
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Update client models, mock handler, and client tests</name>
  <files>internal/client/models_exports.go, internal/testmock/handlers/servers.go, internal/client/servers_test.go</files>
  <action>
1. **models_exports.go**: Delete `ServerDNS` struct entirely. Update `Server` struct:
   - `DNS` field: change type from `[]ServerDNS` to `[]NamedReference`
   - Add `DirectoryServices []NamedReference \`json:"directory_services,omitempty"\`` field
   - Update `ServerPost.DNS` from `[]ServerDNS` to `[]NamedReference`
   - Update `ServerPatch.DNS` from `[]ServerDNS` to `[]NamedReference`

2. **handlers/servers.go** (`AddServer` method): Update seeded server DNS from:
   ```go
   DNS: []client.ServerDNS{{Domain: "test.local", Nameservers: []string{"10.0.0.1"}}}
   ```
   to:
   ```go
   DNS: []client.NamedReference{{Name: "management"}},
   DirectoryServices: []client.NamedReference{{Name: "srv-backup_nfs"}},
   ```
   Also update `handlePost` -- the `srv` creation already copies `body.DNS` which will now be `[]NamedReference`. Update `handlePatch` dns unmarshal from `[]client.ServerDNS` to `[]client.NamedReference`. Add `DirectoryServices` to the created server in handlePost (set to empty `[]client.NamedReference{}`).

3. **servers_test.go**: Update `TestUnit_Server_Post` -- change `ServerPost.DNS` from `[]client.ServerDNS{{Domain: "example.com", Nameservers: []string{"8.8.8.8"}}}` to `[]client.NamedReference{{Name: "management"}}`. Update assertion from checking `DNS[0].Domain` to checking `DNS[0].Name`. Same for `TestUnit_Server_Patch` -- change `ServerPatch.DNS` from `[]client.ServerDNS{{Domain: "updated.example.com"}}` to `[]client.NamedReference{{Name: "updated-dns"}}` and update assertion.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade && go build ./internal/client/... && go test ./internal/client/... -run TestUnit_Server -count=1 -v</automated>
  </verify>
  <done>ServerDNS struct deleted, Server/ServerPost/ServerPatch use []NamedReference for DNS, DirectoryServices added, mock handler updated, client tests pass</done>
</task>

<task type="auto">
  <name>Task 2: Refactor resource, data source, and all provider tests</name>
  <files>internal/provider/server_resource.go, internal/provider/server_data_source.go, internal/provider/server_resource_test.go, internal/provider/server_data_source_test.go</files>
  <action>
**server_resource.go:**

1. **Delete** `serverDNSModel` struct, `serverDNSAttrTypes()`, `serverDNSObjectType()`, `mapModelDNSToClient()` -- all obsolete.

2. **Update `serverResourceModel`**: `DNS` stays `types.List` but will now hold `types.StringType` elements (list of DNS config names). Add `DirectoryServices types.List \`tfsdk:"directory_services"\``.

3. **Update Schema** (version 1 -> 2):
   - `Version: 2`
   - Replace `dns` from `ListNestedAttribute` to `schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "List of DNS configuration names associated with this server.", PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}}`.
   - Add `directory_services`: `schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of directory service names associated with this server.", PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}}`.

4. **Add v1->v2 StateUpgrader**: Create `serverV1StateModel` (copy of old serverResourceModel with old DNS as ListNested + network_interfaces, no directory_services). The v1 PriorSchema must match current v1 schema exactly (ListNestedAttribute dns with domain/nameservers/services, cascade_delete, network_interfaces, timeouts). The upgrader reads old state, converts DNS list-of-objects to list-of-strings by extracting... wait -- v1 DNS was inline objects which don't have a "name" field. The v1->v2 upgrader should set DNS to `types.ListNull(types.StringType)` (data will be refreshed on next Read) and set DirectoryServices to `types.ListNull(types.StringType)`. Keep existing v0->v1 upgrader but update its PriorSchema to keep the old nested DNS + update its output to also include `DirectoryServices: types.ListNull(types.StringType)` in the new state model AND update the DNS to `types.ListNull(types.StringType)` (since the executor target state is now v2 with flat string DNS). Actually, the v0 upgrader should output the v1 state format, not v2. So keep v0->v1 as-is (it outputs the v1 model format). Then add v1->v2 which reads v1 state (with nested DNS + network_interfaces) and outputs v2 state (flat DNS null + directory_services null + network_interfaces preserved). The framework chains upgraders automatically.

   Actually, re-examine: the v0->v1 upgrader currently writes to `serverResourceModel` which after our change will have flat DNS + DirectoryServices. This will break. Fix: The v0->v1 upgrader must output v1 state format. Create a dedicated `serverV1OutputModel` struct for v0->v1 output that has the OLD nested DNS format. OR simpler: update v0->v1 upgrader to output the FINAL v2 format directly (skipping v1), setting DNS and DirectoryServices to null. The Terraform framework calls ALL upgraders in sequence (0->1->2), BUT only the upgrader matching the state version runs. So a state at v0 runs upgrader[0] then upgrader[1]. The issue is upgrader[0] must output v1 format and upgrader[1] reads v1 and outputs v2.

   **Simplest approach**: Update the v0 upgrader to output a model compatible with the v1 PriorSchema (keep nested DNS format in its output). Then the v1->v2 upgrader converts from nested DNS to flat. For v0->v1 output, use `serverV0ToV1Model` that has the OLD dns field format (nested objects) + network_interfaces. For v1->v2, use `serverV1StateModel` (same shape as old serverResourceModel -- nested DNS + network_interfaces + cascade_delete + timeouts) as input, and output the new `serverResourceModel` (flat DNS + DirectoryServices + network_interfaces).

   In practice, the simplest approach: keep the v0 PriorSchema as-is. Change v0 upgrader output to write into v1 format. Add v1 PriorSchema = the current v1 schema (nested DNS, network_interfaces, cascade_delete, timeouts, version 1). v1 upgrader reads this, outputs v2 model with DNS=null, DirectoryServices=null, NetworkInterfaces preserved, CascadeDelete preserved, Timeouts preserved.

   For the v0 upgrader, its current output uses `serverResourceModel` which is about to change shape. Fix: make v0 output use a raw `resp.State.SetAttribute` approach or define an intermediate struct. Easiest: define `serverV1Model` struct with the old nested DNS types.List + NetworkInterfaces types.List + the rest, use that for v0->v1 output.

   **Implementation steps for state upgraders:**
   a. Rename existing `serverV0StateModel` as-is (it's the v0 input).
   b. Create `serverV1StateModel` struct: same fields as old `serverResourceModel` (ID, Name, Created, DNS as types.List of nested objects, CascadeDelete, NetworkInterfaces, Timeouts). This is both the OUTPUT of v0->v1 and the INPUT of v1->v2.
   c. Update v0->v1 upgrader: read `serverV0StateModel`, write `serverV1StateModel` (setting NetworkInterfaces to empty list, DNS carried over as-is).
   d. Create v1->v2 upgrader: PriorSchema = current v1 schema (nested DNS, network_interfaces, etc). Read `serverV1StateModel`. Write new `serverResourceModel` with `DNS: types.ListNull(types.StringType)`, `DirectoryServices: types.ListNull(types.StringType)`, `NetworkInterfaces` preserved, `CascadeDelete` preserved, `Timeouts` preserved.
   e. The v1 PriorSchema must include ALL v1 attributes: id, name, created, dns (ListNested with domain/nameservers/services), cascade_delete, network_interfaces, timeouts.

5. **Simplify `mapServerToModel`**: Replace the 50-line DNS mapping block with:
   ```go
   if len(srv.DNS) > 0 {
       names := make([]string, len(srv.DNS))
       for i, d := range srv.DNS {
           names[i] = d.Name
       }
       dnsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
       diags.Append(listDiags...)
       if diags.HasError() { return }
       data.DNS = dnsList
   } else {
       data.DNS = types.ListNull(types.StringType)
   }
   // Map directory_services
   if len(srv.DirectoryServices) > 0 {
       names := make([]string, len(srv.DirectoryServices))
       for i, ds := range srv.DirectoryServices {
           names[i] = ds.Name
       }
       dsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
       diags.Append(listDiags...)
       if diags.HasError() { return }
       data.DirectoryServices = dsList
   } else {
       data.DirectoryServices = types.ListNull(types.StringType)
   }
   ```

6. **Replace `mapModelDNSToClient`** with a simple helper `dnsNamesToRefs`:
   ```go
   func dnsNamesToRefs(ctx context.Context, data *serverResourceModel, diags *diag.Diagnostics) []client.NamedReference {
       if data.DNS.IsNull() || data.DNS.IsUnknown() || len(data.DNS.Elements()) == 0 {
           return nil
       }
       var names []string
       diags.Append(data.DNS.ElementsAs(ctx, &names, false)...)
       if diags.HasError() { return nil }
       refs := make([]client.NamedReference, len(names))
       for i, n := range names {
           refs[i] = client.NamedReference{Name: n}
       }
       return refs
   }
   ```
   Update Create and Update to call `dnsNamesToRefs` instead of `mapModelDNSToClient`.

7. **Update ImportState**: Initialize `DirectoryServices` to `types.ListNull(types.StringType)`.

**server_data_source.go:**

1. **Update `serverDataSourceModel`**: Add `DirectoryServices types.List \`tfsdk:"directory_services"\``.

2. **Update Schema**: Replace dns `ListNestedAttribute` with `schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of DNS configuration names associated with this server."}`. Add `directory_services`: `schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of directory service names associated with this server."}`.

3. **Delete `mapServerDNSToDataSourceModel`**. Replace its call in Read with inline mapping (same pattern as resource):
   ```go
   // Map DNS names
   if len(srv.DNS) > 0 {
       names := make([]string, len(srv.DNS))
       for i, d := range srv.DNS { names[i] = d.Name }
       dnsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
       resp.Diagnostics.Append(listDiags...)
       if resp.Diagnostics.HasError() { return }
       config.DNS = dnsList
   } else {
       config.DNS = types.ListNull(types.StringType)
   }
   // Map directory_services names
   if len(srv.DirectoryServices) > 0 {
       names := make([]string, len(srv.DirectoryServices))
       for i, ds := range srv.DirectoryServices { names[i] = ds.Name }
       dsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
       resp.Diagnostics.Append(listDiags...)
       if resp.Diagnostics.HasError() { return }
       config.DirectoryServices = dsList
   } else {
       config.DirectoryServices = types.ListNull(types.StringType)
   }
   ```

**server_resource_test.go:**

1. **Update `buildServerType()`**: Remove `dnsType` object. Change `"dns"` from `tftypes.List{ElementType: dnsType}` to `tftypes.List{ElementType: tftypes.String}`. Add `"directory_services": tftypes.List{ElementType: tftypes.String}`.

2. **Update `nullServerConfig()`**: Same changes -- dns becomes `tftypes.List{ElementType: tftypes.String}` with nil value. Add `"directory_services"` with nil value.

3. **Delete `serverPlanWithDNS()`**. Replace with a simpler helper that takes DNS names as `[]string`:
   ```go
   func serverPlanWithDNS(t *testing.T, name string, dnsNames []string) tfsdk.Plan {
       t.Helper()
       s := serverResourceSchema(t).Schema
       cfg := nullServerConfig()
       cfg["name"] = tftypes.NewValue(tftypes.String, name)
       if dnsNames != nil {
           vals := make([]tftypes.Value, len(dnsNames))
           for i, n := range dnsNames { vals[i] = tftypes.NewValue(tftypes.String, n) }
           cfg["dns"] = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, vals)
       }
       return tfsdk.Plan{Raw: tftypes.NewValue(buildServerType(), cfg), Schema: s}
   }
   ```

4. **Update all test call sites**:
   - `TestUnit_Server_Create`: `serverPlanWithDNS(t, "test-server", []string{"management"})`. Verify DNS is not null after create.
   - `TestUnit_Server_Update`: Create with `serverPlanWithDNS(t, "update-server", []string{"management"})`, update with `serverPlanWithDNS(t, "update-server", []string{"updated-dns"})`. Verify DNS after update by extracting `[]string` (not `[]serverDNSModel`). Check `dnsNames[0] == "updated-dns"`.
   - `TestUnit_Server_Import`: `serverPlanWithDNS(t, "import-server", []string{"management"})`.
   - `TestUnit_Server_SchemaVersion`: Change expected version from 1 to 2.

5. **Update `TestUnit_Server_StateUpgradeV0ToV1`**: This test verifies v0->v1 upgrade. Keep it mostly as-is but it now needs to verify the intermediate v1 state. OR: rename to `TestUnit_Server_StateUpgradeV0ToV2` and chain both upgraders. Simplest: test v0->v1 and v1->v2 separately. Add `TestUnit_Server_StateUpgradeV1ToV2` that:
   - Builds a v1 raw state with nested DNS objects + network_interfaces
   - Runs v1->v2 upgrader
   - Verifies DNS is null (types.ListNull of StringType), DirectoryServices is null, NetworkInterfaces preserved
   Update existing v0->v1 test to verify output is v1 format (nested DNS preserved, network_interfaces added).

**server_data_source_test.go:**

1. **Update `buildServerDSType()`**: Remove `dnsType` object. Change `"dns"` to `tftypes.List{ElementType: tftypes.String}`. Add `"directory_services": tftypes.List{ElementType: tftypes.String}`.

2. **Update `nullServerDSConfig()`**: Same -- dns becomes string list. Add `"directory_services"`.

3. **Update test assertions**: DNS is now a string list, not nested objects. Verify `dns` is not null and contains expected names (e.g., `"management"` from mock). Add assertion for `directory_services` not null.
  </action>
  <verify>
    <automated>cd /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade && go build ./... && go test ./internal/provider/... -run "TestUnit_Server" -count=1 -v && go test ./internal/client/... -run "TestUnit_Server" -count=1 -v</automated>
  </verify>
  <done>Schema version is 2. dns is a flat list of strings. directory_services is computed list of strings. v0->v1 and v1->v2 state upgraders work. All server resource/data source/client tests pass. No lint errors.</done>
</task>

</tasks>

<verification>
```bash
cd /home/gule/Workspace/team-infrastructure/terraform-provider-flashblade
go build ./...
go vet ./...
go test ./... -count=1
```
All must pass with zero failures.
</verification>

<success_criteria>
- `ServerDNS` struct deleted from models_exports.go
- `Server.DNS` is `[]NamedReference`, `Server.DirectoryServices` is `[]NamedReference`
- Resource schema version is 2
- v0->v1 and v1->v2 state upgraders exist and work
- dns is `ListAttribute` with `ElementType: types.StringType` in both resource and data source
- directory_services is computed `ListAttribute` with `ElementType: types.StringType` in both
- `go test ./... -count=1` passes
</success_criteria>

<output>
After completion, create `.planning/quick/7-refactor-server-dns-to-namedreference-ad/7-SUMMARY.md`
</output>
