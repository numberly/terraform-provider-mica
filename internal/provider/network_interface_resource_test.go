package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestNetworkInterfaceResource creates a networkInterfaceResource wired to the given mock server.
func newTestNetworkInterfaceResource(t *testing.T, ms *testmock.MockServer) *networkInterfaceResource {
	t.Helper()
	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return &networkInterfaceResource{client: c}
}

// niResourceSchema returns the parsed schema for the network interface resource.
func niResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &networkInterfaceResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildNIType returns the tftypes.Object for the network interface resource schema.
func buildNIType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":               tftypes.String,
		"name":             tftypes.String,
		"address":          tftypes.String,
		"subnet_name":      tftypes.String,
		"type":             tftypes.String,
		"services":         tftypes.String,
		"attached_servers": tftypes.List{ElementType: tftypes.String},
		"enabled":          tftypes.Bool,
		"gateway":          tftypes.String,
		"mtu":              tftypes.Number,
		"netmask":          tftypes.String,
		"vlan":             tftypes.Number,
		"realms":           tftypes.List{ElementType: tftypes.String},
		"timeouts":         timeoutsType,
	}}
}

// nullNIConfig returns a base config map with all resource attributes null.
func nullNIConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":               tftypes.NewValue(tftypes.String, nil),
		"name":             tftypes.NewValue(tftypes.String, nil),
		"address":          tftypes.NewValue(tftypes.String, nil),
		"subnet_name":      tftypes.NewValue(tftypes.String, nil),
		"type":             tftypes.NewValue(tftypes.String, nil),
		"services":         tftypes.NewValue(tftypes.String, nil),
		"attached_servers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"enabled":          tftypes.NewValue(tftypes.Bool, nil),
		"gateway":          tftypes.NewValue(tftypes.String, nil),
		"mtu":              tftypes.NewValue(tftypes.Number, nil),
		"netmask":          tftypes.NewValue(tftypes.String, nil),
		"vlan":             tftypes.NewValue(tftypes.Number, nil),
		"realms":           tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"timeouts":         tftypes.NewValue(timeoutsType, nil),
	}
}

// niPlanWith returns a tfsdk.Plan for the network interface resource.
func niPlanWith(t *testing.T, fields map[string]tftypes.Value) tfsdk.Plan {
	t.Helper()
	s := niResourceSchema(t).Schema
	cfg := nullNIConfig()
	for k, v := range fields {
		cfg[k] = v
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildNIType(), cfg),
		Schema: s,
	}
}

// serverListValue returns a tftypes.Value for a list of server names.
func serverListValue(names ...string) tftypes.Value {
	vals := make([]tftypes.Value, 0, len(names))
	for _, n := range names {
		vals = append(vals, tftypes.NewValue(tftypes.String, n))
	}
	return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, vals)
}

// ---- tests ------------------------------------------------------------------

// TestUnit_NetworkInterface_Create verifies Create populates all attributes from API response.
func TestUnit_NetworkInterface_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestNetworkInterfaceResource(t, ms)
	s := niResourceSchema(t).Schema

	plan := niPlanWith(t, map[string]tftypes.Value{
		"name":             tftypes.NewValue(tftypes.String, "vip0"),
		"address":          tftypes.NewValue(tftypes.String, "10.21.200.10"),
		"subnet_name":      tftypes.NewValue(tftypes.String, "data-subnet"),
		"type":             tftypes.NewValue(tftypes.String, "vip"),
		"services":         tftypes.NewValue(tftypes.String, "data"),
		"attached_servers": serverListValue("server1"),
	})

	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNIType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model networkInterfaceResourceModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "vip0" {
		t.Errorf("expected name=vip0, got %s", model.Name.ValueString())
	}
	if model.Address.ValueString() != "10.21.200.10" {
		t.Errorf("expected address=10.21.200.10, got %s", model.Address.ValueString())
	}
	if model.SubnetName.ValueString() != "data-subnet" {
		t.Errorf("expected subnet_name=data-subnet, got %s", model.SubnetName.ValueString())
	}
	if model.Type.ValueString() != "vip" {
		t.Errorf("expected type=vip, got %s", model.Type.ValueString())
	}
	if model.Services.ValueString() != "data" {
		t.Errorf("expected services=data, got %s", model.Services.ValueString())
	}
	// Computed fields should be populated.
	if model.Enabled.IsNull() {
		t.Error("expected enabled to be populated")
	}
	if model.MTU.IsNull() {
		t.Error("expected mtu to be populated")
	}
	if model.Netmask.IsNull() {
		t.Error("expected netmask to be populated")
	}
	// attached_servers should be a list with one entry.
	if model.AttachedServers.IsNull() {
		t.Error("expected attached_servers to be populated (not null)")
	}
}

