package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestWorkloadResource creates a workloadResource wired to the given mock server.
func newTestWorkloadResource(t *testing.T, ms *testmock.MockServer) *workloadResource {
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
	return &workloadResource{client: c}
}

// workloadResourceSchema returns the parsed schema for the workload resource.
func workloadResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &workloadResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// workloadParameterTFType returns the tftypes.Object type for a workload parameter.
func workloadParameterTFType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":                tftypes.String,
		"value_string":        tftypes.String,
		"value_bool":          tftypes.Bool,
		"value_integer":       tftypes.Number,
		"value_resource_name": tftypes.String,
		"value_resource_id":   tftypes.String,
		"value_resource_type": tftypes.String,
	}}
}

// buildWorkloadType returns the tftypes.Object for the workload resource schema.
func buildWorkloadType() tftypes.Object {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	contextType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":                         tftypes.String,
		"name":                       tftypes.String,
		"preset_name":                tftypes.String,
		"parameters":                 tftypes.List{ElementType: workloadParameterTFType()},
		"destroy_eradicate_on_delete": tftypes.Bool,
		"context":                    contextType,
		"created":                    tftypes.Number,
		"destroyed":                  tftypes.Bool,
		"status":                     tftypes.String,
		"status_details":             tftypes.List{ElementType: tftypes.String},
		"time_remaining":             tftypes.Number,
		"timeouts":                   timeoutsType,
	}}
}

// nullWorkloadConfig returns a base config map with all attributes null.
func nullWorkloadConfig() map[string]tftypes.Value {
	timeoutsType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String,
		"read":   tftypes.String,
		"update": tftypes.String,
		"delete": tftypes.String,
	}}
	contextType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":                         tftypes.NewValue(tftypes.String, nil),
		"name":                       tftypes.NewValue(tftypes.String, nil),
		"preset_name":                tftypes.NewValue(tftypes.String, nil),
		"parameters":                 tftypes.NewValue(tftypes.List{ElementType: workloadParameterTFType()}, nil),
		"destroy_eradicate_on_delete": tftypes.NewValue(tftypes.Bool, nil),
		"context":                    tftypes.NewValue(contextType, nil),
		"created":                    tftypes.NewValue(tftypes.Number, nil),
		"destroyed":                  tftypes.NewValue(tftypes.Bool, nil),
		"status":                     tftypes.NewValue(tftypes.String, nil),
		"status_details":             tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"time_remaining":             tftypes.NewValue(tftypes.Number, nil),
		"timeouts":                   tftypes.NewValue(timeoutsType, nil),
	}
}

// workloadPlanWith returns a tfsdk.Plan with the required fields set.
func workloadPlanWith(t *testing.T, name, presetName string, eradicate bool) tfsdk.Plan {
	t.Helper()
	s := workloadResourceSchema(t).Schema
	cfg := nullWorkloadConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, name)
	cfg["preset_name"] = tftypes.NewValue(tftypes.String, presetName)
	cfg["destroy_eradicate_on_delete"] = tftypes.NewValue(tftypes.Bool, eradicate)
	cfg["parameters"] = tftypes.NewValue(tftypes.List{ElementType: workloadParameterTFType()}, []tftypes.Value{})
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(buildWorkloadType(), cfg),
		Schema: s,
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_WorkloadResource_Lifecycle: create → verify state → read → delete.
func TestUnit_WorkloadResource_Lifecycle(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterWorkloadHandlers(ms.Mux)

	r := newTestWorkloadResource(t, ms)
	s := workloadResourceSchema(t).Schema

	// Step 1: Create.
	plan := workloadPlanWith(t, "wl-lifecycle", "test-preset", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildWorkloadType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create returned error: %s", createResp.Diagnostics)
	}

	var afterCreate workloadModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}
	if afterCreate.ID.IsNull() || afterCreate.ID.ValueString() == "" {
		t.Error("expected non-empty id after Create")
	}
	if afterCreate.Name.ValueString() != "wl-lifecycle" {
		t.Errorf("expected name=wl-lifecycle, got %s", afterCreate.Name.ValueString())
	}
	if afterCreate.Status.ValueString() != "ready" {
		t.Errorf("expected status=ready, got %s", afterCreate.Status.ValueString())
	}
	if afterCreate.Destroyed.ValueBool() {
		t.Error("expected destroyed=false after Create")
	}

	// Step 2: Read (idempotence check).
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var afterRead workloadModel
	if diags := readResp.State.Get(context.Background(), &afterRead); diags.HasError() {
		t.Fatalf("Get read state: %s", diags)
	}
	if afterRead.Name.ValueString() != afterCreate.Name.ValueString() {
		t.Errorf("name changed on Read: create=%s read=%s", afterCreate.Name.ValueString(), afterRead.Name.ValueString())
	}

	// Step 3: Delete (soft-delete only, no eradication).
	deleteResp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{State: readResp.State}, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete returned error: %s", deleteResp.Diagnostics)
	}
}

