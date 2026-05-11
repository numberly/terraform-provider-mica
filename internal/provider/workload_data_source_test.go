package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// ---- helpers ----------------------------------------------------------------

// newTestWorkloadDataSource creates a workloadDataSource wired to the given mock server.
func newTestWorkloadDataSource(t *testing.T, ms *testmock.MockServer) *workloadDataSource {
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
	return &workloadDataSource{client: c}
}

// workloadDSSchema returns the parsed schema for the workload data source.
func workloadDSSchema(t *testing.T) datasource.SchemaResponse {
	t.Helper()
	d := &workloadDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// buildWorkloadDSType returns the tftypes.Object for the workload data source schema.
func buildWorkloadDSType() tftypes.Object {
	contextType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"name":           tftypes.String,
		"preset_name":    tftypes.String,
		"context":        contextType,
		"created":        tftypes.Number,
		"destroyed":      tftypes.Bool,
		"status":         tftypes.String,
		"status_details": tftypes.List{ElementType: tftypes.String},
		"time_remaining": tftypes.Number,
	}}
}

// nullWorkloadDSConfig returns a base config map with all attributes null.
func nullWorkloadDSConfig() map[string]tftypes.Value {
	contextType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	}}
	return map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, nil),
		"name":           tftypes.NewValue(tftypes.String, nil),
		"preset_name":    tftypes.NewValue(tftypes.String, nil),
		"context":        tftypes.NewValue(contextType, nil),
		"created":        tftypes.NewValue(tftypes.Number, nil),
		"destroyed":      tftypes.NewValue(tftypes.Bool, nil),
		"status":         tftypes.NewValue(tftypes.String, nil),
		"status_details": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"time_remaining": tftypes.NewValue(tftypes.Number, nil),
	}
}

// ---- tests ------------------------------------------------------------------

// TestUnit_WorkloadDataSource_Basic seeds a workload in the mock and reads it via the data source.
func TestUnit_WorkloadDataSource_Basic(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterWorkloadHandlers(ms.Mux)

	// Seed a workload for the data source to read.
	store.Seed(&client.Workload{
		ID:     "wl-ds-001",
		Name:   "ds-workload",
		Status: "ready",
		Preset: &client.WorkloadPreset{
			ID:   "preset-ds-1",
			Name: "ds-preset",
		},
		StatusDetails: []string{},
		Destroyed:     false,
	})

	d := newTestWorkloadDataSource(t, ms)
	s := workloadDSSchema(t).Schema
	objType := buildWorkloadDSType()

	cfg := nullWorkloadDSConfig()
	cfg["name"] = tftypes.NewValue(tftypes.String, "ds-workload")

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{Raw: tftypes.NewValue(objType, nil), Schema: s},
	}
	d.Read(context.Background(), datasource.ReadRequest{
		Config: tfsdk.Config{Raw: tftypes.NewValue(objType, cfg), Schema: s},
	}, readResp)

	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read returned error: %s", readResp.Diagnostics)
	}

	var model workloadDataSourceModel
	if diags := readResp.State.Get(context.Background(), &model); diags.HasError() {
		t.Fatalf("Get state: %s", diags)
	}

	if model.ID.ValueString() != "wl-ds-001" {
		t.Errorf("expected id=wl-ds-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "ds-workload" {
		t.Errorf("expected name=ds-workload, got %s", model.Name.ValueString())
	}
	if model.Status.ValueString() != "ready" {
		t.Errorf("expected status=ready, got %s", model.Status.ValueString())
	}
	if model.PresetName.ValueString() != "ds-preset" {
		t.Errorf("expected preset_name=ds-preset, got %s", model.PresetName.ValueString())
	}
	if model.Destroyed.ValueBool() {
		t.Error("expected destroyed=false, got true")
	}
	// Context is null when no context in API response — acceptable for this test fixture.
	_ = model.Context.IsNull()
}
