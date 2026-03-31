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

// newTestSubnetDataSource creates a subnetDataSource wired to the given mock server.
func newTestSubnetDataSource(t *testing.T, ms *testmock.MockServer) *subnetDataSource {
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
	return &subnetDataSource{client: c}
}

// subnetDSSchema returns the schema for the subnet data source.
func subnetDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &subnetDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildSubnetDSType returns the tftypes.Object for the subnet data source schema.
func buildSubnetDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":         tftypes.String,
		"name":       tftypes.String,
		"prefix":     tftypes.String,
		"gateway":    tftypes.String,
		"mtu":        tftypes.Number,
		"vlan":       tftypes.Number,
		"lag_name":   tftypes.String,
		"enabled":    tftypes.Bool,
		"services":   tftypes.List{ElementType: tftypes.String},
		"interfaces": tftypes.List{ElementType: tftypes.String},
	}}
}

// nullSubnetDSConfig returns a base config map with all attributes null.
func nullSubnetDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, nil),
		"name":       tftypes.NewValue(tftypes.String, nil),
		"prefix":     tftypes.NewValue(tftypes.String, nil),
		"gateway":    tftypes.NewValue(tftypes.String, nil),
		"mtu":        tftypes.NewValue(tftypes.Number, nil),
		"vlan":       tftypes.NewValue(tftypes.Number, nil),
		"lag_name":   tftypes.NewValue(tftypes.String, nil),
		"enabled":    tftypes.NewValue(tftypes.Bool, nil),
		"services":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"interfaces": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_SubnetDataSource_Read verifies the data source reads a seeded subnet by name.
func TestUnit_SubnetDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterSubnetHandlers(ms.Mux)
	store.AddSubnet("ds-subnet", "192.168.1.0/24", "lag0")

	d := newTestSubnetDataSource(t, ms)
	s := subnetDSSchema(t).Schema

	cfg := nullSubnetDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-subnet")

	objType := buildSubnetDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model subnetDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "ds-subnet" {
		t.Errorf("expected name=ds-subnet, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id")
	}
	if model.Prefix.ValueString() != "192.168.1.0/24" {
		t.Errorf("expected prefix=192.168.1.0/24, got %s", model.Prefix.ValueString())
	}
	if model.LagName.ValueString() != "lag0" {
		t.Errorf("expected lag_name=lag0, got %s", model.LagName.ValueString())
	}
	if model.MTU.ValueInt64() != 1500 {
		t.Errorf("expected mtu=1500 (AddSubnet default), got %d", model.MTU.ValueInt64())
	}
	if model.Enabled.IsNull() {
		t.Error("expected enabled to be populated")
	}
}

// TestUnit_SubnetDataSource_NotFound verifies error diagnostic when subnet not found.
func TestUnit_SubnetDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	d := newTestSubnetDataSource(t, ms)
	s := subnetDSSchema(t).Schema

	cfg := nullSubnetDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent")

	objType := buildSubnetDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found subnet, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Subnet not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Subnet not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}

// TestUnit_SubnetDataSource_Schema verifies schema: name is Required, others Computed.
func TestUnit_SubnetDataSource_Schema(t *testing.T) {
	resp := subnetDSSchema(t)
	s := resp.Schema

	attr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}
	if !attr.IsRequired() {
		t.Error("name should be Required")
	}

	computedAttrs := []string{"id", "prefix", "gateway", "mtu", "vlan", "lag_name", "enabled", "services", "interfaces"}
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
