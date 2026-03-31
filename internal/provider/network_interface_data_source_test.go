package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestNetworkInterfaceDataSource creates a networkInterfaceDataSource wired to the given mock server.
func newTestNetworkInterfaceDataSource(t *testing.T, ms *testmock.MockServer) *networkInterfaceDataSource {
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
	return &networkInterfaceDataSource{client: c}
}

// niDSSchema returns the schema for the network interface data source.
func niDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &networkInterfaceDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildNIDSType returns the tftypes.Object for the network interface data source schema.
func buildNIDSType() tftypes.Object {
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
	}}
}

// nullNIDSConfig returns a base config map with all data source attributes null.
func nullNIDSConfig() map[string]tftypes.Value {
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
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_NetworkInterfaceDataSource_Read verifies the data source reads a seeded NI by name.
func TestUnit_NetworkInterfaceDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)
	store.AddNetworkInterface("ds-vip", "192.168.1.10", "data-subnet", "vip", "data")

	d := newTestNetworkInterfaceDataSource(t, ms)
	s := niDSSchema(t).Schema

	cfg := nullNIDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-vip")

	objType := buildNIDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model networkInterfaceDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-vip" {
		t.Errorf("expected name=ds-vip, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id")
	}
	if model.Address.ValueString() != "192.168.1.10" {
		t.Errorf("expected address=192.168.1.10, got %s", model.Address.ValueString())
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
	if model.MTU.ValueInt64() != 1500 {
		t.Errorf("expected mtu=1500 (AddNetworkInterface default), got %d", model.MTU.ValueInt64())
	}
	if model.Enabled.IsNull() {
		t.Error("expected enabled to be populated")
	}
	if model.Netmask.IsNull() {
		t.Error("expected netmask to be populated")
	}
}

// TestUnit_NetworkInterfaceDataSource_NotFound verifies error diagnostic when NI not found.
func TestUnit_NetworkInterfaceDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	d := newTestNetworkInterfaceDataSource(t, ms)
	s := niDSSchema(t).Schema

	cfg := nullNIDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-vip")

	objType := buildNIDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found network interface, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Network interface not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Network interface not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}

// TestUnit_NetworkInterfaceDataSource_Schema verifies schema: name is Required, others Computed.
func TestUnit_NetworkInterfaceDataSource_Schema(t *testing.T) {
	resp := niDSSchema(t)
	s := resp.Schema

	attr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}
	if !attr.IsRequired() {
		t.Error("name should be Required")
	}

	computedAttrs := []string{"id", "address", "subnet_name", "type", "services", "attached_servers",
		"enabled", "gateway", "mtu", "netmask", "vlan", "realms"}
	for _, name := range computedAttrs {
		a, ok := s.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found in schema", name)
			continue
		}
		if !a.IsComputed() {
			t.Errorf("attribute %q should be Computed", name)
		}
	}
}

// TestUnit_NetworkInterfaceDataSource_WithServers verifies attached_servers are populated.
func TestUnit_NetworkInterfaceDataSource_WithServers(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterNetworkInterfaceHandlers(ms.Mux)

	// Seed NI.
	ni := store.AddNetworkInterface("vip-servers", "10.0.0.5", "data-subnet", "vip", "data")
	ni.AttachedServers = []client.NamedReference{{Name: "server1"}, {Name: "server2"}}

	d := newTestNetworkInterfaceDataSource(t, ms)
	s := niDSSchema(t).Schema

	cfg := nullNIDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "vip-servers")

	objType := buildNIDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model networkInterfaceDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	var serverNames []string
	if diags := model.AttachedServers.ElementsAs(context.Background(), &serverNames, false); diags.HasError() {
		t.Fatalf("ElementsAs: %s", diags)
	}
	if len(serverNames) != 2 {
		t.Errorf("expected 2 attached_servers, got %d", len(serverNames))
	}
}
