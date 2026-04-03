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

// newTestTargetResource creates a targetResource wired to the given mock server.
func newTestTargetResource(t *testing.T, ms *testmock.MockServer) *targetResource {
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
	return &targetResource{client: c}
}

// targetResourceSchema returns the parsed schema for the target resource.
func targetResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &targetResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// buildTargetType returns the tftypes.Object for the target resource schema.
func buildTargetType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                   tftypes.String,
		"name":                 tftypes.String,
		"address":              tftypes.String,
		"ca_certificate_group": tftypes.String,
		"status":               tftypes.String,
		"status_details":       tftypes.String,
		"timeouts":             timeoutsType,
	}}
}

// nullTargetConfig returns a base config map with all attributes null.
func nullTargetConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                   tftypes.NewValue(tftypes.String, nil),
		"name":                 tftypes.NewValue(tftypes.String, nil),
		"address":              tftypes.NewValue(tftypes.String, nil),
		"ca_certificate_group": tftypes.NewValue(tftypes.String, nil),
		"status":               tftypes.NewValue(tftypes.String, nil),
		"status_details":       tftypes.NewValue(tftypes.String, nil),
		"timeouts":             tftypes.NewValue(timeoutsType, nil),
	}
}

// targetPlanWith returns a tfsdk.Plan with the given field values.
func targetPlanWith(t *testing.T, name, address, caCertGroup string) tfsdk.Plan {
	t.Helper()
	s := targetResourceSchema(t).Schema
	cfg := nullTargetConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["address"] = tftypes.NewValue(tftypes.String, address)
	if caCertGroup != "" {
		cfg["ca_certificate_group"] = tftypes.NewValue(tftypes.String, caCertGroup)
	}
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildTargetType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_TargetResource_Lifecycle: create → verify state → update address → verify → destroy.
func TestUnit_TargetResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterTargetHandlers(ms.Mux)

	r := newTestTargetResource(t, ms)
	s := targetResourceSchema(t).Schema

	// Step 1: Create.
	plan := targetPlanWith(t, "tgt-lifecycle", "192.168.1.10", "")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTargetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate targetModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if afterCreate.Name.ValueString() != "tgt-lifecycle" {
		t.Errorf("expected name=tgt-lifecycle, got %s", afterCreate.Name.ValueString())
	}
	if afterCreate.Address.ValueString() != "192.168.1.10" {
		t.Errorf("expected address=192.168.1.10, got %s", afterCreate.Address.ValueString())
	}
	if afterCreate.Status.ValueString() != "connected" {
		t.Errorf("expected status=connected, got %s", afterCreate.Status.ValueString())
	}

	// Step 2: Read (idempotence check — plan converges).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead targetModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.Address.ValueString() != afterCreate.Address.ValueString() {
		t.Errorf("address drift on Read: create=%s read=%s", afterCreate.Address.ValueString(), afterRead.Address.ValueString())
	}

	// Step 3: Update address.
	updatePlan := targetPlanWith(t, "tgt-lifecycle", "10.0.0.99", "")
	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTargetType(), nil), Schema: s},
	}
	r.Update(context.Background(), resource.UpdateRequest{
		Plan:  updatePlan,
		State: readResp.State,
	}, updateResp)
	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update returned error: %s", updateResp.Diagnostics)
	}

	var afterUpdate targetModel
	if diags := updateResp.State.Get(context.Background(), &afterUpdate); diags.HasError() {
		t.Fatalf("Get update state: %s", diags)
	}
	if afterUpdate.Address.ValueString() != "10.0.0.99" {
		t.Errorf("expected address=10.0.0.99 after update, got %s", afterUpdate.Address.ValueString())
	}

	// Step 4: Delete.
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: updateResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}

	// Verify gone.
	_, err := r.client.GetTarget(context.Background(), "tgt-lifecycle")
	if err == nil || !client.IsNotFound(err) {
		t.Errorf("expected target to be deleted, got: %v", err)
	}
}

// TestUnit_TargetResource_Import: create → import by name → check plan converges.
func TestUnit_TargetResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterTargetHandlers(ms.Mux)

	r := newTestTargetResource(t, ms)
	s := targetResourceSchema(t).Schema

	// Create first.
	plan := targetPlanWith(t, "tgt-import", "192.168.2.20", "")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTargetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTargetType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "tgt-import"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model targetModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	if model.Name.ValueString() != "tgt-import" {
		t.Errorf("expected name=tgt-import, got %s", model.Name.ValueString())
	}
	if model.Address.ValueString() != "192.168.2.20" {
		t.Errorf("expected address=192.168.2.20, got %s", model.Address.ValueString())
	}

	// Verify Read after import gives 0 diff (idempotence).
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read after import: %s", readResp.Diagnostics)
	}

	var afterRead targetModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get post-import-read state: %s", diags)
	}
	if afterRead.Address.ValueString() != model.Address.ValueString() {
		t.Errorf("post-import Read drift on address: import=%s read=%s", model.Address.ValueString(), afterRead.Address.ValueString())
	}
}

// TestUnit_TargetResource_DriftDetection: create → Seed modified target in mock → refresh → check state updated.
func TestUnit_TargetResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterTargetHandlers(ms.Mux)

	r := newTestTargetResource(t, ms)
	s := targetResourceSchema(t).Schema

	// Create.
	plan := targetPlanWith(t, "tgt-drift", "192.168.3.30", "")
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildTargetType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate targetModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Simulate drift: modify address in the mock store directly.
	store.Seed(&client.Target{
		ID:            afterCreate.ID.ValueString(),
		Name:          "tgt-drift",
		Address:       "10.99.99.99", // changed outside Terraform
		Status:        "connected",
		StatusDetails: "drifted",
	})

	// Read to detect drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read drift detection: %s", readResp.Diagnostics)
	}

	var afterDrift targetModel
	if diags := readResp.State.Get(context.Background(), &afterDrift); diags.HasError() {
		t.Fatalf("Get drift state: %s", diags)
	}

	// State should reflect the new API values.
	if afterDrift.Address.ValueString() != "10.99.99.99" {
		t.Errorf("expected state to reflect drifted address=10.99.99.99, got %s", afterDrift.Address.ValueString())
	}
	if afterDrift.StatusDetails.ValueString() != "drifted" {
		t.Errorf("expected status_details=drifted, got %s", afterDrift.StatusDetails.ValueString())
	}
}
