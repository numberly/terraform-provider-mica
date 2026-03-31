package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestSubnetResource creates a subnetResource wired to the given mock server.
func newTestSubnetResource(t *testing.T, ms *testmock.MockServer) *subnetResource {
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
	return &subnetResource{client: c}
}

// subnetResourceSchema returns the parsed schema for the subnet resource.
func subnetResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &subnetResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildSubnetType returns the tftypes.Object for the subnet resource schema.
func buildSubnetType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
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
		"timeouts":   timeoutsType,
	}}
}

// nullSubnetConfig returns a base config map with all resource attributes null.
func nullSubnetConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
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
		"timeouts":   tftypes.NewValue(timeoutsType, nil),
	}
}

// subnetPlanWith returns a tfsdk.Plan for the subnet resource.
func subnetPlanWith(t *testing.T, fields map[string]tftypes.Value) tfsdk.Plan {
	t.Helper()
	s := subnetResourceSchema(t).Schema
	cfg := nullSubnetConfig()
	for k, v := range fields {
		cfg[k] = v
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildSubnetType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_SubnetResource_Create verifies Create populates all attributes from API response.
func TestUnit_SubnetResource_Create(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	lagStore := handlers.RegisterLinkAggregationGroupHandlers(ms.Mux)
	lagStore.Seed(&client.LinkAggregationGroup{
		ID:   "lag-001",
		Name: "lag0",
	})
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":     tftypes.NewValue(tftypes.String, "test-subnet"),
		"prefix":   tftypes.NewValue(tftypes.String, "10.21.200.0/24"),
		"gateway":  tftypes.NewValue(tftypes.String, "10.21.200.1"),
		"mtu":      tftypes.NewValue(tftypes.Number, 9000),
		"vlan":     tftypes.NewValue(tftypes.Number, 100),
		"lag_name": tftypes.NewValue(tftypes.String, "lag0"),
	})

	resp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", resp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := resp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty ID after Create")
	}
	if model.Name.ValueString() != "test-subnet" {
		t.Errorf("expected name=test-subnet, got %s", model.Name.ValueString())
	}
	if model.Prefix.ValueString() != "10.21.200.0/24" {
		t.Errorf("expected prefix=10.21.200.0/24, got %s", model.Prefix.ValueString())
	}
	if model.Gateway.ValueString() != "10.21.200.1" {
		t.Errorf("expected gateway=10.21.200.1, got %s", model.Gateway.ValueString())
	}
	if model.MTU.ValueInt64() != 9000 {
		t.Errorf("expected mtu=9000, got %d", model.MTU.ValueInt64())
	}
	if model.VLAN.ValueInt64() != 100 {
		t.Errorf("expected vlan=100, got %d", model.VLAN.ValueInt64())
	}
	if model.LagName.ValueString() != "lag0" {
		t.Errorf("expected lag_name=lag0, got %s", model.LagName.ValueString())
	}
	if model.Enabled.IsNull() {
		t.Error("expected enabled to be populated")
	}
}

// TestUnit_SubnetResource_Update verifies Update applies partial PATCH and updates state.
func TestUnit_SubnetResource_Update(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	// Create first.
	createPlan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":    tftypes.NewValue(tftypes.String, "update-subnet"),
		"prefix":  tftypes.NewValue(tftypes.String, "10.0.0.0/24"),
		"gateway": tftypes.NewValue(tftypes.String, "10.0.0.1"),
		"mtu":     tftypes.NewValue(tftypes.Number, 1500),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: createPlan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Update gateway and mtu.
	updatePlan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":    tftypes.NewValue(tftypes.String, "update-subnet"),
		"prefix":  tftypes.NewValue(tftypes.String, "10.0.0.0/24"),
		"gateway": tftypes.NewValue(tftypes.String, "10.0.0.254"),
		"mtu":     tftypes.NewValue(tftypes.Number, 9000),
	})
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: createResp.State,
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := updateResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Gateway.ValueString() != "10.0.0.254" {
		t.Errorf("expected gateway=10.0.0.254 after update, got %s", model.Gateway.ValueString())
	}
	if model.MTU.ValueInt64() != 9000 {
		t.Errorf("expected mtu=9000 after update, got %d", model.MTU.ValueInt64())
	}
}

// TestUnit_SubnetResource_Delete verifies Delete removes the subnet.
func TestUnit_SubnetResource_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":   tftypes.NewValue(tftypes.String, "delete-subnet"),
		"prefix": tftypes.NewValue(tftypes.String, "10.10.0.0/24"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
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

	// Verify subnet is gone.
	_, err := r.client.GetSubnet(context.Background(), "delete-subnet")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected subnet to be deleted, got: %v", err)
	}
}

// TestUnit_SubnetResource_Import verifies ImportState populates all attributes.
func TestUnit_SubnetResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	// Create first.
	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":   tftypes.NewValue(tftypes.String, "import-subnet"),
		"prefix": tftypes.NewValue(tftypes.String, "10.20.0.0/24"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "import-subnet"}, importResp)

	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.Name.ValueString() != "import-subnet" {
		t.Errorf("expected name=import-subnet after import, got %s", model.Name.ValueString())
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected ID to be populated after import")
	}
	if model.Prefix.ValueString() != "10.20.0.0/24" {
		t.Errorf("expected prefix=10.20.0.0/24 after import, got %s", model.Prefix.ValueString())
	}
}

// TestUnit_SubnetResource_Drift verifies Read detects drift when subnet is modified externally.
func TestUnit_SubnetResource_Drift(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	// Create via resource.
	plan := subnetPlanWith(t, map[string]tftypes.Value{
		"name":    tftypes.NewValue(tftypes.String, "drift-subnet"),
		"prefix":  tftypes.NewValue(tftypes.String, "10.30.0.0/24"),
		"gateway": tftypes.NewValue(tftypes.String, "10.30.0.1"),
	})
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Simulate external change: modify subnet directly in mock store.
	subnet := store.AddSubnet("drift-subnet-external", "10.30.0.0/24", "")
	// Directly modify the created subnet's gateway via AddSubnet trick —
	// instead seed a modified version via the store.
	_ = subnet // ignore the new one

	// Direct modify: create a new subnet to simulate the seeded one being modified.
	// Use PatchSubnet directly on the client to simulate external change.
	newGateway := "10.30.0.254"
	_, err := r.client.PatchSubnet(context.Background(), "drift-subnet", client.SubnetPatch{
		Gateway: &newGateway,
	})
	if err != nil {
		t.Fatalf("PatchSubnet (external change simulation): %v", err)
	}

	// Read should reflect the new state (drift detected).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model subnetResourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	// State should now reflect the drifted value.
	if model.Gateway.ValueString() != "10.30.0.254" {
		t.Errorf("expected gateway=10.30.0.254 after drift Read, got %s", model.Gateway.ValueString())
	}
}

// TestUnit_SubnetResource_NotFound verifies Read removes resource from state on 404.
func TestUnit_SubnetResource_NotFound(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterSubnetHandlers(ms.Mux)

	r := newTestSubnetResource(t, ms)
	s := subnetResourceSchema(t).Schema

	cfg := nullSubnetConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "nonexistent-subnet")
	cfg["id"] = tftypes.NewValue(tftypes.String, "some-id")
	state := tfsdk.State{Raw: tftypes.NewValue(buildSubnetType(), cfg), Schema: s}

	readResp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}
	if !readResp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) when subnet not found")
	}
}