// TestUnit_WorkloadResource_Import: create → import by name → check state populated.
func TestUnit_WorkloadResource_Import(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	handlers.RegisterWorkloadHandlers(ms.Mux)

	r := newTestWorkloadResource(t, ms)
	s := workloadResourceSchema(t).Schema

	// Create first.
	plan := workloadPlanWith(t, "wl-import", "import-preset", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildWorkloadType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	// Import by name.
	importResp := &resource.ImportStateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildWorkloadType(), nil), Schema: s},
	}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "wl-import"}, importResp)
	if importResp.Diagnostics.HasError() {
		t.Fatalf("ImportState returned error: %s", importResp.Diagnostics)
	}

	var model workloadModel
	if diags := importResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get import state: %s", diags)
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		t.Error("expected non-empty id after import")
	}
	if model.Name.ValueString() != "wl-import" {
		t.Errorf("expected name=wl-import, got %s", model.Name.ValueString())
	}
	if model.PresetName.ValueString() != "import-preset" {
		t.Errorf("expected preset_name=import-preset, got %s", model.PresetName.ValueString())
	}

	// Verify Read after import is idempotent.
	readResp := &resource.ReadResponse{State: importResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: importResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read after import: %s", readResp.Diagnostics)
	}
}

// TestUnit_WorkloadResource_DriftDetection: create → seed modified workload → read → verify state updated.
func TestUnit_WorkloadResource_DriftDetection(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterWorkloadHandlers(ms.Mux)

	r := newTestWorkloadResource(t, ms)
	s := workloadResourceSchema(t).Schema

	// Create.
	plan := workloadPlanWith(t, "wl-drift", "drift-preset", false)
	createResp := &resource.CreateResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(buildWorkloadType(), nil), Schema: s},
	}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create: %s", createResp.Diagnostics)
	}

	var afterCreate workloadModel
	if diags := createResp.State.Get(context.Background(), &afterCreate); diags.HasError() {
		t.Fatalf("Get create state: %s", diags)
	}

	// Simulate drift: modify status in the mock store directly.
	store.Seed(&client.Workload{
		ID:            afterCreate.ID.ValueString(),
		Name:          "wl-drift",
		Status:        "recovering", // changed outside Terraform
		StatusDetails: []string{"component X recovering"},
		Destroyed:     false,
	})

	// Read to detect drift.
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(context.Background(), resource.ReadRequest{State: createResp.State}, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read drift detection: %s", readResp.Diagnostics)
	}

	var afterDrift workloadModel
	if diags := readResp.State.Get(context.Background(), &afterDrift); diags.HasError() {
		t.Fatalf("Get drift state: %s", diags)
	}

	// State should reflect the new API values.
	if afterDrift.Status.ValueString() != "recovering" {
		t.Errorf("expected state to reflect drifted status=recovering, got %s", afterDrift.Status.ValueString())
	}
}
