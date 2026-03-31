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

// newTestLagDataSource creates a linkAggregationGroupDataSource wired to the given mock server.
func newTestLagDataSource(t *testing.T, ms *testmock.MockServer) *linkAggregationGroupDataSource {
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
	return &linkAggregationGroupDataSource{client: c}
}

// lagDSSchema returns the schema for the LAG data source.
func lagDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &linkAggregationGroupDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildLagDSType returns the tftypes.Object for the LAG data source schema.
func buildLagDSType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"status":      tftypes.String,
		"mac_address": tftypes.String,
		"port_speed":  tftypes.Number,
		"lag_speed":   tftypes.Number,
		"ports":       tftypes.List{ElementType: tftypes.String},
	}}
}

// nullLagDSConfig returns a base config map with all attributes null.
func nullLagDSConfig() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, nil),
		"status":      tftypes.NewValue(tftypes.String, nil),
		"mac_address": tftypes.NewValue(tftypes.String, nil),
		"port_speed":  tftypes.NewValue(tftypes.Number, nil),
		"lag_speed":   tftypes.NewValue(tftypes.Number, nil),
		"ports":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_LagDataSource_Read verifies the data source reads a seeded LAG by name.
func TestUnit_LagDataSource_Read(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	lagStore := handlers.RegisterLinkAggregationGroupHandlers(ms.Mux)
	lagStore.Seed(&client.LinkAggregationGroup{
		ID:         "lag-001",
		Name:       "lag0",
		Status:     "healthy",
		MacAddress: "00:11:22:33:44:55",
		PortSpeed:  25000000000,
		LagSpeed:   50000000000,
		Ports:      []string{"eth0", "eth1"},
	})

	d := newTestLagDataSource(t, ms)
	s := lagDSSchema(t).Schema

	cfg := nullLagDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "lag0")

	objType := buildLagDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("DataSource Read returned error: %s", readResp.Diagnostics)
	}

	var model lagDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "lag0" {
		t.Errorf("expected name=lag0, got %s", model.Name.ValueString())
	}
	if model.ID.ValueString() != "lag-001" {
		t.Errorf("expected id=lag-001, got %s", model.ID.ValueString())
	}
	if model.Status.ValueString() != "healthy" {
		t.Errorf("expected status=healthy, got %s", model.Status.ValueString())
	}
	if model.MacAddress.ValueString() != "00:11:22:33:44:55" {
		t.Errorf("expected mac_address=00:11:22:33:44:55, got %s", model.MacAddress.ValueString())
	}
	if model.PortSpeed.ValueInt64() != 25000000000 {
		t.Errorf("expected port_speed=25000000000, got %d", model.PortSpeed.ValueInt64())
	}
	if model.LagSpeed.ValueInt64() != 50000000000 {
		t.Errorf("expected lag_speed=50000000000, got %d", model.LagSpeed.ValueInt64())
	}
	if model.Ports.IsNull() {
		t.Error("expected ports to be populated")
	}
	var ports []string
	if diags := model.Ports.ElementsAs(context.Background(), &ports, false); diags.HasError() {
		t.Fatalf("ElementsAs ports: %s", diags)
	}
	if len(ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(ports))
	}
	if ports[0] != "eth0" || ports[1] != "eth1" {
		t.Errorf("expected ports=[eth0, eth1], got %v", ports)
	}
}

// TestUnit_LagDataSource_NotFound verifies error diagnostic when LAG not found.
func TestUnit_LagDataSource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterLinkAggregationGroupHandlers(ms.Mux)

	d := newTestLagDataSource(t, ms)
	s := lagDSSchema(t).Schema

	cfg := nullLagDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-lag")

	objType := buildLagDSType()
	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if !readResp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for not-found LAG, got none")
	}

	found := false
	for _, diag := range readResp.Diagnostics {
		if diag.Summary() == "Link aggregation group not found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Link aggregation group not found' diagnostic, got: %s", readResp.Diagnostics)
	}
}

// TestUnit_LagDataSource_Schema verifies schema: name is Required, others Computed.
func TestUnit_LagDataSource_Schema(t *testing.T) {
	resp := lagDSSchema(t)
	s := resp.Schema

	attr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("name attribute not found in schema")
	}
	if !attr.IsRequired() {
		t.Error("name should be Required")
	}

	computedAttrs := []string{"id", "status", "mac_address", "port_speed", "lag_speed", "ports"}
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