// TestUnit_NetworkInterface_Update verifies Update applies PATCH and updates state.
func TestUnit_NetworkInterface_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestNetworkInterfaceResource(t, ms)
	s := niResourceSchema(t).Schema

	// Create first.
	createPlan := niPlanWith(t, map[string]tftypes.Value{
		"name":             tftypes.NewValue(tftypes.String, "update-vip"),
		"address":          tftypes.NewValue(tftypes.String, "10.0.0.10"),
		"subnet_name":      tftypes.NewValue(tftypes.String, "data-subnet"),
		"type":             tftypes.NewValue(tftypes.String, "vip"),
		"services":         tftypes.NewValue(tftypes.String, "data"),
		"attached_servers": serverListValue("server1"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNIType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update address and attached_servers.
	updatePlan := niPlanWith(t, map[string]tftypes.Value{
		"name":             tftypes.NewValue(tftypes.String, "update-vip"),
		"address":          tftypes.NewValue(tftypes.String, "10.0.0.20"),
		"subnet_name":      tftypes.NewValue(tftypes.String, "data-subnet"),
		"type":             tftypes.NewValue(tftypes.String, "vip"),
		"services":         tftypes.NewValue(tftypes.String, "data"),
		"attached_servers": serverListValue("server1", "server2"),
	})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNIType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model networkInterfaceResourceModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Address.ValueString() != "10.0.0.20" {
		t.Errorf("expected address=10.0.0.20 after update, got %s", model.Address.ValueString())
	}
	// Verify attached_servers list length.
	var serverNames []string
	if diags := model.AttachedServers.ElementsAs(context.Background(), &serverNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(serverNames) != 2 {
		t.Errorf("expected 2 attached_servers after update, got %d", len(serverNames))
	}
}

// TestUnit_NetworkInterface_Delete verifies Delete removes the network interface.
func TestUnit_NetworkInterface_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestNetworkInterfaceResource(t, ms)
	s := niResourceSchema(t).Schema

	plan := niPlanWith(t, map[string]tftypes.Value{
		"name":             tftypes.NewValue(tftypes.String, "delete-vip"),
		"address":          tftypes.NewValue(tftypes.String, "10.0.0.5"),
		"subnet_name":      tftypes.NewValue(tftypes.String, "data-subnet"),
		"type":             tftypes.NewValue(tftypes.String, "vip"),
		"services":         tftypes.NewValue(tftypes.String, "egress-only"),
		"attached_servers": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNIType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: createResp.State}, deleteResp)

	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify NI is gone.
	_, err := r.client.GetNetworkInterface(context.Background(), "delete-vip")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected network interface to be deleted, got: %v", err)
	}
}

// TestUnit_NetworkInterface_Schema verifies schema attributes and modifiers.
func TestUnit_NetworkInterface_Schema(t *testing.T) {
	resp := niResourceSchema(t)
	s := resp.Schema

	// Required attributes (including RequiresReplace candidates).
	for _, attrName := range []string{"name", "subnet_name", "type", "services", "address"} {
		attr, ok := s.Attributes[attrName]
		if !ok {
			t.Fatalf("attribute %q not found in schema", attrName)
		}
		if !attr.IsRequired() {
			t.Errorf("attribute %q should be Required", attrName)
		}
	}

	// Computed-only attributes.
	for _, attrName := range []string{"id", "enabled", "gateway", "mtu", "netmask", "vlan", "realms"} {
		attr, ok := s.Attributes[attrName]
		if !ok {
			t.Fatalf("attribute %q not found in schema", attrName)
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %q should be Computed", attrName)
		}
	}

	// attached_servers should be Optional+Computed.
	asAttr, ok := s.Attributes["attached_servers"]
	if !ok {
		t.Fatal("attached_servers attribute not found in schema")
	}
	if !asAttr.IsOptional() {
		t.Error("attached_servers should be Optional")
	}
	if !asAttr.IsComputed() {
		t.Error("attached_servers should be Computed")
	}
}

// TestUnit_NetworkInterface_ServicesValidator verifies the services enum validator.
func TestUnit_NetworkInterface_ServicesValidator(t *testing.T) {
	v := serviceTypeValidator()

	// Valid values must pass.
	validValues := []string{"data", "sts", "egress-only", "replication"}
	for _, val := range validValues {
		var resp validator.StringResponse
		v.ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue(val),
		}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("expected %q to pass services validator, got error: %s", val, resp.Diagnostics)
		}
	}

	// Invalid value must fail.
	var resp validator.StringResponse
	v.ValidateString(context.Background(), validator.StringRequest{
		ConfigValue: types.StringValue("invalid-svc"),
	}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected 'invalid-svc' to fail services validator, but it passed")
	}
}

// TestUnit_NetworkInterface_ConfigValidator verifies the cross-field services x attached_servers validator.
func TestUnit_NetworkInterface_ConfigValidator(t *testing.T) {
	cases := []struct {
		name        string
		services    string
		servers     []string // nil = null
		expectError bool
	}{
		{name: "data+one_server=OK", services: "data", servers: []string{"server1"}, expectError: false},
		{name: "data+no_servers=ERROR", services: "data", servers: nil, expectError: true},
		{name: "sts+one_server=OK", services: "sts", servers: []string{"server1"}, expectError: false},
		{name: "sts+no_servers=ERROR", services: "sts", servers: nil, expectError: true},
		{name: "egress-only+no_servers=OK", services: "egress-only", servers: nil, expectError: false},
		{name: "egress-only+servers=ERROR", services: "egress-only", servers: []string{"server1"}, expectError: true},
		{name: "replication+no_servers=OK", services: "replication", servers: nil, expectError: false},
		{name: "replication+servers=ERROR", services: "replication", servers: []string{"server1"}, expectError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &networkInterfaceResource{}
			validators := r.ConfigValidators(context.Background())
			if len(validators) == 0 {
				t.Fatal("expected at least one ConfigValidator")
			}

			s := niResourceSchema(t).Schema

			var serverVal tftypes.Value
			if len(tc.servers) > 0 {
				serverVal = serverListValue(tc.servers...)
			} else {
				serverVal = tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil)
			}

			cfg := nullNIConfig()
			cfg["services"] = tftypes.NewValue(tftypes.String, tc.services)
			cfg["attached_servers"] = serverVal
			config := tfsdk.Config{
				Raw:    tftypes.NewValue(buildNIType(), cfg),
				Schema: s,
			}

			var validateResp resource.ValidateConfigResponse
			for _, v := range validators {
				v.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: config}, &validateResp)
			}

			if tc.expectError && !validateResp.Diagnostics.HasError() {
				t.Errorf("expected validation error for %s+%v, got none", tc.services, tc.servers)
			}
			if !tc.expectError && validateResp.Diagnostics.HasError() {
				t.Errorf("expected no validation error for %s+%v, got: %s", tc.services, tc.servers, validateResp.Diagnostics)
			}
		})
	}
}

// TestUnit_NetworkInterface_Import verifies ImportState populates all attributes.
func TestUnit_NetworkInterface_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	// Seed directly in mock store.
	store.AddNetworkInterface("import-vip", "10.21.200.50", "data-subnet", "vip", "data")

	r := newTestNetworkInterfaceResource(t, ms)
	s := niResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNIType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-vip"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model networkInterfaceResourceModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-vip" {
		t.Errorf("expected name=import-vip after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Address.ValueString() != "10.21.200.50" {
		t.Errorf("expected address=10.21.200.50 after import, got %s", model.Address.ValueString())
	}
	if model.SubnetName.ValueString() != "data-subnet" {
		t.Errorf("expected subnet_name=data-subnet after import, got %s", model.SubnetName.ValueString())
	}
	if model.Services.ValueString() != "data" {
		t.Errorf("expected services=data after import, got %s", model.Services.ValueString())
	}
}

// TestUnit_NetworkInterface_Drift verifies Read detects changes and updates state.
func TestUnit_NetworkInterface_Drift(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestNetworkInterfaceResource(t, ms)
	s := niResourceSchema(t).Schema

	// Create via resource.
	createPlan := niPlanWith(t, map[string]tftypes.Value{
		"name":             tftypes.NewValue(tftypes.String, "drift-vip"),
		"address":          tftypes.NewValue(tftypes.String, "10.30.0.10"),
		"subnet_name":      tftypes.NewValue(tftypes.String, "data-subnet"),
		"type":             tftypes.NewValue(tftypes.String, "vip"),
		"services":         tftypes.NewValue(tftypes.String, "data"),
		"attached_servers": serverListValue("server1"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNIType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Simulate external change: PATCH address directly.
	newAddr := "10.30.0.99"
	_, err := r.client.PatchNetworkInterface(context.Background(), "drift-vip", client.NetworkInterfacePatch{
		Address:         &newAddr,
		Services:        []string{"data"},
		AttachedServers: []client.NamedReference{{Name: "server1"}},
	})
	if err != nil {
		t.Fatalf("PatchNetworkInterface (external change simulation): %v", err)
	}

	// Read should reflect the drifted state.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model networkInterfaceResourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// State should now reflect the drifted address.
	if model.Address.ValueString() != "10.30.0.99" {
		t.Errorf("expected address=10.30.0.99 after drift Read, got %s", model.Address.ValueString())
	}
}

// TestUnit_NetworkInterface_NotFound verifies Read removes resource from state on 404.
func TestUnit_NetworkInterface_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	r := newTestNetworkInterfaceResource(t, ms)
	s := niResourceSchema(t).Schema

	cfg := nullNIConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-vip")
	cfg["id"] = tftypes.NewValue(tftypes.String, "some-id")
	state := tfsdk.State{Raw: tftypes.NewValue(buildNIType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when network interface not found")
	}
}

// TestUnit_NetworkInterface_AttachedServersEmptyList verifies that when the API
// returns no attached_servers, the state uses an empty list (not null) to prevent drift.
func TestUnit_NetworkInterface_AttachedServersEmptyList(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	// Seed an NI with no attached_servers (egress-only).
	store.AddNetworkInterface("egress-vip", "10.99.0.1", "egress-subnet", "vip", "egress-only")

	r := newTestNetworkInterfaceResource(t, ms)
	s := niResourceSchema(t).Schema

	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildNIType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "egress-vip"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model networkInterfaceResourceModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// attached_servers should be an empty list (not null) — prevents spurious drift.
	if model.AttachedServers.IsNull() {
		t.Error("expected attached_servers to be an empty list (not null) when API returns no servers")
	}
	var names []string
	if diags := model.AttachedServers.ElementsAs(context.Background(), &names, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 attached_servers, got %d", len(names))
	}
}
